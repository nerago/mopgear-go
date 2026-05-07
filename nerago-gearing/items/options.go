package items

import (
	"iter"
	"math/big"
	"paladin_gearing_go/util"
	"slices"
)

type FullOptionsMap [16][]FullItem

func (optionsMap *FullOptionsMap) Get(slot SlotEquip) []FullItem {
	return optionsMap[slot]
}

func (optionsMap *FullOptionsMap) Has(slot SlotEquip) bool {
	return len(optionsMap[slot]) > 0
}

func (optionsMap *FullOptionsMap) IncludesItemId(itemId ItemId) bool {
	for _, slotArray := range optionsMap {
		for _, item := range slotArray {
			if item.ItemId() == itemId {
				return true
			}
		}
	}
	return false
}

func (optionsMap *FullOptionsMap) IncludesItemIdInSlot(itemId ItemId, slot SlotEquip) bool {
	for _, item := range optionsMap[slot] {
		if item.ItemId() == itemId {
			return true
		}
	}
	return false
}

func (optionsMap *FullOptionsMap) SlotGroupedByItemId(slot SlotEquip) map[ItemId][]*FullItem {
	grouped := make(map[ItemId][]*FullItem)
	for _, item := range optionsMap[slot] {
		grouped[item.ItemId()] = append(grouped[item.ItemId()], &item)
	}
	return grouped
}

func (optionsMap *FullOptionsMap) IncludesItemNameInSlot(itemName string, slot SlotEquip) bool {
	for _, item := range optionsMap[slot] {
		if item.BaseName == itemName {
			return true
		}
	}
	return false
}

func (optionsMap *FullOptionsMap) MapSlotsAll(mapper func([]FullItem) []FullItem) {
	for i := range optionsMap {
		optionsMap[i] = mapper(optionsMap[i])
	}
}

func (optionsMap *FullOptionsMap) MapSlot(slot SlotEquip, mapper func([]FullItem) []FullItem) {
	optionsMap[slot] = mapper(optionsMap[slot])
}

func (optionsMap *FullOptionsMap) MapEachItem(mapper func(*FullItem) FullItem) {
	for i := range optionsMap {
		optionsMap[i] = util.MapSliceAsNew(optionsMap[i], mapper)
	}
}

func (optionsMap *FullOptionsMap) FilterSlotsAll(filter func(*FullItem) bool) {
	for i := range optionsMap {
		optionsMap[i] = util.FilterSliceAsNew(optionsMap[i], filter)
	}
}

func (optionsMap *FullOptionsMap) FilterSlot(slot SlotEquip, filter func(*FullItem) bool) {
	optionsMap[slot] = util.FilterSliceAsNew(optionsMap[slot], filter)
}

func (optionsMap *FullOptionsMap) FindItemId(itemId ItemId) iter.Seq[FullItem] {
	return func(yield func(FullItem) bool) {
		for _, slotArray := range optionsMap {
			for i := range slotArray {
				item := &slotArray[i]
				if item.ItemId() == itemId {
					if !yield(*item) {
						return
					}
				}
			}
		}
	}
}

func (optionsMap *FullOptionsMap) FindItemIdFirst(itemId ItemId) *FullItem {
	for _, slotArray := range optionsMap {
		for i := range slotArray {
			item := &slotArray[i]
			if item.ItemId() == itemId {
				return item
			}
		}
	}
	panic("no such item")
}

func (optionsMap *FullOptionsMap) FindItemIdFirstOptional(itemId ItemId) (*FullItem, bool) {
	for _, slotArray := range optionsMap {
		for i := range slotArray {
			item := &slotArray[i]
			if item.ItemId() == itemId {
				return item, true
			}
		}
	}
	return nil, false
}

func (optionsMap *FullOptionsMap) AllItems() iter.Seq[*FullItem] {
	return func(yield func(*FullItem) bool) {
		for _, slotArray := range optionsMap {
			for i := range slotArray {
				item := &slotArray[i]
				if !yield(item) {
					return
				}
			}
		}
	}
}

func (optionsMap *FullOptionsMap) AllItemsWithSlot() iter.Seq2[SlotEquip, *FullItem] {
	return func(yield func(SlotEquip, *FullItem) bool) {
		for slot, slotArray := range optionsMap {
			for _, item := range slotArray {
				if !yield(SlotEquip(slot), &item) {
					return
				}
			}
		}
	}
}

func (optionsMap *FullOptionsMap) AddOneOption(item FullItem) {
	for _, slotEquip := range item.Slot.ToSlotEquipOptions() {
		optionsMap[slotEquip] = append(optionsMap[slotEquip], item)
	}
}

func (optionsMap *FullOptionsMap) AddSeveralOptions(slot SlotItem, options []FullItem) {
	for _, slotEquip := range slot.ToSlotEquipOptions() {
		optionsMap[slotEquip] = append(optionsMap[slotEquip], options...)
	}
}

func (optionsMap *FullOptionsMap) AddSeveralOptionsSpecific(slotEquip SlotEquip, options []FullItem) {
	optionsMap[slotEquip] = append(optionsMap[slotEquip], options...)
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

func (equipMap *FullEquipMap) FillSlot_ExpectedEmpty(slotItem SlotItem, item *FullItem) {
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
		if equipMap[Equip_Ring1] == nil {
			slotEquip = Equip_Ring1
		} else {
			slotEquip = Equip_Ring2
		}

	case Item_Trinket:
		if equipMap[Equip_Trinket1] == nil {
			slotEquip = Equip_Trinket1
		} else {
			slotEquip = Equip_Trinket2
		}

	default:
		panic("unexpected SlotItem")
	}

	if equipMap[slotEquip] == nil {
		equipMap[slotEquip] = item
	} else {
		panic("duplicate item")
	}
}

func (optionsMap *FullOptionsMap) Clone() FullOptionsMap {
	result := FullOptionsMap{}
	for slot, content := range optionsMap {
		result[slot] = slices.Clone(content)
	}
	return result
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

func (optionsMap *SolvableOptionsMap) TotalCombinationCountAsInt() uint64 {
	combo := optionsMap.TotalCombinationCount()
	if !combo.IsUint64() {
		panic("too big for 64 bit int")
	}
	return combo.Uint64()
}

func (optionsMap *SolvableOptionsMap) TotalItemCount() int {
	itemCount := 0
	for _, slotArray := range optionsMap {
		itemCount += len(slotArray)
	}
	return itemCount
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
