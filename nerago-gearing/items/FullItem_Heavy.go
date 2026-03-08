//go:build !statslite

package items

import (
	. "paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

// /////////////////////////////////////////////////////////////
type FullItem struct {
	fullItem_common
	totalCap   StatBlock // constant total stats as they contribute to caps
	totalRated StatBlock // averaged variable total stats for rating purposes
}

func FullItem_FromWowSim(ref ItemRef, slot SlotItem, baseName string, statBase StatBlock, armorType ArmorType, socketSlots []SocketType, socketBonus StatBlock, phase int8) FullItem {
	return FullItem{
		fullItem_common{ref, slot, baseName, armorType, statBase.PrimaryStat(),
			socketSlots, socketBonus, phase,
			ReforgeRecipe_empty, nil, 0, 0,
			statBase, StatBlock_empty},
		statBase, statBase}
}

func (item *FullItem) ChangeDerivedStatFields() {
	StatBlock_Add_Into(&item.StatBase, &item.StatEnchant, &item.totalRated)
	if item.Slot.AddEnchantToCap() {
		item.totalCap = item.totalRated
	} else {
		item.totalCap = item.StatBase
	}
}

func (item *FullItem) TotalCap() *StatBlock {
	return &item.totalCap
}

func (item *FullItem) TotalRated() *StatBlock {
	return &item.totalRated
}

// /////////////////////////////////////////////////////////////
type FullItemSet struct {
	Items      FullEquipMap
	totalCap   StatBlock
	totalRated StatBlock
}

func FullItemSet_FromSolved(solvedSet SolvableItemSet, optionsMap *FullOptionsMap) FullItemSet {
	fullMap := FullEquipMap{}
	for slot, solveItem := range solvedSet.Items {
		if solveItem != nil {
			fullItem := findMatch(optionsMap[slot], solveItem)
			fullMap[slot] = fullItem
		}
	}
	return FullItemSet{Items: fullMap, totalCap: solvedSet.totalCap, totalRated: solvedSet.totalRated}
}

func (itemSet *FullItemSet) Equals(other FullItemSet) bool {
	return itemSet.Items.Equals(&other.Items)
}

func (itemSet *FullItemSet) PrintStats(printer *util.PrintRecorder) {
	printer.Printf("RATED %s\n", itemSet.totalRated.String())
	printer.Printf("CAP %s\n", itemSet.totalCap.String())
}

func (itemSet *FullItemSet) TotalCap() *StatBlock {
	return &itemSet.totalCap
}

func (itemSet *FullItemSet) TotalRated() *StatBlock {
	return &itemSet.totalRated
}
