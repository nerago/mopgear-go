package items

import (
	"paladin_gearing_go/stats"
	. "paladin_gearing_go/stats"
)

// /////////////////////////////////////////////////////////////
type SolvableItem struct {
	total  StatBlock
	itemId ItemId
}

func SolvableItem_Of(item FullItem) SolvableItem {
	return SolvableItem{
		itemId: item.Ref.ItemId,
		total:  item.total}
}

func SolvableItem_ForTest(itemid ItemId, block StatBlock) SolvableItem {
	return SolvableItem{
		itemId: itemid,
		total:  block}
}

func (item *SolvableItem) ItemId() ItemId {
	return item.itemId
}

func (item *SolvableItem) IsEmpty() bool {
	return item.itemId == 0
}

func (item *SolvableItem) TotalCap() *StatBlock {
	return &item.total
}

func (item *SolvableItem) TotalRated() *StatBlock {
	return &item.total
}

func (item *SolvableItem) Equals(other *SolvableItem) bool {
	return item.itemId == other.itemId && stats.StatBlock_Equals(&item.total, &other.total)
}

func (item *SolvableItem) EqualsFull(other *FullItem) bool {
	return item.itemId == other.ItemId() && stats.StatBlock_Equals(&item.total, &other.total)
}

// /////////////////////////////////////////////////////////////
type SolvableItemSet struct {
	total StatBlock
	items SolvableEquipMap
}

func SolvableItemSet_Of(equipMap SolvableEquipMap) SolvableItemSet {
	result := SolvableItemSet{items: equipMap, total: StatBlock{}}
	SolvableItemSet_RecalculateTotal(&result)
	return result
}

func SolvableItemSet_SingleItem(slot SlotEquip, item *SolvableItem) SolvableItemSet {
	equip := SolvableEquipMap{}
	equip[slot] = item
	return SolvableItemSet{
		items: equip,
		total: item.total}
}

func (set *SolvableItemSet) Items() *SolvableEquipMap {
	return &set.items
}

func (set *SolvableItemSet) ItemsGeneric() IEquipMap {
	return &set.items
}

func (set *SolvableItemSet) Clear() {
	set.items = SolvableEquipMap{}
	set.total = StatBlock{}
}

func (set *SolvableItemSet) ClearTotals() {
	set.total = StatBlock{}
}

// obsolete, use AddItem_DeferCalc and SolvableItemSet_RecalculateTotal
// func (set *SolvableItemSet) AddItem_Mutating(slot SlotEquip, item *SolvableItem) {
// 	set.items[slot] = item
// 	StatBlock_Increment_Mutating(&set.total, &item.total)
// }

func (set *SolvableItemSet) AddItem_DeferCalc(slot SlotEquip, item *SolvableItem) {
	set.items[slot] = item
}

func (set *SolvableItemSet) AddItem_DeferCalc_ExpectEmpty(slot SlotEquip, item *SolvableItem) {
	if set.items[slot] != nil {
		panic("slot not empty")
	}
	set.items[slot] = item
}

// obsolete
// func (set *SolvableItemSet) AddItem_CreateNew(slot SlotEquip, item *SolvableItem) *SolvableItemSet {
// 	result := new(SolvableItemSet)
// 	set.items.ReplaceItem_Into(slot, item, &result.items)
// 	StatBlock_Add_Into(&set.total, &item.total, &result.total)
// 	return result
// }

func (set *SolvableItemSet) AddItem_CreateNew_DeferCalc(slot SlotEquip, item *SolvableItem) *SolvableItemSet {
	result := new(SolvableItemSet)
	set.items.ReplaceItem_Into(slot, item, &result.items)
	return result
}

func (set *SolvableItemSet) ReplaceItem_Into(slot SlotEquip, item *SolvableItem, dest *SolvableItemSet) {
	oldItem := set.items[slot]
	set.items.ReplaceItem_Into(slot, item, &dest.items)
	StatBlock_AddAndSubtract_Into(&set.total, &item.total, &oldItem.total, &dest.total)
}

func (set *SolvableItemSet) TotalCap() *StatBlock {
	return &set.total
}

func (set *SolvableItemSet) TotalRated() *StatBlock {
	return &set.total
}

func isMatch(fullItem *FullItem, solveItem *SolvableItem) bool {
	return fullItem.ItemId() == solveItem.itemId &&
		StatBlock_Equals(&fullItem.total, &solveItem.total)
}
