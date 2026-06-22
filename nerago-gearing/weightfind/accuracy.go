package weightfind

import (
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

func EvaluateAccuracy(statWeights stathighs.WeightResult, inputData []stathighs.WeightInput, simRatios stats.SimData) float64 {

	// TODO take into acccoount sim's uncertainty ranges
	// make structures
	type accuracyInfo struct {
		input *stathighs.WeightInput

		// simRankDetail        map[stats.SimType]int
		combinedSimRankScore float64

		statRankRange util.HiLoInt
		simRankRange  util.HiLoInt
	}
	accuracyData := util.MapSliceAsNew(inputData, func(input *stathighs.WeightInput) accuracyInfo {
		return accuracyInfo{
			input: input,
			// simRankDetail: make(map[stats.SimType]int),
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
			// entry.simRankDetail[simType] = simDetailRank
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
