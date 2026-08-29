package weight_highs

import (
	"math"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_diffMax          = 2.0
	c_diffHighM        = 10.0
	c_outputPerInclude = -1
	c_statMax          = 1.0
	c_statHighM        = 10.0
	c_statEqualDelta   = c_statMax / 50000.0
	c_stdDevRange      = 1.0 // approximately 68.3% of all values fall within one standard deviation of the average, 95.4% within two standard deviations, and 99.7% within three.
)

type formulaSection3 struct {
	process *FormulaSegmentedProcess

	printer *util.PrintRecorder
	build   *util_highs.LinearBuilder

	objectiveEquationDiff util_highs.ObjectiveIndex
	objectiveInclude      util_highs.ObjectiveIndex

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

func (sect *formulaSection3) init(printer *util.PrintRecorder) {
	sect.printer = printer
}

func (sect *formulaSection3) setMinimumIncludeRate(percent float64) {
	sect.minimumIncludeRate = percent
}

func (sect *formulaSection3) run(timeout *util_highs.TimeLimitToken) (*util_async.FutureCancellable[weight_types.WeightResult2], error) {
	sect.build = new(util_highs.LinearBuilder)
	sect.build.Minimise = true
	sect.build.Solver = util_highs.Solver_MIP_Interior
	timeout.SetLinear(sect.build)

	sect.build.BlendMultiObjectives = true
	sect.objectiveEquationDiff = sect.build.AddObjectiveBlended(1, 0)
	sect.objectiveInclude = sect.build.AddObjectiveBlended(0.1, 0)

	sect.createWeightColumns()
	sect.createThresholds()
	sect.buildDataRows()

	sect.includeCountRow.Build(sect.build,
		float64(len(sect.process.inputData))*sect.minimumIncludeRate,
		float64(len(sect.process.inputData)))

	stopwatch := util.StopwatchMakeStopped()
	solutionFuture := sect.build.RunHighsFuture(stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) weight_types.WeightResult2 {
		solution, err := linearResult.GetSolutionAndSaveLog(sect.printer)
		if err != nil {
			return weight_types.WeightResult2MakeError(stopwatch.Elapsed(), err)
		}

		weight := sect.extractAndReportSolution(solution)
		return weight_types.WeightResult2Make(&weight, stopwatch.Elapsed(), solution.Status)
	})
}

func (sect *formulaSection3) createWeightColumns() {
	for _, statType := range sect.process.requiredStats {
		for _, simType := range sect.process.requiredSims {
			colDetailWeight := sect.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "WEIGHT " + statType.Name() + " " + simType.Name()})
			sect.detailedWeightColumns.Put(statType, simType, colDetailWeight)
		}
	}

	for _, simType := range sect.process.requiredSims {
		sect.offsetColumns.Put(simType, sect.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "OFFSET " + simType.Name()}))
	}
}

func (sect *formulaSection3) createThresholds() {
	for _, statType := range sect.process.requiredStats {
		vars := &thresholdVars{
			loColumn: sect.build.CreateColumnGeneral(highs.Continuous, 0, c_statMax, util_highs.DebugText(statType.Name()+"-lo")),
			hiColumn: sect.build.CreateColumnGeneral(highs.Continuous, 0, c_statMax, util_highs.DebugText(statType.Name()+"-hi")),
		}
		sect.thresholdCols.Put(statType, vars)
	}
}

func (sect *formulaSection3) buildDataRows() {
	for data := range util_collection.ForPointer(sect.process.inputData) {
		sect.buildConstraintsForInput(data)
	}
}

func (sect *formulaSection3) buildConstraintsForInput(data *weight_types.WeightInput) {
	includeColumn := sect.makeIncludeToggleColumn()
	sect.buildIncludeCondition(includeColumn, &data.TotalStat)

	for _, simType := range sect.process.requiredSims {
		sect.buildDataEquation(&data.TotalStat,
			data.SimResult.Get(simType), data.SimResult.GetStdDevOrZero(simType),
			simType, includeColumn)
	}
}

func (sect *formulaSection3) makeIncludeToggleColumn() util_highs.ColumnIndex {
	includeColumn := sect.build.CreateColumnBoolWithObjective(c_outputPerInclude, sect.objectiveInclude, util_highs.DebugString{Text: "include"})
	sect.includeCountRow.Add(includeColumn, 1)
	sect.includeColumns = append(sect.includeColumns, includeColumn)
	return includeColumn
}

func (sect *formulaSection3) buildIncludeCondition(includeColumn util_highs.ColumnIndex, stats *stats.StatBlock) {
	and := util_highs.ConstraintAndBuilder{}
	for _, statType := range sect.process.requiredStats {
		statValue := stats.GetFloat(statType)
		statScale := sect.process.scaleStats.GetOrPanic(statType)
		scaledStatValue := statValue * statScale

		vars := sect.thresholdCols.GetOrPanic(statType)

		isOverMinimum := sect.build.ConstantIsGreaterOrEqualColumn(vars.loColumn, scaledStatValue, c_statHighM, c_statEqualDelta)
		isUnderMaximum := sect.build.ConstantIsLessOrEqualColumn(vars.hiColumn, scaledStatValue, c_statHighM, c_statEqualDelta)
		and.AddInput(isOverMinimum)
		and.AddInput(isUnderMaximum)
	}
	and.SetOutput(includeColumn)
	and.Build(sect.build)
}

func (sect *formulaSection3) buildDataEquation(stats *stats.StatBlock, simValueAverage, simValueStdDev float64, simType stats.SimType, includeColumn util_highs.ColumnIndex) {
	inputVars := make([]util_highs.ColumnIndex, 0)
	inputCoefficients := make([]float64, 0)

	for _, statType := range sect.process.requiredStats {
		weightDetailCol := sect.detailedWeightColumns.GetOrPanic(statType, simType)

		statValue := stats.GetFloat(statType)
		statScale := sect.process.scaleStats.GetOrPanic(statType)
		scaledStatValue := statValue * statScale

		inputVars = append(inputVars, weightDetailCol)
		inputCoefficients = append(inputCoefficients, scaledStatValue)
	}

	deviationOutput := sect.build.CreateColumnWithObjective(highs.Continuous,
		0, c_diffMax,
		1, sect.objectiveEquationDiff,
		util_highs.DebugString{Text: "deviationOutput"})

	sharedOffsetCol := sect.offsetColumns.GetOrPanic(simType)
	inputVars = append(inputVars, sharedOffsetCol)
	inputCoefficients = append(inputCoefficients, 1)

	simScale := sect.process.scaleSims.GetOrPanic(simType)
	scaledSimValueAverage := simScale.Apply(simValueAverage)
	if !sect.process.UseStdDev {
		sect.build.AbsoluteValueFromSumSeveral_WithToggle(
			inputVars, inputCoefficients,
			scaledSimValueAverage,
			includeColumn,
			deviationOutput,
			c_diffHighM,
		)
	} else {
		scaledStdDev := math.Abs(simValueStdDev * simScale.Scale)
		lo := scaledSimValueAverage - scaledStdDev
		hi := scaledSimValueAverage + scaledStdDev
		sect.build.AbsoluteValueFromSumSeveral_ConstRange_WithToggle(
			inputVars, inputCoefficients,
			lo, hi,
			includeColumn,
			deviationOutput,
			c_diffHighM,
		)
	}
}

func (sect *formulaSection3) extractAndReportSolution(solution *highs.Solution) weight_types.Weight2Extended {
	sect.build.DebugPrintColumns(solution, sect.printer)

	sect.printer.Println("WEIGHTS")
	weightExtended := sect.extractDetailWeights(solution)

	sect.reportInclude(solution)

	return weightExtended
}

func (sect *formulaSection3) extractDetailWeights(solution *highs.Solution) weight_types.Weight2Extended {
	// extract and report on detail weights
	weightExtended := weight_types.Weight2Extended_Make(sect.process.requiredSims, sect.process.requiredStats)
	for entry := range sect.detailedWeightColumns.SeqKey1Key2ValueEntries() {
		statType := entry.Key1
		simType := entry.Key2
		column := entry.Value

		modelWeight := solution.ColValues[column]

		scaleStat := sect.process.scaleStats.GetOrPanic(statType)
		usableWeight := modelWeight * scaleStat

		weightExtended.PutWeight(simType, statType, usableWeight)

		sect.printer.Printf("%10s %10s %11.8f (%5.2e) %11.8f (%5.2e)\n", statType.Name(), simType.Name(), modelWeight, modelWeight, usableWeight, usableWeight)
	}
	sect.printer.Println0()

	for simType, offsetColumn := range sect.offsetColumns.SeqKeyValue() {
		offsetValue := solution.ColValues[offsetColumn]
		ratio := sect.process.targetRatios.GetOrPanic(simType)
		weightExtended.SetSimScale(simType, 1, offsetValue, ratio)
	}

	for entry := range weightExtended.SeqBySimThenStat() {
		usableWeight := entry.Value
		sect.printer.Printf("%10s %10s %11.8f (%5.2e)\n", entry.Key1.Name(), entry.Key2.Name(), usableWeight, usableWeight)
	}
	sect.printer.Println0()

	weightExtended.UpdateScaling(sect.process.inputData)
	weightExtended.FinishAndValidate(sect.process.inputData)
	return *weightExtended
}

func (sect *formulaSection3) reportInclude(solution *highs.Solution) {
	var includeCount uint32 = 0
	for _, col := range sect.includeColumns {
		if util.FloatEqualsOne(solution.ColValues[col]) {
			includeCount++
		}
	}
	includePercent := float64(includeCount) / float64(len(sect.process.inputData))
	sect.printer.Printf("Include %d %f\n", includeCount, includePercent)
}
