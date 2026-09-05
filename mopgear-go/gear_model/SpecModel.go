package gear_model

import (
	"slices"
	"strings"

	"github.com/nerago/mopgear-go/gear_model/bonus_set"
	"github.com/nerago/mopgear-go/gear_model/ratings"
	. "github.com/nerago/mopgear-go/items"
	. "github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type SpecModel struct {
	Label string
	Spec  SpecType
	ModelItems
	ModelSolve
	ModelSimulate
	ModelBonus
}

func (model *SpecModel) InitDerives() {
	model.ModelItems.EnchantChoice = EnchantChoice_ForSpec(model.Spec, model.Goal)
	model.ModelItems.GemChoice = GemChoice_ForSpec(model.Spec, model.Goal)
	model.ModelBonus.BonusEnabled.InitFromModel(&model.SimPriority)
}

type ModelItems struct {
	GearFile           string
	SampleDataFile     string
	ReforgeRules       ReforgeRules
	EnchantChoice      EnchantChoice
	GemChoice          GemChoice
	Professions        ProfessionInfo
	BlockSpecificItems []ItemId
}

func (mi *ModelItems) CloneShallow() ModelItems {
	return ModelItems{
		GearFile:           mi.GearFile,
		ReforgeRules:       mi.ReforgeRules,
		EnchantChoice:      mi.EnchantChoice,
		GemChoice:          mi.GemChoice,
		Professions:        mi.Professions,
		BlockSpecificItems: mi.BlockSpecificItems,
	}
}

func (mi *ModelItems) Equals(other *ModelItems) bool {
	return mi.GearFile == other.GearFile &&
		mi.ReforgeRules.Equals(&other.ReforgeRules) &&
		mi.EnchantChoice.Equals(other.EnchantChoice) &&
		mi.GemChoice.Equals(other.GemChoice) &&
		mi.Professions == other.Professions &&
		slices.Equal(mi.BlockSpecificItems, other.BlockSpecificItems)
}

func (mi *ModelItems) GetSampleFileGrid() string {
	return strings.Replace(mi.SampleDataFile, "{}", "grid", -1)
}
func (mi *ModelItems) GetSampleFileFit() string {
	return strings.Replace(mi.SampleDataFile, "{}", "fit", -1)
}
func (mi *ModelItems) GetSampleFileRand() string {
	return strings.Replace(mi.SampleDataFile, "{}", "rand", -1)
}

type ModelSolve struct {
	WeightFile        string
	StatRequirements  StatRequirements
	SimPriority       weight_types.SimPriorityBasic
	StatsForWeighting []StatType

	statWeights *ratings.StatRatingsWeightsExtended
	initialized bool
}

func (ms *ModelSolve) CloneShallow() ModelSolve {
	return ModelSolve{
		WeightFile:        ms.WeightFile,
		StatRequirements:  ms.StatRequirements,
		SimPriority:       ms.SimPriority,
		StatsForWeighting: ms.StatsForWeighting,
		statWeights:       ms.statWeights,
		initialized:       ms.initialized,
	}
}

func (ms *ModelSolve) Equals(other *ModelSolve) bool {
	return ms.WeightFile == other.WeightFile &&
		ms.StatRequirements.Equals(other.StatRequirements) &&
		ms.SimPriority.Equals(&other.SimPriority) &&
		slices.Equal(ms.StatsForWeighting, other.StatsForWeighting) &&
		util.NilSafeEqual(&ms.statWeights, &other.statWeights, (*ratings.StatRatingsWeightsExtended).Equals) &&
		ms.initialized == other.initialized
}

func (ms *ModelSolve) GetStatWeights() *ratings.StatRatingsWeightsExtended {
	if ms.initialized {
		return ms.statWeights
	} else {
		panic("weights not initialized")
	}
}

func (ms *ModelSolve) SetStatWeights(weights *ratings.StatRatingsWeightsExtended) {
	ms.statWeights = weights
	ms.initialized = true
}

type ModelSimulate struct {
	Goal       OptimiseGoal
	SimulateAs WowSim_Fight
	SimSpeedUp int
}

func (mt *ModelSimulate) CloneShallow() ModelSimulate {
	return ModelSimulate{
		Goal:       mt.Goal,
		SimulateAs: mt.SimulateAs,
		SimSpeedUp: mt.SimSpeedUp,
	}
}

func (mt *ModelSimulate) Equals(other *ModelSimulate) bool {
	return mt.Goal == other.Goal &&
		mt.SimulateAs == other.SimulateAs &&
		mt.SimSpeedUp == other.SimSpeedUp
}

type ModelBonus struct {
	BonusEnabled        *bonus_set.SpecSetsEnable
	BonusRequiredSolve  bonus_set.ItemCountsRequiredOptions
	BonusRequiredWeight *bonus_set.ItemCountsRequired
}

func (mb *ModelBonus) CloneShallow() ModelBonus {
	return ModelBonus{
		BonusEnabled:        mb.BonusEnabled,
		BonusRequiredSolve:  mb.BonusRequiredSolve,
		BonusRequiredWeight: mb.BonusRequiredWeight,
	}
}

func (mb *ModelBonus) Equals(other *ModelBonus) bool {
	return mb.BonusEnabled.Equals(other.BonusEnabled) &&
		mb.BonusRequiredSolve.Equals(other.BonusRequiredSolve) &&
		util.NilSafeEqual(mb.BonusRequiredWeight, other.BonusRequiredWeight, bonus_set.ItemCountsRequired.Equals)

}

func (model *SpecModel) Equals(other *SpecModel) bool {
	return model.Label == other.Label &&
		model.Spec == other.Spec &&
		model.ModelItems.Equals(&other.ModelItems) &&
		model.ModelSolve.Equals(&other.ModelSolve) &&
		model.ModelSimulate.Equals(&other.ModelSimulate) &&
		model.ModelBonus.Equals(&other.ModelBonus)
}

func (model *SpecModel) CloneShallow() *SpecModel {
	return &SpecModel{
		Label:         model.Label,
		Spec:          model.Spec,
		ModelItems:    model.ModelItems.CloneShallow(),
		ModelSolve:    model.ModelSolve.CloneShallow(),
		ModelSimulate: model.ModelSimulate.CloneShallow(),
		ModelBonus:    model.ModelBonus.CloneShallow(),
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
	weight := model.ModelSolve.GetStatWeights().GetByWeightTypeForSolve(weightType)
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
	weight := model.ModelSolve.GetStatWeights().GetByWeightTypeForSolve(weightType)
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
	return model.ModelSolve.GetStatWeights().GetByWeightTypeForSolve(weightType).CalcStatScore(item.Total())
}

func (model *SpecModel) CalcRatingFullItem(item *FullItem, weightType weight_types.WeightType) float64 {
	return model.ModelSolve.GetStatWeights().GetByWeightTypeForSolve(weightType).CalcStatScore(item.Total())
}

// ////////// ProfessionInfo
type ProfessionInfo struct {
	IsBlacksmith bool
	IsEngineer   bool
}
