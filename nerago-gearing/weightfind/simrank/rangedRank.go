package simrank

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/weightfind/weight_types"
)

func RankSimsRegularForAccuracyRanged(requiredSims []stats.SimType, data []*weight_types.AccuracyInfo, simRatios *weight_types.SimPriorityBasic) {
	simScoringBasicFastForAccuracy(requiredSims, simRatios, data)
	sortSimScores(data)
	arrayRankToSetSimRankRange(data)
}

func RankSimsStatisticalForAccuracyRanged(requiredSims []stats.SimType, data []*weight_types.AccuracyInfo, simRatios *weight_types.SimPriorityBasic) {
	simScoringStatisticalForAccuracy(requiredSims, simRatios, data)
	sortSimScores(data)
	arrayRankToSetSimRankRange(data)
}

func RankSimsStatisticalFlat[T weight_types.IRankEntryFlatSingle](requiredSims []stats.SimType, data []T, simRatios *weight_types.SimPriorityBasic) {
	simScoringStatistical(requiredSims, simRatios, data)
	sortSimScores(data)
	arrayRankToSetSimBasicSimRank(data)
}

func RankSimsStatisticalForRanged[T weight_types.IRankEntryFlatRange](requiredSims []stats.SimType, data []T, simRatios *weight_types.SimPriorityBasic) {
	simScoringStatistical(requiredSims, simRatios, data)
	sortSimScores(data)
	arrayRankToSetSimRankRange(data)
}
