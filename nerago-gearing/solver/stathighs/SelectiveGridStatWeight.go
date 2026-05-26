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

type SelectiveGridStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios            simulate.SimResultStats
	inputData               []WeightInput
	inputDataIncludeToggles []utilhighs.ColumnIndex

	input           utilhighs.InputBuilder
	unitStatValues  util.MapMapSlice[stats.StatType, simulate.SimResultType, selectiveGridDataSample]
	detailedWeights util.MapMap[stats.StatType, simulate.SimResultType, utilhighs.ColumnIndex]
	finalWeights    map[stats.StatType]utilhighs.ColumnIndex
}

type selectiveGridDataSample struct {
	value          float64
	includeToggles [2]utilhighs.ColumnIndex
}

func (selgrid *SelectiveGridStatWeightProcess) Init(printer *util.PrintRecorder) {
	selgrid.printer = printer
	selgrid.input.Minimise = true
	// grid.input.Solver = "ipm"
	selgrid.finalWeights = make(map[stats.StatType]utilhighs.ColumnIndex)
}

func (selgrid *SelectiveGridStatWeightProcess) SupplyData(inputData []WeightInput) {
	selgrid.inputData = inputData
}

func (selgrid *SelectiveGridStatWeightProcess) SetTargetRatios(targetRatios simulate.SimResultStats) {
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

	selgrid.targetRatios = targetRatios
}

func (selgrid *SelectiveGridStatWeightProcess) Run() map[stats.StatType]float64 {
	selgrid.setupWeightVars()

	selgrid.createIncludeToggles()
	selgrid.dataSamplesFromPairs()
	selgrid.checkSampleRange()
	selgrid.unitValuesToCalcDetailedRatings()
	selgrid.calcTotalRatings()

	solution, log := selgrid.input.RunHighs()
	selgrid.printer.AppendOther(log)
	selgrid.printer.Println(solution.Status.String())

	selgrid.input.DebugPrintColumns(solution, selgrid.printer)

	return selgrid.reportOutputWeightsGrid(solution, selgrid.finalWeights, selgrid.printer)
}

func (selgrid *SelectiveGridStatWeightProcess) setupWeightVars() {
	for _, statType := range G_RequiredStats {
		colFinalWeight := selgrid.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "FINAL WEIGHT: " + statType.Name()})
		// colFinalWeight := basic.input.CreateColumnGeneral(highs.Continuous, -c_finalWeightLimit, c_finalWeightLimit)
		selgrid.finalWeights[statType] = colFinalWeight
	}

	for _, statType := range G_RequiredStats {
		for _, simType := range G_RequiredSims {
			colDetailWeight := selgrid.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "WEIGHT: " + statType.Name() + " " + simType.String()})
			selgrid.detailedWeights.Put(statType, simType, colDetailWeight)
		}
	}

	for _, simType := range G_RequiredSims {
		value := selgrid.targetRatios.Get(simType)
		colDetailWeight := selgrid.detailedWeights.GetOrPanic(c_baseStatType, simType)
		strengthSetToRatio := utilhighs.ConstraintRowBuild{}
		strengthSetToRatio.Add(colDetailWeight, 1)
		strengthSetToRatio.Finish(&selgrid.input, value, value)
	}
}

func (selgrid *SelectiveGridStatWeightProcess) createIncludeToggles() {
	includeScore := 0.1

	rowIncludeReasonableNumber := utilhighs.ConstraintRowBuild{}

	selgrid.inputDataIncludeToggles = make([]utilhighs.ColumnIndex, len(selgrid.inputData))
	for i := range len(selgrid.inputData) {
		// negative score since we're minimising, each variable used is "better"
		column := selgrid.input.CreateColumnWithOutput(highs.Integer, 0, 1, -includeScore, utilhighs.DebugString{Text: "input toggle " + strconv.Itoa(i)})
		selgrid.inputDataIncludeToggles[i] = column

		rowIncludeReasonableNumber.Add(column, 1)
	}

	minimumUseful := len(selgrid.inputData) / 4
	rowIncludeReasonableNumber.Finish(&selgrid.input, float64(minimumUseful), float64(len(selgrid.inputData)))
}

// lazy func, could avoid double processing and put in order, very N**2
func (selgrid *SelectiveGridStatWeightProcess) dataSamplesFromPairs() {
	for a := range selgrid.inputData {
		for b := range selgrid.inputData {
			statType, isGood := selgrid.isGoodPair(&selgrid.inputData[a].TotalStat, &selgrid.inputData[b].TotalStat)
			if isGood {
				selgrid.prepareSample(statType, &selgrid.inputData[a], &selgrid.inputData[b], selgrid.inputDataIncludeToggles[a], selgrid.inputDataIncludeToggles[b])
			}
		}
	}
}

func (selgrid *SelectiveGridStatWeightProcess) isGoodPair(blockHigh, blockLow *stats.StatBlock) (stats.StatType, bool) {
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

func (selgrid *SelectiveGridStatWeightProcess) prepareSample(statType stats.StatType, high, low *WeightInput, highInclude, lowInclude utilhighs.ColumnIndex) {
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

		// if (unitStatValue > 0 && unitStatValue <= 1e-9) || (unitStatValue < 0 && unitStatValue >= -1e-9) {
		// 	panic("small values, highs won't like it")
		// }

		dataSample := selectiveGridDataSample{unitStatValue, [2]utilhighs.ColumnIndex{highInclude, lowInclude}}
		selgrid.unitStatValues.Add(statType, simType, dataSample)
	}
}

func (selgrid *SelectiveGridStatWeightProcess) checkSampleRange() {
	good, bad := 0, 0
	for entry := range selgrid.unitStatValues.SeqValues() {
		if isGoodValueRange(entry.value) {
			good++
		} else {
			bad++
		}
	}
	selgrid.printer.Printf("checkSampleRange good=%d bad=%d\n", good, bad)
	if bad > (good+bad)/10 {
		panic("many values have inconvenient range")
	}
}

func (selgrid *SelectiveGridStatWeightProcess) unitValuesToCalcDetailedRatings() {
	// FORMULA:
	// detailweight_dps_haste = unit_dps_haste / unit_dps_str * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = unit_dps_haste * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = detailweight_dps_str * unit_dps_haste
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste = 0
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste + offset = 0  (allow small offset to optimise on)

	for simType, lookupStat := range selgrid.unitStatValues.SeqGroupsKey2() {
		unitValueBaseSeq := lookupStat(c_baseStatType)
		detailWeightBase := selgrid.detailedWeights.GetOrPanic(c_baseStatType, simType)
		for _, thisStatType := range G_RequiredStats {
			if thisStatType != c_baseStatType {
				thisUnitValueSeq := lookupStat(thisStatType)
				selgrid.unitValuesCalcForGroup(simType, thisStatType, unitValueBaseSeq, thisUnitValueSeq, detailWeightBase)
			}
		}
	}
}

func (selgrid *SelectiveGridStatWeightProcess) unitValuesCalcForGroup(simType simulate.SimResultType, thisStatType stats.StatType, unitValueBaseSeq iter.Seq[selectiveGridDataSample], thisUnitValueSeq iter.Seq[selectiveGridDataSample], detailWeightBase utilhighs.ColumnIndex) {
	debugText := simType.String() + " " + thisStatType.Name()
	thisDetailWeight := selgrid.detailedWeights.GetOrPanic(thisStatType, simType)

	// look at multiple input values of each unitstat value
	index := 0
	for unitValueBase := range unitValueBaseSeq {
		for thisUnitValue := range thisUnitValueSeq {
			if isGoodValueRange(unitValueBase.value) && isGoodValueRange(thisUnitValue.value) {
				selgrid.unitValueCombinationAddToModel(unitValueBase, detailWeightBase, thisUnitValue, thisDetailWeight, debugText+" "+strconv.Itoa(index))
				index++
			}
		}
	}
}

func (selgrid *SelectiveGridStatWeightProcess) unitValueCombinationAddToModel(baseUnitSample selectiveGridDataSample, detailWeightBase utilhighs.ColumnIndex,
	thisUnitSample selectiveGridDataSample, thisDetailWeight utilhighs.ColumnIndex, debugText string) {

	includeDataPointToggle := selgrid.input.CreateColumnBool(utilhighs.DebugString{Text: "DATA TOGGLE " + debugText})
	and := utilhighs.ContraintAndBuilder{}
	and.SetOutput(includeDataPointToggle)
	and.AddInput(baseUnitSample.includeToggles[0])
	and.AddInput(baseUnitSample.includeToggles[1])
	and.AddInput(thisUnitSample.includeToggles[0])
	and.AddInput(thisUnitSample.includeToggles[1])
	and.FinishAndApply(&selgrid.input)

	// detailweight_dps_haste * unit_dps_base - detailweight_dps_base * unit_dps_haste + offset = 0
	// detailweight_dps_haste / unit_dps_haste  - detailweight_dps_base   / unit_dps_base + offset / unit_dps_base / unit_dps_haste = 0
	offsetSigned := selgrid.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "OFFSET SIGNED " + debugText})
	weightRow := utilhighs.ConstraintRowBuild{}
	weightRow.Add(thisDetailWeight, baseUnitSample.value)  // OLD
	weightRow.Add(detailWeightBase, -thisUnitSample.value) // OLD
	// weightRow.Add(thisDetailWeight, 1/thisUnitSample.value) // NEW BUT TOO BIG
	// weightRow.Add(detailWeightBase, -1/baseUnitSample.value) // NEW BUT TOO BIG
	weightRow.Add(offsetSigned, 1)
	weightRow.Finish(&selgrid.input, 0, 0)

	// take absolute value
	offsetAbs := selgrid.input.CreateColumnGeneral(highs.Continuous, 0, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "OFFSET ABS " + debugText})
	utilhighs.AbsoluteValue2(&selgrid.input, offsetSigned, offsetAbs)

	// output for objective function
	output := selgrid.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: "OUTPUT " + debugText})
	utilhighs.ContraintIfBoolCopyValueElseZero(&selgrid.input, includeDataPointToggle, offsetAbs, output, 0, c_finalWeightLimit) // don't know if this can work, especially with zero low
}

func (selgrid *SelectiveGridStatWeightProcess) calcTotalRatings() {
	for _, statType := range G_RequiredStats {
		statFinalRow := utilhighs.ConstraintRowBuild{}
		selgrid.detailedWeights.ForeachInnerWithKey1Value(statType, func(_ simulate.SimResultType, detailColumn utilhighs.ColumnIndex) {
			statFinalRow.Add(detailColumn, 1)
		})

		finalWeightColumn := selgrid.finalWeights[statType]
		statFinalRow.Add(finalWeightColumn, -1)
		statFinalRow.Finish(&selgrid.input, 0, 0)
	}
}

func (selgrid *SelectiveGridStatWeightProcess) reportOutputWeightsGrid(solution *highs.Solution, weightColumns map[stats.StatType]utilhighs.ColumnIndex, printer *util.PrintRecorder) map[stats.StatType]float64 {
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
