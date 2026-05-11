package items

import (
	"iter"
	"paladin_gearing_go/util"
)

type SolvableOptionsMap [16][]SolvableItem

func SolvableOptionsMap_of(fullMap *FullOptionsMap) SolvableOptionsMap {
	result := SolvableOptionsMap{}
	for slot := range fullMap {
		result[slot] = util.CastSliceAsNew(fullMap[slot], SolvableItem_Of)
	}
	return result
}

func (optionsMap *SolvableOptionsMap) Get(slot SlotEquip) []SolvableItem {
	return optionsMap[slot]
}

func (optionsMap *SolvableOptionsMap) Has(slot SlotEquip) bool {
	return len(optionsMap[slot]) > 0
}

func (optionsMap *SolvableOptionsMap) AllItemSeq() iter.Seq[*SolvableItem] {
	return func(yield func(*SolvableItem) bool) {
		for _, slotArray := range optionsMap {
			for _, item := range slotArray {
				if !yield(&item) {
					return
				}
			}
		}
	}
}

func (optionsMap *SolvableOptionsMap) AllItemSlotSeq() iter.Seq2[SlotEquip, *SolvableItem] {
	return func(yield func(SlotEquip, *SolvableItem) bool) {
		for slot := Equip_Iter_First; slot <= Equip_Iter_Last; slot++ {
			for i := range optionsMap[slot] {
				if !yield(slot, &optionsMap[slot][i]) {
					return
				}
			}
		}
	}
}
