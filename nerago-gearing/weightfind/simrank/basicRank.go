package simrank

import (
	"cmp"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
	"slices"
)

func simScoringBasic[T weight_types.IRankEntryFlat](simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) {
	resetSimScores(inputData)
	for _, simType := range simList {
		sortGenericBasic(simType, inputData)
		arrayRankToIncrementSimScore(simType, priority, inputData)
	}
}

func simScoringStatistical[T weight_types.IRankEntryFlat](simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) {
	resetSimScores(inputData)
	for _, simType := range simList {
		sortGenericWithDeviation(simType, inputData)
		arrayRankToIncrementSimScore(simType, priority, inputData)
	}
}

func simScoringBasicFastForAccuracy(simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []*weight_types.AccuracyInfo) {
	for _, simType := range simList {
		sortAccuracyFast(inputData, simType)
		arrayRankToIncrementSimScore(simType, priority, inputData)
	}
}

func simScoringBasicFastForAccuracyPrepare(simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []*weight_types.AccuracyInfoPrePrepare) {
	for _, simType := range simList {
		sortAccuracyPrepareFast(inputData, simType)
		arrayRankToIncrementSimScore(simType, priority, inputData)
	}
}

func simScoringStatisticalForAccuracy(simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []*weight_types.AccuracyInfo) {
	for _, simType := range simList {
		sortAccuracyWithDeviation(simType, inputData)
		arrayRankToIncrementSimScore(simType, priority, inputData)
	}
}

func simScoringStatisticalForAccuracyPrepare(simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []*weight_types.AccuracyInfoPrePrepare) {
	for _, simType := range simList {
		sortAccuracyPrepareWithDeviation(simType, inputData)
		arrayRankToIncrementSimScore(simType, priority, inputData)
	}
}

func resetSimScores[T weight_types.IRankEntryFlat](inputData []T) {
	for _, entry := range inputData {
		entry.ResetSimScore()
	}
}

func arrayRankToIncrementSimScore[T weight_types.IRankEntryFlat](simType stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) {
	ratio := priority.GetOrPanic(simType)
	for rank := range inputData {
		increment := float64(rank) * ratio
		inputData[rank].IncrementSimScore(increment)
	}
}

func sortSimScores[T weight_types.IRankEntryFlat](data []T) {
	// rank combined sims
	slices.SortFunc(data, func(a, b T) int {
		return cmp.Compare(a.GetSimScore(), b.GetSimScore())
	})
}

func rankOrderBasic[T weight_types.IRankEntryFlatSingle](inputData []T) {
	sortSimScores(inputData)
	arrayRankToSetSimBasicSimRank(inputData)
}

func RankingWeightsPrepareBasicRankings[T weight_types.IRankEntryFlatSingle](simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) {
	simScoringBasic(simList, priority, inputData)
	rankOrderBasic(inputData)
}

// currently just in Rank5
func RankingWeightsPrepareBasicRankingsRemoveDuplicates[T weight_types.IRankEntryFlatSingle](simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) []T {
	simScoringBasic(simList, priority, inputData)
	inputData = util_collection.RemoveDuplicatesFunc_NewIfChanged(inputData, func(a, b *T) bool { return (*a).GetSimScore() == (*b).GetSimScore() })
	rankOrderBasic(inputData)
	return inputData
}

func simScoringMidRange[T weight_types.IRankEntryFlatSingle](simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) {
	// score each sim
	resetSimScores(inputData)
	for _, simType := range simList {
		for entry, simDetailRankHiLo := range calculateRankingRanges(simType.IsHighGood(), inputData, func(x T) float64 { return x.GetSimData().Get(simType) }) {
			increment := float64(simDetailRankHiLo.Mid()) * priority.GetOrPanic(simType)
			entry.IncrementSimScore(increment)
		}
	}
}

// this is weird, used in RankingStatWeights4 only
func RankingWeightsPrepareUsingMidRange[T weight_types.IRankEntryFlatSingle](simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) { // reset values
	simScoringMidRange(simList, priority, inputData)

	// rank combined sims
	for entry, simRankHiLo := range calculateRankingRanges(true, inputData, func(x T) float64 { return x.GetSimScore() }) {
		// not sure if this intentional or a normal rank would be fine too
		entry.SetSimRank(simRankHiLo.Lo)
	}

	slices.SortFunc(inputData, func(a, b T) int { return cmp.Compare(a.GetSimRank(), b.GetSimRank()) })
}

// currently just in Rank5
func RankingWeightsPrepareUsingMidRangeRemoveDuplicates[T weight_types.IRankEntryFlatSingle](simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) []T {
	simScoringMidRange(simList, priority, inputData)
	inputData = util_collection.RemoveDuplicatesFunc_NewIfChanged(inputData, func(a, b *T) bool { return (*a).GetSimScore() == (*b).GetSimScore() })
	rankOrderBasic(inputData)
	return inputData
}
