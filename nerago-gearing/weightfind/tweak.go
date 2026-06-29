package weightfind

import (
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
)

const (
	c_tweak_start      = 0.01
	c_tweak_limit      = 0.00001
	c_tweak_iter_count = 1000 // limit to avoid infinite loop
)

func WeightTweaker(startWeight stathighs.WeightResult, weightStats []stats.StatType, targetRatio stats.SimData, inputData []stathighs.WeightInput, printer *util.PrintRecorder) stathighs.WeightResult {
	return weightTweakerCustom(startWeight, c_tweak_start, weightStats, targetRatio, inputData, printer)
}

func weightTweakerCustom(startWeight stathighs.WeightResult, tweakStart float64, weightStats []stats.StatType, targetRatio stats.SimData, inputData []stathighs.WeightInput, printer *util.PrintRecorder) stathighs.WeightResult {
	add := tweakStart
	multiply := 1 + add
	bestWeight := startWeight.Clone()

	for range c_tweak_iter_count {
		best := util_rank.BestCollector1[stathighs.WeightResult]{}
		best.Offer(&bestWeight, EvaluateAccuracy(bestWeight, inputData, targetRatio))
		for i := 1; i < len(weightStats); i++ {
			stat := weightStats[i]
			if !bestWeight.IsZero(stat) {
				hi := bestWeight.Clone()
				hi.MultiplyEquals(stat, multiply)
				best.Offer(&hi, EvaluateAccuracy(hi, inputData, targetRatio))

				lo := bestWeight.Clone()
				lo.DivideEquals(stat, multiply)
				best.Offer(&lo, EvaluateAccuracy(lo, inputData, targetRatio))
			} else {
				hi := bestWeight.Clone()
				hi.PlusEquals(stat, add)
				best.Offer(&hi, EvaluateAccuracy(hi, inputData, targetRatio))

				lo := bestWeight.Clone()
				lo.MinusEquals(stat, add)
				best.Offer(&lo, EvaluateAccuracy(lo, inputData, targetRatio))
			}
		}
		updateWeight := best.GetBestOrPanic()

		if updateWeight.Equals(bestWeight) {
			add /= 2
			multiply = 1 + add
			if add <= c_tweak_limit {
				printer.Printf("DONE\n")
				break
			}
		} else {
			printer.Printf("NEXT %s accuracy=%f\n", updateWeight.String(), EvaluateAccuracy(updateWeight, inputData, targetRatio))
			bestWeight = updateWeight
		}
	}

	return bestWeight
}
