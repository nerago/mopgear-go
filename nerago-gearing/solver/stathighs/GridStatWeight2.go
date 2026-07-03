package stathighs

import (
	"math"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

const c_grid2_maxWeight = 5000.0
const c_grid2_highScore = 10000.0
const c_grid2_scaleTarget = 10.0

type GridStatWeightProcess2 struct {
	printer *util.PrintRecorder

	DIFFINCLUDE   int
	targetRatios  stats.SimData
	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	inputData     []WeightInput

	scaleSims map[stats.SimType]float64
	//scaleStats map[stats.StatType]float64

	build           utilhighs.LinearBuilder
	detailedWeights util.MapMap[stats.StatType, stats.SimType, utilhighs.ColumnIndex]
}

type gridDataSample2 struct {
	value float64
}

func (grid2 *GridStatWeightProcess2) Init(printer *util.PrintRecorder, timeout int) {
	grid2.printer = printer
	grid2.build.Minimise = true
	grid2.build.TimeLimitSeconds = timeout
	if grid2.DIFFINCLUDE == 2 || grid2.DIFFINCLUDE == 12 {
		grid2.build.Solver = utilhighs.Solver_MIP_Interior
	} else {
		grid2.build.Solver = utilhighs.Solver_LP_USE_GPU
		grid2.build.DisablePreSolve = true
	}
}

func (grid2 *GridStatWeightProcess2) SupplyData(inputData []WeightInput) {
	grid2.inputData = inputData
}

func (grid2 *GridStatWeightProcess2) SetRequiredStats(requiredStats []stats.StatType) {
	grid2.requiredStats = requiredStats
}

func (grid2 *GridStatWeightProcess2) SetTargetRatios(targetRatios stats.SimData) {
	grid2.requiredSims = targetRatios.NonZeroTypes()

	sum := 0.0
	for _, simType := range grid2.requiredSims {
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

func (grid2 *GridStatWeightProcess2) Run(stopwatch *util.Stopwatch) *channel_op.FutureCancellable[WeightResult] {
	grid2.setupWeightVars()
	grid2.chooseScalingX()
	grid2.processInputData()

	solutionFuture := grid2.build.RunHighsFuture(stopwatch)
	return channel_op.FutureCancellable_MapValue(solutionFuture, func(linearResult utilhighs.LinearResult) (WeightResult, bool) {
		solution := linearResult.GetSolutionAndSaveLog(grid2.printer)

		grid2.printer.Println(solution.Status.String())
		grid2.build.DebugPrintColumns(solution, grid2.printer)

		return grid2.reportOutputWeightsGrid(solution), true
	})

}

func (grid2 *GridStatWeightProcess2) setupWeightVars() {
	// create detail columns
	for _, statType := range grid2.requiredStats {
		for _, simType := range grid2.requiredSims {
			colDetailWeight := grid2.build.CreateColumnGeneral(highs.Continuous, -c_grid2_maxWeight, c_grid2_maxWeight, utilhighs.DebugString{Text: "WEIGHT: " + statType.Name() + " " + simType.Name()})
			grid2.detailedWeights.Put(statType, simType, colDetailWeight)
		}
	}

	// strength column within each simtype is set to targetratio (0.4 etc)
	baseStat := grid2.requiredStats[0]
	for _, simType := range grid2.requiredSims {
		colDetailWeight := grid2.detailedWeights.GetOrPanic(baseStat, simType)
		strAbs := grid2.build.CreateColumnGeneral(highs.Continuous, 0.001, utilhighs.C_PlusInf, nil)
		grid2.build.AbsoluteValueFromDiffOneToConst(colDetailWeight, 1, 0, strAbs, "")
	}
}

func (grid2 *GridStatWeightProcess2) chooseScalingX() {
	grid2.scaleSims = make(map[stats.SimType]float64)
	for _, simType := range grid2.requiredSims {
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
		grid2.printer.Printf("simDiffs %s min=%f avg=%f max=%f\n", simType.Name(), listMin(listDiffs), listAvg(listDiffs), listMax(listDiffs))

		scale := chooseScale(slices.Values(listDiffs), c_grid2_scaleTarget)
		grid2.scaleSims[simType] = scale
	}
}

func (grid2 *GridStatWeightProcess2) calcSimDiff(one *WeightInput, two *WeightInput, simType stats.SimType) (float64, bool) {
	simOne := one.SimResult.GetFriendly(simType)
	simTwo := two.SimResult.GetFriendly(simType)
	if !util.FloatsApproxEquals(simOne, simTwo) {
		return simOne - simTwo, true
	} else {
		return 0, false
	}
}

func listMin(list []float64) float64 {
	curr := list[0]
	for _, val := range list {
		curr = min(curr, val)
	}
	return curr
}
func listMax(list []float64) float64 {
	curr := list[0]
	for _, val := range list {
		curr = max(curr, val)
	}
	return curr
}
func listAvg(list []float64) float64 {
	total := 0.0
	for _, val := range list {
		total += val
	}
	return total / float64(len(list))
}

func (grid2 *GridStatWeightProcess2) processInputData() {
	for a := range grid2.inputData {
		for b := a + 1; b < len(grid2.inputData); b++ {
			differenceCount, diffStatA, diffStatB, _ := grid2.checkForNumberStatDifferences(&grid2.inputData[a].TotalStat, &grid2.inputData[b].TotalStat)
			switch differenceCount {
			case 1:
				if grid2.DIFFINCLUDE == 1 || grid2.DIFFINCLUDE == 12 {
					grid2.prepareSampleOneDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA)
				}

				if grid2.DIFFINCLUDE == 1001 || grid2.DIFFINCLUDE == 1012 {
					grid2.prepareSampleOneDifferenceStatsX(&grid2.inputData[a], &grid2.inputData[b], diffStatA)
				}
			case 2:
				if grid2.DIFFINCLUDE == 2 || grid2.DIFFINCLUDE == 12 {
					grid2.prepareSampleTwoDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA, diffStatB)
				}

				if grid2.DIFFINCLUDE == 1002 || grid2.DIFFINCLUDE == 1012 {
					grid2.prepareSampleTwoDifferenceStatsX(&grid2.inputData[a], &grid2.inputData[b], diffStatA, diffStatB)
				}
			case 3:
				// grid2.prepareSampleThreeDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA, diffStatB, diffStatC)
			}

			// TODO param for this, use for compare accuracy etc

			// using 1,2,3: 68.5225%
			// using 1,2: 89.3560%
			// using 2: 89.4443%
		}
	}
}

func (grid2 *GridStatWeightProcess2) checkForNumberStatDifferences(one, two *stats.StatBlock) (differenceCount int, diffStatA stats.StatType, diffStatB stats.StatType, diffStatC stats.StatType) {
	// for stat := range one { // was doing fine in tests up until now, up to 89%
	for _, stat := range grid2.requiredStats {
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

	for _, simType := range grid2.requiredSims {
		weightColumn := grid2.detailedWeights.GetOrPanic(statType, simType)

		debugText := "MISMATCH1 " + statType.Name() + " " + simType.Name()
		mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: debugText})

		simDiff, goodDiff := grid2.calcSimDiff(two, one, simType)
		if goodDiff {
			simDiff *= grid2.scaleSims[simType]
			grid2.build.AbsoluteValueFromDiffOneToConst(weightColumn, statDiff, simDiff, mismatchCol, debugText)
		}
	}
}
func (grid2 *GridStatWeightProcess2) prepareSampleOneDifferenceStatsX(one *WeightInput, two *WeightInput, statType stats.StatType) {
	statDiff := one.TotalStat.GetFloat(statType) - two.TotalStat.GetFloat(statType)

	for _, simType := range grid2.requiredSims {
		weightColumn := grid2.detailedWeights.GetOrPanic(statType, simType)

		debugText := "MISMATCH1 " + statType.Name() + " " + simType.Name()
		mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: debugText})

		simDiff, goodDiff := grid2.calcSimDiff(one, two, simType)
		if goodDiff {
			simDiff *= grid2.scaleSims[simType]
			grid2.build.AbsoluteValueFromDiffOneToConst(weightColumn, statDiff, simDiff, mismatchCol, debugText)
		}
	}
}

func (grid2 *GridStatWeightProcess2) prepareSampleTwoDifferenceStats(one *WeightInput, two *WeightInput, statTypeA stats.StatType, statTypeB stats.StatType) {
	statDiffA := one.TotalStat.GetFloat(statTypeA) - two.TotalStat.GetFloat(statTypeA)
	statDiffB := one.TotalStat.GetFloat(statTypeB) - two.TotalStat.GetFloat(statTypeB)

	for _, simType := range grid2.requiredSims {
		weightColumnA := grid2.detailedWeights.GetOrPanic(statTypeA, simType)
		weightColumnB := grid2.detailedWeights.GetOrPanic(statTypeB, simType)

		debugText := "MISMATCH2 " + statTypeA.Name() + " " + statTypeB.Name() + " " + simType.Name()
		mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: debugText})

		simDiff, goodDiff := grid2.calcSimDiff(two, one, simType)
		if goodDiff {
			simDiff *= grid2.scaleSims[simType]
			grid2.build.AbsoluteValueDiffTwoVarsThenDiffConst(
				weightColumnA, statDiffA,
				weightColumnB, statDiffB,
				mismatchCol,
				simDiff,
				c_grid2_highScore,
				debugText)
		}
	}
}

func (grid2 *GridStatWeightProcess2) prepareSampleTwoDifferenceStatsX(one *WeightInput, two *WeightInput, statTypeA stats.StatType, statTypeB stats.StatType) {
	statDiffA := one.TotalStat.GetFloat(statTypeA) - two.TotalStat.GetFloat(statTypeA)
	statDiffB := one.TotalStat.GetFloat(statTypeB) - two.TotalStat.GetFloat(statTypeB)

	for _, simType := range grid2.requiredSims {
		weightColumnA := grid2.detailedWeights.GetOrPanic(statTypeA, simType)
		weightColumnB := grid2.detailedWeights.GetOrPanic(statTypeB, simType)

		debugText := "MISMATCH2 " + statTypeA.Name() + " " + statTypeB.Name() + " " + simType.Name()
		mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: debugText})

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

	for _, simType := range grid2.requiredSims {
		weightColumnA := grid2.detailedWeights.GetOrPanic(statTypeA, simType)
		weightColumnB := grid2.detailedWeights.GetOrPanic(statTypeB, simType)
		weightColumnC := grid2.detailedWeights.GetOrPanic(statTypeC, simType)

		simDiff, goodDiff := grid2.calcSimDiff(one, two, simType)
		if goodDiff {
			simDiff *= grid2.scaleSims[simType]

			debugText := "MISMATCH3 " + statTypeA.Name() + " " + statTypeB.Name() + " " + statTypeC.Name() + " " + simType.Name()
			mismatchSignedCol := grid2.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, nil)

			rowEquate := utilhighs.ConstraintRow{}
			rowEquate.Add(weightColumnA, statDiffA)
			rowEquate.Add(weightColumnB, statDiffB)
			rowEquate.Add(weightColumnC, statDiffC)
			rowEquate.Add(mismatchSignedCol, 1)
			rowEquate.Build(&grid2.build, simDiff, simDiff)

			mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: debugText})
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

	for _, statType := range grid2.requiredStats {
		grid2.printer.Printf("%10s >>>>>\n", statType.Name())

		sumIndividual := 0.0

		for simType, detailWeightCol := range grid2.detailedWeights.SeqInnerWithKey1Value(statType) {
			weight := solution.ColValues[detailWeightCol]

			// usableWeight := weight * grid2.scaleStats[statType] / grid2.scaleSims[simType]
			// TODO we've generally found that multiplying by scaleStat is correct, but this is scaleSim, check
			usableWeight := weight * grid2.scaleSims[simType]

			// if simType.IsHighGood() {
			// 	usableWeight *= -1
			// }

			grid2.printer.Printf("         %5s > %f %f\n", simType.Name(), weight, usableWeight)

			sumIndividual += usableWeight * grid2.targetRatios.Get(simType)
		}

		grid2.printer.Printf("             === %f\n", sumIndividual)
		result.Put(statType, sumIndividual)
	}

	baseStat := grid2.requiredStats[0]
	divideBy := result.Get(baseStat)
	for _, statType := range grid2.requiredStats {
		result.Put(statType, result.Get(statType)/divideBy)
	}

	return result
}
