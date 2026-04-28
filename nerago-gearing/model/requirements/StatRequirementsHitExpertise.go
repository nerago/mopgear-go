package requirements

import (
	"math"
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/stats"
)

type StatRequirementsHitExpertise struct {
	hitMin, hitMax uint32
	expMin, expMax uint32
	// minimiseExpertise bool
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
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_ALLOW_EXCEED*5}
}

func StatRequirementsHitExpertise_ProtFullExpertise() StatRequirementsHitExpertise {
	return StatRequirementsHitExpertise{
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_ALLOW_EXCEED,
		TARGET_RATING_TANK, TARGET_RATING_TANK + DEFAULT_CAP_ALLOW_EXCEED}
}

func StatRequirementsHitExpertise_ProtFlexibleParry() StatRequirementsHitExpertise {
	return StatRequirementsHitExpertise{
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_ALLOW_EXCEED*4,
		TARGET_RATING_MELEE, TARGET_RATING_TANK}
}

func StatRequirementsHitExpertise_None() StatRequirementsHitExpertise {
	return StatRequirementsHitExpertise{0, math.MaxUint32, 0, math.MaxUint32}
}

func (inst StatRequirementsHitExpertise) CheckSet(block *StatBlock) bool {
	hit := block.Hit()
	exp := block.Expertise()
	return inst.hitMin <= hit && hit <= inst.hitMax && inst.expMin <= exp && exp <= inst.expMax
}

func (inst StatRequirementsHitExpertise) CheckSetSkinny(set *SkinnyItemSet) bool {
	hit := set.A
	exp := set.B
	return inst.hitMin <= hit && hit <= inst.hitMax && inst.expMin <= exp && exp <= inst.expMax
}

func (inst StatRequirementsHitExpertise) ToSkinny(item *SolvableItem) SkinnyItem {
	return SkinnyItem{A: item.TotalCap().Hit(), B: item.TotalCap().Expertise(), Exists: true}
}

func (inst StatRequirementsHitExpertise) SkinnyMatch(skinny *SkinnyItem, item *SolvableItem) bool {
	return skinny.A == item.TotalCap().Hit() && skinny.B == item.TotalCap().Expertise()
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
