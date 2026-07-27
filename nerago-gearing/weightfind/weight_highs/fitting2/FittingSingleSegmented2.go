package fitting2

import (
	"cmp"
	"math"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/util_weight"
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

	c_fitting2_segmentSizeMinimumStats = 1000
	c_fitting2_segmentSizeMinimumCount = 0.10

	c_fitting2_minimum_stat_coverage       = 100
	c_fitting2_permitted_overlap_fix       = 50
	c_fitting2_number_nice_number_interval = 5
)

// //////////////////////////////////////////////////////
type SingleSegmented2 struct {
	printer   *util.PrintRecorder
	stopwatch util.Stopwatch
	build     *util_highs.LinearBuilder
	timeout   int

	scaleStat float64
	inputData []util_weight.FittingSample

	targetSegmentCount int
	segments           []*segmentVars

	objectiveLineFitSlack util_highs.ObjectiveIndex
	objectiveThresholds   util_highs.ObjectiveIndex
	objectiveLineOverlap  util_highs.ObjectiveIndex
}

type segmentVars struct {
	process           *SingleSegmented2
	lineSlope         util_highs.ColumnIndex
	lineOffset        util_highs.ColumnIndex
	minimumThreshold  util_highs.ColumnIndex
	maximumThreshold  util_highs.ColumnIndex
	includeColumns    []util_highs.ColumnIndex
	includeColumnRow  util_highs.ConstraintRow
	includeOverlapRow util_highs.ConstraintRow
	isFirst, isLast   bool
}

type InitialResultSet struct {
	Segments []InitialSegment
}

type InitialSegment struct {
	LineSlope             float64
	LineOffset            float64
	StatRange             weight_types.StatRangeFloat
	IncludeCount          uint32
	IncludePercentOfTotal float64
	StopwatchSolver       util.Stopwatch
}

func (fg *SingleSegmented2) Init(targetSegmentCount int, scaleStat float64, printer *util.PrintRecorder, timeout int) {
	fg.targetSegmentCount = targetSegmentCount
	if targetSegmentCount <= 1 {
		panic("don't use this for 1 segment")
	}

	fg.scaleStat = scaleStat
	fg.printer = printer

	fg.build = new(util_highs.LinearBuilder)
	fg.build.Minimise = true
	fg.build.Solver = util_highs.Solver_MIP_Interior
	fg.build.TimeLimitSeconds = timeout

	fg.setupObjectives()
}

func (fg *SingleSegmented2) setupObjectives() {
	fg.build.BlendMultiObjectives = true
	// regular line fitting, compare to scaled sim, average 0.01 but multiplied by 500(N), so expect total about 5.0
	fg.objectiveLineFitSlack = fg.build.AddObjectiveBlended(1, 0)
	// similar scale numbers to regular line fitting, only one per segment, but only moderate importance
	fg.objectiveLineOverlap = fg.build.AddObjectiveBlended(500, 0)
	// smallish numbers need a boost up to 0.01 just per segment, and is important
	fg.objectiveThresholds = fg.build.AddObjectiveBlended(1000, 0)
}

func (fg *SingleSegmented2) setupObjectives2() {
	fg.build.BlendMultiObjectives = false

	// regular line fitting, compare to scaled sim, average 0.01 but multiplied by 500(N), so expect total about 5.0
	fg.objectiveLineFitSlack = fg.build.AddObjectivePrioritised(false, -1, 0.3, 4)
	// smallish numbers need a boost up to 0.01 just per segment, and is important
	fg.objectiveThresholds = fg.build.AddObjectivePrioritised(false, -1, 0.1, 2)
	// similar scale numbers to regular line fitting, only one per segment, but only moderate importance
	fg.objectiveLineOverlap = fg.build.AddObjectivePrioritised(false, -1, -1, 1)
}

func (fg *SingleSegmented2) SupplyData(inputData []util_weight.FittingSample) {
	fg.inputData = slices.Clone(inputData)
	slices.SortFunc(fg.inputData, func(a, b util_weight.FittingSample) int {
		return cmp.Compare(a.StatValue, b.StatValue)
	})
}

func (fg *SingleSegmented2) Run() *util_async.FutureCancellable[InitialResultSet] {
	for i := range fg.targetSegmentCount {
		if i == 0 {
			fg.addSegment(true, false, nil)
		} else if i == fg.targetSegmentCount-1 {
			fg.addSegment(false, true, fg.segments[i-1])
		} else {
			fg.addSegment(false, false, fg.segments[i-1])
		}
	}
	for _, sample := range fg.inputData {
		fg.addSample(sample)
	}
	for i := range len(fg.segments) {
		fg.finishSegment(fg.segments[i], i == len(fg.segments)-1)
	}

	future := fg.build.RunHighsFuture(&fg.stopwatch)
	return util_async.FutureCancellable_MapValue(future, func(res util_highs.LinearResult) (InitialResultSet, bool) {
		solution := res.GetSolution2AndSaveLog(fg.printer)
		solution.DebugPrint(fg.printer)
		if solution.HasSolution() {
			return fg.prepareResult(solution), true
		} else {
			return InitialResultSet{}, false
		}
	})
}

func (fg *SingleSegmented2) addSegment(first, last bool, prev *segmentVars) {
	fs := &segmentVars{isFirst: first, isLast: last}
	fs.lineSlope = fg.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "slope"})
	fs.lineOffset = fg.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "offset"})

	//if first {
	//	fs.minimumThreshold = -1
	//	fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})
	//} else if last {
	//	fs.minimumThreshold = prev.maximumThreshold
	//	fs.maximumThreshold = -1
	//} else {
	//	fs.minimumThreshold = prev.maximumThreshold
	//	fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})

	if first {
		//fs.minimumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, 0, util_highs.DebugString{Text: "minimum"})
		fs.minimumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "minimum"})
		fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})
	} else if last {
		fs.minimumThreshold = prev.maximumThreshold
		//fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, c_fitting2_statScaledMaxValue, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})
		fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})
	} else {
		fs.minimumThreshold = prev.maximumThreshold
		fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})

		//fg.build.ColumnIsGreaterOrEqualColumnEnforce(fs.minimumThreshold, fs.maximumThreshold)

		//row := util_highs.ConstraintRow{Debug: "ColumnIsGreaterOrEqualColumnEnforce"}
		//row.Add(fs.minimumThreshold, -1)
		//row.Add(fs.maximumThreshold, 1)
		//row.Build(fg.build, 0, util_highs.InfPos())

		//segmentSizeRow := util_highs.ConstraintRow{Debug: "segmentSizeRow"}
		//segmentSizeRow.Add(fs.minimumThreshold, -1)
		//segmentSizeRow.Add(fs.maximumThreshold, 1)
		////segmentSizeRow.Build(fg.build, 0, util_highs.InfPos())
		//segmentSizeRow.Build(fg.build, c_fitting2_segmentSizeMinimumStats*fg.scaleStat, util_highs.InfPos())
	}

	segmentSizeRow := util_highs.ConstraintRow{Debug: "segmentSizeRow"}
	segmentSizeRow.Add(fs.minimumThreshold, -1)
	segmentSizeRow.Add(fs.maximumThreshold, 1)
	//segmentSizeRow.Build(fg.build, 0, util_highs.InfPos())
	segmentSizeRow.Build(fg.build, c_fitting2_segmentSizeMinimumStats*fg.scaleStat, util_highs.InfPos())

	fg.segments = append(fg.segments, fs)
}

func (fg *SingleSegmented2) finishSegment(segment *segmentVars, isLast bool) {
	if !isLast {
		segment.includeOverlapRow.Build(fg.build, 1, 1)
	}

	// I want this but makes infeasible
	minimumColumnCount := c_fitting2_segmentSizeMinimumCount * float64(len(fg.inputData))
	segment.includeColumnRow.Build(fg.build, math.Round(minimumColumnCount), util_highs.InfPos())
}

func validateSample(sample util_weight.FittingSample) {
	if sample.SimResult < 0 || sample.SimResult > 1 || sample.StatValue < 0 || sample.StatValue > 1 {
		panic("sample out of range")
	}
}

func (fg *SingleSegmented2) addSample(sample util_weight.FittingSample) {
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

func (fg *SingleSegmented2) sampleIncludeToggleColumn(sample util_weight.FittingSample, segment *segmentVars) util_highs.ColumnIndex {
	includeColumn := fg.build.CreateColumnBool(util_highs.DebugString{Text: "include"})

	//unequalDelta := fg.scaleStat * 0
	unequalDelta := fg.scaleStat * 0.1
	//unequalDelta := c_fitting2_statScaledUnequalDelta

	//if segment.isFirst {
	//	fg.build.ColumnIsGreaterOrEqualThanConstant_Supplied(includeColumn, segment.maximumThreshold, sample.StatValue, c_fitting2_statScaledHighM, unequalDelta)
	//} else if segment.isLast {
	//	fg.build.ColumnIsLessOrEqualThanConstant_Supplied(includeColumn, segment.minimumThreshold, sample.StatValue, c_fitting2_statScaledHighM, unequalDelta)
	//} else {
	//	fg.build.ConstantIsBetweenColumns_NoSequenceCheck(segment.minimumThreshold, segment.maximumThreshold, includeColumn, sample.StatValue, c_fitting2_statScaledHighM, unequalDelta)
	//}
	fg.build.ConstantIsBetweenColumns_NoSequenceCheck(segment.minimumThreshold, segment.maximumThreshold, includeColumn, sample.StatValue, c_fitting2_statScaledHighM, unequalDelta)

	segment.includeColumnRow.Add(includeColumn, 1)
	segment.includeColumns = append(segment.includeColumns, includeColumn)
	return includeColumn
}

func (fg *SingleSegmented2) prepareAsPotentialThreshold(seg1, seg2 *segmentVars, include1, include2 util_highs.ColumnIndex, sample util_weight.FittingSample) {
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

func (fg *SingleSegmented2) sampleToFitLine(sample util_weight.FittingSample, segment *segmentVars, include util_highs.ColumnIndex) {
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

func (fg *SingleSegmented2) prepareResult(solution *util_highs.Solution2) InitialResultSet {
	resultSet := InitialResultSet{}

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
		interim := InitialSegment{
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

func (fg *SingleSegmented2) countSamples(segment *segmentVars, solution *util_highs.Solution2) int {
	count := 0
	for _, column := range segment.includeColumns {
		if solution.ValueIsOne(column) {
			count++
		}
	}
	return count
}
