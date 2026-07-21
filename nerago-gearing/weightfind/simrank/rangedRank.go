package simrank

import (
	"cmp"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
)

func RankSimsRegularForAccuracyRanged(requiredSims []stats.SimType, data []*weight_types.AccuracyInfoSimStatRanged, simRatios *weight_types.SimPriorityBasic) {
	simScoringBasicFastForAccuracy(requiredSims, simRatios, data)
	sortSimScores(data)
	arrayRankToSetSimRank(data)
}

func RankSimsStatisticalForAccuracyRanged(requiredSims []stats.SimType, data []*weight_types.AccuracyInfoSimStatRanged, simRatios *weight_types.SimPriorityBasic) {
	simScoringStatisticalForAccuracy(requiredSims, simRatios, data)
	sortSimScores(data)
	arrayRankToSetSimRank(data)
}

func sortSimScores(data []*weight_types.AccuracyInfoSimStatRanged) {
	// rank combined sims
	slices.SortFunc(data, func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
		return cmp.Compare(a.SimScore, b.SimScore)
	})
}

func arrayRankToSetSimRank(data []*weight_types.AccuracyInfoSimStatRanged) {
	data[0].SimRankRange = &util.HiLoInt{Lo: 0, Hi: 0}
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].SimScore, data[i-1].SimScore) {
			prevRange := data[i-1].SimRankRange
			data[i].SimRankRange = prevRange
			prevRange.Hi = i
		} else {
			data[i].SimRankRange = &util.HiLoInt{Lo: i, Hi: i}
		}
	}
}

// make ranked entries for later, calculating the sim rank as we go
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
