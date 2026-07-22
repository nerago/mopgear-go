package simrank

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_types"
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
		arrayRankToSetSimSpecificRankRange(simType, inputData)
	}
}

func arrayRankToSetSimSpecificSimRank[T weight_types.IRankEntryExtendedSingle](simType stats.SimType, inputData []T) {
	for rank := range inputData {
		inputData[rank].SetSimRankByType(simType, rank)
	}
}

func arrayRankToSetSimSpecificRankRange[T weight_types.IRankEntryExtendedRange](simType stats.SimType, data []T) {
	data[0].SetSimRankRangeByType(simType, &util.HiLoInt{Lo: 0, Hi: 0})
	for rank := 1; rank < len(data); rank++ {
		if util.FloatsApproxEquals(data[rank].GetSimData().Get(simType), data[rank-1].GetSimData().Get(simType)) {
			prevRange := data[rank-1].GetSimRankRangeByType(simType)
			data[rank].SetSimRankRangeByType(simType, prevRange)
			prevRange.Hi = rank
		} else {
			data[rank].SetSimRankRangeByType(simType, &util.HiLoInt{Lo: rank, Hi: rank})
		}
	}
}
