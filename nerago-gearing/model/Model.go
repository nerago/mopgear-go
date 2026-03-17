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
	ReforgeRules  ReforgeRules
	EnchantChoice EnchantChoice
	GemChoice     GemChoice
	SetBonus      SetBonus
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
	baseRating := model.StatRatings.CalcRatingFloat(itemSet.TotalRated())
	setRating := model.SetBonus.CalcBonusSolve(itemSet.Items())
	return uint64(baseRating * setRating)
}

func (model *Model) CalcRatingFull(itemSet *FullItemSet) uint64 {
	baseRating := model.StatRatings.CalcRatingFloat(itemSet.TotalRated())
	setRating := model.SetBonus.CalcBonusFull(itemSet.Items())
	return uint64(baseRating * setRating)
}

func (model *Model) CalcRatingGenericSet(itemSet IItemSet) uint64 {
	baseRating := model.StatRatings.CalcRatingFloat(itemSet.TotalRated())
	setRating := model.SetBonus.CalcBonusGeneric(itemSet.ItemsGeneric())
	return uint64(baseRating * setRating)
}

// combined?
func (model *Model) CheckAndRate(itemSet *SolvableItemSet) uint64 {
	if model.StatRequirements.CheckSet(itemSet.TotalCap()) {
		baseRating := model.StatRatings.CalcRatingFloat(itemSet.TotalRated())
		setRating := model.SetBonus.CalcBonusSolve(itemSet.Items())
		return uint64(baseRating * setRating)
	} else {
		return 0
	}
}

// ////////// items ratings
func (model *Model) CalcRatingFullItem(item *FullItem) uint64 {
	return model.StatRatings.CalcRatingInt(item.TotalRated())
}

func (model *Model) CalcRatingSolveItem(item *SolvableItem) uint64 {
	return model.StatRatings.CalcRatingInt(item.TotalRated())
}

func (model *Model) CalcRatingGenericItem(item IItem) uint64 {
	return model.StatRatings.CalcRatingInt(item.TotalRated())
}
