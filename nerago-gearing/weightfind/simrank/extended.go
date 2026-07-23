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
		arrayRankToSetRangeStatisticalComplicated(inputData,
			func(entry T) averageAndStdDev {
				return averageAndStdDev{
					average: entry.GetSimData().Get(simType),
					stdDev:  entry.GetSimData().GetStdDevOrZero(simType),
				}
			}, func(entry T, hiLo *util.HiLoInt) {
				entry.SetSimRankRangeByType(simType, hiLo)
			})
	}
}
