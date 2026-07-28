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
	c_fitting2_statScaledMaxValue = 1.0
	c_fitting2_simScaledHighM     = 2.0
	c_fitting2_statScaledHighM    = 2.0

	c_fitting2_segmentSizeMinimumStats = 750
	c_fitting2_segmentSizeMinimumCount = 0.10

	c_fitting2_output_lineFit      = 1
	c_fitting2_output_thresholdGap = 50
)

// //////////////////////////////////////////////////////
type SingleSegmented2 struct {
	printer   *util.PrintRecorder
	stopwatch util.Stopwatch
	build     *util_highs.LinearBuilder
	timeout   int

	scaleStat        float64
	unequalStatDelta float64
	inputData        []util_weight.FittingSample

	targetSegmentCount int
	segments           []*segmentVars
}

type segmentVars struct {
	process             *SingleSegmented2
	lineSlope           util_highs.ColumnIndex
	lineOffset          util_highs.ColumnIndex
	minimumThreshold    util_highs.ColumnIndex
	maximumThreshold    util_highs.ColumnIndex
	includeColumns      []util_highs.ColumnIndex
	thresholdColumns    []util_highs.ColumnIndex
	includeColumnRow    util_highs.ConstraintRow
	includeThresholdRow util_highs.ConstraintRow
	isFirst, isLast     bool
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
	fg.unequalStatDelta = fg.scaleStat * 0.1
	fg.printer = printer

	fg.build = new(util_highs.LinearBuilder)
	fg.build.Minimise = true
	fg.build.Solver = util_highs.Solver_MIP_Interior
	fg.build.TimeLimitSeconds = timeout
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

	if first || prev == nil {
		fs.minimumThreshold = -1
		fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})
	} else if last {
		fs.minimumThreshold = prev.maximumThreshold
		fs.maximumThreshold = -1
	} else {
		fs.minimumThreshold = prev.maximumThreshold
		fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})

		segmentSizeRow := util_highs.ConstraintRow{Debug: "segmentSizeRow"}
		segmentSizeRow.Add(fs.minimumThreshold, -1)
		segmentSizeRow.Add(fs.maximumThreshold, 1)
		segmentSizeRow.Build(fg.build, c_fitting2_segmentSizeMinimumStats*fg.scaleStat, util_highs.InfPos())
	}

	fg.segments = append(fg.segments, fs)
}

func (fg *SingleSegmented2) finishSegment(segment *segmentVars, isLast bool) {
	if !isLast {
		// in theory only have one overlap, but needs to be freer due to duplicates
		segment.includeThresholdRow.Build(fg.build, 1, util_highs.InfPos())
	}

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

	rowIncludeInOne := util_highs.ConstraintRow{}
	for _, segment := range fg.segments {
		includeColumn := fg.sampleIncludeToggleColumn(sample, segment)
		fg.sampleToFitLine(sample, segment, includeColumn)
		rowIncludeInOne.Add(includeColumn, 1)
	}
	rowIncludeInOne.Build(fg.build, 1, 2)

	for i := range len(fg.segments) - 1 {
		fg.prepareAsPotentialThreshold(fg.segments[i], fg.segments[i+1], sample)
	}
}

func (fg *SingleSegmented2) sampleIncludeToggleColumn(sample util_weight.FittingSample, segment *segmentVars) util_highs.ColumnIndex {
	includeColumn := fg.build.CreateColumnBool(util_highs.DebugString{Text: "include"})

	if segment.isFirst {
		fg.build.ColumnIsGreaterOrEqualThanConstant_Supplied(includeColumn, segment.maximumThreshold, sample.StatValue, c_fitting2_statScaledHighM, fg.unequalStatDelta)
	} else if segment.isLast {
		fg.build.ColumnIsLessOrEqualThanConstant_Supplied(includeColumn, segment.minimumThreshold, sample.StatValue, c_fitting2_statScaledHighM, fg.unequalStatDelta)
	} else {
		fg.build.ConstantIsBetweenColumns_NoSequenceCheck(segment.minimumThreshold, segment.maximumThreshold, includeColumn, sample.StatValue, c_fitting2_statScaledHighM, fg.unequalStatDelta)
	}

	segment.includeColumnRow.Add(includeColumn, 1)
	segment.includeColumns = append(segment.includeColumns, includeColumn)
	return includeColumn
}

func (fg *SingleSegmented2) prepareAsPotentialThreshold(seg1, seg2 *segmentVars, sample util_weight.FittingSample) {
	isThreshold := fg.build.CreateColumnBool(util_highs.DebugText("isThreshold"))
	fg.build.ColumnIsEqualConstant(seg1.maximumThreshold, isThreshold, sample.StatValue, c_fitting2_statScaledHighM, fg.unequalStatDelta)
	seg1.includeThresholdRow.Add(isThreshold, 1)
	seg1.thresholdColumns = append(seg1.thresholdColumns, isThreshold)

	difference := fg.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), c_fitting2_output_thresholdGap, util_highs.DebugString{Text: "difference"})

	fg.build.AbsoluteValueFromSumSeveral_WithToggle(
		[]util_highs.ColumnIndex{seg1.lineSlope, seg1.lineOffset, seg2.lineSlope, seg2.lineOffset},
		[]float64{sample.StatValue, 1, -sample.StatValue, -1},
		0,
		isThreshold,
		difference,
		c_fitting2_simScaledHighM,
	)
}

func (fg *SingleSegmented2) sampleToFitLine(sample util_weight.FittingSample, segment *segmentVars, include util_highs.ColumnIndex) {
	difference := fg.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), c_fitting2_output_lineFit, util_highs.DebugString{Text: "difference"})

	fg.build.AbsoluteValueFromSumTwoThenDiffToConst_WithToggle(
		segment.lineSlope, sample.StatValue,
		segment.lineOffset, 1,
		sample.SimResult,
		include,
		difference,
		c_fitting2_simScaledHighM,
	)
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
