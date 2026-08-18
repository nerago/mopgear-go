package weight_types

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util/util_collection"
)

type AccuracyInfo struct {
	StatScore     float64
	SimScore      float64
	StatRankRange *util_collection.HiLoInt
	SimRankRange  *util_collection.HiLoInt
	DataSim       *stats.SimData
}

func (a *AccuracyInfo) GetSimData() *stats.SimData {
	return a.DataSim
}

func (a *AccuracyInfo) GetSimScore() float64 {
	return a.SimScore
}

func (a *AccuracyInfo) IncrementSimScore(add float64) {
	a.SimScore += add
}

func (a *AccuracyInfo) ResetSimScore() {
	a.SimScore = 0
}

func (a *AccuracyInfo) SetSimRankRange(targetRange *util_collection.HiLoInt) {
	a.SimRankRange = targetRange
}

func (a *AccuracyInfo) GetSimRankRange() *util_collection.HiLoInt {
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

func (a *AccuracyInfoPrePrepare) ResetSimScore() {
	a.SimScore = 0
}

type AccuracyInfoPrepared struct {
	StatScore    float64
	Stats        *stats.StatBlock
	SimRankRange *util_collection.HiLoInt
}
