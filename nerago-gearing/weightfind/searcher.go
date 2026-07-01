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

type WeightSearcher1 struct {
	weightStats []stats.StatType
	targetRatio stats.SimData
	inputData   []stathighs.WeightInput
	printer     *util.PrintRecorder
}

func (ws *WeightSearcher1) Init(weightStats []stats.StatType, targetRatio stats.SimData, inputData []stathighs.WeightInput, printer *util.PrintRecorder) {
	ws.weightStats = weightStats
	ws.targetRatio = targetRatio
	ws.inputData = inputData
	ws.printer = printer
}

func (ws *WeightSearcher1) Run() stathighs.WeightResult {
	best := util_rank.BestCollector1[stathighs.WeightResult]{}
	for initialWeight := range ws.makeSpacedWeights() {
		updatedWeight, updatedAccuracy := weightTweakerInternal(initialWeight, c_search_tweak_start, ws.weightStats, ws.targetRatio, ws.inputData, ws.printer)
		best.Offer(&updatedWeight, updatedAccuracy)
	}
	return best.GetBestOrPanic()
}

func (ws *WeightSearcher1) makeSpacedWeights() iter.Seq[stathighs.WeightResult] {
	return func(yield func(stathighs.WeightResult) bool) {
		ws.buildSpacedWeightsRecur(ws.weightStats, stathighs.WeightResult_Make(), yield)
	}
}

func (ws *WeightSearcher1) buildSpacedWeightsRecur(weightStats []stats.StatType, current stathighs.WeightResult, yield func(stathighs.WeightResult) bool) bool {
	if len(weightStats) == 0 {
		return yield(current)
	}

	statAdd := weightStats[0]
	statsRemain := weightStats[1:]

	for value := c_search_min; value <= c_search_max; value += c_search_step {
		next := current.Clone()
		next.Put(statAdd, value)
		if !ws.buildSpacedWeightsRecur(statsRemain, next, yield) {
			return false
		}
	}

	return true
}
