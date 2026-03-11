package model

import (
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model/ratings"
	. "paladin_gearing_go/model/requirements"
	. "paladin_gearing_go/stats"
)

type Model struct {
	// StatRequirements StatRequirements
	// StatRatings      StatRatings
	StatRequirements StatRequirementsHitExpertise
	StatRatings      StatRatingsWeights

	Spec SpecType

	ReforgeRules  ReforgeRules
	EnchantChoice EnchantChoice
	GemChoice     GemChoice
	SetBonus      SetBonus
}

func (model *Model) CheckSet(itemSet *SolvableItemSet) bool {
	return model.StatRequirements.CheckSet(itemSet.TotalCap())
}

func (model *Model) CheckSetSkinny(itemSet *SkinnyItemSet) bool {
	return model.StatRequirements.CheckSetSkinny(itemSet)
}

func (model *Model) CalcRatingSolve(itemSet *SolvableItemSet) uint64 {
	rating := model.StatRatings.CalcRating(itemSet.TotalRated())
	rating = model.SetBonus.CalcAndMultiplySolve(&itemSet.Items, rating)
	return rating
}

func (model *Model) CalcRatingFull(itemSet *FullItemSet) uint64 {
	rating := model.StatRatings.CalcRating(itemSet.TotalRated())
	rating = model.SetBonus.CalcAndMultiply(&itemSet.Items, rating)
	return rating
}

func (model *Model) CalcRatingFullItem(item *FullItem) uint64 {
	return model.StatRatings.CalcRating(item.TotalRated())
}

func (model *Model) CalcRatingSolveItem(item *SolvableItem) uint64 {
	return model.StatRatings.CalcRating(item.TotalRated())
}

const weightMitiFile = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\src\main\resources\weight\PaladinProtMitigation.txt`
const weightDpsFile = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\src\main\resources\weight\PaladinProtDps.txt`
const weightRetFile = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\src\main\resources\weight\PaladinRet.txt`

func Model_PallyProtMitigation() Model {
	weightMiti := StatRatingsWeights_ReadFile(weightMitiFile, false, true, false)
	weightDps := StatRatingsWeights_ReadFile(weightDpsFile, false, true, false)
	weight := StatRatingsWeights_Mix(weightMiti, 95, weightDps, 34)
	return Model{
		Spec:             Spec_PaladinProtMitigation,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(Spec_PaladinProtMitigation),
		GemChoice:        GemChoice_ForSpec(Spec_PaladinProtMitigation),
		SetBonus:         SetBonus_Named("Plate of the Lightning Emperor Prot Mitigation")}
}

func Model_PallyProtDps() Model {
	weightMiti := StatRatingsWeights_ReadFile(weightMitiFile, false, true, false)
	weightDps := StatRatingsWeights_ReadFile(weightDpsFile, false, true, false)
	weight := StatRatingsWeights_Mix(weightMiti, 32, weightDps, 146)
	return Model{
		Spec:             Spec_PaladinProtDps,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(Spec_PaladinProtDps),
		GemChoice:        GemChoice_ForSpec(Spec_PaladinProtDps),
		SetBonus:         SetBonus_Named("Plate of the Lightning Emperor Prot Damage")}
}

func Model_PallyRet() Model {
	weight := StatRatingsWeights_ReadFile(weightRetFile, false, false, false)
	return Model{
		Spec:             Spec_PaladinRet,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_RetWideCap(),
		ReforgeRules:     ReforgeRules_melee,
		EnchantChoice:    EnchantChoice_ForSpec(Spec_PaladinRet),
		GemChoice:        GemChoice_ForSpec(Spec_PaladinRet),
		SetBonus:         SetBonus_ForSpec(Spec_PaladinRet)}
}

func Model_Testing() Model {
	weight := StatRatingsWeights_ReadFile(weightDpsFile, false, true, false)
	return Model{
		Spec:             Spec_PaladinProtDps,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_None(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(Spec_PaladinProtDps),
		GemChoice:        GemChoice_ForSpec(Spec_PaladinProtDps),
		SetBonus:         SetBonus_Empty()}
}
