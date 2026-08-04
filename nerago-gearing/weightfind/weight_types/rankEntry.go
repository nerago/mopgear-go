package weight_types

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_collection"
)

type IRankEntry interface {
	GetSimData() *stats.SimData
}

type IRankEntryFlat interface {
	IRankEntry
	GetSimScore() float64
	ResetSimScore()
	IncrementSimScore(add float64)
}

type IRankEntrySingle interface {
	IRankEntry
	SetSimRank(targetRank int)
	GetSimRank() int
}

type IRankEntryRange interface {
	IRankEntry
	SetSimRankRange(targetRange *util_collection.HiLoInt)
	GetSimRankRange() *util_collection.HiLoInt
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
	GetSimRankByType(simType stats.SimType) int
	SetSimRankByType(simType stats.SimType, targetRank int)
}

type IRankEntryExtendedRange interface {
	IRankEntry
	GetSimRankRangeByType(simType stats.SimType) *util_collection.HiLoInt
	SetSimRankRangeByType(simType stats.SimType, targetRank *util_collection.HiLoInt)
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
func (rc *RankStatWeightsCommon) ResetSimScore() {
	rc.SimScore = 0
}
func (rc *RankStatWeightsCommon) SetSimRank(targetRank int) {
	rc.TargetRank = targetRank
}
func (rc *RankStatWeightsCommon) GetSimRank() int {
	return rc.TargetRank
}
