package util_weight

import (
	"iter"
	"math"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func ChooseSimUnfriendlyScalingBasic(inputData []weight_types.WeightInput, scaleTarget float64, keepUnderTarget bool, printer *util.PrintRecorder) stats.SimTypeMap[float64] {
	scaleMap := stats.SimTypeMap[float64]{}
	chooseScalingBasicScale(
		&scaleMap,
		inputData,
		stats.SimTypeList,
		func(data *weight_types.WeightInput, simType stats.SimType) float64 {
			return data.SimResult.Get(simType)
		},
		scaleTarget,
		keepUnderTarget,
		printer,
	)
	return scaleMap
}

func ChooseStatScalingBasic(inputData []weight_types.WeightInput, scaleTarget float64, keepUnderTarget bool, printer *util.PrintRecorder) stats.StatTypeMap[float64] {
	scaleMap := stats.StatTypeMap[float64]{}
	chooseScalingBasicScale(
		&scaleMap,
		inputData,
		stats.StatType_List,
		func(data *weight_types.WeightInput, statType stats.StatType) float64 {
			return data.TotalStat.GetFloat(statType)
		},
		scaleTarget,
		keepUnderTarget,
		printer,
	)
	return scaleMap
}

func chooseScalingBasicScale[E util_collection.EnumBaseType, M util_collection.IMap[E, float64]](scaleMap M, inputData []weight_types.WeightInput, checkTypes []E, getValue func(*weight_types.WeightInput, E) float64, scaleTarget float64, keepUnderTarget bool, printer *util.PrintRecorder) {
	for _, check := range checkTypes {
		valueSeq := util_collection.MapSliceAsSeq(inputData, func(x *weight_types.WeightInput) float64 {
			return getValue(x, check)
		})

		scale := ChooseScale(valueSeq, scaleTarget, keepUnderTarget)
		scaleMap.Put(check, scale)

		printer.Printf("scale %s %e\n", check.Name(), scaleMap.GetOrPanic(check))
	}
}

func ChooseScale(seq iter.Seq[float64], scaleTarget float64, keepUnderTarget bool) float64 {
	minPosValue, maxPosValue, minNegValue, maxNegValue, hasNeg, hasPos, hasZero := sequenceMetrics(seq)

	var scale float64
	if hasPos && hasNeg {
		superMax := max(maxNegValue, maxPosValue)
		scale = scaleTarget / superMax
	} else if hasPos && !hasZero && !keepUnderTarget {
		scale = scaleTarget / minPosValue
	} else if hasPos {
		scale = scaleTarget / maxPosValue
	} else if hasNeg && !hasZero && !keepUnderTarget {
		scale = scaleTarget / minNegValue
	} else if hasNeg {
		scale = scaleTarget / maxNegValue
	} else {
		scale = 1
	}

	scale = util.Clamp(scale, 1e-10, 1e10)
	return scale
}

func sequenceMetrics(seq iter.Seq[float64]) (minPosValue, maxPosValue, minNegValue, maxNegValue float64, hasNeg, hasPos, hasZero bool) {
	minPosValue, maxPosValue = math.MaxFloat64, 0.0
	minNegValue, maxNegValue = math.MaxFloat64, 0.0
	hasNeg, hasPos, hasZero = false, false, false
	for valueRaw := range seq {
		if util.FloatEqualsZero(valueRaw) {
			minNegValue = 0
			minPosValue = 0
			hasZero = true
		} else if valueRaw > 0 {
			minPosValue = min(minPosValue, valueRaw)
			maxPosValue = max(maxPosValue, valueRaw)
			hasPos = true
		} else {
			minNegValue = min(minNegValue, -valueRaw)
			maxNegValue = max(maxNegValue, -valueRaw)
			hasNeg = true
		}
	}
	return minPosValue, maxPosValue, minNegValue, maxNegValue, hasNeg, hasPos, hasZero
}

func ChooseSimUnfriendlyUnitScaleAndOffset(inputData []weight_types.WeightInput, simTypeList []stats.SimType) stats.SimTypeMap[weight_types.ScaleAndOffset] {
	scaleMap := stats.SimTypeMap[weight_types.ScaleAndOffset]{}
	for _, simType := range simTypeList {
		valueSeq := util_collection.MapSliceAsSeq(inputData, func(x *weight_types.WeightInput) float64 {
			return x.SimResult.Get(simType)
		})

		params := chooseUnitScaleAndOffset(valueSeq, simType.IsHighGood())
		scaleMap.Put(simType, params)
	}
	return scaleMap
}

func ChooseSimDetailUnitScaleAndOffset(inputData []weight_types.WeightInput, simTypeList []stats.SimType) stats.SimTypeMap[weight_types.ScaleAndOffset] {
	scaleMap := stats.SimTypeMap[weight_types.ScaleAndOffset]{}
	for _, simType := range simTypeList {
		var valueSeq iter.Seq[float64]

		if simType.ExpectDetail() {
			valueSeqMin := util_collection.MapSliceAsSeq(inputData, func(x *weight_types.WeightInput) float64 {
				return x.SimResult.GetDetailed2(simType).Min
			})
			valueSeqMax := util_collection.MapSliceAsSeq(inputData, func(x *weight_types.WeightInput) float64 {
				return x.SimResult.GetDetailed2(simType).Max
			})
			valueSeq = util_collection.ConcatSeq2(valueSeqMin, valueSeqMax)
		} else {
			valueSeq = util_collection.MapSliceAsSeq(inputData, func(x *weight_types.WeightInput) float64 {
				return x.SimResult.Get(simType)
			})
		}

		params := chooseUnitScaleAndOffset(valueSeq, simType.IsHighGood())
		scaleMap.Put(simType, params)
	}
	return scaleMap
}

func chooseUnitScaleAndOffset(seq iter.Seq[float64], isHighGood bool) weight_types.ScaleAndOffset {
	minPosValue, maxPosValue, minNegValue, maxNegValue, hasNeg, hasPos, hasZero := sequenceMetrics(seq)

	// TODO not sure if flipping negatives makes sense, rarely comes up though probably

	var highestActual, lowestActual float64
	// initially worked out based on isHighGood=true, so best is closest to +inf
	if hasPos && hasNeg {
		highestActual = maxPosValue
		lowestActual = -maxNegValue
	} else if hasPos && hasZero {
		highestActual = maxPosValue
		lowestActual = 0
	} else if hasPos {
		highestActual = maxPosValue
		lowestActual = minPosValue
	} else if hasNeg && hasZero {
		highestActual = 0
		lowestActual = -maxNegValue
	} else if hasNeg {
		highestActual = -minNegValue
		lowestActual = -maxNegValue
	} else {
		panic("can't determine value range for all zeros")
	}

	return weight_types.CalcScaleOffsetForUnitRange(isHighGood, highestActual, lowestActual)
}
