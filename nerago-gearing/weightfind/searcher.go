package weightfind

import (
	"iter"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
)

const (
	c_search_min         = -0.5
	c_search_max         = 5.5
	c_search_step        = 1.0
	c_search_tweak_start = 0.1
)

// let's say we have 8 stats
// let's say we want to search about 1million samples, so 8th root = 5.62, about 6 samples per stat

func WeightSearcher(weightStats []stats.StatType, targetRatio stats.SimData, inputData []stathighs.WeightInput, printer *util.PrintRecorder) stathighs.WeightResult {
	best := util_rank.BestCollector1[stathighs.WeightResult]{}
	for initialWeight := range makeSpacedWeights(weightStats) {
		updatedWeight, updatedAccuracy := weightTweakerInternal(initialWeight, c_search_tweak_start, weightStats, targetRatio, inputData, printer)
		best.Offer(&updatedWeight, updatedAccuracy)
	}
	return best.GetBestOrPanic()
}

func makeSpacedWeights(weightStats []stats.StatType) iter.Seq[stathighs.WeightResult] {
	return func(yield func(stathighs.WeightResult) bool) {
		buildSpacedWeightsRecur(weightStats, stathighs.WeightResult_Make(), yield)
	}
}

func buildSpacedWeightsRecur(weightStats []stats.StatType, current stathighs.WeightResult, yield func(stathighs.WeightResult) bool) bool {
	if len(weightStats) == 0 {
		return yield(current)
	}

	statAdd := weightStats[0]
	statsRemain := weightStats[1:]

	for value := c_search_min; value <= c_search_max; value += c_search_step {
		next := current.Clone()
		next.Put(statAdd, value)
		if !buildSpacedWeightsRecur(statsRemain, next, yield) {
			return false
		}
	}

	return true
}
