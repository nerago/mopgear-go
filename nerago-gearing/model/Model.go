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

// ////////// set ratings
func (model *Model) CalcRatingSolveAsFloat(itemSet *SolvableItemSet) float64 {
	baseRating := model.StatRatings.CalcRatingFloat(itemSet.TotalRated())
	setRating := model.SetBonus.CalcBonusSolve(itemSet.Items())
	return float64(baseRating * setRating)
}

func (model *Model) CalcRatingFullAsFloat(itemSet *FullItemSet) float64 {
	baseRating := model.StatRatings.CalcRatingFloat(itemSet.TotalRated())
	setRating := model.SetBonus.CalcBonusFull(itemSet.Items())
	return float64(baseRating * setRating)
}

// ////////// items ratings
func (model *Model) CalcRatingSolveItemAsFloat(item *SolvableItem) float64 {
	return float64(model.StatRatings.CalcRatingFloat(item.TotalRated()))
}

func (model *Model) CalcRatingFullItemAsFloat(item *FullItem) float64 {
	return float64(model.StatRatings.CalcRatingFloat(item.TotalRated()))
}

// ////////// ProfessionInfo
type ProfessionInfo struct {
	IsBlacksmith bool
	IsEngineer   bool
}
