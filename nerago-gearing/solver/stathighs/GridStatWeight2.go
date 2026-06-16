package stathighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

type GridStatWeightProcess2 struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimData
	inputData    []WeightInput

	scaleSims  map[simulate.SimType]float64
	scaleStats map[stats.StatType]float64

	build           utilhighs.LinearBuilder
	detailedWeights util.MapMap[stats.StatType, simulate.SimType, utilhighs.ColumnIndex]
}

type gridDataSample2 struct {
	value float64
}

func (grid2 *GridStatWeightProcess2) Init(printer *util.PrintRecorder) {
	grid2.printer = printer
	grid2.build.Minimise = true
	// grid2.input.Solver = "pdlp"
}

func (grid2 *GridStatWeightProcess2) SupplyData(inputData []WeightInput) {
	grid2.inputData = inputData
}

func (grid2 *GridStatWeightProcess2) SetTargetRatios(targetRatios simulate.SimData) {
	sum := 0.0
	for _, simType := range G_RequiredSims {
		val := targetRatios.Get(simType)
		if val <= 0 {
			panic("missing ratio")
		}
		sum += val
	}
	if !utilhighs.FloatEqualsOne(sum) {
		panic("ratios don't add to one")
	}

	grid2.targetRatios = targetRatios
}

func (grid2 *GridStatWeightProcess2) Run() WeightResult {
	grid2.setupWeightVars()
	grid2.chooseScaling()
	grid2.processInputData()

	solution, log := grid2.build.RunHighs()
	grid2.printer.AppendOther(log)
	grid2.printer.Println(solution.Status.String())

	grid2.build.DebugPrintColumns(solution, grid2.printer)

	return grid2.reportOutputWeightsGrid(solution)
}

func (grid2 *GridStatWeightProcess2) setupWeightVars() {
	// create detail columns
	for _, statType := range G_RequiredStats {
		for _, simType := range G_RequiredSims {
			colDetailWeight := grid2.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "WEIGHT: " + statType.Name() + " " + simType.Name()})
			grid2.detailedWeights.Put(statType, simType, colDetailWeight)
		}
	}

	// strength column within each simtype is set to targetratio (0.4 etc)
	for _, simType := range G_RequiredSims {
		value := grid2.targetRatios.Get(simType)
		colDetailWeight := grid2.detailedWeights.GetOrPanic(c_baseStatType, simType)
		strengthSetToRatio := utilhighs.ConstraintRow{}
		strengthSetToRatio.Add(colDetailWeight, 1)
		strengthSetToRatio.Build(&grid2.build, value, value)
	}
}

func (grid2 *GridStatWeightProcess2) chooseScaling() {
	grid2.scaleSims = chooseSimScaling(grid2.inputData, grid2.printer)
	grid2.scaleStats = chooseStatScaling(grid2.inputData, grid2.printer)
}

func (grid2 *GridStatWeightProcess2) processInputData() {
	for a := range grid2.inputData {
		for b := a + 1; b < len(grid2.inputData); b++ {
			differenceCount, diffStatA, diffStatB, _ := grid2.checkForNumberStatDifferences(&grid2.inputData[a].TotalStat, &grid2.inputData[b].TotalStat)
			switch differenceCount {
			case 1:
				grid2.prepareSampleOneDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA)
			case 2:
				grid2.prepareSampleTwoDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA, diffStatB)
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
	for _, stat := range G_RequiredStats {
		if one[stat] != two[stat] {
			switch differenceCount {
			case 0:
				diffStatA = stats.StatType(stat)
			case 1:
				diffStatB = stats.StatType(stat)
			case 2:
				diffStatC = stats.StatType(stat)
			}
			differenceCount++
		}
	}
	return differenceCount, diffStatA, diffStatB, diffStatC
}

func (grid2 *GridStatWeightProcess2) prepareSampleOneDifferenceStats(one *WeightInput, two *WeightInput, statType stats.StatType) {
	statDiff := one.TotalStat.GetFloat(statType) - two.TotalStat.GetFloat(statType)
	statDiff *= grid2.scaleStats[statType]

	for _, simType := range G_RequiredSims {
		weightColumn := grid2.detailedWeights.GetOrPanic(statType, simType)

		debugText := "MISMATCH1 " + statType.Name() + " " + simType.Name()
		mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: debugText})

		simDiff := two.SimResult.GetFriendly(simType) - one.SimResult.GetFriendly(simType)
		simDiff *= grid2.scaleSims[simType]

		utilhighs.AbsoluteValueFromDiffOneToConst(&grid2.build, weightColumn, statDiff, simDiff, mismatchCol, debugText)
	}
}

func (grid2 *GridStatWeightProcess2) prepareSampleTwoDifferenceStats(one *WeightInput, two *WeightInput, statTypeA stats.StatType, statTypeB stats.StatType) {
	statDiffA := one.TotalStat.GetFloat(statTypeA) - two.TotalStat.GetFloat(statTypeA)
	statDiffB := one.TotalStat.GetFloat(statTypeB) - two.TotalStat.GetFloat(statTypeB)
	statDiffA *= grid2.scaleStats[statTypeA]
	statDiffB *= grid2.scaleStats[statTypeB]

	for _, simType := range G_RequiredSims {
		weightColumnA := grid2.detailedWeights.GetOrPanic(statTypeA, simType)
		weightColumnB := grid2.detailedWeights.GetOrPanic(statTypeB, simType)

		debugText := "MISMATCH2 " + statTypeA.Name() + " " + statTypeB.Name() + " " + simType.Name()
		mismatchCol := grid2.build.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: debugText})

		simDiff := two.SimResult.GetFriendly(simType) - one.SimResult.GetFriendly(simType)
		simDiff *= grid2.scaleSims[simType]

		utilhighs.AbsoluteValueDiffTwoVarsThenDiffConst(&grid2.build,
			weightColumnA, statDiffA,
			weightColumnB, statDiffB,
			mismatchCol,
			simDiff,
			debugText)
	}
}

func (grid2 *GridStatWeightProcess2) prepareSampleThreeDifferenceStats(one *WeightInput, two *WeightInput, statTypeA stats.StatType, statTypeB stats.StatType, statTypeC stats.StatType) {
	statDiffA := one.TotalStat.GetFloat(statTypeA) - two.TotalStat.GetFloat(statTypeA)
	statDiffA *= grid2.scaleStats[statTypeA]
	statDiffB := one.TotalStat.GetFloat(statTypeB) - two.TotalStat.GetFloat(statTypeB)
	statDiffB *= grid2.scaleStats[statTypeB]
	statDiffC := one.TotalStat.GetFloat(statTypeC) - two.TotalStat.GetFloat(statTypeC)
	statDiffC *= grid2.scaleStats[statTypeC]

	for _, simType := range G_RequiredSims {
		weightColumnA := grid2.detailedWeights.GetOrPanic(statTypeA, simType)
		weightColumnB := grid2.detailedWeights.GetOrPanic(statTypeB, simType)
		weightColumnC := grid2.detailedWeights.GetOrPanic(statTypeC, simType)

		simDiff := two.SimResult.GetFriendly(simType) - one.SimResult.GetFriendly(simType)
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
		utilhighs.AbsoluteValue(&grid2.build, mismatchSignedCol, mismatchCol)
	}
}

func (grid2 *GridStatWeightProcess2) reportOutputWeightsGrid(solution *highs.Solution) WeightResult {
	result := WeightResult_Make()
	grid2.printer.Println("FINAL WEIGHTS:")

	for _, statType := range G_RequiredStats {
		grid2.printer.Printf("%10s >>>>>\n", statType.Name())

		sumIndividual := 0.0

		for simType, detailWeightCol := range grid2.detailedWeights.SeqInnerWithKey1Value(statType) {
			weight := solution.ColValues[detailWeightCol]

			usableWeight := weight / grid2.scaleSims[simType]

			if !simType.IsHighGood() {
				usableWeight *= -1
			}

			grid2.printer.Printf("         %5s > %f %f\n", simType.Name(), weight, usableWeight)

			sumIndividual += usableWeight
		}

		grid2.printer.Printf("             === %f\n", sumIndividual)
		result.Put(statType, sumIndividual)
	}

	divideBy := result[stats.Stat_Strength]
	for _, statType := range G_RequiredStats {
		result[statType] /= divideBy
	}

	return result
}
