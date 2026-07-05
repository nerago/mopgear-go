package weightfind

import (
	"iter"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
)

const (
	c_search1_min         = -0.5
	c_search1_max         = 5.5
	c_search1_step        = 0.5
	c_search1_tweak_start = 0.1
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
	progress := 0
	requiredSims := ws.targetRatio.NonZeroTypes()

	bestCandidates := util_rank.HighestCollector_ForN[stathighs.WeightResult](128, (*stathighs.WeightResult).Equals)
	for possibleWeight := range ws.makeSpacedWeights() {
		accuracy := EvaluateAccuracyNoRangeInlined2(possibleWeight, requiredSims, ws.targetRatio, ws.inputData)
		bestCandidates.Offer(&possibleWeight, accuracy)
		if progress%100 == 0 {
			_, bestAccuracy := bestCandidates.GetBest1()
			ws.printer.Printf("%6d %6.3f %6.3f\n", progress, accuracy, bestAccuracy)
		}
		progress++
	}

	progress = 0
	bestResult := util_rank.BestCollector1[stathighs.WeightResult]{}
	for checkWeight := range bestCandidates.ResultsSeq() {
		updatedWeight, updatedAccuracy := weightTweakerInternal_FastNoRange(*checkWeight, c_search1_tweak_start, ws.weightStats, requiredSims, ws.targetRatio, ws.inputData)
		bestResult.Offer(&updatedWeight, updatedAccuracy)
		ws.printer.Printf("%6d %6.3f %6.3f\n", progress, updatedAccuracy, bestResult.BestValue)
		progress++
	}
	return bestResult.GetBestOrPanic()
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

	for value := c_search1_min; value <= c_search1_max; value += c_search1_step {
		next := current.Clone()
		next.Put(statAdd, value)
		if !ws.buildSpacedWeightsRecur(statsRemain, next, yield) {
			return false
		}
	}

	return true
}
