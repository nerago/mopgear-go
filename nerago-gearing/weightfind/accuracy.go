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

	diffAsRatio := float64(fullLength-diff) / float64(fullLength)
	percentScore := diffAsRatio * 100.0
	return percentScore
}

func EvaluateAccuracyNoRangeInlined1(statWeights stathighs.WeightResult, requiredSims []stats.SimType, simRatios stats.SimData, inputData []stathighs.WeightInput) float64 {
	type accuracyInfo struct {
		statScore float64
		simScore  float64
		statRank  int32
		simRank   int32
		input     stathighs.WeightInput
	}
	accuracyData := make([]accuracyInfo, len(inputData))
	for i := range len(inputData) {
		input := &inputData[i]
		accuracyData[i] = accuracyInfo{
			input:     *input,
			statScore: statWeights.CalcStatScore(input),
		}
	}

	// rank stats scores
	slices.SortFunc(accuracyData, func(a, b accuracyInfo) int { return cmp.Compare(a.statScore, b.statScore) })
	for rank := range accuracyData {
		entry := &accuracyData[rank]
		entry.statRank = int32(rank)
	}

	// score each sim
	for _, simType := range requiredSims {
		if simType.IsHighGood() {
			slices.SortFunc(accuracyData, func(a, b accuracyInfo) int {
				return cmp.Compare(a.input.SimResult.Get(simType), b.input.SimResult.Get(simType))
			})
		} else {
			slices.SortFunc(accuracyData, func(a, b accuracyInfo) int {
				return cmp.Compare(b.input.SimResult.Get(simType), a.input.SimResult.Get(simType))
			})
		}

		ratio := simRatios.Get(simType)
		for rank := range accuracyData {
			entry := &accuracyData[rank]
			entry.simScore += float64(rank) * ratio
		}
	}

	// rank combined sims
	slices.SortFunc(accuracyData, func(a, b accuracyInfo) int { return cmp.Compare(a.simScore, b.simScore) })
	for rank := range accuracyData {
		entry := &accuracyData[rank]
		entry.simRank = int32(rank)
	}

	// compute average difference between stat rank and sim rank
	totalComparePercents := 0.0
	for i := range accuracyData {
		entry := &accuracyData[i]
		diff := util.AbsInt32Diff(entry.simRank, entry.statRank)
		diffAsRatio := float64(int32(len(accuracyData))-diff) / float64(len(accuracyData))
		percentScore := diffAsRatio
		totalComparePercents += percentScore
	}

	return 100.0 * totalComparePercents / float64(len(accuracyData))
}

type accuracyInfoX struct {
	statScore float64
	simScore  float64
	statRank  int32
	simRank   int32
	input     *stathighs.WeightInput
}

func compareStatScores(a, b *accuracyInfoX) int { return cmp.Compare(a.statScore, b.statScore) }
func compareSimScores(a, b *accuracyInfoX) int  { return cmp.Compare(a.simScore, b.simScore) }

func EvaluateAccuracyNoRangeInlined2(statWeights stathighs.WeightResult, requiredSims []stats.SimType, simRatios stats.SimData, inputData []stathighs.WeightInput) float64 {
	accuracyData := buildAccuracySlice(statWeights, inputData)

	// rank stats scores
	slices.SortFunc(accuracyData, compareStatScores)
	calcStatRanks(accuracyData)

	// score each sim
	for _, simType := range requiredSims {
		calcOneSim(simType, accuracyData, simRatios)
	}

	// rank combined sims
	slices.SortFunc(accuracyData, compareSimScores)
	calcSimRanks(accuracyData)

	// compute average difference between stat rank and sim rank
	return finalResult(accuracyData)
}

func finalResult(accuracyData []*accuracyInfoX) float64 {
	totalComparePercents := 0.0
	for i := range accuracyData {
		percentScore := calcDiff(accuracyData[i], len(accuracyData))
		totalComparePercents += percentScore
	}

	return 100.0 * totalComparePercents / float64(len(accuracyData))
}

func calcDiff(entry *accuracyInfoX, size int) float64 {
	diff := util.AbsInt32Diff(entry.simRank, entry.statRank)
	diffAsRatio := float64(int32(size)-diff) / float64(size)
	percentScore := diffAsRatio
	return percentScore
}

func calcSimRanks(accuracyData []*accuracyInfoX) {
	for rank := range accuracyData {
		entry := accuracyData[rank]
		entry.simRank = int32(rank)
	}
}

func calcOneSim(simType stats.SimType, accuracyData []*accuracyInfoX, simRatios stats.SimData) {
	sortSimType(accuracyData, simType)

	addToSimScore(simRatios, simType, accuracyData)
}

var simSortFuncs = [6]func(a, b *accuracyInfoX) int{
	func(a, b *accuracyInfoX) int {
		return cmp.Compare(a.input.SimResult.Get(stats.Sim_DPS), b.input.SimResult.Get(stats.Sim_DPS))
	},
	func(a, b *accuracyInfoX) int {
		return cmp.Compare(a.input.SimResult.Get(stats.Sim_TPS), b.input.SimResult.Get(stats.Sim_TPS))
	},
	func(a, b *accuracyInfoX) int {
		return cmp.Compare(b.input.SimResult.Get(stats.Sim_DTPS), a.input.SimResult.Get(stats.Sim_DTPS))
	},
	func(a, b *accuracyInfoX) int {
		return cmp.Compare(a.input.SimResult.Get(stats.Sim_HPS), b.input.SimResult.Get(stats.Sim_HPS))
	},
	func(a, b *accuracyInfoX) int {
		return cmp.Compare(b.input.SimResult.Get(stats.Sim_TMI), a.input.SimResult.Get(stats.Sim_TMI))
	},
	func(a, b *accuracyInfoX) int {
		return cmp.Compare(b.input.SimResult.Get(stats.Sim_DEATH), b.input.SimResult.Get(stats.Sim_DEATH))
	},
}

func sortSimType(accuracyData []*accuracyInfoX, simType stats.SimType) {
	slices.SortFunc(accuracyData, simSortFuncs[simType])
}

func addToSimScore(simRatios stats.SimData, simType stats.SimType, accuracyData []*accuracyInfoX) {
	ratio := simRatios.Get(simType)
	for rank := range accuracyData {
		entry := accuracyData[rank]
		entry.simScore += float64(rank) * ratio
	}
}

func calcStatRanks(accuracyData []*accuracyInfoX) {
	for rank := range accuracyData {
		entry := accuracyData[rank]
		entry.statRank = int32(rank)
	}
}

func buildAccuracySlice(statWeights stathighs.WeightResult, inputData []stathighs.WeightInput) []*accuracyInfoX {
	accuracyData := make([]*accuracyInfoX, len(inputData))
	for i := range len(inputData) {
		input := &inputData[i]
		accuracyData[i] = &accuracyInfoX{
			input:     input,
			statScore: statWeights.CalcStatScore(input),
		}
	}
	return accuracyData
}

func EvaluateAccuracyFullRangeInlined(statWeights stathighs.WeightResult, requiredSims []stats.SimType, simRatios stats.SimData, inputData []stathighs.WeightInput) float64 {
	type accuracyInfo struct {
		input     *stathighs.WeightInput
		statScore float64
		simScore  float64
		statRank  *util.HiLoInt
		simRank   *util.HiLoInt
	}
	accuracyData := make([]*accuracyInfo, len(inputData))
	for i := range len(inputData) {
		input := &inputData[i]
		accuracyData[i] = &accuracyInfo{
			input:     input,
			statScore: statWeights.CalcStatScore(input),
		}
	}

	// rank stats scores
	slices.SortFunc(accuracyData, func(a, b *accuracyInfo) int { return cmp.Compare(a.statScore, b.statScore) })
	{
		prevScore := accuracyData[0].statScore
		prevHiLo := &util.HiLoInt{Lo: 0, Hi: 0}
		for index := range accuracyData {
			entry := accuracyData[index]
			if util.FloatsApproxEquals(entry.statScore, prevScore) {
				prevHiLo.Hi = index
				entry.statRank = prevHiLo
			} else {
				prevHiLo = &util.HiLoInt{Lo: index, Hi: index}
				entry.statRank = prevHiLo
			}
			prevScore = entry.statScore
		}
	}

	// score each sim
	for _, simType := range requiredSims {
		if simType.IsHighGood() {
			slices.SortFunc(accuracyData, func(a, b *accuracyInfo) int {
				return cmp.Compare(a.input.SimResult.Get(simType), b.input.SimResult.Get(simType))
			})
		} else {
			slices.SortFunc(accuracyData, func(a, b *accuracyInfo) int {
				return cmp.Compare(b.input.SimResult.Get(simType), a.input.SimResult.Get(simType))
			})
		}

		ratio := simRatios.Get(simType)
		for rank := range accuracyData {
			entry := accuracyData[rank]
			entry.simScore += float64(rank) * ratio
		}
	}

	// rank combined sims
	slices.SortFunc(accuracyData, func(a, b *accuracyInfo) int { return cmp.Compare(a.simScore, b.simScore) })
	{
		prevScore := accuracyData[0].simScore
		prevHiLo := &util.HiLoInt{Lo: 0, Hi: 0}
		for index := range accuracyData {
			entry := accuracyData[index]
			if util.FloatsApproxEquals(entry.simScore, prevScore) {
				prevHiLo.Hi = index
				entry.simRank = prevHiLo
			} else {
				prevHiLo = &util.HiLoInt{Lo: index, Hi: index}
				entry.simRank = prevHiLo
			}
			prevScore = entry.simScore
		}
	}

	// compute average difference between stat rank and sim rank
	totalComparePercents := 0.0
	for i := range accuracyData {
		entry := accuracyData[i]
		percentScore := rangePercentDiff(*entry.simRank, *entry.statRank, len(accuracyData))
		totalComparePercents += percentScore
	}

	return 100.0 * totalComparePercents / float64(len(accuracyData))
}
