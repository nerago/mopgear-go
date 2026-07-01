package weightfind

import (
	"cmp"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
)

func EvaluateAccuracy(statWeights stathighs.WeightResult, inputData []stathighs.WeightInput, simRatios stats.SimData) float64 {
	if statWeights.IsEmpty() {
		return 0
	}

	// TODO take into account sim's uncertainty ranges
	// make structures
	type accuracyInfo struct {
		input *stathighs.WeightInput

		combinedSimRankScore float64

		statRankRange util.HiLoInt
		simRankRange  util.HiLoInt
	}
	accuracyData := util.MapSliceAsNew(inputData, func(input *stathighs.WeightInput) accuracyInfo {
		return accuracyInfo{
			input: input,
		}
	})

	// score stats
	for entry, statRank := range util.CalculateRankingRanges(true, accuracyData, func(x *accuracyInfo) float64 { return statWeights.CalcStatScore(x.input) }) {
		entry.statRankRange = statRank
	}

	// score each sim
	requiredSims := simRatios.NonZeroTypes()
	for _, simType := range requiredSims {
		for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), accuracyData, func(x *accuracyInfo) float64 { return x.input.SimResult.Get(simType) }) {
			entry.combinedSimRankScore += float64(simDetailRank) * simRatios.Get(simType)
		}
	}

	// rank combined sims
	for entry, simRank := range util.CalculateRankingRanges(true, accuracyData, func(x *accuracyInfo) float64 { return x.combinedSimRankScore }) {
		entry.simRankRange = simRank
	}

	// compute average difference between stat rank and sim rank
	totalComparePercents := 0.0
	for info := range util.ForPointer(accuracyData) {
		percentScore := rangePercentDiff(info.simRankRange, info.statRankRange, len(accuracyData))
		totalComparePercents += percentScore
	}
	return totalComparePercents / float64(len(accuracyData))
}

// 100% if ranks are equal, 90% if average 10% difference, etc
func rangePercentDiff(one, two util.HiLoInt, fullLength int) float64 {
	var diff int
	if one.Overlap(two) {
		return 100.0
	} else if one.Hi < two.Lo {
		diff = two.Lo - one.Hi
	} else if two.Hi < one.Lo {
		diff = one.Lo - two.Hi
	} else {
		panic("logic issue")
	}

	diffAsRatio := float64(diff) / float64(fullLength)
	percentScore := 100.0 - (diffAsRatio * 100.0)
	return percentScore
}

// doesn't understand ranges
func EvaluateAccuracyOriginal(statWeights stathighs.WeightResult, inputData []stathighs.WeightInput, simRatios stats.SimData) float64 {
	type accuracyInfo struct {
		input *stathighs.WeightInput

		simRankDetail        map[stats.SimType]int
		combinedSimRankScore float64

		statRank int
		simRank  int
	}
	accuracyData := util.MapSliceAsNew(inputData, func(input *stathighs.WeightInput) accuracyInfo {
		return accuracyInfo{
			input:         input,
			simRankDetail: make(map[stats.SimType]int),
		}
	})

	// score stats
	for entry, statRank := range util.CalculateRanking(true, accuracyData, func(x *accuracyInfo) float64 { return statWeights.CalcStatScore(x.input) }) {
		entry.statRank = statRank
	}

	// score each sim
	requiredSims := simRatios.NonZeroTypes()
	for _, simType := range requiredSims {
		for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), accuracyData, func(x *accuracyInfo) float64 { return x.input.SimResult.Get(simType) }) {
			entry.simRankDetail[simType] = simDetailRank
			entry.combinedSimRankScore += float64(simDetailRank) * simRatios.Get(simType)
		}
	}

	// rank combined sims
	for entry, simRank := range util.CalculateRanking(true, accuracyData, func(x *accuracyInfo) float64 { return x.combinedSimRankScore }) {
		entry.simRank = simRank
	}

	// compute average difference between stat rank and sim rank
	totalComparePercents := 0.0
	for info := range util.ForPointer(accuracyData) {
		// 100% if ranks are equal, 90% if average 10% difference, etc
		diff := util.AbsIntDiff(info.simRank, info.statRank)
		diffAsRatio := float64(diff) / float64(len(accuracyData))
		percentScore := 100.0 - (diffAsRatio * 100.0)
		totalComparePercents += percentScore
	}

	return totalComparePercents / float64(len(accuracyData))
}

func EvaluateAccuracyInlined(statWeights stathighs.WeightResult, inputData []stathighs.WeightInput, simRatios stats.SimData) float64 {
	if statWeights.IsEmpty() {
		return 0
	}

	requiredSims := simRatios.NonZeroTypes()

	type accuracyInfo struct {
		input                *stathighs.WeightInput
		combinedSimRankScore float64
		statRankRange        util.HiLoInt
		simRankRange         util.HiLoInt
	}
	accuracyData := util.MapSliceAsNew(inputData, func(input *stathighs.WeightInput) accuracyInfo {
		return accuracyInfo{
			input: input,
		}
	})

	// score stats using preovider weights
	for entry, statRank := range util.CalculateRankingRanges(true, accuracyData,
		func(x *accuracyInfo) float64 { return statWeights.CalcStatScore(x.input) }) {
		entry.statRankRange = statRank
	}

	// score each sim
	type internalEntry struct {
		score   float64
		pointer *accuracyInfo
	}
	rankArray := make([]internalEntry, len(accuracyData))
	for _, simType := range requiredSims {
		for i := range len(accuracyData) {
			data := &accuracyData[i]
			rankArray[i] = internalEntry{
				score:   data.input.SimResult.Get(simType),
				pointer: data,
			}
		}

		if simType.IsHighGood() {
			slices.SortFunc(rankArray, func(a, b internalEntry) int { return cmp.Compare(a.score, b.score) })
		} else {
			slices.SortFunc(rankArray, func(a, b internalEntry) int { return cmp.Compare(b.score, a.score) })
		}

		for simRank := range rankArray {
			entry := &rankArray[simRank]
			entry.pointer.combinedSimRankScore += float64(simRank) * simRatios.Get(simType)
		}
	}

	// rank combined sims
	for entry, simRank := range util.CalculateRankingRanges(true, accuracyData, func(x *accuracyInfo) float64 { return x.combinedSimRankScore }) {
		entry.simRankRange = simRank
	}

	// compute average difference between stat rank and sim rank
	totalComparePercents := 0.0
	for info := range util.ForPointer(accuracyData) {
		percentScore := rangePercentDiff(info.simRankRange, info.statRankRange, len(accuracyData))
		totalComparePercents += percentScore
	}
	return totalComparePercents / float64(len(accuracyData))
}
