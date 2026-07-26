package weight_highs

import (
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_fitting2_targetSegments      = 4
	c_fitting2_statScaledRangeHigh = 1.0
	c_fitting2_simScaledRangeHigh  = 1.0

	c_fitting2_maxStatsGapBetweenSegments          = 0.02 // similar to dropped range in fitting1
	c_fitting2_maxStatsAllowOverlapBetweenSegments = 0.01 // 1% is about 150 stat given standard 15000 range
	c_fitting2_output_thresholdCompareSlack        = 1
	c_fitting2_outputFittingPerInclude

	c_fitting2_statUnscaledHigh       = 50000
	c_fitting2_statScaledUnequalDelta = c_fitting2_statScaledRangeHigh / float64(c_fitting2_statUnscaledHigh)
)

type FittingSingleStatSegmentsProcess2 struct {
	printer   *util.PrintRecorder
	stopwatch util.Stopwatch
	build     *util_highs.LinearBuilder
	timeout   int

	inputData []FittingSample

	segments []*fittingSingleSegment2

	objectiveLineDiff   util_highs.ObjectiveIndex
	objectiveInclude    util_highs.ObjectiveIndex
	objectiveThresholds util_highs.ObjectiveIndex
}

type fittingSingleSegment2 struct {
	process          *FittingSingleStatSegmentsProcess2
	lineSlope        util_highs.ColumnIndex
	lineOffset       util_highs.ColumnIndex
	minimumThreshold util_highs.ColumnIndex
	maximumThreshold util_highs.ColumnIndex
	includeColumns   []util_highs.ColumnIndex
	includeCountRow  util_highs.ConstraintRow
}

type Fitting2InterimResult struct {
	LineSlope             float64
	LineOffset            float64
	StatRange             weight_types.StatRange
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
}

func (fg *FittingSingleStatSegmentsProcess2) SupplyData(inputData []FittingSample) {
	fg.inputData = slices.Clone(inputData)
}

func (fg *FittingSingleStatSegmentsProcess2) Run() *util_async.FutureCancellable[Fitting2InterimResult] {
	for range c_fitting2_targetSegments {
		fg.addSegment()
	}
	fg.enforceFirstSegment(fg.segments[0])
	for i := range len(fg.segments) - 1 {
		fg.enforceCrossSegmentRules(fg.segments[i], fg.segments[i+1])
	}
	fg.enforceLastSegment(fg.segments[len(fg.segments)-1])

	for _, sample := range fg.inputData {
		fg.addSample(sample)
	}

	//fw.includeCountRow.Build(fw.build, float64(len(fw.inputData))*fw.minimumIncludeRate, util_highs.C_PlusInf)
	return nil
}

func (fg *FittingSingleStatSegmentsProcess2) addSegment() {
	fs := &fittingSingleSegment2{}
	fs.lineSlope = fg.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugString{Text: "slope"})
	fs.lineOffset = fg.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugString{Text: "offset"})
	fs.minimumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledRangeHigh, util_highs.DebugString{Text: "minimum"})
	fs.maximumThreshold = fg.build.CreateColumnGeneral(highs.Continuous, 0, c_fitting2_statScaledRangeHigh, util_highs.DebugString{Text: "maximum"})

	// TODO this condition could also enforce a minimum range if we want one
	fg.build.ColumnIsGreaterOrEqualColumnEnforce(fs.minimumThreshold, fs.maximumThreshold)

	fg.segments = append(fg.segments, fs)
}

func (fg *FittingSingleStatSegmentsProcess2) enforceCrossSegmentRules(one *fittingSingleSegment2, two *fittingSingleSegment2) {
	thresholdCompareSlack := fg.build.CreateColumnGeneral(highs.Continuous, c_fitting2_maxStatsGapBetweenSegments, c_fitting2_maxStatsAllowOverlapBetweenSegments, util_highs.DebugString{Text: "thresholdCompareSlack"})
	compareThreshold := util_highs.ConstraintRow{}
	compareThreshold.Add(one.maximumThreshold, -1)
	compareThreshold.Add(two.minimumThreshold, 1)
	compareThreshold.Add(thresholdCompareSlack, 1)
	compareThreshold.Build(fg.build, 0, 0)

	thresholdCompareSlackOutput := fg.build.CreateColumnWithObjective(highs.Continuous, 0, util_highs.C_PlusInf, c_fitting2_output_thresholdCompareSlack, fg.objectiveThresholds, util_highs.DebugString{Text: "thresholdCompareSlackOutput"})
	fg.build.AbsoluteValue(thresholdCompareSlack, thresholdCompareSlackOutput)

	// TODO also lines should meet
	// simA=slopeA*threshold+offsetA       simB=slopeB*threshold+offsetB
	// heightDiff = simA-simB = slopeA*threshold+offsetA - (slopeB*threshold+offsetB)
	//                        = slopeA*threshold + offsetA - slopeB*threshold - offsetB
	//                        = (slopeA-slopeB)*threshold + offsetA - offsetB
	//                      0 = (slopeA-slopeB)*threshold + offsetA - offsetB - heightDiff
	//                      0 = (slopeA-slopeB)*threshold + stuffCols
	//

	// simA=slopeA*threshold+offsetA       simB=slopeB*threshold+offsetB
	// threshold=(simA-offset)/slopeA      simB=slopeB*((simA-offsetA)/slopeA)+offsetB
	//                                simB-offsetB=slopeB*((simA-offsetA)/slopeA)
	//                                  (simB-offsetB)*slopeA = slopeB*(simA-offsetA)
	//                                  (sim-offsetB)*slopeA = slopeB*(sim-offsetA)
	//

	// this is kinda more a where do the lines meet question
	//     simA-simB = 0 = slopeA*threshold+offsetA - (slopeB*threshold+offsetB)
	//                 0 = slopeA*threshold + offsetA - slopeB*threshold - offsetB
	// offsetB - offsetA = (slopeA-slopeB)*threshold
	//         threshold = (offsetB - offsetA) / (slopeA - slopeB)

	//compareHeightAtJoin := util_highs.ConstraintRow{}
	//compareHeightAtJoin.Add(one.lineOffset, ?)
	//compareHeightAtJoin.Add(one.lineSlope, ?)

	// this is all looking like non-linear steps, maybe going about this wrong
	// would enforced overlap do enough to help??
	// or some kinda multi round solve where I have more control than the linear objectives...
	// could use values from last round, pick the thresholds then optimize the slopes around that

	//
	// what if we have forced overlap segments with double bool includes ANDed
	// so normally during an overlap we're separately checking:
	// sampleRow: one.lineSlope*StatValue + one.lineOffset + difference = simResult
	// sampleRow: two.lineSlope*StatValue + two.lineOffset + difference = simResult
	// how about as well or instead
	// overlapRow: one.lineSlope*StatValue + one.lineOffset - two.lineSlope*StatValue - two.lineOffset + difference = 0

	// so we could get to a situation where if include=1 then sampleRow, include=1&&2 then overlapRow. or both
	// but ANDing every combination is horrible var explosion
	// or maybe its just sample*segment? we're already checking for include in each,
	// AND between segments would just be adjacent ones
	// we aren't sorting stats but easily could be, then adopt some idea that a particular sample is the pivot of segments
}

func (fg *FittingSingleStatSegmentsProcess2) enforceFirstSegment(segment *fittingSingleSegment2) {
	rowLimit := util_highs.ConstraintRow{}
	rowLimit.Add(segment.minimumThreshold, 1)
	rowLimit.Build(fg.build, 0, 0)
}

func (fg *FittingSingleStatSegmentsProcess2) enforceLastSegment(segment *fittingSingleSegment2) {
	rowLimit := util_highs.ConstraintRow{}
	rowLimit.Add(segment.maximumThreshold, 1)
	rowLimit.Build(fg.build, c_fitting2_statScaledRangeHigh, c_fitting2_statScaledRangeHigh)
}

func (fg *FittingSingleStatSegmentsProcess2) addSample(sample FittingSample) {
	if sample.SimResult < 0 || sample.SimResult > 1 || sample.StatValue < 0 || sample.StatValue > 1 {
		panic("sample out of range")
	}

	for _, segment := range fg.segments {
		includeColumn := fg.sampleIncludeToggleColumn(sample, segment)
		fg.sampleToFitLine(sample, segment, includeColumn)
	}
}

func (fg *FittingSingleStatSegmentsProcess2) sampleIncludeToggleColumn(sample FittingSample, segment *fittingSingleSegment2) util_highs.ColumnIndex {
	includeColumn := fg.build.CreateColumnBoolWithObjective(c_fitting2_outputFittingPerInclude, fg.objectiveInclude, util_highs.DebugString{Text: "include"})
	segment.includeCountRow.Add(includeColumn, 1)
	segment.includeColumns = append(segment.includeColumns, includeColumn)

	fg.build.ConstantIsBetweenColumns_NoSequenceCheck(segment.minimumThreshold, segment.maximumThreshold, includeColumn, sample.StatValue, c_fitting2_statScaledRangeHigh, c_fitting2_statScaledUnequalDelta)
	return includeColumn
}

func (fg *FittingSingleStatSegmentsProcess2) sampleToFitLine(sample FittingSample, segment *fittingSingleSegment2, toggle util_highs.ColumnIndex) {
	//// this would need another way to include toggle AbsoluteValueFromSumTwoDiffConst_WithToggle may be possible
	//difference := fg.build.CreateColumnWithObjective(highs.Continuous, 0, util_highs.C_PlusInf, c_fitting_outputDifference, fg.objectiveLineDiff, util_highs.DebugString{Text: "difference"})
	//fg.build.AbsoluteValueFromSumTwoThenDiffToConst(
	//	segment.lineSlope, sample.StatValue,
	//	segment.lineOffset, 1,
	//	sample.SimResult,
	//	difference,
	//	"sampleFit")

	differenceSigned := fg.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugString{Text: "differenceSigned"})
	difference := fg.build.CreateColumnWithObjective(highs.Continuous, 0, util_highs.C_PlusInf, c_fitting_outputDifference, fg.objectiveLineDiff, util_highs.DebugString{Text: "difference"})

	// sampleRow: one.lineSlope*StatValue + one.lineOffset + difference = simResult
	sampleRow := util_highs.ConstraintRow{Debug: "sampleRow"}
	sampleRow.Add(segment.lineSlope, sample.StatValue)
	sampleRow.Add(segment.lineOffset, 1)
	sampleRow.Add(difference, 1)
	sampleRow.Build(fg.build, sample.SimResult, sample.SimResult)

	fg.build.AbsoluteValue_WithToggle_NoExtraCheck(differenceSigned, difference, toggle, c_fitting_simScaledRangeHigh)
}
