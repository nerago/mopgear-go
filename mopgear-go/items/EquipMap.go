package items

import (
	"iter"
)

const ITEM_SLOT_COUNT = 16

type FullEquipMap [ITEM_SLOT_COUNT]*FullItem

func (equipMap *FullEquipMap) Get(slot SlotEquip) *FullItem {
	return equipMap[slot]
}

func (equipMap *FullEquipMap) Has(slot SlotEquip) bool {
	return equipMap[slot] != nil
}

func (equipMap *FullEquipMap) GetAsId(slot SlotEquip) ItemId {
	item := equipMap[slot]
	if item != nil {
		return item.ItemId()
	} else {
		return 0
	}
}

func (equipMap *FullEquipMap) IncludesItemId(itemId ItemId) bool {
	for _, item := range equipMap {
		if item != nil && item.ItemId() == itemId {
			return true
		}
	}
	return false
}

func (equipMap *FullEquipMap) FindItemId(itemId ItemId) *FullItem {
	for _, item := range equipMap {
		if item != nil && item.ItemId() == itemId {
			return item
		}
	}
	return nil
}

func (equipMap *FullEquipMap) Equals(other *FullEquipMap) bool {
	for slot := range equipMap {
		a := equipMap[slot]
		b := other[slot]
		if a != nil && b != nil {
			if !a.Equals(b) {
				return false
			}
		} else if a != nil || b != nil {
			return false
		}
	}
	return true
}

func (equipMap *FullEquipMap) AllItemSeq() iter.Seq[*FullItem] {
	return func(yield func(*FullItem) bool) {
		for _, item := range equipMap {
			if item != nil {
				if !yield(item) {
					return
				}
			}
		}
	}
}

// so export doesn't flip around
// TODO fix this upstream
func (equipMap *FullEquipMap) AllItemSeqPairedConsistent() iter.Seq[*FullItem] {
	return func(yield func(*FullItem) bool) {
		for slot := Equip_Iter_First; slot < Equip_Ring1; slot++ {
			if equipMap[slot] != nil {
				if !yield(equipMap[slot]) {
					return
				}
			}
		}
		if !withConsistentOrder(equipMap[Equip_Ring1], equipMap[Equip_Ring2], yield) {
			return
		}
		if !withConsistentOrder(equipMap[Equip_Trinket1], equipMap[Equip_Trinket2], yield) {
			return
		}
		for slot := Equip_Weapon; slot <= Equip_Iter_Last; slot++ {
			if equipMap[slot] != nil {
				if !yield(equipMap[slot]) {
					return
				}
			}
		}
	}
}

func withConsistentOrder(a, b *FullItem, yield func(*FullItem) bool) bool {
	if a.baseName < b.baseName {
		return yield(a) && yield(b)
	} else {
		return yield(b) && yield(a)
	}
}

// //////////////////////////////////////////////////////
type SolvableEquipMap [ITEM_SLOT_COUNT]*SolvableItem

func (equipMap *SolvableEquipMap) WithAdditional(slot SlotEquip, item *SolvableItem) SolvableEquipMap {
	var result SolvableEquipMap = *equipMap
	result[slot] = item
	return result
}

func (equipMap *SolvableEquipMap) ReplaceItem_Into(slot SlotEquip, item *SolvableItem, dest *SolvableEquipMap) {
	// for i := Equip_Iter_First; i < slot; i++ {
	// 	dest[i] = equipMap[i]
	// }
	// dest[slot] = item
	// for i := slot + 1; i <= Equip_Iter_Last; i++ {
	// 	dest[i] = equipMap[i]
	// }

	*dest = *equipMap
	dest[slot] = item
}

func (equipMap *SolvableEquipMap) Get(slot SlotEquip) *SolvableItem {
	return equipMap[slot]
}

func (equipMap *SolvableEquipMap) Has(slot SlotEquip) bool {
	return equipMap[slot] != nil
}

func (equipMap *SolvableEquipMap) GetAsId(slot SlotEquip) ItemId {
	item := equipMap[slot]
	if item != nil {
		return item.itemId
	} else {
		return 0
	}
}

func (equipMap *SolvableEquipMap) CountNonEmptySlots() int {
	count := 0
	for i := range *equipMap {
		if equipMap[i] != nil {
			count++
		}
	}
	return count
}

func (equipMap *SolvableEquipMap) AllItemSeq() iter.Seq[*SolvableItem] {
	return func(yield func(*SolvableItem) bool) {
		for _, item := range equipMap {
			if item != nil {
				if !yield(item) {
					return
				}
			}
		}
	}
}
