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
	c_fitting2_each_threadCount = 8

	c_fitting2_statScaledMaxValue = 1.0
	c_fitting2_simScaledHighM     = 2.0
	c_fitting2_statScaledHighM    = 2.0

	//c_fitting2_minStatsRequireOverlapBetweenSegments = 0.001 // need a bit of overlap (zero maybe ok) to fix the common entries as line meeting points
	//c_fitting2_minStatsRequireOverlapBetweenSegments = 0.0  // need a bit of overlap (zero maybe ok) to fix the common entries as line meeting points
	//c_fitting2_maxStatsAllowOverlapBetweenSegments   = 0.01 // 1% is about 150 stat given standard 15000 range

	c_fitting2_statUnscaledHigh       = 50000
	c_fitting2_statScaledUnequalDelta = c_fitting2_statScaledMaxValue / float64(c_fitting2_statUnscaledHigh)

	c_fitting2_minimum_stat_coverage       = 100
	c_fitting2_permitted_overlap_fix       = 50
	c_fitting2_number_nice_number_interval = 5
)

type FittingEachStatWeightProcess2 struct {
	printer *util.PrintRecorder
	timeout int

	targetSegmentCount int
	inputData          []weight_types.WeightInput
	requiredStats      []stats.StatType
	requiredSims       []stats.SimType

	scaleSims  util_collection.EnumMap[stats.SimType, scaleAndOffset]
	scaleStats util_collection.EnumMap[stats.StatType, float64]

	each util_collection.MapMap[stats.StatType, stats.SimType, *fitting2EachFields]
}

type fitting2EachFields struct {
	statType stats.StatType
	simType  stats.SimType
	process  FittingSingleStatSegmentsProcess2
	//resultSet Fitting2InitialResultSet
	resultSlice []FittingInterimResult
}

func (fe *FittingEachStatWeightProcess2) Init(targetSegmentCount int, printer *util.PrintRecorder, timeout int) {
	fe.targetSegmentCount = targetSegmentCount
	fe.printer = printer
	fe.timeout = timeout
}

func (fe *FittingEachStatWeightProcess2) SetRequiredStats(requiredStats []stats.StatType, requiredSims []stats.SimType) {
	fe.requiredStats = requiredStats
	fe.requiredSims = requiredSims
}

func (fe *FittingEachStatWeightProcess2) SupplyData(inputData []weight_types.WeightInput) {
	fe.inputData = inputData
}

func (fe *FittingEachStatWeightProcess2) Run(stopwatch *util.Stopwatch, cancel util_async.CancelSignal) weight_types.Weight3ExtendedRanged {
	fe.chooseScaling()
	fe.launchEachNested(cancel)
	weights := fe.buildResult()
	fe.calcMetrics(stopwatch)
	return *weights
}

func (fe *FittingEachStatWeightProcess2) calcMetrics(stopwatch *util.Stopwatch) {
	for fields := range fe.each.SeqValues() {
		for _, detail := range fields.resultSlice {
			stopwatch.AddElapsedFrom(&detail.StopwatchSolver)
		}
	}
}

func (fe *FittingEachStatWeightProcess2) buildResult() *weight_types.Weight3ExtendedRanged {
	weights := weight_types.Weight3ExtendedRanged_Make(fe.requiredStats, fe.requiredSims)
	fe.each.ForeachWithKeys(func(statType stats.StatType, simType stats.SimType, value *fitting2EachFields) {
		for _, detail := range value.resultSlice {
			weights.AddDetailWeight(simType, statType, detail.StatRange, detail.LineSlope, detail.LineOffset, detail.IncludePercentOfTotal)
		}
	})
	weights.FinishAndValidate()
	return weights
}

func (fe *FittingEachStatWeightProcess2) launchEachNested(cancel util_async.CancelSignal) {
	for _, statType := range fe.requiredStats {
		for _, simType := range fe.requiredSims {
			printer := util.PrintRecorder_HoldAll()
			fields := fitting2EachFields{statType: statType, simType: simType}
			fields.process.Init(fe.targetSegmentCount, printer, fe.timeout)
			fields.process.SupplyData(fe.prepareSamples(statType, simType))
			fe.each.Put(statType, simType, &fields)
		}
	}

	channelEach := util_async.SeqToChannel_Cancellable(fe.each.SeqValues(), cancel)
	util_async.ForEach_Channel(c_fitting2_each_threadCount, channelEach, func(fields *fitting2EachFields) {
		initialResultFuture := fields.process.Run()
		initialResult := initialResultFuture.WaitForResultOrPanic()
		fe.rescaleAndCleanup(initialResult, fields)
	})
}

func (fe *FittingEachStatWeightProcess2) chooseScaling() {
	fe.scaleStats = chooseStatScalingBasic(fe.inputData, c_fitting2_statScaledMaxValue, true, fe.printer)
	fe.scaleSims = chooseSimUnfriendlyUnitScaleAndOffset(fe.inputData, fe.requiredSims)
}

func (fe *FittingEachStatWeightProcess2) prepareSamples(statType stats.StatType, simType stats.SimType) []FittingSample {
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

func (fe *FittingEachStatWeightProcess2) rescaleAndCleanup(initialSet Fitting2InitialResultSet, fields *fitting2EachFields) {
	fields.resultSlice = fe.convertAndScaleResult(initialSet, fields.statType)
	fields.resultSlice = fe.cleanupRanges(fields.resultSlice)
}

func (fe *FittingEachStatWeightProcess2) convertAndScaleResult(initialSet Fitting2InitialResultSet, statType stats.StatType) []FittingInterimResult {
	scaleStat := fe.scaleStats.GetOrPanic(statType)

	resultSlice := make([]FittingInterimResult, 0, len(initialSet.Segments))
	for i, resultInitial := range initialSet.Segments {
		interim := FittingInterimResult{
			LineSlope:  resultInitial.LineSlope * scaleStat,
			LineOffset: resultInitial.LineOffset,
			StatRange: weight_types.StatRange{
				Minimum: uint32(math.Round(resultInitial.StatRange.Minimum / scaleStat)),
				Maximum: uint32(math.Round(resultInitial.StatRange.Maximum / scaleStat)),
			},
			IncludeCount:               resultInitial.IncludeCount,
			IncludePercentOfTotal:      resultInitial.IncludePercentOfTotal,
			IncludePercentOfStageInput: 0,
			BuiltSequence:              []int{i},
			StopwatchSolver:            resultInitial.StopwatchSolver,
		}
		resultSlice = append(resultSlice, interim)
	}
	return resultSlice
}

func (fe *FittingEachStatWeightProcess2) cleanupRanges(results []FittingInterimResult) []FittingInterimResult {
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

func (fe *FittingEachStatWeightProcess2) updateBreakpoint(one, two *FittingInterimResult) (deleteSecond bool) {
	if one.StatRange.RangeSize() < c_fitting2_minimum_stat_coverage || two.StatRange.RangeSize() < c_fitting2_minimum_stat_coverage {
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
	} else if one.StatRange.Maximum >= two.StatRange.Minimum && one.StatRange.Maximum <= two.StatRange.Minimum+c_fitting2_permitted_overlap_fix {
		// if maximum has small overlap into next minimum, fix
		newMinimum := fe.chooseNicerNumber(one.StatRange.Maximum-1, two.StatRange.Minimum+1)
		two.StatRange.Minimum = newMinimum
		one.StatRange.Maximum = newMinimum - 1
	} else if one.StatRange.Overlap(two.StatRange) {
		// more overlap than threshold in previous, fail
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

func (fe *FittingEachStatWeightProcess2) chooseNicerNumber(lo uint32, hi uint32) uint32 {
	if lo > hi {
		lo, hi = hi, lo
	}

	if hi-lo > c_fitting2_number_nice_number_interval {
		mid := lo + (hi-lo)/2
		mid -= mid % c_fitting2_number_nice_number_interval
		return mid
	} else {
		for v := lo; v <= hi; v++ {
			if v%c_fitting2_number_nice_number_interval == 0 {
				return v
			}
		}
		mid := lo + (hi-lo)/2
		return mid
	}
}

// //////////////////////////////////////////////////////
type FittingSingleStatSegmentsProcess2 struct {
	printer   *util.PrintRecorder
	stopwatch util.Stopwatch
	build     *util_highs.LinearBuilder
	timeout   int

	inputData []FittingSample

	targetSegmentCount int
	segments           []*fitting2SegmentVars

	objectiveLineFitSlack util_highs.ObjectiveIndex
	objectiveInclude      util_highs.ObjectiveIndex
	objectiveThresholds   util_highs.ObjectiveIndex
	objectiveLineOverlap  util_highs.ObjectiveIndex
}

type fitting2SegmentVars struct {
	process          *FittingSingleStatSegmentsProcess2
	lineSlope        util_highs.ColumnIndex
	lineOffset       util_highs.ColumnIndex
	minimumThreshold util_highs.ColumnIndex
	maximumThreshold util_highs.ColumnIndex
	//includeSampleRow   util_highs.ConstraintRow
	//includeSampleCount util_highs.ColumnIndex
	includeColumns    []util_highs.ColumnIndex
	includeOverlapRow util_highs.ConstraintRow
	isFirst, isLast   bool
}

type Fitting2InitialResultSet struct {
	Segments []Fitting2InitialSegment
}

type Fitting2InitialSegment struct {
	LineSlope             float64
	LineOffset            float64
	StatRange             weight_types.StatRangeFloat
	IncludeCount          uint32
	IncludePercentOfTotal float64
	StopwatchSolver       util.Stopwatch
}

func (fg *FittingSingleStatSegmentsProcess2) Init(targetSegmentCount int, printer *util.PrintRecorder, timeout int) {
	fg.targetSegmentCount = targetSegmentCount
	fg.printer = printer

	fg.build = new(util_highs.LinearBuilder)
	fg.build.Minimise = true
	fg.build.Solver = util_highs.Solver_MIP_Interior
	fg.build.TimeLimitSeconds = timeout

	fg.setupObjectives()
}

func (fg *FittingSingleStatSegmentsProcess2) setupObjectives() {
	fg.build.BlendMultiObjectives = true
	// regular line fitting, compare to scaled sim, average 0.01 but multiplied by 500(N), so expect total about 5.0
	fg.objectiveLineFitSlack = fg.build.AddObjectiveBlended(1, 0)
	// similar scale numbers to regular line fitting, only one per segment, but only moderate importance
	fg.objectiveLineOverlap = fg.build.AddObjectiveBlended(500, 0)
	// smallish numbers need a boost up to 0.01 just per segment, and is important
	fg.objectiveThresholds = fg.build.AddObjectiveBlended(1000, 0)
	// negative weight for includes since we want to maximize, should slightly overpower the fitting cost of them
	fg.objectiveInclude = fg.build.AddObjectiveBlended(-0.05, 0)
}

func (fg *FittingSingleStatSegmentsProcess2) setupObjectives2() {
	fg.build.BlendMultiObjectives = false

	// regular line fitting, compare to scaled sim, average 0.01 but multiplied by 500(N), so expect total about 5.0
	fg.objectiveLineFitSlack = fg.build.AddObjectivePrioritised(false, -1, 0.3, 4)
	// negative weight for includes since we want to maximize, should slightly overpower the fitting cost of them
	fg.objectiveInclude = fg.build.AddObjectivePrioritised(true, -1, 0.1, 3)
	// smallish numbers need a boost up to 0.01 just per segment, and is important
	fg.objectiveThresholds = fg.build.AddObjectivePrioritised(false, -1, 0.1, 2)
	// similar scale numbers to regular line fitting, only one per segment, but only moderate importance
	fg.objectiveLineOverlap = fg.build.AddObjectivePrioritised(false, -1, -1, 1)
}

func (fg *FittingSingleStatSegmentsProcess2) SupplyData(inputData []FittingSample) {
	fg.inputData = slices.Clone(inputData)
}

func (fg *FittingSingleStatSegmentsProcess2) Run() *util_async.FutureCancellable[Fitting2InitialResultSet] {
	for i := range fg.targetSegmentCount {
		if i == 0 && fg.targetSegmentCount == 1 {
			fg.addSegment(true, true, nil)
		} else if i == 0 {
			fg.addSegment(true, false, nil)
		} else if i == fg.targetSegmentCount-1 {
			fg.addSegment(false, true, fg.segments[i-1])
		} else {
			fg.addSegment(false, false, fg.segments[i-1])
		}
		//fg.addSegment(i == 0, i == fg.targetSegmentCount-1)
	}
	for i := range len(fg.segments) - 1 {
		fg.enforceCrossSegmentRules(fg.segments[i], fg.segments[i+1])
	}
	fg.processData()
	for i := range len(fg.segments) {
		fg.finishSegment(fg.segments[i], i == len(fg.segments)-1)
	}

	future := fg.build.RunHighsFuture(&fg.stopwatch)
	return util_async.FutureCancellable_MapValue(future, func(res util_highs.LinearResult) (Fitting2InitialResultSet, bool) {
		solution := res.GetSolution2AndSaveLog(fg.printer)
		solution.DebugPrint(fg.printer)
		if solution.HasSolution() {
			return fg.prepareResult(solution), true
		} else {
			return Fitting2InitialResultSet{}, false
		}
	})
}

func (fg *FittingSingleStatSegmentsProcess2) addSegment(first, last bool, prev *fitting2SegmentVars) {
	fs := &fitting2SegmentVars{isFirst: first, isLast: last}
	fs.lineSlope = fg.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "slope"})
	fs.lineOffset = fg.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "offset"})

	if first && last {
		fs.minimumThreshold = -1
		fs.maximumThreshold = -1
	} else if first {
		fs.minimumThreshold = -1
		fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})
	} else if last {
		fs.minimumThreshold = prev.maximumThreshold
		fs.maximumThreshold = -1
	} else {
		fs.minimumThreshold = prev.maximumThreshold
		fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})
		fg.build.ColumnIsGreaterOrEqualColumnEnforce(fs.minimumThreshold, fs.maximumThreshold)
	}

	//if first {
	//	fs.minimumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, 0, util_highs.DebugString{Text: "minimum"})
	//} else {
	//	fs.minimumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "minimum"})
	//}
	//if last {
	//	fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, c_fitting2_statScaledMaxValue, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})
	//} else {
	//	fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})
	//}
	//fg.build.ColumnIsGreaterOrEqualColumnEnforce(fs.minimumThreshold, fs.maximumThreshold)

	fg.segments = append(fg.segments, fs)
}

func (fg *FittingSingleStatSegmentsProcess2) finishSegment(segment *fitting2SegmentVars, isLast bool) {
	if !isLast {
		segment.includeOverlapRow.Build(fg.build, 1, 1)
	}

	//segment.includeSampleCount = fg.build.CreateColumnGeneral(highs.Integer, 0, util_highs.InfPos(), util_highs.DebugText("includeSampleCount"))
	//segment.includeSampleRow.Add(segment.includeSampleCount, -1)
	//segment.includeSampleRow.Build(fg.build, 0, 0)
}

func (fg *FittingSingleStatSegmentsProcess2) enforceCrossSegmentRules(one *fitting2SegmentVars, two *fitting2SegmentVars) {
	//thresholdCompareSlack := fg.build.CreateColumnWithObjective(highs.Continuous, c_fitting2_minStatsRequireOverlapBetweenSegments, c_fitting2_maxStatsAllowOverlapBetweenSegments, 1, fg.objectiveThresholds, util_highs.DebugString{Text: "thresholdCompareSlack"})
	//compareThreshold := util_highs.ConstraintRow{}
	//compareThreshold.Add(one.maximumThreshold, 1)
	//compareThreshold.Add(two.minimumThreshold, -1)
	//compareThreshold.Add(thresholdCompareSlack, -1)
	//compareThreshold.Build(fg.build, 0, 0)

	// enforce some range between segments
	//compareThreshold := util_highs.ConstraintRow{}
	//compareThreshold.Add(one.maximumThreshold, 1)
	//compareThreshold.Add(two.minimumThreshold, -1)
	//compareThreshold.Build(fg.build, c_fitting2_minStatsRequireOverlapBetweenSegments, c_fitting2_maxStatsAllowOverlapBetweenSegments)

	// two>one
	//compareThreshold := util_highs.ConstraintRow{}
	//compareThreshold.Add(one.maximumThreshold, -1)
	//compareThreshold.Add(two.minimumThreshold, 1)
	//compareThreshold.Build(fg.build, 0, util_highs.InfPos())

	// simple equality
	//compareThreshold := util_highs.ConstraintRow{}
	//compareThreshold.Add(one.maximumThreshold, -1)
	//compareThreshold.Add(two.minimumThreshold, 1)
	//compareThreshold.Build(fg.build, 0, 0)

	if one.maximumThreshold != -1 && two.minimumThreshold != -1 && one.maximumThreshold != two.minimumThreshold {
		panic("columns wrong")
	}
}

func (fg *FittingSingleStatSegmentsProcess2) processData() {
	slices.SortFunc(fg.inputData, func(a, b FittingSample) int {
		return cmp.Compare(a.StatValue, b.StatValue)
	})

	for _, sample := range fg.inputData {
		fg.addSample(sample)
	}
}

func validateSample(sample FittingSample) {
	if sample.SimResult < 0 || sample.SimResult > 1 || sample.StatValue < 0 || sample.StatValue > 1 {
		panic("sample out of range")
	}
}

func (fg *FittingSingleStatSegmentsProcess2) addSample(sample FittingSample) {
	validateSample(sample)

	includeInSegments := make([]util_highs.ColumnIndex, len(fg.segments))
	for segIndex, segment := range fg.segments {
		includeColumn := fg.sampleIncludeToggleColumn(sample, segment)
		includeInSegments[segIndex] = includeColumn

		fg.sampleToFitLine(sample, segment, includeColumn)
	}

	for i := range len(fg.segments) - 1 {
		fg.prepareAsPotentialThreshold(fg.segments[i], fg.segments[i+1], includeInSegments[i], includeInSegments[i+1], sample)
	}

	rowIncludeInOne := util_highs.ConstraintRow{}
	for _, col := range includeInSegments {
		rowIncludeInOne.Add(col, 1)
	}
	rowIncludeInOne.Build(fg.build, 1, 2)
}

func (fg *FittingSingleStatSegmentsProcess2) sampleIncludeToggleColumn(sample FittingSample, segment *fitting2SegmentVars) util_highs.ColumnIndex {
	//includeColumn := fg.build.CreateColumnBoolWithObjective(1, fg.objectiveInclude, util_highs.DebugString{Text: "include"})
	includeColumn := fg.build.CreateColumnBool(util_highs.DebugString{Text: "include"})

	if segment.isFirst && segment.isLast {
		row := util_highs.ConstraintRow{}
		row.Add(includeColumn, 1)
		row.Build(fg.build, 1, 1)
	} else if segment.isFirst {
		fg.build.ColumnIsGreaterOrEqualThanConstant_Supplied(includeColumn, segment.maximumThreshold, sample.StatValue, c_fitting2_statScaledHighM, c_fitting2_statScaledUnequalDelta)
	} else if segment.isLast {
		fg.build.ColumnIsLessOrEqualThanConstant_Supplied(includeColumn, segment.minimumThreshold, sample.StatValue, c_fitting2_statScaledHighM, c_fitting2_statScaledUnequalDelta)
	} else {
		fg.build.ConstantIsBetweenColumns_NoSequenceCheck(segment.minimumThreshold, segment.maximumThreshold, includeColumn, sample.StatValue, c_fitting2_statScaledHighM, c_fitting2_statScaledUnequalDelta)
	}

	segment.includeColumns = append(segment.includeColumns, includeColumn)
	return includeColumn
}

func (fg *FittingSingleStatSegmentsProcess2) prepareAsPotentialThreshold(seg1, seg2 *fitting2SegmentVars, include1, include2 util_highs.ColumnIndex, sample FittingSample) {
	includeBoth := fg.build.CreateColumnBool(util_highs.DebugText("includeBoth"))
	seg1.includeOverlapRow.Add(includeBoth, 1)
	fg.build.ConstraintAnd(includeBoth, include1, include2)
	// TODO can this be simplified further depending

	differenceSigned := fg.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "differenceSigned"})
	difference := fg.build.CreateColumnWithObjective(highs.Continuous, 0, util_highs.InfPos(), 1, fg.objectiveLineOverlap, util_highs.DebugString{Text: "difference"})

	// overlapRow: one.lineSlope*StatValue + one.lineOffset - two.lineSlope*StatValue - two.lineOffset + difference = 0
	overlapRow := util_highs.ConstraintRow{Debug: "overlapRow"}
	overlapRow.Add(seg1.lineSlope, sample.StatValue)
	overlapRow.Add(seg1.lineOffset, 1)
	overlapRow.Add(seg2.lineSlope, -sample.StatValue)
	overlapRow.Add(seg2.lineOffset, -1)
	overlapRow.Add(differenceSigned, 1)
	overlapRow.Build(fg.build, 0, 0)

	fg.build.AbsoluteValue_WithToggle_NoExtraCheck(differenceSigned, difference, includeBoth, c_fitting2_simScaledHighM)
}

func (fg *FittingSingleStatSegmentsProcess2) sampleToFitLine(sample FittingSample, segment *fitting2SegmentVars, include util_highs.ColumnIndex) {
	differenceSigned := fg.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "differenceSigned"})
	difference := fg.build.CreateColumnWithObjective(highs.Continuous, 0, util_highs.InfPos(), 1, fg.objectiveLineFitSlack, util_highs.DebugString{Text: "difference"})

	// sampleRow: one.lineSlope*StatValue + one.lineOffset + difference = simResult
	sampleRow := util_highs.ConstraintRow{Debug: "sampleRow"}
	sampleRow.Add(segment.lineSlope, sample.StatValue)
	sampleRow.Add(segment.lineOffset, 1)
	sampleRow.Add(differenceSigned, 1)
	sampleRow.Build(fg.build, sample.SimResult, sample.SimResult)

	fg.build.AbsoluteValue_WithToggle_NoExtraCheck(differenceSigned, difference, include, c_fitting2_simScaledHighM)
}

func (fg *FittingSingleStatSegmentsProcess2) prepareResult(solution *util_highs.Solution2) Fitting2InitialResultSet {
	resultSet := Fitting2InitialResultSet{}

	for _, segment := range fg.segments {
		var statRange weight_types.StatRangeFloat
		if segment.minimumThreshold != -1 {
			statRange.Minimum = solution.GetValue(segment.minimumThreshold)
		}
		if segment.maximumThreshold != -1 {
			statRange.Maximum = solution.GetValue(segment.maximumThreshold)
		} else {
			statRange.Maximum = math.MaxUint32
		}

		includeSampleCount := fg.countSamples(segment, solution)
		interim := Fitting2InitialSegment{
			LineSlope:             solution.GetValue(segment.lineSlope),
			LineOffset:            solution.GetValue(segment.lineOffset),
			StatRange:             statRange,
			IncludeCount:          uint32(includeSampleCount),
			IncludePercentOfTotal: float64(includeSampleCount) / float64(len(fg.inputData)),
			StopwatchSolver:       util.Stopwatch{},
		}

		resultSet.Segments = append(resultSet.Segments, interim)
	}

	return resultSet
}

func (fg *FittingSingleStatSegmentsProcess2) countSamples(segment *fitting2SegmentVars, solution *util_highs.Solution2) int {
	count := 0
	for _, column := range segment.includeColumns {
		if solution.ValueIsOne(column) {
			count++
		}
	}
	return count
}
