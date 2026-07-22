package simrank

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_types"
)

func RankSimsRegularForAccuracyRanged(requiredSims []stats.SimType, data []*weight_types.AccuracyInfoSimStatRanged, simRatios *weight_types.SimPriorityBasic) {
	simScoringBasicFastForAccuracy(requiredSims, simRatios, data)
	sortSimScores(data)
	arrayRankToSetSimRankRange(data)
}

func RankSimsStatisticalForAccuracyRanged(requiredSims []stats.SimType, data []*weight_types.AccuracyInfoSimStatRanged, simRatios *weight_types.SimPriorityBasic) {
	simScoringStatisticalForAccuracy(requiredSims, simRatios, data)
	sortSimScores(data)
	arrayRankToSetSimRankRange(data)
}

// TODO new algo that doesn't overextend range, especially for stat based

func arrayRankToSetSimRankRange[T weight_types.IRankEntryFlatRange](data []T) {
	data[0].SetSimRankRange(&util.HiLoInt{Lo: 0, Hi: 0})
	for rank := 1; rank < len(data); rank++ {
		if util.FloatsApproxEquals(data[rank].GetSimScore(), data[rank-1].GetSimScore()) {
			prevRange := data[rank-1].GetSimRankRange()
			data[rank].SetSimRankRange(prevRange)
			prevRange.Hi = rank
		} else {
			data[rank].SetSimRankRange(&util.HiLoInt{Lo: rank, Hi: rank})
		}
	}
}

// make ranked entries for later, calculating the sim rank as we go
// this uses the input array to make a new one, generic form would be a mapping operation
func AccuracyPrepareCalcHiLo(inputData []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyPreparedEntry {
	prepare := make([]*weight_types.AccuracyPreparedEntry, len(inputData))

	prepare[0] = &weight_types.AccuracyPreparedEntry{
		SimRankRange: &util.HiLoInt{Lo: 0, Hi: 0},
		Stats:        inputData[0].DataStat,
	}
	for i := 1; i < len(inputData); i++ {
		if util.FloatsApproxEquals(inputData[i].SimScore, inputData[i-1].SimScore) {
			prevRange := prepare[i-1].SimRankRange
			prevRange.Hi = i
			prepare[i] = &weight_types.AccuracyPreparedEntry{SimRankRange: prevRange, Stats: inputData[i].DataStat}
		} else {
			newRange := &util.HiLoInt{Lo: i, Hi: i}
			prepare[i] = &weight_types.AccuracyPreparedEntry{SimRankRange: newRange, Stats: inputData[i].DataStat}
		}
	}

	return prepare
}
