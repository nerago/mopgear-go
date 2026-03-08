package items

import "iter"

type FullEquipMap [16]*FullItem

func (equipMap *FullEquipMap) Get(slot SlotEquip) *FullItem {
	return equipMap[slot]
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

// //////////////////////////////////////////////////////
type SolvableEquipMap [16]*SolvableItem

func (equipMap SolvableEquipMap) WithAdditional(slot SlotEquip, item *SolvableItem) SolvableEquipMap {
	var result SolvableEquipMap = equipMap
	result[slot] = item
	return result
}

func (equipMap SolvableEquipMap) AllItemSeq() iter.Seq[*SolvableItem] {
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

// //////////////////////////////////////////////////////
type SkinnyEquipMap [16]SkinnyItem
