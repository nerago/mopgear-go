package weight_highs

import (
	"iter"
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/weightfind/weight_types"
)

func isGoodValueRange(value float64) bool {
	value = math.Abs(value)
	return 1e-6 <= value && value <= 1e6
}

func chooseSimUnfriendlyScalingBasic(inputData []weight_types.WeightInput, scaleTarget float64, keepUnderTarget bool, printer *util.PrintRecorder) util_collection.EnumMap[stats.SimType, float64] {
	return chooseScalingBasicScale(inputData,
		stats.SimTypeList,
		func(data *weight_types.WeightInput, simType stats.SimType) float64 {
			return data.SimResult.Get(simType)
		},
		scaleTarget,
		keepUnderTarget,
		printer,
		stats.SimTypeEnum)
}

func chooseStatScalingBasic(inputData []weight_types.WeightInput, scaleTarget float64, keepUnderTarget bool, printer *util.PrintRecorder) util_collection.EnumMap[stats.StatType, float64] {
	return chooseScalingBasicScale(inputData,
		stats.StatType_List,
		func(data *weight_types.WeightInput, statType stats.StatType) float64 {
			return data.TotalStat.GetFloat(statType)
		},
		scaleTarget,
		keepUnderTarget,
		printer,
		stats.StatTypeEnum)
}

func chooseScalingBasicScale[E util_collection.EnumBaseType](inputData []weight_types.WeightInput, checkTypes []E, getValue func(*weight_types.WeightInput, E) float64, scaleTarget float64, keepUnderTarget bool, printer *util.PrintRecorder, enumType util_collection.EnumType[E]) util_collection.EnumMap[E, float64] {
	scaleMap := util_collection.EnumMapMake[E, float64](enumType)
	for _, check := range checkTypes {
		valueSeq := util_collection.MapSliceAsSeq(inputData, func(x *weight_types.WeightInput) float64 {
			return getValue(x, check)
		})

		scale := chooseScale(valueSeq, scaleTarget, keepUnderTarget)
		scaleMap.Put(check, scale)

		printer.Printf("scale %s %e\n", check.Name(), scaleMap.GetOrPanic(check))
	}
	return scaleMap
}

func chooseScale(seq iter.Seq[float64], scaleTarget float64, keepUnderTarget bool) float64 {
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

type scaleAndOffset struct {
	scale  float64
	offset float64
}

func (so scaleAndOffset) Apply(value float64) float64 {
	return (value + so.offset) * so.scale
}

func chooseSimUnfriendlyUnitScaleAndOffset(inputData []weight_types.WeightInput, simTypeList []stats.SimType) util_collection.EnumMap[stats.SimType, scaleAndOffset] {
	scaleMap := util_collection.EnumMapMake[stats.SimType, scaleAndOffset](stats.SimTypeEnum)
	for _, simType := range simTypeList {
		valueSeq := util_collection.MapSliceAsSeq(inputData, func(x *weight_types.WeightInput) float64 {
			return x.SimResult.Get(simType)
		})

		params := chooseUnitScaleAndOffset(valueSeq, simType.IsHighGood())
		scaleMap.Put(simType, params)
	}
	return scaleMap
}

func chooseUnitScaleAndOffset(seq iter.Seq[float64], isHighGood bool) scaleAndOffset {
	minPosValue, maxPosValue, minNegValue, maxNegValue, hasNeg, hasPos, hasZero := sequenceMetrics(seq)

	var bestActual, worstActual float64
	// initially worked out based on isHighGood=true, so best is closest to +inf
	if hasPos && hasNeg {
		bestActual = maxPosValue
		worstActual = -maxNegValue
	} else if hasPos && hasZero {
		bestActual = maxPosValue
		worstActual = 0
	} else if hasPos {
		bestActual = maxPosValue
		worstActual = minPosValue
	} else if hasNeg && hasZero {
		bestActual = 0
		worstActual = -maxNegValue
	} else if hasNeg {
		bestActual = -minNegValue
		worstActual = -maxNegValue
	} else {
		panic("can't determine value range for all zeros")
	}

	if !isHighGood {
		// flip meaning so that best is closest to -inf
		bestActual, worstActual = worstActual, bestActual
	}

	worstTarget := 0.0
	bestTarget := 1.0

	// scaleAndOffset logic: output = (input + so.offset) * so.scale
	// worstTarget = (worstActual + offset) * scale
	//   -> worstTarget = worstActual*scale + offset*scale
	//   -> worstTarget - worstActual*scale = offset*scale
	// bestTarget = (bestActual + offset) * scale
	//   -> bestTarget = bestActual*scale + offset*scale
	//   -> bestTarget = bestActual*scale + (worstTarget - worstActual*scale)
	//   -> bestTarget - worstTarget = bestActual*scale - worstActual*scale
	//   -> bestTarget - worstTarget = (bestActual - worstActual) * scale
	//   -> (bestTarget - worstTarget) / (bestActual - worstActual) = scale
	// bestTarget = (bestActual + offset) * scale
	//   -> bestTarget / scale = bestActual + offset
	//   -> bestTarget / scale - bestActual = offset

	scale := (bestTarget - worstTarget) / (bestActual - worstActual)
	offset := (bestTarget / scale) - bestActual
	return scaleAndOffset{
		scale:  scale,
		offset: offset,
	}
}
