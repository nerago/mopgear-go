package weightfind

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_rank"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
	"iter"
	"math/rand"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_search0_min        = -0.5
	c_search0_max        = 5.5
	c_search0_tweakStart = 0.1
)

type WeightSearcher0 struct {
	weightStats         []stats.StatType
	targetRatio         weight_types.SimPriorityBasic
	evaluateAccuracy    EvaluateAccuracyPrepared
	printer             *util.PrintRecorder
	AccuracyStatistical bool
}

func (ws *WeightSearcher0) Init(weightStats []stats.StatType, targetRatio weight_types.SimPriorityBasic, printer *util.PrintRecorder) {
	ws.weightStats = weightStats
	ws.targetRatio = targetRatio
	ws.printer = printer
}

func (ws *WeightSearcher0) SupplyData(inputData []weight_types.WeightInput) {
	ws.evaluateAccuracy.Init(inputData, &ws.targetRatio, ws.AccuracyStatistical)
}

func (ws *WeightSearcher0) Run(cancel util_async.CancelSignal) weight_types.WeightResult {
	stopwatch := util.StopwatchMakeStarted()
	best := util_rank.BestCollector1[weight_types.Weight1Basic]{}
	progress := 0
	for initialWeight := range ws.makeRandomWeights(5000) {
		updatedWeight, updatedAccuracy := weightTweaker_internal_FastCached(initialWeight, c_search0_tweakStart, ws.weightStats, &ws.evaluateAccuracy)
		best.Offer(&updatedWeight, updatedAccuracy)
		if progress%1000 == 0 {
			ws.printer.Printf("%6d %6.3f %6.3f\n", progress, updatedAccuracy, best.BestValue)
		}
		progress++

		if cancel.ShouldFinish() {
			break
		}
	}
	bestWeight := best.GetBestOrNilValue()
	bestWeight.NormalizeForBase(ws.weightStats)
	return weight_types.WeightResult{Weight: &bestWeight, SolveTime: stopwatch.Elapsed(), Status: highs.ModelStatusOptimal}
}

func (ws *WeightSearcher0) makeRandomWeights(count int) iter.Seq[weight_types.Weight1Basic] {
	return func(yield func(weight_types.Weight1Basic) bool) {
		rng := rand.New(rand.NewSource(rand.Int63()))
		for range count {
			weight := ws.buildWeightsRandom(rng)
			if !yield(weight) {
				return
			}
		}
	}
}

func (ws *WeightSearcher0) buildWeightsRandom(rng *rand.Rand) weight_types.Weight1Basic {
	weight := weight_types.Weight1Basic_Make()
	for _, statType := range ws.weightStats {
		value := c_search0_min + rng.Float64()*(c_search0_max-c_search0_min)
		weight.Put(statType, value)
	}
	return weight
}

//accuracy = 84.770372
//Duration = 4m16.736176s
//88.901605 Duration = 4m9.1343515s
