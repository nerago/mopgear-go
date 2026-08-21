package simrank

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func simScoringBasic[T weight_types.IRankEntryFlat](simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) {
	for _, simType := range simList {
		sortGenericBasic(simType, inputData)
		arrayRankToIncrementSimScore(simType, priority, inputData)
	}
}

func simScoringStatistical[T weight_types.IRankEntryFlat](simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) {
	for _, simType := range simList {
		sortGenericWithDeviation(simType, inputData)
		arrayRankToIncrementSimScore(simType, priority, inputData)
	}
}

func arrayRankToIncrementSimScore[T weight_types.IRankEntryFlat](simType stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) {
	ratio := priority.GetOrPanic(simType)
	for rank := range inputData {
		fmt.Printf("%s rank %d %f %f\n", simType.Name(), rank, inputData[rank].GetSimData().Get(simType), ratio)
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
	arrayRankToSetSimRankFlat(inputData)
}

func simScoringMidRange[T weight_types.IRankEntryFlatSingle](simList []stats.SimType, priority *weight_types.SimPriorityBasic, inputData []T) {
	// score each sim
	ResetSimScores(inputData)
	for _, simType := range simList {
		for entry, simDetailRankHiLo := range calculateRankingRanges(simType.IsHighGood(), inputData, func(x T) float64 { return x.GetSimData().Get(simType) }) {
			increment := float64(simDetailRankHiLo.Mid()) * priority.GetOrPanic(simType)
			entry.IncrementSimScore(increment)
		}
	}
}
