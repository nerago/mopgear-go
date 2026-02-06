package items

import (
	. "paladin_gearing_go/types/common"
	. "paladin_gearing_go/types/stats"
)

type SolvableItem struct {
	itemId     int64
	itemLevel  int16
	slot       SlotItem
	reforge    ReforgeRecipe
	gemChoice  []int32
	totalCap   StatBlock
	totalRated StatBlock
}

type SolvableEquipMap [16]SolvableItem

type SolvableItemSet struct {
	items      SolvableEquipMap
	totalCap   StatBlock
	totalRated StatBlock
}
