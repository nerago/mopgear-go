package util_weight

import (
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_types"
)

type FittingSample struct {
	StatValue float64
	SimResult float64
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
