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

	targetRatios simulate.SimResultStats
	inputData    []WeightInput

	scaleSims  map[simulate.SimResultType]float64
	scaleStats map[stats.StatType]float64

	input           utilhighs.InputBuilder
	detailedWeights util.MapMap[stats.StatType, simulate.SimResultType, utilhighs.ColumnIndex]
	finalWeights    map[stats.StatType]utilhighs.ColumnIndex
}

type gridDataSample2 struct {
	value float64
}

func (grid2 *GridStatWeightProcess2) Init(printer *util.PrintRecorder) {
	grid2.printer = printer
	grid2.input.Minimise = true
	// grid2.input.Solver = "pdlp"
	grid2.finalWeights = make(map[stats.StatType]utilhighs.ColumnIndex)
}

func (grid2 *GridStatWeightProcess2) SupplyData(inputData []WeightInput) {
	grid2.inputData = inputData
}

func (grid2 *GridStatWeightProcess2) SetTargetRatios(targetRatios simulate.SimResultStats) {
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
	grid2.chooseSimScaling()
	grid2.chooseStatScaling()
	grid2.processInputData()
	grid2.calcTotalRatings()

	solution, log := grid2.input.RunHighs()
	grid2.printer.AppendOther(log)
	grid2.printer.Println(solution.Status.String())

	grid2.input.DebugPrintColumns(solution, grid2.printer)

	return grid2.reportOutputWeightsGrid(solution, grid2.finalWeights, grid2.printer)
}

func (grid2 *GridStatWeightProcess2) setupWeightVars() {
	for _, statType := range G_RequiredStats {
		colFinalWeight := grid2.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "FINAL WEIGHT: " + statType.Name()})
		// colFinalWeight := basic.input.CreateColumnGeneral(highs.Continuous, -c_finalWeightLimit, c_finalWeightLimit)
		grid2.finalWeights[statType] = colFinalWeight
	}

	for _, statType := range G_RequiredStats {
		for _, simType := range G_RequiredSims {
			colDetailWeight := grid2.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "WEIGHT: " + statType.Name() + " " + simType.String()})
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

func (grid2 *GridStatWeightProcess2) chooseSimScaling() {
	c_targetNumber := 1.0
	grid2.scaleSims = make(map[simulate.SimResultType]float64)
	for _, simType := range G_RequiredSims {
		total := 0.0
		for data := range util.ForPointer(grid2.inputData) {
			total += data.SimResult.Get(simType)
		}

		average := total / float64(len(grid2.inputData))
		if average != 0 {
			scale := c_targetNumber / average
			grid2.scaleSims[simType] = scale
		} else {
			grid2.scaleSims[simType] = 1
		}

		grid2.printer.Printf("scale %s %e\n", simType.String(), grid2.scaleSims[simType])
	}
}

func (grid2 *GridStatWeightProcess2) chooseStatScaling() {
	c_targetNumber := 1.0
	grid2.scaleStats = make(map[stats.StatType]float64)
	for _, statType := range G_RequiredStats {
		total := 0.0
		for data := range util.ForPointer(grid2.inputData) {
			total += float64(data.TotalStat.Get(statType))
		}

		average := total / float64(len(grid2.inputData))
		if average != 0 {
			scale := c_targetNumber / average
			grid2.scaleStats[statType] = scale
		} else {
			grid2.scaleStats[statType] = 1
		}

		grid2.printer.Printf("scale %s %.8f\n", statType.Name(), grid2.scaleStats[statType])
	}
}

func (grid2 *GridStatWeightProcess2) processInputData() {
	for a := range grid2.inputData {
		for b := a + 1; b < len(grid2.inputData); b++ {
			differenceCount, diffStatA, diffStatB := grid2.checkForNumberStatDifferences(&grid2.inputData[a].TotalStat, &grid2.inputData[b].TotalStat)
			switch differenceCount {
			case 1:
				grid2.prepareSampleOneDifferenceStats(&grid2.inputData[a], &grid2.inputData[b], diffStatA)
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

	statDiff := float64(one.TotalStat.Get(statType) - two.TotalStat.Get(statType))
	statDiff *= grid2.scaleStats[statType]

	for _, simType := range G_RequiredSims {
		weightColumn := grid2.detailedWeights.GetOrPanic(statType, simType)

		debugText := "MISMATCH1 signed " + statType.Name() + " " + simType.String()
		mismatchSignedCol := grid2.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: debugText})
		debugText = "MISMATCH1 " + statType.Name() + " " + simType.String()
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

	statDiffA := float64(one.TotalStat.Get(statTypeA) - two.TotalStat.Get(statTypeA))
	statDiffB := float64(one.TotalStat.Get(statTypeB) - two.TotalStat.Get(statTypeB))
	statDiffA *= grid2.scaleStats[statTypeA]
	statDiffB *= grid2.scaleStats[statTypeB]

	for _, simType := range G_RequiredSims {
		weightColumnA := grid2.detailedWeights.GetOrPanic(statTypeA, simType)
		weightColumnB := grid2.detailedWeights.GetOrPanic(statTypeB, simType)

		debugText := "MISMATCH2 " + statTypeA.Name() + " " + statTypeB.Name() + " " + simType.String()
		mismatchCol := grid2.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: debugText})

		simDiff := one.SimResult.GetFriendly(simType) - two.SimResult.GetFriendly(simType)
		simDiff *= grid2.scaleSims[simType]

		// not based directly on above
		// weightA * statDiffA - weightB * statDiffB + mismatch = simDiff
		utilhighs.AbsoluteValueFromDiffWithOffset(&grid2.input,
			weightColumnA, statDiffA,
			weightColumnB, statDiffB,
			mismatchCol,
			simDiff,
			debugText)
	}
}

func (grid2 *GridStatWeightProcess2) twoSamplesDifferenceAddToModel(oneSample float64, oneWeightCol utilhighs.ColumnIndex, twoSample float64, twoWeightCol utilhighs.ColumnIndex, debugText string) {
	offsetAbs := grid2.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: "OFFSET " + debugText})
	utilhighs.AbsoluteValueFromDiff(&grid2.input,
		oneWeightCol, twoSample,
		twoWeightCol, oneSample,
		offsetAbs, "OFFSET "+debugText)
}

func (grid2 *GridStatWeightProcess2) calcTotalRatings() {
	for _, statType := range G_RequiredStats {
		statFinalRow := utilhighs.ConstraintRowBuild{}
		for _, detailColumn := range grid2.detailedWeights.SeqInnerWithKey1Value(statType) {
			statFinalRow.Add(detailColumn, 1)
		}

		finalWeightColumn := grid2.finalWeights[statType]
		statFinalRow.Add(finalWeightColumn, -1)
		statFinalRow.Finish(&grid2.input, 0, 0)
	}
}

func (grid2 *GridStatWeightProcess2) reportOutputWeightsGrid(solution *highs.Solution, weightColumns map[stats.StatType]utilhighs.ColumnIndex, printer *util.PrintRecorder) map[stats.StatType]float64 {
	result := make(map[stats.StatType]float64)
	printer.Println("FINAL WEIGHTS:")
	for _, statType := range G_RequiredStats {
		columnIndex := weightColumns[statType]
		value := solution.ColValues[columnIndex]
		printer.Printf("%10s %f\n", statType.Name(), value)
		result[statType] = value
	}
	return result
}
