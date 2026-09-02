package simrank

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

//func RankSimsStatisticalFlatSingle[T weight_types.IRankEntryFlatSingle](simList []stats.SimType, data []T, priority *weight_types.SimPriorityBasic) {
//	simScoringStatistical(simList, priority, data)
//	sortSimScores(data)
//	arrayRankToSetSimRankFlat(data)
//
//	// TODO this method could be considered obsolete?
//}

func RankSimsStatisticalFlatSingleExtended_Experimental[T weight_types.IRankEntryFlatSingle](simList []stats.SimType, data []T, priority *weight_types.SimPriorityBasic) {
	adaptedData := util_collection.MapSliceAsNew_NoPointer(data, func(x T) *rankEntryFlatSingle_IRankEntryExtendedRangeAndSummary[T] {
		return &rankEntryFlatSingle_IRankEntryExtendedRangeAndSummary[T]{inner: x}
	})
	simScoringStatisticalComplicated(simList, adaptedData)
	multiplyFloatRangesByRatio(adaptedData, priority)
	sortSimRankComplex(adaptedData)
	arrayRankToSetSimRankRangeComplexCompare(adaptedData)
}

func RankSimsBasicForRanged[T weight_types.IRankEntryFlatRange](simList []stats.SimType, data []T, priority *weight_types.SimPriorityBasic) {
	simScoringBasic(simList, priority, data)
	sortSimScores(data)
	arrayRankToSetSimRankRange(data)
}

func RankSimsStatisticalForExtendedRanged[T weight_types.IRankEntryExtendedRangeAndSummary](simList []stats.SimType, data []T, priority *weight_types.SimPriorityBasic) {
	simScoringStatisticalComplicated(simList, data)
	multiplyFloatRangesByRatio(data, priority)
	sortSimRankComplex(data)
	arrayRankToSetSimRankRangeComplexCompare(data)

	// something like this for everything?
}

// just used on accuracyPrepare init
func AccuracyPrepareRankSimsStatisticalExtended(simList []stats.SimType, priority *weight_types.SimPriorityBasic, data []*weight_types.AccuracyInfoPrePrepareExtended) []*weight_types.AccuracyInfoPrepared {
	simScoringStatisticalComplicated(simList, data)
	multiplyFloatRangesByRatio(data, priority)
	sortSimRankComplex(data)
	return accuracyPrepareCalcRangeComplicated(data)
}

// just used on accuracyPrepare init
//func AccuracyPrepareRankSimsStatistical(simList []stats.SimType, priority *weight_types.SimPriorityBasic, data []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyInfoPrepared {
//	simScoringStatistical(simList, priority, data)
//	sortSimScores(data)
//	return accuracyPrepareCalcRangeRegular(data)
//}

// just used on accuracyPrepare init
func AccuracyPrepareRankSimsBasic(simList []stats.SimType, priority *weight_types.SimPriorityBasic, data []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyInfoPrepared {
	simScoringBasic(simList, priority, data)
	sortSimScores(data)
	return accuracyPrepareCalcRangeRegular(data)
}

func RankSimsForRankingSeparated[T weight_types.IRankEntryExtendedSingle](simList []stats.SimType, data []T) {
	simScoringSeparated(simList, data)
}

func RankSimsForRangedRankSeparated[T weight_types.IRankEntryExtendedRangeInt](simList []stats.SimType, data []T) {
	simScoringStatisticalComplicated(simList, data)
}

func ResetSimScores[T weight_types.IRankEntryFlat](data []T) {
	for _, entry := range data {
		entry.ResetSimScore()
	}
}
