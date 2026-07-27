package weight_highs

import (
	"paladin_gearing_go/util/util_highs"

	"github.com/bartolsthoorn/gohighs/highs"
)

const c_fitting3_std_deviation_accept = 1.0

type FittingSample3 struct {
	StatValue float64
	SimResult FittingSimDetail
}

type FittingSimDetail struct {
	Average float64
	Min     float64
	Max     float64
	StdDev  float64
}

type FittingSingleStatSegmentsProcess3 struct {
	//printer   *util.PrintRecorder
	//stopwatch util.Stopwatch
	build *util_highs.LinearBuilder
	//timeout   int

	inputData []FittingSample3

	//segments []*fitting2SegmentVars
	//
	objectiveLineFitSlack util_highs.ObjectiveIndex
	//objectiveInclude      util_highs.ObjectiveIndex
	//objectiveThresholds   util_highs.ObjectiveIndex
	//objectiveLineOverlap  util_highs.ObjectiveIndex
}

func (fg *FittingSingleStatSegmentsProcess3) sampleToFitLine(sample FittingSample3, segment *fitting2SegmentVars, include util_highs.ColumnIndex) {
	differenceSigned := fg.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "differenceSigned"})
	difference := fg.build.CreateColumnWithObjective(highs.Continuous, 0, util_highs.InfPos(), 1, fg.objectiveLineFitSlack, util_highs.DebugString{Text: "difference"})

	lo := sample.SimResult.Average - sample.SimResult.StdDev*c_fitting3_std_deviation_accept
	hi := sample.SimResult.Average + sample.SimResult.StdDev*c_fitting3_std_deviation_accept

	// sampleRow: one.lineSlope*StatValue + one.lineOffset + difference = simResult
	sampleRow := util_highs.ConstraintRow{Debug: "sampleRow"}
	sampleRow.Add(segment.lineSlope, sample.StatValue)
	sampleRow.Add(segment.lineOffset, 1)
	sampleRow.Add(differenceSigned, 1)
	sampleRow.Build(fg.build, lo, hi)

	fg.build.AbsoluteValue_WithToggle_NoExtraCheck(differenceSigned, difference, include, c_fitting2_simScaledRangeHigh)
}
