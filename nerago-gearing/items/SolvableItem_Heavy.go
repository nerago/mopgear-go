//go:build !statslite

package items

import (
	. "paladin_gearing_go/stats"
)

// /////////////////////////////////////////////////////////////
type SolvableItem struct {
	totalCap   StatBlock
	totalRated StatBlock
	ItemId     uint32
}

func SolvableItem_Of(item FullItem) SolvableItem {
	return SolvableItem{
		ItemId: item.Ref.ItemId,
		totalCap: item.totalCap,
		totalRated: item.totalRated}
}

func SolvableItem_ForTest(itemid uint32, block StatBlock) SolvableItem {
	return SolvableItem{
		ItemId: itemid,
		totalCap: block,
		totalRated: block}
}

func (item *SolvableItem) IsEmpty() bool {
	return item.ItemId == 0
}

func (item *SolvableItem) TotalCap() *StatBlock {
	return &item.totalCap
}

func (item *SolvableItem) TotalRated() *StatBlock {
	return &item.totalRated
}

// /////////////////////////////////////////////////////////////
type SolvableItemSet struct {
	Items      SolvableEquipMap
	totalCap   StatBlock
	totalRated StatBlock
}

func SolvableItemSet_Of(equipMap SolvableEquipMap) SolvableItemSet {
	result := SolvableItemSet{equipMap, StatBlock{}, StatBlock{}}
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
		Items:      equip,
		totalCap:   item.totalCap,
		totalRated: item.totalRated}
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
	set.Items[slot] = item
	StatBlock_Increment_Mutating(&set.totalCap, &item.totalCap)
	StatBlock_Increment_Mutating(&set.totalRated, &item.totalRated)
}

func (set *SolvableItemSet) AddItem_CreateNew(slot SlotEquip, item *SolvableItem) SolvableItemSet {
	result := SolvableItemSet{}
	result.Items = set.Items
	result.Items[slot] = item
	StatBlock_Add_Into(&set.totalCap, &item.totalCap, &result.totalCap)
	StatBlock_Add_Into(&set.totalRated, &item.totalRated, &result.totalRated)
	return result
}

func (set *SolvableItemSet) TotalCap() *StatBlock {
	return &set.totalCap
}

func (set *SolvableItemSet) TotalRated() *StatBlock {
	return &set.totalRated
}

func isMatch(fullItem *FullItem, solveItem *SolvableItem) bool {
	return fullItem.ItemId() == solveItem.ItemId &&
		StatBlock_Equals(&fullItem.totalCap, &solveItem.totalCap) &&
		StatBlock_Equals(&fullItem.totalRated, &solveItem.totalRated)
}
