package fitting1

import (
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_highs"
	util_weight2 "github.com/nerago/mopgear-go/weightfind/util_weight"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_fitting_each_threadCount = 8

	c_fitting_statScaledRangeHigh    = 1.0
	c_fitting_simScaledRangeHigh     = 1.0
	c_fitting_statUnscaledHigh       = 50000
	c_fitting_statScaledUnequalDelta = c_fitting_statScaledRangeHigh / float64(c_fitting_statUnscaledHigh)

	c_fitting_outputDifference              = 1
	c_fitting_outputFittingPerInclude       = -1
	c_fitting_objectiveSlackForFullCoverage = 0.8

	c_fitting_minimum_stat_coverage       = 100
	c_fitting_permitted_overlap_fix       = 5
	c_fitting_number_nice_number_interval = 5

	c_fitting_initial_range_required    = 0.35
	c_fitting_large_range_threshold     = 0.50
	c_fitting_large_range_required      = 0.20
	c_fitting_medium_range_threshold    = 0.25
	c_fitting_medium_range_required     = 0.10
	c_fitting_small_range_threshold     = 0.15
	c_fitting_small_range_required      = 0.10
	c_fitting_tiny_range_threshold      = 0.05
	c_fitting_tiny_range_over_required  = 0.80
	c_fitting_tiny_range_under_required = 0.95
	c_fitting_dropped_range_threshold   = 0.02
)

// so we want to define a line of best fit for each stat/sim
// but also only for certain ranges of each stat, others excluded

// question is do we work each stat separately, even though ranges may not line up
// yes, we can always compose the individual parts later

// but do we allow ranges on the other stats we're not checking
// we might end up very fragmented and noisy if we do

// do we consider simType separately? why not - some aren't correlated, going to hard enough to reconcile across one dimension
//                                    why - some are highly correlated

// we need to consider that we're developing a function that correlates to the sim output, not predicts it in any summation sense
// maybe suggests start with the individual ones, less tempting to try to hit totals

type FittingSingleStatWeightProcess struct {
	printer   *util.PrintRecorder
	stopwatch util.Stopwatch
	build     *util_highs.LinearBuilder

	minimumIncludeRate float64
	inputData          []util_weight2.FittingSample

	objectiveLineDiff util_highs.ObjectiveIndex
	objectiveInclude  util_highs.ObjectiveIndex

	lineSlope        util_highs.ColumnIndex
	lineOffset       util_highs.ColumnIndex
	minimumThreshold util_highs.ColumnIndex
	maximumThreshold util_highs.ColumnIndex

	includeColumns  []util_highs.ColumnIndex
	includeCountRow util_highs.ConstraintRow
}

type FittingSingleStatResult struct {
	LineSlope                  float64
	LineOffset                 float64
	Minimum                    float64
	Maximum                    float64
	IncludeCount               uint32
	IncludePercentOfStageInput float64
	BuiltSequence              int
	StopwatchSolver            util.Stopwatch
}

func (fw *FittingSingleStatWeightProcess) Init(printer *util.PrintRecorder, timeout int) {
	fw.printer = printer
	fw.build = new(util_highs.LinearBuilder)
	fw.build.Minimise = true
	fw.build.Solver = util_highs.Solver_MIP_Interior
	fw.build.TimeLimitSeconds = timeout

	fw.lineSlope = fw.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "slope"})
	fw.lineOffset = fw.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "offset"})
	fw.minimumThreshold = fw.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting_statScaledRangeHigh, util_highs.DebugString{Text: "minimum"})
	fw.maximumThreshold = fw.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting_statScaledRangeHigh, util_highs.DebugString{Text: "maximum"})

	fw.build.ColumnIsGreaterOrEqualColumnEnforce(fw.minimumThreshold, fw.maximumThreshold)
}

func (fw *FittingSingleStatWeightProcess) SetMinimumIncludeRate(percent float64) {
	fw.minimumIncludeRate = percent
}

func (fw *FittingSingleStatWeightProcess) SupplySamples(inputData []util_weight2.FittingSample) {
	fw.inputData = inputData
}

func (fw *FittingSingleStatWeightProcess) Run() *util_async.FutureCancellable[FittingSingleStatResult] {
	fw.setupLinearObjectives()

	for _, sample := range fw.inputData {
		fw.addSample(sample)
	}

	fw.includeCountRow.Build(fw.build, float64(len(fw.inputData))*fw.minimumIncludeRate, util_highs.InfPos())

	solutionFuture := fw.build.RunHighsFuture(&fw.stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) (FittingSingleStatResult, bool) {
		solution := linearResult.GetSolution2AndSaveLog(fw.printer)
		solution.DebugPrint(fw.printer)
		if solution.Status() == highs.ModelStatusOptimal {
			return fw.buildResult(solution), true
		} else {
			return FittingSingleStatResult{}, false
		}
	})
}

func (fw *FittingSingleStatWeightProcess) setupLinearObjectives0() {
	fw.build.BlendMultiObjectives = true

	fw.objectiveLineDiff = fw.build.AddObjectiveBlended(5, 0)
	fw.objectiveInclude = fw.build.AddObjectiveBlended(1, 0)
}

func (fw *FittingSingleStatWeightProcess) setupLinearObjectives() {
	fw.build.BlendMultiObjectives = false

	var relativeToleranceParam float64
	if fw.minimumIncludeRate < 1 {
		// first linear step find a regular solution to the line fit
		// will probably follow the minimum required include
		// will get us a positive initial result from the sum of differenceAbs
		// let it expand to full coverage if it wants, but without worsening the average difference
		multiplierToFullCoverage := 1 / fw.minimumIncludeRate
		// add a bit of factor to this, only 80%, otherwise might get too greedy
		multiplierToFullCoverage *= c_fitting_objectiveSlackForFullCoverage
		// highs logic is "objective * (1.0 + linear_objective.rel_tolerance)", so need to minus one in compensation
		// don't let it go negative or below a small value
		relativeToleranceParam = max(multiplierToFullCoverage-1, 0.1)
	} else {
		relativeToleranceParam = 0
	}
	fw.objectiveLineDiff = fw.build.AddObjectivePrioritised(false, -1, relativeToleranceParam, 2)

	// second priority is sum of includeColumn which are negative one each, can lead to negative total objective
	// but we don't need to care about offsets much since it's the last one, highs shouldn't even look at them
	fw.objectiveInclude = fw.build.AddObjectivePrioritised(false, -1, -1, 1)
}

func (fw *FittingSingleStatWeightProcess) buildResult(solution *util_highs.Solution2) FittingSingleStatResult {
	result := FittingSingleStatResult{}
	result.LineSlope = solution.GetValue(fw.lineSlope)
	result.LineOffset = solution.GetValue(fw.lineOffset)
	result.Minimum = solution.GetValue(fw.minimumThreshold)
	result.Maximum = solution.GetValue(fw.maximumThreshold)

	var includeCount uint32 = 0
	for _, col := range fw.includeColumns {
		if solution.ValueIsOne(col) {
			includeCount++
		}
	}
	result.IncludeCount = includeCount
	result.IncludePercentOfStageInput = float64(includeCount) / float64(len(fw.inputData))

	if includeCount == 0 {
		panic("shouldn't this have failed in model")
	}

	result.StopwatchSolver = fw.stopwatch
	return result
}

func (fw *FittingSingleStatWeightProcess) addSample(sample util_weight2.FittingSample) {
	if sample.SimResult < 0 || sample.SimResult > 1 || sample.StatValue < 0 || sample.StatValue > 1 {
		panic("sample out of range")
	}

	includeColumn := fw.sampleIncludeToggleColumn(sample)
	fw.sampleToFitLine(sample, includeColumn)
}

func (fw *FittingSingleStatWeightProcess) sampleIncludeToggleColumn(sample util_weight2.FittingSample) util_highs.ColumnIndex {
	includeColumn := fw.build.CreateColumnBoolWithObjective(c_fitting_outputFittingPerInclude, fw.objectiveInclude, util_highs.DebugString{Text: "include"})
	fw.includeCountRow.Add(includeColumn, 1)
	fw.includeColumns = append(fw.includeColumns, includeColumn)

	fw.build.ConstantIsBetweenColumns_NoSequenceCheck(fw.minimumThreshold, fw.maximumThreshold, includeColumn, sample.StatValue, c_fitting_statScaledRangeHigh, c_fitting_statScaledUnequalDelta)

	return includeColumn
}

// Want lineSlope to look like sim/stat
// basic line formula:       y = lineSlope * x + lineOffset
//
//	            y - lineOffset = lineSlope * x
//	        y/x - lineOffset/x = lineSlope
//	sim/stat - lineOffset/stat = lineSlope
//	                  sim/stat = lineSlope + lineOffset/stat
//	                       sim = lineSlope*stat + lineOffset
func (fw *FittingSingleStatWeightProcess) sampleToFitLine(sample util_weight2.FittingSample, toggle util_highs.ColumnIndex) {
	difference := fw.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "difference"})
	differenceAbs := fw.build.CreateColumnWithObjective(highs.Continuous, 0, util_highs.InfPos(), c_fitting_outputDifference, fw.objectiveLineDiff, util_highs.DebugString{Text: "differenceAbs"})

	sampleRow := util_highs.ConstraintRow{Debug: "sampleRow"}
	sampleRow.Add(fw.lineSlope, sample.StatValue)
	sampleRow.Add(fw.lineOffset, 1)
	sampleRow.Add(difference, 1) // this is vertical difference, not true minimum distance. would be proportional within similar slope ranges only
	sampleRow.Build(fw.build, sample.SimResult, sample.SimResult)

	fw.build.AbsoluteValue_WithToggle_NoExtraCheck(difference, differenceAbs, toggle, c_fitting_simScaledRangeHigh)
}

// what would be the ideal form
// distance would be perpendicular to line right?
// that distance is not necessarily proportional to the vertical distance, which could be inflated for a steep line.
// thus existing algorithm favors shallower lines

// we have coords x1,y1 line is y=mx+c
// if m=2, rises from 0,0 to 1,2 tan says 63degrees, with another 90 degrees, m=-0.5, rise=-1/run=2.
// so basically perpendicular line is mb=-1/m
// then we'd have to compute an intercept, a bit more to that
