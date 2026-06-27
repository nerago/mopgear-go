package stathighs

import (
	"iter"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"strconv"
	"strings"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_selGrid_finalWeightLimit = 50.0
	c_selGrid_formulaHigh      = 1000
)

// we use unscaled sims and stats so
// stat = 1000..50000.
// sim =  0..500000.
// statDiff be exactly 250/500
// simDiff for dps 50000, but bigger is possible
// simDiff for tmi about 10
// simDiff for death about 1
// unitStatValue := simValueDiff / statDiff
// unitValue = approx 100 for dps, 0.004 for death
// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste + offset = 0
// formula always within same simtype
// so scale of the formula: detailweight_dps_haste * unit_dps_str: 50 * 100 = 5000
//                          detailweight_dps_str * unit_dps_haste: 0.4 * 100 = 40

// changing to scaled sim and stat
// stat, sim = 0..1
// statDiff be around 0.01
// simDiff be around 0.1
// unitValue = approx 10
// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste + offset = 0
// formula always within same simtype
// so scale of the formula: detailweight_dps_haste * unit_dps_str: 50 * 10 = 500
//                          detailweight_dps_str * unit_dps_haste: 0.4 * 10 = 4

type SelectiveGridStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios            stats.SimData
	requiredStats           []stats.StatType
	requiredSims            []stats.SimType
	inputData               []WeightInput
	inputDataIncludeToggles []utilhighs.ColumnIndex

	scaleStat map[stats.StatType]float64
	scaleSim  map[stats.SimType]float64

	build           *utilhighs.LinearBuilder
	unitStatValues  util.MapMapSlice[stats.StatType, stats.SimType, selectiveGridDataSample]
	detailedWeights util.MapMap[stats.StatType, stats.SimType, utilhighs.ColumnIndex]
}

type selectiveGridDataSample struct {
	value          float64
	includeToggles [2]utilhighs.ColumnIndex
}

func (selgrid *SelectiveGridStatWeightProcess) Init(printer *util.PrintRecorder) {
	selgrid.printer = printer
	selgrid.build = new(utilhighs.LinearBuilder)
	selgrid.build.Minimise = true
	selgrid.build.Solver = utilhighs.Solver_MIP_Interior
	// selgrid.build.DisablePreSolve = true
}

func (selgrid *SelectiveGridStatWeightProcess) SupplyData(inputData []WeightInput) {
	selgrid.scaleStat = chooseStatScaling(inputData, selgrid.printer)
	selgrid.scaleSim = chooseSimScalingUnfriendly(inputData, selgrid.printer)
	selgrid.inputData = inputData
}

func (selgrid *SelectiveGridStatWeightProcess) SetRequiredStats(requiredStats []stats.StatType) {
	selgrid.requiredStats = requiredStats
}

func (selgrid *SelectiveGridStatWeightProcess) SetTargetRatios(targetRatios stats.SimData) {
	selgrid.requiredSims = targetRatios.NonZeroTypes()

	sum := 0.0
	for _, simType := range selgrid.requiredSims {
		val := targetRatios.Get(simType)
		if val <= 0 {
			panic("missing ratio")
		}
		sum += val
	}
	if !util.FloatEqualsOne(sum) {
		panic("ratios don't add to one")
	}

	selgrid.targetRatios = targetRatios
}

func (selgrid *SelectiveGridStatWeightProcess) Run() WeightResult {
	selgrid.setupWeightVars()

	selgrid.createIncludeToggles()
	selgrid.dataSamplesFromPairs()
	selgrid.unitValuesToCalcDetailedRatings()

	solution := selgrid.build.RunHighs(selgrid.printer)
	selgrid.printer.Println(solution.Status.String())

	selgrid.build.DebugPrintColumns(solution, selgrid.printer)

	return selgrid.reportOutputWeightsGrid(solution)
}

func (selgrid *SelectiveGridStatWeightProcess) setupWeightVars() {
	for _, statType := range selgrid.requiredStats {
		for _, simType := range selgrid.requiredSims {
			colDetailWeight := selgrid.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "WEIGHT: " + statType.Name() + " " + simType.Name()})
			selgrid.detailedWeights.Put(statType, simType, colDetailWeight)
		}
	}

	baseStat := selgrid.requiredStats[0]
	for _, simType := range selgrid.requiredSims {
		colDetailWeight := selgrid.detailedWeights.GetOrPanic(baseStat, simType)
		strengthSetToRatio := utilhighs.ConstraintRow{}
		strengthSetToRatio.Add(colDetailWeight, 1)
		strengthSetToRatio.Build(selgrid.build, 0.1, 100)
	}
}

func (selgrid *SelectiveGridStatWeightProcess) createIncludeToggles() {
	// negative score since we're minimising, each variable used is "better"
	includeScore := -10.0

	rowIncludeReasonableNumber := utilhighs.ConstraintRow{}

	selgrid.inputDataIncludeToggles = make([]utilhighs.ColumnIndex, len(selgrid.inputData))
	for i := range len(selgrid.inputData) {
		column := selgrid.build.CreateColumnBoolWithOutput(includeScore, utilhighs.DebugString{Text: "input toggle " + strconv.Itoa(i)})
		selgrid.inputDataIncludeToggles[i] = column

		rowIncludeReasonableNumber.Add(column, 1)
	}

	minimumUseful := len(selgrid.inputData) / 4
	rowIncludeReasonableNumber.Build(selgrid.build, float64(minimumUseful), float64(len(selgrid.inputData)))
}

// lazy func, could avoid double processing and put in order, very N**2
func (selgrid *SelectiveGridStatWeightProcess) dataSamplesFromPairs() {
	for a := range selgrid.inputData {
		for b := a + 1; b < len(selgrid.inputData); b++ {
			diffCount, statType := selgrid.checkForNumberStatDifferences(&selgrid.inputData[a].TotalStat, &selgrid.inputData[b].TotalStat)
			if diffCount == 1 {
				selgrid.prepareSample(statType, &selgrid.inputData[a], &selgrid.inputData[b], selgrid.inputDataIncludeToggles[a], selgrid.inputDataIncludeToggles[b])
			}
		}
	}
}

func (selgrid *SelectiveGridStatWeightProcess) checkForNumberStatDifferences(one, two *stats.StatBlock) (differenceCount int, diffStatA stats.StatType) {
	// for stat := range one { // was doing fine in tests up until now, up to 89%
	for _, stat := range selgrid.requiredStats {
		if one[stat] != two[stat] {
			if differenceCount == 0 {
				diffStatA = stats.StatType(stat)
			}
			differenceCount++
		}
	}
	return differenceCount, diffStatA
}

func (selgrid *SelectiveGridStatWeightProcess) prepareSample(statType stats.StatType, high, low *WeightInput, highInclude, lowInclude utilhighs.ColumnIndex) {
	// basic approach (spreadsheet "build ratings miti_2")
	// unit_dps_haste = (this_dps[haste] - base_dps) / this_haste_value
	// detailweight_dps_haste = unit_dps_haste / unit_dps_str * detailweight_str

	statDiff := high.TotalStat.GetFloat(statType) - low.TotalStat.GetFloat(statType)
	statDiff *= selgrid.scaleStat[statType]

	for _, simType := range selgrid.requiredSims {
		simValueDiff := high.SimResult.Get(simType) - low.SimResult.Get(simType)
		simValueDiff *= selgrid.scaleSim[simType]
		if !simType.IsHighGood() {
			simValueDiff *= -1
		}
		unitStatValue := simValueDiff / statDiff

		dataSample := selectiveGridDataSample{unitStatValue, [2]utilhighs.ColumnIndex{highInclude, lowInclude}}
		selgrid.unitStatValues.Add(statType, simType, dataSample)
	}
}

func (selgrid *SelectiveGridStatWeightProcess) unitValuesToCalcDetailedRatings() {
	// FORMULA:
	// detailweight_dps_haste = unit_dps_haste / unit_dps_str * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = unit_dps_haste * detailweight_dps_str
	// detailweight_dps_haste * unit_dps_str = detailweight_dps_str * unit_dps_haste
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste = 0
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste + offset = 0  (allow small offset to optimise on)

	baseStat := selgrid.requiredStats[0]
	for simType, lookupStat := range selgrid.unitStatValues.SeqGroupsKey2Lookup() {
		unitValueBaseSeq := lookupStat(baseStat)
		for _, thisStatType := range selgrid.requiredStats {
			if thisStatType != baseStat {
				thisUnitValueSeq := lookupStat(thisStatType)
				selgrid.unitValuesCalcForGroup(simType, thisStatType, unitValueBaseSeq, thisUnitValueSeq)
			}
		}
	}
}

func (selgrid *SelectiveGridStatWeightProcess) unitValuesCalcForGroup(simType stats.SimType, thisStatType stats.StatType, unitValueBaseSeq iter.Seq[selectiveGridDataSample], thisUnitValueSeq iter.Seq[selectiveGridDataSample]) {
	debugText := simType.Name() + " " + thisStatType.Name()

	baseStat := selgrid.requiredStats[0]
	baseDetailWeight := selgrid.detailedWeights.GetOrPanic(baseStat, simType)
	thisDetailWeight := selgrid.detailedWeights.GetOrPanic(thisStatType, simType)

	// look at multiple input values of each unitstat value
	index := 0
	for unitValueBase := range unitValueBaseSeq {
		for thisUnitValue := range thisUnitValueSeq {
			// TODO does this reject unevenly by type
			if isGoodValueRange(unitValueBase.value) && isGoodValueRange(thisUnitValue.value) {
				selgrid.unitValueCombinationAddToModel(unitValueBase, baseDetailWeight, thisUnitValue, thisDetailWeight, debugText+" "+strconv.Itoa(index))
				index++
			}
		}
	}
}

func (selgrid *SelectiveGridStatWeightProcess) unitValueCombinationAddToModel(baseUnitSample selectiveGridDataSample, detailWeightBase utilhighs.ColumnIndex,
	thisUnitSample selectiveGridDataSample, thisDetailWeight utilhighs.ColumnIndex, debugText string) {

	includeDataPointToggle := selgrid.build.CreateColumnBool(utilhighs.DebugString{Text: "DATA TOGGLE " + debugText})
	and := utilhighs.ConstraintAndBuilder{}
	and.SetOutput(includeDataPointToggle)
	and.AddInput(baseUnitSample.includeToggles[0])
	and.AddInput(baseUnitSample.includeToggles[1])
	and.AddInput(thisUnitSample.includeToggles[0])
	and.AddInput(thisUnitSample.includeToggles[1])
	and.Build(selgrid.build)

	// detailweight_dps_haste * unit_dps_base - detailweight_dps_base * unit_dps_haste + offset = 0
	// detailweight_dps_haste / unit_dps_haste  - detailweight_dps_base   / unit_dps_base + offset / unit_dps_base / unit_dps_haste = 0
	offsetSigned := selgrid.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "OFFSET SIGNED " + debugText})
	weightRow := utilhighs.ConstraintRow{}
	weightRow.Add(thisDetailWeight, baseUnitSample.value)
	weightRow.Add(detailWeightBase, -thisUnitSample.value)
	weightRow.Add(offsetSigned, 1)
	weightRow.Build(selgrid.build, 0, 0)

	// take absolute value
	offsetAbs := selgrid.build.CreateColumnWithOutput(highs.Continuous, 0, c_selGrid_formulaHigh, 1, utilhighs.DebugString{Text: "OFFSET ABS " + debugText})
	selgrid.build.AbsoluteValue_WithToggle(offsetSigned, offsetAbs, includeDataPointToggle, c_selGrid_formulaHigh*2)
}

func (selgrid *SelectiveGridStatWeightProcess) reportOutputWeightsGrid(solution *highs.Solution) WeightResult {
	selgrid.printer.Println("FINAL WEIGHTS:")

	// weights are mutally scaled within simtype
	// detailweight_dps_haste * unit_dps_str - detailweight_dps_str * unit_dps_haste + offset = 0
	// unitValue := (DPS_Diff * DPS_SCALE) / (HASTE_DIFF * HASTE_SCALE)
	// detailweight_dps_haste * (DPS_Diff * DPS_SCALE) / (STR_DIFF * STR_SCALE) = detailweight_dps_str * (DPS_Diff * DPS_SCALE) / (HASTE_DIFF * HASTE_SCALE)
	// detailweight_dps_haste * (DPS_Diff * DPS_SCALE) / (STR_DIFF * STR_SCALE) = detailweight_dps_str * (DPS_Diff * DPS_SCALE) / (HASTE_DIFF * HASTE_SCALE)
	// detailweight_dps_haste / (STR_DIFF * STR_SCALE) = detailweight_dps_str / (HASTE_DIFF * HASTE_SCALE)
	// detailweight_dps_haste / detailweight_dps_str = (STR_DIFF * STR_SCALE) / (HASTE_DIFF * HASTE_SCALE)
	// detailweight_dps_haste / detailweight_dps_str = (STR_DIFF / HASTE_DIFF) * (STR_SCALE / HASTE_SCALE)
	// detailweight_dps_haste / detailweight_dps_str * (HASTE_SCALE / STR_SCALE) = (STR_DIFF / HASTE_DIFF)

	weightValues := util.MapMap_FromExitingMapMap_WithApplyPlusKeys(&selgrid.detailedWeights,
		func(statType stats.StatType, _ stats.SimType, weightCol utilhighs.ColumnIndex) float64 {
			scale := selgrid.scaleStat[statType]
			return solution.ColValues[weightCol] * scale
		})

	baseStat := selgrid.requiredStats[0]
	for simType := range weightValues.SeqKey2() {
		strengthValue := weightValues.GetOrPanic(baseStat, simType)
		for statType := range weightValues.SeqKey1() {
			weightValues.Apply(statType, simType, func(oldValue float64) float64 {
				value := oldValue / strengthValue * selgrid.targetRatios.Get(simType)
				// if !simType.IsHighGood() {
				// 	value *= -1
				// }
				return value
			})
		}
	}

	finalWeightResult := WeightResult_Make()
	for statType, valueSeq := range weightValues.SeqGroupsKey1NestedKeyValue() {
		total := 0.0
		for _, value := range valueSeq {
			total += value
		}
		finalWeightResult.Put(statType, total)
	}

	included := 0
	for _, colInclude := range selgrid.inputDataIncludeToggles {
		if util.FloatEqualsOne(solution.ColValues[colInclude]) {
			included++
		}
	}
	selgrid.printer.Printf("SELGRID include rate %d / %d = %f\n", included, len(selgrid.inputData), float64(included)/float64(len(selgrid.inputData)))

	dataToggleTotal, dataToggleOn := 0, 0
	for colIndex := range solution.ColValues {
		debug := selgrid.build.DebugTextFor(utilhighs.ColumnIndex(colIndex))
		if strings.Contains(debug, "DATA TOGGLE") {
			if util.FloatEqualsOne(solution.ColValues[colIndex]) {
				dataToggleOn++
			}
			dataToggleTotal++
		}
	}
	selgrid.printer.Printf("SELGRID pair include rate %d / %d = %f\n", dataToggleOn, dataToggleTotal, float64(dataToggleOn)/float64(dataToggleTotal))

	return finalWeightResult
}
