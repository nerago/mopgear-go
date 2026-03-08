//go:build statslite

package items

import (
	. "paladin_gearing_go/stats"
)

// /////////////////////////////////////////////////////////////
type SolvableItem struct {
	ItemId uint32
	total  StatBlock
}

func SolvableItem_Of(item FullItem) SolvableItem {
	return SolvableItem{
		item.Ref.ItemId,
		item.total}
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
	Items SolvableEquipMap
	total StatBlock
}

func SolvableItemSet_Of(equipMap SolvableEquipMap) SolvableItemSet {
	result := SolvableItemSet{equipMap, StatBlock{}}
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

func (set *SolvableItemSet) AddItem_Mutating(slot SlotEquip, item *SolvableItem) {
	if set.Items[slot] != nil {
		panic("slot not empty")
	}

	set.Items[slot] = item
	StatBlock_Increment_Mutating(&set.total, &item.total)
}

func (set *SolvableItemSet) AddItem_CreateNew(slot SlotEquip, item *SolvableItem) SolvableItemSet {
	if set.Items[slot] != nil {
		panic("slot not empty")
	}

	result := SolvableItemSet{}
	result.Items = set.Items
	result.Items[slot] = item
	StatBlock_Add_Into(&set.total, &item.total, &result.total)
	return result
}

func (item *SolvableItemSet) TotalCap() *StatBlock {
	return &item.total
}

func (item *SolvableItemSet) TotalRated() *StatBlock {
	return &item.total
}

func isMatch(fullItem *FullItem, solveItem *SolvableItem) bool {
	// TODO is it okay to not check item level
	return fullItem.ItemId() == solveItem.ItemId &&
		fullItem.total == solveItem.total
}
