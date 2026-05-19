package stathighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_rangeHigh        = 100.0
	c_baseStatType     = stats.Stat_Strength
	c_finalWeightLimit = 50
	c_offsetLimit      = 0.1
)

type BasicStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimResultStats
	simBase      simulate.SimResultStats
	simData      util.MapMap[stats.StatType, uint32, simulate.SimResultStats]

	input           utilhighs.InputBuilder
	colNames        []string
	unitStatValues  util.MapMap[stats.StatType, simulate.SimResultType, []float64]
	detailedWeights util.MapMap[stats.StatType, simulate.SimResultType, utilhighs.ColumnIndex]
	finalWeights    map[stats.StatType]utilhighs.ColumnIndex
}

func (basic *BasicStatWeightProcess) Init(printer *util.PrintRecorder) {
	basic.printer = printer
	basic.input.Minimise = true
	basic.input.Solver = "ipm"
	basic.finalWeights = make(map[stats.StatType]utilhighs.ColumnIndex)
}

func (basic *BasicStatWeightProcess) SetBaseline(simBase simulate.SimResultStats) {
	basic.simBase = simBase
}

func (basic *BasicStatWeightProcess) SetTargetRatios(targetRatios simulate.SimResultStats) {
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

func (basic *BasicStatWeightProcess) AddSimData(statType stats.StatType, statValue uint32, sim simulate.SimResultStats) {
	basic.simData.Put(statType, statValue, sim)
}

// alternately we could baseline each other with a full array of +100 perumtations etc

func (basic *BasicStatWeightProcess) Run() {
	for _, statType := range G_RequiredStats {
		colFinalWeight := basic.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
		// colFinalWeight := basic.input.CreateColumnGeneral(highs.Continuous, -c_finalWeightLimit, c_finalWeightLimit)
		basic.finalWeights[statType] = colFinalWeight
		basic.colNames = append(basic.colNames, "FINAL WEIGHT: "+statType.Name())
	}

	for _, statType := range G_RequiredStats {
		for _, simType := range G_RequiredSims {
			colDetailWeight := basic.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
			basic.detailedWeights.Put(statType, simType, colDetailWeight)
			basic.colNames = append(basic.colNames, "WEIGHT: "+statType.Name()+" "+simType.String())
		}
	}

	for _, simType := range G_RequiredSims {
		value := basic.targetRatios.Get(simType)
		colDetailWeight := basic.detailedWeights.GetOrPanic(c_baseStatType, simType)
		strengthSetToRatio := utilhighs.ConstraintRowBuild{}
		strengthSetToRatio.Add(colDetailWeight, 1)
		strengthSetToRatio.Finish(&basic.input, value, value)
	}

	basic.simData.ForeachWithKeys(func(statType stats.StatType, statValue uint32, sim simulate.SimResultStats) {
		basic.incorporateSample(statType, statValue, sim)
	})
	basic.unitValuesToCalcDetailedRatings()
	basic.calcTotalRatings()

	solution, log := basic.input.RunHighs()
	basic.printer.AppendOther(log)
	basic.printer.Println(solution.Status.String())

	debugPrintColumns(solution, basic)

	reportOutputWeights(solution, basic.finalWeights, basic.printer)
}

// this is a single diff value, ideally we want to push data in and average across multiple
func (basic *BasicStatWeightProcess) unitDiffValue(sim simulate.SimResultStats, simType simulate.SimResultType, statValue uint32) float64 {
	simValueDiff := sim.Get(simType) - basic.simBase.Get(simType)
	simValueDiffPerStat := simValueDiff / float64(statValue)
	return simValueDiffPerStat
}

func (basic *BasicStatWeightProcess) incorporateSample(statType stats.StatType, statValue uint32, sim simulate.SimResultStats) {
	// basic approach (spreadsheet "build ratings miti_2")
	// unit_dps_haste = (this_dps[haste] - base_dps) / this_haste_value
	// detailweight_dps_haste = unit_dps_haste / unit_dps_str * detailweight_str

	for _, simType := range G_RequiredSims {
		// TODO validateIsRelevantBase(basic.simBase // can't valiidate since we don't actually have full stat blocks, just this one difference value
		unitStatValue := basic.unitDiffValue(sim, simType, statValue)
		basic.unitStatValues.Apply(statType, simType, func(oldValue []float64) []float64 { return append(oldValue, unitStatValue) })
	}
}

func (basic *BasicStatWeightProcess) unitValuesToCalcDetailedRatings() {
	// FORMULA:
	// detailweight_dps_haste = unit_dps_haste / unit_dps_str * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = unit_dps_haste * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = detailweight_dps_str * unit_dps_haste
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste = 0
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste + offset = 0  (allow small offset to optimise on)

	basic.unitStatValues.ForeachGroupForKey2(func(simType simulate.SimResultType, lookupStat func(stats.StatType) []float64) {
		unitValueBaseArray := lookupStat(c_baseStatType)
		detailWeightBase := basic.detailedWeights.GetOrPanic(c_baseStatType, simType)
		for _, thisStatType := range G_RequiredStats {
			if thisStatType != c_baseStatType {
				thisUnitValueArray := lookupStat(thisStatType)
				thisDetailWeight := basic.detailedWeights.GetOrPanic(thisStatType, simType)

				// look at multiple input values of each unitstat value
				for _, unitValueBase := range unitValueBaseArray {
					for _, thisUnitValue := range thisUnitValueArray {
						basic.unitValuesToCalcDetailedRatings_single(unitValueBase, detailWeightBase, thisUnitValue, thisDetailWeight, simType, thisStatType)
					}
				}
			}
		}
	})
}

func (basic *BasicStatWeightProcess) unitValuesToCalcDetailedRatings_single(unitValueBase float64, detailWeightBase utilhighs.ColumnIndex,
	thisUnitValue float64, thisdetailWeight utilhighs.ColumnIndex, simType simulate.SimResultType, statType stats.StatType) {

	offsetSigned := basic.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
	basic.colNames = append(basic.colNames, "OFFSET SIGNED "+simType.String()+" "+statType.Name())
	offsetAbs := basic.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1) // outputs for objective function
	basic.colNames = append(basic.colNames, "OFFSET ABS "+simType.String()+" "+statType.Name())
	utilhighs.AbsoluteValue2(&basic.input, offsetSigned, offsetAbs, c_rangeHigh)

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
		basic.detailedWeights.ForeachInnerWithKey1Value(statType, func(_ simulate.SimResultType, detailColumn utilhighs.ColumnIndex) {
			statFinalRow.Add(detailColumn, 1)
		})

		finalWeightColumn := basic.finalWeights[statType]
		statFinalRow.Add(finalWeightColumn, -1)
		statFinalRow.Finish(&basic.input, 0, 0)
	}
}

func debugPrintColumns(solution *highs.Solution, basic *BasicStatWeightProcess) {
	if utilhighs.C_DebugHighs {
		for i, x := range solution.ColValues {
			basic.printer.Printf("%3d %14f %s\n", i, x, basic.colNames[i])
		}
	}
}

func reportOutputWeights(solution *highs.Solution, weightColumns map[stats.StatType]utilhighs.ColumnIndex, printer *util.PrintRecorder) {
	printer.Println("FINAL WEIGHTS:")
	for _, statType := range G_RequiredStats {
		columnIndex := weightColumns[statType]
		value := solution.ColValues[columnIndex]
		printer.Printf("%10s %f\n", statType.Name(), value)
	}
}
