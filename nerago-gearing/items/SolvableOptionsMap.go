package items

import (
	"iter"
	"maps"
	"paladin_gearing_go/util/util_collection"
	"slices"
)

type SolvableOptionsMap struct {
	array              [ITEM_SLOT_COUNT][]SolvableItem
	uniqueEquippedSets [][]ItemId
}

func SolvableOptionsMap_of(fullMap *FullOptionsMap) SolvableOptionsMap {
	result := SolvableOptionsMap{}
	for slot := range fullMap {
		result.array[slot] = util_collection.MapSliceAsNew(fullMap[slot], SolvableItem_Of)
	}
	result.uniqueEquippedSets = UniqueEquipSetsInOptions(fullMap)
	return result
}

func (optionsMap *SolvableOptionsMap) Get(slot SlotEquip) []SolvableItem {
	return optionsMap.array[slot]
}

func (optionsMap *SolvableOptionsMap) Has(slot SlotEquip) bool {
	return len(optionsMap.array[slot]) > 0
}

func (optionsMap *SolvableOptionsMap) Set(slot SlotEquip, options []SolvableItem) {
	optionsMap.array[slot] = options
}

func (optionsMap *SolvableOptionsMap) Clone() SolvableOptionsMap {
	// send existing uniqueEquipped as-is, assuming we're cutting options after Clone
	// don't need to manipulate some will just become non-applicable
	clone := SolvableOptionsMap{uniqueEquippedSets: optionsMap.uniqueEquippedSets}

	// do copy the item slices
	for i := range optionsMap.array {
		clone.array[i] = slices.Clone(optionsMap.array[i])
	}
	return clone
}

func (optionsMap *SolvableOptionsMap) RemoveItemIdFromAll(itemId ItemId) (makesSlotEmpty bool) {
	for slot := range optionsMap.array {
		if len(optionsMap.array[slot]) > 0 {
			util_collection.FilterSliceInPlace(&optionsMap.array[slot], func(x *SolvableItem) bool { return x.ItemId() != itemId })
			if len(optionsMap.array[slot]) == 0 {
				return true
			}
		}
	}
	return false
}

func (optionsMap *SolvableOptionsMap) AllItemSeq() iter.Seq[*SolvableItem] {
	return func(yield func(*SolvableItem) bool) {
		for slot := range optionsMap.array {
			for i := range optionsMap.array[slot] {
				if !yield(&optionsMap.array[slot][i]) {
					return
				}
			}
		}
	}
}

func (optionsMap *SolvableOptionsMap) AllItemSlotSeq() iter.Seq2[SlotEquip, *SolvableItem] {
	return func(yield func(SlotEquip, *SolvableItem) bool) {
		for slot := Equip_Iter_First; slot <= Equip_Iter_Last; slot++ {
			for i := range optionsMap.array[slot] {
				if !yield(slot, &optionsMap.array[slot][i]) {
					return
				}
			}
		}
	}
}

func (optionsMap *SolvableOptionsMap) CountSlotUniqueItemIds(slotEquip SlotEquip) int {
	idMap := make(map[ItemId]bool)
	for i := range optionsMap.array[slotEquip] {
		id := optionsMap.array[slotEquip][i].itemId
		idMap[id] = true
	}
	return len(idMap)
}

func (optionsMap *SolvableOptionsMap) SeqSlotUniqueItemIds(slotEquip SlotEquip) iter.Seq[ItemId] {
	idMap := make(map[ItemId]bool)
	for i := range optionsMap.array[slotEquip] {
		id := optionsMap.array[slotEquip][i].itemId
		idMap[id] = true
	}
	return maps.Keys(idMap)
}

func (optionsMap *SolvableOptionsMap) SlotItemSeq(slotEquip SlotEquip) iter.Seq[*SolvableItem] {
	return func(yield func(*SolvableItem) bool) {
		for i := range optionsMap.array[slotEquip] {
			if !yield(&optionsMap.array[slotEquip][i]) {
				return
			}
		}
	}
}

func (optionsMap *SolvableOptionsMap) SlotSliceSeq() iter.Seq2[SlotEquip, []SolvableItem] {
	return func(yield func(SlotEquip, []SolvableItem) bool) {
		for slot := Equip_Iter_First; slot <= Equip_Iter_Last; slot++ {
			if !yield(slot, optionsMap.array[slot]) {
				return
			}
		}
	}
}

func (optionsMap *SolvableOptionsMap) SlotNestedSeq() iter.Seq2[SlotEquip, iter.Seq[*SolvableItem]] {
	return func(yield func(SlotEquip, iter.Seq[*SolvableItem]) bool) {
		for slot := Equip_Iter_First; slot <= Equip_Iter_Last; slot++ {
			if !yield(slot, func(yield2 func(*SolvableItem) bool) {
				for i := range optionsMap.array[slot] {
					if !yield2(&optionsMap.array[slot][i]) {
						return
					}
				}
			}) {
				return
			}
		}
	}
}

func (optionsMap *SolvableOptionsMap) UniqueEquippedSets() [][]ItemId {
	return optionsMap.uniqueEquippedSets
}
