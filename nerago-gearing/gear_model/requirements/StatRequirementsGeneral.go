package requirements

import (
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_collection"
)

type StatRequirementsGeneral struct {
	lo    stats.StatBlock
	hi    stats.StatBlock
	asMap map[stats.StatType]util_collection.HiLoUInt32
}

func StatRequirementsGeneral_RetWideCap() *StatRequirementsGeneral {
	inst := &StatRequirementsGeneral{asMap: make(map[stats.StatType]util_collection.HiLoUInt32, 2), hi: fullHighBlock()}
	inst.addRange(stats.Stat_Hit, TARGET_RATING_MELEE, TARGET_RATING_MELEE+EXTREME_CAP_EXCEED)
	inst.addRange(stats.Stat_Expertise, TARGET_RATING_MELEE, TARGET_RATING_MELEE+EXTREME_CAP_EXCEED)
	return inst
}

func StatRequirementsGeneral_ProtFullExpertise() *StatRequirementsGeneral {
	inst := &StatRequirementsGeneral{asMap: make(map[stats.StatType]util_collection.HiLoUInt32, 2), hi: fullHighBlock()}
	inst.addRange(stats.Stat_Hit, TARGET_RATING_MELEE, TARGET_RATING_MELEE+DEFAULT_CAP_EXCEED)
	inst.addRange(stats.Stat_Expertise, TARGET_RATING_TANK, TARGET_RATING_TANK+DEFAULT_CAP_EXCEED)
	return inst
}

func StatRequirementsGeneral_ProtFlexibleParry() *StatRequirementsGeneral {
	inst := &StatRequirementsGeneral{asMap: make(map[stats.StatType]util_collection.HiLoUInt32, 2), hi: fullHighBlock()}
	inst.addRange(stats.Stat_Hit, TARGET_RATING_MELEE, TARGET_RATING_MELEE+DEFAULT_CAP_EXCEED)
	inst.addRange(stats.Stat_Expertise, TARGET_RATING_MELEE, TARGET_RATING_TANK) // TODO review this when we get falloff of value
	return inst
}

func StatRequirementsGeneral_ProtFlexibleParry_PlusAdditional(additional *stats.StatAndValue) *StatRequirementsGeneral {
	inst := &StatRequirementsGeneral{asMap: make(map[stats.StatType]util_collection.HiLoUInt32, 3), hi: fullHighBlock()}
	inst.addRange(stats.Stat_Hit, TARGET_RATING_MELEE, TARGET_RATING_MELEE+DEFAULT_CAP_EXCEED)
	inst.addRange(stats.Stat_Expertise, TARGET_RATING_MELEE, TARGET_RATING_TANK)
	if additional != nil {
		inst.addRange(additional.StatType, additional.Value, math.MaxUint32)
	}
	return inst
}

func StatRequirementsGeneral_None() *StatRequirementsGeneral {
	inst := &StatRequirementsGeneral{asMap: nil, hi: fullHighBlock()}
	return inst
}

func fullHighBlock() stats.StatBlock {
	highBlock := stats.StatBlock{}
	for i := range highBlock {
		highBlock[i] = math.MaxUint32
	}
	return highBlock
}

func (inst *StatRequirementsGeneral) addRange(statType stats.StatType, lo uint32, hi uint32) {
	inst.asMap[statType] = util_collection.HiLoUInt32{Lo: lo, Hi: hi}
	inst.lo[statType] = lo
	inst.hi[statType] = hi
}

func (inst *StatRequirementsGeneral) CheckSet(block *stats.StatBlock) bool {
	for i := range block {
		if block[i] < inst.lo[i] || block[i] > inst.hi[i] {
			return false
		}
	}
	return true
}

func (inst *StatRequirementsGeneral) EqualsTyped(other *StatRequirementsGeneral) bool {
	return inst.lo.Equals(&other.lo) && inst.hi.Equals(&other.hi)
}

func (inst *StatRequirementsGeneral) Equals(other any) bool {
	if typed, isThisType := other.(*StatRequirementsGeneral); isThisType {
		return inst.EqualsTyped(typed)
	} else {
		return false
	}
}

func (inst *StatRequirementsGeneral) IsLow(statType stats.StatType, value uint32) bool {
	return value < inst.lo[statType]
}

func (inst *StatRequirementsGeneral) IsHigh(statType stats.StatType, value uint32) bool {
	return value > inst.hi[statType]
}

func (inst *StatRequirementsGeneral) GetLow(statType stats.StatType) uint32 {
	return inst.lo[statType]
}

func (inst *StatRequirementsGeneral) GetHigh(statType stats.StatType) uint32 {
	return inst.hi[statType]
}

func (inst *StatRequirementsGeneral) AsMap() map[stats.StatType]util_collection.HiLoUInt32 {
	return inst.asMap
}
