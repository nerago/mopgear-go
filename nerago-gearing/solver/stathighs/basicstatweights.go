package stathighs

import (
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"

	"github.com/bartolsthoorn/gohighs/highs"
)

type BasicStatWeightProcess struct {
	printer *util.PrintRecorder

	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	targetRatios  stats.SimData
	simBase       stats.SimData
	simData       map[stats.StatType]basicDataEntry

	build           utilhighs.LinearBuilder
	unitStatValues  util.MapMap[stats.StatType, stats.SimType, float64]
	detailedWeights util.MapMap[stats.StatType, stats.SimType, utilhighs.ColumnIndex]
	finalWeights    map[stats.StatType]utilhighs.ColumnIndex
}

type basicDataEntry struct {
	increment uint32
	sim       stats.SimData
}

func (basic *BasicStatWeightProcess) Init(printer *util.PrintRecorder) {
	basic.printer = printer
	basic.build.Minimise = true
	basic.build.Solver = utilhighs.Solver_LP_NO_GPU // needs sync and would be overkill
	basic.simData = make(map[stats.StatType]basicDataEntry)
	basic.finalWeights = make(map[stats.StatType]utilhighs.ColumnIndex)
}

func (basic *BasicStatWeightProcess) SetBaseline(simBase stats.SimData) {
	basic.simBase = simBase
}

func (basic *BasicStatWeightProcess) SetRequiredStats(requiredStats []stats.StatType) {
	basic.requiredStats = requiredStats
}

func (basic *BasicStatWeightProcess) SetTargetRatios(targetRatios stats.SimData) {
	basic.requiredSims = targetRatios.NonZeroTypes()

	sum := 0.0
	for _, simType := range basic.requiredSims {
		val := targetRatios.Get(simType)
		if val <= 0 {
			panic("missing ratio")
		}
		sum += val
	}
	if !util.FloatEqualsOne(sum) {
		panic("ratios don't add to one")
	}
	basic.targetRatios = targetRatios
}

func (basic *BasicStatWeightProcess) AddSimData(statType stats.StatType, statValue uint32, sim stats.SimData) {
	if _, hasValue := basic.simData[statType]; hasValue {
		panic("entry already set for stat type")
	}
	basic.simData[statType] = basicDataEntry{statValue, sim}
}

func (basic *BasicStatWeightProcess) Run(stopwatch *util.Stopwatch) *channel_op.FutureCancellable[WeightResult] {
	for _, statType := range basic.requiredStats {
		colName := "FINAL WEIGHT: " + statType.Name()
		colFinalWeight := basic.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: colName})
		// colFinalWeight := basic.input.CreateColumnGeneral(highs.Continuous, -c_finalWeightLimit, c_finalWeightLimit)
		basic.finalWeights[statType] = colFinalWeight
	}

	for _, statType := range basic.requiredStats {
		for _, simType := range basic.requiredSims {
			colName := "WEIGHT: " + statType.Name() + " " + simType.Name()
			colDetailWeight := basic.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: colName})
			basic.detailedWeights.Put(statType, simType, colDetailWeight)
		}
	}

	baseStat := basic.requiredStats[0]
	for _, simType := range basic.requiredSims {
		value := basic.targetRatios.Get(simType)
		colDetailWeight := basic.detailedWeights.GetOrPanic(baseStat, simType)
		strengthSetToRatio := utilhighs.ConstraintRow{}
		strengthSetToRatio.Add(colDetailWeight, 1)
		strengthSetToRatio.Build(&basic.build, value, value)
	}

	for statType, entry := range basic.simData {
		basic.incorporateSample(statType, entry.increment, entry.sim)
	}
	basic.unitValuesToCalcDetailedRatings()
	basic.calcTotalRatings()

	solutionFuture := basic.build.RunHighsFuture(stopwatch)

	return channel_op.FutureCancellable_MapValue(solutionFuture, func(linearResult utilhighs.LinearResult) (WeightResult, bool) {
		solution := linearResult.GetSolutionAndSaveLog(basic.printer)

		basic.printer.Println(solution.Status.String())
		basic.build.DebugPrintColumns(solution, basic.printer)

		return basic.reportOutputWeights(solution), true
	})
}

// this is a single diff value, ideally we want to push data in and average across multiple
func (basic *BasicStatWeightProcess) unitDiffValue(sim stats.SimData, simType stats.SimType, statValue uint32) float64 {
	simValueDiff := sim.Get(simType) - basic.simBase.Get(simType)
	simValueDiffPerStat := simValueDiff / float64(statValue)
	return simValueDiffPerStat
}

func (basic *BasicStatWeightProcess) incorporateSample(statType stats.StatType, statValue uint32, sim stats.SimData) {
	// basic approach (spreadsheet "build ratings miti_2")
	// unit_dps_haste = (this_dps[haste] - base_dps) / this_haste_value
	// detailweight_dps_haste = unit_dps_haste / unit_dps_str * detailweight_str

	for _, simType := range basic.requiredSims {
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

	baseStat := basic.requiredStats[0]
	for simType, lookupStat := range basic.unitStatValues.SeqGroupsKey2Lookup() {
		unitValueBase := lookupStat(baseStat)
		detailWeightBase := basic.detailedWeights.GetOrPanic(baseStat, simType)
		for _, thisStatType := range basic.requiredStats {
			if thisStatType != baseStat {
				thisUnitValue := lookupStat(thisStatType)
				thisDetailWeight := basic.detailedWeights.GetOrPanic(thisStatType, simType)

				basic.unitValuesToCalcDetailedRatings_single(unitValueBase, detailWeightBase, thisUnitValue, thisDetailWeight, simType, thisStatType)
			}
		}
	}
}

func (basic *BasicStatWeightProcess) unitValuesToCalcDetailedRatings_single(unitValueBase float64, detailWeightBase utilhighs.ColumnIndex,
	thisUnitValue float64, thisdetailWeight utilhighs.ColumnIndex, simType stats.SimType, statType stats.StatType) {

	colName := "OFFSET SIGNED " + simType.Name() + " " + statType.Name()
	offsetSigned := basic.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: colName})

	colName = "OFFSET ABS " + simType.Name() + " " + statType.Name()
	offsetAbs := basic.build.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: colName}) // outputs for objective function
	basic.build.AbsoluteValue(offsetSigned, offsetAbs)

	// detailweight_dps_haste * unit_dps_base - detailweight_dps_base * unit_dps_haste + offset = 0
	weightRow := utilhighs.ConstraintRow{}
	weightRow.Add(thisdetailWeight, unitValueBase)
	weightRow.Add(detailWeightBase, -thisUnitValue)
	weightRow.Add(offsetSigned, 1)
	weightRow.Build(&basic.build, 0, 0)
}

func (basic *BasicStatWeightProcess) calcTotalRatings() {
	for _, statType := range basic.requiredStats {
		statFinalRow := utilhighs.ConstraintRow{}
		for _, detailColumn := range basic.detailedWeights.SeqInnerWithKey1Value(statType) {
			statFinalRow.Add(detailColumn, 1)
		}

		finalWeightColumn := basic.finalWeights[statType]
		statFinalRow.Add(finalWeightColumn, -1)
		statFinalRow.Build(&basic.build, 0, 0)
	}
}

func (basic *BasicStatWeightProcess) reportOutputWeights(solution *highs.Solution) WeightResult {
	result := WeightResult_Make()
	basic.printer.Println("FINAL WEIGHTS:")
	for _, statType := range basic.requiredStats {
		columnIndex := basic.finalWeights[statType]
		value := solution.ColValues[columnIndex]
		basic.printer.Printf("%10s %f\n", statType.Name(), value)
		result.Put(statType, value)
	}
	return result
}
