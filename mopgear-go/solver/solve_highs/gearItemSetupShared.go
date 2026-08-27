package solve_highs

import (
	"errors"
	"strconv"

	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"

	"github.com/bartolsthoorn/gohighs/highs"
)

type gearItemSetupShared struct {
	itemColumns      util_collection.MapSlice[items.ItemId, *columnInfo]
	createItemColumn func(*columnInfo)

	slotsOneEachRow [items.ITEM_SLOT_COUNT]util_highs.ConstraintRow // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot

	uniqueEquipRowsById util_collection.Map[items.ItemId, *util_highs.ConstraintRow] // lookup by id, may have multiple mappings for an item so need pointers
	uniqueEquipRowsAll  []*util_highs.ConstraintRow                                  // definitive copy of each unique equip row constraint

	countSetItemsRow []util_highs.ConstraintRow // use to count items used from this set, has 1 or 0 flags
	activeSetIndex   func(id items.ItemId) (int, bool)
}

func (sit *gearItemSetupShared) prepare(model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap, createItemColumn func(*columnInfo)) error {
	err := sit.prepareUniqueEquipped(itemOptions)
	if err != nil {
		return err
	}
	sit.countSetItemsRow = make([]util_highs.ConstraintRow, model.SetBonus.TotalCount)
	sit.activeSetIndex = model.SetBonus.IndexForItem
	sit.createItemColumn = createItemColumn
	return nil
}

func (sit *gearItemSetupShared) prepareUniqueEquipped(itemOptions *items.SolvableOptionsMap) error {
	sit.uniqueEquipRowsAll = make([]*util_highs.ConstraintRow, 0)
	seen := make(map[items.ItemId]bool)

	// add items from predefined unique equipped sets
	for _, set := range itemOptions.UniqueEquippedSets() {
		row := new(util_highs.ConstraintRow)
		row.Debug = "uniqueEquipped" + set[0].String()
		sit.uniqueEquipRowsAll = append(sit.uniqueEquipRowsAll, row)

		for _, itemId := range set {
			if seen[itemId] {
				return errors.New("unique equipped data has duplicate")
			}
			sit.uniqueEquipRowsById.Put(itemId, row)
			seen[itemId] = true
		}
	}

	return nil
}

func (sit *gearItemSetupShared) addItemCommon(itemSlot items.SlotEquip, item *items.SolvableItem) util_highs.ColumnIndex {
	entry := &columnInfo{entryType: entry_item, itemSlot: itemSlot, item: item}

	// boolean value to flag use of specific item, in exact reforge/gem state
	sit.createItemColumn(entry)
	columnIndex := entry.columnIndex
	sit.itemColumns.Add(item.ItemId(), entry)

	// 1 for that slot that matches the item, so we can tell solver only one item per slot
	sit.slotsOneEachRow[itemSlot].Add(columnIndex, 1.0)

	// if this item belongs to any item set then flag with a 1
	activeSetIndex, hasSet := sit.activeSetIndex(item.ItemId())
	if hasSet {
		sit.countSetItemsRow[activeSetIndex].Add(columnIndex, 1)
	}

	// if this item is unique equipped (mostly checked for ring/trinket)
	uniqueRow := sit.uniqueEquipRowsById.GetOrNilValue(item.ItemId())
	if uniqueRow != nil {
		uniqueRow.Add(columnIndex, 1)
	}

	return columnIndex
}

func (sit *gearItemSetupShared) finishItemsEquipped(itemOptions *items.SolvableOptionsMap, build *util_highs.LinearBuilder) error {
	// constrain: exactly one item for each slot
	for slot, row := range sit.slotsOneEachRow {
		slotEquip := items.SlotEquip(slot)
		row.Debug = "slotsOneEachRow_" + slotEquip.Name()
		if itemOptions.Has(slotEquip) && !row.IsEmpty() {
			row.Build(build, 1, 1)
		} else if !row.IsEmpty() {
			row.Build(build, 0, 0)
		} else if itemOptions.Has(slotEquip) {
			return errors.New("lost some item options for " + slotEquip.Name())
		}
	}

	// constrain: unique item by itemid/unique set
	for _, row := range sit.uniqueEquipRowsAll {
		if !row.IsEmpty() {
			row.Build(build, 0, 1)
		}
	}

	return nil
}

func (sit *gearItemSetupShared) finishSetCounts(build *util_highs.LinearBuilder) map[solve_highs_types.SetBonusIndex]*columnInfo {
	countSetItemsCol := make(map[solve_highs_types.SetBonusIndex]*columnInfo)

	// constrain: matching number of items from each given set
	for i := range sit.countSetItemsRow {
		setIndex := solve_highs_types.SetBonusIndex(i)
		setRow := &sit.countSetItemsRow[setIndex]
		setRow.Debug = "countSetItemsRow" + strconv.Itoa(i)

		entry := &columnInfo{entryType: entry_set_total_count, setIndex: setIndex}
		entry.columnIndex = build.CreateColumnGeneral(highs.Integer, 0, c_maxSetItems, entry)
		countSetItemsCol[setIndex] = entry

		setRow.Add(entry.columnIndex, -1)
		setRow.Build(build, 0, 0)
	}

	return countSetItemsCol
}
