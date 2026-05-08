package items

import "paladin_gearing_go/stats"

// /////////////////////////////////////////////////////////////
type SolvableItemSet struct {
	total stats.StatBlock
	items SolvableEquipMap
}

func SolvableItemSet_Of(equipMap SolvableEquipMap) SolvableItemSet {
	result := SolvableItemSet{items: equipMap, total: stats.StatBlock{}}
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

func (set *SolvableItemSet) Clear() {
	set.items = SolvableEquipMap{}
	set.total = stats.StatBlock{}
}

func (set *SolvableItemSet) ClearTotals() {
	set.total = stats.StatBlock{}
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

func (set *SolvableItemSet) ReplaceItem_Into(slot SlotEquip, item *SolvableItem, dest *SolvableItemSet) {
	oldItem := set.items[slot]
	set.items.ReplaceItem_Into(slot, item, &dest.items)
	stats.StatBlock_AddAndSubtract_Into(&set.total, &item.total, &oldItem.total, &dest.total)
}

func (set *SolvableItemSet) TotalCap() *stats.StatBlock {
	return &set.total
}

func (set *SolvableItemSet) TotalRated() *stats.StatBlock {
	return &set.total
}

func isMatch(fullItem *FullItem, solveItem *SolvableItem) bool {
	return fullItem.ItemId() == solveItem.itemId &&
		stats.StatBlock_Equals(&fullItem.total, &solveItem.total)
}
