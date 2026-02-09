package model

import (
	. "paladin_gearing_go/model/ratings"
	. "paladin_gearing_go/model/requirements"
	. "paladin_gearing_go/types/common"
	. "paladin_gearing_go/types/items"
)

type Model struct {
	spec             SpecType
	statRatings      StatRatings
	statRequirements StatRequirements
	reforgeRules     ReforgeRules
	enchantChoice    EnchantChoice
	gemChoice        GemChoice
	// setBonus         SetBonus
}

func (model *Model) CheckSet(itemSet *SolvableItemSet) bool {
	return model.statRequirements.CheckSet(&itemSet.TotalCap)
}

func (model *Model) CalcRating(itemSet *SolvableItemSet) uint64 {
	return model.statRatings.CalcRating(&itemSet.TotalRated)
}

func Model_PallyProtMitigation() Model {
	return Model{
		Spec_PaladinProtMitigation,
		StatRatingsWeights_readFile("C:\\Users\\nicholas\\Dropbox\\prog\\paladin_gearing\\src\\main\\resources\\weight\\PaladinProtMitigation.txt"),
		StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules_tank,
		EnchantChoice_ForSpec(Spec_PaladinProtMitigation),
		GemChoice_ForSpec(Spec_PaladinProtMitigation)}
}
