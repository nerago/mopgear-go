package weight_highs

import (
	"iter"
	"math"
	"slices"
	"strconv"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const c_grid1c_scaleTarget = 10.0

type GridStatWeightProcess1C struct {
	printer *util.PrintRecorder

	targetRatios  weight_types.SimPriorityBasic
	inputData     []weight_types.WeightInput
	requiredStats []stats.StatType
	simTypes      []stats.SimType
	testMode      bool
	ROUNDMODE     int
	RESCALE       int

	build           util_highs.LinearBuilder
	unitStatValues  util_collection.MapMapSlice[stats.StatType, stats.SimType, float64]
	scales          map[stats.SimType]float64
	detailedWeights util_collection.MapMap[stats.StatType, stats.SimType, util_highs.ColumnIndex]
	finalWeights    map[stats.StatType]util_highs.ColumnIndex
}

func (grid *GridStatWeightProcess1C) Init(printer *util.PrintRecorder, timeout int) {
	grid.printer = printer
	grid.build.Minimise = true
	grid.build.Solver = util_highs.Solver_LP_USE_GPU
	grid.build.TimeLimitSeconds = timeout
	grid.build.SetEachTolerance(1e-3)
	grid.finalWeights = make(map[stats.StatType]util_highs.ColumnIndex)
	grid.scales = make(map[stats.SimType]float64)
}

func (grid *GridStatWeightProcess1C) SupplyData(inputData []weight_types.WeightInput) {
	if grid.testMode {
		grid.inputData = inputData[0:10]
	} else {
		grid.inputData = inputData
	}
}

func (grid *GridStatWeightProcess1C) SetRequiredStats(requiredStats []stats.StatType) {
	grid.requiredStats = requiredStats
}

func (grid *GridStatWeightProcess1C) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	targetRatios.ValidateRatioAddsToOne()
	grid.simTypes = targetRatios.SimTypes()
	grid.targetRatios = targetRatios
}

func (grid *GridStatWeightProcess1C) SetTestMode(testMode bool) {
	grid.testMode = testMode
	if testMode {
		grid.build.TimeLimitSeconds = 60
	}
}

func (grid *GridStatWeightProcess1C) Run() (*util_async.FutureCancellable[weight_types.WeightResult1], error) {
	grid.setupWeightVars()
	grid.dataSamplesFromPairs()
	grid.removeOutliers()
	grid.chooseScalesBySim()
	grid.unitValuesToCalcDetailedRatings()
	grid.finalWeightVars()

	stopwatch := util.StopwatchMakeStopped()
	solutionFuture := grid.build.RunHighsFuture(stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) weight_types.WeightResult1 {
		if solution, err := linearResult.GetSolutionAndSaveLog(grid.printer); err == nil {
			grid.printer.Println(solution.Status.String())
			grid.build.DebugPrintColumns(solution, grid.printer)

			weight := grid.reportOutputWeightsGrid(solution, grid.finalWeights, grid.printer)
			return weight_types.WeightResult1Make(&weight, stopwatch.Elapsed(), solution.Status)
		} else {
			return weight_types.WeightResult1MakeError(stopwatch.Elapsed(), err)
		}
	})
}

func (grid *GridStatWeightProcess1C) setupWeightVars() {
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

	// TODO assumes base stat is positive
	baseStat := grid.requiredStats[0]
	for _, simType := range grid.simTypes {
		value := grid.targetRatios.GetOrPanic(simType)
		colDetailWeight := grid.detailedWeights.GetOrPanic(baseStat, simType)
		strengthSetToRatio := util_highs.ConstraintRow{}
		strengthSetToRatio.Add(colDetailWeight, 1)
		strengthSetToRatio.Build(&grid.build, value, value)
	}
}

func (grid *GridStatWeightProcess1C) finalWeightVars() {
	for _, statType := range grid.requiredStats {
		statFinalRow := util_highs.ConstraintRow{}
		for simType, detailColumn := range grid.detailedWeights.SeqKey2ValueWithKey1(statType) {
			scale := grid.scales[simType]
			var coeff float64
			if grid.RESCALE == 0 {
				coeff = 1
			} else if grid.RESCALE == 1 {
				coeff = scale
			} else if grid.RESCALE == 2 {
				coeff = 1 / scale
			}
			statFinalRow.Add(detailColumn, coeff) // 1/scale[100] is too small, =scale[100] is too big. but all harmless when fixed like that
		}

		finalWeightColumn := grid.finalWeights[statType]
		statFinalRow.Add(finalWeightColumn, -1)
		statFinalRow.Build(&grid.build, 0, 0)
	}
}

func (grid *GridStatWeightProcess1C) dataSamplesFromPairs() {
	for a := range grid.inputData {
		for b := a + 1; b < len(grid.inputData); b++ {
			statType, isGood := grid.hasOneStatDifferent(&grid.inputData[a].TotalStat, &grid.inputData[b].TotalStat)
			if isGood {
				grid.prepareSample(statType, &grid.inputData[a], &grid.inputData[b])
			}
		}
	}
}

func (grid *GridStatWeightProcess1C) hasOneStatDifferent(blockHigh, blockLow *stats.StatBlock) (stats.StatType, bool) {
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

func (grid *GridStatWeightProcess1C) prepareSample(statType stats.StatType, high, low *weight_types.WeightInput) {
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

func (grid *GridStatWeightProcess1C) chooseScalesBySim() {
	for _, simType := range grid.simTypes {
		scale := util_weight.ChooseScale(grid.unitStatValues.SeqValuesWithKey2(simType), c_grid1c_scaleTarget, false)
		grid.scales[simType] = scale
	}
}

func (grid *GridStatWeightProcess1C) removeOutliers() {
	for _, statType := range grid.requiredStats {
		for _, simType := range grid.simTypes {
			grid.unitStatValues.MapInternalSliceIfExists(statType, simType, func(dataSlice []float64) []float64 {
				if len(dataSlice) >= 12 {
					slices.Sort(dataSlice)
					dataSlice = dataSlice[len(dataSlice)/6 : len(dataSlice)*5/6]
				}
				return dataSlice
			})
		}
	}
}

func (grid *GridStatWeightProcess1C) unitValuesToCalcDetailedRatings() {
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

func (grid *GridStatWeightProcess1C) unitValuesCalcForGroup(simType stats.SimType, thisStatType stats.StatType, baseUnitValueSeq iter.Seq[float64], thisUnitValueSeq iter.Seq[float64]) {
	debugText := simType.Name() + " " + thisStatType.Name()
	baseStat := grid.requiredStats[0]
	baseDetailWeightCol := grid.detailedWeights.GetOrPanic(baseStat, simType)
	thisDetailWeightCol := grid.detailedWeights.GetOrPanic(thisStatType, simType)

	scale := grid.scales[simType]

	index := 0
	for baseUnitSample := range baseUnitValueSeq {
		for thisUnitSample := range thisUnitValueSeq {
			debugText := debugText + " " + strconv.Itoa(index)
			offsetAbs := grid.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), 1, util_highs.DebugString{Text: "OFFSET ABS " + debugText})

			grid.build.AbsoluteValueFromDiffTwoVars_ScaleOutput(thisDetailWeightCol, baseUnitSample*scale, baseDetailWeightCol, thisUnitSample*scale, offsetAbs, 1/scale, "OFFSET ABS "+debugText)

			index++
		}
	}
}

func (grid *GridStatWeightProcess1C) reportOutputWeightsGrid(solution *highs.Solution, weightColumns map[stats.StatType]util_highs.ColumnIndex, printer *util.PrintRecorder) weight_types.Weight1Basic {
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
