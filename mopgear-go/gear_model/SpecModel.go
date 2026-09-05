package gear_model

import (
	"slices"

	"github.com/nerago/mopgear-go/gear_model/bonus_set"
	"github.com/nerago/mopgear-go/gear_model/ratings"
	. "github.com/nerago/mopgear-go/items"
	. "github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type SpecModel struct {
	Spec SpecType
	ModelItems
	ModelSolve
	ModelSimulate
	ModelBonus
	Initialized bool
}

func (model *SpecModel) Init() {
	model.ModelItems.EnchantChoice = EnchantChoice_ForSpec(model.Spec, model.Goal)
	model.ModelItems.GemChoice = GemChoice_ForSpec(model.Spec, model.Goal)
	model.BonusEnabled.InitFromModel(&model.SimPriority)
}

type ModelItems struct {
	GearFile           string
	ReforgeRules       ReforgeRules
	EnchantChoice      EnchantChoice
	GemChoice          GemChoice
	Professions        ProfessionInfo
	BlockSpecificItems []ItemId
}

type ModelSolve struct {
	WeightFile        string
	StatWeights       ratings.StatRatingsWeightsExtended
	StatRequirements  StatRequirements
	SimPriority       weight_types.SimPriorityBasic
	StatsForWeighting []StatType
}

type ModelSimulate struct {
	Goal       OptimiseGoal
	SimulateAs WowSim_Fight
	SimSpeedUp int
}

type ModelBonus struct {
	BonusEnabled        *bonus_set.SpecSetsEnable
	BonusRequiredSolve  bonus_set.ItemCountsRequiredOptions
	BonusRequiredWeight *bonus_set.ItemCountsRequired
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
		model.BonusRequiredSolve.Equals(other.BonusRequiredSolve) &&
		util.NilSafeEqual(model.BonusRequiredWeight, other.BonusRequiredWeight, bonus_set.ItemCountsRequired.Equals) &&
		model.Professions == other.Professions &&
		model.SimPriority.Equals(&other.SimPriority) &&
		slices.Equal(model.StatsForWeighting, other.StatsForWeighting) &&
		slices.Equal(model.BlockSpecificItems, other.BlockSpecificItems) &&
		model.GearFile == other.GearFile
}

func (model *SpecModel) CloneShallow() *SpecModel {
	return &SpecModel{
		StatRequirements:    model.StatRequirements,
		StatWeights:         model.StatWeights,
		Spec:                model.Spec,
		Goal:                model.Goal,
		SimulateAs:          model.SimulateAs,
		SimSpeedUp:          model.SimSpeedUp,
		ReforgeRules:        model.ReforgeRules,
		EnchantChoice:       model.EnchantChoice,
		GemChoice:           model.GemChoice,
		BonusEnabled:        model.BonusEnabled,
		BonusRequiredSolve:  model.BonusRequiredSolve,
		BonusRequiredWeight: model.BonusRequiredWeight,
		Professions:         model.Professions,
		SimPriority:         model.SimPriority,
		StatsForWeighting:   model.StatsForWeighting,
		BlockSpecificItems:  model.BlockSpecificItems,
		GearFile:            model.GearFile,
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
		return model.BonusRequiredWeight.ItemsMatchRuleFull(itemSet.Items(), bonus_set.CountMode_Exact)
	} else if !model.BonusRequiredSolve.IsEmpty() {
		return model.BonusRequiredSolve.ItemsMatchAnyRuleFull(itemSet.Items())
	}
	return true
}

func (model *SpecModel) ValidateSet(itemSet *FullItemSet) error {
	if err := itemSet.DebugValidate(); err != nil {
		return err
	}
	if err := itemSet.ValidateItemRules(); err != nil {
		return err
	}
	return model.GemChoice.ValidateMetaGemInItemSet(itemSet)
}

// ////////// set ratings
func (model *SpecModel) CalcRatingSolve(itemSet *SolvableItemSet, weightType weight_types.WeightType) float64 {
	weight := model.StatWeights.GetByWeightTypeForSolve(weightType)
	if weightType == 1 {
		baseRating := weight.CalcStatScore(itemSet.Total())
		setRating := model.BonusEnabled.CalcBonusSolveFlat(itemSet.Items())
		return baseRating * setRating
	} else {
		bonusMap := bonus_set.BonusBySim{}
		model.BonusEnabled.CalcBonusSolveBySim(itemSet.Items(), &bonusMap)
		return weight.CalcStatScoreWithBonus(itemSet.Total(), (*SimTypeMap[float64])(&bonusMap))
	}
}

func (model *SpecModel) CalcRatingFull(itemSet *FullItemSet, weightType weight_types.WeightType) float64 {
	weight := model.StatWeights.GetByWeightTypeForSolve(weightType)
	if weightType == 1 {
		baseRating := weight.CalcStatScore(itemSet.Total())
		setRating := model.BonusEnabled.CalcBonusFullFlat(itemSet.Items())
		return baseRating * setRating
	} else {
		bonusMap := bonus_set.BonusBySim{}
		model.BonusEnabled.CalcBonusFullBySim(itemSet.Items(), &bonusMap)
		return weight.CalcStatScoreWithBonus(itemSet.Total(), (*SimTypeMap[float64])(&bonusMap))
	}
}

// ////////// items ratings
func (model *SpecModel) CalcRatingSolveItem(item *SolvableItem, weightType weight_types.WeightType) float64 {
	return model.StatWeights.GetByWeightTypeForSolve(weightType).CalcStatScore(item.Total())
}

func (model *SpecModel) CalcRatingFullItem(item *FullItem, weightType weight_types.WeightType) float64 {
	return model.StatWeights.GetByWeightTypeForSolve(weightType).CalcStatScore(item.Total())
}

// ////////// ProfessionInfo
type ProfessionInfo struct {
	IsBlacksmith bool
	IsEngineer   bool
}
