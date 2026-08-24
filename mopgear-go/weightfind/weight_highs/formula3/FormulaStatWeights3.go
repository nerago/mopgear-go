package weight_highs

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	util_weight2 "github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_formula3DiffMax          = 2.0
	c_formula3DiffHighM        = 10.0
	c_formula3OutputPerInclude = -1
	c_formula3StatMax          = 1.0
	c_formula3StatHighM        = 10.0
	c_formula3StatEqualDelta   = c_formula3StatMax / 50000.0
)

type FormulaStatWeightProcess3 struct {
	printer *util.PrintRecorder

	targetRatios  weight_types.SimPriorityBasic
	requiredStats []stats.StatType
	requiredSims  []stats.SimType

	inputData []weight_types.WeightInput

	build *util_highs.LinearBuilder

	objectiveEquationDiff util_highs.ObjectiveIndex
	objectiveInclude      util_highs.ObjectiveIndex

	scaleSims             stats.SimTypeMap[util_weight.ScaleAndOffset]
	scaleStats            stats.StatTypeMap[float64]
	detailedWeightColumns util_collection.MapMap[stats.StatType, stats.SimType, util_highs.ColumnIndex]
	offsetColumns         stats.SimTypeMap[util_highs.ColumnIndex]

	minimumIncludeRate float64
	includeColumns     []util_highs.ColumnIndex
	includeCountRow    util_highs.ConstraintRow

	thresholdCols stats.StatTypeMap[*thresholdVars]
}

type thresholdVars struct {
	loColumn util_highs.ColumnIndex
	hiColumn util_highs.ColumnIndex
	// hiloDiff row, just minimum/or slack/or target
	// diff to top/bottom vars?
	// number of stat types with cutoffs?

}

func (form *FormulaStatWeightProcess3) Init(printer *util.PrintRecorder) {
	form.printer = printer
}

func (form *FormulaStatWeightProcess3) SupplyData(inputData []weight_types.WeightInput) {
	form.inputData = inputData
}

func (form *FormulaStatWeightProcess3) SetRequiredStats(requiredStats []stats.StatType) {
	form.requiredStats = requiredStats
}

func (form *FormulaStatWeightProcess3) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	form.targetRatios = targetRatios
	form.requiredSims = targetRatios.SimTypes()
}

func (form *FormulaStatWeightProcess3) SetMinimumIncludeRate(percent float64) {
	form.minimumIncludeRate = percent
}

func (form *FormulaStatWeightProcess3) Run(timeout int) *util_async.FutureCancellable[weight_types.WeightResult] {
	form.build = new(util_highs.LinearBuilder)
	form.build.Minimise = true
	form.build.Solver = util_highs.Solver_MIP_Interior
	form.build.TimeLimitSeconds = timeout

	form.build.BlendMultiObjectives = true
	form.objectiveEquationDiff = form.build.AddObjectiveBlended(1, 0)
	form.objectiveInclude = form.build.AddObjectiveBlended(0.1, 0)

	form.chooseScaling()
	form.createWeightColumns()
	form.createThresholds()
	form.buildDataRows()

	form.includeCountRow.Build(form.build, float64(len(form.inputData))*form.minimumIncludeRate, util_highs.InfPos())

	stopwatch := util.StopwatchMakeStopped()
	solutionFuture := form.build.RunHighsFuture(stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) (weight_types.WeightResult, bool) {
		solution := linearResult.GetSolutionAndSaveLog(form.printer)
		weight := form.extractAndReportSolution(solution)
		return weight_types.WeightResult{Weight: &weight, SolveTime: stopwatch.Elapsed(), Status: solution.Status}, true
	})
}

func (form *FormulaStatWeightProcess3) chooseScaling() {
	target := c_formula3StatMax
	form.scaleStats = util_weight2.ChooseStatScalingBasic(form.inputData, target, true, form.printer)
	form.scaleSims = util_weight2.ChooseSimUnfriendlyUnitScaleAndOffset(form.inputData, form.requiredSims)
}

func (form *FormulaStatWeightProcess3) createWeightColumns() {
	for _, statType := range form.requiredStats {
		for _, simType := range form.requiredSims {
			colDetailWeight := form.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "WEIGHT " + statType.Name() + " " + simType.Name()})
			form.detailedWeightColumns.Put(statType, simType, colDetailWeight)
		}
	}

	for _, simType := range form.requiredSims {
		form.offsetColumns.Put(simType, form.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "OFFSET " + simType.Name()}))
	}
}

func (form *FormulaStatWeightProcess3) createThresholds() {
	for _, statType := range form.requiredStats {
		vars := &thresholdVars{
			loColumn: form.build.CreateColumnGeneral(highs.Continuous, 0, c_formula3StatMax, util_highs.DebugText(statType.Name()+"-lo")),
			hiColumn: form.build.CreateColumnGeneral(highs.Continuous, 0, c_formula3StatMax, util_highs.DebugText(statType.Name()+"-hi")),
		}
		form.thresholdCols.Put(statType, vars)
	}
}

func (form *FormulaStatWeightProcess3) buildDataRows() {
	for data := range util_collection.ForPointer(form.inputData) {
		form.buildDataRowsForInput(data)
	}
}

func (form *FormulaStatWeightProcess3) buildDataRowsForInput(data *weight_types.WeightInput) {
	includeColumn := form.makeIncludeToggleColumn()
	form.buildIncludeCondition(includeColumn, &data.TotalStat)

	for _, simType := range form.requiredSims {
		form.buildDataEquationForInputSim(&data.TotalStat, data.SimResult.Get(simType), simType, includeColumn)
	}
}

func (form *FormulaStatWeightProcess3) makeIncludeToggleColumn() util_highs.ColumnIndex {
	includeColumn := form.build.CreateColumnBoolWithObjective(c_formula3OutputPerInclude, form.objectiveInclude, util_highs.DebugString{Text: "include"})
	form.includeCountRow.Add(includeColumn, 1)
	form.includeColumns = append(form.includeColumns, includeColumn)
	return includeColumn
}

func (form *FormulaStatWeightProcess3) buildIncludeCondition(includeColumn util_highs.ColumnIndex, stats *stats.StatBlock) {
	and := util_highs.ConstraintAndBuilder{}
	for _, statType := range form.requiredStats {
		statValue := stats.GetFloat(statType)
		statScale := form.scaleStats.GetOrPanic(statType)
		scaledStatValue := statValue * statScale

		vars := form.thresholdCols.GetOrPanic(statType)

		isOverMinimum := form.build.ColumnIsLessOrEqualThanConstant(vars.loColumn, scaledStatValue, c_formula3StatHighM, c_formula3StatEqualDelta)
		isUnderMaximum := form.build.ColumnIsGreaterOrEqualThanConstant(vars.hiColumn, scaledStatValue, c_formula3StatHighM, c_formula3StatEqualDelta)
		and.AddInput(isOverMinimum)
		and.AddInput(isUnderMaximum)
	}
	and.SetOutput(includeColumn)
	and.Build(form.build)
}

func (form *FormulaStatWeightProcess3) buildDataEquationForInputSim(stats *stats.StatBlock, simValue float64, simType stats.SimType, includeColumn util_highs.ColumnIndex) {
	//matchSimValue := util_highs.ConstraintRow{}
	inputVars := make([]util_highs.ColumnIndex, 0)
	inputCoefficients := make([]float64, 0)

	for _, statType := range form.requiredStats {
		weightDetailCol := form.detailedWeightColumns.GetOrPanic(statType, simType)

		statValue := stats.GetFloat(statType)
		statScale := form.scaleStats.GetOrPanic(statType)
		scaledStatValue := statValue * statScale

		inputVars = append(inputVars, weightDetailCol)
		inputCoefficients = append(inputCoefficients, scaledStatValue)
	}

	deviationOutput := form.build.CreateColumnWithObjective(highs.Continuous,
		0, c_formula3DiffMax,
		1, form.objectiveEquationDiff,
		util_highs.DebugString{Text: "deviationAbsOutput"})

	sharedOffsetCol := form.offsetColumns.GetOrPanic(simType)
	inputVars = append(inputVars, sharedOffsetCol)
	inputCoefficients = append(inputCoefficients, 1)

	simScale := form.scaleSims.GetOrPanic(simType)
	scaledSimValue := simScale.Apply(simValue)

	form.build.AbsoluteValueFromSumSeveral_WithToggle(
		inputVars, inputCoefficients,
		scaledSimValue,
		includeColumn,
		deviationOutput,
		c_formula3DiffHighM,
	)
}

func (form *FormulaStatWeightProcess3) extractAndReportSolution(solution *highs.Solution) weight_types.Weight2Extended {
	form.build.DebugPrintColumns(solution, form.printer)

	form.printer.Println("WEIGHTS")
	weightExtended := form.extractDetailWeights(solution)

	form.reportInclude(solution)

	return weightExtended
}

func (form *FormulaStatWeightProcess3) extractDetailWeights(solution *highs.Solution) weight_types.Weight2Extended {
	// extract and report on detail weights
	weightExtended := weight_types.Weight2Extended_Make(form.requiredSims, form.requiredStats)
	for entry := range form.detailedWeightColumns.SeqKey1Key2ValueEntries() {
		statType := entry.Key1
		simType := entry.Key2
		column := entry.Value

		modelWeight := solution.ColValues[column]

		scaleStat := form.scaleStats.GetOrPanic(statType)
		usableWeight := modelWeight * scaleStat

		weightExtended.PutWeight(simType, statType, usableWeight)

		form.printer.Printf("%10s %10s %11.8f (%5.2e) %11.8f (%5.2e)\n", statType.Name(), simType.Name(), modelWeight, modelWeight, usableWeight, usableWeight)
	}
	form.printer.Println0()

	for simType, offsetColumn := range form.offsetColumns.SeqKeyValue() {
		//simScale := form.scaleSims[simType]
		offsetValue := solution.ColValues[offsetColumn]
		ratio := form.targetRatios.GetOrPanic(simType)
		weightExtended.SetSimScale(simType, 1, offsetValue, ratio)
	}

	for entry := range weightExtended.SeqBySimThenStat() {
		usableWeight := entry.Value
		form.printer.Printf("%10s %10s %11.8f (%5.2e)\n", entry.Key1.Name(), entry.Key2.Name(), usableWeight, usableWeight)
	}
	form.printer.Println0()

	weightExtended.FinishAndValidate()
	return *weightExtended
}

func (form *FormulaStatWeightProcess3) reportInclude(solution *highs.Solution) {
	var includeCount uint32 = 0
	for _, col := range form.includeColumns {
		if util.FloatEqualsOne(solution.ColValues[col]) {
			includeCount++
		}
	}
	includePercent := float64(includeCount) / float64(len(form.inputData))
	form.printer.Printf("Include %d %f\n", includeCount, includePercent)
}
