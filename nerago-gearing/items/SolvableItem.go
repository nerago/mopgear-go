package items

import (
	"iter"
	. "paladin_gearing_go/stats"
)

type SolvableItem struct {
	ItemId uint32
	// ItemLevel  int16
	// Slot       SlotItem
	// Reforge    ReforgeRecipe
	// GemChoice  []int32
	TotalCap   StatBlock
	TotalRated StatBlock
}

func (item *SolvableItem) IsEmpty() bool {
	return item.ItemId == 0
}

type SolvableEquipMap [16]*SolvableItem

func (equipMap SolvableEquipMap) WithAdditional(slot SlotEquip, item *SolvableItem) SolvableEquipMap {
	var result SolvableEquipMap = equipMap
	result[slot] = item
	return result
}

func (equipMap SolvableEquipMap) AllItemSeq() iter.Seq[*SolvableItem] {
	return func(yield func(*SolvableItem) bool) {
		for _, item := range equipMap {
			if item != nil {
				if !yield(item) {
					return
				}
			}
		}
	}
}

type SolvableItemSet struct {
	Items      SolvableEquipMap
	TotalCap   StatBlock
	TotalRated StatBlock
}

func SolvableItemSet_Of(equipMap SolvableEquipMap) SolvableItemSet {
	result := SolvableItemSet{equipMap, StatBlock{}, StatBlock{}}
	for _, item := range equipMap {
		if item != nil {
			StatBlock_Increment_Mutating(&result.TotalCap, &item.TotalCap)
			StatBlock_Increment_Mutating(&result.TotalRated, &item.TotalRated)
		}
	}
	return result
}

func SolvableItemSet_SingleItem(slot SlotEquip, item *SolvableItem) SolvableItemSet {
	equip := SolvableEquipMap{}
	equip[slot] = item
	return SolvableItemSet{
		Items:      equip,
		TotalCap:   item.TotalCap,
		TotalRated: item.TotalRated}
}

func (set *SolvableItemSet) Clear() {
	set.Items = SolvableEquipMap{}
	set.TotalCap = StatBlock{}
	set.TotalRated = StatBlock{}
}

func (set *SolvableItemSet) AddItem_Mutating(slot SlotEquip, item *SolvableItem) {
	if set.Items[slot] != nil {
		panic("slot not empty")
	}

	set.Items[slot] = item
	StatBlock_Increment_Mutating(&set.TotalCap, &item.TotalCap)
	StatBlock_Increment_Mutating(&set.TotalRated, &item.TotalRated)
}

func (set *SolvableItemSet) AddItem_CreateNew(slot SlotEquip, item *SolvableItem) SolvableItemSet {
	if set.Items[slot] != nil {
		panic("slot not empty")
	}

	result := SolvableItemSet{}
	result.Items = set.Items
	result.Items[slot] = item
	StatBlock_Add_Into(&set.TotalCap, &item.TotalCap, &result.TotalCap)
	StatBlock_Add_Into(&set.TotalRated, &item.TotalRated, &result.TotalRated)
	return result
}
