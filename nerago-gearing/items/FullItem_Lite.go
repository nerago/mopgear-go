//go:build statslite

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
	Items FullEquipMap
	total StatBlock
}

func FullItemSet_FromSolved(solvedSet SolvableItemSet, optionsMap *FullOptionsMap) FullItemSet {
	fullMap := FullEquipMap{}
	for slot, solveItem := range solvedSet.Items {
		if solveItem != nil {
			fullItem := findMatch(optionsMap[slot], solveItem)
			fullMap[slot] = fullItem
		}
	}
	return FullItemSet{Items: fullMap, total: solvedSet.total}
}

func (itemSet *FullItemSet) Equals(other FullItemSet) bool {
	return itemSet.Items.Equals(&other.Items)
}

func (itemSet *FullItemSet) PrintStats(printer *util.PrintRecorder) {
	printer.Printf("STATS %s\n", itemSet.total.String())
}

func (item *FullItemSet) TotalCap() *StatBlock {
	return &item.total
}

func (item *FullItemSet) TotalRated() *StatBlock {
	return &item.total
}
