package items

import (
	"iter"
	"math/big"
	"paladin_gearing_go/types/common"
)

type FullOptionsMap [16][]FullItem

func (optionsMap FullOptionsMap) Get(slot common.SlotEquip) []FullItem {
	return optionsMap[slot]
}

func (optionsMap FullOptionsMap) Has(slot common.SlotEquip) bool {
	return len(optionsMap[slot]) > 0
}

func (optionsMap *FullOptionsMap) Put(slot common.SlotEquip, items []FullItem) {
	optionsMap[slot] = items
}

func (optionsMap *FullOptionsMap) MapSlots(mapper func([]FullItem) []FullItem) {
	for i := range optionsMap {
		optionsMap[i] = mapper(optionsMap[i])
	}
}

func (optionsMap *FullOptionsMap) AllItems() iter.Seq[FullItem] {
	return func(yield func(FullItem) bool) {
		for _, slotArray := range optionsMap {
			for _, item := range slotArray {
				if !yield(item) {
					return
				}
			}
		}
	}
}

type SolvableOptionsMap [16][]SolvableItem

func (optionsMap *SolvableOptionsMap) Get(slot common.SlotEquip) []SolvableItem {
	return optionsMap[slot]
}

func (optionsMap *SolvableOptionsMap) Has(slot common.SlotEquip) bool {
	return len(optionsMap[slot]) > 0
}

func (optionsMap *SolvableOptionsMap) TotalCombinationCount() *big.Int {
	total := big.NewInt(1)
	for _, slotArray := range optionsMap {
		slotSize := int64(len(slotArray))
		total.Mul(total, big.NewInt(slotSize))
	}
	return total
}

type SkinnyOptionsMap [16][]SkinnyItem

func (optionsMap *SkinnyOptionsMap) TotalCombinationCount() *big.Int {
	total := big.NewInt(1)
	for _, slotArray := range optionsMap {
		slotSize := int64(len(slotArray))
		total.Mul(total, big.NewInt(slotSize))
	}
	return total
}
