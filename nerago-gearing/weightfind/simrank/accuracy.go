package simrank

import (
	"cmp"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
)

func RankSimsStatisticalForAccuracyPrepare(simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyInfoPrepared {
	simScoringStatisticalForAccuracyPrepare(simList, priority, inputData)

	// rank combined sims
	slices.SortFunc(inputData, func(a, b *weight_types.AccuracyInfoPrePrepare) int {
		return cmp.Compare(a.SimScore, b.SimScore)
	})

	return AccuracyPrepareCalcHiLo(inputData)
}

func AccuracyPrepareRankSimsBasic(simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyInfoPrepared {
	simScoringBasicFastForAccuracyPrepare(simList, priority, inputData)

	// rank combined sims
	slices.SortFunc(inputData, func(a, b *weight_types.AccuracyInfoPrePrepare) int {
		return cmp.Compare(a.SimScore, b.SimScore)
	})

	return AccuracyPrepareCalcHiLo(inputData)
}
