package requirements

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util/util_collection"
	"math"
)

type StatRequirementsHitExpertise struct {
	hitMin, hitMax uint32
	expMin, expMax uint32

	AdditionalMinimumRequirement *stats.StatAndValue

	asMap map[stats.StatType]util_collection.HiLoUInt32
}

const (
	// RATING_PER_PERCENT    float64 = 339.9534
	// TARGET_PERCENT_MELEE  float64 = 7.5
	TARGET_RATING_MELEE uint32 = 2550
	TARGET_RATING_TANK  uint32 = 5100
	TARGET_RATING_CAST  uint32 = 5100
	DEFAULT_CAP_EXCEED  uint32 = 600
	EXTREME_CAP_EXCEED  uint32 = 2000
)

func StatRequirementsHitExpertise_RetWideCap() *StatRequirementsHitExpertise {
	inst := &StatRequirementsHitExpertise{
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + EXTREME_CAP_EXCEED,
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + EXTREME_CAP_EXCEED,
		nil, nil}
	inst.asMap = mapOf(inst)
	return inst
}

func StatRequirementsHitExpertise_ProtFullExpertise() *StatRequirementsHitExpertise {
	inst := &StatRequirementsHitExpertise{
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_EXCEED,
		TARGET_RATING_TANK, TARGET_RATING_TANK + DEFAULT_CAP_EXCEED,
		nil, nil}
	inst.asMap = mapOf(inst)
	return inst
}

func StatRequirementsHitExpertise_ProtFlexibleParry() *StatRequirementsHitExpertise {
	inst := &StatRequirementsHitExpertise{
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_EXCEED,
		TARGET_RATING_MELEE, TARGET_RATING_TANK, // TODO review this when we get falloff of value
		nil, nil}
	inst.asMap = mapOf(inst)
	return inst
}

func StatRequirementsHitExpertise_ProtFlexibleParry_PlusAdditional(additional *stats.StatAndValue) *StatRequirementsHitExpertise {
	inst := &StatRequirementsHitExpertise{
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_EXCEED,
		TARGET_RATING_MELEE, TARGET_RATING_TANK,
		additional, nil}
	inst.asMap = mapOf(inst)
	return inst
}

func StatRequirementsHitExpertise_None() *StatRequirementsHitExpertise {
	return &StatRequirementsHitExpertise{0, math.MaxUint32, 0, math.MaxUint32, nil, nil}
}

func mapOf(inst *StatRequirementsHitExpertise) map[stats.StatType]util_collection.HiLoUInt32 {
	asMap := make(map[stats.StatType]util_collection.HiLoUInt32, 3)
	asMap[stats.Stat_Hit] = util_collection.HiLoUInt32{Lo: inst.hitMin, Hi: inst.hitMax}
	asMap[stats.Stat_Expertise] = util_collection.HiLoUInt32{Lo: inst.expMin, Hi: inst.expMax}
	if inst.AdditionalMinimumRequirement != nil {
		asMap[inst.AdditionalMinimumRequirement.StatType] = util_collection.HiLoUInt32{Lo: inst.AdditionalMinimumRequirement.Value, Hi: math.MaxUint32}
	}
	return asMap
}

func (inst *StatRequirementsHitExpertise) CheckSet(block *stats.StatBlock) (bool, string) {
	hit := block.Hit()
	exp := block.Expertise()
	if inst.hitMin <= hit && hit <= inst.hitMax {
		if inst.expMin <= exp && exp <= inst.expMax {
			if inst.AdditionalMinimumRequirement != nil {
				if block.GetUInt(inst.AdditionalMinimumRequirement.StatType) >= inst.AdditionalMinimumRequirement.Value {
					return true, ""
				} else {
					return false, "additional minimum stat too low"
				}
			} else {
				return true, ""
			}
		} else {
			return false, "expertise out of range"
		}
	} else {
		return false, "hit out of range"
	}
}

func (inst *StatRequirementsHitExpertise) EqualsTyped(other *StatRequirementsHitExpertise) bool {
	if inst.hitMin == other.hitMin && inst.hitMax == other.hitMax && inst.expMin == other.expMin && inst.expMax == other.expMax {
		if inst.AdditionalMinimumRequirement == nil && other.AdditionalMinimumRequirement == nil {
			return true
		} else if inst.AdditionalMinimumRequirement != nil && other.AdditionalMinimumRequirement != nil {
			return *inst.AdditionalMinimumRequirement == *other.AdditionalMinimumRequirement
		}
	}
	return false
}

func (inst *StatRequirementsHitExpertise) Equals(other any) bool {
	if typed, isThisType := other.(*StatRequirementsHitExpertise); isThisType {
		return inst.EqualsTyped(typed)
	} else {
		return false
	}
}

func (inst *StatRequirementsHitExpertise) IsLow(statType stats.StatType, value uint32) bool {
	switch statType {
	case stats.Stat_Hit:
		return value < inst.hitMin
	case stats.Stat_Expertise:
		return value < inst.expMin
	}

	if inst.AdditionalMinimumRequirement != nil && inst.AdditionalMinimumRequirement.StatType == statType {
		return value < inst.AdditionalMinimumRequirement.Value
	}

	return false
}

func (inst *StatRequirementsHitExpertise) IsHigh(statType stats.StatType, value uint32) bool {
	switch statType {
	case stats.Stat_Hit:
		return value > inst.hitMax
	case stats.Stat_Expertise:
		return value > inst.expMax
	}
	return false
}

func (inst *StatRequirementsHitExpertise) GetLow(statType stats.StatType) uint32 {
	switch statType {
	case stats.Stat_Hit:
		return inst.hitMin
	case stats.Stat_Expertise:
		return inst.expMin
	}

	if inst.AdditionalMinimumRequirement != nil && inst.AdditionalMinimumRequirement.StatType == statType {
		return inst.AdditionalMinimumRequirement.Value
	}

	return 0
}

func (inst *StatRequirementsHitExpertise) GetHigh(statType stats.StatType) uint32 {
	switch statType {
	case stats.Stat_Hit:
		return inst.hitMax
	case stats.Stat_Expertise:
		return inst.expMax
	}
	return math.MaxUint32
}

func (inst *StatRequirementsHitExpertise) HitMin() uint32 {
	return inst.hitMin
}

func (inst *StatRequirementsHitExpertise) HitMax() uint32 {
	return inst.hitMax
}

func (inst *StatRequirementsHitExpertise) ExpertMin() uint32 {
	return inst.expMin
}

func (inst *StatRequirementsHitExpertise) ExpertMax() uint32 {
	return inst.expMax
}

func (inst *StatRequirementsHitExpertise) AsMap() map[stats.StatType]util_collection.HiLoUInt32 {
	return inst.asMap
}
