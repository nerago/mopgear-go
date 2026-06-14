package stathighs

import (
	"math"
	"paladin_gearing_go/simulate"
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

func chooseSimScaling(inputData []WeightInput, printer *util.PrintRecorder) map[simulate.SimType]float64 {
	return chooseScalingNumbers(inputData,
		G_RequiredSims,
		func(data *WeightInput, simType simulate.SimType) float64 { return data.SimResult.Get(simType) },
		printer)
}

func chooseStatScaling(inputData []WeightInput, printer *util.PrintRecorder) map[stats.StatType]float64 {
	return chooseScalingNumbers(inputData,
		G_RequiredStats,
		func(data *WeightInput, statType stats.StatType) float64 { return data.TotalStat.GetFloat(statType) },
		printer)
}

func chooseStatScalingAll(inputData []WeightInput, printer *util.PrintRecorder) float64 {
	c_targetNumber := 1.0
	max := 0.0
	for _, check := range G_RequiredStats {
		for _, data := range inputData {
			value := data.TotalStat.GetFloat(check)
			if value > max {
				max = value
			}
		}
	}

	scale := c_targetNumber / max
	printer.Printf("scale %e\n", scale)
	return scale
}

func chooseScalingNumbers[E enumWithName](inputData []WeightInput, checkTypes []E, getValue func(*WeightInput, E) float64, printer *util.PrintRecorder) map[E]float64 {
	c_targetNumber := 1.0
	scaleMap := make(map[E]float64)
	for _, check := range checkTypes {
		max := 0.0
		for data := range util.ForPointer(inputData) {
			value := getValue(data, check)
			if value > max {
				max = value
			}
		}

		scale := c_targetNumber / max
		scaleMap[check] = scale

		printer.Printf("scale %s %e\n", check.Name(), scaleMap[check])
	}
	return scaleMap
}
