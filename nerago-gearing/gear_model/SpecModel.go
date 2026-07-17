package gear_model

import (
	. "paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	. "paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
)

type SpecModel struct {
	// interface
	StatRequirements StatRequirements
	StatRatings      StatRatings

	// hardcode implementation
	//StatRequirements StatRequirementsHitExpertise
	//StatRatings      StatRatingsWeights

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

func (model *SpecModel) Equals(other *SpecModel) bool {
	return model.StatRequirements.Equals(other.StatRequirements) &&
		model.StatRatings == other.StatRatings &&
		model.Spec == other.Spec &&
		model.Goal == other.Goal &&
		model.SimulateAs == other.SimulateAs &&
		model.SimSpeedUp == other.SimSpeedUp &&
		model.ReforgeRules.Equals(&other.ReforgeRules) &&
		model.EnchantChoice.Equals(other.EnchantChoice) &&
		model.GemChoice.Equals(other.GemChoice) &&
		model.SetBonus.Equals(&other.SetBonus) &&
		slices.EqualFunc(model.SetBonusRequired, other.SetBonusRequired, ActiveSetCountsRequired.Equals) &&
		util.NilSafeEqual(model.FixedWeightsSetBonus, other.FixedWeightsSetBonus, ActiveSetCountsRequired.Equals) &&
		model.Professions == other.Professions &&
		model.SimRatioWeighting.Equals(&other.SimRatioWeighting) &&
		slices.Equal(model.StatsForWeighting, other.StatsForWeighting) &&
		model.ReferenceGearFile == other.ReferenceGearFile
}

// ////////// requirements
func (model *SpecModel) CheckSet(itemSet *SolvableItemSet) bool {
	if !model.StatRequirements.CheckSet(itemSet.Total()) {
		return false
	}
	if len(model.SetBonusRequired) > 0 {
		return ActiveSetCountsMeetAny(model.SetBonusRequired, itemSet.Items())
	}
	return true
}

func (model *SpecModel) CheckSetFull(itemSet *FullItemSet) bool {
	if !model.StatRequirements.CheckSet(itemSet.Total()) {
		return false
	}
	if len(model.SetBonusRequired) > 0 {
		return ActiveSetCountsMeetAny_FullItem(model.SetBonusRequired, itemSet.Items())
	}
	return true
}

func (model *SpecModel) CheckSetFull_ForWeightProcess(itemSet *FullItemSet) bool {
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
func (model *SpecModel) CalcRatingSolve(itemSet *SolvableItemSet) float64 {
	baseRating := model.StatRatings.CalcRating(itemSet.Total())
	setRating := model.SetBonus.CalcBonusSolve(itemSet.Items())
	return baseRating * setRating
}

func (model *SpecModel) CalcRatingFull(itemSet *FullItemSet) float64 {
	baseRating := model.StatRatings.CalcRating(itemSet.Total())
	setRating := model.SetBonus.CalcBonusFull(itemSet.Items())
	return baseRating * setRating
}

// ////////// items ratings
func (model *SpecModel) CalcRatingSolveItem(item *SolvableItem) float64 {
	return model.StatRatings.CalcRating(item.Total())
}

func (model *SpecModel) CalcRatingFullItem(item *FullItem) float64 {
	return model.StatRatings.CalcRating(item.Total())
}

// ////////// ProfessionInfo
type ProfessionInfo struct {
	IsBlacksmith bool
	IsEngineer   bool
}
