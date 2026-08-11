package solve_highs

import (
	"iter"
	"paladin_gearing_go/gear_model/bonus_set"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"
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
	entry_item                  entryType = iota
	entry_set_total_count       entryType = iota
	entry_set_exact_count       entryType = iota
	entry_sum_rating            entryType = iota
	entry_combo_active          entryType = iota
	entry_main_output           entryType = iota
	entry_multi_enable_forge    entryType = iota
	entry_multi_output          entryType = iota
	entry_stat_total            entryType = iota
	entry_sim_value             entryType = iota
	entry_sim_stat_value        entryType = iota
	entry_sim_stat_value_option entryType = iota
	entry_sim_value_combo       entryType = iota
)

type columnInfo struct {
	entryType   entryType
	columnIndex util_highs.ColumnIndex

	statType stats.StatType
	simType  stats.SimType
	itemSlot items.SlotEquip
	item     *items.SolvableItem
	itemFull *items.FullItem

	setIndex   solve_highs_types.SetBonusIndex
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

type bonusInfo struct {
	setCountItems       func(*items.SolvableEquipMap) uint8
	setMultipliers      bonus_set.BonusByCountFlat
	setMultipliersBySim bonus_set.BonusByCountBySim
	setIndex            solve_highs_types.SetBonusIndex

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

func (combo bonusCombo) totalFlatMultiplier() float64 {
	bonusMultiplier := 1.0
	for _, setAndCount := range combo.condition {
		bonusForCount := setAndCount.setInfo.setMultipliers[setAndCount.count]
		bonusMultiplier *= bonusForCount
	}
	return bonusMultiplier
}

func (combo bonusCombo) totalMultiplierForSim(simType stats.SimType) float64 {
	bonusMultiplier := 1.0
	for _, setAndCount := range combo.condition {
		simMap := setAndCount.setInfo.setMultipliersBySim[setAndCount.count]
		bonusForCount := simMap.GetOrDefault(simType, 1)
		bonusMultiplier *= bonusForCount
	}
	return bonusMultiplier
}
