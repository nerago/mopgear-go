package weight_types

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_collection"
)

type AccuracyInfoSimStatRanged struct {
	StatScore     float64
	SimScore      float64
	StatRankRange *util_collection.HiLoInt
	SimRankRange  *util_collection.HiLoInt
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

func (a *AccuracyInfoSimStatRanged) SetSimRankRange(targetRange *util_collection.HiLoInt) {
	a.SimRankRange = targetRange
}

func (a *AccuracyInfoSimStatRanged) GetSimRankRange() *util_collection.HiLoInt {
	return a.SimRankRange
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
	SimRankRange *util_collection.HiLoInt
}
