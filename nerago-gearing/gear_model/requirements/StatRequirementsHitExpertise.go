package requirements

import (
	"math"
	"paladin_gearing_go/stats"
	. "paladin_gearing_go/stats"
)

type StatRequirementsHitExpertise struct {
	hitMin, hitMax uint32
	expMin, expMax uint32

	AdditionalMinimumRequirement *stats.StatAndValue
}

const (
	// RATING_PER_PERCENT    float64 = 339.9534
	// TARGET_PERCENT_MELEE  float64 = 7.5
	TARGET_RATING_MELEE      uint32 = 2550
	TARGET_RATING_TANK       uint32 = 5100
	TARGET_RATING_CAST       uint32 = 5100
	DEFAULT_CAP_ALLOW_EXCEED uint32 = 400
)

func StatRequirementsHitExpertise_RetWideCap() StatRequirementsHitExpertise {
	return StatRequirementsHitExpertise{
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_ALLOW_EXCEED*5,
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_ALLOW_EXCEED*5, nil}
}

func StatRequirementsHitExpertise_ProtFullExpertise() StatRequirementsHitExpertise {
	return StatRequirementsHitExpertise{
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_ALLOW_EXCEED,
		TARGET_RATING_TANK, TARGET_RATING_TANK + DEFAULT_CAP_ALLOW_EXCEED, nil}
}

func StatRequirementsHitExpertise_ProtFlexibleParry() StatRequirementsHitExpertise {
	return StatRequirementsHitExpertise{
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_ALLOW_EXCEED*4,
		TARGET_RATING_MELEE, TARGET_RATING_TANK, nil}
}

func StatRequirementsHitExpertise_ProtFlexibleParry_PlusAdditional(additional *stats.StatAndValue) StatRequirementsHitExpertise {
	return StatRequirementsHitExpertise{
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_ALLOW_EXCEED*4,
		TARGET_RATING_MELEE, TARGET_RATING_TANK, additional}
}

func StatRequirementsHitExpertise_None() StatRequirementsHitExpertise {
	return StatRequirementsHitExpertise{0, math.MaxUint32, 0, math.MaxUint32, nil}
}

func (inst *StatRequirementsHitExpertise) CheckSet(block *StatBlock) bool {
	hit := block.Hit()
	exp := block.Expertise()
	if inst.hitMin <= hit && hit <= inst.hitMax && inst.expMin <= exp && exp <= inst.expMax {
		if inst.AdditionalMinimumRequirement != nil {
			return block.GetUInt(inst.AdditionalMinimumRequirement.StatType) >= inst.AdditionalMinimumRequirement.Value
		} else {
			return true
		}
	} else {
		return false
	}
}

func (inst *StatRequirementsHitExpertise) Equals(other *StatRequirementsHitExpertise) bool {
	if inst.hitMin == other.hitMin && inst.hitMax == other.hitMax && inst.expMin == other.expMin && inst.expMax == other.expMax {
		if inst.AdditionalMinimumRequirement == nil && other.AdditionalMinimumRequirement == nil {
			return true
		} else if inst.AdditionalMinimumRequirement != nil && other.AdditionalMinimumRequirement != nil {
			return *inst.AdditionalMinimumRequirement == *other.AdditionalMinimumRequirement
		}
	}
	return false
}

func (inst *StatRequirementsHitExpertise) IsLow(stat StatType, value uint32) bool {
	switch stat {
	case Stat_Hit:
		return value < inst.hitMin
	case Stat_Expertise:
		return value < inst.expMin
	default:
		return false
	}
}

func (inst *StatRequirementsHitExpertise) IsHigh(stat StatType, value uint32) bool {
	switch stat {
	case Stat_Hit:
		return value > inst.hitMax
	case Stat_Expertise:
		return value > inst.expMax
	default:
		return false
	}
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
