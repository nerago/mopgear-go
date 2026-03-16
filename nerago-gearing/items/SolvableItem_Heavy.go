//go:build !statslite

package items

import (
	. "paladin_gearing_go/stats"
)

// /////////////////////////////////////////////////////////////
type SolvableItem struct {
	totalCap   StatBlock
	totalRated StatBlock
	itemId     uint32
	// TODO try with precalculated ratings from model
	// TODO try with set bonus info
}

func SolvableItem_Of(item FullItem) SolvableItem {
	return SolvableItem{
		itemId:     item.Ref.ItemId,
		totalCap:   item.totalCap,
		totalRated: item.totalRated}
}

func SolvableItem_ForTest(itemid uint32, block StatBlock) SolvableItem {
	return SolvableItem{
		itemId:     itemid,
		totalCap:   block,
		totalRated: block}
}

func (item *SolvableItem) ItemId() uint32 {
	return item.itemId
}

func (item *SolvableItem) IsEmpty() bool {
	return item.itemId == 0
}

func (item *SolvableItem) TotalCap() *StatBlock {
	return &item.totalCap
}

func (item *SolvableItem) TotalRated() *StatBlock {
	return &item.totalRated
}

// /////////////////////////////////////////////////////////////
type SolvableItemSet struct {
	totalCap   StatBlock
	totalRated StatBlock
	items      SolvableEquipMap
}

func SolvableItemSet_Of(equipMap SolvableEquipMap) SolvableItemSet {
	result := SolvableItemSet{items: equipMap, totalCap: StatBlock{}, totalRated: StatBlock{}}
	for _, item := range equipMap {
		if item != nil {
			StatBlock_Increment_Mutating(&result.totalCap, &item.totalCap)
			StatBlock_Increment_Mutating(&result.totalRated, &item.totalRated)
		}
	}
	return result
}

func SolvableItemSet_SingleItem(slot SlotEquip, item *SolvableItem) SolvableItemSet {
	equip := SolvableEquipMap{}
	equip[slot] = item
	return SolvableItemSet{
		items:      equip,
		totalCap:   item.totalCap,
		totalRated: item.totalRated}
}

func (set *SolvableItemSet) Items() *SolvableEquipMap {
	return &set.items
}

func (set *SolvableItemSet) ItemsGeneric() IEquipMap {
	return &set.items
}

func (set *SolvableItemSet) Clear() {
	*set = SolvableItemSet{}
	// set.Items = SolvableEquipMap{}
	// set.totalCap = StatBlock{}
	// set.totalRated = StatBlock{}
}

func (set *SolvableItemSet) ClearTotals() {
	set.totalCap = StatBlock{}
	set.totalRated = StatBlock{}
}

func (set *SolvableItemSet) AddItem_Mutating(slot SlotEquip, item *SolvableItem) {
	set.items[slot] = item
	StatBlock_Increment_Mutating(&set.totalCap, &item.totalCap)
	StatBlock_Increment_Mutating(&set.totalRated, &item.totalRated)
}

func (set *SolvableItemSet) AddItem_CreateNew(slot SlotEquip, item *SolvableItem) *SolvableItemSet {
	result := new(SolvableItemSet)
	set.items.ReplaceItem_Into(slot, item, &result.items)
	StatBlock_Add_Into(&set.totalCap, &item.totalCap, &result.totalCap)
	StatBlock_Add_Into(&set.totalRated, &item.totalRated, &result.totalRated)
	return result
}

func (set *SolvableItemSet) AddItem_Into(slot SlotEquip, item *SolvableItem, dest *SolvableItemSet) {
	set.items.ReplaceItem_Into(slot, item, &dest.items)
	StatBlock_Add_Into(&set.totalCap, &item.totalCap, &dest.totalCap)
	StatBlock_Add_Into(&set.totalRated, &item.totalRated, &dest.totalRated)
}

func (set *SolvableItemSet) ReplaceItem_Into(slot SlotEquip, item *SolvableItem, dest *SolvableItemSet) {
	oldItem := set.items[slot] // assumed to be non-nil
	set.items.ReplaceItem_Into(slot, item, &dest.items)
	StatBlock_AddAndSubtract_Into(&set.totalCap, &item.totalCap, &oldItem.totalCap, &dest.totalCap)
	StatBlock_AddAndSubtract_Into(&set.totalRated, &item.totalRated, &oldItem.totalRated, &dest.totalRated)
}

func (set *SolvableItemSet) TotalCap() *StatBlock {
	return &set.totalCap
}

func (set *SolvableItemSet) TotalRated() *StatBlock {
	return &set.totalRated
}

func isMatch(fullItem *FullItem, solveItem *SolvableItem) bool {
	return fullItem.ItemId() == solveItem.itemId &&
		StatBlock_Equals(&fullItem.totalCap, &solveItem.totalCap) &&
		StatBlock_Equals(&fullItem.totalRated, &solveItem.totalRated)
}
