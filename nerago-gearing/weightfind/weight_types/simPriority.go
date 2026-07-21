package weight_types

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

//type WeightAlternateSimPriority struct {
//	orderedList []AlternateSimPriority
//}
//type AlternateSimPriority struct {
//	SimType                    stats.SimType
//	CompromisePermittedPercent float64
//}

type SimPriorityBasic struct {
	content util.EnumMap[stats.SimType, float64]
}

func SimPriorityBasic_MakeEmpty(parts ...any) SimPriorityBasic {
	return SimPriorityBasic{
		content: util.EnumMapMake[stats.SimType, float64](stats.SimTypeEnum),
	}
}
func SimPriorityBasic_Make(parts ...any) SimPriorityBasic {
	sim := SimPriorityBasic{
		content: util.EnumMapMake[stats.SimType, float64](stats.SimTypeEnum),
	}
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

func (sr *SimPriorityBasic) Equals(other *SimPriorityBasic) bool {
	return sr.content.Equals(&other.content, func(a *float64, b *float64) bool { return *a == *b })
}

func (sr *SimPriorityBasic) ScaleForTotalSum(targetTotal float64) *SimPriorityBasic {
	currentTotal := 0.0
	for value := range sr.content.SeqValue() {
		currentTotal += value
	}
	scale := targetTotal / currentTotal

	result := SimPriorityBasic_MakeEmpty()
	for simType := range sr.content.SeqKey() {
		value, hasValue := sr.content.Get(simType)
		if hasValue {
			result.content.Put(simType, value*scale)
		}
	}
	return &result
}

func (sr *SimPriorityBasic) ValidateRatioAddsToOne() {
	currentTotal := 0.0
	for value := range sr.content.SeqValue() {
		if value < 0 {
			panic("negative sim ratio not supported")
		}
		currentTotal += value
	}
	if !util.FloatEqualsOne(currentTotal) {
		panic("ratios don't add to one")
	}
}

// for extended stats planned calculation is:
// statA*weight1A + statB*weight1B + statC*weight1C = sim1
// statA*weight2A + statB*weight2B + statC*weight2C = sim2
// (sim1+offset1)*scale1 = 0-1.0 (better is higher)
type SimPriorityExtended struct {
	entries util.EnumMap[stats.SimType, SimPriorityEntry]
}
type SimPriorityEntry struct {
	RangingScale  float64 // calculates values so that range is consistent (e.g. 0-1.0)
	RangingOffset float64
	RatioScale    float64 // relative factor to other sim entries to establish priority
}

func (se SimPriorityEntry) Apply(subtotal float64) float64 {
	return (subtotal + se.RangingOffset) * se.RangingScale * se.RatioScale
}

func SimPriorityExtended_Make() SimPriorityExtended {
	return SimPriorityExtended{util.EnumMapMake[stats.SimType, SimPriorityEntry](stats.SimTypeEnum)}
}

func (sre *SimPriorityExtended) checkInitialized() {
	if sre == nil || sre.entries.IsUninitialized() {
		panic("SimPriorityExtended not initialized")
	}
}

func (sre *SimPriorityExtended) Get(simType stats.SimType) (SimPriorityEntry, bool) {
	sre.checkInitialized()
	return sre.entries.Get(simType)
}

func (sre *SimPriorityExtended) GetOrPanic(simType stats.SimType) SimPriorityEntry {
	sre.checkInitialized()
	entry, hasEntry := sre.entries.Get(simType)
	if !hasEntry {
		panic("missing entry")
	}
	return entry
}

func (sre *SimPriorityExtended) SetSimScale(simType stats.SimType, rangingScale, rangingOffset, ratioScale float64) {
	sre.checkInitialized()
	if sre.entries.Has(simType) {
		panic("duplicate")
	}
	sre.entries.Put(simType, SimPriorityEntry{
		RangingScale:  rangingScale,
		RangingOffset: rangingOffset,
		RatioScale:    ratioScale,
	})
}

func (sre *SimPriorityExtended) Validate() {
	sre.checkInitialized()
}

func (sre *SimPriorityExtended) ConvertToBasic() SimPriorityBasic {
	sre.checkInitialized()
	simRatio := SimPriorityBasic_MakeEmpty()
	for simType, entry := range sre.entries.SeqKeyValue() {
		simRatio.Set(simType, entry.RatioScale)
	}
	simRatio = *simRatio.ScaleForTotalSum(1.0)
	return simRatio
}
