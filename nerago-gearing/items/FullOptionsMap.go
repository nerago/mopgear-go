package items

import (
	"iter"
	"paladin_gearing_go/util"
	"slices"
)

type FullOptionsMap [ITEM_SLOT_COUNT][]FullItem

func (optionsMap *FullOptionsMap) Get(slot SlotEquip) []FullItem {
	return optionsMap[slot]
}

func (optionsMap *FullOptionsMap) Has(slot SlotEquip) bool {
	return len(optionsMap[slot]) > 0
}

func (optionsMap *FullOptionsMap) IncludesItemId(itemId ItemId) bool {
	for slot := range optionsMap {
		for i := range optionsMap[slot] {
			if optionsMap[slot][i].ItemId() == itemId {
				return true
			}
		}
	}
	return false
}

func (optionsMap *FullOptionsMap) IncludesItemIdInSlot(itemId ItemId, slot SlotEquip) bool {
	for i := range optionsMap[slot] {
		if optionsMap[slot][i].ItemId() == itemId {
			return true
		}
	}
	return false
}

func (optionsMap *FullOptionsMap) IncludesItemNameInSlot(itemName string, slot SlotEquip) bool {
	for i := range optionsMap[slot] {
		if optionsMap[slot][i].BaseName() == itemName {
			return true
		}
	}
	return false
}

func (optionsMap *FullOptionsMap) FindItemId(itemId ItemId) iter.Seq[FullItem] {
	return func(yield func(FullItem) bool) {
		for slot := range optionsMap {
			for i := range optionsMap[slot] {
				item := &optionsMap[slot][i]
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
	for slot := range optionsMap {
		for i := range optionsMap[slot] {
			item := &optionsMap[slot][i]
			if item.ItemId() == itemId {
				return item
			}
		}
	}
	panic("no such item")
}

func (optionsMap *FullOptionsMap) FindItemIdFirstOptional(itemId ItemId) (*FullItem, bool) {
	for slot := range optionsMap {
		for i := range optionsMap[slot] {
			item := &optionsMap[slot][i]
			if item.ItemId() == itemId {
				return item, true
			}
		}
	}
	return nil, false
}

func (optionsMap *FullOptionsMap) FindItemIdSlotUnique(itemId ItemId) SlotEquip {
	var slotFound SlotEquip
	found := false
	for slot := range optionsMap {
		for i := range optionsMap[slot] {
			item := &optionsMap[slot][i]
			if item.ItemId() == itemId {
				if !found {
					found = true
					slotFound = SlotEquip(slot)
				} else if slotFound != SlotEquip(slot) {
					panic("duplicate slot for item")
				}
			}
		}
	}
	return slotFound
}

func (optionsMap *FullOptionsMap) SlotItemSeq(slotEquip SlotEquip) iter.Seq[*FullItem] {
	return func(yield func(*FullItem) bool) {
		for i := range optionsMap[slotEquip] {
			if !yield(&optionsMap[slotEquip][i]) {
				return
			}
		}
	}
}

func (optionsMap *FullOptionsMap) AllItems() iter.Seq[*FullItem] {
	return func(yield func(*FullItem) bool) {
		for slot := range optionsMap {
			for i := range optionsMap[slot] {
				item := &optionsMap[slot][i]
				if !yield(item) {
					return
				}
			}
		}
	}
}

func (optionsMap *FullOptionsMap) AllItemsWithSlot() iter.Seq2[SlotEquip, *FullItem] {
	return func(yield func(SlotEquip, *FullItem) bool) {
		for slot := range optionsMap {
			for i := range optionsMap[slot] {
				if !yield(SlotEquip(slot), &optionsMap[slot][i]) {
					return
				}
			}
		}
	}
}

func (optionsMap *FullOptionsMap) AddOneOption(item FullItem) {
	for _, slotEquip := range item.slot.ToSlotEquipOptions() {
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

func (optionsMap *FullOptionsMap) AddSeveralOptionsSpecific_WhereNotExist(slotEquip SlotEquip, options []FullItem) {
	for _, newItem := range options {
		if !util.ContainsFunc_Pointer(optionsMap[slotEquip], newItem.Equals) {
			optionsMap[slotEquip] = append(optionsMap[slotEquip], newItem)
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

func (optionsMap *FullOptionsMap) MapEachItem(mapper func(*FullItem) FullItem) {
	for i := range optionsMap {
		optionsMap[i] = util.MapSliceAsNew(optionsMap[i], mapper)
	}
}

func (optionsMap *FullOptionsMap) FilterAllItems(filter func(*FullItem) bool) {
	for i := range optionsMap {
		if len(optionsMap[i]) > 0 {
			optionsMap[i] = util.FilterSliceAsNew(optionsMap[i], filter)
			if len(optionsMap[i]) == 0 {
				panic("removing items leaves slot empty")
			}
		}
	}
}

func (optionsMap *FullOptionsMap) FilterSlot(slot SlotEquip, filter func(*FullItem) bool) {
	if len(optionsMap[slot]) > 0 {
		optionsMap[slot] = util.FilterSliceAsNew(optionsMap[slot], filter)
		if len(optionsMap[slot]) == 0 {
			panic("removing items leaves slot empty " + slot.Name())
		}
	}
}

func (optionsMap *FullOptionsMap) FilterSlotNoValidate(slot SlotEquip, filter func(*FullItem) bool) {
	if len(optionsMap[slot]) > 0 {
		optionsMap[slot] = util.FilterSliceAsNew(optionsMap[slot], filter)
	}
}

func (optionsMap *FullOptionsMap) RemoveItemIdFromAll(itemId ItemId) {
	for slot := range optionsMap {
		if len(optionsMap[slot]) > 0 {
			optionsMap[slot] = util.FilterSliceAsNew(optionsMap[slot], func(x *FullItem) bool { return x.ItemId() != itemId })
			if len(optionsMap[slot]) == 0 {
				panic("removing items leaves slot empty")
			}
		}
	}
}

func (optionsMap *FullOptionsMap) RemoveItemIdFromSlot(slot SlotEquip, itemId ItemId) {
	if len(optionsMap[slot]) > 0 {
		optionsMap[slot] = util.FilterSliceAsNew(optionsMap[slot], func(x *FullItem) bool { return x.ItemId() != itemId })
		if len(optionsMap[slot]) == 0 {
			panic("removing items leaves slot empty " + slot.Name())
		}
	}
}

func (optionsMap *FullOptionsMap) ForceSlotOnlySpecifiedItemId(slot SlotEquip, itemId ItemId) {
	if len(optionsMap[slot]) > 0 {
		optionsMap[slot] = util.FilterSliceAsNew(optionsMap[slot], func(x *FullItem) bool { return x.ItemId() == itemId })
		if len(optionsMap[slot]) == 0 {
			panic("removing items leaves slot empty " + slot.Name())
		}
	}
}

func (optionsMap *FullOptionsMap) RemoveDuplicates() {
	for slot := range optionsMap {
		if len(optionsMap[slot]) > 0 {
			optionsMap[slot] = util.RemoveDuplicatesFunc_NewIfChanged(optionsMap[slot], (*FullItem).Equals)
		}
	}
}

func (optionsMap *FullOptionsMap) Clone() FullOptionsMap {
	result := FullOptionsMap{}
	for slot := range optionsMap {
		result[slot] = slices.Clone(optionsMap[slot])
	}
	return result
}
