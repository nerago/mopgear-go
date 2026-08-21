package simrank

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func RankSimsStatisticalForAccuracyPrepare(simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyInfoPrepared {
	simScoringStatisticalForAccuracy(simList, priority, inputData)

	// rank combined sims
	sortSimScores(inputData)

	return AccuracyPrepareCalcHiLo(inputData)
}

func AccuracyPrepareRankSimsBasic(simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyInfoPrepared {
	simScoringBasicFastForAccuracyPrepare(simList, priority, inputData)

	// rank combined sims
	sortSimScores(inputData)

	return AccuracyPrepareCalcHiLo(inputData)
}
