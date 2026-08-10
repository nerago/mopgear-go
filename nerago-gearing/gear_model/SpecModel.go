package gear_model

import (
	"paladin_gearing_go/gear_model/bonus_set"
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
	BonusEnabled             *bonus_set.SpecSetsEnable
	BonusRequiredSolve       bonus_set.ItemCountsRequiredOptions
	BonusRequiredWeight      *bonus_set.ItemCountsRequired
	BonusAvoidNextStep       bool
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
		model.BonusEnabled.Equals(other.BonusEnabled) &&
		slices.EqualFunc(model.BonusRequiredSolve, other.BonusRequiredSolve, bonus_set.ItemCountsRequired.Equals) &&
		util.NilSafeEqual(model.BonusRequiredWeight, other.BonusRequiredWeight, bonus_set.ItemCountsRequired.Equals) &&
		model.Professions == other.Professions &&
		model.SimPriority.Equals(&other.SimPriority) &&
		slices.Equal(model.StatsForWeighting, other.StatsForWeighting) &&
		slices.Equal(model.SpecificIncompatibleList, other.SpecificIncompatibleList) &&
		model.ReferenceGearFile == other.ReferenceGearFile
}

func (model *SpecModel) CloneShallow() *SpecModel {
	return &SpecModel{
		StatRequirements:         model.StatRequirements,
		StatWeights:              model.StatWeights,
		Spec:                     model.Spec,
		Goal:                     model.Goal,
		SimulateAs:               model.SimulateAs,
		SimSpeedUp:               model.SimSpeedUp,
		ReforgeRules:             model.ReforgeRules,
		EnchantChoice:            model.EnchantChoice,
		GemChoice:                model.GemChoice,
		BonusEnabled:             model.BonusEnabled,
		BonusRequiredSolve:       model.BonusRequiredSolve,
		BonusRequiredWeight:      model.BonusRequiredWeight,
		BonusAvoidNextStep:       model.BonusAvoidNextStep,
		Professions:              model.Professions,
		SimPriority:              model.SimPriority,
		StatsForWeighting:        model.StatsForWeighting,
		SpecificIncompatibleList: model.SpecificIncompatibleList,
		ReferenceGearFile:        model.ReferenceGearFile,
	}
}

// ////////// requirements
func (model *SpecModel) CheckSetForSolver(itemSet *SolvableItemSet) (bool, string) {
	if isStatOk, message := model.StatRequirements.CheckSet(itemSet.Total()); !isStatOk {
		return false, message
	}

	if isBonusOk, message := model.BonusRequiredSolve.ItemsMatchAnyRuleSolve(itemSet.Items()); !isBonusOk {
		return false, message
	}

	return true, ""
}

func (model *SpecModel) CheckSetFull_ForWeightProcess(itemSet *FullItemSet) bool {
	isStatOk, _ := model.StatRequirements.CheckSet(itemSet.Total())
	if !isStatOk {
		return false
	}
	if model.BonusRequiredWeight != nil {
		return model.BonusRequiredWeight.ItemsMatchRuleFull(itemSet.Items())
	} else if len(model.BonusRequiredSolve) > 0 {
		return model.BonusRequiredSolve.ItemsMatchAnyRuleFull(itemSet.Items())
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
	weight := model.StatWeights.GetByWeightType(weightType)
	if weightType == 1 {
		baseRating := weight.CalcStatScore(itemSet.Total())
		setRating := model.BonusEnabled.CalcBonusSolveFlat(itemSet.Items(), model.SimPriority)
		return baseRating * setRating
	} else {
		bonusMap := SimTypeMap[float64]{}
		model.BonusEnabled.CalcBonusSolveBySim(itemSet.Items(), &bonusMap)
		return weight.CalcStatScoreWithBonus(itemSet.Total(), &bonusMap)
	}
}

func (model *SpecModel) CalcRatingFull(itemSet *FullItemSet, weightType weight_types.WeightType) float64 {
	weight := model.StatWeights.GetByWeightType(weightType)
	if weightType == 1 {
		baseRating := weight.CalcStatScore(itemSet.Total())
		setRating := model.BonusEnabled.CalcBonusFullFlat(itemSet.Items(), model.SimPriority)
		return baseRating * setRating
	} else {
		bonusMap := SimTypeMap[float64]{}
		model.BonusEnabled.CalcBonusFullBySim(itemSet.Items(), &bonusMap)
		return weight.CalcStatScoreWithBonus(itemSet.Total(), &bonusMap)
	}
}

// ////////// items ratings
func (model *SpecModel) CalcRatingSolveItem(item *SolvableItem, weightType weight_types.WeightType) float64 {
	return model.StatWeights.GetByWeightType(weightType).CalcStatScore(item.Total())
}

func (model *SpecModel) CalcRatingFullItem(item *FullItem, weightType weight_types.WeightType) float64 {
	return model.StatWeights.GetByWeightType(weightType).CalcStatScore(item.Total())
}

// ////////// ProfessionInfo
type ProfessionInfo struct {
	IsBlacksmith bool
	IsEngineer   bool
}
