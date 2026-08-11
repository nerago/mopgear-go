package solve_highs_types

import (
	"math"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/gear_model/bonus_set"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
)

type SetBonusIndex int8
type SetBonusRequiredCounts []uint8

type SolverModel struct {
	WeightsGeneric weight_types.IWeight
	Weights1       *weight_types.Weight1Basic
	Weights2       *weight_types.Weight2Extended
	Weights3       *weight_types.Weight3ExtendedRanged

	CheckSet       func(itemSet *items.SolvableItemSet) (bool, string)
	CalcRatingItem func(item *items.SolvableItem) float64
	CalcRatingSet  func(item *items.SolvableItemSet) float64

	StatRequirements stats.StatTypeMap[weight_types.StatRangeFloat]

	SetBonusCountMode        bonus_set.ItemCountsRequiredMode
	SetBonusRequiredCounts   []SetBonusRequiredCounts
	SetBonusTotalCount       int
	SetBonusIndexForItem     func(id items.ItemId) (int, bool)
	SetBonusCountItems       []func(*items.SolvableEquipMap) uint8
	SetBonusMultipliersFlat  []bonus_set.BonusByCountFlat
	SetBonusMultipliersBySim []bonus_set.BonusByCountBySim
	SetBonusExtendedUseSim   bool
}

type OverrideBonusCounts struct {
	Specific bonus_set.ItemCountsRequired
}

func SolverModelBuild(model *gear_model.SpecModel, weightType weight_types.WeightType, bonusCount *OverrideBonusCounts) *SolverModel {
	solveModel := &SolverModel{
		CheckSet:               model.CheckSetForSolver,
		StatRequirements:       toEnumMap(model.StatRequirements.AsMap()),
		SetBonusCountMode:      model.BonusRequiredSolve.Mode,
		SetBonusRequiredCounts: convertBonusRequired(model.BonusRequiredSolve, model.BonusEnabled.EnabledSets),
		SetBonusTotalCount:     len(model.BonusEnabled.EnabledSets),
		SetBonusIndexForItem:   model.BonusEnabled.BonusSetIndexForItem,
	}

	if bonusCount != nil {
		solveModel.SetBonusRequiredCounts = []SetBonusRequiredCounts{convertBonusRequiredSingle(&bonusCount.Specific, model.BonusEnabled.EnabledSets)}
	}

	weightExt := model.StatWeights
	switch weightType {
	case 1:
		solveModel.Weights1 = &weightExt.Weight1
		solveModel.WeightsGeneric = &weightExt.Weight1
	case 2:
		solveModel.Weights2 = weightExt.Weight2
		solveModel.WeightsGeneric = weightExt.Weight2
		solveModel.SetBonusExtendedUseSim = true
	case 3:
		solveModel.Weights3 = weightExt.Weight3
		solveModel.WeightsGeneric = weightExt.Weight3
		solveModel.SetBonusExtendedUseSim = true
	default:
		panic("invalid weight number")
	}

	if solveModel.WeightsGeneric.IsEmpty() {
		panic("requested weight missing")
	}

	solveModel.CalcRatingSet = func(itemSet *items.SolvableItemSet) float64 {
		return model.CalcRatingSolve(itemSet, weightType)
	}

	solveModel.CalcRatingItem = func(item *items.SolvableItem) float64 {
		return model.CalcRatingSolveItem(item, weightType)
	}

	for _, prepBonus := range model.BonusEnabled.EnabledSets {
		solveModel.SetBonusCountItems = append(solveModel.SetBonusCountItems, prepBonus.CountItemsSolve)
		solveModel.SetBonusMultipliersFlat = append(solveModel.SetBonusMultipliersFlat, prepBonus.BonusByCount())
		solveModel.SetBonusMultipliersBySim = append(solveModel.SetBonusMultipliersBySim, prepBonus.BonusByCountBySim())
	}

	return solveModel
}

func convertBonusRequired(required bonus_set.ItemCountsRequiredOptions, enabledSets []bonus_set.PreparedBonus) []SetBonusRequiredCounts {
	return util_collection.MapSliceAsNew(required.Options, func(setCounts *bonus_set.ItemCountsRequired) SetBonusRequiredCounts {
		return convertBonusRequiredSingle(setCounts, enabledSets)
	})
}

func convertBonusRequiredSingle(setCounts *bonus_set.ItemCountsRequired, enabledSets []bonus_set.PreparedBonus) SetBonusRequiredCounts {
	countsAsSlice := make(SetBonusRequiredCounts, len(enabledSets))
	for set, count := range setCounts.Pairs() {
		setIndex := slices.IndexFunc(enabledSets, func(bonus bonus_set.PreparedBonus) bool {
			return bonus.Name() == set.Name()
		})
		if setIndex == -1 {
			panic("set not found")
		}
		countsAsSlice[setIndex] = count
	}
	return countsAsSlice
}

func toEnumMap(goMap map[stats.StatType]util_collection.HiLoUInt32) stats.StatTypeMap[weight_types.StatRangeFloat] {
	enumMap := stats.StatTypeMap[weight_types.StatRangeFloat]{}
	for stat, hiLo := range goMap {
		enumMap.Put(stat, weight_types.StatRangeFloat{Minimum: float64(hiLo.Lo), Maximum: convertHigh(hiLo.Hi)})
	}
	return enumMap
}

func convertHigh(high uint32) float64 {
	if high == math.MaxUint32 {
		return util_highs.InfPos()
	} else {
		return float64(high)
	}
}
