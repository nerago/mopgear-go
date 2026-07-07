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
	weightStats []stats.StatType
	targetRatio stats.SimData
	inputData   []stathighs.WeightInput
	printer     *util.PrintRecorder
}

func (ws *WeightSearcher0) Init(weightStats []stats.StatType, targetRatio stats.SimData, printer *util.PrintRecorder) {
	ws.weightStats = weightStats
	ws.targetRatio = targetRatio
	ws.printer = printer
}

func (ws *WeightSearcher0) SupplyData(inputData []stathighs.WeightInput) {
	ws.inputData = inputData
}

func (ws *WeightSearcher0) Run(cancel channel_op.CancelSignal) stathighs.WeightResult {
	simTypes := ws.targetRatio.NonZeroTypes()
	best := util_rank.BestCollector1[stathighs.WeightResult]{}
	progress := 0
	for initialWeight := range ws.makeRandomWeights(100) {
		updatedWeight, updatedAccuracy := weightTweakerInternal_FastNoRange(initialWeight, c_search0_tweak_start, ws.weightStats, simTypes, ws.targetRatio, ws.inputData)
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
