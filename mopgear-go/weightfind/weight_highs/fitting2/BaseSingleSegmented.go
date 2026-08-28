package fitting2

import (
	"math"

	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

type BaseSingleSegmented[S any] struct {
	Printer   *util.PrintRecorder
	Stopwatch util.Stopwatch
	Build     *util_highs.LinearBuilder
	timeout   int

	scaleStat        float64
	unequalStatDelta float64
	InputData        []S

	TargetSegmentCount int
	Segments           []*SegmentVars
}

type SegmentVars struct {
	//process             *SingleSegmented2
	LineSlope           util_highs.ColumnIndex
	LineOffset          util_highs.ColumnIndex
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
}

func (bss *BaseSingleSegmented[S]) Init(targetSegmentCount int, scaleStat float64, printer *util.PrintRecorder, timeout int) {
	bss.TargetSegmentCount = targetSegmentCount
	if targetSegmentCount <= 1 {
		panic("don't use this for 1 segment")
	}

	bss.scaleStat = scaleStat
	bss.unequalStatDelta = bss.scaleStat * 0.1
	bss.Printer = printer

	bss.Build = new(util_highs.LinearBuilder)
	bss.Build.Minimise = true
	bss.Build.Solver = util_highs.Solver_MIP_Interior
	bss.Build.TimeLimitSeconds = timeout
}

func (bss *BaseSingleSegmented[S]) PrepareSegments(enforceMinimumStatRange bool) {
	for i := range bss.TargetSegmentCount {
		if i == 0 {
			bss.addSegment(true, false, nil, enforceMinimumStatRange)
		} else if i == bss.TargetSegmentCount-1 {
			bss.addSegment(false, true, bss.Segments[i-1], enforceMinimumStatRange)
		} else {
			bss.addSegment(false, false, bss.Segments[i-1], enforceMinimumStatRange)
		}
	}
}

func (bss *BaseSingleSegmented[S]) FinishSegments(enforceMinimumIncludeCount bool) {
	for i := range len(bss.Segments) {
		bss.finishSegment(bss.Segments[i], i == len(bss.Segments)-1, enforceMinimumIncludeCount)
	}
}

func (bss *BaseSingleSegmented[S]) RunSolve() (*util_async.FutureCancellable[InitialResultSet], error) {
	future := bss.Build.RunHighsFuture(&bss.Stopwatch)
	return util_async.FutureCancellable_MapValue(future, func(res util_highs.LinearResult) (InitialResultSet, bool) {
		solution := res.GetSolution2AndSaveLog(bss.Printer)
		solution.DebugPrint(bss.Printer)
		if solution.HasSolution() {
			return bss.prepareResult(solution), true
		} else {
			return InitialResultSet{}, false
		}
	})
}

func (bss *BaseSingleSegmented[S]) addSegment(first, last bool, prev *SegmentVars, enforceMinimumStatRange bool) {
	fs := &SegmentVars{isFirst: first, isLast: last}
	fs.LineSlope = bss.Build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "slope"})
	fs.LineOffset = bss.Build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "offset"})

	if first || prev == nil {
		fs.minimumThreshold = -1
		fs.maximumThreshold = bss.Build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})
	} else if last {
		fs.minimumThreshold = prev.maximumThreshold
		fs.maximumThreshold = -1
	} else {
		fs.minimumThreshold = prev.maximumThreshold
		fs.maximumThreshold = bss.Build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledMaxValue, util_highs.DebugString{Text: "maximum"})

		if enforceMinimumStatRange {
			segmentSizeRow := util_highs.ConstraintRow{Debug: "segmentSizeRow"}
			segmentSizeRow.Add(fs.minimumThreshold, -1)
			segmentSizeRow.Add(fs.maximumThreshold, 1)
			segmentSizeRow.Build(bss.Build, c_fitting2_segmentSizeMinimumStats*bss.scaleStat, util_highs.InfPos())
		}
	}

	bss.Segments = append(bss.Segments, fs)
}

func (bss *BaseSingleSegmented[S]) finishSegment(segment *SegmentVars, isLast bool, enforceMinimumIncludeCount bool) {
	if !isLast {
		// in theory only have one overlap, but needs to be freer due to duplicates
		segment.includeThresholdRow.Build(bss.Build, 1, util_highs.InfPos())
	}

	if enforceMinimumIncludeCount {
		minimumColumnCount := c_fitting2_segmentSizeMinimumCount * float64(len(bss.InputData))
		segment.includeColumnRow.Build(bss.Build, math.Round(minimumColumnCount), util_highs.InfPos())
	}
}

func (bss *BaseSingleSegmented[S]) SampleIncludeToggleColumn(statValue float64, segment *SegmentVars) util_highs.ColumnIndex {
	includeColumn := bss.Build.CreateColumnBool(util_highs.DebugString{Text: "include"})

	if segment.isFirst {
		bss.Build.ColumnIsGreaterOrEqualThanConstant_Supplied(includeColumn, segment.maximumThreshold, statValue, c_fitting2_statScaledHighM, bss.unequalStatDelta)
	} else if segment.isLast {
		bss.Build.ColumnIsLessOrEqualThanConstant_Supplied(includeColumn, segment.minimumThreshold, statValue, c_fitting2_statScaledHighM, bss.unequalStatDelta)
	} else {
		bss.Build.ConstantIsBetweenColumns_NoSequenceCheck(segment.minimumThreshold, segment.maximumThreshold, includeColumn, statValue, c_fitting2_statScaledHighM, bss.unequalStatDelta)
	}

	segment.includeColumnRow.Add(includeColumn, 1)
	segment.includeColumns = append(segment.includeColumns, includeColumn)
	return includeColumn
}

func (bss *BaseSingleSegmented[S]) PrepareThresholdColumn(seg1 *SegmentVars, statValue float64) util_highs.ColumnIndex {
	isThreshold := bss.Build.CreateColumnBool(util_highs.DebugText("isThreshold"))
	bss.Build.ColumnIsEqualConstant(seg1.maximumThreshold, isThreshold, statValue, c_fitting2_statScaledHighM, bss.unequalStatDelta)
	seg1.includeThresholdRow.Add(isThreshold, 1)
	seg1.thresholdColumns = append(seg1.thresholdColumns, isThreshold)
	return isThreshold
}

func (bss *BaseSingleSegmented[S]) prepareResult(solution *util_highs.Solution2) InitialResultSet {
	resultSet := InitialResultSet{}

	for _, segment := range bss.Segments {
		var statRange weight_types.StatRangeFloat
		if segment.minimumThreshold != -1 {
			statRange.Minimum = solution.GetValue(segment.minimumThreshold)
		}
		if segment.maximumThreshold != -1 {
			statRange.Maximum = solution.GetValue(segment.maximumThreshold)
		} else {
			statRange.Maximum = math.MaxUint32
		}

		includeSampleCount := bss.countSamples(segment, solution)
		interim := InitialSegment{
			LineSlope:             solution.GetValue(segment.LineSlope),
			LineOffset:            solution.GetValue(segment.LineOffset),
			StatRange:             statRange,
			IncludeCount:          uint32(includeSampleCount),
			IncludePercentOfTotal: float64(includeSampleCount) / float64(len(bss.InputData)),
		}

		resultSet.Segments = append(resultSet.Segments, interim)
	}

	return resultSet
}

func (bss *BaseSingleSegmented[S]) countSamples(segment *SegmentVars, solution *util_highs.Solution2) int {
	count := 0
	for _, column := range segment.includeColumns {
		if solution.ValueIsOne(column) {
			count++
		}
	}
	return count
}
