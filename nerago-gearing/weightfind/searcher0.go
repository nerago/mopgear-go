package weightfind

import (
	"iter"
	"math/rand"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_rank"
	"paladin_gearing_go/weightfind/weight_highs"
)

const (
	c_search0_min         = -0.5
	c_search0_max         = 5.5
	c_search0_tweak_start = 0.1
)

type WeightSearcher0 struct {
	weightStats      []stats.StatType
	targetRatio      stats.SimData
	evaluateAccuracy EvaluateAccuracyPrepared
	printer          *util.PrintRecorder
}

func (ws *WeightSearcher0) Init(weightStats []stats.StatType, targetRatio stats.SimData, printer *util.PrintRecorder) {
	ws.weightStats = weightStats
	ws.targetRatio = targetRatio
	ws.printer = printer
}

func (ws *WeightSearcher0) SupplyData(inputData []weight_highs.WeightInput) {
	ws.evaluateAccuracy.Init(inputData, ws.targetRatio)
}

func (ws *WeightSearcher0) Run(cancel util_async.CancelSignal) weight_highs.WeightResult {
	best := util_rank.BestCollector1[weight_highs.WeightResult]{}
	progress := 0
	for initialWeight := range ws.makeRandomWeights(20000) {
		updatedWeight, updatedAccuracy := weightTweaker_internal_FastCached(initialWeight, c_search0_tweak_start, ws.weightStats, &ws.evaluateAccuracy)
		best.Offer(&updatedWeight, updatedAccuracy)
		ws.printer.Printf("%6d %6.3f %6.3f\n", progress, updatedAccuracy, best.BestValue)
		progress++

		if cancel.ShouldFinish() {
			break
		}
	}
	bestWeight := best.GetBestOrNilValue()
	return bestWeight.ScaleForBaseStat(ws.weightStats[0])
}

func (ws *WeightSearcher0) makeRandomWeights(count int) iter.Seq[weight_highs.WeightResult] {
	return func(yield func(weight_highs.WeightResult) bool) {
		rng := rand.New(rand.NewSource(rand.Int63()))
		for range count {
			weight := ws.buildWeightsRandom(rng)
			if !yield(weight) {
				return
			}
		}
	}
}

func (ws *WeightSearcher0) buildWeightsRandom(rng *rand.Rand) weight_highs.WeightResult {
	weight := weight_highs.WeightResult_Make()
	for _, statType := range ws.weightStats {
		value := c_search0_min + rng.Float64()*(c_search0_max-c_search0_min)
		weight.Put(statType, value)
	}
	return weight
}

//accuracy = 84.770372
//Duration = 4m16.736176s
//88.901605 Duration = 4m9.1343515s
