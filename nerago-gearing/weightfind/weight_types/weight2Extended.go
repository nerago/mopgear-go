package weight_types

import (
	"iter"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
)

// Weight2Extended
type Weight2Extended struct {
	detailedWeights   util.MapMap[stats.StatType, stats.SimType, float64]
	statList          []stats.StatType
	simList           []stats.SimType
	simRatioWeighting stats.SimData // should also include component to bring values to similar size
}

func Weight2Extended_Make(simRatio stats.SimData, statList []stats.StatType) Weight2Extended {
	return Weight2Extended{simRatioWeighting: simRatio, statList: statList}
}

func (we *Weight2Extended) Put(statType stats.StatType, simType stats.SimType, weight float64) {
	we.detailedWeights.Put(statType, simType, weight)
}

func (we *Weight2Extended) GetOrPanic(statType stats.StatType, simType stats.SimType) float64 {
	return we.detailedWeights.GetOrPanic(statType, simType)
}

func (we *Weight2Extended) SeqBySimThenStat() iter.Seq[util.MapMapEntry[stats.StatType, stats.SimType, float64]] {
	return we.detailedWeights.SeqWithKeysOtherOrder()
}

func (we *Weight2Extended) FinishedMake() {
	we.validate()
	we.scaleEachSimForBase()
}

func (we *Weight2Extended) validate() {
	for statType := range we.detailedWeights.SeqKey1() {
		if !slices.Contains(we.statList, statType) {
			panic("weight given for unlisted stat")
		}
	}
}

func (we *Weight2Extended) scaleEachSimForBase() {

}

func (we *Weight2Extended) ConvertToWeight1() Weight1Basic {
	weight1 := Weight1Basic_Make()
	// do we want to scale all the detailed on basestat first
	for _, statType := range we.statList {
		sumIndividual := 0.0

		for simType, thisDetailWeight := range we.detailedWeights.SeqInnerWithKey1Value(statType) {
			targetRatio := we.simRatioWeighting.Get(simType)
			componentValue := thisDetailWeight * targetRatio

			if !simType.IsHighGood() {
				componentValue *= -1
			}

			//componentValue := targetRatio * thisDetailWeight / strengthDetailWeight

			sumIndividual += componentValue
		}
		weight1.Put(statType, sumIndividual)
	}

	weight1 = weight1.ScaleForBaseStat(we.statList[0])
	return weight1
}

//func (grid2 *GridStatWeightProcess2) reportOutputWeightsGrid(solution *highs.Solution) weight_types.WeightBasic {
//	result := weight_types.WeightBasic_Make()
//	for _, statType := range grid2.statTypes {
//		grid2.printer.Printf("%10s >>>>>\n", statType.Name())
//
//		sumIndividual := 0.0
//
//		for simType, detailWeightCol := range grid2.detailedWeights.SeqInnerWithKey1Value(statType) {
//			weight := solution.ColValues[detailWeightCol]
//			usableWeight := weight / grid2.scaleSims[simType]
//
//			if !simType.IsHighGood() {
//				usableWeight *= -1
//			}
//
//			grid2.printer.Printf("         %5s > %f %f\n", simType.Name(), weight, usableWeight)
//
//			sumIndividual += usableWeight * grid2.targetRatios.Get(simType)
//		}
//
//		grid2.printer.Printf("             === %f\n", sumIndividual)
//		result.Put(statType, sumIndividual)
//	}
//
//	baseStat := grid2.statTypes[0]
//	divideBy := result.Get(baseStat)
//	for _, statType := range grid2.statTypes {
//		result.Put(statType, result.Get(statType)/divideBy)
//	}
//
//	return result
//}

//	baseStat := fiteach.requiredStats[0]
//	standardResult := weight_types.WeightBasic_Make()
//	standardResult.Put(baseStat, 1)
//	for _, statType := range fiteach.requiredStats {
//		if statType != baseStat {
//			totalSum := 0.0
//			for _, simType := range fiteach.requiredSims {
//				thisRating := bestRatingEach.GetOrPanic(statType, simType)
//				strengthRating := bestRatingEach.GetOrPanic(stats.Stat_Strength, simType)
//				relative := thisRating / strengthRating * fiteach.targetRatios.Get(simType)
//				totalSum += relative
//			}
//			standardResult.Put(statType, totalSum)
//		}
//	}
//	return standardResult
//}
