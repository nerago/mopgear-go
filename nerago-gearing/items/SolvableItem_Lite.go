//go:build statslite

package items

import (
	. "paladin_gearing_go/stats"
)

// /////////////////////////////////////////////////////////////
type SolvableItem struct {
	total  StatBlock
	ItemId uint32
}

func SolvableItem_Of(item FullItem) SolvableItem {
	return SolvableItem{
		ItemId: item.Ref.ItemId,
		total: item.total}
}

func (item *SolvableItem) IsEmpty() bool {
	return item.ItemId == 0
}

func (item *SolvableItem) TotalCap() *StatBlock {
	return &item.total
}

func (item *SolvableItem) TotalRated() *StatBlock {
	return &item.total
}

// /////////////////////////////////////////////////////////////
type SolvableItemSet struct {
	total StatBlock
	Items SolvableEquipMap
}

func SolvableItemSet_Of(equipMap SolvableEquipMap) SolvableItemSet {
	result := SolvableItemSet{Items: equipMap, total: StatBlock{}}
	for _, item := range equipMap {
		if item != nil {
			StatBlock_Increment_Mutating(&result.total, &item.total)
		}
	}
	return result
}

func SolvableItemSet_SingleItem(slot SlotEquip, item *SolvableItem) SolvableItemSet {
	equip := SolvableEquipMap{}
	equip[slot] = item
	return SolvableItemSet{
		Items: equip,
		total: item.total}
}

func (set *SolvableItemSet) Clear() {
	set.Items = SolvableEquipMap{}
	set.total = StatBlock{}
}

func (set *SolvableItemSet) ClearTotals() {
	set.total = StatBlock{}
}

func (set *SolvableItemSet) AddItem_Mutating(slot SlotEquip, item *SolvableItem) {
	set.Items[slot] = item
	StatBlock_Increment_Mutating(&set.total, &item.total)
}

func (set *SolvableItemSet) AddItem_CreateNew(slot SlotEquip, item *SolvableItem) *SolvableItemSet {
	result := new(SolvableItemSet)
	set.Items.ReplaceItem_Into(slot, item, &result.Items)
	StatBlock_Add_Into(&set.total, &item.total, &result.total)
	return result
}

func (set *SolvableItemSet) ReplaceItem_Into(slot SlotEquip, item *SolvableItem, dest *SolvableItemSet) {
	oldItem := set.Items[slot]
	set.Items.ReplaceItem_Into(slot, item, &dest.Items)
	StatBlock_AddAndSubtract_Into(&set.total, &item.total, &oldItem.total, &dest.total)
}

func (set *SolvableItemSet) TotalCap() *StatBlock {
	return &set.total
}

func (set *SolvableItemSet) TotalRated() *StatBlock {
	return &set.total
}

func isMatch(fullItem *FullItem, solveItem *SolvableItem) bool {
	return fullItem.ItemId() == solveItem.ItemId &&
		StatBlock_Equals(&fullItem.total, &solveItem.total)
}
