package stathighs

import (
	"iter"
	"math"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

type GridStatWeightProcess1B struct {
	printer *util.PrintRecorder

	targetRatios  stats.SimData
	inputData     []WeightInput
	requiredStats []stats.StatType
	simTypes      []stats.SimType
	testMode      bool
	EXTRAMODE     int

	build           utilhighs.LinearBuilder
	unitStatValues  util.MapMapSlice[stats.StatType, stats.SimType, float64]
	scales          util.MapMap[stats.StatType, stats.SimType, float64]
	detailedWeights util.MapMap[stats.StatType, stats.SimType, utilhighs.ColumnIndex]
	finalWeights    map[stats.StatType]utilhighs.ColumnIndex
}

func (grid *GridStatWeightProcess1B) Init(printer *util.PrintRecorder) {
	grid.printer = printer
	grid.build.Minimise = true
	grid.build.Solver = utilhighs.Solver_LP_USE_GPU
	grid.build.DisablePreSolve = true
	grid.build.TimeLimitSeconds = 120
	grid.finalWeights = make(map[stats.StatType]utilhighs.ColumnIndex)
}

func (grid *GridStatWeightProcess1B) SupplyData(inputData []WeightInput) {
	if grid.testMode {
		grid.inputData = inputData[0:10]
	} else {
		grid.inputData = inputData
	}
}

func (grid *GridStatWeightProcess1B) SetRequiredStats(requiredStats []stats.StatType) {
	grid.requiredStats = requiredStats
}

func (grid *GridStatWeightProcess1B) SetTargetRatios(targetRatios stats.SimData) {
	sum := 0.0
	for simType, ratio := range targetRatios.Seq() {
		if ratio >= 0 {
			grid.simTypes = append(grid.simTypes, simType)
		}
		sum += ratio
	}
	if !utilhighs.FloatEqualsOne(sum) {
		panic("ratios don't add to one")
	}

	grid.targetRatios = targetRatios
}

func (grid *GridStatWeightProcess1B) SetTestMode(testMode bool) {
	grid.testMode = testMode
	if testMode {
		grid.build.TimeLimitSeconds = 60
	}
}

func (grid *GridStatWeightProcess1B) Run() WeightResult {
	grid.setupWeightVars()
	grid.dataSamplesFromPairs()
	grid.chooseScalesBySim()
	grid.unitValuesToCalcDetailedRatings()
	grid.finalWeightVars()

	solution := grid.build.RunHighs(grid.printer)
	grid.printer.Println(solution.Status.String())

	grid.build.DebugPrintColumns(solution, grid.printer)

	return grid.reportOutputWeightsGrid(solution, grid.finalWeights, grid.printer)
}

func (grid *GridStatWeightProcess1B) setupWeightVars() {
	for _, statType := range grid.requiredStats {
		colFinalWeight := grid.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "FINAL WEIGHT: " + statType.Name()})
		grid.finalWeights[statType] = colFinalWeight
	}

	for _, statType := range grid.requiredStats {
		for _, simType := range grid.simTypes {
			colDetailWeight := grid.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "WEIGHT: " + statType.Name() + " " + simType.Name()})
			grid.detailedWeights.Put(statType, simType, colDetailWeight)
		}
	}

	baseStat := grid.requiredStats[0]
	for _, simType := range grid.simTypes {
		value := grid.targetRatios.Get(simType)
		colDetailWeight := grid.detailedWeights.GetOrPanic(baseStat, simType)
		strengthSetToRatio := utilhighs.ConstraintRow{}
		strengthSetToRatio.Add(colDetailWeight, 1)
		strengthSetToRatio.Build(&grid.build, value, value)
	}
}

func (grid *GridStatWeightProcess1B) finalWeightVars() {
	for _, statType := range grid.requiredStats {
		statFinalRow := utilhighs.ConstraintRow{}
		for simType, detailColumn := range grid.detailedWeights.SeqInnerWithKey1Value(statType) {
			_ = simType
			// scale := grid.scales.GetOrPanic(statType, simType)
			statFinalRow.Add(detailColumn, 1) // 1/scale[100] is too small, =scale[100] is too big. but all harmless when fixed like that
		}

		finalWeightColumn := grid.finalWeights[statType]
		statFinalRow.Add(finalWeightColumn, -1)
		statFinalRow.Build(&grid.build, 0, 0)
	}
}

// lazy func, could avoid double processing and put in order, very N**2
func (grid *GridStatWeightProcess1B) dataSamplesFromPairs() {
	for a := range grid.inputData {
		for b := a + 1; b < len(grid.inputData); b++ {
			statType, isGood := grid.hasOneStatDifferent(&grid.inputData[a].TotalStat, &grid.inputData[b].TotalStat)
			if isGood {
				grid.prepareSample(statType, &grid.inputData[a], &grid.inputData[b])
			}
		}
	}
}

func (grid *GridStatWeightProcess1B) hasOneStatDifferent(blockHigh, blockLow *stats.StatBlock) (stats.StatType, bool) {
	var changedStat stats.StatType
	foundChange := false
	for stat := range blockHigh {
		if blockHigh[stat] != blockLow[stat] {
			if !foundChange {
				foundChange = true
				changedStat = stats.StatType(stat)
			} else {
				// another inequality is fail
				return changedStat, false
			}
		}
	}
	return changedStat, foundChange
}

func (grid *GridStatWeightProcess1B) prepareSample(statType stats.StatType, high, low *WeightInput) {
	statDiff := high.TotalStat.GetFloat(statType) - low.TotalStat.GetFloat(statType)

	for _, simType := range grid.simTypes {
		var simValueDiff float64
		if simType.IsHighGood() {
			simValueDiff = high.SimResult.GetFriendly(simType) - low.SimResult.GetFriendly(simType)
		} else {
			simValueDiff = low.SimResult.GetFriendly(simType) - high.SimResult.GetFriendly(simType)
		}
		unitStatValue := simValueDiff / statDiff

		grid.unitStatValues.Add(statType, simType, unitStatValue)
	}
}

func (grid *GridStatWeightProcess1B) chooseScalesBySim() {
	scaleTarget := 1.0
	for _, simType := range grid.simTypes {
		minValue, maxValue := math.MaxFloat64, 0.0
		total := 0.0
		count := 0
		for value := range grid.unitStatValues.ValuesForKey2AsSeq(simType) {
			value = math.Abs(value)
			minValue = min(minValue, value)
			maxValue = max(maxValue, value)
			total += value
			count++
		}

		var scale float64
		if count != 0 {
			// average := total / float64(count)
			switch grid.EXTRAMODE {
			case 0:
				average := total / float64(count)
				scale = scaleTarget / average
			case 1:
				scale = scaleTarget / maxValue
			default:
				scale = scaleTarget / minValue
			}
			scale = util.Clamp(scale, 1e-5, 1e5)
		} else {
			scale = 1
		}

		for _, statType := range grid.requiredStats {
			grid.scales.Put(statType, simType, scale)
			// grid.scales.Put(statType, simType, 10) // 100 range, marginally better, 0.001 marginally worse
		}

		valueArray := slices.Collect(grid.unitStatValues.ValuesForKey2AsSeq(simType))
		slices.Sort(valueArray)
		grid.printer.Printf("[")
		for i := range valueArray {
			v := valueArray[i] * scale
			grid.printer.Printf("%f ", v)
		}
		grid.printer.Println("]")
	}
}

func (grid *GridStatWeightProcess1B) chooseScalesEachCombo() {
	scaleTarget := 1.0
	for group := range grid.unitStatValues.SeqGroupsKeysNestedValueSeq() {
		minValue, maxValue := math.MaxFloat64, 0.0
		total := 0.0
		count := 0
		for value := range group.ValueSeq {
			value = math.Abs(value)
			minValue = min(minValue, value)
			maxValue = max(maxValue, value)
			total += value
			count++
		}

		var scale float64
		if count != 0 {
			switch grid.EXTRAMODE {
			case 0:
				average := total / float64(count)
				scale = scaleTarget / average
			case 1:
				scale = scaleTarget / maxValue
			default:
				scale = scaleTarget / minValue
			}
			scale = scaleTarget / minValue
			scale = util.Clamp(scale, 1e-5, 1e5)
		} else {
			scale = 1
		}

		grid.scales.Put(group.Key1, group.Key2, scale)

		valueArray := slices.Collect(group.ValueSeq)
		slices.Sort(valueArray)
		grid.printer.Printf("[")
		for i := range valueArray {
			v := valueArray[i] * scale
			grid.printer.Printf("%f ", v)
		}
		grid.printer.Println("]")
	}
}

func (grid *GridStatWeightProcess1B) unitValuesToCalcDetailedRatings() {
	baseStat := grid.requiredStats[0]
	for simType, lookupStat := range grid.unitStatValues.SeqGroupsKey2Lookup() {
		unitValueBaseSeq := lookupStat(baseStat)
		for _, thisStatType := range grid.requiredStats {
			if thisStatType != baseStat {
				thisUnitValueSeq := lookupStat(thisStatType)
				grid.unitValuesCalcForGroup(simType, thisStatType, unitValueBaseSeq, thisUnitValueSeq)
			}
		}
	}
}

func (grid *GridStatWeightProcess1B) unitValuesCalcForGroup(simType stats.SimType, thisStatType stats.StatType, baseUnitValueSeq iter.Seq[float64], thisUnitValueSeq iter.Seq[float64]) {
	debugText := simType.Name() + " " + thisStatType.Name()
	baseStat := grid.requiredStats[0]
	baseDetailWeightCol := grid.detailedWeights.GetOrPanic(baseStat, simType)
	thisDetailWeightCol := grid.detailedWeights.GetOrPanic(thisStatType, simType)

	baseScale := grid.scales.GetOrPanic(baseStat, simType)
	thisScale := grid.scales.GetOrPanic(thisStatType, simType)
	if baseScale != thisScale {
		panic("what")
	}

	index := 0
	for baseUnitSample := range baseUnitValueSeq {
		for thisUnitSample := range thisUnitValueSeq {
			var debugText string = debugText + " " + strconv.Itoa(index)
			offsetAbs := grid.build.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: "OFFSET ABS " + debugText})

			grid.build.AbsoluteValueFromDiffTwoVars_ScaleOutput(thisDetailWeightCol, baseUnitSample*baseScale, baseDetailWeightCol, thisUnitSample*thisScale, offsetAbs, 1/baseScale, "OFFSET ABS "+debugText)

			index++
		}
	}
}

func (grid *GridStatWeightProcess1B) reportOutputWeightsGrid(solution *highs.Solution, weightColumns map[stats.StatType]utilhighs.ColumnIndex, printer *util.PrintRecorder) WeightResult {
	result := WeightResult_Make()
	printer.Println("FINAL WEIGHTS:")
	for _, statType := range grid.requiredStats {
		columnIndex := weightColumns[statType]
		value := solution.ColValues[columnIndex]
		printer.Printf("%10s %f\n", statType.Name(), value)
		result.Put(statType, value)
	}
	return result
}
