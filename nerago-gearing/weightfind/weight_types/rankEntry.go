package weight_types

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

type IRankEntry interface {
	GetSimData() *stats.SimData
}

type IRankEntryFlat interface {
	IRankEntry
	GetSimScore() float64
	IncrementSimScore(add float64)
}

type IRankEntrySingle interface {
	IRankEntry
	SetTargetRank(targetRank int)
	GetTargetRank() int
}

type IRankEntryRange interface {
	IRankEntry
	SetTargetRange(targetRange *util.HiLoInt)
	GetTargetRange() *util.HiLoInt
}

type IRankEntryFlatSingle interface {
	IRankEntryFlat
	IRankEntrySingle
}

type IRankEntryFlatRange interface {
	IRankEntryFlat
	IRankEntryRange
}

type IRankEntryExtendedSingle interface {
	IRankEntry
	GetTargetRankBySim(simType stats.SimType) int
	SetTargetRankBySim(simType stats.SimType, targetRank int)
}

type IRankEntryExtendedRange interface {
	IRankEntry
	GetTargetRangeBySim(simType stats.SimType) *util.HiLoInt
	SetTargetRangeBySim(simType stats.SimType, targetRank *util.HiLoInt)
}

type RankStatWeightsCommon struct {
	Data       *WeightInput
	SimScore   float64
	TargetRank int
}

func (rc *RankStatWeightsCommon) GetSimData() *stats.SimData {
	return &rc.Data.SimResult
}
func (rc *RankStatWeightsCommon) GetSimScore() float64 {
	return rc.SimScore
}
func (rc *RankStatWeightsCommon) IncrementSimScore(add float64) {
	rc.SimScore += add
}
func (rc *RankStatWeightsCommon) SetTargetRank(targetRank int) {
	rc.TargetRank = targetRank
}
func (rc *RankStatWeightsCommon) GetTargetRank() int {
	return rc.TargetRank
}
