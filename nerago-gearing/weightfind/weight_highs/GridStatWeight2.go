package weight_highs

import (
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_grid2_minBaseWeight = 0.01
	c_grid2_maxWeight     = 100.0

	c_grid2_simDiffTempCalcHighRange = 1000.0
	c_grid2_simDiffScaleMax          = 100.0
	c_grid2_simDiffMinimumUseful     = 0.1
)

type GridStatWeightProcess2 struct {
	printer *util.PrintRecorder

	IncludeDiffs1 bool
	IncludeDiffs2 bool
	IncludeDiffs3 bool

	targetRatios weight_types.SimPriorityBasic
	statTypes    []stats.StatType
	simTypes     []stats.SimType
	inputData    []weight_types.WeightInput

	scaleSims map[stats.SimType]float64

	build           util_highs.LinearBuilder
	detailedWeights util_collection.MapMap[stats.StatType, stats.SimType, util_highs.ColumnIndex]
}

func (grid2 *GridStatWeightProcess2) Init(printer *util.PrintRecorder, timeout int) {
	grid2.printer = printer
	grid2.initBuilder(timeout)
}

func (grid2 *GridStatWeightProcess2) initBuilder(timeout int) {
	grid2.build.Minimise = true
	grid2.build.TimeLimitSeconds = timeout
	grid2.build.Solver = util_highs.Solver_Force_IPX
}

func (grid2 *GridStatWeightProcess2) SupplyData(inputData []weight_types.WeightInput) {
	grid2.inputData = inputData
}

func (grid2 *GridStatWeightProcess2) SetRequiredStats(requiredStats []stats.StatType) {
	grid2.statTypes = requiredStats
}

func (grid2 *GridStatWeightProcess2) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {

	grid2.simTypes = targetRatios.SimTypes()
	grid2.targetRatios = targetRatios
}

func (grid2 *GridStatWeightProcess2) Run(stopwatch *util.Stopwatch) *util_async.FutureCancellable[weight_types.Weight2Extended] {
	grid2.setupWeightVars()
	grid2.chooseSimDiffScaling()
	grid2.processInputData()

	solutionFuture := grid2.build.RunHighsFuture(stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) (weight_types.Weight2Extended, bool) {
		solution := linearResult.GetSolutionAndSaveLog(grid2.printer)

		grid2.printer.Println(solution.Status.String())
		grid2.build.DebugPrintColumns(solution, grid2.printer)

		return grid2.reportOutputWeightsGrid(solution), true
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
		grid2.scaleSims[simType] = 1

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

func (grid2 *GridStatWeightProcess2) calcSimDiff(one *weight_types.WeightInput, two *weight_types.WeightInput, simType stats.SimType) (float64, bool) {
	simOne := one.SimResult.GetFriendly(simType)
	simTwo := two.SimResult.GetFriendly(simType)
	diff := simOne - simTwo
	diff *= grid2.scaleSims[simType]
	if math.Abs(diff) >= c_grid2_simDiffMinimumUseful {
		return diff, true
	} else {
		return 0, false
	}
}

func (grid2 *GridStatWeightProcess2) processInputData() {
	for a := range grid2.inputData {
		for b := a + 1; b < len(grid2.inputData); b++ {
			differenceCount, diffStatA, diffStatB, diffStatC := grid2.checkForNumberStatDifferences(&grid2.inputData[a].TotalStat, &grid2.inputData[b].TotalStat)
			switch differenceCount {
			case 1:
				if grid2.IncludeDiffs1 {
					grid2.prepareSampleOneDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA)
				}

			case 2:
				if grid2.IncludeDiffs2 {
					grid2.prepareSampleTwoDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA, diffStatB)
				}

			case 3:
				if grid2.IncludeDiffs3 {
					grid2.prepareSampleThreeDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA, diffStatB, diffStatC)
				}
			}
		}
	}
}

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

func (grid2 *GridStatWeightProcess2) prepareSampleOneDifferenceStats(one *weight_types.WeightInput, two *weight_types.WeightInput, statType stats.StatType) {
	statDiff := one.TotalStat.GetFloat(statType) - two.TotalStat.GetFloat(statType)

	for _, simType := range grid2.simTypes {
		weightColumn := grid2.detailedWeights.GetOrPanic(statType, simType)

		simDiff, goodDiff := grid2.calcSimDiff(one, two, simType)
		if goodDiff {
			debugText := "MISMATCH1 " + statType.Name() + " " + simType.Name()
			mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), 1, util_highs.DebugString{Text: debugText})

			grid2.build.AbsoluteValueFromDiffOneToConst(weightColumn, statDiff, simDiff, mismatchCol, debugText)
		}
	}
}

func (grid2 *GridStatWeightProcess2) prepareSampleTwoDifferenceStats(one *weight_types.WeightInput, two *weight_types.WeightInput, statTypeA stats.StatType, statTypeB stats.StatType) {
	statDiffA := one.TotalStat.GetFloat(statTypeA) - two.TotalStat.GetFloat(statTypeA)
	statDiffB := one.TotalStat.GetFloat(statTypeB) - two.TotalStat.GetFloat(statTypeB)

	for _, simType := range grid2.simTypes {
		weightColumnA := grid2.detailedWeights.GetOrPanic(statTypeA, simType)
		weightColumnB := grid2.detailedWeights.GetOrPanic(statTypeB, simType)

		simDiff, goodDiff := grid2.calcSimDiff(one, two, simType)
		if goodDiff {
			debugText := "MISMATCH2 " + statTypeA.Name() + " " + statTypeB.Name() + " " + simType.Name()
			mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), 1, util_highs.DebugString{Text: debugText})

			grid2.build.AbsoluteValueFromSumTwoThenDiffToConst(
				weightColumnA, statDiffA,
				weightColumnB, statDiffB,
				simDiff,
				mismatchCol,
				debugText)
		}
	}
}

func (grid2 *GridStatWeightProcess2) prepareSampleThreeDifferenceStats(one *weight_types.WeightInput, two *weight_types.WeightInput, statTypeA stats.StatType, statTypeB stats.StatType, statTypeC stats.StatType) {
	statDiffA := one.TotalStat.GetFloat(statTypeA) - two.TotalStat.GetFloat(statTypeA)
	statDiffB := one.TotalStat.GetFloat(statTypeB) - two.TotalStat.GetFloat(statTypeB)
	statDiffC := one.TotalStat.GetFloat(statTypeC) - two.TotalStat.GetFloat(statTypeC)

	for _, simType := range grid2.simTypes {
		weightColumnA := grid2.detailedWeights.GetOrPanic(statTypeA, simType)
		weightColumnB := grid2.detailedWeights.GetOrPanic(statTypeB, simType)
		weightColumnC := grid2.detailedWeights.GetOrPanic(statTypeC, simType)

		simDiff, goodDiff := grid2.calcSimDiff(one, two, simType)
		if goodDiff {
			debugText := "MISMATCH3 " + statTypeA.Name() + " " + statTypeB.Name() + " " + statTypeC.Name() + " " + simType.Name()
			mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), 1, util_highs.DebugString{Text: debugText})

			grid2.build.AbsoluteValueFromSumSeveralThenDiffToConst(
				[]util_highs.ColumnIndex{weightColumnA, weightColumnB, weightColumnC},
				[]float64{statDiffA, statDiffB, statDiffC},
				simDiff,
				mismatchCol,
				debugText)
		}
	}
}

func (grid2 *GridStatWeightProcess2) reportOutputWeightsGrid(solution *highs.Solution) weight_types.Weight2Extended {
	result := weight_types.Weight2Extended_Make(grid2.statTypes, grid2.simTypes)
	grid2.printer.Println("FINAL WEIGHTS:")

	// weight * (statOne - statTwo) * statScale[type] = (simOne - simTwo) * simScale[type]
	// weight = (simOne - simTwo) / (statOne - statTwo) * (simScale[type] / statScale[type])
	// usableweight = (simOne - simTwo) / (statOne - statTwo)
	// weight = usableweight * (simScale[type] / statScale[type])
	// usableweight = weight * (statScale[type] / simScale[type])

	for _, statType := range grid2.statTypes {
		grid2.printer.Printf("%10s >>>>>\n", statType.Name())

		for simType, detailWeightCol := range grid2.detailedWeights.SeqKey2ValueWithKey1(statType) {
			weight := solution.ColValues[detailWeightCol]
			usableWeight := weight / grid2.scaleSims[simType]

			result.PutWeight(statType, simType, usableWeight)

			grid2.printer.Printf("         %5s > %f %f\n", simType.Name(), weight, usableWeight)
		}
	}

	for _, simType := range grid2.simTypes {
		result.SetSimScale(simType, 1, 0, grid2.targetRatios.GetOrPanic(simType))
	}

	result.FinishAndValidate()
	return *result
}
