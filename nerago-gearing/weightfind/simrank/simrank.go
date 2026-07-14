package simrank

import (
	"cmp"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
)

func RankSimsRegularForAccuracyRanged(requiredSims []stats.SimType, data []*weight_types.AccuracyInfoSimStatRanged, simRatios stats.SimData) {
	// score each sim
	for _, simType := range requiredSims {
		slices.SortFunc(data, simSortRangedCompares[simType])
		ratio := simRatios.Get(simType)
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

func RankSimsStatisticalForAccuracyRanged(requiredSims []stats.SimType, data []*weight_types.AccuracyInfoSimStatRanged, simRatios stats.SimData) {
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
		ratio := simRatios.Get(simType)
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

func RankSimsStatisticalForAccuracyPrepare(simRatios stats.SimData, data []*weight_types.AccuracyInfoPrePrepare, requiredSims []stats.SimType) {
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
		ratio := simRatios.Get(simType)
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

func RankSimsBasicForAccuracyPrepare(simRatios stats.SimData, data []*weight_types.AccuracyInfoPrePrepare, requiredSims []stats.SimType) {
	for _, simType := range requiredSims {
		ratio := simRatios.Get(simType)
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

func RankingWeight4RankSims(runData []weight_types.RankEntry4, requiredSims []stats.SimType, targetRatios stats.SimData) { // reset values
	for i := range runData {
		runData[i].SimScore = 0
		runData[i].TargetRank = 0
	}

	// score each sim
	for _, simType := range requiredSims {
		for entry, simDetailRankHiLo := range util.CalculateRankingRanges(simType.IsHighGood(), runData, func(x *weight_types.RankEntry4) float64 { return x.Data.SimResult.Get(simType) }) {
			entry.SimScore += float64(simDetailRankHiLo.Mid()) * targetRatios.Get(simType)
		}
	}

	// rank combined sims
	for entry, simRankHiLo := range util.CalculateRankingRanges(true, runData, func(x *weight_types.RankEntry4) float64 { return x.SimScore }) {
		entry.TargetRank = simRankHiLo.Lo
	}
}

func RankWeights5RankSims(runData []weight_types.RankEntry5, requiredSims []stats.SimType, targetRatios stats.SimData, printer *util.PrintRecorder) []weight_types.RankEntry5 {
	// score each sim
	for _, simType := range requiredSims {
		for entry, simDetailRankHiLo := range util.CalculateRankingRanges(simType.IsHighGood(), runData, func(x *weight_types.RankEntry5) float64 { return x.Data.SimResult.Get(simType) }) {
			entry.SimScore += float64(simDetailRankHiLo.Mid()) * targetRatios.Get(simType)
		}
	}

	runData = util.RemoveDuplicatesFuncNotify(runData,
		func(a, b *weight_types.RankEntry5) bool { return a.SimScore == b.SimScore },
		func(entry *weight_types.RankEntry5) { printer.Println("removing duplicate score") },
	)

	slices.SortFunc(runData, func(a, b weight_types.RankEntry5) int { return cmp.Compare(a.SimScore, b.SimScore) })

	return runData
}
