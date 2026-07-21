package simrank

import (
	"cmp"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
)

func RankSimsRegularForAccuracyRanged(requiredSims []stats.SimType, data []*weight_types.AccuracyInfoSimStatRanged, simRatios weight_types.SimPriorityBasic) {
	// score each sim
	for _, simType := range requiredSims {
		slices.SortFunc(data, simSortRangedCompares[simType])
		ratio := simRatios.GetOrPanic(simType)
		for rank := range data {
			entry := data[rank]
			entry.SimScore += float64(rank) * ratio
		}
	}

	// rank combined sims
	slices.SortFunc(data, func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
		return cmp.Compare(a.SimScore, b.SimScore)
	})
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

func RankSimsStatisticalForAccuracyRanged(requiredSims []stats.SimType, data []*weight_types.AccuracyInfoSimStatRanged, simRatios weight_types.SimPriorityBasic) {
	// score each sim
	for _, simType := range requiredSims {
		if simType == stats.Sim_DEATH {
			// death data never has detail
			slices.SortFunc(data, simSortRangedCompares[stats.Sim_DEATH])
		} else if simType.IsHighGood() {
			slices.SortFunc(data, func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
				return deviationCompareSims(a.DataSim, b.DataSim, simType)
			})
		} else {
			slices.SortFunc(data, func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
				return deviationCompareSims(b.DataSim, a.DataSim, simType)
			})
		}
		ratio := simRatios.GetOrPanic(simType)
		for rank := range data {
			entry := data[rank]
			entry.SimScore += float64(rank) * ratio
		}
	}

	// rank combined sims
	slices.SortFunc(data, func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
		return cmp.Compare(a.SimScore, b.SimScore)
	})
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

func RankSimsStatisticalForAccuracyPrepare(simRatios weight_types.SimPriorityBasic, data []*weight_types.AccuracyInfoPrePrepare, requiredSims []stats.SimType) {
	for _, simType := range requiredSims {
		if simType == stats.Sim_DEATH {
			// death data never has detail
			slices.SortFunc(data, simSortSimSingledCompares[simType])
		} else if simType.IsHighGood() {
			slices.SortFunc(data, func(a, b *weight_types.AccuracyInfoPrePrepare) int {
				return deviationCompareSims(a.DataSim, b.DataSim, simType)
			})
		} else {
			slices.SortFunc(data, func(a, b *weight_types.AccuracyInfoPrePrepare) int {
				return deviationCompareSims(b.DataSim, a.DataSim, simType)
			})
		}
		ratio := simRatios.GetOrPanic(simType)
		for rank := range data {
			entry := data[rank]
			entry.SimScore += float64(rank) * ratio
		}
	}

	// rank combined sims
	slices.SortFunc(data, func(a, b *weight_types.AccuracyInfoPrePrepare) int {
		return cmp.Compare(a.SimScore, b.SimScore)
	})
}

func RankSimsBasicForAccuracyPrepare(simRatios weight_types.SimPriorityBasic, data []*weight_types.AccuracyInfoPrePrepare, requiredSims []stats.SimType) {
	for _, simType := range requiredSims {
		ratio := simRatios.GetOrPanic(simType)
		slices.SortFunc(data, simSortSimSingledCompares[simType])
		for rank := range data {
			data[rank].SimScore += float64(rank) * ratio
		}
	}

	// rank combined sims
	slices.SortFunc(data, func(a, b *weight_types.AccuracyInfoPrePrepare) int {
		return cmp.Compare(a.SimScore, b.SimScore)
	})
}

func CalcHiLoForAccuracyPrepare(data []*weight_types.AccuracyInfoPrePrepare, prepare []*weight_types.AccuracyPreparedEntry) {
	prepare[0] = &weight_types.AccuracyPreparedEntry{
		SimRankRange: &util.HiLoInt{Lo: 0, Hi: 0},
		Stats:        data[0].DataStat,
	}
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].SimScore, data[i-1].SimScore) {
			prevRange := prepare[i-1].SimRankRange
			prevRange.Hi = i
			prepare[i] = &weight_types.AccuracyPreparedEntry{SimRankRange: prevRange, Stats: data[i].DataStat}
		} else {
			newRange := &util.HiLoInt{Lo: i, Hi: i}
			prepare[i] = &weight_types.AccuracyPreparedEntry{SimRankRange: newRange, Stats: data[i].DataStat}
		}
	}
}

func simScoringBasic[T weight_types.IRankEntryHasCommon](simList []stats.SimType, inputData []T, priority weight_types.SimPriorityBasic) {
	// score each sim
	for _, simType := range simList {
		for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), inputData, func(x *T) float64 { return (*x).ToCommon().Data.SimResult.Get(simType) }) {
			(*entry).ToCommon().SimScore += float64(simDetailRank) * priority.GetOrPanic(simType)
		}
	}
}

func rankOrderBasic[T weight_types.IRankEntryHasCommon](inputData []T) {
	// rank combined sims
	for entry, simRank := range util.CalculateRanking(true, inputData, func(x *T) float64 { return (*x).ToCommon().SimScore }) {
		(*entry).ToCommon().TargetRank = simRank
	}

	slices.SortFunc(inputData, func(a, b T) int { return cmp.Compare(a.ToCommon().TargetRank, b.ToCommon().TargetRank) })
}

func RankingWeightsPrepareBasicRankings[T weight_types.IRankEntryHasCommon](simList []stats.SimType, inputData []T, priority weight_types.SimPriorityBasic) {
	simScoringBasic(simList, inputData, priority)
	rankOrderBasic(inputData)
}

// currently just in Rank5
func RankingWeightsPrepareBasicRankingsRemoveDuplicates[T weight_types.IRankEntryHasCommon](simList []stats.SimType, inputData []T, priority weight_types.SimPriorityBasic) []T {
	simScoringBasic(simList, inputData, priority)
	inputData = util.RemoveDuplicatesFunc_NewIfChanged(inputData, func(a, b *T) bool { return (*a).ToCommon().SimScore == (*b).ToCommon().SimScore })
	rankOrderBasic(inputData)
	return inputData
}

func simScoringMidRange[T weight_types.IRankEntryHasCommon](simList []stats.SimType, inputData []T, priority weight_types.SimPriorityBasic) {
	// score each sim
	for _, simType := range simList {
		for entry, simDetailRankHiLo := range util.CalculateRankingRanges(simType.IsHighGood(), inputData, func(x *T) float64 { return (*x).ToCommon().Data.SimResult.Get(simType) }) {
			(*entry).ToCommon().SimScore += float64(simDetailRankHiLo.Mid()) * priority.GetOrPanic(simType)
		}
	}
}

// this is weird, used in RankingStatWeights4 only
func RankingWeightsPrepareUsingMidRange[T weight_types.IRankEntryHasCommon](simList []stats.SimType, inputData []T, priority weight_types.SimPriorityBasic) { // reset values
	simScoringMidRange(simList, inputData, priority)

	// rank combined sims
	for entry, simRankHiLo := range util.CalculateRankingRanges(true, inputData, func(x *T) float64 { return (*x).ToCommon().SimScore }) {
		// not sure if this intentional or a normal rank would be fine too
		(*entry).ToCommon().TargetRank = simRankHiLo.Lo
	}

	slices.SortFunc(inputData, func(a, b T) int { return cmp.Compare(a.ToCommon().TargetRank, b.ToCommon().TargetRank) })
}

// currently just in Rank5
func RankingWeightsPrepareUsingMidRangeRemoveDuplicates[T weight_types.IRankEntryHasCommon](simList []stats.SimType, inputData []T, priority weight_types.SimPriorityBasic) []T {
	simScoringMidRange(simList, inputData, priority)
	inputData = util.RemoveDuplicatesFunc_NewIfChanged(inputData, func(a, b *T) bool { return (*a).ToCommon().SimScore == (*b).ToCommon().SimScore })
	rankOrderBasic(inputData)
	return inputData
}
