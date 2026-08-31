package weight_types

import (
	"iter"
	"math"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_rank"
)

const c_priorityRounding = 4

type SimPriorityBasic struct {
	content stats.SimTypeMap[float64]
}

func SimPriorityBasic_Make(parts ...any) SimPriorityBasic {
	sim := SimPriorityBasic{}
	for i := 0; i < len(parts); i += 2 {
		simType := parts[i].(stats.SimType)

		switch value := parts[i+1].(type) {
		case int:
			sim.Set(simType, float64(value))
		case float64:
			sim.Set(simType, value)
		}
	}
	return sim
}

func (sr *SimPriorityBasic) IsEmpty() bool {
	for value := range sr.content.SeqValues() {
		if stats.IsUsefulWeightNumber(value) {
			return false
		}
	}
	return true
}

func (sr *SimPriorityBasic) Set(simType stats.SimType, value float64) {
	sr.content.Put(simType, value)
}

func (sr *SimPriorityBasic) Get(simType stats.SimType) (float64, bool) {
	return sr.content.Get(simType)
}

func (sr *SimPriorityBasic) GetOrPanic(simType stats.SimType) float64 {
	value, _ := sr.content.Get(simType)
	return value
}

func (sr *SimPriorityBasic) SimTypes() []stats.SimType {
	return sr.content.KeySlice()
}

func (sr *SimPriorityBasic) SeqTypeValue() iter.Seq2[stats.SimType, float64] {
	return sr.content.SeqKeyValue()
}

func (sr *SimPriorityBasic) Equals(other *SimPriorityBasic) bool {
	return sr.content.Equals(&other.content, func(a *float64, b *float64) bool { return *a == *b })
}

func (sr *SimPriorityBasic) ScaleForTotalSum(targetTotal float64) error {
	currentTotal, err := sr.currentTotal()
	if err != nil {
		return err
	}

	if currentTotal == 0 {
		return util.ErrorTracedNew("can't scale empty ratio")
	} else if !util.FloatEqualsOne(currentTotal) {
		sr.scaleForMathSum(currentTotal, targetTotal)
		sr.fixSumWithRounding(targetTotal)
	}
	return nil
}

func (sr *SimPriorityBasic) fixSumWithRounding(targetTotal float64) {
	tensMultiplier := math.Pow10(c_priorityRounding)
	biggestPart := util_rank.BestCollector1Lite[stats.SimType]{}

	roundedTotal := 0.0
	for simType, value := range sr.content.SeqKeyValue() {
		roundedTotal += math.Round(value*tensMultiplier) / tensMultiplier
		biggestPart.Offer(simType, value)
	}

	adjust := targetTotal - roundedTotal
	biggestSimValue := biggestPart.GetBestOrPanic()
	sr.content.Compute(biggestSimValue, func(v float64) float64 { return v + adjust })
}

func (sr *SimPriorityBasic) scaleForMathSum(currentTotal, targetTotal float64) {
	scale := targetTotal / currentTotal

	for simType := range sr.content.SeqKey() {
		value, hasValue := sr.content.Get(simType)
		if hasValue {
			sr.content.Put(simType, value*scale)
		}
	}
}

func (sr *SimPriorityBasic) ValidateRatioAddsToOne() error {
	currentTotal, err := sr.currentTotal()
	if err != nil {
		return err
	}

	if !util.FloatEqualsOne(currentTotal) {
		return util.ErrorTracedNew("ratios don't add to one")
	}
	return nil
}

func (sr *SimPriorityBasic) currentTotal() (float64, error) {
	currentTotal := 0.0
	for value := range sr.content.SeqValues() {
		if value < 0 {
			return 0, util.ErrorTracedNew("negative sim ratio not supported")
		}
		currentTotal += value
	}
	return currentTotal, nil
}

func (sr *SimPriorityBasic) String() string {
	sb := util.StringBuild2{}
	sr.AppendString(&sb)
	return sb.String()
}

func (sr *SimPriorityBasic) AppendString(sb *util.StringBuild2) {
	for typ, value := range sr.content.SeqKeyValue() {
		sb.WriteString(typ.Name())
		sb.WriteRune('=')
		sb.WriteFloat64(value, c_priorityRounding)
		sb.WriteRune(' ')
	}
	sb.Rewind(1)
}

// for extended stats planned calculation is:
// statA*weight1A + statB*weight1B + statC*weight1C = sim1
// statA*weight2A + statB*weight2B + statC*weight2C = sim2
// (sim1+offset1)*scale1 = 0-1.0 (better is higher)
type SimPriorityExtended struct {
	entries stats.SimTypeMap[SimPriorityEntry]
}
type SimPriorityEntry struct {
	RangingScale  float64 // calculates values so that range is consistent (e.g. 0-1.0)
	RangingOffset float64
	RatioScale    float64 // relative factor to other sim entries to establish priority
}

func (se *SimPriorityEntry) Equals(other *SimPriorityEntry) bool {
	return se.RangingScale == other.RangingScale &&
		se.RangingOffset == other.RangingOffset &&
		se.RatioScale == other.RatioScale
}

func (se *SimPriorityEntry) Apply(subtotal float64) float64 {
	return (subtotal + se.RangingOffset) * se.RangingScale * se.RatioScale
}

func (se *SimPriorityEntry) ApplyRanging(subtotal float64) float64 {
	return (subtotal + se.RangingOffset) * se.RangingScale
}

func SimPriorityExtended_Make() SimPriorityExtended {
	return SimPriorityExtended{}
}

func (sre *SimPriorityExtended) IsEmpty() bool {
	for entry := range sre.entries.SeqValues() {
		if stats.IsUsefulWeightNumber(entry.RatioScale) &&
			stats.IsUsefulWeightNumber(entry.RangingScale) &&
			stats.IsValidWeightNumber(entry.RangingOffset) {
			return false
		}
	}
	return true
}

func (sre *SimPriorityExtended) Clone() SimPriorityExtended {
	return SimPriorityExtended{*sre.entries.Clone()}
}

func (sre *SimPriorityExtended) Get(simType stats.SimType) (SimPriorityEntry, bool) {
	return sre.entries.Get(simType)
}

func (sre *SimPriorityExtended) GetOrPanic(simType stats.SimType) SimPriorityEntry {
	entry, hasEntry := sre.entries.Get(simType)
	if !hasEntry {
		panic("missing entry")
	}
	return entry
}

func (sre *SimPriorityExtended) SetSimScale(simType stats.SimType, rangingScale, rangingOffset, ratioScale float64) error {
	if sre.entries.Has(simType) {
		return util.ErrorTracedNew("duplicate")
	}
	sre.entries.Put(simType, SimPriorityEntry{
		RangingScale:  rangingScale,
		RangingOffset: rangingOffset,
		RatioScale:    ratioScale,
	})
	return nil
}

func (sre *SimPriorityExtended) Validate() {
}

func (sre *SimPriorityExtended) ConvertToBasic() (SimPriorityBasic, error) {
	simRatio := SimPriorityBasic{}
	for simType, entry := range sre.entries.SeqKeyValue() {
		simRatio.Set(simType, entry.RatioScale)
	}
	err := simRatio.ScaleForTotalSum(1.0)
	return simRatio, err
}

func (sre *SimPriorityExtended) Equals(other *SimPriorityExtended) bool {
	return sre.entries.Equals(&other.entries, (*SimPriorityEntry).Equals)
}

func (sre *SimPriorityExtended) Delete(simType stats.SimType) {
	sre.entries.Delete(simType)
}
