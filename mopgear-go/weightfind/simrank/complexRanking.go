package simrank

import (
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type averageAndStdDev struct {
	average float64
	stdDev  float64
}

func averageAndStdDevMake[T weight_types.IRankEntry](entry T, simType stats.SimType) averageAndStdDev {
	return averageAndStdDev{
		average: entry.GetSimData().Get(simType),
		stdDev:  entry.GetSimData().GetStdDevOrZero(simType),
	}
}

func arrayFindChainFromFunc[T weight_types.IRankEntry](data []T, runStart int, simType stats.SimType) (int, int) {
	countChainedRun := 1
	countEqualFirstRun := 1

	firstScore := averageAndStdDevMake(data[runStart], simType)
	prevScore := firstScore

	for check := runStart + 1; check < len(data); check++ {
		currScore := averageAndStdDevMake(data[check], simType)
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

func arrayApplyChainFunc[T weight_types.IRankEntryExtendedRangeInt](data []T, runStart int, countChainedRun int, countEqualFirstRun int, simType stats.SimType) int {
	if countChainedRun == 1 {
		data[runStart].SetSimRankRangeByType(simType, runStart, runStart)
		return runStart + 1
	} else {
		var end int
		if countChainedRun == countEqualFirstRun || countChainedRun <= 4 || countEqualFirstRun >= countChainedRun*3/4 {
			end = runStart + countChainedRun - 1
		} else {
			end = runStart + countEqualFirstRun - 1
		}

		for i := runStart; i <= end; i++ {
			data[i].SetSimRankRangeByType(simType, runStart, end)
		}

		return end + 1
	}
}

func sortSimRankComplex[T weight_types.IRankEntryExtendedRangeFloat](dataSlice []T) {
	slices.SortFunc(dataSlice, compareForComplexSummary)
}

func multiplyFloatRangesByRatio[T weight_types.IRankEntryExtendedRangeFloat](dataSlice []T, ratios *weight_types.SimPriorityBasic) {
	for _, data := range dataSlice {
		for simType, ratio := range ratios.SeqTypeValue() {
			currRange := data.GetSimRankRangeFloatByType(simType)
			data.SetSimRankRangeFloatByType(
				simType,
				currRange.Lo*ratio,
				currRange.Hi*ratio,
			)
		}
	}
}

func compareForComplexSummary[T weight_types.IRankEntryExtendedRangeFloat](one T, two T) int {
	totalDiff := complexSummaryDiff(one, two)
	return diffSignToCmp(totalDiff)
}

func complexSummaryDiff[T weight_types.IRankEntryExtendedRangeFloat](one T, two T) float64 {
	totalDiff := 0.0
	for simType, oneHiLo := range one.SeqSimRankRangeFloatByType() {
		twoHiLo := two.GetSimRankRangeFloatByType(simType)
		diff := oneHiLo.GapSigned(twoHiLo)
		totalDiff += diff
	}
	return totalDiff
}

func diffSignToCmp(totalDiff float64) int {
	if totalDiff < 0 {
		return -1
	} else if totalDiff > 0 {
		return 1
	} else {
		return 0
	}
}

func arrayRankToSetSimRankRangeComplexCompare[T weight_types.IRankEntryExtendedRangeAndSummary](data []T) {
	data[0].SetSimRankRange(&util_collection.HiLoInt{Lo: 0, Hi: 0})
	for i := 1; i < len(data); i++ {
		diff := complexSummaryDiff(data[i], data[i-1])
		if util.FloatEqualsZero(diff) {
			prevRange := data[i-1].GetSimRankRange()
			data[i].SetSimRankRange(prevRange)
			prevRange.Hi = i
		} else {
			data[i].SetSimRankRange(&util_collection.HiLoInt{Lo: i, Hi: i})
		}
	}
}
