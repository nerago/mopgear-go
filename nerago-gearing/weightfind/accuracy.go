package weightfind

import (
	"cmp"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/weightfind/simrank"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
)

func EvaluateAccuracy[W weight_types.IWeight](statWeights W, requiredSims []stats.SimType, simRatios *weight_types.SimPriorityBasic, inputData []weight_types.WeightInput) float64 {
	if statWeights.IsEmpty() {
		return 0
	}
	data := evaluateStatScoreAndCreateStructure(statWeights, inputData)
	deriveStatRanks(data)
	simrank.RankSimsRegularForAccuracyRanged(requiredSims, data, simRatios)
	return calcAverageDifference(data)
}

func EvaluateAccuracyStatistical[W weight_types.IWeight](statWeights W, requiredSims []stats.SimType, simRatios *weight_types.SimPriorityBasic, inputData []weight_types.WeightInput) float64 {
	if statWeights.IsEmpty() {
		return 0
	}
	data := evaluateStatScoreAndCreateStructure(statWeights, inputData)
	deriveStatRanks(data)
	simrank.RankSimsStatisticalForAccuracyRanged(requiredSims, data, simRatios)
	return calcAverageDifference(data)
}

func evaluateStatScoreAndCreateStructure[W weight_types.IWeight](statWeights W, inputData []weight_types.WeightInput) []*weight_types.AccuracyInfo {
	return util_collection.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *weight_types.AccuracyInfo {
		return &weight_types.AccuracyInfo{
			DataSim:       &input.SimResult,
			StatScore:     statWeights.CalcStatScore(&input.TotalStat),
			SimScore:      0,
			StatRankRange: nil,
			SimRankRange:  nil,
		}
	})
}

func deriveStatRanks(data []*weight_types.AccuracyInfo) {
	// rank stats scores
	slices.SortFunc(data, func(a, b *weight_types.AccuracyInfo) int {
		return cmp.Compare(a.StatScore, b.StatScore)
	})
	data[0].StatRankRange = &util_collection.HiLoInt{Lo: 0, Hi: 0}
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].StatScore, data[i-1].StatScore) {
			prevRange := data[i-1].StatRankRange
			data[i].StatRankRange = prevRange
			prevRange.Hi = i
		} else {
			data[i].StatRankRange = &util_collection.HiLoInt{Lo: i, Hi: i}
		}
	}
}

func calcAverageDifference(data []*weight_types.AccuracyInfo) float64 {
	// compute average difference between stat rank and sim rank
	sumRatioScores := 0.0
	for i := range data {
		entry := data[i]
		ratioScore := rangesToAccuracyRatio(*entry.SimRankRange, *entry.StatRankRange, len(data))
		sumRatioScores += ratioScore
	}

	return checkValue(100.0 * sumRatioScores / float64(len(data)))
}
func checkValue(value float64) float64 {
	if value < 0 || value >= 100.0 {
		//panic("accuracy value out of expected range")
	}
	return value
}

// 100% if ranks are equal, 90% if average 10% difference, etc
func rangesToAccuracyRatio(one, two util_collection.HiLoInt, fullLength int) float64 {
	var diff int
	if one.Overlap(two) {
		return 1.0
	} else if one.Hi < two.Lo {
		diff = two.Lo - one.Hi
	} else if two.Hi < one.Lo {
		diff = one.Lo - two.Hi
	} else {
		panic("logic issue")
	}

	return float64(fullLength-diff) / float64(fullLength)
}
