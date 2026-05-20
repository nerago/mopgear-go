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

	Spec             SpecType
	SimulateAs       WowSim_Fight
	ReforgeRules     ReforgeRules
	EnchantChoice    EnchantChoice
	GemChoice        GemChoice
	SetBonus         SetBonus
	SetBonusRequired uint8
	Professions      ProfessionInfo
}

func (model *Model) Equals(other *Model) bool {
	return model.StatRequirements.Equals(&other.StatRequirements) &&
		model.StatRatings == other.StatRatings &&
		model.Spec == other.Spec &&
		model.ReforgeRules.Equals(&other.ReforgeRules) &&
		model.EnchantChoice.Equals(other.EnchantChoice) &&
		model.GemChoice.Equals(other.GemChoice) &&
		model.SetBonus.Equals(&other.SetBonus) &&
		model.SetBonusRequired == other.SetBonusRequired &&
		model.Professions == other.Professions
}

// ////////// requirements
func (model *Model) CheckSet(itemSet *SolvableItemSet) bool {
	if model.StatRequirements.CheckSet(itemSet.Total()) {
		if model.SetBonusRequired > 0 {
			if len(model.SetBonus.activeSets) != 1 {
				panic("set bonus required only available for single set")
			}
			count := model.SetBonus.CountInAnySetSolve(itemSet.Items())
			return count >= model.SetBonusRequired
		} else {
			return true
		}
	} else {
		return false
	}
}

func (model *Model) CheckSetFull(itemSet *FullItemSet) bool {
	if model.StatRequirements.CheckSet(itemSet.Total()) {
		if model.SetBonusRequired > 0 {
			if len(model.SetBonus.activeSets) != 1 {
				panic("set bonus required only available for single set")
			}
			count := model.SetBonus.CountInAnySet(itemSet.Items())
			return count >= model.SetBonusRequired
		} else {
			return true
		}
	} else {
		return false
	}
}

// ////////// set ratings
func (model *Model) CalcRatingSolve(itemSet *SolvableItemSet) float64 {
	baseRating := model.StatRatings.CalcRating(itemSet.Total())
	setRating := model.SetBonus.CalcBonusSolve(itemSet.Items())
	return baseRating * setRating
}

func (model *Model) CalcRatingFull(itemSet *FullItemSet) float64 {
	baseRating := model.StatRatings.CalcRating(itemSet.Total())
	setRating := model.SetBonus.CalcBonusFull(itemSet.Items())
	return baseRating * setRating
}

// ////////// items ratings
func (model *Model) CalcRatingSolveItem(item *SolvableItem) float64 {
	return model.StatRatings.CalcRating(item.Total())
}

func (model *Model) CalcRatingFullItem(item *FullItem) float64 {
	return model.StatRatings.CalcRating(item.Total())
}

// ////////// ProfessionInfo
type ProfessionInfo struct {
	IsBlacksmith bool
	IsEngineer   bool
}
