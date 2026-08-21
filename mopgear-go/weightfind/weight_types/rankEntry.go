package weight_types

import (
	"iter"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util/util_collection"
)

type IRankEntry interface {
	GetSimData() *stats.SimData
}

type IRankEntryFlatRead interface {
	IRankEntry
	GetSimScore() float64
}

type IRankEntryFlat interface {
	IRankEntryFlatRead
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

type IRankEntryExtendedRangeInt interface {
	IRankEntry
	SetSimRankRangeByType(simType stats.SimType, lo int, hi int)
}

type IRankEntryExtendedRangeFloat interface {
	IRankEntry
	GetSimRankRangeFloatByType(simType stats.SimType) util_collection.HiLoFloat
	SeqSimRankRangeFloatByType() iter.Seq2[stats.SimType, util_collection.HiLoFloat]
	SetSimRankRangeFloatByType(simType stats.SimType, lo float64, hi float64)
}

type IRankEntryExtendedRangeAndSummary interface {
	IRankEntryRange
	IRankEntryExtendedRangeInt
	IRankEntryExtendedRangeFloat
}

type IRankStatFlatRange interface {
	GetStatScore() float64
	SetStatRankRange(targetRange *util_collection.HiLoInt)
	GetStatRankRange() *util_collection.HiLoInt
}

type IRankDoubleFlatRange interface {
	//IRankEntryFlatRead
	IRankStatFlatRange
	GetSimRankRange() *util_collection.HiLoInt
}

type RankStatWeightsCommon struct {
	Data       *WeightInput
	SimScore   float64
	TargetRank int
}

var _ IRankEntryFlatSingle = &RankStatWeightsCommon{}

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
