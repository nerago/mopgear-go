package simrank

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func arrayRankToSetSimBasicSimRank[T weight_types.IRankEntryFlatSingle](inputData []T) {
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

func arrayRankToSetSimSpecificRankRange[T weight_types.IRankEntryExtendedRange](simType stats.SimType, data []T) {
	data[0].SetSimRankRangeByType(simType, &util_collection.HiLoInt{Lo: 0, Hi: 0})
	for rank := 1; rank < len(data); rank++ {
		if util.FloatsApproxEquals(data[rank].GetSimData().Get(simType), data[rank-1].GetSimData().Get(simType)) {
			prevRange := data[rank-1].GetSimRankRangeByType(simType)
			data[rank].SetSimRankRangeByType(simType, prevRange)
			prevRange.Hi = rank
		} else {
			data[rank].SetSimRankRangeByType(simType, &util_collection.HiLoInt{Lo: rank, Hi: rank})
		}
	}
}

func arrayRankToSetRangeStatisticalComplicated[T weight_types.IRankEntry](data []T, getSimScore func(T) averageAndStdDev, setSimRankRange func(T, *util_collection.HiLoInt)) {
	for runStart := 0; runStart < len(data); {
		countChainedRun, countEqualFirstRun := arrayFindChainFromFunc(data, runStart, getSimScore)
		runStart = arrayApplyChainFunc(data, runStart, countChainedRun, countEqualFirstRun, setSimRankRange)
	}
}

type averageAndStdDev struct {
	average float64
	stdDev  float64
}

func arrayFindChainFromFunc[T weight_types.IRankEntry](data []T, runStart int, getSimScore func(T) averageAndStdDev) (int, int) {
	countChainedRun := 1
	countEqualFirstRun := 1

	firstScore := getSimScore(data[runStart])
	prevScore := firstScore

	for check := runStart + 1; check < len(data); check++ {
		currScore := getSimScore(data[check])
		if equalSimsDetailStatistical(currScore.average, currScore.stdDev, firstScore.average, firstScore.stdDev) {
			countEqualFirstRun++
			countChainedRun++
		} else if equalSimsDetailStatistical(currScore.average, currScore.stdDev, prevScore.average, prevScore.stdDev) {
			countChainedRun++
		} else {
			break
		}
		prevScore = currScore
	}

	return countChainedRun, countEqualFirstRun
}

func arrayApplyChainFunc[T weight_types.IRankEntry](data []T, runStart int, countChainedRun int, countEqualFirstRun int, setSimRankRange func(T, *util_collection.HiLoInt)) int {
	if countChainedRun == 1 {
		setSimRankRange(data[runStart], &util_collection.HiLoInt{Lo: runStart, Hi: runStart})
		return runStart + 1
	} else {
		var end int
		if countChainedRun == countEqualFirstRun || countChainedRun <= 4 || countEqualFirstRun >= countChainedRun*3/4 {
			end = runStart + countChainedRun - 1
		} else {
			end = runStart + countEqualFirstRun - 1
		}

		hilo := &util_collection.HiLoInt{Lo: runStart, Hi: end}
		for i := runStart; i <= end; i++ {
			setSimRankRange(data[i], hilo)
		}

		return end + 1
	}
}

// make ranked entries for later, calculating the sim rank as we go
// this uses the input array to make a new one, generic form would be a mapping operation
func AccuracyPrepareCalcHiLo(inputData []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyInfoPrepared {
	prepare := make([]*weight_types.AccuracyInfoPrepared, len(inputData))

	prepare[0] = &weight_types.AccuracyInfoPrepared{
		SimRankRange: &util_collection.HiLoInt{Lo: 0, Hi: 0},
		Stats:        inputData[0].DataStat,
	}
	for i := 1; i < len(inputData); i++ {
		if util.FloatsApproxEquals(inputData[i].SimScore, inputData[i-1].SimScore) {
			prevRange := prepare[i-1].SimRankRange
			prevRange.Hi = i
			prepare[i] = &weight_types.AccuracyInfoPrepared{SimRankRange: prevRange, Stats: inputData[i].DataStat}
		} else {
			newRange := &util_collection.HiLoInt{Lo: i, Hi: i}
			prepare[i] = &weight_types.AccuracyInfoPrepared{SimRankRange: newRange, Stats: inputData[i].DataStat}
		}
	}

	return prepare
}
