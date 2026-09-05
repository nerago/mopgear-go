package weight_types

import (
	"iter"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
)

// for extended stats planned calculation is:
// statA*weight1A + statB*weight1B + statC*weight1C = sim1
// statA*weight2A + statB*weight2B + statC*weight2C = sim2
// sim1*scale1+offset = 0-100 (better is higher)

// Weight2
type Weight2 struct {
	DetailedWeights util_collection.MapMap[stats.SimType, stats.StatType, float64]
	SimList         []stats.SimType
	StatList        []stats.StatType
	SimPriority     SimPriorityExtended
}

func Weight2Extended_Make(simList []stats.SimType, statList []stats.StatType) *Weight2 {
	return &Weight2{StatList: statList, SimList: simList, SimPriority: SimPriorityExtended_Make()}
}

func (we *Weight2) IsEmpty() bool {
	for value := range we.DetailedWeights.SeqValues() {
		if stats.IsUsefulWeightNumber(value) {
			return false
		}
	}
	return true
}

func (we *Weight2) PutWeight(simType stats.SimType, statType stats.StatType, weight float64) {
	we.DetailedWeights.Put(simType, statType, weight)
}

func (we *Weight2) GetWeightOrPanic(simType stats.SimType, statType stats.StatType) float64 {
	return we.DetailedWeights.GetOrPanic(simType, statType)
}

func (we *Weight2) SeqBySimThenStat() iter.Seq[util_collection.MapMapEntry[stats.SimType, stats.StatType, float64]] {
	return we.DetailedWeights.SeqKey2Key1ValueEntries()
}

func (we *Weight2) SeqBySimNestedPairs() iter.Seq2[stats.SimType, iter.Seq2[stats.StatType, float64]] {
	return we.DetailedWeights.SeqKey1NestedKey2Value()
}

func (we *Weight2) GetSimPriority() *SimPriorityExtended {
	return &we.SimPriority
}

func (we *Weight2) SetSimScale(simType stats.SimType, rangingScaleOffset ScaleAndOffset, ratioScale float64) error {
	return we.SimPriority.SetSimScale(simType, rangingScaleOffset, ratioScale)
}

func (we *Weight2) FinishAndValidate(sampleInputs []WeightInput) error {
	err := we.validateTypes()
	if err != nil {
		return err
	}

	err = we.verifyGoodRange(sampleInputs)
	if err != nil {
		return err
	}

	return nil
}

func (we *Weight2) FinishAndValidateNoVerify() error {
	return we.validateTypes()
}

func (we *Weight2) CalcStatScore(statBlock *stats.StatBlock) float64 {
	totalSum := 0.0
	for _, simType := range we.SimList {
		subTotal := we.scoreForSimWeighted(statBlock, simType)
		totalSum += subTotal
	}
	return totalSum
}

func (we *Weight2) CalcStatScoreRaw(statBlock *stats.StatBlock) float64 {
	totalSum := 0.0
	for _, simType := range we.SimList {
		subTotal := we.scoreForSimRaw(statBlock, simType)
		totalSum += subTotal
	}
	return totalSum
}

func (we *Weight2) CalcStatScoreWithBonus(statBlock *stats.StatBlock, simBonus *stats.SimTypeMap[float64]) float64 {
	totalSum := 0.0
	for _, simType := range we.SimList {
		subTotal := we.scoreForSimWeighted(statBlock, simType)
		subTotal *= simBonus.GetOrDefault(simType, 1)
		totalSum += subTotal
	}
	return totalSum
}

func (we *Weight2) scoreForSimWeighted(statBlock *stats.StatBlock, simType stats.SimType) float64 {
	subTotal := we.scoreForSimRaw(statBlock, simType)

	priorityEntry := we.SimPriority.GetOrPanic(simType)
	return priorityEntry.Apply(subTotal)
}

func (we *Weight2) scoreForSimRaw(statBlock *stats.StatBlock, simType stats.SimType) float64 {
	subTotal := 0.0
	for statType, detailWeight := range we.DetailedWeights.SeqKey2ValueWithKey1(simType) {
		subTotal += detailWeight * statBlock.GetFloat(statType)
	}
	return subTotal
}

func (we *Weight2) validateTypes() error {
	for statType := range we.DetailedWeights.SeqKey2() {
		if !slices.Contains(we.StatList, statType) {
			return util.ErrorTracedNew("weight given for unlisted stat " + statType.Name())
		}
	}
	for _, simType := range we.SimList {
		for _, statType := range we.StatList {
			if !we.DetailedWeights.Has(simType, statType) {
				return util.ErrorTracedNew("missing weight for " + statType.Name() + " " + simType.Name())
			}
		}
	}
	for simType, statType := range we.DetailedWeights.SeqKeysAll() {
		if !slices.Contains(we.StatList, statType) {
			return util.ErrorTracedNew("unexpected weight for " + statType.Name())
		}
		if !slices.Contains(we.SimList, simType) {
			return util.ErrorTracedNew("unexpected weight for " + simType.Name())
		}
	}

	for simType := range we.SimPriority.entries.SeqKey() {
		if !slices.Contains(we.SimList, simType) {
			return util.ErrorTracedNew("priority given for unlisted sim")
		}
	}
	for _, simType := range we.SimList {
		_, hasValue := we.SimPriority.Get(simType)
		if !hasValue {
			return util.ErrorTracedNew("priority missing for " + simType.Name())
		}
	}
	return nil
}

func (we *Weight2) Equals(other *Weight2) bool {
	return slices.Equal(we.StatList, other.StatList) &&
		slices.Equal(we.SimList, other.SimList) &&
		we.SimPriority.Equals(&other.SimPriority) &&
		we.DetailedWeights.Equals(&other.DetailedWeights, func(a *float64, b *float64) bool { return *a == *b })
}

func (we *Weight2) ConvertToWeight1(sampleInputs []WeightInput) *Weight1_ScaledSolvable {
	weight1 := Weight1_Make_ScaledSolvable()

	for _, statType := range we.StatList {
		sumForStat := 0.0
		for simType, detailWeight := range we.DetailedWeights.SeqKey1ValueWithKey2(statType) {
			simEntry := we.SimPriority.GetOrPanic(simType)
			componentValue := detailWeight * simEntry.Ranging.Scale * simEntry.RatioScale
			sumForStat += componentValue
		}
		weight1.Put(statType, sumForStat)
	}

	weight1.UpdateScaling(sampleInputs)
	return weight1
}

func (we *Weight2) Print(printer *util.PrintRecorder) {
	sb := util.StringBuild2{}
	we.AppendString(&sb)
	printer.PrintlnFromBuild(sb)
}

func (we *Weight2) String() string {
	sb := util.StringBuild2{}
	we.AppendString(&sb)
	return sb.String()
}

func (we *Weight2) AppendString(sb *util.StringBuild2) {
	sb.WriteString("(")
	for _, simType := range we.SimList {
		priority := we.SimPriority.GetOrPanic(simType)
		sb.WriteString(simType.Name())
		sb.WriteString("(scale1=")
		sb.WriteFloatScientific64(priority.RatioScale)
		sb.WriteString(",scale2=")
		sb.WriteFloatScientific64(priority.Ranging.Scale)
		sb.WriteString(",offset=")
		sb.WriteFloatScientific64(priority.Ranging.Offset)
		sb.WriteRune(',')
		for statType, value := range we.DetailedWeights.SeqKey2ValueWithKey1(simType) {
			sb.WriteString(statType.Name())
			sb.WriteRune(',')
			sb.WriteFloatScientific64(value)
			sb.WriteRune(',')
		}
		sb.Rewind(1)
		sb.WriteRune(')')
	}
	sb.WriteRune(')')
}
