package simrank

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func RankSimsStatisticalFlatSingle[T weight_types.IRankEntryFlatSingle](requiredSims []stats.SimType, data []T, simRatios *weight_types.SimPriorityBasic) {
	simScoringStatistical(requiredSims, simRatios, data)
	sortSimScores(data)
	arrayRankToSetSimRankFlat(data)
}

func RankSimsBasicForRanged[T weight_types.IRankEntryFlatRange](requiredSims []stats.SimType, data []T, simRatios *weight_types.SimPriorityBasic) {
	simScoringBasic(requiredSims, simRatios, data)
	sortSimScores(data)
	arrayRankToSetSimRankRange(data)
}

func RankSimsStatisticalForRanged[T weight_types.IRankEntryFlatRange](requiredSims []stats.SimType, data []T, simRatios *weight_types.SimPriorityBasic) {
	simScoringStatistical(requiredSims, simRatios, data)
	sortSimScores(data)
	arrayRankToSetSimRankRange(data)
}

func RankSimsStatisticalForExtendedRanged[T weight_types.IRankEntryExtendedRangeAndSummary](requiredSims []stats.SimType, data []T, simRatios *weight_types.SimPriorityBasic) {
	for _, simType := range requiredSims {
		sortGenericWithDeviation(simType, data)
		arrayRankToSetRangeStatisticalComplicated(simType, data)
	}
	complexSimRankRangesToSummaryRange(data, simRatios)
}

// just used on accuracyPrepare init
func AccuracyPrepareRankSimsStatisticalExtended(simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyInfoPrepared {
	simScoringStatistical(simList, priority, inputData)
	sortSimScores(inputData)
	return accuracyPrepareCalcHiLo(inputData)
}

// just used on accuracyPrepare init
func AccuracyPrepareRankSimsStatistical(simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyInfoPrepared {
	simScoringStatistical(simList, priority, inputData)
	sortSimScores(inputData)
	return accuracyPrepareCalcHiLo(inputData)
}

// just used on accuracyPrepare init
func AccuracyPrepareRankSimsBasic(simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyInfoPrepared {
	simScoringBasic(simList, priority, inputData)
	sortSimScores(inputData)
	return accuracyPrepareCalcHiLo(inputData)
}

func RankSimsForRankingSeparated[T weight_types.IRankEntryExtendedSingle](requiredSims []stats.SimType, inputData []T) {
	for _, simType := range requiredSims {
		sortGenericBasic(simType, inputData)
		arrayRankToSetSimSpecificSimRank(simType, inputData)
	}
}

func RankSimsForRangedRankSeparated[T weight_types.IRankEntryExtendedRangeInt](requiredSims []stats.SimType, inputData []T) {
	for _, simType := range requiredSims {
		sortGenericWithDeviation(simType, inputData)
		arrayRankToSetRangeStatisticalComplicated(simType, inputData)
	}
}

func ResetSimScores[T weight_types.IRankEntryFlat](inputData []T) {
	for _, entry := range inputData {
		entry.ResetSimScore()
	}
}
