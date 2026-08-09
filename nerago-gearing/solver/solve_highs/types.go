package solve_highs

import (
	"iter"
	"math"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
)

const (
	c_maxSetItems    = 5 // fundamental in MoP gear sets
	c_setItemsCounts = c_maxSetItems + 1
)

type ISingleGearSet interface {
	ColumnsForItemId(id items.ItemId) iter.Seq[*columnInfo]
	AllColumns() []*columnInfo
	buildResultSet(solution util_highs.ISolution) items.SolvableItemSet
	MainOutputVar() *columnInfo
}

type entryType int8

const (
	entry_item            entryType = iota
	entry_set_total_count entryType = iota
	entry_set_exact_count entryType = iota
	entry_sum_rating      entryType = iota
	entry_combo_active    entryType = iota
	//entry_combo_output_weighted entryType = iota
	entry_main_output           entryType = iota
	entry_multi_enable_forge    entryType = iota
	entry_multi_output          entryType = iota
	entry_stat_total            entryType = iota
	entry_sim_value             entryType = iota
	entry_sim_stat_value        entryType = iota
	entry_sim_stat_value_option entryType = iota
)

type columnInfo struct {
	entryType   entryType
	columnIndex util_highs.ColumnIndex

	statType stats.StatType
	simType  stats.SimType
	itemSlot items.SlotEquip
	item     *items.SolvableItem
	itemFull *items.FullItem

	setIndex   setBonusIndex
	itemCount  int
	combo      *bonusCombo
	multiplier float64
	statRange  weight_types.StatRange
}

func (colEntry columnInfo) ItemId() items.ItemId {
	if colEntry.item != nil {
		return colEntry.item.ItemId()
	} else if colEntry.itemFull != nil {
		return colEntry.itemFull.ItemId()
	} else {
		return 0
	}
}

type setBonusIndex int8
type setBonusRequiredCounts []uint8
type setBonusMultiplierByCount [6]float64

type bonusInfo struct {
	setCountItems  func(*items.SolvableEquipMap) uint8
	setMultipliers setBonusMultiplierByCount
	setIndex       setBonusIndex

	countSetItemsRow  util_highs.ConstraintRow      // use to count items used from this set, has 1 or 0 flags
	setTotalCountVar  *columnInfo                   // total count of items used
	setExactCountVars [c_setItemsCounts]*columnInfo // specific bools for different counts
}

type bonusWithCount struct {
	setInfo *bonusInfo
	count   int
}

type bonusCombo struct {
	condition     []bonusWithCount
	activatingVar *columnInfo
}

func (combo bonusCombo) debugStr() string {
	build := util.StringBuild2{}
	for _, set := range combo.condition {
		build.WriteInt32(int32(set.setInfo.setIndex))
		build.WriteRune('=')
		build.WriteInt64(int64(set.count))
		build.WriteRune(' ')
	}
	return build.String()
}

type SolverModel struct {
	WeightsGeneric weight_types.IWeight
	Weights1       *weight_types.Weight1Basic
	Weights2       *weight_types.Weight2Extended
	Weights3       *weight_types.Weight3ExtendedRanged

	CheckSet       func(itemSet *items.SolvableItemSet) bool
	CalcRatingItem func(item *items.SolvableItem) float64
	CalcRatingSet  func(item *items.SolvableItemSet) float64

	StatRequirements util_collection.EnumMap[stats.StatType, weight_types.StatRangeFloat]

	SetBonusRequiredCounts []setBonusRequiredCounts
	SetBonusTotalCount     int
	SetBonusIndexForItem   func(id items.ItemId) (int, bool)
	SetBonusCountItems     []func(*items.SolvableEquipMap) uint8
	SetBonusMultipliers    []setBonusMultiplierByCount
}

func SolverModelBuild(model *gear_model.SpecModel, weightType weight_types.WeightType) *SolverModel {
	solveModel := &SolverModel{
		CheckSet:               model.CheckSet,
		StatRequirements:       toEnumMap(model.StatRequirements.AsMap()),
		SetBonusRequiredCounts: convertBonusRequired(model.SetBonusRequired, model.SetBonus.BonusSets()),
		SetBonusTotalCount:     len(model.SetBonus.BonusSets()),
		SetBonusIndexForItem:   model.SetBonus.BonusSetIndexForItem,
	}

	weightExt := model.StatWeights
	switch weightType {
	case 1:
		solveModel.Weights1 = &weightExt.Weight1
		solveModel.WeightsGeneric = &weightExt.Weight1
	case 2:
		solveModel.Weights2 = &weightExt.Weight2
		solveModel.WeightsGeneric = &weightExt.Weight2
	case 3:
		solveModel.Weights3 = &weightExt.Weight3
		solveModel.WeightsGeneric = &weightExt.Weight3
	default:
		panic("invalid weight number")
	}

	if solveModel.WeightsGeneric.IsEmpty() {
		panic("requested weight missing")
	}

	solveModel.CalcRatingSet = func(itemSet *items.SolvableItemSet) float64 {
		return model.CalcRatingSolveForGivenWeight(itemSet, solveModel.WeightsGeneric)
	}

	solveModel.CalcRatingItem = func(item *items.SolvableItem) float64 {
		return model.CalcRatingSolveItemForGivenWeight(item, solveModel.WeightsGeneric)
	}

	solveModel.SetBonusCountItems = util_collection.MapSliceAsNew_NoPointer(
		model.SetBonus.BonusSets(),
		func(set gear_model.BonusSet) func(*items.SolvableEquipMap) uint8 {
			return set.CountItems
		},
	)

	solveModel.SetBonusMultipliers = util_collection.MapSliceAsNew_NoPointer(
		model.SetBonus.BonusSets(),
		func(set gear_model.BonusSet) setBonusMultiplierByCount {
			return setBonusMultiplierByCount(set.BonusByCount())
		},
	)

	return solveModel
}

func convertBonusRequired(required []gear_model.BonusSetCountsRequired, activeSets []gear_model.BonusSet) []setBonusRequiredCounts {
	return util_collection.MapSliceAsNew(required, func(setCounts *gear_model.BonusSetCountsRequired) setBonusRequiredCounts {
		countsAsSlice := make(setBonusRequiredCounts, len(activeSets))
		for set, count := range setCounts.Pairs() {
			setIndex := slices.IndexFunc(activeSets, set.Equals)
			if setIndex == -1 {
				panic("set not found")
			}
			countsAsSlice[setIndex] = count
		}
		return countsAsSlice
	})
}

func toEnumMap(goMap map[stats.StatType]util_collection.HiLoUInt32) util_collection.EnumMap[stats.StatType, weight_types.StatRangeFloat] {
	enumMap := util_collection.EnumMapMake[stats.StatType, weight_types.StatRangeFloat](stats.StatTypeEnum)
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
