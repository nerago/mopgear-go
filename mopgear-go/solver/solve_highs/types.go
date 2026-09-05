package solve_highs

import (
	"iter"

	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

const (
	c_maxSetItems    = 5 // fundamental in MoP gear sets
	c_setItemsCounts = c_maxSetItems + 1
)

type ISingleGearSet interface {
	setup(model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap) (*columnInfo, error)
	columnsForItemId(id items.ItemId) iter.Seq[*columnInfo]
	buildResultSet(solution util_highs.ISolution, model *solve_highs_types.SolverModel) (items.SolvableItemSet, error)
	checkSetRatingIsObjective(solution *util_highs.Solution2, itemSet *items.SolvableItemSet, calcRating func(item *items.SolvableItemSet) float64, ratingOutputScale float64) error
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

func (colEntry columnInfo) itemId() items.ItemId {
	if colEntry.item != nil {
		return colEntry.item.ItemId()
	} else if colEntry.itemFull != nil {
		return colEntry.itemFull.ItemId()
	} else {
		return 0
	}
}

//type bonusInfo struct {
//	//setTotalCountVar  *columnInfo
//	setIndex          solve_highs_types.SetBonusIndex
//	setExactCountVars [c_setItemsCounts]*columnInfo // specific bools for different counts
//}

type bonusColsByCount [c_setItemsCounts]*columnInfo

type comboEntry struct {
	bonusSetIndex    solve_highs_types.SetBonusIndex
	count            uint32
	exactSetCountVar *columnInfo
}

type bonusCombo struct {
	condition     []comboEntry
	activatingVar *columnInfo
}

func (combo bonusCombo) debugStr() string {
	build := util.StringBuild2{}
	for _, set := range combo.condition {
		build.WriteInt32(int32(set.bonusSetIndex))
		build.WriteRune('=')
		build.WriteInt64(int64(set.count))
		build.WriteRune(' ')
	}
	return build.String()
}
