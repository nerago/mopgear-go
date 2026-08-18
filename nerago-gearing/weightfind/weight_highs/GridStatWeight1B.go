package weight_highs

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
	"iter"
	"math"
	"slices"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

const c_grid1b_scaleTarget = 10.0
const c_grid1b_minWeightSum = 0.01
const c_grid1b_targetWeightSum = 0.5

type GridStatWeightProcess1B struct {
	printer *util.PrintRecorder

	targetRatios  weight_types.SimPriorityBasic
	inputData     []weight_types.WeightInput
	requiredStats []stats.StatType
	simTypes      []stats.SimType
	SCALEMODE     int
	ROUNDMODE     int
	OUTLIER       int
	CALCMODE      int
	RATIO         int

	build           util_highs.LinearBuilder
	unitStatValues  util_collection.MapMapSlice[stats.StatType, stats.SimType, float64]
	scales          util_collection.MapMap[stats.StatType, stats.SimType, float64]
	detailedWeights util_collection.MapMap[stats.StatType, stats.SimType, util_highs.ColumnIndex]
	finalWeights    map[stats.StatType]util_highs.ColumnIndex
}

func (grid *GridStatWeightProcess1B) Init(printer *util.PrintRecorder, timeoutSeconds int) {
	grid.printer = printer
	grid.build.Minimise = true
	//grid.build.Solver = util_highs.Solver_Force_IPX
	//grid.build.Solver = util_highs.Solver_Force_Simplex
	grid.build.Solver = util_highs.Solver_LP_USE_GPU

	grid.build.TimeLimitSeconds = timeoutSeconds
	grid.build.AddOptionFloat("small_matrix_value", 1e-8)
	grid.finalWeights = make(map[stats.StatType]util_highs.ColumnIndex)
}

func (grid *GridStatWeightProcess1B) SupplyData(inputData []weight_types.WeightInput) {
	grid.inputData = inputData
}

func (grid *GridStatWeightProcess1B) SetRequiredStats(requiredStats []stats.StatType) {
	grid.requiredStats = requiredStats
}

func (grid *GridStatWeightProcess1B) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	targetRatios.ValidateRatioAddsToOne()
	grid.simTypes = targetRatios.SimTypes()
	grid.targetRatios = targetRatios
}

func (grid *GridStatWeightProcess1B) Run() *util_async.FutureCancellable[weight_types.WeightResult] {
	grid.setupWeightVars()
	grid.dataSamplesFromPairs()
	grid.removeOutliers()
	if grid.SCALEMODE < 3 {
		grid.chooseScalesBySim()
	} else {
		grid.chooseScalesEachCombo()
	}
	grid.unitValuesToCalcDetailedRatings()
	grid.finalWeightVars()

	stopwatch := util.StopwatchMakeStopped()
	solutionFuture := grid.build.RunHighsFuture(stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) (weight_types.WeightResult, bool) {
		solution := linearResult.GetSolutionAndSaveLog(grid.printer)

		grid.printer.Println(solution.Status.String())
		grid.build.DebugPrintColumns(solution, grid.printer)

		weight := grid.reportOutputWeightsGrid(solution, grid.finalWeights, grid.printer)
		return weight_types.WeightResult{Weight: &weight, SolveTime: stopwatch.Elapsed(), Status: solution.Status}, true
	})
}

func (grid *GridStatWeightProcess1B) setupWeightVars() {
	for _, statType := range grid.requiredStats {
		colFinalWeight := grid.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "FINAL WEIGHT: " + statType.Name()})
		grid.finalWeights[statType] = colFinalWeight
	}

	for _, statType := range grid.requiredStats {
		for _, simType := range grid.simTypes {
			colDetailWeight := grid.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "WEIGHT: " + statType.Name() + " " + simType.Name()})
			grid.detailedWeights.Put(statType, simType, colDetailWeight)
		}
	}

	if grid.RATIO == 0 {
		baseStat := grid.requiredStats[0]
		for _, simType := range grid.simTypes {
			value := grid.targetRatios.GetOrPanic(simType)
			colDetailWeight := grid.detailedWeights.GetOrPanic(baseStat, simType)
			strengthSetToRatio := util_highs.ConstraintRow{}
			strengthSetToRatio.Add(colDetailWeight, 1)
			strengthSetToRatio.Build(&grid.build, value, value)
		}
	} else if grid.RATIO == 1 {
		for _, simType := range grid.simTypes {
			weightSum := grid.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), nil)
			weightSumOut := grid.build.CreateColumnWithOutput(highs.Continuous, c_grid1b_minWeightSum, util_highs.InfPos(), 1, nil)

			sumSimWeights := util_highs.ConstraintRow{}
			for _, detailColumn := range grid.detailedWeights.SeqKey1ValueWithKey2(simType) {
				sumSimWeights.Add(detailColumn, 1)
			}
			sumSimWeights.Add(weightSum, -1)
			sumSimWeights.Build(&grid.build, 0, 0)

			grid.build.AbsoluteValue(weightSum, weightSumOut)
		}
	} else {
		for _, simType := range grid.simTypes {
			weightSum := grid.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), nil)
			weightSumOut := grid.build.CreateColumnWithOutput(highs.Continuous, c_grid1b_targetWeightSum, c_grid1b_targetWeightSum, 1, nil)

			sumSimWeights := util_highs.ConstraintRow{}
			for _, detailColumn := range grid.detailedWeights.SeqKey1ValueWithKey2(simType) {
				sumSimWeights.Add(detailColumn, 1)
			}
			sumSimWeights.Add(weightSum, -1)
			sumSimWeights.Build(&grid.build, 0, 0)

			grid.build.AbsoluteValue(weightSum, weightSumOut)
		}
	}
}

func (grid *GridStatWeightProcess1B) finalWeightVars() {
	for _, statType := range grid.requiredStats {
		statFinalRow := util_highs.ConstraintRow{}
		for simType, detailColumn := range grid.detailedWeights.SeqKey2ValueWithKey1(statType) {
			_ = simType
			// scale := grid.scales.GetOrPanic(statType, simType)
			if grid.RATIO == 3 {
				if simType.IsHighGood() {
					statFinalRow.Add(detailColumn, 1)
				} else {
					statFinalRow.Add(detailColumn, -1)
				}
			} else {
				statFinalRow.Add(detailColumn, 1)
			}
		}

		finalWeightColumn := grid.finalWeights[statType]
		statFinalRow.Add(finalWeightColumn, -1)
		statFinalRow.Build(&grid.build, 0, 0)
	}
}

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

func (grid *GridStatWeightProcess1B) prepareSample(statType stats.StatType, high, low *weight_types.WeightInput) {
	statDiff := high.TotalStat.GetFloat(statType) - low.TotalStat.GetFloat(statType)

	for _, simType := range grid.simTypes {
		var simValueDiff float64
		var unitStatValue float64

		switch grid.ROUNDMODE {
		case 0:
			if simType.IsHighGood() {
				simValueDiff = high.SimResult.GetFriendly(simType) - low.SimResult.GetFriendly(simType)
			} else {
				simValueDiff = low.SimResult.GetFriendly(simType) - high.SimResult.GetFriendly(simType)
			}
			unitStatValue = simValueDiff / statDiff
		case 1:
			simValueDiff = high.SimResult.GetFriendly(simType) - low.SimResult.GetFriendly(simType)
			unitStatValue = simValueDiff / statDiff
		default:
			simValueDiff = high.SimResult.GetFriendly(simType) - low.SimResult.GetFriendly(simType)
			unitStatValue = math.Abs(simValueDiff / statDiff)
		}

		grid.unitStatValues.Add(statType, simType, unitStatValue)
	}
}

func (grid *GridStatWeightProcess1B) chooseScalesBySim() {
	scaleTarget := 1.0
	for _, simType := range grid.simTypes {
		minPosValue, maxPosValue := math.MaxFloat64, 0.0
		minNegValue, maxNegValue := math.MaxFloat64, 0.0
		hasNeg, hasPos, hasZero := false, false, false
		total := 0.0
		count := 0
		for valueRaw := range grid.unitStatValues.SeqValuesWithKey2(simType) {
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
			total += math.Abs(valueRaw)
			count++
		}

		var scale float64
		if count != 0 {
			// scale = scaleTarget / minValue
			switch grid.SCALEMODE {
			case 0:
				scale = 1
			case 1:
				average := total / float64(count)
				scale = scaleTarget / average
			case 2:
				if hasPos && !hasNeg && !hasZero {
					scale = scaleTarget / minPosValue
				} else if hasPos && !hasNeg && hasZero {
					scale = scaleTarget / maxPosValue
				} else if !hasPos && hasNeg && !hasZero {
					scale = scaleTarget / minNegValue
				} else if !hasPos && hasNeg && hasZero {
					scale = scaleTarget / maxNegValue
				} else if hasPos && hasNeg {
					superMax := max(maxNegValue, maxPosValue)
					scale = scaleTarget / superMax
				} else {
					scale = 1
				}
			default:
				panic("unknown")
			}
			scale = util.Clamp(scale, 1e-5, 1e5)
		} else {
			scale = 1
		}

		if scale >= 1 {
			grid.build.SetEachTolerance(1e-3)
		}

		for _, statType := range grid.requiredStats {
			grid.scales.Put(statType, simType, scale)
			// grid.scales.Put(statType, simType, 10) // 100 range, marginally better, 0.001 marginally worse
		}

		//valueArray := slices.Collect(grid.unitStatValues.ValuesForKey2AsSeq(simType))
		//slices.Sort(valueArray)
		//grid.printer.Printf("[")
		//for i := range valueArray {
		//	v := valueArray[i] * scale
		//	grid.printer.Printf("%f ", v)
		//}
		//grid.printer.Println("]")
	}
}

func (grid *GridStatWeightProcess1B) removeOutliers() {
	if grid.OUTLIER != 0 {
		for _, statType := range grid.requiredStats {
			for _, simType := range grid.simTypes {
				grid.unitStatValues.MapInternalSliceIfExists(statType, simType, func(dataSlice []float64) []float64 {
					total := 0.0
					count := 0
					for _, value := range dataSlice {
						total += value
						count++
					}

					average := total / float64(count)

					totalDiff := 0.0
					for _, value := range dataSlice {
						totalDiff += math.Abs(average - value)
					}
					stdDev := totalDiff / float64(count)

					if grid.OUTLIER == 1 {
						acceptDeviations := 2.0
						util_collection.FilterSliceInPlace(&dataSlice, func(value *float64) bool {
							deviation := math.Abs(average-*value) / stdDev
							return deviation <= acceptDeviations
						})
					} else if grid.OUTLIER == 2 {
						acceptDeviations := 4.0
						util_collection.FilterSliceInPlace(&dataSlice, func(value *float64) bool {
							deviation := math.Abs(average-*value) / stdDev
							return deviation <= acceptDeviations
						})
					} else if grid.OUTLIER == 3 {
						if len(dataSlice) >= 12 {
							slices.Sort(dataSlice)
							dataSlice = dataSlice[len(dataSlice)/6 : len(dataSlice)*5/6]
						}
					}

					return dataSlice
				})
			}
		}
	}
}

func (grid *GridStatWeightProcess1B) chooseScalesEachCombo() {
	for group := range grid.unitStatValues.SeqKey1Key2ValueSeqEntries() {
		if grid.SCALEMODE == 3 {
			scale := util_weight.ChooseScale(group.ValueSeq, c_grid1b_scaleTarget, false)
			grid.scales.Put(group.Key1, group.Key2, scale)
		} else {
			scale := util_weight.ChooseScale(group.ValueSeq, c_grid1b_scaleTarget, true)
			grid.scales.Put(group.Key1, group.Key2, scale)
		}
	}
}

// func (grid *GridStatWeightProcess1B) chooseScalesEachCombo() {
// 	scaleTarget := 1.0
// 	for group := range grid.unitStatValues.SeqGroupsKeysNestedValueSeq() {
// 		minValue, maxValue := math.MaxFloat64, 0.0
// 		total := 0.0
// 		count := 0
// 		for valueRaw := range group.ValueSeq {
// 			value := math.Abs(valueRaw)
// 			minValue = min(minValue, value)
// 			maxValue = max(maxValue, value)
// 			total += value
// 			count++
// 		}

// 		var scale float64
// 		if count != 0 {
// 			scale = scaleTarget / minValue
// 			scale = util.Clamp(scale, 1e-5, 1e5)
// 		} else {
// 			scale = 1
// 		}

// 		grid.scales.Put(group.Key1, group.Key2, scale)

// 		valueArray := slices.Collect(group.ValueSeq)
// 		slices.Sort(valueArray)
// 		grid.printer.Printf("%s %s [", group.Key1.Name(), group.Key2.Name())
// 		for i := range valueArray {
// 			v := valueArray[i] * scale
// 			grid.printer.Printf("%f ", v)
// 		}
// 		grid.printer.Println("]")
// 	}
// }

func (grid *GridStatWeightProcess1B) unitValuesToCalcDetailedRatings() {
	baseStat := grid.requiredStats[0]
	for _, simType := range grid.simTypes {
		unitValueBaseSeq := grid.unitStatValues.GetAsSeq(baseStat, simType)
		for _, thisStatType := range grid.requiredStats {
			if thisStatType != baseStat {
				thisUnitValueSeq := grid.unitStatValues.GetAsSeq(thisStatType, simType)
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
	outputScale := 1 / baseScale

	index := 0
	for baseUnitSample := range baseUnitValueSeq {
		for thisUnitSample := range thisUnitValueSeq {
			if grid.OUTLIER == 4 {
				if !isGoodValueRange(baseUnitSample) || !isGoodValueRange(thisUnitSample) {
					continue
				}
			}

			if grid.CALCMODE == 0 {
				var debugText string = debugText + " " + strconv.Itoa(index)
				offsetAbs := grid.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), 1, util_highs.DebugString{Text: "OFFSET ABS " + debugText})
				grid.build.AbsoluteValueFromDiffTwoVars_ScaleOutput(thisDetailWeightCol, baseUnitSample*thisScale, baseDetailWeightCol, thisUnitSample*baseScale, offsetAbs, outputScale, "OFFSET ABS "+debugText)
				index++
			} else if grid.CALCMODE == 1 {
				var debugText string = debugText + " " + strconv.Itoa(index)
				offsetAbs := grid.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), 1, util_highs.DebugString{Text: "OFFSET ABS " + debugText})
				grid.build.AbsoluteValueFromDiffTwoVars_ScaleOutput(thisDetailWeightCol, thisUnitSample*thisScale, baseDetailWeightCol, baseUnitSample*baseScale, offsetAbs, outputScale, "OFFSET ABS "+debugText)
				index++
			} else if grid.CALCMODE == 2 {
				var debugText string = debugText + " " + strconv.Itoa(index)
				offsetAbs := grid.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), 1, util_highs.DebugString{Text: "OFFSET ABS " + debugText})
				grid.build.AbsoluteValueFromDiffTwoVars_ScaleOutput(thisDetailWeightCol, baseUnitSample*baseScale, baseDetailWeightCol, thisUnitSample*thisScale, offsetAbs, outputScale, "OFFSET ABS "+debugText)
				index++
			} else {
				var debugText string = debugText + " " + strconv.Itoa(index)
				offsetAbs := grid.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), 1, util_highs.DebugString{Text: "OFFSET ABS " + debugText})
				grid.build.AbsoluteValueFromDiffTwoVars_ScaleOutput(thisDetailWeightCol, thisUnitSample*baseScale, baseDetailWeightCol, baseUnitSample*thisScale, offsetAbs, outputScale, "OFFSET ABS "+debugText)
				index++
			}

		}
	}
}

func (grid *GridStatWeightProcess1B) reportOutputWeightsGrid(solution *highs.Solution, weightColumns map[stats.StatType]util_highs.ColumnIndex, printer *util.PrintRecorder) weight_types.Weight1Basic {
	result := weight_types.Weight1Basic_Make()
	printer.Println("FINAL WEIGHTS:")
	for _, statType := range grid.requiredStats {
		columnIndex := weightColumns[statType]
		value := solution.ColValues[columnIndex]
		printer.Printf("%10s %f\n", statType.Name(), value)
		result.Put(statType, value)
	}
	return result
}
