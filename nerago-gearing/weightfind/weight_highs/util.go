package weight_highs

import (
	"iter"
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

func isGoodValueRange(value float64) bool {
	value = math.Abs(value)
	return 1e-6 <= value && value <= 1e6
}

type enumWithName interface {
	~uint8
	Name() string
}

func chooseSimScalingUnfriendly(inputData []WeightInput, scaleTarget float64, printer *util.PrintRecorder) map[stats.SimType]float64 {
	return chooseScalingNumbers(inputData,
		stats.SimTypeList,
		func(data *WeightInput, simType stats.SimType) float64 { return data.SimResult.Get(simType) },
		scaleTarget,
		printer)
}

func chooseStatScaling(inputData []WeightInput, scaleTarget float64, printer *util.PrintRecorder) map[stats.StatType]float64 {
	return chooseScalingNumbers(inputData,
		stats.StatType_List,
		func(data *WeightInput, statType stats.StatType) float64 { return data.TotalStat.GetFloat(statType) },
		scaleTarget,
		printer)
}

func chooseScalingNumbers[E enumWithName](inputData []WeightInput, checkTypes []E, getValue func(*WeightInput, E) float64, scaleTarget float64, printer *util.PrintRecorder) map[E]float64 {
	scaleMap := make(map[E]float64)
	for _, check := range checkTypes {
		valueSeq := util.MapSliceAsSeq(inputData, func(x *WeightInput) float64 {
			return getValue(x, check)
		})

		scale := chooseScale(valueSeq, scaleTarget)
		scaleMap[check] = scale

		printer.Printf("scale %s %e\n", check.Name(), scaleMap[check])
	}
	return scaleMap
}

func chooseScale(seq iter.Seq[float64], scaleTarget float64) float64 {
	minPosValue, maxPosValue := math.MaxFloat64, 0.0
	minNegValue, maxNegValue := math.MaxFloat64, 0.0
	hasNeg, hasPos, hasZero := false, false, false
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

	var scale float64
	if hasPos && hasNeg {
		superMax := max(maxNegValue, maxPosValue)
		scale = scaleTarget / superMax
	} else if hasPos && !hasZero {
		scale = scaleTarget / minPosValue
	} else if hasPos {
		scale = scaleTarget / maxPosValue
	} else if hasNeg && !hasZero {
		scale = scaleTarget / minNegValue
	} else if hasNeg {
		scale = scaleTarget / maxNegValue
	} else {
		scale = 1
	}

	scale = util.Clamp(scale, 1e-5, 1e5)
	return scale
}
