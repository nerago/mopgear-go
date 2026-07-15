package solve_highs

import (
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_highs"
)

type entryType int8

const (
	entry_item                        entryType = iota
	entry_set_total_count             entryType = iota
	entry_set_exact_count             entryType = iota
	entry_sum_rating                  entryType = iota
	entry_permutation_active          entryType = iota
	entry_permutation_output_weighted entryType = iota
	entry_main_output                 entryType = iota
	entry_multi_enable_forge          entryType = iota
	entry_multi_output                entryType = iota
)

type columnInfo struct {
	entryType   entryType
	columnIndex util_highs.ColumnIndex

	itemSlot items.SlotEquip
	item     *items.SolvableItem
	itemFull *items.FullItem

	set         gear_model.ActiveSet
	itemCount   int
	permutation *setPermutation
	weight      float64
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

type setInfo struct {
	activeSet gear_model.ActiveSet
	setIndex  int

	countSetItemsRow  util_highs.ConstraintRow      // use to count items used from this set, has 1 or 0 flags
	setTotalCountVar  *columnInfo                   // total count of items used
	setExactCountVars [c_setItemsCounts]*columnInfo // specific bools for different counts
}

type setPermutation struct {
	content []setWithCount

	outputVar     *columnInfo
	activatingVar *columnInfo
	weight        float64
}

func (perm setPermutation) debugStr() string {
	build := util.StringBuild2{}
	for _, set := range perm.content {
		build.WriteString(set.setInfo.activeSet.Name())
		build.WriteRune('=')
		build.WriteInt64(int64(set.count))
		build.WriteRune(' ')
	}
	return build.String()
}

type setWithCount struct {
	setInfo setInfo
	count   int
}
