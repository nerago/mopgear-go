package simrank

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/weightfind/weight_types"
)

func RankSimsForRankingSeparated[T weight_types.IRankEntryExtendedSingle](requiredSims []stats.SimType, inputData []T) {
	for _, simType := range requiredSims {
		sortGenericBasic(simType, inputData)
		arrayRankToSetSimSpecificSimRank(simType, inputData)
	}
}

func RankSimsForRangedRankSeparated[T weight_types.IRankEntryExtendedSingle](requiredSims []stats.SimType, inputData []T) {
	for _, simType := range requiredSims {
		sortGenericWithDeviation(simType, inputData)
		//arrayRankToSetSimSpecificSimRank(simType, inputData)
		panic("TODO")
	}
}

func arrayRankToSetSimSpecificSimRank[T weight_types.IRankEntryExtendedSingle](simType stats.SimType, inputData []T) {
	for rank := range inputData {
		inputData[rank].SetTargetRankBySim(simType, rank)
	}
}
