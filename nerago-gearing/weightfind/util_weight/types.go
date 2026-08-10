package util_weight

import (
	"math"
	"paladin_gearing_go/stats"
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

func MakeFittingDetail(average float64, data *stats.SimDataDetail, scaleSim ScaleAndOffset) FittingSimDetail {
	fit := FittingSimDetail{
		Average:   scaleSim.Apply(average),
		Min:       scaleSim.Apply(data.Min),
		Max:       scaleSim.Apply(data.Max),
		StdDev:    math.Abs(scaleSim.Scale) * data.StdDev,
		HasDetail: true,
	}
	fit.FlipMinMaxAsNeeded()
	return fit
}

func MakeFittingDetailFromAverage(average float64, scaleSim ScaleAndOffset) FittingSimDetail {
	return FittingSimDetail{
		Average:   scaleSim.Apply(average),
		HasDetail: false,
	}
}

func (d *FittingSimDetail) FlipMinMaxAsNeeded() {
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
