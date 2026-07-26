package weight_highs

import (
	"cmp"
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"

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
)

type FittingSample struct {
	StatValue float64
	SimResult float64
}

type statRangeFloat struct {
	Minimum float64
	Maximum float64
}

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

	onlyComputeSingleSegmentEach bool
	inputData                    []weight_types.WeightInput
	requiredStats                []stats.StatType
	requiredSims                 []stats.SimType

	scaleSims  util_collection.EnumMap[stats.SimType, scaleAndOffset]
	scaleStats util_collection.EnumMap[stats.StatType, float64]

	each util_collection.MapMap[stats.StatType, stats.SimType, *fittingEachFields]
}

type fittingEachFields struct {
	statType    stats.StatType
	simType     stats.SimType
	process     FittingSingleStatSegmentsProcess
	resultSlice []FittingInterimResult
}

func (fe *FittingEachStatWeightProcess) Init(printer *util.PrintRecorder, timeout int) {
	fe.printer = printer
	fe.timeout = timeout
}

func (fe *FittingEachStatWeightProcess) SetRequiredStats(requiredStats []stats.StatType, requiredSims []stats.SimType) {
	fe.requiredStats = requiredStats
	fe.requiredSims = requiredSims
}

// if just converting directly to weights1
func (fe *FittingEachStatWeightProcess) SetOnlyComputeSingleSegmentEach(lazy bool) {
	fe.onlyComputeSingleSegmentEach = lazy
}

func (fe *FittingEachStatWeightProcess) SupplyData(inputData []weight_types.WeightInput) {
	fe.inputData = inputData
}

func (fe *FittingEachStatWeightProcess) Run(stopwatch *util.Stopwatch, cancel util_async.CancelSignal) weight_types.Weight3ExtendedRanged {
	fe.chooseScaling()
	fe.launchEachNested(cancel)
	weights := fe.buildResult()
	fe.calcMetrics(stopwatch)
	return *weights
}

func (fe *FittingEachStatWeightProcess) calcMetrics(stopwatch *util.Stopwatch) {
	for fields := range fe.each.SeqValues() {
		for _, detail := range fields.resultSlice {
			stopwatch.AddElapsedFrom(&detail.StopwatchSolver)
		}
	}
}

func (fe *FittingEachStatWeightProcess) buildResult() *weight_types.Weight3ExtendedRanged {
	weights := weight_types.Weight3ExtendedRanged_Make(fe.requiredStats, fe.requiredSims)
	fe.each.ForeachWithKeys(func(statType stats.StatType, simType stats.SimType, value *fittingEachFields) {
		for _, detail := range value.resultSlice {
			weights.AddDetailWeight(simType, statType, detail.StatRange, detail.LineSlope, detail.LineOffset, detail.IncludePercentOfTotal)
		}
	})
	weights.FinishAndValidate()
	return weights
}

func (fe *FittingEachStatWeightProcess) launchEachNested(cancel util_async.CancelSignal) {
	for _, statType := range fe.requiredStats {
		for _, simType := range fe.requiredSims {
			printer := util.PrintRecorder_HoldAll()
			fields := fittingEachFields{statType: statType, simType: simType}
			fields.process.Init(printer, fe.timeout)
			fields.process.SetOnlyComputeSingleSegment(fe.onlyComputeSingleSegmentEach)
			fields.process.SupplyData(fe.prepareSamples(statType, simType))
			fe.each.Put(statType, simType, &fields)
		}
	}

	channelEach := util_async.SeqToChannel_Cancellable(fe.each.SeqValues(), cancel)
	util_async.ForEach_Channel(c_fitting_each_threadCount, channelEach, func(fields *fittingEachFields) {
		initialResult := fields.process.Run(cancel)
		fe.rescaleAndCleanup(initialResult, fields)
	})
}

func (fe *FittingEachStatWeightProcess) chooseScaling() {
	fe.scaleStats = chooseStatScalingBasic(fe.inputData, c_fitting_statScaledRangeHigh, true, fe.printer)
	fe.scaleSims = chooseSimUnfriendlyUnitScaleAndOffset(fe.inputData, fe.requiredSims)
}

func (fe *FittingEachStatWeightProcess) prepareSamples(statType stats.StatType, simType stats.SimType) []FittingSample {
	scaleStat := fe.scaleStats.GetOrPanic(statType)
	scaleSim := fe.scaleSims.GetOrPanic(simType)

	samples := make([]FittingSample, len(fe.inputData))
	for i := range fe.inputData {
		statValue := fe.inputData[i].TotalStat.GetFloat(statType)
		simResult := fe.inputData[i].SimResult.Get(simType)
		samples[i] = FittingSample{
			StatValue: statValue * scaleStat,
			SimResult: scaleSim.Apply(simResult),
		}
	}
	return samples
}

// inner process ran: simScaled = lineSlopeInternal*statScaled + lineOffsetInternal
// for the sake of Weight3 we want to retain the simScaling, but undo the stats
//
//	want to end up with: simScaled = lineSlopeUsable * stat + lineOffsetUsable
//
// expanding on the first one: simScaled = lineSlopeInternal*(stat*scaleStat) + lineOffsetInternal
// when stat=0: simScaled = lineOffsetInternal. not sure if it's a true equality, but doesn't seem to need scale at least
// to fix the stat scale lineSlopeUsable * stat = lineSlopeInternal * (stat * scaleStat)
//
//	lineSlopeUsable = lineSlopeInternal * scaleStat
func (fe *FittingEachStatWeightProcess) rescaleAndCleanup(initialMap map[statRangeFloat]FittingSingleStatResult, fields *fittingEachFields) {
	fields.resultSlice = fe.convertAndScaleResult(initialMap, fields.statType)
	fields.resultSlice = fe.cleanupRanges(fields.resultSlice)
}

func (fe *FittingEachStatWeightProcess) convertAndScaleResult(initialMap map[statRangeFloat]FittingSingleStatResult, statType stats.StatType) []FittingInterimResult {
	scaleStat := fe.scaleStats.GetOrPanic(statType)

	resultSlice := make([]FittingInterimResult, 0, len(initialMap))
	for rangeScaled, resultInitial := range initialMap {
		interim := FittingInterimResult{
			LineSlope:  resultInitial.LineSlope * scaleStat,
			LineOffset: resultInitial.LineOffset,
			StatRange: weight_types.StatRange{
				Minimum: uint32(math.Round(rangeScaled.Minimum / scaleStat)),
				Maximum: uint32(math.Round(rangeScaled.Maximum / scaleStat)),
			},
			IncludeCount:               resultInitial.IncludeCount,
			IncludePercentOfTotal:      float64(resultInitial.IncludeCount) / float64(len(fe.inputData)),
			IncludePercentOfStageInput: resultInitial.IncludePercentOfStageInput,
			BuiltSequence:              []int{resultInitial.BuiltSequence},
			StopwatchSolver:            resultInitial.StopwatchSolver,
		}
		resultSlice = append(resultSlice, interim)
	}
	return resultSlice
}

func (fe *FittingEachStatWeightProcess) cleanupRanges(results []FittingInterimResult) []FittingInterimResult {
	slices.SortFunc(results, func(a, b FittingInterimResult) int {
		return cmp.Or(cmp.Compare(a.StatRange.Minimum, b.StatRange.Minimum), cmp.Compare(a.StatRange.Maximum, b.StatRange.Maximum))
	})

	results[0].StatRange.Minimum = 0
	results[len(results)-1].StatRange.Maximum = math.MaxUint32

	for i := range len(results) - 1 {
		if fe.updateBreakpoint(&results[i], &results[i+1]) {
			util_collection.DeleteIndexInPlace(&results, i+1)
		}
	}

	return results
}

func (fe *FittingEachStatWeightProcess) updateBreakpoint(one, two *FittingInterimResult) (deleteSecond bool) {
	if one.StatRange.RangeSize() < c_fitting_minimum_stat_coverage || two.StatRange.RangeSize() < c_fitting_minimum_stat_coverage {
		// if covers less than 100 stat numbers, merge them
		if one.StatRange.RangeSize() < two.StatRange.RangeSize() {
			one.LineSlope = two.LineSlope
			one.LineOffset = two.LineOffset
		}
		one.StatRange.Maximum = two.StatRange.Maximum
		one.IncludePercentOfStageInput += two.IncludePercentOfStageInput
		one.IncludePercentOfTotal += two.IncludePercentOfTotal
		one.IncludeCount += two.IncludeCount
		one.BuiltSequence = slices.Concat(one.BuiltSequence, two.BuiltSequence)
		one.StopwatchSolver.AddElapsedFrom(&two.StopwatchSolver)
		return true
	} else if one.StatRange.Maximum >= two.StatRange.Minimum && one.StatRange.Maximum <= two.StatRange.Minimum+c_fitting_permitted_overlap_fix {
		// if maximum has small overlap into next minimum, fix
		newMinimum := fe.chooseNicerNumber(one.StatRange.Maximum-1, two.StatRange.Minimum+1)
		two.StatRange.Minimum = newMinimum
		one.StatRange.Maximum = newMinimum - 1
	} else if one.StatRange.Overlap(two.StatRange) {
		// more overlap than threshold in previous, fail
		// shouldn't happen, anything more than exact equality due to scaling and rounding shouldn't either
		panic("overlapping ranges created")
	} else if one.StatRange.Maximum < two.StatRange.Minimum-1 {
		// part of range got dropped earlier
		newMinimum := fe.chooseNicerNumber(one.StatRange.Maximum-1, two.StatRange.Minimum+1)
		two.StatRange.Minimum = newMinimum
		one.StatRange.Maximum = newMinimum - 1
	} else if one.StatRange.Minimum < one.StatRange.Maximum && one.StatRange.Maximum+1 == two.StatRange.Minimum && two.StatRange.Minimum < two.StatRange.Maximum {
		// passes all expected conditions
	} else {
		panic("internal logic error in updateBreakpoint")
	}
	return false
}

func (fe *FittingEachStatWeightProcess) chooseNicerNumber(lo uint32, hi uint32) uint32 {
	if lo > hi {
		lo, hi = hi, lo
	}

	if hi-lo > c_fitting_number_nice_number_interval {
		mid := lo + (hi-lo)/2
		mid -= mid % c_fitting_number_nice_number_interval
		return mid
	} else {
		for v := lo; v <= hi; v++ {
			if v%c_fitting_number_nice_number_interval == 0 {
				return v
			}
		}
		mid := lo + (hi-lo)/2
		return mid
	}
}

////////////////////////////////////////////////////////

type FittingSingleStatSegmentsProcess struct {
	printer *util.PrintRecorder

	timeout                  int
	onlyComputeSingleSegment bool

	samplesOriginal       []FittingSample
	samplesRemainingParts map[statRangeFloat][]FittingSample

	foundSegments map[statRangeFloat]FittingSingleStatResult
}

func (fg *FittingSingleStatSegmentsProcess) Init(printer *util.PrintRecorder, timeout int) {
	fg.printer = printer
	fg.timeout = timeout
	fg.foundSegments = make(map[statRangeFloat]FittingSingleStatResult)
	fg.samplesRemainingParts = make(map[statRangeFloat][]FittingSample)
}

func (fg *FittingSingleStatSegmentsProcess) SetOnlyComputeSingleSegment(lazy bool) {
	fg.onlyComputeSingleSegment = lazy
}

func (fg *FittingSingleStatSegmentsProcess) SupplyData(inputData []FittingSample) {
	fg.samplesOriginal = slices.Clone(inputData)
}

const c_fitting_initial_range_required = 0.35

const c_fitting_large_range_threshold = 0.50
const c_fitting_large_range_required = 0.20

const c_fitting_medium_range_threshold = 0.25
const c_fitting_medium_range_required = 0.10

const c_fitting_small_range_threshold = 0.15
const c_fitting_small_range_required = 0.10

const c_fitting_tiny_range_threshold = 0.05
const c_fitting_tiny_range_over_required = 0.80
const c_fitting_tiny_range_under_required = 0.95

const c_fitting_dropped_range_threshold = 0.02

// let's say a standard output is about 4 line segments
// each segment thus should cover on average 25%
// initial range is a stronger requirement of at least 35%, assume remaining actually 60% (20% each)
// next ones we want to give them some slack but hoping for 15-30%
func (fg *FittingSingleStatSegmentsProcess) Run(cancel util_async.CancelSignal) map[statRangeFloat]FittingSingleStatResult {
	fg.runInitial(cancel)
	if fg.onlyComputeSingleSegment {
		return fg.foundSegments
	}

	overallSize := len(fg.samplesOriginal)
	for len(fg.samplesRemainingParts) > 0 {
		fg.mergeAnyPossibleRemainingSamples()

		nextRange, nextData := util_collection.MapFirstEntry(fg.samplesRemainingParts)
		delete(fg.samplesRemainingParts, nextRange)

		processRatioOfOverall := float64(len(nextData)) / float64(overallSize)
		targetInclude := 0.0

		if processRatioOfOverall < c_fitting_dropped_range_threshold || len(nextData) <= 8 {
			// drop it, can't get good results from small sample
			continue
		} else if processRatioOfOverall >= c_fitting_large_range_threshold {
			targetRatioOfOverall := c_fitting_large_range_required
			targetInclude = targetRatioOfOverall / processRatioOfOverall
		} else if processRatioOfOverall >= c_fitting_medium_range_threshold {
			targetRatioOfOverall := c_fitting_medium_range_required
			targetInclude = targetRatioOfOverall / processRatioOfOverall
		} else if processRatioOfOverall >= c_fitting_small_range_threshold {
			targetRatioOfOverall := c_fitting_small_range_required
			targetInclude = targetRatioOfOverall / processRatioOfOverall
		} else if processRatioOfOverall >= c_fitting_tiny_range_threshold {
			targetInclude = c_fitting_tiny_range_over_required
		} else {
			targetInclude = c_fitting_tiny_range_under_required
		}

		fg.runNextSegment(nextData, nextRange, targetInclude, cancel)
	}

	return fg.foundSegments
}

func (fg *FittingSingleStatSegmentsProcess) mergeAnyPossibleRemainingSamples() {
	type statRangeFlagged struct {
		statRange statRangeFloat
		remain    bool
	}
	flaggedSegments := make([]statRangeFlagged, 0)
	for key := range fg.samplesRemainingParts {
		flaggedSegments = append(flaggedSegments, statRangeFlagged{
			key,
			true,
		})
	}
	for key := range fg.foundSegments {
		flaggedSegments = append(flaggedSegments, statRangeFlagged{
			key,
			false,
		})
	}

	slices.SortFunc(flaggedSegments, func(a, b statRangeFlagged) int { return cmp.Compare(a.statRange.Minimum, b.statRange.Minimum) })
	for i := range len(flaggedSegments) - 1 {
		a := flaggedSegments[i]
		b := flaggedSegments[i+1]
		if a.remain && b.remain {
			combinedData := slices.Concat(fg.samplesRemainingParts[a.statRange], fg.samplesRemainingParts[b.statRange])
			combinedRange := statRangeFloat{
				Minimum: min(a.statRange.Minimum, b.statRange.Minimum),
				Maximum: max(a.statRange.Maximum, b.statRange.Maximum),
			}

			delete(fg.samplesRemainingParts, a.statRange)
			delete(fg.samplesRemainingParts, b.statRange)
			fg.samplesRemainingParts[combinedRange] = combinedData
		}
	}
}

func (fg *FittingSingleStatSegmentsProcess) runInitial(cancel util_async.CancelSignal) {
	fit := FittingSingleStatWeightProcess{}
	fit.Init(fg.printer, fg.timeout)
	fit.SetMinimumIncludeRate(c_fitting_initial_range_required)
	fit.SupplySamples(fg.samplesOriginal)

	resultOptionalFuture := fit.Run()
	util_async.ChainCancel(cancel, resultOptionalFuture)

	resultOptional := resultOptionalFuture.WaitForResultAsOptional()
	if segmentResult, hasResult := resultOptional.GetWithFlag(); hasResult {
		statRange := statRangeFloat{Minimum: segmentResult.Minimum, Maximum: segmentResult.Maximum}
		segmentResult.BuiltSequence = 0
		fg.foundSegments[statRange] = segmentResult

		totalRange := statRangeFloat{Minimum: 0, Maximum: c_fitting_statScaledRangeHigh}
		fg.addToRemainingData(fg.samplesOriginal, totalRange, statRange)
	} else {
		panic("failed to get any useful stat fit")
	}
}

func (fg *FittingSingleStatSegmentsProcess) runNextSegment(inputData []FittingSample, inputRange statRangeFloat, includeRate float64, cancel util_async.CancelSignal) {
	fit := FittingSingleStatWeightProcess{}
	fit.Init(fg.printer, fg.timeout)
	fit.SetMinimumIncludeRate(includeRate)
	fit.SupplySamples(inputData)

	resultOptionalFuture := fit.Run()
	util_async.ChainCancel(cancel, resultOptionalFuture)
	resultOptional := resultOptionalFuture.WaitForResultAsOptional()

	if segmentResult, hasResult := resultOptional.GetWithFlag(); hasResult {
		minimum := max(inputRange.Minimum, segmentResult.Minimum)
		maximum := min(inputRange.Maximum, segmentResult.Maximum)

		statRange := statRangeFloat{Minimum: minimum, Maximum: maximum}
		segmentResult.Minimum = minimum
		segmentResult.Maximum = maximum
		segmentResult.BuiltSequence = len(fg.foundSegments)
		fg.foundSegments[statRange] = segmentResult

		fg.addToRemainingData(inputData, inputRange, statRange)
	}
}

func (fg *FittingSingleStatSegmentsProcess) addToRemainingData(processedData []FittingSample, inputRange statRangeFloat, removeRange statRangeFloat) {
	if removeRange.Minimum < inputRange.Minimum || removeRange.Maximum > inputRange.Maximum || removeRange.Minimum > removeRange.Maximum || inputRange.Minimum > inputRange.Maximum {
		panic("range isn't within bounds")
	}

	loData := make([]FittingSample, 0)
	hiData := make([]FittingSample, 0)
	for _, input := range processedData {
		stat := input.StatValue
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
		loRange := statRangeFloat{Minimum: inputRange.Minimum, Maximum: removeRange.Minimum - c_fitting_statScaledUnequalDelta}
		fg.samplesRemainingParts[loRange] = loData
	}

	if len(hiData) > 0 {
		hiRange := statRangeFloat{Minimum: removeRange.Maximum + c_fitting_statScaledUnequalDelta, Maximum: inputRange.Maximum}
		fg.samplesRemainingParts[hiRange] = hiData
	}
}

////////////////////////////////////////////////////////

type FittingSingleStatWeightProcess struct {
	printer   *util.PrintRecorder
	stopwatch util.Stopwatch
	build     *util_highs.LinearBuilder

	minimumIncludeRate float64
	inputData          []FittingSample

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
type FittingInterimResult struct {
	LineSlope                  float64
	LineOffset                 float64
	StatRange                  weight_types.StatRange
	IncludeCount               uint32
	IncludePercentOfStageInput float64
	IncludePercentOfTotal      float64
	BuiltSequence              []int
	StopwatchSolver            util.Stopwatch
}

func (fw *FittingSingleStatWeightProcess) Init(printer *util.PrintRecorder, timeout int) {
	fw.printer = printer
	fw.build = new(util_highs.LinearBuilder)
	fw.build.Minimise = true
	fw.build.Solver = util_highs.Solver_MIP_Interior
	fw.build.TimeLimitSeconds = timeout

	fw.lineSlope = fw.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugString{Text: "slope"})
	fw.lineOffset = fw.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugString{Text: "offset"})
	fw.minimumThreshold = fw.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting_statScaledRangeHigh, util_highs.DebugString{Text: "minimum"})
	fw.maximumThreshold = fw.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting_statScaledRangeHigh, util_highs.DebugString{Text: "maximum"})

	fw.build.ColumnIsGreaterOrEqualColumnEnforce(fw.minimumThreshold, fw.maximumThreshold)
}

func (fw *FittingSingleStatWeightProcess) SetMinimumIncludeRate(percent float64) {
	fw.minimumIncludeRate = percent
}

func (fw *FittingSingleStatWeightProcess) SupplySamples(inputData []FittingSample) {
	fw.inputData = inputData
}

func (fw *FittingSingleStatWeightProcess) Run() *util_async.FutureCancellable[FittingSingleStatResult] {
	fw.setupLinearObjectives()

	for _, sample := range fw.inputData {
		fw.addSample(sample)
	}

	fw.includeCountRow.Build(fw.build, float64(len(fw.inputData))*fw.minimumIncludeRate, util_highs.C_PlusInf)

	solutionFuture := fw.build.RunHighsFuture(&fw.stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) (FittingSingleStatResult, bool) {
		solution := linearResult.GetSolution2AndSaveLog(fw.printer)
		solution.DebugPrint(fw.printer)
		if solution.HasSolution() {
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

func (fw *FittingSingleStatWeightProcess) addSample(sample FittingSample) {
	if sample.SimResult < 0 || sample.SimResult > 1 || sample.StatValue < 0 || sample.StatValue > 1 {
		panic("sample out of range")
	}

	includeColumn := fw.sampleIncludeToggleColumn(sample)
	fw.sampleToFitLine(sample, includeColumn)
}

func (fw *FittingSingleStatWeightProcess) sampleIncludeToggleColumn(sample FittingSample) util_highs.ColumnIndex {
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
func (fw *FittingSingleStatWeightProcess) sampleToFitLine(sample FittingSample, toggle util_highs.ColumnIndex) {
	difference := fw.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugString{Text: "difference"})
	differenceAbs := fw.build.CreateColumnWithObjective(highs.Continuous, 0, util_highs.C_PlusInf, c_fitting_outputDifference, fw.objectiveLineDiff, util_highs.DebugString{Text: "differenceAbs"})

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
