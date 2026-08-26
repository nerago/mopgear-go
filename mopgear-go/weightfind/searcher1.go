package weightfind

import (
	"iter"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_rank"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_search1_min        = -1.0
	c_search1_max        = 10.0
	c_search1_step       = 3.13
	c_search1_tweakStart = 0.1
)

// let's say we have 8 stats
// let's say we want to search about 1million samples, so 8th root = 5.62, about 6 samples per stat

type WeightSearcher1 struct {
	weightStats         []stats.StatType
	targetRatio         weight_types.SimPriorityBasic
	evaluateAccuracy    EvaluateAccuracyPrepared
	printer             *util.PrintRecorder
	AccuracyStatistical bool
}

func (ws *WeightSearcher1) Init(weightStats []stats.StatType, targetRatio weight_types.SimPriorityBasic, printer *util.PrintRecorder) {
	ws.weightStats = weightStats
	ws.targetRatio = targetRatio
	ws.printer = printer
}

func (ws *WeightSearcher1) SupplyData(inputData []weight_types.WeightInput) {
	ws.evaluateAccuracy.Init(inputData, &ws.targetRatio, ws.AccuracyStatistical, false)
}

func (ws *WeightSearcher1) Run(cancel util_async.CancelSignal) weight_types.WeightResult1 {
	progress := 0
	stopwatch := util.StopwatchMakeStarted()

	bestCandidates := util_rank.HighestCollector_ForN[weight_types.Weight1Basic](128, (*weight_types.Weight1Basic).Equals)
	for possibleWeight := range ws.makeSpacedWeights() {
		accuracy := ws.evaluateAccuracy.EvaluateWeight1(&possibleWeight)
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
	bestResult := util_rank.BestCollector1[weight_types.Weight1Basic]{}
	for checkWeight := range bestCandidates.ResultsSeq() {
		updatedWeight, updatedAccuracy := weightTweaker_internal_FastCached(*checkWeight, c_search1_tweakStart, ws.weightStats, &ws.evaluateAccuracy)
		bestResult.Offer(&updatedWeight, updatedAccuracy)
		ws.printer.Printf("%6d %6.3f %6.3f\n", progress, updatedAccuracy, bestResult.BestValue)
		progress++

		if cancel.ShouldFinish() {
			break
		}
	}

	bestWeight := bestResult.GetBestOrNilValue()
	bestWeight.NormalizeForBase(ws.weightStats)
	return weight_types.WeightResult1Make(&bestWeight, stopwatch.Elapsed(), highs.ModelStatusOptimal)
}

func (ws *WeightSearcher1) makeSpacedWeights() iter.Seq[weight_types.Weight1Basic] {
	return func(yield func(weight_types.Weight1Basic) bool) {
		ws.buildSpacedWeightsRecur(ws.weightStats, weight_types.Weight1Basic_Make(), yield)
	}
}

func (ws *WeightSearcher1) buildSpacedWeightsRecur(weightStats []stats.StatType, current weight_types.Weight1Basic, yield func(weight_types.Weight1Basic) bool) bool {
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
