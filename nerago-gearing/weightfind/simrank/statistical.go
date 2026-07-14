package simrank

import (
	"cmp"
	"math"
	"paladin_gearing_go/stats"
)

// const c_accuracy_statistical_critical_value = 1.6449 // 1.6449 corresponds to 10% false equals cases
const c_accuracy_statistical_critical_value = 1.4

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
