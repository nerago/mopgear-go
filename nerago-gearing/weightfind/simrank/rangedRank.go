package simrank

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_types"
)

func RankSimsRegularForAccuracyRanged(requiredSims []stats.SimType, data []*weight_types.AccuracyInfoSimStatRanged, simRatios *weight_types.SimPriorityBasic) {
	simScoringBasicFastForAccuracy(requiredSims, simRatios, data)
	sortSimScores(data)
	arrayRankToSetSimRankRange(data)
}

func RankSimsStatisticalForAccuracyRanged(requiredSims []stats.SimType, data []*weight_types.AccuracyInfoSimStatRanged, simRatios *weight_types.SimPriorityBasic) {
	simScoringStatisticalForAccuracy(requiredSims, simRatios, data)
	sortSimScores(data)
	arrayRankToSetSimRankRange(data)
}

func arrayRankToSetSimRankRange[T weight_types.IRankEntryFlatRange](data []T) {
	data[0].SetSimRankRange(&util.HiLoInt{Lo: 0, Hi: 0})
	for rank := 1; rank < len(data); rank++ {
		if util.FloatsApproxEquals(data[rank].GetSimScore(), data[rank-1].GetSimScore()) {
			prevRange := data[rank-1].GetSimRankRange()
			data[rank].SetSimRankRange(prevRange)
			prevRange.Hi = rank
		} else {
			data[rank].SetSimRankRange(&util.HiLoInt{Lo: rank, Hi: rank})
		}
	}
}

func arrayRankToSetSimRankRangeComplicated[T weight_types.IRankEntryFlatRange](data []T) {
	for runStart := 0; runStart < len(data); {
		countChainedRun := 1
		countEqualFirstRun := 1

		firstScore := data[runStart].GetSimScore()
		prevScore := firstScore

		for check := runStart + 1; check < len(data); check++ {
			currScore := data[check].GetSimScore()
			if util.FloatsApproxEquals(currScore, firstScore) {
				countEqualFirstRun++
				countChainedRun++
			} else if util.FloatsApproxEquals(currScore, prevScore) {
				countChainedRun++
			} else {
				break
			}
			prevScore = currScore
		}

		if countChainedRun == 1 {
			data[runStart].SetSimRankRange(&util.HiLoInt{Lo: runStart, Hi: runStart})
			runStart++
		} else {
			var end int
			if countChainedRun == countEqualFirstRun || countChainedRun <= 4 || countEqualFirstRun >= countChainedRun*3/4 {
				end = runStart + countChainedRun - 1
			} else {
				end = runStart + countEqualFirstRun - 1
			}

			hilo := &util.HiLoInt{Lo: runStart, Hi: end}
			for i := runStart; i <= end; i++ {
				data[i].SetSimRankRange(hilo)
			}

			runStart = end + 1
		}
	}
}

// make ranked entries for later, calculating the sim rank as we go
// this uses the input array to make a new one, generic form would be a mapping operation
func AccuracyPrepareCalcHiLo(inputData []*weight_types.AccuracyInfoPrePrepare) []*weight_types.AccuracyPreparedEntry {
	prepare := make([]*weight_types.AccuracyPreparedEntry, len(inputData))

	prepare[0] = &weight_types.AccuracyPreparedEntry{
		SimRankRange: &util.HiLoInt{Lo: 0, Hi: 0},
		Stats:        inputData[0].DataStat,
	}
	for i := 1; i < len(inputData); i++ {
		if util.FloatsApproxEquals(inputData[i].SimScore, inputData[i-1].SimScore) {
			prevRange := prepare[i-1].SimRankRange
			prevRange.Hi = i
			prepare[i] = &weight_types.AccuracyPreparedEntry{SimRankRange: prevRange, Stats: inputData[i].DataStat}
		} else {
			newRange := &util.HiLoInt{Lo: i, Hi: i}
			prepare[i] = &weight_types.AccuracyPreparedEntry{SimRankRange: newRange, Stats: inputData[i].DataStat}
		}
	}

	return prepare
}
