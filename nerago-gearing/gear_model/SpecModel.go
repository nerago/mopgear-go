package gear_model

import (
	"paladin_gearing_go/gear_model/ratings"
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
)

type SpecModel struct {
	StatRequirements         StatRequirements
	StatWeights              ratings.StatRatingsWeightsExtended
	Spec                     SpecType
	Goal                     OptimiseGoal
	SimulateAs               WowSim_Fight
	SimSpeedUp               int
	ReforgeRules             ReforgeRules
	EnchantChoice            EnchantChoice
	GemChoice                GemChoice
	SetBonus                 SetBonus
	SetBonusRequired         []ActiveSetCountsRequired
	FixedWeightsSetBonus     *ActiveSetCountsRequired
	Professions              ProfessionInfo
	SimPriority              weight_types.SimPriorityBasic
	StatsForWeighting        []StatType
	SpecificIncompatibleList []ItemId
	ReferenceGearFile        string // should just be used by exporters etc
}

func (model *SpecModel) Equals(other *SpecModel) bool {
	return model.Spec == other.Spec &&
		model.Goal == other.Goal &&
		model.StatRequirements.Equals(other.StatRequirements) &&
		model.StatWeights.Equals(&other.StatWeights) &&
		model.SimulateAs == other.SimulateAs &&
		model.SimSpeedUp == other.SimSpeedUp &&
		model.ReforgeRules.Equals(&other.ReforgeRules) &&
		model.EnchantChoice.Equals(other.EnchantChoice) &&
		model.GemChoice.Equals(other.GemChoice) &&
		model.SetBonus.Equals(&other.SetBonus) &&
		slices.EqualFunc(model.SetBonusRequired, other.SetBonusRequired, ActiveSetCountsRequired.Equals) &&
		util.NilSafeEqual(model.FixedWeightsSetBonus, other.FixedWeightsSetBonus, ActiveSetCountsRequired.Equals) &&
		model.Professions == other.Professions &&
		model.SimPriority.Equals(&other.SimPriority) &&
		slices.Equal(model.StatsForWeighting, other.StatsForWeighting) &&
		slices.Equal(model.SpecificIncompatibleList, other.SpecificIncompatibleList) &&
		model.ReferenceGearFile == other.ReferenceGearFile
}

func (model *SpecModel) CloneShallow(other *SpecModel) *SpecModel {
	return &SpecModel{
		StatRequirements:         other.StatRequirements,
		StatWeights:              other.StatWeights,
		Spec:                     other.Spec,
		Goal:                     other.Goal,
		SimulateAs:               other.SimulateAs,
		SimSpeedUp:               other.SimSpeedUp,
		ReforgeRules:             other.ReforgeRules,
		EnchantChoice:            other.EnchantChoice,
		GemChoice:                other.GemChoice,
		SetBonus:                 other.SetBonus,
		SetBonusRequired:         other.SetBonusRequired,
		FixedWeightsSetBonus:     other.FixedWeightsSetBonus,
		Professions:              other.Professions,
		SimPriority:              other.SimPriority,
		StatsForWeighting:        other.StatsForWeighting,
		SpecificIncompatibleList: other.SpecificIncompatibleList,
		ReferenceGearFile:        other.ReferenceGearFile,
	}
}

// ////////// requirements
func (model *SpecModel) CheckSet(itemSet *SolvableItemSet) bool {
	//TODO CheckSetWithFailMessage

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

func (model *SpecModel) ValidateSet(itemSet *FullItemSet) {
	itemSet.DebugValidate()
	itemSet.ValidateItemRules()
	model.GemChoice.ValidateMetaGemInItemSet(itemSet)
}

// ////////// set ratings
func (model *SpecModel) CalcRatingSolve(itemSet *SolvableItemSet, weightType weight_types.WeightType) float64 {
	baseRating := model.StatWeights.GetByWeightType(weightType).CalcStatScore(itemSet.Total())
	setRating := model.SetBonus.CalcBonusSolve(itemSet.Items())
	return baseRating * setRating
}

func (model *SpecModel) CalcRatingFull(itemSet *FullItemSet, weightType weight_types.WeightType) float64 {
	baseRating := model.StatWeights.GetByWeightType(weightType).CalcStatScore(itemSet.Total())
	setRating := model.SetBonus.CalcBonusFull(itemSet.Items())
	return baseRating * setRating
}

func (model *SpecModel) CalcRatingSolveForGivenWeight(itemSet *SolvableItemSet, weight weight_types.IWeight) float64 {
	baseRating := weight.CalcStatScore(itemSet.Total())
	setRating := model.SetBonus.CalcBonusSolve(itemSet.Items())
	return baseRating * setRating
}

// ////////// items ratings
func (model *SpecModel) CalcRatingSolveItem(item *SolvableItem, weightType weight_types.WeightType) float64 {
	return model.StatWeights.GetByWeightType(weightType).CalcStatScore(item.Total())
}

func (model *SpecModel) CalcRatingFullItem(item *FullItem, weightType weight_types.WeightType) float64 {
	return model.StatWeights.GetByWeightType(weightType).CalcStatScore(item.Total())
}

func (model *SpecModel) CalcRatingSolveItemForGivenWeight(item *SolvableItem, weight weight_types.IWeight) float64 {
	return weight.CalcStatScore(item.Total())
}

// ////////// ProfessionInfo
type ProfessionInfo struct {
	IsBlacksmith bool
	IsEngineer   bool
}
