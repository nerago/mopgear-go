package items

import (
	. "paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

// /////////////////////////////////////////////////////////////
type FullItem struct {
	fullItem_common
	total StatBlock // constant total stats as they contribute to caps
}

func FullItem_FromWowSim(ref ItemRef, slot SlotItem, baseName string, statBase StatBlock, armorType ArmorType, socketSlots []SocketType, socketBonus StatBlock, phase int8) FullItem {
	return FullItem{
		fullItem_common{ref, slot, baseName, armorType, statBase.PrimaryStat(),
			socketSlots, socketBonus, phase,
			ReforgeRecipe_empty, nil, 0, 0,
			statBase, StatBlock_empty},
		statBase}
}

func FullItem_ForTest(itemId uint32, slot SlotItem, statBase StatBlock) FullItem {
	return FullItem{
		fullItem_common{ItemRef{itemId, 400, 400},
			slot, slot.Name(), Armor_None, statBase.PrimaryStat(),
			nil, StatBlock_empty, 0,
			ReforgeRecipe_empty, nil, 0, 0,
			statBase, StatBlock_empty},
		statBase}
}

func (item *FullItem) ChangeDerivedStatFields() {
	StatBlock_Add_Into(&item.StatBase, &item.StatEnchant, &item.total)
}

func (item *FullItem) TotalCap() *StatBlock {
	return &item.total
}

func (item *FullItem) TotalRated() *StatBlock {
	return &item.total
}

// /////////////////////////////////////////////////////////////
type FullItemSet struct {
	items FullEquipMap
	total StatBlock
}

func FullItemSet_FromSolved(solvedSet SolvableItemSet, optionsMap *FullOptionsMap) FullItemSet {
	fullMap := FullEquipMap{}
	for slot, solveItem := range solvedSet.items {
		if solveItem != nil {
			fullItem := findMatch(optionsMap[slot], solveItem)
			fullMap[slot] = fullItem
		}
	}
	return FullItemSet{items: fullMap, total: solvedSet.total}
}

func (itemSet *FullItemSet) PrintStats(printer *util.PrintRecorder) {
	printer.Printf("STATS %s\n", itemSet.total.CreateString())
}

func (itemSet *FullItemSet) TotalCap() *StatBlock {
	return &itemSet.total
}

func (itemSet *FullItemSet) TotalRated() *StatBlock {
	return &itemSet.total
}

func (itemSet *FullItemSet) Items() *FullEquipMap {
	return &itemSet.items
}

func (itemSet *FullItemSet) ItemsGeneric() IEquipMap {
	return &itemSet.items
}
