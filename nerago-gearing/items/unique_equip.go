package items

import (
	"maps"
	"paladin_gearing_go/util/util_collection"
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

	if (a.ItemId() == b.ItemId() && a.RandomSuffix() == b.RandomSuffix()) || a.BaseName() == b.BaseName() {
		return true
	}

	for _, set := range UniqueItemIdSets {
		if slices.Contains(set, a.ItemId()) && slices.Contains(set, b.ItemId()) {
			return true
		}
	}

	return false
}

func UniqueEquipViolationSolve(a, b *SolvableItem, solvableOptionsMap *SolvableOptionsMap) bool {
	if a == nil || b == nil {
		return false
	}

	for _, set := range solvableOptionsMap.uniqueEquippedSets {
		if slices.Contains(set, a.ItemId()) && slices.Contains(set, b.ItemId()) {
			return true
		}
	}

	return false
}

// NOTE: multiple items with different names but same itemid implies RandomSuffix or similar;
// consider they need uniqueequipped even if they don't really
func UniqueEquipSetsInOptions(itemOptions *FullOptionsMap) [][]ItemId {
	uniqueSets := make([][]ItemId, 0, len(UniqueItemIdSets))

	// mark predefined sets out of scope, but add them directly too
	predefined := make(map[ItemId]bool)
	for _, set := range UniqueItemIdSets {
		for _, itemId := range set {
			predefined[itemId] = true
		}
		uniqueSets = append(uniqueSets, set)
	}

	// build unique lookup maps for all names and ids
	groupedByName := util_collection.MapMap[string, ItemId, bool]{}
	groupedById := util_collection.MapMap[ItemId, string, bool]{}
	for _, slotEquip := range PairedSlotList {
		for item := range itemOptions.SlotItemSeq(slotEquip) {
			if !predefined[item.ItemId()] {
				groupedByName.Put(item.BaseName(), item.ItemId(), true)
				groupedById.Put(item.ItemId(), item.BaseName(), true)
			}
		}
	}

	// start with any one element then go through and find all other names/ids linked to it
	for !groupedByName.IsEmpty() {
		namesFound := make(map[string]bool)
		idsFound := make(map[ItemId]bool)

		startingName := groupedByName.FirstKey1()
		namesFound[startingName] = true

		// dig around the maps, expanding search to all id and names for each
		for {
			foundMore := false

			for name := range namesFound {
				for id := range groupedByName.SeqKey2ValueWithKey1(name) {
					if !idsFound[id] {
						idsFound[id] = true
						foundMore = true
					}
				}
			}

			for id := range idsFound {
				for name := range groupedById.SeqKey2ValueWithKey1(id) {
					if !namesFound[name] {
						namesFound[name] = true
						foundMore = true
					}
				}
			}

			if !foundMore {
				break
			}
		}

		// go back and remove them from the main maps
		for name := range namesFound {
			groupedByName.DeleteAllForKey1(name)
		}
		for id := range idsFound {
			groupedById.DeleteAllForKey1(id)
		}

		// finally collect the resolved ids as a set
		idList := slices.Collect(maps.Keys(idsFound))
		uniqueSets = append(uniqueSets, idList)
	}

	return uniqueSets
}
