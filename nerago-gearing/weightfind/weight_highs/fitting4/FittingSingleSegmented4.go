package fitting4

import (
	"cmp"
	"fmt"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting2"
	"math"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_fitting4_std_deviation_accept = 1.0

	c_fitting4_simScaledHighM     = 2.0
	c_fitting4_statScaledHighM    = 2.0
	c_fitting4_statScaledMaxValue = 1.0

	c_fitting4_output_lineFit      = 1
	c_fitting4_output_thresholdGap = 50

	c_fitting4_segmentSizeMinimumTarget = 0.1
)

type FittingSingleSegmented4 struct {
	fitting2.BaseSingleSegmented[util_weight.FittingSample3]
	segmentOnData bool
}

func (ss *FittingSingleSegmented4) SupplyData(inputData []util_weight.FittingSample3) {
	ss.InputData = slices.Clone(inputData)
	slices.SortFunc(ss.InputData, func(a, b util_weight.FittingSample3) int {
		return cmp.Compare(a.StatValue, b.StatValue)
	})
}

// Two alternates, segment based on stat ranges, segment based on sample indexes
func (ss *FittingSingleSegmented4) Run() *util_async.FutureCancellable[fitting2.InitialResultSet] {
	ss.PrepareSegments(false)

	if ss.segmentOnData {
		ss.splitInputDataEvenlyBetweenSegments()
	} else {
		ss.splitStatRangeEvenlyBetweenSegments()
	}

	ss.FinishSegments(false)
	return ss.RunSolve()
}

func (ss *FittingSingleSegmented4) splitInputDataEvenlyBetweenSegments() {
	splitPeriod := len(ss.InputData) / ss.TargetSegmentCount
	for segmentNum := range len(ss.Segments) {
		firstIndex := segmentNum * splitPeriod
		lastIndex := (segmentNum+1)*splitPeriod - 1

		segment := ss.Segments[segmentNum]
		for i := firstIndex; i <= lastIndex; i++ {
			ss.addSample(ss.InputData[i], segment)
		}

		if segmentNum < ss.TargetSegmentCount-1 {
			lastSample := ss.InputData[lastIndex]
			ss.prepareAsThreshold(ss.Segments[segmentNum], ss.Segments[segmentNum+1], lastSample)
		}
	}
}

func (ss *FittingSingleSegmented4) splitStatRangeEvenlyBetweenSegments() {
	minValue := ss.InputData[0].StatValue
	maxValue := ss.InputData[len(ss.InputData)-1].StatValue
	valueRange := maxValue - minValue
	valueInterval := valueRange / float64(ss.TargetSegmentCount)

	dataBySegment := make([]util_collection.List[util_weight.FittingSample3], ss.TargetSegmentCount)
	for _, sample := range ss.InputData {
		floatyIndex := (sample.StatValue - minValue) / valueInterval
		index := util.Clamp(int(math.Floor(floatyIndex)), 0, len(dataBySegment)-1)
		dataBySegment[index].AppendLast(sample)
	}

	minimumSampleCount := int(math.Floor(c_fitting4_segmentSizeMinimumTarget * float64(len(ss.InputData))))
	for blockIndex := range len(dataBySegment) - 1 {
		balanceBlocks(&dataBySegment[blockIndex], &dataBySegment[blockIndex+1], minimumSampleCount)
	}

	for segmentNum := range len(ss.Segments) {
		data := dataBySegment[segmentNum]
		segment := ss.Segments[segmentNum]
		for sample := range data.SeqValues() {
			ss.addSample(sample, segment)
		}

		if segmentNum < ss.TargetSegmentCount-1 {
			lastSample, _ := data.GetLast()
			ss.prepareAsThreshold(ss.Segments[segmentNum], ss.Segments[segmentNum+1], lastSample)
		}
	}
}

func balanceBlocks(first *util_collection.List[util_weight.FittingSample3], second *util_collection.List[util_weight.FittingSample3], targetMinCount int) {
	if first.Size() < targetMinCount && first.Size() < second.Size() {
		for first.Size() < targetMinCount && first.Size() < second.Size() {
			if sample, hasFirst := second.RemoveFirstAndReturn(); hasFirst {
				first.AppendLast(sample)
			} else {
				break
			}
		}
	} else if second.Size() < targetMinCount {
		for second.Size() < targetMinCount && second.Size() < first.Size() {
			if sample, hasLast := first.RemoveLastAndReturn(); hasLast {
				second.InsertFirst(sample)
			} else {
				break
			}
		}
	}
}

func (ss *FittingSingleSegmented4) addSample(sample util_weight.FittingSample3, explicitSegment *fitting2.SegmentVars) {
	ss.validateSample(sample)

	includeColumn := ss.SampleIncludeToggleColumn(sample.StatValue, explicitSegment)
	ss.sampleToFitLine(sample, explicitSegment, includeColumn)
}

func (ss *FittingSingleSegmented4) validateSample(sample util_weight.FittingSample3) {
	slack := 0.000001
	if sample.SimResult.Average < -slack || sample.SimResult.Average > 1+slack || sample.StatValue < -slack || sample.StatValue > 1+slack {
		panic(fmt.Sprintf("sample out of range %e %e", sample.SimResult.Average, sample.StatValue))
	}
}

func (ss *FittingSingleSegmented4) prepareAsThreshold(seg1, seg2 *fitting2.SegmentVars, sample util_weight.FittingSample3) {
	isThreshold := ss.PrepareThresholdColumn(seg1, sample.StatValue)

	difference := ss.Build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), c_fitting4_output_thresholdGap, util_highs.DebugString{Text: "difference"})

	ss.Build.AbsoluteValueFromSumSeveral_WithToggle(
		[]util_highs.ColumnIndex{seg1.LineSlope, seg1.LineOffset, seg2.LineSlope, seg2.LineOffset},
		[]float64{sample.StatValue, 1, -sample.StatValue, -1},
		0,
		isThreshold,
		difference,
		c_fitting4_simScaledHighM,
	)
}

func (ss *FittingSingleSegmented4) sampleToFitLine(sample util_weight.FittingSample3, segment *fitting2.SegmentVars, include util_highs.ColumnIndex) {
	differenceSigned := ss.Build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "differenceSigned"})

	lo := sample.SimResult.Average - sample.SimResult.StdDev*c_fitting4_std_deviation_accept
	hi := sample.SimResult.Average + sample.SimResult.StdDev*c_fitting4_std_deviation_accept

	sampleRow := util_highs.ConstraintRow{Debug: "sampleRow"}
	sampleRow.Add(segment.LineSlope, sample.StatValue)
	sampleRow.Add(segment.LineOffset, 1)
	sampleRow.Add(differenceSigned, 1)
	sampleRow.Build(ss.Build, lo, hi)

	difference := ss.Build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), c_fitting4_output_lineFit, util_highs.DebugString{Text: "difference"})
	ss.Build.AbsoluteValue_WithToggle_NoExtraCheck(differenceSigned, difference, include, c_fitting4_simScaledHighM)
}
