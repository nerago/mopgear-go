package model

import (
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model/ratings"
	. "paladin_gearing_go/model/requirements"
	"paladin_gearing_go/stats"
	. "paladin_gearing_go/stats"
	"slices"
)

type Model struct {
	// interface
	// StatRequirements StatRequirements
	// StatRatings      StatRatings

	// hardcode implementation
	StatRequirements StatRequirementsHitExpertise
	StatRatings      StatRatingsWeights

	Spec                 SpecType
	Goal                 OptimiseGoal
	SimulateAs           WowSim_Fight
	SimSpeedUp           int
	ReforgeRules         ReforgeRules
	EnchantChoice        EnchantChoice
	GemChoice            GemChoice
	SetBonus             SetBonus
	SetBonusRequired     []ActiveSetCountsRequired
	FixedWeightsSetBonus *ActiveSetCountsRequired
	Professions          ProfessionInfo
	SimRatioWeighting    stats.SimData
	StatsForWeighting    []StatType
	ReferenceGearFile    string // should just be used by exporters etc
}

func (model *Model) Equals(other *Model) bool {
	return model.StatRequirements.Equals(&other.StatRequirements) &&
		model.StatRatings == other.StatRatings &&
		model.Spec == other.Spec &&
		model.ReforgeRules.Equals(&other.ReforgeRules) &&
		model.EnchantChoice.Equals(other.EnchantChoice) &&
		model.GemChoice.Equals(other.GemChoice) &&
		model.SetBonus.Equals(&other.SetBonus) &&
		slices.EqualFunc(model.SetBonusRequired, other.SetBonusRequired, ActiveSetCountsRequired.Equals) &&
		model.Professions == other.Professions
}

// ////////// requirements
func (model *Model) CheckSet(itemSet *SolvableItemSet) bool {
	if !model.StatRequirements.CheckSet(itemSet.Total()) {
		return false
	}
	if len(model.SetBonusRequired) > 0 {
		return ActiveSetCountsMeetAny(model.SetBonusRequired, itemSet.Items())
	}
	return true
}

func (model *Model) CheckSetFull(itemSet *FullItemSet) bool {
	if !model.StatRequirements.CheckSet(itemSet.Total()) {
		return false
	}
	if len(model.SetBonusRequired) > 0 {
		return ActiveSetCountsMeetAny_FullItem(model.SetBonusRequired, itemSet.Items())
	}
	return true
}

func (model *Model) CheckSetFull_ForWeightProcess(itemSet *FullItemSet) bool {
	if !model.StatRequirements.CheckSet(itemSet.Total()) {
		return false
	}
	if model.FixedWeightsSetBonus != nil {
		return ActiveSetCountsMeetExact_FullItem(*model.FixedWeightsSetBonus, itemSet.Items())
	} else if len(model.SetBonusRequired) > 0 {
		return ActiveSetCountsMeetAny_FullItem(model.SetBonusRequired, itemSet.Items())
	}
	return true
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
