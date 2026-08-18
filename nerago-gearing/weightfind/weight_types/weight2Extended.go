package weight_types

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"iter"
	"slices"
)

// for extended stats planned calculation is:
// statA*weight1A + statB*weight1B + statC*weight1C = sim1
// statA*weight2A + statB*weight2B + statC*weight2C = sim2
// sim1*scale1+offset = 0-100 (better is higher)

// Weight2Extended
type Weight2Extended struct {
	DetailedWeights util_collection.MapMap[stats.StatType, stats.SimType, float64]
	StatList        []stats.StatType
	SimList         []stats.SimType
	SimPriority     SimPriorityExtended
}

func Weight2Extended_Make(statList []stats.StatType, simList []stats.SimType) *Weight2Extended {
	return &Weight2Extended{StatList: statList, SimList: simList, SimPriority: SimPriorityExtended_Make()}
}

func (we *Weight2Extended) IsEmpty() bool {
	for value := range we.DetailedWeights.SeqValues() {
		if stats.IsUsefulWeightNumber(value) {
			return false
		}
	}
	return true
}

func (we *Weight2Extended) PutWeight(statType stats.StatType, simType stats.SimType, weight float64) {
	we.DetailedWeights.Put(statType, simType, weight)
}

func (we *Weight2Extended) GetWeightOrPanic(statType stats.StatType, simType stats.SimType) float64 {
	return we.DetailedWeights.GetOrPanic(statType, simType)
}

func (we *Weight2Extended) SeqBySimThenStat() iter.Seq[util_collection.MapMapEntry[stats.StatType, stats.SimType, float64]] {
	return we.DetailedWeights.SeqKey2Key1ValueEntries()
}

func (we *Weight2Extended) SeqBySimNestedPairs() iter.Seq2[stats.SimType, iter.Seq2[stats.StatType, float64]] {
	return we.DetailedWeights.SeqKey2NestedKey1Value()
}

func (we *Weight2Extended) GetSimPriority() *SimPriorityExtended {
	return &we.SimPriority
}

func (we *Weight2Extended) SetSimScale(simType stats.SimType, rangingScale, rangingOffset, ratioScale float64) {
	we.SimPriority.SetSimScale(simType, rangingScale, rangingOffset, ratioScale)
}

func (we *Weight2Extended) FinishAndValidate() {
	we.validate()
	//we.scaleEachSimForBase()
}

func (we *Weight2Extended) CalcStatScoreForInput(input *WeightInput) float64 {
	return we.CalcStatScore(&input.TotalStat)
}

func (we *Weight2Extended) CalcStatScore(statBlock *stats.StatBlock) float64 {
	totalSum := 0.0
	for simType, nested := range we.DetailedWeights.SeqKey2NestedKey1Value() {
		subTotal := scoreForSim2(nested, statBlock)

		priorityEntry := we.SimPriority.GetOrPanic(simType)
		subTotal = priorityEntry.Apply(subTotal)

		totalSum += subTotal
	}
	return totalSum
}

func (we *Weight2Extended) CalcStatScoreWithBonus(statBlock *stats.StatBlock, simBonus *stats.SimTypeMap[float64]) float64 {
	totalSum := 0.0
	for simType, nested := range we.DetailedWeights.SeqKey2NestedKey1Value() {
		subTotal := scoreForSim2(nested, statBlock)

		priorityEntry := we.SimPriority.GetOrPanic(simType)
		subTotal = priorityEntry.Apply(subTotal)

		subTotal *= simBonus.GetOrDefault(simType, 1)

		totalSum += subTotal
	}
	return totalSum
}

func scoreForSim2(nested iter.Seq2[stats.StatType, float64], statBlock *stats.StatBlock) float64 {
	subTotal := 0.0
	for statType, detailWeight := range nested {
		specificValue := detailWeight * statBlock.GetFloat(statType)
		subTotal += specificValue
	}
	return subTotal
}

func (we *Weight2Extended) validate() {
	for statType := range we.DetailedWeights.SeqKey1() {
		if !slices.Contains(we.StatList, statType) {
			panic("weight given for unlisted stat")
		}
	}
	for _, simType := range we.SimList {
		for _, statType := range we.StatList {
			if !we.DetailedWeights.Has(statType, simType) {
				panic("missing weight for " + statType.Name() + " " + simType.Name())
			}
		}
	}

	for simType := range we.SimPriority.entries.SeqKey() {
		if !slices.Contains(we.SimList, simType) {
			panic("priority given for unlisted sim")
		}
	}
	for _, simType := range we.SimList {
		_, hasValue := we.SimPriority.Get(simType)
		if !hasValue {
			panic("priority missing for " + simType.Name())
		}
	}
}

func (we *Weight2Extended) Equals(other *Weight2Extended) bool {
	return slices.Equal(we.StatList, other.StatList) &&
		slices.Equal(we.SimList, other.SimList) &&
		we.SimPriority.Equals(&other.SimPriority) &&
		we.DetailedWeights.Equals(&other.DetailedWeights, func(a *float64, b *float64) bool { return *a == *b })
}

func (we *Weight2Extended) ConvertToWeight1() *Weight1Basic {
	// NOTE: assuming that scaleEachSimForBase has run, all on equal basis
	weight1 := Weight1Basic_Make()

	for _, statType := range we.StatList {
		sumForStat := 0.0
		for simType, detailWeight := range we.DetailedWeights.SeqKey2ValueWithKey1(statType) {
			simEntry := we.SimPriority.GetOrPanic(simType)
			componentValue := detailWeight * simEntry.RangingScale * simEntry.RatioScale
			sumForStat += componentValue
		}
		weight1.Put(statType, sumForStat)
	}

	weight1.NormalizeForBase(we.StatList)
	return &weight1
}

func (we *Weight2Extended) Print(printer *util.PrintRecorder) {
	sb := util.StringBuild2{}
	we.AppendString(&sb)
	printer.PrintlnFromBuild(sb)
}

func (we *Weight2Extended) String() string {
	sb := util.StringBuild2{}
	we.AppendString(&sb)
	return sb.String()
}

func (we *Weight2Extended) AppendString(sb *util.StringBuild2) {
	sb.WriteString("(")
	for _, simType := range we.SimList {
		priority := we.SimPriority.GetOrPanic(simType)
		sb.WriteString(simType.Name())
		sb.WriteString("(scale1=")
		sb.WriteFloatScientific64(priority.RatioScale)
		sb.WriteString(",scale2=")
		sb.WriteFloatScientific64(priority.RangingScale)
		sb.WriteString(",offset=")
		sb.WriteFloatScientific64(priority.RangingOffset)
		sb.WriteRune(',')
		for statType, value := range we.DetailedWeights.SeqKey1ValueWithKey2(simType) {
			sb.WriteString(statType.Name())
			sb.WriteRune('=')
			sb.WriteFloatScientific64(value)
			sb.WriteRune(',')
		}
		sb.Rewind(1)
		sb.WriteRune(')')
	}
	sb.WriteRune(')')
}
