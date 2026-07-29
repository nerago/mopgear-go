package util_weight

import (
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_types"
)

type FittingSample struct {
	StatValue float64
	SimResult float64
}

type FittingSample3 struct {
	StatValue float64
	SimResult FittingSimDetail
}

type FittingSimDetail struct {
	Average   float64
	Min       float64
	Max       float64
	StdDev    float64
	HasDetail bool
}

func (d FittingSimDetail) FlipMinMaxAsNeeded() {
	if d.Min >= d.Average && d.Average >= d.Max {
		d.Max, d.Min = d.Min, d.Max
	}
}

type FittingInterimResult struct {
	LineSlope                  float64
	LineOffset                 float64
	StatRange                  weight_types.StatRange
	IncludeCount               uint32
	IncludePercentOfStageInput float64
	IncludePercentOfTotal      float64
	BuiltSequence              []int
	StopwatchSolver            util.Stopwatch
}

type FittingInterimResult2 struct {
	LineSlope                  float64
	LineOffset                 float64
	StatRange                  weight_types.StatRange
	IncludeCount               uint32
	IncludePercentOfStageInput float64
	IncludePercentOfTotal      float64
	BuiltSequence              []int
}
