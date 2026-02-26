package items

import (
	"iter"
	"math/big"
	. "paladin_gearing_go/stats"
)

type FullOptionsMap [16][]FullItem

func (optionsMap *FullOptionsMap) Get(slot SlotEquip) []FullItem {
	return optionsMap[slot]
}

func (optionsMap *FullOptionsMap) Has(slot SlotEquip) bool {
	return len(optionsMap[slot]) > 0
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

func (optionsMap *FullOptionsMap) FillSlot_ExpectedEmpty(slotItem SlotItem, optionList []FullItem) {
	var slotEquip SlotEquip
	switch slotItem {
	case Item_Back:
		slotEquip = Equip_Back
	case Item_Belt:
		slotEquip = Equip_Belt
	case Item_Chest:
		slotEquip = Equip_Chest
	case Item_Foot:
		slotEquip = Equip_Foot
	case Item_Hand:
		slotEquip = Equip_Hand
	case Item_Head:
		slotEquip = Equip_Head
	case Item_Leg:
		slotEquip = Equip_Leg
	case Item_Neck:
		slotEquip = Equip_Neck
	case Item_Offhand:
		slotEquip = Equip_Offhand
	case Item_Shoulder:
		slotEquip = Equip_Shoulder
	case Item_Wrist:
		slotEquip = Equip_Wrist
	case Item_Weapon1H:
		slotEquip = Equip_Weapon
	case Item_Weapon2H:
		slotEquip = Equip_Weapon

	case Item_Ring:
		if optionsMap[Equip_Ring1] == nil {
			slotEquip = Equip_Ring1
		} else {
			slotEquip = Equip_Ring2
		}

	case Item_Trinket:
		if optionsMap[Equip_Trinket1] == nil {
			slotEquip = Equip_Trinket1
		} else {
			slotEquip = Equip_Trinket2
		}

	default:
		panic("unexpected SlotItem")
	}

	if optionsMap[slotEquip] == nil {
		optionsMap[slotEquip] = optionList
	} else {
		panic("duplicate item")
	}
}

type SolvableOptionsMap [16][]SolvableItem

func (optionsMap *SolvableOptionsMap) Get(slot SlotEquip) []SolvableItem {
	return optionsMap[slot]
}

func (optionsMap *SolvableOptionsMap) Has(slot SlotEquip) bool {
	return len(optionsMap[slot]) > 0
}

func (optionsMap *SolvableOptionsMap) TotalCombinationCount() *big.Int {
	valueCount := 0
	total := big.NewInt(1)
	for _, slotArray := range optionsMap {
		slotSize := int64(len(slotArray))
		if slotSize > 0 {
			total.Mul(total, big.NewInt(slotSize))
			valueCount++
		}
	}
	if valueCount == 0 {
		panic("empty options")
	}
	return total
}

func (optionsMap *SolvableOptionsMap) AllItemSeq() iter.Seq[*SolvableItem] {
	return func(yield func(*SolvableItem) bool) {
		for _, slotArray := range optionsMap {
			for _, item := range slotArray {
				if !item.IsEmpty() {
					if !yield(&item) {
						return
					}
				}
			}
		}
	}
}

type SkinnyOptionsMap [16][]SkinnyItem

func (optionsMap *SkinnyOptionsMap) TotalCombinationCount() *big.Int {
	valueCount := 0
	total := big.NewInt(1)
	for _, slotArray := range optionsMap {
		slotSize := int64(len(slotArray))
		if slotSize > 0 {
			total.Mul(total, big.NewInt(slotSize))
			valueCount++
		}
	}
	if valueCount == 0 {
		panic("empty options")
	}
	return total
}
