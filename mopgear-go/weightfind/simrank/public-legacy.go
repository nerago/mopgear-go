package simrank

import (
	"cmp"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func RankingWeightsPrepareBasicRankings[T weight_types.IRankEntryFlatSingle](simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) {
	ResetSimScores(inputData)
	simScoringBasic(simList, priority, inputData)
	rankOrderBasic(inputData)
}

// currently just in Rank5
func RankingWeightsPrepareBasicRankingsRemoveDuplicates[T weight_types.IRankEntryFlatSingle](simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) []T {
	ResetSimScores(inputData)
	simScoringBasic(simList, priority, inputData)
	inputData = util_collection.RemoveDuplicatesFunc_NewIfChanged(inputData, func(a, b *T) bool { return (*a).GetSimScore() == (*b).GetSimScore() })
	rankOrderBasic(inputData)
	return inputData
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
