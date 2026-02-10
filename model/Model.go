package model

import (
	. "paladin_gearing_go/model/ratings"
	. "paladin_gearing_go/model/requirements"
	. "paladin_gearing_go/types/common"
	. "paladin_gearing_go/types/items"
)

type Model struct {
	Spec             SpecType
	StatRatings      StatRatings
	StatRequirements StatRequirements
	ReforgeRules     ReforgeRules
	EnchantChoice    EnchantChoice
	GemChoice        GemChoice
	// setBonus         SetBonus
}

func (model *Model) CheckSet(itemSet *SolvableItemSet) bool {
	return model.StatRequirements.CheckSet(&itemSet.TotalCap)
}

func (model *Model) CalcRatingSolve(itemSet *SolvableItemSet) uint64 {
	return model.StatRatings.CalcRating(&itemSet.TotalRated)
}

func (model *Model) CalcRatingFull(itemSet *FullItemSet) uint64 {
	return model.StatRatings.CalcRating(&itemSet.TotalRated)
}

func Model_PallyProtMitigation() Model {
	return Model{
		Spec_PaladinProtMitigation,
		StatRatingsWeights_ReadFile("C:\\Users\\nicholas\\Dropbox\\prog\\paladin_gearing\\src\\main\\resources\\weight\\PaladinProtMitigation.txt",
			false, true, false),
		StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules_tank,
		EnchantChoice_ForSpec(Spec_PaladinProtMitigation),
		GemChoice_ForSpec(Spec_PaladinProtMitigation)}
}
