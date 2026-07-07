package weightfind

import (
	"iter"
	"math/rand"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"paladin_gearing_go/util/util_rank"
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

func (ws *WeightSearcher0) SupplyData(inputData []stathighs.WeightInput) {
	ws.evaluateAccuracy.Init(inputData, ws.targetRatio)
}

func (ws *WeightSearcher0) Run(cancel channel_op.CancelSignal) stathighs.WeightResult {
	best := util_rank.BestCollector1[stathighs.WeightResult]{}
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

func (ws *WeightSearcher0) makeRandomWeights(count int) iter.Seq[stathighs.WeightResult] {
	return func(yield func(stathighs.WeightResult) bool) {
		rng := rand.New(rand.NewSource(rand.Int63()))
		for range count {
			weight := ws.buildWeightsRandom(rng)
			if !yield(weight) {
				return
			}
		}
	}
}

func (ws *WeightSearcher0) buildWeightsRandom(rng *rand.Rand) stathighs.WeightResult {
	weight := stathighs.WeightResult_Make()
	for _, statType := range ws.weightStats {
		value := c_search0_min + rng.Float64()*(c_search0_max-c_search0_min)
		weight.Put(statType, value)
	}
	return weight
}

//accuracy = 84.770372
//Duration = 4m16.736176s
//88.901605 Duration = 4m9.1343515s
