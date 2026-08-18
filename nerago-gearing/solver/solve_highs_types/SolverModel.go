package solve_highs_types

import (
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/gear_model/bonus_set"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
	"math"
	"slices"
)

type SetBonusIndex int8
type SetBonusRequiredCounts map[SetBonusIndex]uint8

type SolverModel struct {
	WeightsGeneric weight_types.IWeight
	Weights1       *weight_types.Weight1Basic
	Weights2       *weight_types.Weight2Extended
	Weights3       *weight_types.Weight3ExtendedRanged

	CheckSet       func(itemSet *items.SolvableItemSet) (bool, string)
	CalcRatingItem func(item *items.SolvableItem) float64
	CalcRatingSet  func(item *items.SolvableItemSet) float64

	StatRequirements stats.StatTypeMap[weight_types.StatRangeFloat]

	SetBonus SolverModelSetBonus
}

type SolverModelSetBonus struct {
	CountMode        bonus_set.ItemCountsRequiredMode
	RequiredCounts   []SetBonusRequiredCounts
	TotalCount       int
	IndexForItem     func(id items.ItemId) (int, bool)
	CountItems       []func(*items.SolvableEquipMap) uint8
	MultipliersFlat  []bonus_set.BonusByCountFlat
	MultipliersBySim []bonus_set.BonusByCountBySim
}

type OverrideBonusCounts struct {
	Specific bonus_set.ItemCountsRequired
}

func SolverModelBuild(model *gear_model.SpecModel, weightType weight_types.WeightType, bonusCount *OverrideBonusCounts) *SolverModel {
	solveModel := &SolverModel{
		CheckSet:         model.CheckSetForSolver,
		StatRequirements: toEnumMap(model.StatRequirements.AsMap()),
		SetBonus: SolverModelSetBonus{
			CountMode:      model.BonusRequiredSolve.Mode,
			RequiredCounts: convertBonusRequired(model.BonusRequiredSolve, model.BonusEnabled.EnabledSets),
			TotalCount:     len(model.BonusEnabled.EnabledSets),
			IndexForItem:   model.BonusEnabled.BonusSetIndexForItem,
		},
	}

	if bonusCount != nil {
		solveModel.SetBonus.RequiredCounts = []SetBonusRequiredCounts{convertBonusRequiredSingle(&bonusCount.Specific, model.BonusEnabled.EnabledSets)}
	}

	weightExt := model.StatWeights
	switch weightType {
	case 1:
		solveModel.Weights1 = &weightExt.Weight1
		solveModel.WeightsGeneric = &weightExt.Weight1
	case 2:
		solveModel.Weights2 = weightExt.Weight2
		solveModel.WeightsGeneric = weightExt.Weight2
	case 3:
		solveModel.Weights3 = weightExt.Weight3
		solveModel.WeightsGeneric = weightExt.Weight3
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
		solveModel.SetBonus.CountItems = append(solveModel.SetBonus.CountItems, prepBonus.CountItemsSolve)
		solveModel.SetBonus.MultipliersFlat = append(solveModel.SetBonus.MultipliersFlat, prepBonus.BonusByCount())
		solveModel.SetBonus.MultipliersBySim = append(solveModel.SetBonus.MultipliersBySim, prepBonus.BonusByCountBySim())
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
		countsAsSlice[SetBonusIndex(setIndex)] = count
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
