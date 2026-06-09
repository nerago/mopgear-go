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

	input           utilhighs.InputBuilder
	detailedWeights util.MapMap[stats.StatType, simulate.SimType, utilhighs.ColumnIndex]
	// finalWeights    map[stats.StatType]utilhighs.ColumnIndex
}

type gridDataSample2 struct {
	value float64
}

func (grid2 *GridStatWeightProcess2) Init(printer *util.PrintRecorder) {
	grid2.printer = printer
	grid2.input.Minimise = true
	// grid2.input.Solver = "pdlp"
	// grid2.finalWeights = make(map[stats.StatType]utilhighs.ColumnIndex)
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

func (grid2 *GridStatWeightProcess2) Run() map[stats.StatType]float64 {
	grid2.setupWeightVars()
	grid2.chooseScaling()
	grid2.processInputData()
	// grid2.calcTotalRatings()

	solution, log := grid2.input.RunHighs()
	grid2.printer.AppendOther(log)
	grid2.printer.Println(solution.Status.String())

	grid2.input.DebugPrintColumns(solution, grid2.printer)

	return grid2.reportOutputWeightsGrid(solution)
}

func (grid2 *GridStatWeightProcess2) setupWeightVars() {
	// for _, statType := range G_RequiredStats {
	// 	colFinalWeight := grid2.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "FINAL WEIGHT: " + statType.Name()})
	// 	// colFinalWeight := basic.input.CreateColumnGeneral(highs.Continuous, -c_finalWeightLimit, c_finalWeightLimit)
	// 	grid2.finalWeights[statType] = colFinalWeight
	// }

	for _, statType := range G_RequiredStats {
		for _, simType := range G_RequiredSims {
			colDetailWeight := grid2.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "WEIGHT: " + statType.Name() + " " + simType.Name()})
			grid2.detailedWeights.Put(statType, simType, colDetailWeight)
		}
	}

	for _, simType := range G_RequiredSims {
		value := grid2.targetRatios.Get(simType)
		colDetailWeight := grid2.detailedWeights.GetOrPanic(c_baseStatType, simType)
		strengthSetToRatio := utilhighs.ConstraintRowBuild{}
		strengthSetToRatio.Add(colDetailWeight, 1)
		strengthSetToRatio.Finish(&grid2.input, value, value)
	}
}

func (grid2 *GridStatWeightProcess2) chooseScaling() {
	grid2.scaleSims = chooseSimScaling(grid2.inputData, grid2.printer)
	grid2.scaleStats = chooseStatScaling(grid2.inputData, grid2.printer)
}

func (grid2 *GridStatWeightProcess2) processInputData() {
	for a := range grid2.inputData {
		for b := a + 1; b < len(grid2.inputData); b++ {
			differenceCount, diffStatA, diffStatB := grid2.checkForNumberStatDifferences(&grid2.inputData[a].TotalStat, &grid2.inputData[b].TotalStat)
			switch differenceCount {
			case 1:
				// grid2.prepareSampleOneDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA)
			case 2:
				grid2.prepareSampleTwoDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA, diffStatB)
			}
		}
	}
}

func (grid2 *GridStatWeightProcess2) checkForNumberStatDifferences(one, two *stats.StatBlock) (differenceCount int, diffStatA stats.StatType, diffStatB stats.StatType) {
	for stat := range one {
		if one[stat] != two[stat] {
			switch differenceCount {
			case 0:
				diffStatA = stats.StatType(stat)
			case 1:
				diffStatB = stats.StatType(stat)
			}
			differenceCount++
		}
	}
	return differenceCount, diffStatA, diffStatB
}

func (grid2 *GridStatWeightProcess2) prepareSampleOneDifferenceStats(one *WeightInput, two *WeightInput, statType stats.StatType) {
	// maybe old formula applies somehow still? dunno, thats a separates pair of differences
	// formula: detailweight_dps_haste * unit_dps_strength - detailweight_dps_strength * unit_dps_haste + offset = 0

	statDiff := one.TotalStat.GetFloat(statType) - two.TotalStat.GetFloat(statType)
	statDiff *= grid2.scaleStats[statType]

	for _, simType := range G_RequiredSims {
		weightColumn := grid2.detailedWeights.GetOrPanic(statType, simType)

		debugText := "MISMATCH1 signed " + statType.Name() + " " + simType.Name()
		mismatchSignedCol := grid2.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: debugText})
		debugText = "MISMATCH1 " + statType.Name() + " " + simType.Name()
		mismatchCol := grid2.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: debugText})

		simDiff := one.SimResult.GetFriendly(simType) - two.SimResult.GetFriendly(simType)
		simDiff *= grid2.scaleSims[simType]

		rowCompare := utilhighs.ConstraintRowBuild{}
		rowCompare.Add(weightColumn, statDiff)
		rowCompare.Add(mismatchSignedCol, 1)
		rowCompare.Finish(&grid2.input, simDiff, simDiff)

		utilhighs.AbsoluteValue(&grid2.input, mismatchSignedCol, mismatchCol)
	}
}

func (grid2 *GridStatWeightProcess2) prepareSampleTwoDifferenceStats(one *WeightInput, two *WeightInput, statTypeA stats.StatType, statTypeB stats.StatType) {
	// formula: detailweight_dps_haste * unit_dps_strength - detailweight_dps_strength * unit_dps_haste + offset = 0
	// general form: weightA / unitValueA == weightB / unitValueB
	//   -->         weightA / (simA / statA) == weightB / (simB / statB)
	//   -->         (simB / statB) / (simA / statA) == weightB / weightA				(in words, we're checking on the ratio of different valuation of the two stats)

	// so the old method we're comparing   baseSim[0, 0]<=>strengthSimPlus[0, +s] vs baseSim[0, 0]<=>hasteSimPlus[+h, 0]
	// what if we pretend that we have that sim too even though its unavailable at least at this point
	// pretendSim[a b]<=>strengthSimPlus[a, b+strengthIncrement] vs baseSim[0 0]<=>hasteSimPlus[a+hasteIncrement, b]
	// (strengthSimPlus-pretendSim)/strengthIncrement vs (hasteSimPlus-baseSim)/hasteIncrement

	// basic method formula, as with spreadsheet
	// unit_dps_haste = (this_dps[haste] - base_dps) / this_haste_increment
	// unit_dps_strength = (this_dps[strength] - base_dps) / this_strength_increment
	// Weight[x] = unit_stat[x] / unit_stat[str] * weight[str]
	// Weight[x] / weight[str] = unit_stat[x] / unit_stat[str]

	// formula: detailweight_dps_haste * (simValSingleDiff / strengthSingleDiff) - detailweight_dps_strength * (simValOtherDiff / hasteSingleDiff) + offset = 0

	statDiffA := one.TotalStat.GetFloat(statTypeA) - two.TotalStat.GetFloat(statTypeA)
	statDiffB := one.TotalStat.GetFloat(statTypeB) - two.TotalStat.GetFloat(statTypeB)
	statDiffA *= grid2.scaleStats[statTypeA]
	statDiffB *= grid2.scaleStats[statTypeB]

	for _, simType := range G_RequiredSims {
		weightColumnA := grid2.detailedWeights.GetOrPanic(statTypeA, simType)
		weightColumnB := grid2.detailedWeights.GetOrPanic(statTypeB, simType)

		debugText := "MISMATCH2 " + statTypeA.Name() + " " + statTypeB.Name() + " " + simType.Name()
		mismatchCol := grid2.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: debugText})

		simDiff := one.SimResult.GetFriendly(simType) - two.SimResult.GetFriendly(simType)
		simDiff *= grid2.scaleSims[simType]

		// not based directly on above
		// weightA * statDiffA - weightB * statDiffB + mismatch = simDiff
		utilhighs.AbsoluteValueFromDiffTwoVarsWithOffset(&grid2.input,
			weightColumnA, statDiffA,
			weightColumnB, statDiffB,
			mismatchCol,
			simDiff,
			debugText)
	}
}

func (grid2 *GridStatWeightProcess2) twoSamplesDifferenceAddToModel(oneSample float64, oneWeightCol utilhighs.ColumnIndex, twoSample float64, twoWeightCol utilhighs.ColumnIndex, debugText string) {
	offsetAbs := grid2.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: "OFFSET " + debugText})
	utilhighs.AbsoluteValueFromDiffTwoVars(&grid2.input,
		oneWeightCol, twoSample,
		twoWeightCol, oneSample,
		offsetAbs, "OFFSET "+debugText)
}

// func (grid2 *GridStatWeightProcess2) calcTotalRatings() {
// 	for _, statType := range G_RequiredStats {
// 		statFinalRow := utilhighs.ConstraintRowBuild{}
// 		for simType, detailColumn := range grid2.detailedWeights.SeqInnerWithKey1Value(statType) {
// 			scale := 1.0 / grid2.scaleSims[simType]
// 			if simType.IsHighGood() || statType == stats.Stat_Strength {
// 				scale *= 1
// 			} else {
// 				scale *= -1
// 			}
// 			statFinalRow.Add(detailColumn, scale)
// 		}

// 		finalWeightColumn := grid2.finalWeights[statType]
// 		statFinalRow.Add(finalWeightColumn, -1)
// 		statFinalRow.Finish(&grid2.input, 0, 0)
// 	}
// }

func (grid2 *GridStatWeightProcess2) reportOutputWeightsGrid(solution *highs.Solution) map[stats.StatType]float64 {
	result := make(map[stats.StatType]float64)
	grid2.printer.Println("FINAL WEIGHTS:")

	for _, statType := range G_RequiredStats {
		grid2.printer.Printf("%10s >>>>>\n", statType.Name())

		sumIndividual := 0.0

		for simType, detailWeightCol := range grid2.detailedWeights.SeqInnerWithKey1Value(statType) {
			weight := solution.ColValues[detailWeightCol]

			scaleFix := grid2.scaleStats[statType] / grid2.scaleSims[simType]
			usableWeight := weight * scaleFix

			if !simType.IsHighGood() {
				usableWeight *= -1
			}

			grid2.printer.Printf("         %5s > %f %f\n", simType.Name(), weight, usableWeight)

			sumIndividual += usableWeight
		}

		// grid2.printer.Printf("%10s %f\n", statType.Name(), sumIndividual)
		grid2.printer.Printf("             === %f\n", sumIndividual)
		result[statType] = sumIndividual
	}
	return result
}
