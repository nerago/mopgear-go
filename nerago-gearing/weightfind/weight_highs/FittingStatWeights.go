package weight_highs

import (
	"cmp"
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/util/util_rank"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (

	// STR example: 30554
	c_statRangeHigh = 50000

	c_scaleBigSim = 1000
	// TPS example:  1957667
	c_simRangeHigh = 5000000 / c_scaleBigSim // we try to scale values like above towards nicer range

	c_outputFittingDifference = 1
	c_outputFittingPerInclude = -1
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

type FittingEachStatWeightProcess struct {
	printer *util.PrintRecorder
	timeout int

	lazyMode      bool
	inputData     []weight_types.WeightInput
	targetRatios  stats.SimData
	requiredStats []stats.StatType
	requiredSims  []stats.SimType

	each util.MapMap[stats.StatType, stats.SimType, *fittingEachFields]
}

type fittingEachFields struct {
	statType  stats.StatType
	simType   stats.SimType
	process   FittingSingleStatSegmentsProcess
	resultMap map[weight_types.StatRange]FittingSingleStatResult
}

func (fiteach *FittingEachStatWeightProcess) Init(printer *util.PrintRecorder, timeout int) {
	fiteach.printer = printer
	fiteach.timeout = timeout
}

func (fiteach *FittingEachStatWeightProcess) SetRequiredStats(requiredStats []stats.StatType) {
	fiteach.requiredStats = requiredStats
}

func (fiteach *FittingEachStatWeightProcess) SetTargetRatios(targetRatios stats.SimData) {
	fiteach.targetRatios = targetRatios
	fiteach.requiredSims = targetRatios.NonZeroTypes()
}

func (fiteach *FittingEachStatWeightProcess) SetLazyMode(lazy bool) {
	fiteach.lazyMode = lazy
}

func (fiteach *FittingEachStatWeightProcess) SupplyDataFromStandard(inputData []weight_types.WeightInput) {
	fiteach.inputData = inputData
}

func (fiteach *FittingEachStatWeightProcess) RunDetailedResults(cancel util_async.CancelSignal) weight_types.Weight3ExtendedRanged {
outer:
	for _, statType := range fiteach.requiredStats {
		for _, simType := range fiteach.requiredSims {
			// TODO holding printer?
			fields := fittingEachFields{statType: statType, simType: simType}
			fields.process.Init(fiteach.printer, statType, simType, fiteach.timeout)
			fields.process.SetLazyMode(fiteach.lazyMode)
			fields.process.SupplyDataFromStandard(fiteach.inputData)
			fiteach.each.Put(statType, simType, &fields)

			if cancel.ShouldFinish() {
				break outer
			}
		}
	}

	channelEach := util_async.SeqToChannel_Cancellable(fiteach.each.SeqValues(), cancel)
	util_async.ForEach_Channel(10, channelEach, func(fields *fittingEachFields) {
		fields.resultMap = fields.process.Run(cancel)
	})

	weights := weight_types.Weight3ExtendedRanged_Make()
	fiteach.each.ForeachWithKeys(func(statType stats.StatType, simType stats.SimType, value *fittingEachFields) {
		for statRange, detail := range value.resultMap {
			weights.Add(simType, statType, statRange, detail.LineSlope, detail.LineOffset, detail.IncludePercent)
		}
	})
	return weights
}

func (fiteach *FittingEachStatWeightProcess) Run(stopwatch *util.Stopwatch, cancel util_async.CancelSignal) weight_types.Weight1Basic {
	detailResult := fiteach.RunDetailedResults(cancel)

	for fields := range fiteach.each.SeqValues() {
		for _, detail := range fields.resultMap {
			stopwatch.AddElapsedFrom(&detail.StopwatchSolver)
		}
	}

	bestRatingEach := util.MapMap_FromExitingMapMap_WithApply(&detailResult, func(byRange map[weight_types.StatRange]FittingSingleStatResult) float64 {
		best := util_rank.BestCollector1[FittingSingleStatResult]{}
		for _, entry := range byRange {
			best.Offer(&entry, float64(entry.IncludeCount))
		}
		bestEntry := best.GetBestOptional()
		if bestEntry.HasValue() {
			return bestEntry.GetOrPanic().LineSlope
		} else {
			return 1
		}
	})

	baseStat := fiteach.requiredStats[0]
	standardResult := weight_types.Weight1Basic_Make()
	standardResult.Put(baseStat, 1)
	for _, statType := range fiteach.requiredStats {
		if statType != baseStat {
			totalSum := 0.0
			for _, simType := range fiteach.requiredSims {
				thisRating := bestRatingEach.GetOrPanic(simType, statType)
				strengthRating := bestRatingEach.GetOrPanic(baseStat, simType)
				relative := thisRating / strengthRating * fiteach.targetRatios.Get(simType)
				totalSum += relative
			}
			standardResult.Put(statType, totalSum)
		}
	}
	return standardResult
}

////////////////////////////////////////////////////////

type FittingSingleStatSegmentsProcess struct {
	printer *util.PrintRecorder

	timeout  int
	lazyMode bool

	inputDataOriginal       []*weight_types.WeightInput
	inputDataRemainingParts map[weight_types.StatRange][]*weight_types.WeightInput
	stat                    stats.StatType
	sim                     stats.SimType

	segments map[weight_types.StatRange]FittingSingleStatResult
}

func (fitseg *FittingSingleStatSegmentsProcess) Init(printer *util.PrintRecorder, stat stats.StatType, sim stats.SimType, timeout int) {
	fitseg.printer = printer
	fitseg.segments = make(map[weight_types.StatRange]FittingSingleStatResult)
	fitseg.inputDataRemainingParts = make(map[weight_types.StatRange][]*weight_types.WeightInput)
	fitseg.stat = stat
	fitseg.sim = sim
	fitseg.timeout = timeout
}

func (fitseg *FittingSingleStatSegmentsProcess) SetLazyMode(lazy bool) {
	fitseg.lazyMode = lazy
}

func (fitseg *FittingSingleStatSegmentsProcess) SupplyDataFromStandard(inputData []weight_types.WeightInput) {
	fitseg.inputDataOriginal = util.MapSliceAsNew(inputData, func(w *weight_types.WeightInput) *weight_types.WeightInput { return w })
}

func (fitseg *FittingSingleStatSegmentsProcess) Run(cancel util_async.CancelSignal) map[weight_types.StatRange]FittingSingleStatResult {
	// fitseg.runFitAll()

	fitseg.runInitial(cancel)
	if fitseg.lazyMode {
		return fitseg.segments
	}

	overallSize := len(fitseg.inputDataOriginal)
	for len(fitseg.inputDataRemainingParts) > 0 {
		nextRange, nextData := util.MapFirstEntry(fitseg.inputDataRemainingParts)
		delete(fitseg.inputDataRemainingParts, nextRange)

		ratioOfOverall := percentRatio(len(nextData), overallSize)
		if ratioOfOverall < 0.02 || len(nextData) < 3 {
			// drop it
		} else if ratioOfOverall < 0.05 || len(nextData) < 8 {
			fitseg.runNextSegment(nextData, nextRange, 1, cancel)
		} else if ratioOfOverall < 0.15 || len(nextData) < 20 {
			fitseg.runNextSegment(nextData, nextRange, 0.8, cancel)
		} else if ratioOfOverall < 0.30 {
			fitseg.runNextSegment(nextData, nextRange, 0.4, cancel)
		} else {
			fitseg.runNextSegment(nextData, nextRange, 0.2, cancel)
		}
	}

	return fitseg.segments
}

func percentRatio(value, total int) float64 {
	return float64(value) / float64(total)
}

func (fitseg *FittingSingleStatSegmentsProcess) runFitAll() {
	fit := FittingSingleStatWeightProcess{}
	fit.Init(fitseg.printer, fitseg.timeout)
	fit.SetMinimumIncludeRate(1)
	fit.SupplyDataFromStandard(fitseg.inputDataOriginal, fitseg.stat, fitseg.sim)
	weightOptionalFuture := fit.Run()
	weightOptional := weightOptionalFuture.WaitForResultAsOptional()
	if weight, hasWeight := weightOptional.GetWithFlag(); hasWeight {
		statRange := weight_types.StatRange{weight.Minimum, weight.Maximum}
		fitseg.segments[statRange] = weight
	}
}

func (fitseg *FittingSingleStatSegmentsProcess) runInitial(cancel util_async.CancelSignal) {
	fit := FittingSingleStatWeightProcess{}
	fit.Init(fitseg.printer, fitseg.timeout)
	fit.SetMinimumIncludeRate(0.3)
	fit.SupplyDataFromStandard(fitseg.inputDataOriginal, fitseg.stat, fitseg.sim)
	weightOptionalFuture := fit.Run()
	util_async.ChainCancel(cancel, weightOptionalFuture)
	weightOptional := weightOptionalFuture.WaitForResultAsOptional()
	if weight, hasWeight := weightOptional.GetWithFlag(); hasWeight {
		statRange := weight_types.StatRange{weight.Minimum, weight.Maximum}
		fitseg.segments[statRange] = weight

		totalRange := weight_types.StatRange{0, c_statRangeHigh}
		fitseg.addToRemainingData(fitseg.inputDataOriginal, totalRange, statRange)
	}
}

func (fitseg *FittingSingleStatSegmentsProcess) runNextSegment(inputData []*weight_types.WeightInput, inputRange weight_types.StatRange, includeRate float64, cancel util_async.CancelSignal) {
	fit := FittingSingleStatWeightProcess{}
	fit.Init(fitseg.printer, fitseg.timeout)
	fit.SetMinimumIncludeRate(includeRate)
	fit.SupplyDataFromStandard(inputData, fitseg.stat, fitseg.sim)
	weightOptionalFuture := fit.Run()
	util_async.ChainCancel(cancel, weightOptionalFuture)
	weightOptional := weightOptionalFuture.WaitForResultAsOptional()
	if weight, hasWeight := weightOptional.GetWithFlag(); hasWeight {
		minimum := max(inputRange.Minimum, weight.Minimum)
		maximum := min(inputRange.Maximum, weight.Maximum)

		statRange := weight_types.StatRange{Minimum: minimum, Maximum: maximum}
		weight.Minimum = minimum
		weight.Maximum = maximum
		fitseg.segments[statRange] = weight

		fitseg.addToRemainingData(inputData, inputRange, statRange)
	}
}

func (fitseg *FittingSingleStatSegmentsProcess) addToRemainingData(processedData []*weight_types.WeightInput, inputRange weight_types.StatRange, removeRange weight_types.StatRange) {
	if removeRange.Minimum < inputRange.Minimum || removeRange.Maximum > inputRange.Maximum || removeRange.Minimum > removeRange.Maximum || inputRange.Minimum > inputRange.Maximum {
		panic("range isn't within bounds")
	}

	loData := make([]*weight_types.WeightInput, 0)
	hiData := make([]*weight_types.WeightInput, 0)
	for _, input := range processedData {
		stat := input.TotalStat.GetUInt(fitseg.stat)
		if stat < inputRange.Minimum {
			panic("sample isn't within bounds")
		} else if inputRange.Minimum <= stat && stat < removeRange.Minimum {
			loData = append(loData, input)
		} else if removeRange.Minimum <= stat && stat <= removeRange.Maximum {
			// drop
		} else if removeRange.Maximum < stat && stat <= inputRange.Maximum {
			hiData = append(hiData, input)
		} else {
			panic("sample isn't within bounds")
		}
	}

	if len(loData) > 0 {
		loRange := weight_types.StatRange{Minimum: inputRange.Minimum, Maximum: removeRange.Minimum - 1}
		fitseg.inputDataRemainingParts[loRange] = loData
	}

	if len(hiData) > 0 {
		hiRange := weight_types.StatRange{Minimum: removeRange.Maximum + 1, Maximum: inputRange.Maximum}
		fitseg.inputDataRemainingParts[hiRange] = hiData
	}
}

func (fitseg *FittingSingleStatSegmentsProcess) filterDataStatRange(inputData []weight_types.WeightInput, lo, hi uint32) {
	util.FilterSliceAsNew(inputData, func(in *weight_types.WeightInput) bool {
		value := in.TotalStat.GetUInt(fitseg.stat)
		return lo <= value && value <= hi
	})
}

////////////////////////////////////////////////////////

type FittingSingleStatWeightProcess struct {
	printer   *util.PrintRecorder
	stopwatch util.Stopwatch
	build     *util_highs.LinearBuilder

	minimumIncludeRate float64
	inputData          []fittingSample
	inputDataSimScale  float64

	objectiveLineDiff util_highs.ObjectiveIndex
	objectiveInclude  util_highs.ObjectiveIndex

	lineSlope        util_highs.ColumnIndex
	lineOffset       util_highs.ColumnIndex
	minimumThreshold util_highs.ColumnIndex
	maximumThreshold util_highs.ColumnIndex
	includeColumns   []util_highs.ColumnIndex

	includeCountRow util_highs.ConstraintRow
}

type FittingSingleStatResult struct {
	LineSlope       float64
	LineOffset      float64
	Minimum         uint32
	Maximum         uint32
	IncludeCount    uint32
	IncludePercent  float64
	StopwatchSolver util.Stopwatch
}

type fittingSample struct {
	statValue     float64
	simResult     float64
	includeColumn util_highs.ColumnIndex
}

func (fit *FittingSingleStatWeightProcess) Init(printer *util.PrintRecorder, timeout int) {
	fit.printer = printer
	fit.build = new(util_highs.LinearBuilder)
	fit.build.Minimise = true
	fit.build.Solver = util_highs.Solver_MIP_Interior
	fit.build.TimeLimitSeconds = timeout

	fit.lineSlope = fit.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugString{Text: "slope"})
	fit.lineOffset = fit.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugString{Text: "offset"})
	fit.minimumThreshold = fit.build.CreateColumnGeneral(highs.Continuous, 0, c_statRangeHigh, util_highs.DebugString{Text: "minimum"})
	fit.maximumThreshold = fit.build.CreateColumnGeneral(highs.Continuous, 0, c_statRangeHigh, util_highs.DebugString{Text: "maximum"})

	maxVsMin := util_highs.ConstraintRow{}
	maxVsMin.Add(fit.minimumThreshold, -1)
	maxVsMin.Add(fit.maximumThreshold, 1)
	maxVsMin.Build(fit.build, 0, util_highs.C_PlusInf)
}

func (fit *FittingSingleStatWeightProcess) SetMinimumIncludeRate(percent float64) {
	fit.minimumIncludeRate = percent
}

func (fit *FittingSingleStatWeightProcess) SupplyDataFromStandard(inputData []*weight_types.WeightInput, stat stats.StatType, sim stats.SimType) {
	fit.inputData = util.MapSliceAsNew(inputData, func(input **weight_types.WeightInput) fittingSample {
		return fittingSample{
			(*input).TotalStat.GetFloat(stat),
			scaleSimItem((*input).SimResult.Get(sim), sim),
			-1,
		}
	})
	fit.inputDataSimScale = scaleSimItem(1, sim)
}

func scaleSimItem(value float64, sim stats.SimType) float64 {
	// example values 1671858.348 10396269.605 117613.197 217148.877 180.467 21.1
	// with scaleBig  1671.858    10396.269    117.613197 217.148877
	switch sim {
	case stats.Sim_DPS, stats.Sim_TPS, stats.Sim_DTPS, stats.Sim_HPS:
		return value / c_scaleBigSim
	case stats.Sim_TMI:
		return value
	case stats.Sim_DEATH:
		return value * 100
	default:
		panic("unknown type")
	}
}

/*

  if (this->options_.blend_multi_objectives) {
    HighsLinearObjective& multi_linear_objective = this->multi_linear_objective_[iObj];
    lp.col_cost_[iCol] += multi_linear_objective.weight * multi_linear_objective.coefficients[iCol];
    lp.sense_ = ObjSense::kMinimize;

  ELSE PRIORITIES

  for (HighsInt iObj = 0; iObj < num_linear_objective; iObj++)
    priority_objective.push_back(std::make_pair(this->multi_linear_objective_[iObj].priority, iObj));
  std::sort(priority_objective.begin(), priority_objective.end(), comparison);

  for (HighsInt iIx = 0; iIx < num_linear_objective; iIx++) {
    lp.col_cost_ = linear_objective.coefficients;
	lp.sense_ = linear_objective.weight > 0 ? ObjSense::kMinimize : ObjSense::kMaximize;
	use previous round's solution as start, needs to be valid for this too
    HighsStatus optimize_model_status = this->optimizeModel();
	if (lp.isMip())
		save col values for next

  	std::vector<HighsInt> index(lp.num_col_);
  	std::vector<double> value(lp.num_col_);
	for (HighsInt iCol = 0; iCol < lp.num_col_; iCol++) {
      if (lp.col_cost_[iCol]) {
        index[nnz] = iCol;
        value[nnz] = lp.col_cost_[iCol];
        nnz++;
      }

	double lower_bound = -kHighsInf;
    double upper_bound = kHighsInf;
	if (lp.sense_ == ObjSense::kMinimize)
	  if (linear_objective.abs_tolerance >= 0)
        upper_bound = objective + linear_objective.abs_tolerance;
      if (linear_objective.rel_tolerance >= 0) {
	    if (objective >= 0) {
		  upper_bound = std::min(objective * (1.0 + linear_objective.rel_tolerance), upper_bound);
		else
		  upper_bound = std::min(objective * (1.0 - linear_objective.rel_tolerance), upper_bound);
	addRow(lower_bound, upper_bound, nnz, index, value);
*/

func (fit *FittingSingleStatWeightProcess) Run() *util_async.FutureCancellable[FittingSingleStatResult] {
	fit.setupLinearObjectives()

	for sample := range util.ForPointer(fit.inputData) {
		fit.addSample(sample)
	}

	fit.includeCountRow.Build(fit.build, float64(len(fit.inputData))*fit.minimumIncludeRate, util_highs.C_PlusInf)

	solutionFuture := fit.build.RunHighsFuture(&fit.stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) (FittingSingleStatResult, bool) {
		solution := linearResult.GetSolutionAndSaveLog(fit.printer)
		fit.printer.Println(solution.Status.String())
		fit.build.DebugPrintColumns(solution, fit.printer)
		if solution.HasSolution() {
			return fit.buildResult(solution), true
		} else {
			return FittingSingleStatResult{}, false
		}
	})
}

func (fit *FittingSingleStatWeightProcess) setupLinearObjectives() {

	// var scaleIncludeObjective float64
	// scaleIncludeObjective = 20 // 20 used previously for the 256 samples, maybe a touch too high
	// scaleIncludeObjective = 19 // 15 is too low for 256, 19 about right
	// scaleIncludeObjective = 12 // for 2000: 10-12 is a bit low. 15 was nice maybe too high? ideal range somewhere 13-14, but runs take 1H
	// these numbers are just too finicky

	// fit.input.BlendMultiObjectives = true
	// fit.linearInclude = fit.input.AddLinearObjective(scaleIncludeObjective, 0, 1000, 1, 1)
	// fit.linearLineDiff = fit.input.AddLinearObjective(1, 0, 10000, 5, 2)

	// averageIncludedDifference := 14 // based on data, haste_dps. rather not use this

	fit.build.BlendMultiObjectives = false

	var relativeToleranceParam float64
	if fit.minimumIncludeRate < 1 {
		// first linear step find a regular solution to the line fit
		// will probably follow the minimum required include
		// will get us a positive initial result from the sum of differenceAbs
		// let it expand to full coverage if it wants, but without worsening the average difference
		multiplierToFullCoverage := 1 / fit.minimumIncludeRate
		// add consider a bit of factor to this, only 80% etc, otherwise might get too greedy
		multiplierToFullCoverage *= 0.5
		// highs logic is "objective * (1.0 + linear_objective.rel_tolerance)", so need to minus one in compenstation
		// don't let it go negative or below a small value
		relativeToleranceParam = max(multiplierToFullCoverage-1, 0.1)
	} else {
		relativeToleranceParam = 0
	}
	fit.objectiveLineDiff = fit.build.AddObjectivePrioritised(false, -1, relativeToleranceParam, 2)

	// second priority is sum of includeColumn which are negative one each, can lead to negative total objective
	// but we don't need to care about offsets much since its the last one, highs shouldn't even look at them
	fit.objectiveInclude = fit.build.AddObjectivePrioritised(false, -1, -1, 1)

	// we might want to increase c_outputIncludePerInclude a bit since average at least for our first test case is 3-10
	// but actually unless we combine the objectives then they aren't getting scaled against each other anyway
}

func (fit *FittingSingleStatWeightProcess) buildResult(solution *highs.Solution) FittingSingleStatResult {
	result := FittingSingleStatResult{}
	result.LineSlope = solution.ColValues[fit.lineSlope] / fit.inputDataSimScale
	result.LineOffset = solution.ColValues[fit.lineOffset] / fit.inputDataSimScale
	result.Minimum = uint32(math.Round(solution.ColValues[fit.minimumThreshold]))
	result.Maximum = uint32(math.Round(solution.ColValues[fit.maximumThreshold]))

	var includeCount uint32 = 0
	for _, col := range fit.includeColumns {
		if util.FloatEqualsOne(solution.ColValues[col]) {
			includeCount++
		}
	}
	result.IncludeCount = includeCount
	result.IncludePercent = float64(includeCount) / float64(len(fit.inputData))

	if includeCount == 0 {
		panic("shouldn't this have failed in model")
	}

	inputSorted := slices.Clone(fit.inputData)
	slices.SortFunc(inputSorted, func(a, b fittingSample) int { return cmp.Compare(a.statValue, b.statValue) })
	for _, sample := range inputSorted {
		fit.printer.Printf("INC %f %f\n", sample.statValue, solution.ColValues[sample.includeColumn])
	}

	result.StopwatchSolver = fit.stopwatch

	return result
}

func (fit *FittingSingleStatWeightProcess) addSample(sample *fittingSample) {
	includeColumn := fit.sampleIncludeToggleColumn(sample)
	fit.sampleToFitLine(sample, includeColumn)
}

func (fit *FittingSingleStatWeightProcess) sampleIncludeToggleColumn(sample *fittingSample) util_highs.ColumnIndex {
	includeColumn := fit.build.CreateColumnBoolWithObjective(c_outputFittingPerInclude, fit.objectiveInclude, util_highs.DebugString{Text: "include"})
	fit.includeCountRow.Add(includeColumn, 1)
	fit.includeColumns = append(fit.includeColumns, includeColumn)
	sample.includeColumn = includeColumn

	fit.build.ConstantIsBetweenColumns(fit.minimumThreshold, fit.maximumThreshold, includeColumn, sample.statValue, c_statRangeHigh, 1.0)

	// another thought, samples could be presorted and indexed, then we setup relationships between adjactent pairs,
	// they pull each other up, until we reach a sample marked as THE high/low cutoff

	return includeColumn
}

func (fit *FittingSingleStatWeightProcess) sampleToFitLine(sample *fittingSample, toggle util_highs.ColumnIndex) {
	difference := fit.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugString{Text: "difference"})
	differenceAbs := fit.build.CreateColumnWithObjective(highs.Continuous, 0, util_highs.C_PlusInf, c_outputFittingDifference, fit.objectiveLineDiff, util_highs.DebugString{Text: "differenceAbs"})

	// i'd like lineSlope to look like sim/stat
	// i don't really care what lineOffset looks like, don't expect to use it at all
	// basic line formula:               y = lineSlope * x + lineOffset
	//                      y - lineOffset = lineSlope * x
	//                  y/x - lineOffset/x = lineSlope
	//          sim/stat - lineOffset/stat = lineSlope
	//                            sim/stat = lineSlope + lineOffset/stat
	//                                 sim = lineSlope*stat + lineOffset
	sampleRow := util_highs.ConstraintRow{Debug: "sampleRow"}
	sampleRow.Add(fit.lineSlope, sample.statValue)
	sampleRow.Add(fit.lineOffset, 1)
	sampleRow.Add(difference, 1) // now technically this is a "vertical" difference, not a anything squared, but hopefully proportional...
	sampleRow.Build(fit.build, sample.simResult, sample.simResult)

	// new absolute val with toggle, this is its test
	fit.build.AbsoluteValue_WithToggle(difference, differenceAbs, toggle, c_simRangeHigh)
}
