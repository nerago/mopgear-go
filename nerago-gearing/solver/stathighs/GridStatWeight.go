package stathighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

type GridStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimResultStats
	inputData    []WeightInput

	input           utilhighs.InputBuilder
	colNames        []string
	unitStatValues  util.MapMap[stats.StatType, simulate.SimResultType, []float64]
	detailedWeights util.MapMap[stats.StatType, simulate.SimResultType, utilhighs.ColumnIndex]
	finalWeights    map[stats.StatType]utilhighs.ColumnIndex
}

func (grid *GridStatWeightProcess) Init(printer *util.PrintRecorder) {
	grid.printer = printer
	grid.input.Minimise = true
	// grid.input.Solver = "ipm"
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

	// grid.simData.ForeachWithKeys(func(statType stats.StatType, statValue uint32, sim simulate.SimResultStats) {
	// 	grid.incorporateSample(statType, statValue, sim)
	// })
	grid.sweepData()
	grid.unitValuesToCalcDetailedRatings()
	grid.calcTotalRatings()

	solution, log := grid.input.RunHighs()
	grid.printer.AppendOther(log)
	grid.printer.Println(solution.Status.String())

	debugPrintColumnsGrid(solution, grid)

	return reportOutputWeightsGrid(solution, grid.finalWeights, grid.printer)
}

func (grid *GridStatWeightProcess) setupWeightVars() {
	for _, statType := range G_RequiredStats {
		colFinalWeight := grid.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, nil)
		// colFinalWeight := basic.input.CreateColumnGeneral(highs.Continuous, -c_finalWeightLimit, c_finalWeightLimit)
		grid.finalWeights[statType] = colFinalWeight
		grid.colNames = append(grid.colNames, "FINAL WEIGHT: "+statType.Name())
	}

	for _, statType := range G_RequiredStats {
		for _, simType := range G_RequiredSims {
			colDetailWeight := grid.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, nil)
			grid.detailedWeights.Put(statType, simType, colDetailWeight)
			grid.colNames = append(grid.colNames, "WEIGHT: "+statType.Name()+" "+simType.String())
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
func (grid *GridStatWeightProcess) sweepData() {
	for a := range grid.inputData {
		for b := range grid.inputData {
			statType, isGood := isGoodSample(&grid.inputData[a].TotalStat, &grid.inputData[b].TotalStat)
			if isGood {
				grid.incorporateSample(statType, &grid.inputData[a], &grid.inputData[b])
			}
		}
	}
}

func isGoodSample(blockHigh, blockLow *stats.StatBlock) (stats.StatType, bool) {
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

func (grid *GridStatWeightProcess) incorporateSample(statType stats.StatType, high, low *WeightInput) {
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

		grid.unitStatValues.Apply(statType, simType, func(oldValue []float64) []float64 { return append(oldValue, unitStatValue) })
	}
}

func (grid *GridStatWeightProcess) unitValuesToCalcDetailedRatings() {
	// FORMULA:
	// detailweight_dps_haste = unit_dps_haste / unit_dps_str * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = unit_dps_haste * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = detailweight_dps_str * unit_dps_haste
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste = 0
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste + offset = 0  (allow small offset to optimise on)

	grid.unitStatValues.ForeachGroupForKey2(func(simType simulate.SimResultType, lookupStat func(stats.StatType) []float64) {
		unitValueBaseArray := lookupStat(c_baseStatType)
		detailWeightBase := grid.detailedWeights.GetOrPanic(c_baseStatType, simType)
		for _, thisStatType := range G_RequiredStats {
			if thisStatType != c_baseStatType {
				thisUnitValueArray := lookupStat(thisStatType)
				thisDetailWeight := grid.detailedWeights.GetOrPanic(thisStatType, simType)

				// look at multiple input values of each unitstat value
				for _, unitValueBase := range unitValueBaseArray {
					for _, thisUnitValue := range thisUnitValueArray {
						grid.unitValuesToCalcDetailedRatings_single(unitValueBase, detailWeightBase, thisUnitValue, thisDetailWeight, simType, thisStatType)
					}
				}
			}
		}
	})
}

func (grid *GridStatWeightProcess) unitValuesToCalcDetailedRatings_single(unitValueBase float64, detailWeightBase utilhighs.ColumnIndex,
	thisUnitValue float64, thisdetailWeight utilhighs.ColumnIndex, simType simulate.SimResultType, statType stats.StatType) {

	offsetSigned := grid.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, nil)
	grid.colNames = append(grid.colNames, "OFFSET SIGNED "+simType.String()+" "+statType.Name())
	offsetAbs := grid.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, nil) // outputs for objective function
	grid.colNames = append(grid.colNames, "OFFSET ABS "+simType.String()+" "+statType.Name())
	utilhighs.AbsoluteValue2(&grid.input, offsetSigned, offsetAbs)

	// detailweight_dps_haste * unit_dps_base - detailweight_dps_base * unit_dps_haste + offset = 0
	weightRow := utilhighs.ConstraintRowBuild{}
	weightRow.Add(thisdetailWeight, unitValueBase)
	weightRow.Add(detailWeightBase, -thisUnitValue)
	weightRow.Add(offsetSigned, 1)
	weightRow.Finish(&grid.input, 0, 0)
}

func (grid *GridStatWeightProcess) calcTotalRatings() {
	for _, statType := range G_RequiredStats {
		statFinalRow := utilhighs.ConstraintRowBuild{}
		grid.detailedWeights.ForeachInnerWithKey1Value(statType, func(_ simulate.SimResultType, detailColumn utilhighs.ColumnIndex) {
			statFinalRow.Add(detailColumn, 1)
		})

		finalWeightColumn := grid.finalWeights[statType]
		statFinalRow.Add(finalWeightColumn, -1)
		statFinalRow.Finish(&grid.input, 0, 0)
	}
}

func debugPrintColumnsGrid(solution *highs.Solution, basic *GridStatWeightProcess) {
	if utilhighs.C_DebugHighs {
		for i, x := range solution.ColValues {
			basic.printer.Printf("%3d %14f %s\n", i, x, basic.colNames[i])
		}
	}
}

func reportOutputWeightsGrid(solution *highs.Solution, weightColumns map[stats.StatType]utilhighs.ColumnIndex, printer *util.PrintRecorder) map[stats.StatType]float64 {
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
