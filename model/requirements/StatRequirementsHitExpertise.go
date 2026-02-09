package requirements

import (
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/types/items"
	. "paladin_gearing_go/types/stats"
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

func StatRequirementsHitExpertise_RetWideCap() StatRequirements {
	return StatRequirementsHitExpertise{
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_ALLOW_EXCEED*5,
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_ALLOW_EXCEED*5}
}

func StatRequirementsHitExpertise_ProtFullExpertise() StatRequirements {
	return StatRequirementsHitExpertise{
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_ALLOW_EXCEED,
		TARGET_RATING_TANK, TARGET_RATING_TANK + DEFAULT_CAP_ALLOW_EXCEED}
}

func StatRequirementsHitExpertise_ProtFlexibleParry() StatRequirements {
	return StatRequirementsHitExpertise{
		TARGET_RATING_MELEE, TARGET_RATING_MELEE + DEFAULT_CAP_ALLOW_EXCEED*2, // TODO why so high?
		TARGET_RATING_MELEE, TARGET_RATING_TANK}
}

func (inst StatRequirementsHitExpertise) CheckSet(block *StatBlock) bool {
	hit := block.Hit()
	exp := block.Expertise()
	return inst.hitMin <= hit && hit <= inst.hitMax &&
		inst.expMin <= exp && exp <= inst.expMax
}

func (inst StatRequirementsHitExpertise) ToSkinny(item *SolvableItem) *SkinnyItem {
	return &SkinnyItem{A: item.TotalCap.Hit(), B: item.TotalCap.Expertise(), Exists: true}
}

func (inst StatRequirementsHitExpertise) SkinnyMatch(skinny *SkinnyItem, item *SolvableItem) bool {
	return skinny.A == item.TotalCap.Hit() && skinny.B == item.TotalCap.Expertise()
}
