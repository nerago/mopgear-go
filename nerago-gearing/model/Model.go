package model

import (
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model/ratings"
	. "paladin_gearing_go/model/requirements"
	. "paladin_gearing_go/stats"
)

type Model struct {
	// interface
	// StatRequirements StatRequirements
	// StatRatings      StatRatings

	// hardcode implementation
	StatRequirements StatRequirementsHitExpertise
	StatRatings      StatRatingsWeights

	Spec          SpecType
	SimulateAs    WowSim_Fight
	ReforgeRules  ReforgeRules
	EnchantChoice EnchantChoice
	GemChoice     GemChoice
	SetBonus      SetBonus
	Professions   ProfessionInfo
}

func (model *Model) Equals(other *Model) bool {
	return model.StatRequirements == other.StatRequirements &&
		model.StatRatings == other.StatRatings &&
		model.Spec == other.Spec &&
		model.ReforgeRules.Equals(&other.ReforgeRules) &&
		model.EnchantChoice.Equals(other.EnchantChoice) &&
		model.GemChoice.Equals(other.GemChoice) &&
		model.SetBonus.Equals(&other.SetBonus) &&
		model.Professions == other.Professions
}

// ////////// requirements
func (model *Model) CheckSet(itemSet *SolvableItemSet) bool {
	return model.StatRequirements.CheckSet(itemSet.TotalCap())
}

func (model *Model) CheckSetSkinny(itemSet *SkinnyItemSet) bool {
	return model.StatRequirements.CheckSetSkinny(itemSet)
}

// ////////// set ratings
func (model *Model) CalcRatingSolve(itemSet *SolvableItemSet) uint64 {
	return uint64(model.CalcRatingSolveAsFloat(itemSet))
}

func (model *Model) CalcRatingSolveAsFloat(itemSet *SolvableItemSet) float32 {
	baseRating := model.StatRatings.CalcRatingFloat(itemSet.TotalRated())
	setRating := model.SetBonus.CalcBonusSolve(itemSet.Items())
	return baseRating * setRating
}

func (model *Model) CalcRatingFull(itemSet *FullItemSet) uint64 {
	return uint64(model.CalcRatingFullAsFloat(itemSet))
}

func (model *Model) CalcRatingFullAsFloat(itemSet *FullItemSet) float32 {
	baseRating := model.StatRatings.CalcRatingFloat(itemSet.TotalRated())
	setRating := model.SetBonus.CalcBonusFull(itemSet.Items())
	return baseRating * setRating
}

func (model *Model) CalcRatingGenericSet(itemSet IItemSet) uint64 {
	baseRating := model.StatRatings.CalcRatingFloat(itemSet.TotalRated())
	setRating := model.SetBonus.CalcBonusGeneric(itemSet.ItemsGeneric())
	return uint64(baseRating * setRating)
}

// ////////// items ratings
func (model *Model) CalcRatingFullItem(item *FullItem) uint64 {
	return model.StatRatings.CalcRatingInt(item.TotalRated())
}

func (model *Model) CalcRatingSolveItem(item *SolvableItem) uint64 {
	return model.StatRatings.CalcRatingInt(item.TotalRated())
}

func (model *Model) CalcRatingSolveItemAsFloat(item *SolvableItem) float32 {
	return model.StatRatings.CalcRatingFloat(item.TotalRated())
}

func (model *Model) CalcRatingGenericItem(item IItem) uint64 {
	return model.StatRatings.CalcRatingInt(item.TotalRated())
}

// ////////// ProfessionInfo
type ProfessionInfo struct {
	IsBlacksmith bool
	IsEngineer   bool
}
