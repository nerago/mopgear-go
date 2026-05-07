package items

import (
	"iter"
)

type FullEquipMap [16]*FullItem

func (equipMap *FullEquipMap) Get(slot SlotEquip) *FullItem {
	return equipMap[slot]
}

func (equipMap *FullEquipMap) Has(slot SlotEquip) bool {
	return equipMap[slot] != nil
}

func (equipMap *FullEquipMap) GetAsId(slot SlotEquip) ItemId {
	item := equipMap[slot]
	if item != nil {
		return item.Ref.ItemId
	} else {
		return 0
	}
}

func (equipMap *FullEquipMap) GetGeneric(slot SlotEquip) IItem {
	return equipMap[slot]
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

func (equipMap *FullEquipMap) AllItemSeqGeneric() iter.Seq[IItem] {
	return func(yield func(IItem) bool) {
		for _, item := range equipMap {
			if item != nil {
				if !yield(item) {
					return
				}
			}
		}
	}
}

// //////////////////////////////////////////////////////
type SolvableEquipMap [16]*SolvableItem

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

func (equipMap SolvableEquipMap) GetGeneric(slot SlotEquip) IItem {
	return equipMap[slot]
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

func (equipMap SolvableEquipMap) AllItemSeqGeneric() iter.Seq[IItem] {
	return func(yield func(IItem) bool) {
		for _, item := range equipMap {
			if item != nil {
				if !yield(item) {
					return
				}
			}
		}
	}
}
