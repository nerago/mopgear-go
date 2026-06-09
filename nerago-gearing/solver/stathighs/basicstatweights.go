package stathighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

type BasicStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimData
	simBase      simulate.SimData
	simData      map[stats.StatType]basicDataEntry

	input           utilhighs.InputBuilder
	unitStatValues  util.MapMap[stats.StatType, simulate.SimType, float64]
	detailedWeights util.MapMap[stats.StatType, simulate.SimType, utilhighs.ColumnIndex]
	finalWeights    map[stats.StatType]utilhighs.ColumnIndex
}

type basicDataEntry struct {
	increment uint32
	sim       simulate.SimData
}

func (basic *BasicStatWeightProcess) Init(printer *util.PrintRecorder) {
	basic.printer = printer
	basic.input.Minimise = true
	basic.input.Solver = "ipm"
	basic.simData = make(map[stats.StatType]basicDataEntry)
	basic.finalWeights = make(map[stats.StatType]utilhighs.ColumnIndex)
}

func (basic *BasicStatWeightProcess) SetBaseline(simBase simulate.SimData) {
	basic.simBase = simBase
}

func (basic *BasicStatWeightProcess) SetTargetRatios(targetRatios simulate.SimData) {
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

	basic.targetRatios = targetRatios
}

func (basic *BasicStatWeightProcess) AddSimData(statType stats.StatType, statValue uint32, sim simulate.SimData) {
	if _, hasValue := basic.simData[statType]; hasValue {
		panic("entry already set for stat type")
	}
	basic.simData[statType] = basicDataEntry{statValue, sim}
}

// alternately we could baseline each other with a full array of +100 perumtations etc

func (basic *BasicStatWeightProcess) Run() WeightResult {
	for _, statType := range G_RequiredStats {
		colName := "FINAL WEIGHT: " + statType.Name()
		colFinalWeight := basic.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: colName})
		// colFinalWeight := basic.input.CreateColumnGeneral(highs.Continuous, -c_finalWeightLimit, c_finalWeightLimit)
		basic.finalWeights[statType] = colFinalWeight
	}

	for _, statType := range G_RequiredStats {
		for _, simType := range G_RequiredSims {
			colName := "WEIGHT: " + statType.Name() + " " + simType.Name()
			colDetailWeight := basic.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: colName})
			basic.detailedWeights.Put(statType, simType, colDetailWeight)
		}
	}

	for _, simType := range G_RequiredSims {
		value := basic.targetRatios.Get(simType)
		colDetailWeight := basic.detailedWeights.GetOrPanic(c_baseStatType, simType)
		strengthSetToRatio := utilhighs.ConstraintRowBuild{}
		strengthSetToRatio.Add(colDetailWeight, 1)
		strengthSetToRatio.Finish(&basic.input, value, value)
	}

	for statType, entry := range basic.simData {
		basic.incorporateSample(statType, entry.increment, entry.sim)
	}
	basic.unitValuesToCalcDetailedRatings()
	basic.calcTotalRatings()

	solution, log := basic.input.RunHighs()
	basic.printer.AppendOther(log)
	basic.printer.Println(solution.Status.String())

	basic.input.DebugPrintColumns(solution, basic.printer)

	return reportOutputWeights(solution, basic.finalWeights, basic.printer)
}

// this is a single diff value, ideally we want to push data in and average across multiple
func (basic *BasicStatWeightProcess) unitDiffValue(sim simulate.SimData, simType simulate.SimType, statValue uint32) float64 {
	simValueDiff := sim.Get(simType) - basic.simBase.Get(simType)
	simValueDiffPerStat := simValueDiff / float64(statValue)
	return simValueDiffPerStat
}

func (basic *BasicStatWeightProcess) incorporateSample(statType stats.StatType, statValue uint32, sim simulate.SimData) {
	// basic approach (spreadsheet "build ratings miti_2")
	// unit_dps_haste = (this_dps[haste] - base_dps) / this_haste_value
	// detailweight_dps_haste = unit_dps_haste / unit_dps_str * detailweight_str

	for _, simType := range G_RequiredSims {
		unitStatValue := basic.unitDiffValue(sim, simType, statValue)
		if basic.unitStatValues.Has(statType, simType) {
			panic("value already set")
		}
		basic.unitStatValues.Put(statType, simType, unitStatValue)
	}
}

func (basic *BasicStatWeightProcess) unitValuesToCalcDetailedRatings() {
	// FORMULA:
	// detailweight_dps_haste = unit_dps_haste / unit_dps_str * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = unit_dps_haste * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = detailweight_dps_str * unit_dps_haste
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste = 0
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste + offset = 0  (allow small offset to optimise on)

	for simType, lookupStat := range basic.unitStatValues.SeqGroupsKey2Lookup() {
		unitValueBase := lookupStat(c_baseStatType)
		detailWeightBase := basic.detailedWeights.GetOrPanic(c_baseStatType, simType)
		for _, thisStatType := range G_RequiredStats {
			if thisStatType != c_baseStatType {
				thisUnitValue := lookupStat(thisStatType)
				thisDetailWeight := basic.detailedWeights.GetOrPanic(thisStatType, simType)

				basic.unitValuesToCalcDetailedRatings_single(unitValueBase, detailWeightBase, thisUnitValue, thisDetailWeight, simType, thisStatType)
			}
		}
	}
}

func (basic *BasicStatWeightProcess) unitValuesToCalcDetailedRatings_single(unitValueBase float64, detailWeightBase utilhighs.ColumnIndex,
	thisUnitValue float64, thisdetailWeight utilhighs.ColumnIndex, simType simulate.SimType, statType stats.StatType) {

	colName := "OFFSET SIGNED " + simType.Name() + " " + statType.Name()
	offsetSigned := basic.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: colName})

	colName = "OFFSET ABS " + simType.Name() + " " + statType.Name()
	offsetAbs := basic.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: colName}) // outputs for objective function
	utilhighs.AbsoluteValue(&basic.input, offsetSigned, offsetAbs)

	// detailweight_dps_haste * unit_dps_base - detailweight_dps_base * unit_dps_haste + offset = 0
	weightRow := utilhighs.ConstraintRowBuild{}
	weightRow.Add(thisdetailWeight, unitValueBase)
	weightRow.Add(detailWeightBase, -thisUnitValue)
	weightRow.Add(offsetSigned, 1)
	weightRow.Finish(&basic.input, 0, 0)
}

func (basic *BasicStatWeightProcess) calcTotalRatings() {
	for _, statType := range G_RequiredStats {
		statFinalRow := utilhighs.ConstraintRowBuild{}
		for _, detailColumn := range basic.detailedWeights.SeqInnerWithKey1Value(statType) {
			statFinalRow.Add(detailColumn, 1)
		}

		finalWeightColumn := basic.finalWeights[statType]
		statFinalRow.Add(finalWeightColumn, -1)
		statFinalRow.Finish(&basic.input, 0, 0)
	}
}

func reportOutputWeights(solution *highs.Solution, weightColumns map[stats.StatType]utilhighs.ColumnIndex, printer *util.PrintRecorder) WeightResult {
	result := WeightResult_Make()
	printer.Println("FINAL WEIGHTS:")
	for _, statType := range G_RequiredStats {
		columnIndex := weightColumns[statType]
		value := solution.ColValues[columnIndex]
		printer.Printf("%10s %f\n", statType.Name(), value)
		result.Put(statType, value)
	}
	return result
}

