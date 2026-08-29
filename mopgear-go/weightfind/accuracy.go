package weightfind

import (
	"cmp"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/simrank"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func EvaluateAccuracySwitch[W weight_types.IWeight](useStat bool, statWeights W, requiredSims []stats.SimType, simRatios *weight_types.SimPriorityBasic, inputData []weight_types.WeightInput) float64 {
	if useStat {
		return EvaluateAccuracyStatisticalExtended(statWeights, requiredSims, simRatios, inputData)
	} else {
		return EvaluateAccuracyBasic(statWeights, requiredSims, simRatios, inputData)
	}
}

func EvaluateAccuracyBasic[W weight_types.IWeight](statWeights W, requiredSims []stats.SimType, simRatios *weight_types.SimPriorityBasic, inputData []weight_types.WeightInput) float64 {
	if util.IsNil(statWeights) || statWeights.IsEmpty() {
		return 0
	}
	data := evaluateStatScoreAndCreateStructure(statWeights, inputData)
	sortStatScores(data)
	deriveStatRanksFromArrayOrder(data)
	simrank.RankSimsBasicForRanged(requiredSims, data, simRatios)
	return calcAverageDifference(data)
}

//func EvaluateAccuracyStatistical[W weight_types.IWeight](statWeights W, requiredSims []stats.SimType, simRatios *weight_types.SimPriorityBasic, inputData []weight_types.WeightInput) float64 {
//	if isNil(statWeights) || statWeights.IsEmpty() {
//		return 0
//	}
//	data := evaluateStatScoreAndCreateStructure(statWeights, inputData)
//	sortStatScores(data)
//	deriveStatRanksFromArrayOrder(data)
//	simrank.RankSimsStatisticalForRanged(requiredSims, data, simRatios)
//	return calcAverageDifference(data)
//}

func EvaluateAccuracyStatisticalExtended[W weight_types.IWeight](statWeights W, requiredSims []stats.SimType, simRatios *weight_types.SimPriorityBasic, inputData []weight_types.WeightInput) float64 {
	if util.IsNil(statWeights) || statWeights.IsEmpty() {
		return 0
	}
	data := evaluateStatScoreAndCreateStructureExtended(statWeights, inputData)
	sortStatScores(data)
	deriveStatRanksFromArrayOrder(data)
	simrank.RankSimsStatisticalForExtendedRanged(requiredSims, data, simRatios)
	return calcAverageDifference(data)
}

func sortStatScores[A weight_types.IRankStatFlatRange](data []A) {
	slices.SortFunc(data, func(a, b A) int {
		return cmp.Compare(a.GetStatScore(), b.GetStatScore())
	})
}

func evaluateStatScoreAndCreateStructure[W weight_types.IWeight](statWeights W, inputData []weight_types.WeightInput) []*weight_types.AccuracyInfo {
	return util_collection.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *weight_types.AccuracyInfo {
		return &weight_types.AccuracyInfo{
			DataSim:       &input.SimResult,
			StatScore:     statWeights.CalcStatScore(&input.TotalStat),
			SimScore:      0,
			StatRankRange: nil,
			SimRankRange:  nil,
		}
	})
}
func evaluateStatScoreAndCreateStructureExtended[W weight_types.IWeight](statWeights W, inputData []weight_types.WeightInput) []*weight_types.AccuracyInfoExtended {
	return util_collection.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *weight_types.AccuracyInfoExtended {
		return &weight_types.AccuracyInfoExtended{
			AccuracyInfo: weight_types.AccuracyInfo{
				DataSim:       &input.SimResult,
				StatScore:     statWeights.CalcStatScore(&input.TotalStat),
				SimScore:      0,
				StatRankRange: nil,
				SimRankRange:  nil,
			},
			//SimRankByType: zero,
		}
	})
}

func deriveStatRanksFromArrayOrder[A weight_types.IRankStatFlatRange](data []A) {
	data[0].SetStatRankRange(&util_collection.HiLoInt{Lo: 0, Hi: 0})
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].GetStatScore(), data[i-1].GetStatScore()) {
			prevRange := data[i-1].GetStatRankRange()
			data[i].SetStatRankRange(prevRange)
			prevRange.Hi = i
		} else {
			data[i].SetStatRankRange(&util_collection.HiLoInt{Lo: i, Hi: i})
		}
	}
}

func calcAverageDifference[A weight_types.IRankDoubleFlatRange](data []A) float64 {
	return calcAverageDifferenceNormal2(data)
}

func calcAverageDifferenceNormal2[A weight_types.IRankDoubleFlatRange](data []A) float64 {
	// compute average difference between stat rank and sim rank
	dataLength := float64(len(data))
	sumDiff := 0
	for i := range len(data) {
		entry := data[i]
		diff := entry.GetSimRankRange().Gap(*entry.GetStatRankRange())
		sumDiff += diff
	}
	averageDiff := float64(sumDiff) / dataLength
	relative := averageDiff / dataLength
	return 100.0 * (1.0 - relative)
}

//func calcAverageDifferenceDebug[A weight_types.IRankDoubleFlatRange](data []A) float64 {
//	// compute average difference between stat rank and sim rank
//	sumRatioScores := 0.0
//	for i := range len(data) {
//		entry := data[i]
//		var fullLength int = len(data)
//		diff := (*entry.GetSimRankRange()).Gap(*entry.GetStatRankRange())
//		ratioScore := float64(fullLength-diff) / float64(fullLength)
//
//		fmt.Printf("%d:  %d-%d %f %d-%d %f\n", i,
//			//entry.GetSimScore(),
//			entry.GetSimRankRange().Lo, entry.GetSimRankRange().Hi,
//			entry.GetStatScore(),
//			entry.GetStatRankRange().Lo, entry.GetStatRankRange().Hi,
//			ratioScore,
//		)
//		sumRatioScores += ratioScore
//	}
//	fmt.Printf("cad %f %d %f\n", sumRatioScores, len(data),
//		100.0*sumRatioScores/float64(len(data)),
//	)
//	return checkValue(100.0 * sumRatioScores / float64(len(data)))
//}

func checkValue(value float64) float64 {
	if value < 0 || value >= 100.0 {
		//panic("accuracy value out of expected range")
	}
	return value
}
