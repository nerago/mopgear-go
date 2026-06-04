package stathighs

import (
	"iter"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

type GridStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimResultStats
	inputData    []WeightInput

	input           utilhighs.InputBuilder
	unitStatValues  util.MapMapSlice[stats.StatType, simulate.SimResultType, gridDataSample]
	detailedWeights util.MapMap[stats.StatType, simulate.SimResultType, utilhighs.ColumnIndex]
	finalWeights    map[stats.StatType]utilhighs.ColumnIndex
}

type gridDataSample struct {
	value float64
}

func (grid *GridStatWeightProcess) Init(printer *util.PrintRecorder) {
	grid.printer = printer
	grid.input.Minimise = true
	grid.input.Solver = "pdlp"
	grid.finalWeights = make(map[stats.StatType]utilhighs.ColumnIndex)
}

func (grid *GridStatWeightProcess) SupplyData(inputData []WeightInput) {
	grid.inputData = inputData
}

func (grid *GridStatWeightProcess) SetTargetRatios(targetRatios simulate.SimResultStats) {
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

func (grid *GridStatWeightProcess) Run() map[stats.StatType]float64 {
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

func (grid *GridStatWeightProcess) setupWeightVars() {
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

// lazy func, could avoid double processing and put in order, very N**2
func (grid *GridStatWeightProcess) dataSamplesFromPairs() {
	for a := range grid.inputData {
		for b := range grid.inputData {
			statType, isGood := grid.isGoodPair(&grid.inputData[a].TotalStat, &grid.inputData[b].TotalStat)
			if isGood {
				grid.prepareSample(statType, &grid.inputData[a], &grid.inputData[b])
			}
		}
	}
}

func (grid *GridStatWeightProcess) isGoodPair(blockHigh, blockLow *stats.StatBlock) (stats.StatType, bool) {
	var changedStat stats.StatType
	foundChange := false
	for stat := range blockHigh {
		if blockHigh[stat] < blockLow[stat] {
			// any inequality with high side lower is fail
			return changedStat, false
		} else if blockHigh[stat] > blockLow[stat] {
			if !foundChange {
				// first inequality is high good
				foundChange = true
				changedStat = stats.StatType(stat)
			} else {
				// another inequality is fail, we want just one difference (for now!)
				return changedStat, false
			}
		}
	}
	return changedStat, foundChange
}

func (grid *GridStatWeightProcess) prepareSample(statType stats.StatType, high, low *WeightInput) {
	// basic approach (spreadsheet "build ratings miti_2")
	// unit_dps_haste = (this_dps[haste] - base_dps) / this_haste_value
	// detailweight_dps_haste = unit_dps_haste / unit_dps_str * detailweight_str

	statDiff := high.TotalStat.GetFloat(statType) - low.TotalStat.GetFloat(statType)

	for _, simType := range G_RequiredSims {
		var simValueDiff float64
		if simType.IsHighGood() {
			simValueDiff = high.SimResult.GetFriendly(simType) - low.SimResult.GetFriendly(simType)
		} else {
			simValueDiff = low.SimResult.GetFriendly(simType) - high.SimResult.GetFriendly(simType)
		}
		unitStatValue := simValueDiff / statDiff

		// if (unitStatValue > 0 && unitStatValue <= 1e-9) || (unitStatValue < 0 && unitStatValue >= -1e-9) {
		// 	panic("small values, highs won't like it")
		// }

		dataSample := gridDataSample{unitStatValue}
		grid.unitStatValues.Add(statType, simType, dataSample)
	}
}

func (grid *GridStatWeightProcess) checkSampleRange() {
	good, bad := 0, 0
	for entry := range grid.unitStatValues.SeqValues() {
		if isGoodValueRange(entry.value) {
			good++
		} else {
			bad++
		}
	}
	grid.printer.Printf("checkSampleRange good=%d bad=%d\n", good, bad)
	if bad > (good+bad)/5 {
		panic("many values have inconvenient range")
	}

	// TODO port scaling process from complex weighter
	// TODO check per type rather than on all, or maybe mix. maybe post scaling
}

func (grid *GridStatWeightProcess) unitValuesToCalcDetailedRatings() {
	// FORMULA:
	// detailweight_dps_haste = unit_dps_haste / unit_dps_str * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = unit_dps_haste * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = detailweight_dps_str * unit_dps_haste
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste = 0
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste + offset = 0  (allow small offset to optimise on)

	for simType, lookupStat := range grid.unitStatValues.SeqGroupsKey2Lookup() {
		unitValueBaseSeq := lookupStat(c_baseStatType)
		detailWeightBase := grid.detailedWeights.GetOrPanic(c_baseStatType, simType)
		for _, thisStatType := range G_RequiredStats {
			if thisStatType != c_baseStatType {
				thisUnitValueSeq := lookupStat(thisStatType)
				grid.unitValuesCalcForGroup(simType, thisStatType, unitValueBaseSeq, thisUnitValueSeq, detailWeightBase)
			}
		}
	}

	// TODO could be interesting experiment to setup all stat pairings, not just strength base
}

func (grid *GridStatWeightProcess) unitValuesCalcForGroup(simType simulate.SimResultType, thisStatType stats.StatType, baseUnitValueSeq iter.Seq[gridDataSample], thisUnitValueSeq iter.Seq[gridDataSample], baseDetailWeightCol utilhighs.ColumnIndex) {
	debugText := simType.String() + " " + thisStatType.Name()
	thisDetailWeightCol := grid.detailedWeights.GetOrPanic(thisStatType, simType)

	// look at multiple input values of each unitstat value
	index := 0
	for baseUnitSample := range baseUnitValueSeq {
		for thisUnitSample := range thisUnitValueSeq {
			if isGoodValueRange(baseUnitSample.value) && isGoodValueRange(thisUnitSample.value) {
				grid.unitValueCombinationAddToModel(baseUnitSample, baseDetailWeightCol, thisUnitSample, thisDetailWeightCol, debugText+" "+strconv.Itoa(index))
				index++
			}
		}
	}
}

func (grid *GridStatWeightProcess) unitValueCombinationAddToModel(baseUnitSample gridDataSample, baseDetailWeightCol utilhighs.ColumnIndex,
	thisUnitSample gridDataSample, thisDetailWeightCol utilhighs.ColumnIndex, debugText string) {

	// detailweight_dps_haste * unit_dps_base - detailweight_dps_base * unit_dps_haste + offset = 0
	// detailweight_dps_haste / unit_dps_haste  - detailweight_dps_base   / unit_dps_base + offset / unit_dps_base / unit_dps_haste = 0
	// offsetSigned := grid.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "OFFSET SIGNED " + debugText})
	// weightRow := utilhighs.ConstraintRowBuild{}
	// weightRow.Add(thisDetailWeight, baseUnitSample.value)  // OLD
	// weightRow.Add(detailWeightBase, -thisUnitSample.value) // OLD
	// // weightRow.Add(thisDetailWeight, 1/thisUnitSample.value) // NEW BUT TOO BIG
	// // weightRow.Add(detailWeightBase, -1/baseUnitSample.value) // NEW BUT TOO BIG
	// weightRow.Add(offsetSigned, 1)
	// weightRow.Finish(&grid.input, 0, 0)

	// take absolute value, output for objective function
	offsetAbs := grid.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: "OFFSET ABS " + debugText})
	// utilhighs.AbsoluteValueFromDiff(&grid.input, offsetSigned, offsetAbs)

	utilhighs.AbsoluteValueFromDiff(&grid.input, thisDetailWeightCol, baseUnitSample.value, baseDetailWeightCol, thisUnitSample.value, offsetAbs, "OFFSET ABS "+debugText)
}

func (grid *GridStatWeightProcess) calcTotalRatings() {
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

func (grid *GridStatWeightProcess) reportOutputWeightsGrid(solution *highs.Solution, weightColumns map[stats.StatType]utilhighs.ColumnIndex, printer *util.PrintRecorder) map[stats.StatType]float64 {
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
