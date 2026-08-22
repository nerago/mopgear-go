package simrank

import (
	"cmp"
	"math"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
)

// const c_accuracy_statistical_critical_value = 1.6449 // 1.6449 corresponds to 10% false equals cases
const c_accuracy_statistical_critical_value = 1.4

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
func compareSimsStatisticalByType(a *stats.SimData, b *stats.SimData, simType stats.SimType) int {
	detailA := a.GetDetailed2(simType)
	detailB := b.GetDetailed2(simType)
	if detailA == nil || detailB == nil {
		panic("missing sim detail")
	}

	averageA := a.Get(simType)
	averageB := b.Get(simType)
	stdDevA := detailA.StdDev
	stdDevB := detailB.StdDev

	if equalSimsDetailStatistical(averageA, stdDevA, averageB, stdDevB) {
		return 0
	} else {
		return cmp.Compare(averageA, averageB)
	}
}

//func compareSimsStatisticalByTypeOrig(a *stats.SimData, b *stats.SimData, simType stats.SimType) int {
//	averageA, _, _, stdDevA, hasDetailA := a.GetDetailed(simType)
//	averageB, _, _, stdDevB, hasDetailB := b.GetDetailed(simType)
//	//iterationsA, iterationsB := float64(a.SimIterations), float64(b.SimIterations)
//
//	if !hasDetailA || !hasDetailB {
//		panic("missing sim detail")
//	}
//
//	if equalSimsDetailStatistical(averageA, stdDevA, averageB, stdDevB) {
//		return 0
//	} else {
//		return cmp.Compare(averageA, averageB)
//	}
//}

func equalSimsDetailStatistical(averageA, stdDevA, averageB, stdDevB float64) bool {
	if stdDevA == 0 || stdDevB == 0 {
		return util.FloatsApproxEquals(averageA, averageB)
	}

	zScore := (averageA - averageB) / math.Sqrt((stdDevA*stdDevA)+(stdDevB*stdDevB))
	if math.Abs(zScore) > c_accuracy_statistical_critical_value {
		// exceed z score, these sims are different
		return false
	} else {
		// null hypothesis accepted, these sim results are equal
		return true
	}
}
