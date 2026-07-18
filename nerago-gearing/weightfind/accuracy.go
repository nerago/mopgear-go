package weightfind

import (
	"cmp"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/simrank"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
)

func EvaluateAccuracyRanged(statWeights weight_types.WeightBasic, requiredSims []stats.SimType, simRatios stats.SimData, inputData []weight_types.WeightInput) float64 {
	if statWeights.IsEmpty() {
		return 0
	}

	data := util.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *weight_types.AccuracyInfoSimStatRanged {
		return &weight_types.AccuracyInfoSimStatRanged{
			DataSim:       &input.SimResult,
			StatScore:     statWeights.CalcStatScore(input),
			SimScore:      0,
			StatRankRange: nil,
			SimRankRange:  nil,
		}
	})

	// rank stats scores
	slices.SortFunc(data, func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
		return cmp.Compare(a.StatScore, b.StatScore)
	})
	data[0].StatRankRange = &util.HiLoInt{Lo: 0, Hi: 0}
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].StatScore, data[i-1].StatScore) {
			prevRange := data[i-1].StatRankRange
			data[i].StatRankRange = prevRange
			prevRange.Hi = i
		} else {
			data[i].StatRankRange = &util.HiLoInt{Lo: i, Hi: i}
		}
	}

	simrank.RankSimsRegularForAccuracyRanged(requiredSims, data, simRatios)

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

func EvaluateAccuracyStatisticalDeviations(statWeights weight_types.WeightBasic, requiredSims []stats.SimType, simRatios stats.SimData, inputData []weight_types.WeightInput) float64 {
	if statWeights.IsEmpty() {
		return 0
	}

	data := util.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *weight_types.AccuracyInfoSimStatRanged {
		return &weight_types.AccuracyInfoSimStatRanged{
			DataSim:       &input.SimResult,
			StatScore:     statWeights.CalcStatScore(input),
			SimScore:      0,
			StatRankRange: nil,
			SimRankRange:  nil,
		}
	})

	// rank stats scores
	slices.SortFunc(data, func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
		return cmp.Compare(a.StatScore, b.StatScore)
	})
	data[0].StatRankRange = &util.HiLoInt{Lo: 0, Hi: 0}
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].StatScore, data[i-1].StatScore) {
			prevRange := data[i-1].StatRankRange
			data[i].StatRankRange = prevRange
			prevRange.Hi = i
		} else {
			data[i].StatRankRange = &util.HiLoInt{Lo: i, Hi: i}
		}
	}

	simrank.RankSimsStatisticalForAccuracyRanged(requiredSims, data, simRatios)

	// compute average difference between stat rank and sim rank
	sumRatioScores := 0.0
	for i := range data {
		entry := data[i]
		ratioScore := rangesToAccuracyRatio(*entry.SimRankRange, *entry.StatRankRange, len(data))
		sumRatioScores += ratioScore
	}

	return checkValue(100.0 * sumRatioScores / float64(len(data)))
}

// 100% if ranks are equal, 90% if average 10% difference, etc
func rangesToAccuracyRatio(one, two util.HiLoInt, fullLength int) float64 {
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
