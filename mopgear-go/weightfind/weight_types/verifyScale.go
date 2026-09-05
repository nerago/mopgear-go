package weight_types

import (
	"fmt"
	"math"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_rank"
)

const (
	c_verify_permittedSlack = 0.1
	c_verify_targetLoValue  = 0.0
	c_verify_targetHiValue  = 1.0
)

// scaleAndOffset logic: output = (input + so.offset) * so.scale
// worstTarget = (worstActual + offset) * scale
//
//	-> worstTarget = worstActual*scale + offset*scale
//	-> worstTarget - worstActual*scale = offset*scale
//
// bestTarget = (bestActual + offset) * scale
//
//	-> bestTarget = bestActual*scale + offset*scale
//	-> bestTarget = bestActual*scale + (worstTarget - worstActual*scale)
//	-> bestTarget - worstTarget = bestActual*scale - worstActual*scale
//	-> bestTarget - worstTarget = (bestActual - worstActual) * scale
//	-> (bestTarget - worstTarget) / (bestActual - worstActual) = scale
//
// bestTarget = (bestActual + offset) * scale
//
//	-> bestTarget / scale = bestActual + offset
//	-> bestTarget / scale - bestActual = offset
func CalcScaleOffsetForUnitRange(isHighGood bool, highestActual float64, lowestActual float64) ScaleAndOffset {
	if !isHighGood {
		// flip meaning so that best is closest to -inf
		highestActual, lowestActual = lowestActual, highestActual
	}

	worstTarget := 0.0
	bestTarget := 1.0

	scale := (bestTarget - worstTarget) / (highestActual - lowestActual)
	offset := (bestTarget / scale) - highestActual
	return ScaleAndOffset{
		Scale:  scale,
		Offset: offset,
	}
}

type ScaleAndOffset struct {
	Scale  float64
	Offset float64
}

func (so ScaleAndOffset) Apply(value float64) float64 {
	initial := (value + so.Offset) * so.Scale
	if math.Abs(initial) <= 1e-9 {
		return 0
	} else {
		return initial
	}
}

type scoreForBasicFunc func(statBlock *stats.StatBlock) float64
type scoreForSimFunc func(statBlock *stats.StatBlock, simType stats.SimType) float64

func (wbs *Weight1_ScaledSolvable) verifyGoodRange(sampleInputs []WeightInput) error {
	return verifyGoodRangeBasic1(sampleInputs, wbs.calcStatScoreRaw)
}

func (we *Weight2) verifyGoodRange(sampleInputs []WeightInput) error {
	return verifyGoodRangeGeneral(sampleInputs, we.SimList, &we.SimPriority, we.scoreForSimRaw)
}

func (wer *Weight3) verifyGoodRange(sampleInputs []WeightInput) error {
	return verifyGoodRangeGeneral(sampleInputs, wer.SimList, &wer.SimPriority, wer.scoreForSimRaw)
}

func (we *Weight2) UpdateScaling(inputData []WeightInput) error {
	return updateScalingGeneral(inputData, we.SimList, &we.SimPriority, we.scoreForSimRaw)
}

func (wer *Weight3) UpdateScaling(inputData []WeightInput) error {
	return updateScalingGeneral(inputData, wer.SimList, &wer.SimPriority, wer.scoreForSimRaw)
}

func verifyGoodRangeBasic1(sampleInputs []WeightInput, statToValueFunc scoreForBasicFunc) error {
	if len(sampleInputs) == 0 {
		return util.ErrorTracedNew("no inputs for verification")
	}

	loValue, hiValue := calcBasicScoreRangeForInputsRaw(sampleInputs, statToValueFunc)

	if util.AbsDiff(loValue, c_verify_targetLoValue) > c_verify_permittedSlack ||
		util.AbsDiff(hiValue, c_verify_targetHiValue) > c_verify_permittedSlack {
		return fmt.Errorf("weights fail to produce expected value range, actual: %f - %f", loValue, hiValue)
	}

	return nil
}

func verifyGoodRangeGeneral(sampleInputs []WeightInput, simList []stats.SimType, simPriority *SimPriorityExtended, statToSimValueFunc scoreForSimFunc) error {
	if len(sampleInputs) == 0 {
		return util.ErrorTracedNew("no inputs for verification")
	}

	for _, simType := range simList {
		loValue, hiValue := calcSimScoreRangeForInputsScaled(sampleInputs, simType, statToSimValueFunc, simPriority)
		if util.AbsDiff(loValue, c_verify_targetLoValue) > c_verify_permittedSlack ||
			util.AbsDiff(hiValue, c_verify_targetHiValue) > c_verify_permittedSlack {
			return fmt.Errorf("weights fail to produce expected value range, actual: %f - %f", loValue, hiValue)
		}
	}

	return nil
}

func updateScalingBasic1(sampleInputs []WeightInput, statToValueFunc scoreForBasicFunc) (ScaleAndOffset, error) {
	if len(sampleInputs) == 0 {
		return ScaleAndOffset{}, util.ErrorTracedNew("no inputData for scaling")
	}

	loValue, hiValue := calcBasicScoreRangeForInputsRaw(sampleInputs, statToValueFunc)

	scaleOffset := CalcScaleOffsetForUnitRange(true, hiValue, loValue)
	return scaleOffset, nil
}

func updateScalingGeneral(sampleInputs []WeightInput, simList []stats.SimType, simPriority *SimPriorityExtended, statToSimValueFunc scoreForSimFunc) error {
	if len(sampleInputs) == 0 {
		return util.ErrorTracedNew("no inputData for scaling")
	}

	for _, simType := range simList {
		oldPriorityEntry := simPriority.GetOrPanic(simType)

		loValue, hiValue := calcSimScoreRangeForInputsRaw(sampleInputs, simType, statToSimValueFunc)
		scaleOffset := CalcScaleOffsetForUnitRange(simType.IsHighGood(), hiValue, loValue)

		simPriority.Delete(simType)
		err := simPriority.SetSimScale(simType, scaleOffset, oldPriorityEntry.RatioScale)
		if err != nil {
			return err
		}
	}
	return nil
}

func calcBasicScoreRangeForInputsRaw(sampleInputs []WeightInput, statToValueFunc scoreForBasicFunc) (float64, float64) {
	lo := util_rank.BestCollector1Lite[float64]{}
	hi := util_rank.BestCollector1Lite[float64]{}

	for input := range util_collection.ForPointer(sampleInputs) {
		score := statToValueFunc(&input.TotalStat)
		lo.Offer(score, -score)
		hi.Offer(score, score)
	}

	loValue, hiValue := lo.GetBestOrNilValue(), hi.GetBestOrNilValue()
	return loValue, hiValue
}

func calcSimScoreRangeForInputsRaw(sampleInputs []WeightInput, simType stats.SimType, statToSimValueFunc scoreForSimFunc) (float64, float64) {
	lo := util_rank.BestCollector1Lite[float64]{}
	hi := util_rank.BestCollector1Lite[float64]{}

	for input := range util_collection.ForPointer(sampleInputs) {
		score := statToSimValueFunc(&input.TotalStat, simType)
		lo.Offer(score, -score)
		hi.Offer(score, score)
	}

	return lo.GetBestOrNilValue(), hi.GetBestOrNilValue()
}

func calcSimScoreRangeForInputsScaled(sampleInputs []WeightInput, simType stats.SimType, statToSimValueFunc scoreForSimFunc, simPriority *SimPriorityExtended) (float64, float64) {
	lo := util_rank.BestCollector1Lite[float64]{}
	hi := util_rank.BestCollector1Lite[float64]{}
	priorityEntry := simPriority.GetOrPanic(simType)

	for input := range util_collection.ForPointer(sampleInputs) {
		score := statToSimValueFunc(&input.TotalStat, simType)
		score = priorityEntry.ApplyRangingOnly(score)
		lo.Offer(score, -score)
		hi.Offer(score, score)
	}

	return lo.GetBestOrNilValue(), hi.GetBestOrNilValue()
}
