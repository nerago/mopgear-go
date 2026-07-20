package weight_types

import (
	"iter"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
)

// Weight2Extended
type Weight2Extended struct {
	detailedWeights util.MapMap[stats.StatType, stats.SimType, float64]
	statList        []stats.StatType
	simList         []stats.SimType
	simPriority     SimPriorityExtended
}

func Weight2Extended_Make(statList []stats.StatType, simList []stats.SimType) *Weight2Extended {
	return &Weight2Extended{statList: statList, simList: simList}
}

func (we *Weight2Extended) PutWeight(statType stats.StatType, simType stats.SimType, weight float64) {
	we.detailedWeights.Put(statType, simType, weight)
}

func (we *Weight2Extended) GetWeightOrPanic(statType stats.StatType, simType stats.SimType) float64 {
	return we.detailedWeights.GetOrPanic(statType, simType)
}

func (we *Weight2Extended) SeqBySimThenStat() iter.Seq[util.MapMapEntry[stats.StatType, stats.SimType, float64]] {
	return we.detailedWeights.SeqWithKeysOtherOrder()
}

func (we *Weight2Extended) SeqBySimNestedPairs() iter.Seq2[stats.SimType, iter.Seq2[stats.StatType, float64]] {
	return we.detailedWeights.SeqGroupsKey2NestedKeyValue()
}

func (we *Weight2Extended) SimPriority() *SimPriorityExtended {
	return &we.simPriority
}

func (we *Weight2Extended) FinishAndValidate() {
	we.validate()
	we.scaleEachSimForBase()
}

func (we *Weight2Extended) validate() {
	for statType := range we.detailedWeights.SeqKey1() {
		if !slices.Contains(we.statList, statType) {
			panic("weight given for unlisted stat")
		}
	}
	for _, simType := range we.simList {
		for _, statType := range we.statList {
			if !we.detailedWeights.Has(statType, simType) {
				panic("missing weight for " + statType.Name() + " " + simType.Name())
			}
		}
	}
}

func (we *Weight2Extended) scaleEachSimForBase() {
	baseStat := we.statList[0]
	for _, simType := range we.simList {
		baseValue := we.detailedWeights.GetOrPanic(baseStat, simType)
		for _, statType := range we.statList {
			we.detailedWeights.Apply(statType, simType, func(oldValue float64) float64 {
				return oldValue / baseValue
			})
		}
	}
}

func (we *Weight2Extended) ConvertToWeight1() Weight1Basic {
	// NOTE: assuming that scaleEachSimForBase has run, all on equal basis
	weight1 := Weight1Basic_Make(we.simPriority.ConvertToBasic())
	for _, statType := range we.statList {
		sumIndividual := 0.0
		for simType, thisDetailWeight := range we.detailedWeights.SeqInnerWithKey1Value(statType) {
			simEntry := we.simPriority.GetOrPanic(simType)

			componentValue := thisDetailWeight * simEntry.Scale
			if simEntry.Scale < 0 {
				componentValue *= -1
			}

			sumIndividual += componentValue
		}
		weight1.Put(statType, sumIndividual)
	}

	weight1 = weight1.ScaleForBaseStat(we.statList[0])
	return weight1
}
