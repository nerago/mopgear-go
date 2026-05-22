package items

import "slices"

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
	{95513, 96500}, // Band of the Scaled Tyrant: normal/heroic
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
