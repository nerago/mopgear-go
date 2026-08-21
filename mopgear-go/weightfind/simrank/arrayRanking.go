package simrank

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func arrayRankToSetSimRankFlat[T weight_types.IRankEntryFlatSingle](inputData []T) {
	for rank := range inputData {
		inputData[rank].SetSimRank(rank)
	}
}

func arrayRankToSetSimSpecificSimRank[T weight_types.IRankEntryExtendedSingle](simType stats.SimType, inputData []T) {
	for rank := range inputData {
		inputData[rank].SetSimRankByType(simType, rank)
	}
}

func arrayRankToSetSimRankRange[T weight_types.IRankEntryFlatRange](data []T) {
	data[0].SetSimRankRange(&util_collection.HiLoInt{Lo: 0, Hi: 0})
	for rank := 1; rank < len(data); rank++ {
		if util.FloatsApproxEquals(data[rank].GetSimScore(), data[rank-1].GetSimScore()) {
			prevRange := data[rank-1].GetSimRankRange()
			data[rank].SetSimRankRange(prevRange)
			prevRange.Hi = rank
		} else {
			data[rank].SetSimRankRange(&util_collection.HiLoInt{Lo: rank, Hi: rank})
		}
	}
}

func arrayRankToSetRangeStatisticalComplicated[T weight_types.IRankEntryExtendedRangeInt](simType stats.SimType, data []T) {
	for runStart := 0; runStart < len(data); {
		countChainedRun, countEqualFirstRun := arrayFindChainFromFunc(data, runStart, simType)
		runStart = arrayApplyChainFunc(data, runStart, countChainedRun, countEqualFirstRun, simType)
	}
}

// make ranked entries for later, calculating the sim rank as we go
// this uses the input array to make a new one, generic form would be a mapping operation
func accuracyPrepareCalcHiLo(inputData []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyInfoPrepared {
	prepare := make([]*weight_types.AccuracyInfoPrepared, len(inputData))

	prepare[0] = &weight_types.AccuracyInfoPrepared{
		SimRankRange: &util_collection.HiLoInt{Lo: 0, Hi: 0},
		Stats:        inputData[0].DataStat,
	}
	for i := 1; i < len(inputData); i++ {
		if util.FloatsApproxEquals(inputData[i].SimScore, inputData[i-1].SimScore) {
			prevRange := prepare[i-1].SimRankRange
			prevRange.Hi = i
			prepare[i] = &weight_types.AccuracyInfoPrepared{
				SimRankRange: prevRange,
				Stats:        inputData[i].DataStat,
			}
		} else {
			newRange := &util_collection.HiLoInt{Lo: i, Hi: i}
			prepare[i] = &weight_types.AccuracyInfoPrepared{
				SimRankRange: newRange,
				Stats:        inputData[i].DataStat,
			}
		}
	}

	return prepare
}

func accuracyPrepareCalcHiLoComplicated(inputData []*weight_types.AccuracyInfoPrePrepareExtended) []*weight_types.AccuracyInfoPrepared {
	return accuracyPrepareGeneral(inputData, func(a, b *weight_types.AccuracyInfoPrePrepareExtended) bool {
		diff := complexSummaryDiff(a, b)
		return util.FloatEqualsZero(diff)
	})
	
	//prepare := make([]*weight_types.AccuracyInfoPrepared, len(inputData))
	//
	//prepare[0] = &weight_types.AccuracyInfoPrepared{
	//	SimRankRange: &util_collection.HiLoInt{Lo: 0, Hi: 0},
	//	Stats:        inputData[0].DataStat,
	//}
	//for i := 1; i < len(inputData); i++ {
	//	diff := complexSummaryDiff(inputData[i], inputData[i-1])
	//	if util.FloatEqualsZero(diff) {
	//		prevRange := prepare[i-1].SimRankRange
	//		prevRange.Hi = i
	//		prepare[i] = &weight_types.AccuracyInfoPrepared{
	//			SimRankRange: prevRange,
	//			Stats:        inputData[i].DataStat,
	//		}
	//	} else {
	//		newRange := &util_collection.HiLoInt{Lo: i, Hi: i}
	//		prepare[i] = &weight_types.AccuracyInfoPrepared{
	//			SimRankRange: newRange,
	//			Stats:        inputData[i].DataStat,
	//		}
	//	}
	//}
	//
	//return prepare
}

func accuracyPrepareGeneral[A interface{ GetStatData() *stats.StatBlock }](inputData []A, equalEntries func(A, A) bool) []*weight_types.AccuracyInfoPrepared {
	prepare := make([]*weight_types.AccuracyInfoPrepared, len(inputData))

	prepare[0] = &weight_types.AccuracyInfoPrepared{
		SimRankRange: &util_collection.HiLoInt{Lo: 0, Hi: 0},
		Stats:        inputData[0].GetStatData(),
	}
	for i := 1; i < len(inputData); i++ {
		if equalEntries(inputData[i], inputData[i-1]) {
			prevRange := prepare[i-1].SimRankRange
			prevRange.Hi = i
			prepare[i] = &weight_types.AccuracyInfoPrepared{
				SimRankRange: prevRange,
				Stats:        inputData[i].GetStatData(),
			}
		} else {
			newRange := &util_collection.HiLoInt{Lo: i, Hi: i}
			prepare[i] = &weight_types.AccuracyInfoPrepared{
				SimRankRange: newRange,
				Stats:        inputData[i].GetStatData(),
			}
		}
	}

	return prepare
}
