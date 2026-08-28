package weightfind

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type SearcherModel interface {
	comboCount() int8
	initialRanges() []weight_types.StatRangeFloat
	newLocals() ModelThreadLocals
	makeFinalResult(point *pointType) any
}

type ModelThreadLocals interface {
	evaluateScore(weightArray *pointType) float64
}

type ModelWeight struct {
	statTypes               []stats.StatType
	simTypes                []stats.SimType
	targetRatio             weight_types.SimPriorityBasic
	initialEvaluateAccuracy weightfind.EvaluateAccuracyPrepared
}

type ModelWeightLocal struct {
	main             *ModelWeight
	evaluateAccuracy *weightfind.EvaluateAccuracyPrepared
}

func (mw *ModelWeight) Init(statTypes []stats.StatType, targetRatio weight_types.SimPriorityBasic) {
	mw.statTypes = statTypes
	mw.simTypes = targetRatio.SimTypes()
	mw.targetRatio = targetRatio
}

func (mw *ModelWeight) SupplyData(inputData []weight_types.WeightInput) {
	mw.initialEvaluateAccuracy.Init(inputData, &mw.targetRatio, true, true)
}

func (mw *ModelWeight) comboCount() int8 {
	return int8(len(mw.statTypes) * len(mw.simTypes))
}

func (mw *ModelWeight) initialRanges() []weight_types.StatRangeFloat {
	rangeSlice := make([]weight_types.StatRangeFloat, mw.comboCount())
	for i := range mw.comboCount() {
		rangeSlice[i].Minimum = -1.0
		rangeSlice[i].Maximum = 1.0
	}
	return rangeSlice
}

func (mw *ModelWeight) newLocals() ModelThreadLocals {
	return &ModelWeightLocal{
		main:             mw,
		evaluateAccuracy: mw.initialEvaluateAccuracy.Clone(),
	}
}

func (loc *ModelWeightLocal) evaluateScore(weightArray *pointType) float64 {
	main := loc.main
	weights := weight_types.Weight2Extended_Make(main.simTypes, main.statTypes)

	index := 0
	for _, statType := range main.statTypes {
		for _, simType := range main.simTypes {
			weights.PutWeight(simType, statType, weightArray[index])
			index++
		}
	}
	for _, simType := range main.simTypes {
		ratio := main.targetRatio.GetOrPanic(simType)
		weights.SetSimScale(simType, 1, 0, ratio)
	}

	accuracy := loc.evaluateAccuracy.EvaluateWeight2(weights)
	return accuracy
}
