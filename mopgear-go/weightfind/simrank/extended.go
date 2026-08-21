package simrank

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func RankSimsForRankingSeparated[T weight_types.IRankEntryExtendedSingle](requiredSims []stats.SimType, inputData []T) {
	for _, simType := range requiredSims {
		sortGenericBasic(simType, inputData)
		arrayRankToSetSimSpecificSimRank(simType, inputData)
	}
}

func RankSimsForRangedRankSeparated[T weight_types.IRankEntryExtendedRange](requiredSims []stats.SimType, inputData []T) {
	for _, simType := range requiredSims {
		sortGenericWithDeviation(simType, inputData)
		arrayRankToSetRangeStatisticalComplicated(simType, inputData)
	}
}
