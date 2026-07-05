package weightfind

import (
	"iter"
	"math/rand"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
)

const (
	c_search0_min         = -0.5
	c_search0_max         = 5.5
	c_search0_tweak_start = 0.1
)

type WeightSearcher0 struct {
	weightStats []stats.StatType
	targetRatio stats.SimData
	inputData   []stathighs.WeightInput
	printer     *util.PrintRecorder
}

func (ws *WeightSearcher0) Init(weightStats []stats.StatType, targetRatio stats.SimData, inputData []stathighs.WeightInput, printer *util.PrintRecorder) {
	ws.weightStats = weightStats
	ws.targetRatio = targetRatio
	ws.inputData = inputData
	ws.printer = printer
}

func (ws *WeightSearcher0) Run() stathighs.WeightResult {
	best := util_rank.BestCollector1[stathighs.WeightResult]{}
	for initialWeight := range ws.makeRandomWeights(1000000) {
		updatedWeight, updatedAccuracy := weightTweakerInternal(initialWeight, c_search0_tweak_start, ws.weightStats, ws.targetRatio, ws.inputData, ws.printer)
		best.Offer(&updatedWeight, updatedAccuracy)
	}
	return best.GetBestOrPanic()
}

func (ws *WeightSearcher0) makeRandomWeights(count int) iter.Seq[stathighs.WeightResult] {
	return func(yield func(stathighs.WeightResult) bool) {
		for range count {
			weight := ws.buildWeightsRandom()
			if !yield(weight) {
				return
			}
		}
	}
}

func (ws *WeightSearcher0) buildWeightsRandom() stathighs.WeightResult {
	weight := stathighs.WeightResult_Make()
	for _, statType := range ws.weightStats {
		value := c_search0_min + rand.Float64()*(c_search0_max-c_search0_min)
		weight.Put(statType, value)
	}
	return weight
}
