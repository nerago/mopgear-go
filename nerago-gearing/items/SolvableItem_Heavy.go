//go:build !statslite

package items

import (
	. "paladin_gearing_go/stats"
)

// /////////////////////////////////////////////////////////////
type SolvableItem struct {
	ItemId     uint32
	totalCap   StatBlock
	totalRated StatBlock
}

func SolvableItem_Of(item FullItem) SolvableItem {
	return SolvableItem{
		item.Ref.ItemId,
		item.totalCap,
		item.totalRated}
}

func SolvableItem_ForTest(itemid uint32, block StatBlock) SolvableItem {
	return SolvableItem{
		itemid,
		block,
		block}
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
	set.Items = SolvableEquipMap{}
	set.totalCap = StatBlock{}
	set.totalRated = StatBlock{}
}

func (set *SolvableItemSet) AddItem_Mutating(slot SlotEquip, item *SolvableItem) {
	if set.Items[slot] != nil {
		panic("slot not empty")
	}

	set.Items[slot] = item
	StatBlock_Increment_Mutating(&set.totalCap, &item.totalCap)
	StatBlock_Increment_Mutating(&set.totalRated, &item.totalRated)
}

func (set *SolvableItemSet) AddItem_CreateNew(slot SlotEquip, item *SolvableItem) SolvableItemSet {
	if set.Items[slot] != nil {
		panic("slot not empty")
	}

	result := SolvableItemSet{}
	result.Items = set.Items
	result.Items[slot] = item
	StatBlock_Add_Into(&set.totalCap, &item.totalCap, &result.totalCap)
	StatBlock_Add_Into(&set.totalRated, &item.totalRated, &result.totalRated)
	return result
}

func (item *SolvableItemSet) TotalCap() *StatBlock {
	return &item.totalCap
}

func (item *SolvableItemSet) TotalRated() *StatBlock {
	return &item.totalRated
}

func isMatch(fullItem *FullItem, solveItem *SolvableItem) bool {
	// TODO is it okay to not check item level
	return fullItem.ItemId() == solveItem.ItemId &&
		fullItem.totalCap == solveItem.totalCap &&
		fullItem.totalRated == solveItem.totalRated
}
