package simrank

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func RankSimsStatisticalFlatSingle[T weight_types.IRankEntryFlatSingle](simList []stats.SimType, data []T, priority *weight_types.SimPriorityBasic) {
	simScoringStatistical(simList, priority, data)
	sortSimScores(data)
	arrayRankToSetSimRankFlat(data)
}

func RankSimsBasicForRanged[T weight_types.IRankEntryFlatRange](simList []stats.SimType, data []T, priority *weight_types.SimPriorityBasic) {
	simScoringBasic(simList, priority, data)
	sortSimScores(data)
	arrayRankToSetSimRankRange(data)
}

func RankSimsStatisticalForRanged[T weight_types.IRankEntryFlatRange](simList []stats.SimType, data []T, priority *weight_types.SimPriorityBasic) {
	simScoringStatistical(simList, priority, data)
	sortSimScores(data)
	arrayRankToSetSimRankRange(data)
}

func RankSimsStatisticalForExtendedRanged[T weight_types.IRankEntryExtendedRangeAndSummary](simList []stats.SimType, data []T, priority *weight_types.SimPriorityBasic) {
	simScoringStatisticalComplicated(simList, data)
	multiplyFloatRangesByRatio(data, priority)
	sortSimRankComplex(data)
	arrayRankToSetSimRankRangeComplexCompare(data)
}

// just used on accuracyPrepare init
func AccuracyPrepareRankSimsStatisticalExtended(simList []stats.SimType, priority *weight_types.SimPriorityBasic, data []*weight_types.AccuracyInfoPrePrepareExtended) []*weight_types.AccuracyInfoPrepared {
	simScoringStatisticalComplicated(simList, data)
	multiplyFloatRangesByRatio(data, priority)
	sortSimScores(data)
	return accuracyPrepareCalcHiLo(data)
}

// just used on accuracyPrepare init
func AccuracyPrepareRankSimsStatistical(simList []stats.SimType, priority *weight_types.SimPriorityBasic, data []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyInfoPrepared {
	simScoringStatistical(simList, priority, data)
	sortSimScores(data)
	return accuracyPrepareCalcHiLo(data)
}

// just used on accuracyPrepare init
func AccuracyPrepareRankSimsBasic(simList []stats.SimType, priority *weight_types.SimPriorityBasic, data []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyInfoPrepared {
	simScoringBasic(simList, priority, data)
	sortSimScores(data)
	return accuracyPrepareCalcHiLo(data)
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
