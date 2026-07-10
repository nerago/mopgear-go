package weightfind

import (
	"iter"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_rank"
	"paladin_gearing_go/weightfind/weight_highs"
)

const (
	c_search1_min  = -1.0
	c_search1_max  = 10.0
	c_search1_step = 3.13
	//c_search1_step        = 2.13
	c_search1_tweak_start = 0.1
)

// let's say we have 8 stats
// let's say we want to search about 1million samples, so 8th root = 5.62, about 6 samples per stat

type WeightSearcher1 struct {
	weightStats      []stats.StatType
	targetRatio      stats.SimData
	evaluateAccuracy EvaluateAccuracyPrepared
	printer          *util.PrintRecorder
}

func (ws *WeightSearcher1) Init(weightStats []stats.StatType, targetRatio stats.SimData, printer *util.PrintRecorder) {
	ws.weightStats = weightStats
	ws.targetRatio = targetRatio
	ws.printer = printer
}

func (ws *WeightSearcher1) SupplyData(inputData []weight_highs.WeightInput) {
	ws.evaluateAccuracy.Init(inputData, ws.targetRatio)
}

func (ws *WeightSearcher1) Run(cancel util_async.CancelSignal) weight_highs.WeightResult {
	progress := 0

	bestCandidates := util_rank.HighestCollector_ForN[weight_highs.WeightResult](128, (*weight_highs.WeightResult).Equals)
	for possibleWeight := range ws.makeSpacedWeights() {
		accuracy := ws.evaluateAccuracy.EvaluateWeight(possibleWeight)
		bestCandidates.Offer(&possibleWeight, accuracy)
		if progress%100 == 0 {
			_, bestAccuracy := bestCandidates.GetBest1()
			ws.printer.Printf("%6d %6.3f %6.3f\n", progress, accuracy, bestAccuracy)
		}
		progress++

		if cancel.ShouldFinish() {
			break
		}
	}

	progress = 0
	bestResult := util_rank.BestCollector1[weight_highs.WeightResult]{}
	for checkWeight := range bestCandidates.ResultsSeq() {
		updatedWeight, updatedAccuracy := weightTweaker_internal_FastCached(*checkWeight, c_search1_tweak_start, ws.weightStats, &ws.evaluateAccuracy)
		bestResult.Offer(&updatedWeight, updatedAccuracy)
		ws.printer.Printf("%6d %6.3f %6.3f\n", progress, updatedAccuracy, bestResult.BestValue)
		progress++

		if cancel.ShouldFinish() {
			break
		}
	}

	bestWeight := bestResult.GetBestOrNilValue()
	return bestWeight.ScaleForBaseStat(ws.weightStats[0])
}

func (ws *WeightSearcher1) makeSpacedWeights() iter.Seq[weight_highs.WeightResult] {
	return func(yield func(weight_highs.WeightResult) bool) {
		ws.buildSpacedWeightsRecur(ws.weightStats, weight_highs.WeightResult_Make(), yield)
	}
}

func (ws *WeightSearcher1) buildSpacedWeightsRecur(weightStats []stats.StatType, current weight_highs.WeightResult, yield func(weight_highs.WeightResult) bool) bool {
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
