package weight_types

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

type AccuracyInfoSimStatRanged struct {
	StatScore     float64
	SimScore      float64
	StatRankRange *util.HiLoInt
	SimRankRange  *util.HiLoInt
	DataSim       *stats.SimData
}

func (a *AccuracyInfoSimStatRanged) GetSimData() *stats.SimData {
	return a.DataSim
}

func (a *AccuracyInfoSimStatRanged) GetSimScore() float64 {
	return a.SimScore
}

func (a *AccuracyInfoSimStatRanged) IncrementSimScore(add float64) {
	a.SimScore += add
}

type AccuracyInfoPrePrepare struct {
	SimScore float64
	DataSim  *stats.SimData
	DataStat *stats.StatBlock
}

func (a *AccuracyInfoPrePrepare) GetSimData() *stats.SimData {
	return a.DataSim
}

func (a *AccuracyInfoPrePrepare) GetSimScore() float64 {
	return a.SimScore
}

func (a *AccuracyInfoPrePrepare) IncrementSimScore(add float64) {
	a.SimScore += add
}

type AccuracyPreparedEntry struct {
	StatScore    float64
	Stats        *stats.StatBlock
	SimRankRange *util.HiLoInt
}
