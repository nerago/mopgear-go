package weight_highs

import (
	"cmp"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_fitting2_targetSegments     = 4
	c_fitting2_statScaledMaxValue = 1.0
	c_fitting2_statScaledHighM    = 2.0

	//c_fitting2_minStatsRequireOverlapBetweenSegments = 0.001 // need a bit of overlap (zero maybe ok) to fix the common entries as line meeting points
	c_fitting2_minStatsRequireOverlapBetweenSegments = 0.0  // need a bit of overlap (zero maybe ok) to fix the common entries as line meeting points
	c_fitting2_maxStatsAllowOverlapBetweenSegments   = 0.01 // 1% is about 150 stat given standard 15000 range

	c_fitting2_statUnscaledHigh       = 50000
	c_fitting2_statScaledUnequalDelta = c_fitting2_statScaledMaxValue / float64(c_fitting2_statUnscaledHigh)
)

type FittingSingleStatSegmentsProcess2 struct {
	printer   *util.PrintRecorder
	stopwatch util.Stopwatch
	build     *util_highs.LinearBuilder
	timeout   int

	inputData []FittingSample

	segments []*fitting2SegmentVars

	objectiveLineFitSlack util_highs.ObjectiveIndex
	objectiveInclude      util_highs.ObjectiveIndex
	objectiveThresholds   util_highs.ObjectiveIndex
	objectiveLineOverlap  util_highs.ObjectiveIndex
}

type fitting2SegmentVars struct {
	process            *FittingSingleStatSegmentsProcess2
	lineSlope          util_highs.ColumnIndex
	lineOffset         util_highs.ColumnIndex
	minimumThreshold   util_highs.ColumnIndex
	maximumThreshold   util_highs.ColumnIndex
	includeSampleRow   util_highs.ConstraintRow
	includeSampleCount util_highs.ColumnIndex
	includeOverlapRow  util_highs.ConstraintRow
}

type Fitting2InterimResultSet struct {
	Segments []Fitting2InterimSegment
}

type Fitting2InterimSegment struct {
	LineSlope             float64
	LineOffset            float64
	StatRange             weight_types.StatRangeFloat
	IncludeCount          uint32
	IncludePercentOfTotal float64
	StopwatchSolver       util.Stopwatch
}

func (fg *FittingSingleStatSegmentsProcess2) Init(printer *util.PrintRecorder, timeout int) {
	fg.printer = printer
	fg.build = new(util_highs.LinearBuilder)
	fg.build.Minimise = true
	fg.build.Solver = util_highs.Solver_MIP_Interior
	fg.build.TimeLimitSeconds = timeout
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

func (fg *FittingSingleStatSegmentsProcess2) SupplyData(inputData []FittingSample) {
	fg.inputData = slices.Clone(inputData)
}

func (fg *FittingSingleStatSegmentsProcess2) Run() *util_async.FutureCancellable[Fitting2InterimResultSet] {
	for i := range c_fitting2_targetSegments {
		fg.addSegment(i == 0, i == c_fitting2_targetSegments-1)
	}
	for i := range len(fg.segments) - 1 {
		fg.enforceCrossSegmentRules(fg.segments[i], fg.segments[i+1])
	}
	fg.processData()
	for i := range len(fg.segments) {
		fg.finishSegment(fg.segments[i], i == len(fg.segments)-1)
	}

	future := fg.build.RunHighsFuture(&fg.stopwatch)
	return util_async.FutureCancellable_MapValue(future, func(res util_highs.LinearResult) (Fitting2InterimResultSet, bool) {
		solution := res.GetSolution2AndSaveLog(fg.printer)
		solution.DebugPrint(fg.printer)
		if solution.HasSolution() {
			return fg.prepareResult(solution), true
		} else {
			return Fitting2InterimResultSet{}, false
		}
	})
}

func (fg *FittingSingleStatSegmentsProcess2) addSegment(first, last bool) {
	fs := &fitting2SegmentVars{}
	fs.lineSlope = fg.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "slope"})
	fs.lineOffset = fg.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "offset"})

	if first {
		fs.minimumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, 0, util_highs.DebugString{Text: "minimum"})
	} else {
		fs.minimumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "minimum"})
	}
	if last {
		fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, c_fitting2_statScaledMaxValue, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})
	} else {
		fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})
	}

	// TODO this condition could also enforce a minimum range if we want one
	fg.build.ColumnIsGreaterOrEqualColumnEnforce(fs.minimumThreshold, fs.maximumThreshold)

	fg.segments = append(fg.segments, fs)
}

func (fg *FittingSingleStatSegmentsProcess2) finishSegment(segment *fitting2SegmentVars, isLast bool) {
	if !isLast {
		segment.includeOverlapRow.Build(fg.build, 1, util_highs.InfPos())
	}

	segment.includeSampleCount = fg.build.CreateColumnGeneral(highs.Integer, 0, util_highs.InfPos(), util_highs.DebugText("includeSampleCount"))
	segment.includeSampleRow.Add(segment.includeSampleCount, -1)
	segment.includeSampleRow.Build(fg.build, 0, 0)
}

func (fg *FittingSingleStatSegmentsProcess2) enforceCrossSegmentRules(one *fitting2SegmentVars, two *fitting2SegmentVars) {
	//thresholdCompareSlack := fg.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "thresholdCompareSlack"})
	//thresholdCompareSlackOutput := fg.build.CreateColumnWithObjective(highs.Continuous, 0, util_highs.InfPos(), 1, fg.objectiveThresholds, util_highs.DebugString{Text: "thresholdCompareSlackOutput"})

	//fg.build.AbsoluteValueFromDiffTwoVars(
	//	one.maximumThreshold, 1,
	//	two.minimumThreshold, 1,
	//	thresholdCompareSlackOutput,
	//	"compareThreshold")

	thresholdCompareSlack := fg.build.CreateColumnWithObjective(highs.Continuous, c_fitting2_minStatsRequireOverlapBetweenSegments, c_fitting2_maxStatsAllowOverlapBetweenSegments, 1, fg.objectiveThresholds, util_highs.DebugString{Text: "thresholdCompareSlack"})
	compareThreshold := util_highs.ConstraintRow{}
	compareThreshold.Add(one.maximumThreshold, 1)
	compareThreshold.Add(two.minimumThreshold, -1)
	compareThreshold.Add(thresholdCompareSlack, -1)
	compareThreshold.Build(fg.build, 0, 0)

	// enforce some range between segments
	//compareThreshold := util_highs.ConstraintRow{}
	//compareThreshold.Add(one.maximumThreshold, 1)
	//compareThreshold.Add(two.minimumThreshold, -1)
	//compareThreshold.Build(fg.build, c_fitting2_minStatsRequireOverlapBetweenSegments, c_fitting2_maxStatsAllowOverlapBetweenSegments)

	// simple equality
	//compareThreshold := util_highs.ConstraintRow{}
	//compareThreshold.Add(one.maximumThreshold, -1)
	//compareThreshold.Add(two.minimumThreshold, 1)
	//compareThreshold.Build(fg.build, 0, util_highs.InfPos())

	//thresholdCompareSlackOutput := fg.build.CreateColumnWithObjective(highs.Continuous, 0, util_highs.InfPos(), 1, fg.objectiveThresholds, util_highs.DebugString{Text: "thresholdCompareSlackOutput"})
	//fg.build.AbsoluteValue(thresholdCompareSlack, thresholdCompareSlackOutput)
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
}

func (fg *FittingSingleStatSegmentsProcess2) sampleIncludeToggleColumn(sample FittingSample, segment *fitting2SegmentVars) util_highs.ColumnIndex {
	includeColumn := fg.build.CreateColumnBoolWithObjective(1, fg.objectiveInclude, util_highs.DebugString{Text: "include"})

	fg.build.ConstantIsBetweenColumns_NoSequenceCheck(segment.minimumThreshold, segment.maximumThreshold, includeColumn, sample.StatValue, c_fitting2_statScaledHighM, c_fitting2_statScaledUnequalDelta)

	segment.includeSampleRow.Add(includeColumn, 1)
	return includeColumn
}

func (fg *FittingSingleStatSegmentsProcess2) prepareAsPotentialThreshold(seg1, seg2 *fitting2SegmentVars, include1, include2 util_highs.ColumnIndex, sample FittingSample) {
	includeBoth := fg.build.CreateColumnBool(util_highs.DebugText("includeBoth"))
	seg1.includeOverlapRow.Add(includeBoth, 1)
	fg.build.ConstraintAnd(includeBoth, include1, include2)

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

	fg.build.AbsoluteValue_WithToggle_NoExtraCheck(differenceSigned, difference, includeBoth, c_fitting_simScaledRangeHigh)
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

	fg.build.AbsoluteValue_WithToggle_NoExtraCheck(differenceSigned, difference, include, c_fitting_simScaledRangeHigh)
}

func (fg *FittingSingleStatSegmentsProcess2) prepareResult(solution *util_highs.Solution2) Fitting2InterimResultSet {
	resultSet := Fitting2InterimResultSet{}

	for _, segment := range fg.segments {
		interim := Fitting2InterimSegment{
			LineSlope:  solution.GetValue(segment.lineSlope),
			LineOffset: solution.GetValue(segment.lineOffset),
			StatRange: weight_types.StatRangeFloat{
				Minimum: solution.GetValue(segment.minimumThreshold),
				Maximum: solution.GetValue(segment.maximumThreshold),
			},
			IncludeCount:          solution.GetValueUInt32(segment.includeSampleCount),
			IncludePercentOfTotal: solution.GetValue(segment.includeSampleCount) / float64(len(fg.inputData)),
			StopwatchSolver:       util.Stopwatch{},
		}

		resultSet.Segments = append(resultSet.Segments, interim)
	}

	return resultSet
}
