package weightfind

import (
	"cmp"
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_highs"
	"slices"
)

const c_accuracy_statistical_critical_value = 1.6449 // 1.6449 corresponds to 10% false equals cases

type accuracyInfoSimStatRanged struct {
	statScore     float64
	simScore      float64
	statRankRange *util.HiLoInt
	simRankRange  *util.HiLoInt
	dataSim       *stats.SimData
}

func EvaluateAccuracyRanged(statWeights weight_highs.WeightResult, requiredSims []stats.SimType, simRatios stats.SimData, inputData []weight_highs.WeightInput) float64 {
	if statWeights.IsEmpty() {
		return 0
	}

	data := util.MapSliceAsNew(inputData, func(input *weight_highs.WeightInput) *accuracyInfoSimStatRanged {
		return &accuracyInfoSimStatRanged{
			dataSim:       &input.SimResult,
			statScore:     statWeights.CalcStatScore(input),
			simScore:      0,
			statRankRange: nil,
			simRankRange:  nil,
		}
	})

	// rank stats scores
	slices.SortFunc(data, func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(a.statScore, b.statScore)
	})
	data[0].statRankRange = &util.HiLoInt{Lo: 0, Hi: 0}
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].statScore, data[i-1].statScore) {
			prevRange := data[i-1].statRankRange
			data[i].statRankRange = prevRange
			prevRange.Hi = i
		} else {
			data[i].statRankRange = &util.HiLoInt{Lo: i, Hi: i}
		}
	}

	// score each sim
	for _, simType := range requiredSims {
		slices.SortFunc(data, simSortRangedCompares[simType])
		ratio := simRatios.Get(simType)
		for rank := range data {
			entry := data[rank]
			entry.simScore += float64(rank) * ratio
		}
	}

	// rank combined sims
	slices.SortFunc(data, func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(a.simScore, b.simScore)
	})
	data[0].simRankRange = &util.HiLoInt{Lo: 0, Hi: 0}
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].simScore, data[i-1].simScore) {
			prevRange := data[i-1].simRankRange
			data[i].simRankRange = prevRange
			prevRange.Hi = i
		} else {
			data[i].simRankRange = &util.HiLoInt{Lo: i, Hi: i}
		}
	}

	// compute average difference between stat rank and sim rank
	sumRatioScores := 0.0
	for i := range data {
		entry := data[i]
		ratioScore := rangesToAccuracyRatio(*entry.simRankRange, *entry.statRankRange, len(data))
		sumRatioScores += ratioScore
	}
	return 100.0 * sumRatioScores / float64(len(data))
}

func EvaluateAccuracyStatisticalDeviations(statWeights weight_highs.WeightResult, requiredSims []stats.SimType, simRatios stats.SimData, inputData []weight_highs.WeightInput) float64 {
	if statWeights.IsEmpty() {
		return 0
	}

	data := util.MapSliceAsNew(inputData, func(input *weight_highs.WeightInput) *accuracyInfoSimStatRanged {
		return &accuracyInfoSimStatRanged{
			dataSim:       &input.SimResult,
			statScore:     statWeights.CalcStatScore(input),
			simScore:      0,
			statRankRange: nil,
			simRankRange:  nil,
		}
	})

	// rank stats scores
	slices.SortFunc(data, func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(a.statScore, b.statScore)
	})
	data[0].statRankRange = &util.HiLoInt{Lo: 0, Hi: 0}
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].statScore, data[i-1].statScore) {
			prevRange := data[i-1].statRankRange
			data[i].statRankRange = prevRange
			prevRange.Hi = i
		} else {
			data[i].statRankRange = &util.HiLoInt{Lo: i, Hi: i}
		}
	}

	// score each sim
	for _, simType := range requiredSims {
		if simType == stats.Sim_DEATH {
			// death data never has detail
			slices.SortFunc(data, simSortRangedCompares[stats.Sim_DEATH])
		} else if simType.IsHighGood() {
			slices.SortFunc(data, func(a, b *accuracyInfoSimStatRanged) int {
				return deviationCompareSims(a.dataSim, b.dataSim, simType)
			})
		} else {
			slices.SortFunc(data, func(a, b *accuracyInfoSimStatRanged) int {
				return -1 * deviationCompareSims(a.dataSim, b.dataSim, simType)
			})
		}
		ratio := simRatios.Get(simType)
		for rank := range data {
			entry := data[rank]
			entry.simScore += float64(rank) * ratio
		}
	}

	// rank combined sims
	slices.SortFunc(data, func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(a.simScore, b.simScore)
	})
	data[0].simRankRange = &util.HiLoInt{Lo: 0, Hi: 0}
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].simScore, data[i-1].simScore) {
			prevRange := data[i-1].simRankRange
			data[i].simRankRange = prevRange
			prevRange.Hi = i
		} else {
			data[i].simRankRange = &util.HiLoInt{Lo: i, Hi: i}
		}
	}

	// compute average difference between stat rank and sim rank
	sumRatioScores := 0.0
	for i := range data {
		entry := data[i]
		ratioScore := rangesToAccuracyRatio(*entry.simRankRange, *entry.statRankRange, len(data))
		sumRatioScores += ratioScore
	}
	return 100.0 * sumRatioScores / float64(len(data))
}

func deviationCompareSims(a *stats.SimData, b *stats.SimData, simType stats.SimType) int {
	//averageA, minA, maxA, stdDevA, hasDetailA := a.GetDetailed(simType)
	//averageB, minB, maxB, stdDevB, hasDetailB := b.GetDetailed(simType)
	averageA, _, _, stdDevA, hasDetailA := a.GetDetailed(simType)
	averageB, _, _, stdDevB, hasDetailB := b.GetDetailed(simType)
	iterationsA, iterationsB := float64(a.SimIterations), float64(b.SimIterations)

	if !hasDetailA || !hasDetailB || iterationsA == 0 || iterationsB == 0 {
		panic("missing sim detail")
	}

	// quick exit if no overlap to ranges
	//if maxA < minB {
	//	return -1
	//} else if minA > maxB {
	//	return 1
	//}

	// null hypothesis: these sim results are equal
	// significance 10% means wrongly rejecting a true null hypothesis happens 10%
	//                        wrongly rejecting a true equality happens 10%
	//                        wrongly accepting an inequality happens 10%
	// z score that exceed critical value mean reject the null, these sims are different
	// even 10% is stricter than we may want

	// the two-sample z-test
	// z = (x̄₁ − x̄₂) / √(σ₁²/n₁ + σ₂²/n₂)
	// σ² is population variance = sum of squares of differences / N
	// σ = stdev = sqrt(sum of squares of differences / N)
	// from wowsim the figure I have is: base.Stdev = math.Sqrt(base.AggregatorData.SumSq/float64(base.AggregatorData.N) - base.Avg*base.Avg)

	// z = (x̄₁ − x̄₂) / √(σ₁²/n₁ + σ₂²/n₂)
	//   = (meanA - meanB) / sqrt(deviationA²/iterationsA + deviationB²/iterationsB)
	//zScore := (averageA - averageB) / math.Sqrt((stdDevA*stdDevA)/iterationsA+(stdDevB*stdDevB)/iterationsB)

	// values are more sensible and as expected without the iterations division which may be redundant anyway since we usually have the same sample size
	zScore := (averageA - averageB) / math.Sqrt((stdDevA*stdDevA)+(stdDevB*stdDevB))

	if math.Abs(zScore) > c_accuracy_statistical_critical_value {
		// exceed z score, these sims are different
		return cmp.Compare(averageA, averageB)
	} else {
		// null hypothesis accepted, these sim results are equal
		return 0
	}
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

//goland:noinspection DuplicatedCode
var simSortRangedCompares = [6]func(a, b *accuracyInfoSimStatRanged) int{
	func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_DPS), b.dataSim.Get(stats.Sim_DPS))
	},
	func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_TPS), b.dataSim.Get(stats.Sim_TPS))
	},
	func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_DTPS), a.dataSim.Get(stats.Sim_DTPS))
	},
	func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_HPS), b.dataSim.Get(stats.Sim_HPS))
	},
	func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_TMI), a.dataSim.Get(stats.Sim_TMI))
	},
	func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_DEATH), a.dataSim.Get(stats.Sim_DEATH))
	},
}
