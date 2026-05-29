package items

import (
	"paladin_gearing_go/util"
	"slices"
)

func (optionsMap *FullOptionsMap) IncludesUniqueEquippedViolationInSlot(search *FullItem, slot SlotEquip) bool {
	for i := range optionsMap[slot] {
		item := &optionsMap[slot][i]
		if UniqueEquipViolation(search, item) {
			return true
		}
	}
	return false
}

var UniqueItemIdSets = [][]ItemId{
	{95140, 95141}, // shado assault band, shado assault loop
}

func UniqueEquipViolation(a, b *FullItem) bool {
	if a == nil || b == nil {
		return false
	}

	if a.ItemId() == b.ItemId() || a.BaseName() == b.BaseName() {
		return true
	}

	for _, set := range UniqueItemIdSets {
		if slices.Contains(set, a.ItemId()) && slices.Contains(set, b.ItemId()) {
			return true
		}
	}

	return false
}

// TODO should have the name collision items in UniqueItemIdSets
func UniqueEquipViolationSolve(a, b *SolvableItem) bool {
	if a == nil || b == nil {
		return false
	}

	if a.ItemId() == b.ItemId() {
		return true
	}

	for _, set := range UniqueItemIdSets {
		if slices.Contains(set, a.ItemId()) && slices.Contains(set, b.ItemId()) {
			return true
		}
	}

	return false
}

func UniqueEquipSetsInOptions(itemOptions *FullOptionsMap) [][]ItemId {
	uniqueSets := make([][]ItemId, 0, len(UniqueItemIdSets))

	predefined := make(map[ItemId]bool)
	for _, set := range UniqueItemIdSets {
		for _, itemId := range set {
			predefined[itemId] = true
		}
		uniqueSets = append(uniqueSets, set)
	}

	grouped := util.MapMap[string, ItemId, bool]{}
	pairedSlots := []SlotEquip{Equip_Ring1, Equip_Ring2, Equip_Trinket1, Equip_Trinket2}
	for _, slotEquip := range pairedSlots {
		for item := range itemOptions.SlotItemSeq(slotEquip) {
			if !predefined[item.ItemId()] {
				grouped.Put(item.BaseName(), item.ItemId(), true)
			}
		}
	}

	for _, itemSeq := range grouped.SeqKey1Key2Nested() {
		idList := slices.Collect(itemSeq)
		uniqueSets = append(uniqueSets, idList)
	}

	return uniqueSets
}
