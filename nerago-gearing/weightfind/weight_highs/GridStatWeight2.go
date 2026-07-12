package weight_highs

import (
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_grid2_minBaseWeight = 0.01
	c_grid2_maxWeight     = 10.0

	c_grid2_simDiffTempCalcHighRange = 1000.0
	c_grid2_simDiffScaleMax          = 100.0
	c_grid2_simDiffMinimumUseful     = 0.1
)

type GridStatWeightProcess2 struct {
	printer *util.PrintRecorder

	IncludeDiffs1   bool
	IncludeDiffs2   bool
	MethodX         bool
	MULTIPLY_OUTPUT int
	SIM_HIGH        int

	targetRatios stats.SimData
	statTypes    []stats.StatType
	simTypes     []stats.SimType
	inputData    []WeightInput

	scaleSims map[stats.SimType]float64

	build           util_highs.LinearBuilder
	detailedWeights util.MapMap[stats.StatType, stats.SimType, util_highs.ColumnIndex]
}

func (grid2 *GridStatWeightProcess2) Init(printer *util.PrintRecorder, timeout int) {
	grid2.printer = printer
	grid2.initBuilder(timeout)
}

func (grid2 *GridStatWeightProcess2) initBuilder(timeout int) {
	grid2.build.Minimise = true
	grid2.build.TimeLimitSeconds = timeout
	if grid2.IncludeDiffs2 && !grid2.MethodX {
		grid2.build.Solver = util_highs.Solver_MIP_Interior
	} else {
		grid2.build.Solver = util_highs.Solver_LP_USE_GPU
	}
}

func (grid2 *GridStatWeightProcess2) SupplyData(inputData []WeightInput) {
	grid2.inputData = inputData
}

func (grid2 *GridStatWeightProcess2) SetRequiredStats(requiredStats []stats.StatType) {
	grid2.statTypes = requiredStats
}

func (grid2 *GridStatWeightProcess2) SetTargetRatios(targetRatios stats.SimData) {
	grid2.simTypes = targetRatios.NonZeroTypes()

	sum := 0.0
	for _, simType := range grid2.simTypes {
		val := targetRatios.Get(simType)
		if val <= 0 {
			panic("missing ratio")
		}
		sum += val
	}
	if !util.FloatEqualsOne(sum) {
		panic("ratios don't add to one")
	}

	grid2.targetRatios = targetRatios
}

func (grid2 *GridStatWeightProcess2) Run(stopwatch *util.Stopwatch) *util_async.FutureCancellable[WeightResult] {
	grid2.setupWeightVars()
	grid2.chooseSimDiffScaling()
	//grid2.processInputData()
	if grid2.IncludeDiffs1 {
		grid2.processInputDataOnes()
	}
	if grid2.IncludeDiffs2 {
		grid2.processInputDataTwos()
	}

	solutionFuture := grid2.build.RunHighsFuture(stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) (WeightResult, bool) {
		solution := linearResult.GetSolutionAndSaveLog(grid2.printer)

		grid2.printer.Println(solution.Status.String())
		grid2.build.DebugPrintColumns(solution, grid2.printer)

		return grid2.reportOutputWeightsGrid(solution), true
	})
}

func (grid2 *GridStatWeightProcess2) RunOneTwo(stopwatch *util.Stopwatch) *util_async.FutureCancellable[WeightResult] {
	grid2.setupWeightVars()
	grid2.chooseSimDiffScaling()
	grid2.processInputDataOnes()

	grid2.build.Solver = util_highs.Solver_LP_USE_GPU
	solutionFuture := grid2.build.RunHighsFuture(stopwatch)

	futureNext := util_async.FutureCancellable_MapToFuture(solutionFuture, func(linearResult util_highs.LinearResult) *util_async.FutureCancellable[util_highs.LinearResult] {
		solution := linearResult.GetSolutionAndSaveLog(grid2.printer)
		grid2.build.DebugPrintColumns(solution, grid2.printer)

		if solution.Status == highs.ModelStatusOptimal {
			grid2.processInputDataTwos()
			grid2.copySolutionAsNextInitial(solution)
			if !grid2.MethodX {
				grid2.build.Solver = util_highs.Solver_MIP_Interior
			}
			return grid2.build.RunHighsFuture(stopwatch)
		} else {
			return nil
		}
	})

	return util_async.FutureCancellable_MapValue(futureNext, func(linearResult util_highs.LinearResult) (WeightResult, bool) {
		solution := linearResult.GetSolutionAndSaveLog(grid2.printer)
		grid2.build.DebugPrintColumns(solution, grid2.printer)

		if solution.HasSolution() {
			return grid2.reportOutputWeightsGrid(solution), true
		} else {
			return WeightResult{}, false
		}
	})
}

func (grid2 *GridStatWeightProcess2) setupWeightVars() {
	baseStat := grid2.statTypes[0]

	// create detail columns
	for _, statType := range grid2.statTypes {
		for _, simType := range grid2.simTypes {
			minWeight := -c_grid2_maxWeight
			maxWeight := c_grid2_maxWeight
			if statType == baseStat {
				minWeight = c_grid2_minBaseWeight
			}
			colDetailWeight := grid2.build.CreateColumnGeneral(highs.Continuous, minWeight, maxWeight, util_highs.DebugString{Text: "WEIGHT: " + statType.Name() + " " + simType.Name()})
			grid2.detailedWeights.Put(statType, simType, colDetailWeight)
		}
	}
}

func (grid2 *GridStatWeightProcess2) chooseSimDiffScaling() {
	grid2.scaleSims = make(map[stats.SimType]float64)
	for _, simType := range grid2.simTypes {
		listDiffs := make([]float64, 0)
		for a := range grid2.inputData {
			for b := a + 1; b < len(grid2.inputData); b++ {
				one, two := &grid2.inputData[a], &grid2.inputData[b]
				simDiff, goodDiff := grid2.calcSimDiff(one, two, simType)
				if goodDiff {
					listDiffs = append(listDiffs, math.Abs(simDiff))
				}
			}
		}

		scale := chooseScale(slices.Values(listDiffs), c_grid2_simDiffScaleMax, true)
		grid2.scaleSims[simType] = scale
	}
}

func (grid2 *GridStatWeightProcess2) calcSimDiff(one *WeightInput, two *WeightInput, simType stats.SimType) (float64, bool) {
	simOne := one.SimResult.GetFriendly(simType)
	simTwo := two.SimResult.GetFriendly(simType)
	diff := simOne - simTwo
	if grid2.SIM_HIGH == 3 {
		if simType.IsHighGood() {
			diff *= -1
		}
	} else if grid2.SIM_HIGH == 4 {
		if simType.IsHighGood() {
			diff *= -1
		}
	}
	if math.Abs(diff) >= c_grid2_simDiffMinimumUseful {
		return diff, true
	} else {
		return 0, false
	}
}

func (grid2 *GridStatWeightProcess2) processInputDataOnes() {
	for a := range grid2.inputData {
		for b := a + 1; b < len(grid2.inputData); b++ {
			differenceCount, diffStatA, _, _ := grid2.checkForNumberStatDifferences(&grid2.inputData[a].TotalStat, &grid2.inputData[b].TotalStat)
			switch differenceCount {
			case 1:
				if grid2.MethodX {
					grid2.prepareSampleOneDifferenceStatsX(&grid2.inputData[a], &grid2.inputData[b], diffStatA)
				} else {
					grid2.prepareSampleOneDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA)
				}
			}
		}
	}
}

func (grid2 *GridStatWeightProcess2) processInputDataTwos() {
	for a := range grid2.inputData {
		for b := a + 1; b < len(grid2.inputData); b++ {
			differenceCount, diffStatA, diffStatB, _ := grid2.checkForNumberStatDifferences(&grid2.inputData[a].TotalStat, &grid2.inputData[b].TotalStat)
			switch differenceCount {
			case 2:
				if grid2.MethodX {
					grid2.prepareSampleTwoDifferenceStatsMIPFree(&grid2.inputData[a], &grid2.inputData[b], diffStatA, diffStatB)
				} else {
					grid2.prepareSampleTwoDifferenceStatsMIPNeeded(&grid2.inputData[a], &grid2.inputData[b], diffStatA, diffStatB)
				}
			}
		}
	}
}

//
//func (grid2 *GridStatWeightProcess2) processInputData() {
//	for a := range grid2.inputData {
//		for b := a + 1; b < len(grid2.inputData); b++ {
//			differenceCount, diffStatA, diffStatB, _ := grid2.checkForNumberStatDifferences(&grid2.inputData[a].TotalStat, &grid2.inputData[b].TotalStat)
//			switch differenceCount {
//			case 1:
//				if grid2.IncludeDiffs1 {
//					if grid2.MethodX {
//						grid2.prepareSampleOneDifferenceStatsX(&grid2.inputData[a], &grid2.inputData[b], diffStatA)
//					} else {
//						grid2.prepareSampleOneDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA)
//					}
//				}
//
//			case 2:
//				if grid2.IncludeDiffs2 {
//					if grid2.MethodX {
//						grid2.prepareSampleTwoDifferenceStatsMIPFree(&grid2.inputData[a], &grid2.inputData[b], diffStatA, diffStatB)
//					} else {
//						grid2.prepareSampleTwoDifferenceStatsMIPNeeded(&grid2.inputData[a], &grid2.inputData[b], diffStatA, diffStatB)
//					}
//				}
//
//			case 3:
//				// grid2.prepareSampleThreeDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA, diffStatB, diffStatC)
//			}
//		}
//	}
//}

func (grid2 *GridStatWeightProcess2) checkForNumberStatDifferences(one, two *stats.StatBlock) (differenceCount int, diffStatA stats.StatType, diffStatB stats.StatType, diffStatC stats.StatType) {
	// for stat := range one { // was doing fine in tests up until now, up to 89%
	for _, stat := range grid2.statTypes {
		if one[stat] != two[stat] {
			switch differenceCount {
			case 0:
				diffStatA = stat
			case 1:
				diffStatB = stat
			case 2:
				diffStatC = stat
			}
			differenceCount++
		}
	}
	return differenceCount, diffStatA, diffStatB, diffStatC
}

func (grid2 *GridStatWeightProcess2) prepareSampleOneDifferenceStats(one *WeightInput, two *WeightInput, statType stats.StatType) {
	statDiff := one.TotalStat.GetFloat(statType) - two.TotalStat.GetFloat(statType)

	for _, simType := range grid2.simTypes {
		weightColumn := grid2.detailedWeights.GetOrPanic(statType, simType)

		debugText := "MISMATCH1 " + statType.Name() + " " + simType.Name()
		mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, 1, util_highs.DebugString{Text: debugText})

		simDiff, goodDiff := grid2.calcSimDiff(two, one, simType) // opposite, matches order in MIPNeeded two
		if goodDiff {
			simDiff *= grid2.scaleSims[simType]
			grid2.build.AbsoluteValueFromDiffOneToConst(weightColumn, statDiff, simDiff, mismatchCol, debugText)
		}
	}
}
func (grid2 *GridStatWeightProcess2) prepareSampleOneDifferenceStatsX(one *WeightInput, two *WeightInput, statType stats.StatType) {
	statDiff := one.TotalStat.GetFloat(statType) - two.TotalStat.GetFloat(statType)

	for _, simType := range grid2.simTypes {
		weightColumn := grid2.detailedWeights.GetOrPanic(statType, simType)

		debugText := "MISMATCH1 " + statType.Name() + " " + simType.Name()
		mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, 1, util_highs.DebugString{Text: debugText})

		simDiff, goodDiff := grid2.calcSimDiff(one, two, simType) // opposite
		if goodDiff {
			simDiff *= grid2.scaleSims[simType]
			grid2.build.AbsoluteValueFromDiffOneToConst(weightColumn, statDiff, simDiff, mismatchCol, debugText)
		}
	}
}

func (grid2 *GridStatWeightProcess2) prepareSampleTwoDifferenceStatsMIPNeeded(one *WeightInput, two *WeightInput, statTypeA stats.StatType, statTypeB stats.StatType) {
	statDiffA := one.TotalStat.GetFloat(statTypeA) - two.TotalStat.GetFloat(statTypeA)
	statDiffB := one.TotalStat.GetFloat(statTypeB) - two.TotalStat.GetFloat(statTypeB)

	for _, simType := range grid2.simTypes {
		weightColumnA := grid2.detailedWeights.GetOrPanic(statTypeA, simType)
		weightColumnB := grid2.detailedWeights.GetOrPanic(statTypeB, simType)

		debugText := "MISMATCH2 " + statTypeA.Name() + " " + statTypeB.Name() + " " + simType.Name()
		mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, 1, util_highs.DebugString{Text: debugText})

		simDiff, goodDiff := grid2.calcSimDiff(two, one, simType)
		if goodDiff {
			simDiff *= grid2.scaleSims[simType]
			grid2.build.AbsoluteValueDiffTwoVarsThenDiffConst_NeedMIP(
				weightColumnA, statDiffA,
				weightColumnB, statDiffB,
				mismatchCol,
				simDiff,
				c_grid2_simDiffTempCalcHighRange,
				debugText)
		}
	}
}

func (grid2 *GridStatWeightProcess2) prepareSampleTwoDifferenceStatsMIPFree(one *WeightInput, two *WeightInput, statTypeA stats.StatType, statTypeB stats.StatType) {
	statDiffA := one.TotalStat.GetFloat(statTypeA) - two.TotalStat.GetFloat(statTypeA)
	statDiffB := one.TotalStat.GetFloat(statTypeB) - two.TotalStat.GetFloat(statTypeB)

	for _, simType := range grid2.simTypes {
		weightColumnA := grid2.detailedWeights.GetOrPanic(statTypeA, simType)
		weightColumnB := grid2.detailedWeights.GetOrPanic(statTypeB, simType)

		debugText := "MISMATCH2 " + statTypeA.Name() + " " + statTypeB.Name() + " " + simType.Name()
		mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, 1, util_highs.DebugString{Text: debugText})

		simDiff, goodDiff := grid2.calcSimDiff(one, two, simType)
		if goodDiff {
			simDiff *= grid2.scaleSims[simType]
			grid2.build.AbsoluteValueFromSumTwoThenDiffToConst(
				weightColumnA, statDiffA,
				weightColumnB, statDiffB,
				simDiff,
				mismatchCol,
				debugText)
		}
	}
}

func (grid2 *GridStatWeightProcess2) prepareSampleThreeDifferenceStatsX(one *WeightInput, two *WeightInput, statTypeA stats.StatType, statTypeB stats.StatType, statTypeC stats.StatType) {
	statDiffA := one.TotalStat.GetFloat(statTypeA) - two.TotalStat.GetFloat(statTypeA)
	statDiffB := one.TotalStat.GetFloat(statTypeB) - two.TotalStat.GetFloat(statTypeB)
	statDiffC := one.TotalStat.GetFloat(statTypeC) - two.TotalStat.GetFloat(statTypeC)

	for _, simType := range grid2.simTypes {
		weightColumnA := grid2.detailedWeights.GetOrPanic(statTypeA, simType)
		weightColumnB := grid2.detailedWeights.GetOrPanic(statTypeB, simType)
		weightColumnC := grid2.detailedWeights.GetOrPanic(statTypeC, simType)

		simDiff, goodDiff := grid2.calcSimDiff(one, two, simType)
		if goodDiff {
			simDiff *= grid2.scaleSims[simType]

			debugText := "MISMATCH3 " + statTypeA.Name() + " " + statTypeB.Name() + " " + statTypeC.Name() + " " + simType.Name()
			mismatchSignedCol := grid2.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, nil)

			rowEquate := util_highs.ConstraintRow{}
			rowEquate.Add(weightColumnA, statDiffA)
			rowEquate.Add(weightColumnB, statDiffB)
			rowEquate.Add(weightColumnC, statDiffC)
			rowEquate.Add(mismatchSignedCol, 1)
			rowEquate.Build(&grid2.build, simDiff, simDiff)

			mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, 1, util_highs.DebugString{Text: debugText})
			grid2.build.AbsoluteValue(mismatchSignedCol, mismatchCol)
		}
	}
}

func (grid2 *GridStatWeightProcess2) reportOutputWeightsGrid(solution *highs.Solution) WeightResult {
	result := WeightResult_Make()
	grid2.printer.Println("FINAL WEIGHTS:")

	// weight * (statOne - statTwo) * statScale[type] = (simOne - simTwo) * simScale[type]
	// weight = (simOne - simTwo) / (statOne - statTwo) * (simScale[type] / statScale[type])
	// usableweight = (simOne - simTwo) / (statOne - statTwo)
	// weight = usableweight * (simScale[type] / statScale[type])
	// usableweight = weight * (statScale[type] / simScale[type])

	for _, statType := range grid2.statTypes {
		grid2.printer.Printf("%10s >>>>>\n", statType.Name())

		sumIndividual := 0.0

		for simType, detailWeightCol := range grid2.detailedWeights.SeqInnerWithKey1Value(statType) {
			weight := solution.ColValues[detailWeightCol]

			var usableWeight float64
			switch grid2.MULTIPLY_OUTPUT {
			case 0:
				usableWeight = weight
			case 1:
				usableWeight = weight / grid2.scaleSims[simType]
			case 2:
				usableWeight = weight * grid2.scaleSims[simType]
			}
			// TODO we've generally found that multiplying by scaleStat is correct, but this is scaleSim, check

			if grid2.SIM_HIGH == 1 {
				if simType.IsHighGood() {
					usableWeight *= -1
				}
			} else if grid2.SIM_HIGH == 2 {
				if !simType.IsHighGood() {
					usableWeight *= -1
				}
			}

			grid2.printer.Printf("         %5s > %f %f\n", simType.Name(), weight, usableWeight)

			sumIndividual += usableWeight * grid2.targetRatios.Get(simType)
		}

		grid2.printer.Printf("             === %f\n", sumIndividual)
		result.Put(statType, sumIndividual)
	}

	baseStat := grid2.statTypes[0]
	divideBy := result.Get(baseStat)
	for _, statType := range grid2.statTypes {
		result.Put(statType, result.Get(statType)/divideBy)
	}

	return result
}

func (grid2 *GridStatWeightProcess2) copySolutionAsNextInitial(solution *highs.Solution) {
	for columnIndex, value := range solution.ColValues {
		grid2.build.SetInitialSolutionValue(util_highs.ColumnIndex(columnIndex), value)
	}
}
