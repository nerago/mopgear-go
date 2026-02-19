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

func (model *Model) CheckSetSkinny(itemSet *SkinnyItemSet) bool {
	return model.StatRequirements.CheckSetSkinny(itemSet)
}

func (model *Model) CalcRatingSolve(itemSet *SolvableItemSet) uint64 {
	return model.StatRatings.CalcRating(&itemSet.TotalRated)
}

func (model *Model) CalcRatingFull(itemSet *FullItemSet) uint64 {
	return model.StatRatings.CalcRating(&itemSet.TotalRated)
}

func (model *Model) CalcRatingFullItem(item *FullItem) uint64 {
	return model.StatRatings.CalcRating(&item.TotalRated)
}

func (model *Model) CalcRatingSolveItem(item *SolvableItem) uint64 {
	return model.StatRatings.CalcRating(&item.TotalRated)
}

const weightMitiFile = "C:\\Users\\nicholas\\Dropbox\\prog\\paladin_gearing\\src\\main\\resources\\weight\\PaladinProtMitigation.txt"
const weightDpsFile = "C:\\Users\\nicholas\\Dropbox\\prog\\paladin_gearing\\src\\main\\resources\\weight\\PaladinProtDps.txt"

func Model_PallyProtMitigation() Model {
	weightMiti := StatRatingsWeights_ReadFile(weightMitiFile, false, true, false)
	weightDps := StatRatingsWeights_ReadFile(weightDpsFile, false, true, false)
	weight := StatRatingsWeights_Mix(weightMiti, 121, weightDps, 41)
	return Model{
		Spec_PaladinProtMitigation,
		weight,
		StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules_tank,
		EnchantChoice_ForSpec(Spec_PaladinProtMitigation),
		GemChoice_ForSpec(Spec_PaladinProtMitigation)}
}

func Model_PallyProtDps() Model {
	weightMiti := StatRatingsWeights_ReadFile(weightMitiFile, false, true, false)
	weightDps := StatRatingsWeights_ReadFile(weightDpsFile, false, true, false)
	weight := StatRatingsWeights_Mix(weightMiti, 15, weightDps, 185)
	return Model{
		Spec_PaladinProtDps,
		weight,
		StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules_tank,
		EnchantChoice_ForSpec(Spec_PaladinProtDps),
		GemChoice_ForSpec(Spec_PaladinProtDps)}
}
