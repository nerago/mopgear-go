package items

import (
	// . "paladin_gearing_go/types/common"
	. "paladin_gearing_go/types/stats"
)

type SolvableItem struct {
	ItemId int32
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

func SolvableItem_Of(item FullItem) SolvableItem {
	return SolvableItem{
		item.ref.itemId,
		// item.ref.itemLevel,
		// item.slot,
		// item.reforge,
		// item.gemChoice,
		item.totalCap,
		item.totalRated}
}

type SolvableEquipMap [16]*SolvableItem

type SolvableItemSet struct {
	Items      SolvableEquipMap
	TotalCap   StatBlock
	TotalRated StatBlock
}

func SolvableItemSet_Of(equipMap SolvableEquipMap) *SolvableItemSet {
	result := SolvableItemSet{equipMap, StatBlock{}, StatBlock{}}
	for _, item := range equipMap {
		result.TotalCap.Increment_Mutating(&item.TotalCap)
		result.TotalRated.Increment_Mutating(&item.TotalRated)
	}
	return &result
}
