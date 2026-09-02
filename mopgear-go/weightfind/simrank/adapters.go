package simrank

import (
	"iter"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type rankEntryFlatSingle_IRankEntryExtendedRangeAndSummary[T weight_types.IRankEntryFlatSingle] struct {
	inner         T
	simRankRange  *util_collection.HiLoInt
	simRankByType stats.SimTypeMap[util_collection.HiLoFloat]
}

func (ad *rankEntryFlatSingle_IRankEntryExtendedRangeAndSummary[T]) GetSimData() *stats.SimData {
	return ad.inner.GetSimData()
}

func (ad *rankEntryFlatSingle_IRankEntryExtendedRangeAndSummary[T]) SetSimRankRange(targetRange *util_collection.HiLoInt) {
	ad.simRankRange = targetRange
	ad.inner.SetSimRank(targetRange.Mid())
}

func (ad *rankEntryFlatSingle_IRankEntryExtendedRangeAndSummary[T]) GetSimRankRange() *util_collection.HiLoInt {
	return ad.simRankRange
}

func (ad *rankEntryFlatSingle_IRankEntryExtendedRangeAndSummary[T]) SetSimRankRangeByType(simType stats.SimType, lo int, hi int) {
	ad.simRankByType.Put(simType, util_collection.HiLoFloat{Lo: float64(lo), Hi: float64(hi)})
}

func (ad *rankEntryFlatSingle_IRankEntryExtendedRangeAndSummary[T]) GetSimRankRangeFloatByType(simType stats.SimType) util_collection.HiLoFloat {
	return ad.simRankByType.GetOrNilValue(simType)
}

func (ad *rankEntryFlatSingle_IRankEntryExtendedRangeAndSummary[T]) SeqSimRankRangeFloatByType() iter.Seq2[stats.SimType, util_collection.HiLoFloat] {
	return ad.simRankByType.SeqKeyValue()
}

func (ad *rankEntryFlatSingle_IRankEntryExtendedRangeAndSummary[T]) SetSimRankRangeFloatByType(simType stats.SimType, lo float64, hi float64) {
	ad.simRankByType.Put(simType, util_collection.HiLoFloat{Lo: lo, Hi: hi})
}
