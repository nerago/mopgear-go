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

	input           utilhighs.InputBuilder
	detailedWeights util.MapMap[stats.StatType, simulate.SimResultType, utilhighs.ColumnIndex]
	finalWeights    map[stats.StatType]utilhighs.ColumnIndex
}

type gridDataSample2 struct {
	value float64
}

func (grid *GridStatWeightProcess2) Init(printer *util.PrintRecorder) {
	grid.printer = printer
	grid.input.Minimise = true
	grid.input.Solver = "pdlp"
	grid.finalWeights = make(map[stats.StatType]utilhighs.ColumnIndex)
}

func (grid *GridStatWeightProcess2) SupplyData(inputData []WeightInput) {
	grid.inputData = inputData
}

func (grid *GridStatWeightProcess2) SetTargetRatios(targetRatios simulate.SimResultStats) {
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

	grid.targetRatios = targetRatios
}

func (grid *GridStatWeightProcess2) Run() map[stats.StatType]float64 {
	grid.setupWeightVars()

	grid.dataSamplesFromPairs()
	grid.checkSampleRange()
	grid.unitValuesToCalcDetailedRatings()
	grid.calcTotalRatings()

	solution, log := grid.input.RunHighs()
	grid.printer.AppendOther(log)
	grid.printer.Println(solution.Status.String())

	grid.input.DebugPrintColumns(solution, grid.printer)

	return grid.reportOutputWeightsGrid(solution, grid.finalWeights, grid.printer)
}

func (grid *GridStatWeightProcess2) setupWeightVars() {
	for _, statType := range G_RequiredStats {
		colFinalWeight := grid.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "FINAL WEIGHT: " + statType.Name()})
		// colFinalWeight := basic.input.CreateColumnGeneral(highs.Continuous, -c_finalWeightLimit, c_finalWeightLimit)
		grid.finalWeights[statType] = colFinalWeight
	}

	for _, statType := range G_RequiredStats {
		for _, simType := range G_RequiredSims {
			colDetailWeight := grid.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "WEIGHT: " + statType.Name() + " " + simType.String()})
			grid.detailedWeights.Put(statType, simType, colDetailWeight)
		}
	}

	for _, simType := range G_RequiredSims {
		value := grid.targetRatios.Get(simType)
		colDetailWeight := grid.detailedWeights.GetOrPanic(c_baseStatType, simType)
		strengthSetToRatio := utilhighs.ConstraintRowBuild{}
		strengthSetToRatio.Add(colDetailWeight, 1)
		strengthSetToRatio.Finish(&grid.input, value, value)
	}
}

func (grid *GridStatWeightProcess2) dataSamplesFromPairs() {
	for a := range grid.inputData {
		for b := a + 1; b < len(grid.inputData); b++ {
			differenceCount, diffStatA, diffStatB := grid.checkForNumberStatDifferences(&grid.inputData[a].TotalStat, &grid.inputData[b].TotalStat)
			if differenceCount == 1 {
				grid.prepareSampleOneDifferenceStats(&grid.inputData[a], &grid.inputData[b], diffStatA)
			} else if differenceCount == 2 {
				grid.prepareSampleTwoDifferenceStats(&grid.inputData[a], &grid.inputData[b], diffStatA, diffStatB)
			}
		}
	}
}

func (grid *GridStatWeightProcess2) checkForNumberStatDifferences(one, two *stats.StatBlock) (differenceCount int, diffStatA stats.StatType, diffStatB stats.StatType) {
	for stat := range one {
		if one[stat] != two[stat] {
			if differenceCount == 0 {
				diffStatA = stats.StatType(stat)
			} else if differenceCount == 1 {
				diffStatB = stats.StatType(stat)
			}
			differenceCount++
		}
	}
	return differenceCount, diffStatA, diffStatB
}

func (grid *GridStatWeightProcess2) prepareSampleOneDifferenceStats(one *WeightInput, two *WeightInput, statType stats.StatType) {
	// maybe old formula applies somehow still? dunno, thats a separates pair of differences
	// formula: detailweight_dps_haste * unit_dps_strength - detailweight_dps_strength * unit_dps_haste + offset = 0

	statDiff := one.TotalStat.Get(statType) - two.TotalStat.Get(statType)

	for _, simType := range G_RequiredSims {
		simValueDiff := one.SimResult.GetFriendly(simType) - two.SimResult.GetFriendly(simType)

		valuePerUnitOfStat := simValueDiff / float64(statDiff)

		dataSample := gridDataSample2{valuePerUnitOfStat}
		dataSample = dataSample

		// grid.unitStatValues.Add(statType, simType, dataSample)
	}
}

func (grid *GridStatWeightProcess2) prepareSampleTwoDifferenceStats(one *WeightInput, two *WeightInput, statTypeA stats.StatType, statTypeB stats.StatType) {
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

	for _, simType := range G_RequiredSims {
		weightColumnA := grid.detailedWeights.GetOrPanic(statTypeA, simType)
		weightColumnB := grid.detailedWeights.GetOrPanic(statTypeB, simType)

		debugText := "MISMATCH " + statTypeA.Name() + " " + statTypeB.Name() + " " + simType.String()
		mismatchCol := grid.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: debugText})

		simDiff := one.SimResult.GetFriendly(simType) - two.SimResult.GetFriendly(simType)

		// not based directly on above
		// weightA * statDiffA - weightB * statDiffB + mismatch = simDiff
		utilhighs.AbsoluteValueFromDiffWithOffset(&grid.input,
			weightColumnA, statDiffA,
			weightColumnB, statDiffB,
			mismatchCol,
			simDiff,
			debugText)
	}
}

func (grid *GridStatWeightProcess2) twoSamplesDifferenceAddToModel(oneSample float64, oneWeightCol utilhighs.ColumnIndex, twoSample float64, twoWeightCol utilhighs.ColumnIndex, debugText string) {
	offsetAbs := grid.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: "OFFSET " + debugText})
	utilhighs.AbsoluteValueFromDiff(&grid.input,
		oneWeightCol, twoSample,
		twoWeightCol, oneSample,
		offsetAbs, "OFFSET "+debugText)
}

func (grid *GridStatWeightProcess2) prepareSample(statType stats.StatType, high, low *WeightInput) {
	// basic approach (spreadsheet "build ratings miti_2")
	// unit_dps_haste = (this_dps[haste] - base_dps) / this_haste_value
	// detailweight_dps_haste = unit_dps_haste / unit_dps_str * detailweight_str

	statDiff := high.TotalStat.Get(statType) - low.TotalStat.Get(statType)

	for _, simType := range G_RequiredSims {
		var simValueDiff float64
		if simType.IsHighGood() {
			simValueDiff = high.SimResult.GetFriendly(simType) - low.SimResult.GetFriendly(simType)
		} else {
			simValueDiff = low.SimResult.GetFriendly(simType) - high.SimResult.GetFriendly(simType)
		}
		unitStatValue := simValueDiff / float64(statDiff)

		dataSample := gridDataSample2{unitStatValue}
		dataSample = dataSample

		// grid.unitStatValues.Add(statType, simType, dataSample)
	}
}

func (grid *GridStatWeightProcess2) checkSampleRange() {
	// good, bad := 0, 0
	// for entry := range grid.unitStatValues.SeqValues() {
	// 	if isGoodValueRange(entry.value) {
	// 		good++
	// 	} else {
	// 		bad++
	// 	}
	// }
	// grid.printer.Printf("checkSampleRange good=%d bad=%d\n", good, bad)
	// if bad > (good+bad)/5 {
	// 	panic("many values have inconvenient range")
	// }

	// TODO port scaling process from complex weighter
	// TODO check per type rather than on all, or maybe mix. maybe post scaling
}

func (grid *GridStatWeightProcess2) unitValuesToCalcDetailedRatings() {
	// FORMULA:
	// detailweight_dps_haste = unit_dps_haste / unit_dps_str * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = unit_dps_haste * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = detailweight_dps_str * unit_dps_haste
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste = 0
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste + offset = 0  (allow small offset to optimise on)

	// for simType, lookupStat := range grid.unitStatValues.SeqGroupsKey2Lookup() {
	// 	unitValueBaseSeq := lookupStat(c_baseStatType)
	// 	detailWeightBase := grid.detailedWeights.GetOrPanic(c_baseStatType, simType)
	// 	for _, thisStatType := range G_RequiredStats {
	// 		if thisStatType != c_baseStatType {
	// 			thisUnitValueSeq := lookupStat(thisStatType)
	// 			grid.unitValuesCalcForGroup(simType, thisStatType, unitValueBaseSeq, thisUnitValueSeq, detailWeightBase)
	// 		}
	// 	}
	// }

	// TODO could be interesting experiment to setup all stat pairings, not just strength base
}

// func (grid *GridStatWeightProcess2) unitValuesCalcForGroup(simType simulate.SimResultType, thisStatType stats.StatType, baseUnitValueSeq iter.Seq[gridDataSample2], thisUnitValueSeq iter.Seq[gridDataSample2], baseDetailWeightCol utilhighs.ColumnIndex) {
// 	debugText := simType.String() + " " + thisStatType.Name()
// 	thisDetailWeightCol := grid.detailedWeights.GetOrPanic(thisStatType, simType)

// 	// look at multiple input values of each unitstat value
// 	index := 0
// 	for baseUnitSample := range baseUnitValueSeq {
// 		for thisUnitSample := range thisUnitValueSeq {
// 			if isGoodValueRange(baseUnitSample.value) && isGoodValueRange(thisUnitSample.value) {
// 				grid.unitValueCombinationAddToModel(baseUnitSample, baseDetailWeightCol, thisUnitSample, thisDetailWeightCol, debugText+" "+strconv.Itoa(index))
// 				index++
// 			}
// 		}
// 	}
// }

func (grid *GridStatWeightProcess2) calcTotalRatings() {
	for _, statType := range G_RequiredStats {
		statFinalRow := utilhighs.ConstraintRowBuild{}
		for _, detailColumn := range grid.detailedWeights.SeqInnerWithKey1Value(statType) {
			statFinalRow.Add(detailColumn, 1)
		}

		finalWeightColumn := grid.finalWeights[statType]
		statFinalRow.Add(finalWeightColumn, -1)
		statFinalRow.Finish(&grid.input, 0, 0)
	}
}

func (grid *GridStatWeightProcess2) reportOutputWeightsGrid(solution *highs.Solution, weightColumns map[stats.StatType]utilhighs.ColumnIndex, printer *util.PrintRecorder) map[stats.StatType]float64 {
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
