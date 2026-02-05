package items

import . "paladin_gearing_go/types/stats"

type SolvableItem struct {
	itemId    int64
	itemLevel int16
	//slot          SlotItem
	reforge    ReforgeRecipe
	gemChoice  []int32
	totalCap   StatBlock
	totalRated StatBlock
}

type SolveEquipMap [16]SolvableItem

type SolvableItemSet struct {
	items      FullEquipMap
	totalCap   StatBlock
	totalRated StatBlock
}
