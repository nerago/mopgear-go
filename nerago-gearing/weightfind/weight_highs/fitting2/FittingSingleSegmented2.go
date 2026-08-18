package fitting2

import (
	"cmp"
	"fmt"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
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

type SingleSegmented2 struct {
	BaseSingleSegmented[util_weight.FittingSample]
}

func (ss *SingleSegmented2) SupplyData(inputData []util_weight.FittingSample) {
	ss.InputData = slices.Clone(inputData)
	slices.SortFunc(ss.InputData, func(a, b util_weight.FittingSample) int {
		return cmp.Compare(a.StatValue, b.StatValue)
	})
}

func (ss *SingleSegmented2) Run() *util_async.FutureCancellable[InitialResultSet] {
	ss.PrepareSegments(true)
	for _, sample := range ss.InputData {
		ss.addSample(sample)
	}
	ss.FinishSegments(true)
	return ss.RunSolve()
}

func (ss *SingleSegmented2) addSample(sample util_weight.FittingSample) {
	ss.validateSample(sample)

	rowIncludeInOne := util_highs.ConstraintRow{}
	for _, segment := range ss.Segments {
		includeColumn := ss.SampleIncludeToggleColumn(sample.StatValue, segment)
		ss.sampleToFitLine(sample, segment, includeColumn)
		rowIncludeInOne.Add(includeColumn, 1)
	}
	rowIncludeInOne.Build(ss.Build, 1, 2)

	for i := range len(ss.Segments) - 1 {
		ss.prepareAsPotentialThreshold(ss.Segments[i], ss.Segments[i+1], sample)
	}
}

func (ss *SingleSegmented2) validateSample(sample util_weight.FittingSample) {
	slack := 0.000001
	if sample.SimResult < -slack || sample.SimResult > 1+slack || sample.StatValue < -slack || sample.StatValue > 1+slack {
		panic(fmt.Sprintf("sample out of range %e %e", sample.SimResult, sample.StatValue))
	}
}

func (ss *SingleSegmented2) prepareAsPotentialThreshold(seg1, seg2 *SegmentVars, sample util_weight.FittingSample) {
	isThreshold := ss.PrepareThresholdColumn(seg1, sample.StatValue)

	difference := ss.Build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), c_fitting2_output_thresholdGap, util_highs.DebugString{Text: "difference"})

	ss.Build.AbsoluteValueFromSumSeveral_WithToggle(
		[]util_highs.ColumnIndex{seg1.LineSlope, seg1.LineOffset, seg2.LineSlope, seg2.LineOffset},
		[]float64{sample.StatValue, 1, -sample.StatValue, -1},
		0,
		isThreshold,
		difference,
		c_fitting2_simScaledHighM,
	)
}

func (ss *SingleSegmented2) sampleToFitLine(sample util_weight.FittingSample, segment *SegmentVars, include util_highs.ColumnIndex) {
	difference := ss.Build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), c_fitting2_output_lineFit, util_highs.DebugString{Text: "difference"})

	ss.Build.AbsoluteValueFromSumTwoThenDiffToConst_WithToggle(
		segment.LineSlope, sample.StatValue,
		segment.LineOffset, 1,
		sample.SimResult,
		include,
		difference,
		c_fitting2_simScaledHighM,
	)
}
