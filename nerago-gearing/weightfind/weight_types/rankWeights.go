package weight_types

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_highs"
)

type IRankEntryMinimum interface {
	GetSimData() *stats.SimData
	GetSimScore() float64
	IncrementSimScore(add float64)
}

type IRankEntryCommon interface {
	IRankEntryMinimum
	SetTargetRank(targetRank int)
	GetTargetRank() int
}

type RankEntryCommon struct {
	Data       *WeightInput
	SimScore   float64
	TargetRank int
}

func (rc *RankEntryCommon) GetSimData() *stats.SimData {
	return &rc.Data.SimResult
}
func (rc *RankEntryCommon) GetSimScore() float64 {
	return rc.SimScore
}
func (rc *RankEntryCommon) IncrementSimScore(add float64) {
	rc.SimScore += add
}
func (rc *RankEntryCommon) SetTargetRank(targetRank int) {
	rc.TargetRank = targetRank
}
func (rc *RankEntryCommon) GetTargetRank() int {
	return rc.TargetRank
}

type RankEntry struct {
	RankEntryCommon

	ScoreColumn util_highs.ColumnIndex
	RankColumn  util_highs.ColumnIndex
}
