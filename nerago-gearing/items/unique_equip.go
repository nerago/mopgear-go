package items

import "slices"

func (optionsMap *FullOptionsMap) IncludesUniqueEquippedViolationInSlot(itemName string, slot SlotEquip) bool {
	for _, item := range optionsMap[slot] {
		if UniqueEquipViolation(item.BaseName(), itemName) {
			return true
		}
	}
	return false
}

var _uniqueSets = [][]string{
	{"Loop of the Shado-Pan Assault", "Band of the Shado-Pan Assault"},
}

var UniqueItemIdSets = [][]ItemId{
	{95140, 95141}, // shado assault band, shado assault loop
	{95513, 96500}, // Band of the Scaled Tyrant: normal/heroic
}

func UniqueEquipViolation(a, b string) bool {
	if a == b {
		return true
	}

	for _, set := range _uniqueSets {
		if slices.Contains(set, a) && slices.Contains(set, b) {
			return true
		}
	}

	return false
}
