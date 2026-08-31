package fitting3

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting2"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_fitting3_std_deviation_accept = 1.0

	c_fitting3_simScaledHighM     = 10.0
	c_fitting3_statScaledHighM    = 10.0
	c_fitting3_statScaledMaxValue = 1.0

	c_fitting3_segmentMinimumColumnIncludePercent = 0.10

	c_fitting3_output_lineFit         = 1
	c_fitting3_output_thresholdGap    = 50
	c_fitting3_output_missingMinimums = 1000
)

type FittingSingleSegmented3 struct {
	fitting2.BaseSingleSegmented[util_weight.FittingSample3]
}

func (ss *FittingSingleSegmented3) SupplyData(inputData []util_weight.FittingSample3) {
	ss.InputData = slices.Clone(inputData)
	slices.SortFunc(ss.InputData, func(a, b util_weight.FittingSample3) int {
		return cmp.Compare(a.StatValue, b.StatValue)
	})
}

func (ss *FittingSingleSegmented3) Run() *util_async.FutureCancellableWithError[fitting2.InitialResultSet] {
	// Disable presolve probing, very slow
	ss.Build.AddOptionInt("presolve_rule_off", 32768)

	ss.PrepareSegments(true, c_fitting3_statScaledMaxValue)
	for _, sample := range ss.InputData {
		ss.addSample(sample)
	}
	ss.FinishSegments(c_fitting3_segmentMinimumColumnIncludePercent, c_fitting3_output_missingMinimums)
	return ss.RunSolve()
}

func (ss *FittingSingleSegmented3) addSample(sample util_weight.FittingSample3) {
	ss.validateSample(sample)

	rowIncludeInOne := util_highs.ConstraintRow{}
	for _, segment := range ss.Segments {
		includeColumn := ss.SampleIncludeToggleColumn(sample.StatValue, segment, c_fitting3_statScaledHighM)
		ss.sampleToFitLine(sample, segment, includeColumn)
		rowIncludeInOne.Add(includeColumn, 1)
	}
	rowIncludeInOne.Build(ss.Build, 1, 2)

	for i := range len(ss.Segments) - 1 {
		ss.prepareAsPotentialThreshold(ss.Segments[i], ss.Segments[i+1], sample)
	}
}

func (ss *FittingSingleSegmented3) validateSample(sample util_weight.FittingSample3) {
	slack := 0.000001
	if sample.SimResult.Average < -slack || sample.SimResult.Average > 1+slack || sample.StatValue < -slack || sample.StatValue > 1+slack {
		panic(fmt.Sprintf("sample out of range %e %e", sample.SimResult.Average, sample.StatValue))
	}
}

func (ss *FittingSingleSegmented3) prepareAsPotentialThreshold(seg1, seg2 *fitting2.SegmentVars, sample util_weight.FittingSample3) {
	isThreshold := ss.PrepareThresholdColumn(seg1, sample.StatValue, c_fitting3_statScaledHighM)

	difference := ss.Build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), c_fitting3_output_thresholdGap, util_highs.DebugString{Text: "difference"})

	ss.Build.AbsoluteValueFromSumSeveral_WithToggle(
		[]util_highs.ColumnIndex{seg1.LineSlope, seg1.LineOffset, seg2.LineSlope, seg2.LineOffset},
		[]float64{sample.StatValue, 1, -sample.StatValue, -1},
		0,
		isThreshold,
		difference,
		c_fitting3_simScaledHighM,
	)
}
func (ss *FittingSingleSegmented3) sampleToFitLine(sample util_weight.FittingSample3, segment *fitting2.SegmentVars, include util_highs.ColumnIndex) {
	differenceSigned := ss.Build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "differenceSigned"})

	lo := sample.SimResult.Average - sample.SimResult.StdDev*c_fitting3_std_deviation_accept
	hi := sample.SimResult.Average + sample.SimResult.StdDev*c_fitting3_std_deviation_accept

	sampleRow := util_highs.ConstraintRow{Debug: "sampleRow"}
	sampleRow.Add(segment.LineSlope, sample.StatValue)
	sampleRow.Add(segment.LineOffset, 1)
	sampleRow.Add(differenceSigned, 1)
	sampleRow.Build(ss.Build, lo, hi)

	difference := ss.Build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), c_fitting3_output_lineFit, util_highs.DebugString{Text: "difference"})
	ss.Build.AbsoluteValue_WithToggle_NoExtraCheck(differenceSigned, difference, include, c_fitting3_simScaledHighM)
}
