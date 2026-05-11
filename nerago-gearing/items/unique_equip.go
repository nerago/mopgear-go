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
