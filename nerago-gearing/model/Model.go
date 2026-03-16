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
	setRating := model.SetBonus.CalcBonusSolveUnrollAndSpecial0(itemSet.Items())
	return multiplyRatings(baseRating, setRating)
}

func (model *Model) CalcRatingFull(itemSet *FullItemSet) uint64 {
	baseRating := model.StatRatings.CalcRatingFloat(itemSet.TotalRated())
	setRating := model.SetBonus.CalcBonus(itemSet.Items())
	return multiplyRatings(baseRating, setRating)
}

func (model *Model) CalcRatingGenericSet(itemSet IItemSet) uint64 {
	baseRating := model.StatRatings.CalcRatingFloat(itemSet.TotalRated())
	setRating := model.SetBonus.CalcBonusGenericInterfaceLoopySomeInlined(itemSet.ItemsGeneric())
	return multiplyRatings(baseRating, setRating)
}

func multiplyRatings(baseRating float32, setRating float32) uint64 {
	return uint64(baseRating * setRating)
}

// ////////// items ratings
func (model *Model) CalcRatingFullItem(item *FullItem) uint64 {
	rating := model.StatRatings.CalcRatingInt(item.TotalRated())
	return rating
}

func (model *Model) CalcRatingSolveItem(item *SolvableItem) uint64 {
	rating := model.StatRatings.CalcRatingInt(item.TotalRated())
	return rating
}

func (model *Model) CalcRatingGenericItem(item IItem) uint64 {
	rating := model.StatRatings.CalcRatingInt(item.TotalRated())
	return rating
}
