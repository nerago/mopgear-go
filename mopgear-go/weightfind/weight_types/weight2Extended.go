package weight_types

import (
	"fmt"
	"iter"
	"math"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_rank"
)

// for extended stats planned calculation is:
// statA*weight1A + statB*weight1B + statC*weight1C = sim1
// statA*weight2A + statB*weight2B + statC*weight2C = sim2
// sim1*scale1+offset = 0-100 (better is higher)

// Weight2Extended
type Weight2Extended struct {
	DetailedWeights util_collection.MapMap[stats.SimType, stats.StatType, float64]
	SimList         []stats.SimType
	StatList        []stats.StatType
	SimPriority     SimPriorityExtended
}

func Weight2Extended_Make(simList []stats.SimType, statList []stats.StatType) *Weight2Extended {
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

func (we *Weight2Extended) PutWeight(simType stats.SimType, statType stats.StatType, weight float64) {
	we.DetailedWeights.Put(simType, statType, weight)
}

func (we *Weight2Extended) GetWeightOrPanic(simType stats.SimType, statType stats.StatType) float64 {
	return we.DetailedWeights.GetOrPanic(simType, statType)
}

func (we *Weight2Extended) SeqBySimThenStat() iter.Seq[util_collection.MapMapEntry[stats.SimType, stats.StatType, float64]] {
	return we.DetailedWeights.SeqKey2Key1ValueEntries()
}

func (we *Weight2Extended) SeqBySimNestedPairs() iter.Seq2[stats.SimType, iter.Seq2[stats.StatType, float64]] {
	return we.DetailedWeights.SeqKey1NestedKey2Value()
}

func (we *Weight2Extended) GetSimPriority() *SimPriorityExtended {
	return &we.SimPriority
}

func (we *Weight2Extended) SetSimScale(simType stats.SimType, rangingScale, rangingOffset, ratioScale float64) {
	we.SimPriority.SetSimScale(simType, rangingScale, rangingOffset, ratioScale)
}

func (we *Weight2Extended) FinishAndValidate(verificationInputs []WeightInput) {
	we.validateTypes()
	we.verifyGoodRange(verificationInputs)
}

func (we *Weight2Extended) FinishAndValidateNoVerify() {
	we.validateTypes()
}

func (we *Weight2Extended) CalcStatScoreForInput(input *WeightInput) float64 {
	return we.CalcStatScore(&input.TotalStat)
}

func (we *Weight2Extended) CalcStatScore(statBlock *stats.StatBlock) float64 {
	totalSum := 0.0
	for simType, nested := range we.DetailedWeights.SeqKey1NestedKey2Value() {
		subTotal := we.scoreForSimWeighted(statBlock, nested, simType)

		totalSum += subTotal
	}
	return totalSum
}

func (we *Weight2Extended) CalcStatScoreWithBonus(statBlock *stats.StatBlock, simBonus *stats.SimTypeMap[float64]) float64 {
	totalSum := 0.0
	for simType, nested := range we.DetailedWeights.SeqKey1NestedKey2Value() {
		subTotal := we.scoreForSim(nested, statBlock)

		priorityEntry := we.SimPriority.GetOrPanic(simType)
		subTotal = priorityEntry.Apply(subTotal)

		subTotal *= simBonus.GetOrDefault(simType, 1)

		totalSum += subTotal
	}
	return totalSum
}

func (we *Weight2Extended) scoreForSimWeighted(statBlock *stats.StatBlock, nested iter.Seq2[stats.StatType, float64], simType stats.SimType) float64 {
	subTotal := we.scoreForSim(nested, statBlock)

	priorityEntry := we.SimPriority.GetOrPanic(simType)
	return priorityEntry.Apply(subTotal)
}

func (we *Weight2Extended) scoreForSim(nested iter.Seq2[stats.StatType, float64], statBlock *stats.StatBlock) float64 {
	subTotal := 0.0
	for statType, detailWeight := range nested {
		specificValue := detailWeight * statBlock.GetFloat(statType)
		subTotal += specificValue
	}
	return subTotal
}

func (we *Weight2Extended) validateTypes() {
	for statType := range we.DetailedWeights.SeqKey2() {
		if !slices.Contains(we.StatList, statType) {
			panic("weight given for unlisted stat")
		}
	}
	for _, simType := range we.SimList {
		for _, statType := range we.StatList {
			if !we.DetailedWeights.Has(simType, statType) {
				panic("missing weight for " + statType.Name() + " " + simType.Name())
			}
		}
	}
	we.DetailedWeights.Foreach(func(simType stats.SimType, statType stats.StatType, value float64) {
		if !slices.Contains(we.StatList, statType) {
			panic("unexpected weight for " + statType.Name())
		}
		if !slices.Contains(we.SimList, simType) {
			panic("unexpected weight for " + simType.Name())
		}
	})

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

func (we *Weight2Extended) verifyGoodRange(verificationInputs []WeightInput) {
	if len(verificationInputs) == 0 {
		panic("no inputs for verification")
	}

	for _, simType := range we.SimList {
		loValue, hiValue, _, _ := we.actualOutputValueRangeForInputs(verificationInputs, simType)

		permittedSlack := 0.1
		if math.Abs(loValue) > permittedSlack || math.Abs(hiValue-1) > permittedSlack {
			panic(fmt.Sprintf("weights fail to produce expected value range, actual: %f - %f", loValue, hiValue))
		}
	}
}

func (we *Weight2Extended) UpdateScaling(inputData []WeightInput) {
	if len(inputData) == 0 {
		panic("no inputData for scaling")
	}

	targetLoValue := 0.0
	targetHiValue := 1.0

	for _, simType := range we.SimList {
		_, _, loValueRaw, hiValueRaw := we.actualOutputValueRangeForInputs(inputData, simType)
		oldPriorityEntry := we.SimPriority.GetOrPanic(simType)

		// lo+offset = targetLo
		offset := targetLoValue - loValueRaw
		// (hi+offset)*scale=targetHi
		// scale = targetHi / (hi+offset)
		scale := targetHiValue / (hiValueRaw + offset)

		we.SimPriority.Delete(simType)
		we.SimPriority.SetSimScale(simType, scale, offset, oldPriorityEntry.RatioScale)
	}
}

func (we *Weight2Extended) actualOutputValueRangeForInputs(verificationInputs []WeightInput, simType stats.SimType) (float64, float64, float64, float64) {
	lo := util_rank.BestCollector1[float64]{Minimise: true}
	hi := util_rank.BestCollector1[float64]{Minimise: false}
	priorityEntry := we.SimPriority.GetOrPanic(simType)

	weightSeq := we.DetailedWeights.SeqKey2ValueWithKey1(simType)
	for input := range util_collection.ForPointer(verificationInputs) {
		score := we.scoreForSim(weightSeq, &input.TotalStat)
		scoreWithRange := priorityEntry.ApplyRanging(score)
		lo.Offer(&score, scoreWithRange)
		hi.Offer(&score, scoreWithRange)
	}

	loValue := lo.BestValue
	loValueRaw := lo.GetBestOrNilValue()
	hiValue := hi.BestValue
	hiValueRaw := hi.GetBestOrNilValue()
	return loValue, hiValue, loValueRaw, hiValueRaw
}

func (we *Weight2Extended) Equals(other *Weight2Extended) bool {
	return slices.Equal(we.StatList, other.StatList) &&
		slices.Equal(we.SimList, other.SimList) &&
		we.SimPriority.Equals(&other.SimPriority) &&
		we.DetailedWeights.Equals(&other.DetailedWeights, func(a *float64, b *float64) bool { return *a == *b })
}

func (we *Weight2Extended) ConvertToWeight1() *Weight1Basic {
	weight1 := Weight1Basic_Make()

	for _, statType := range we.StatList {
		sumForStat := 0.0
		for simType, detailWeight := range we.DetailedWeights.SeqKey1ValueWithKey2(statType) {
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
