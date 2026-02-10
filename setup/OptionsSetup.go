package setup

import (
	. "paladin_gearing_go/loaders"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/process"
	. "paladin_gearing_go/types/common"
	. "paladin_gearing_go/types/items"
)

func OptionsSetup_FromGearFile(filename string, model *Model) FullOptionsMap {
	equipped := GearFileReader_Read(filename)
	return convertToOptions(equipped, model)
}

func convertToOptions(equipped []EquippedItem, model *Model) FullOptionsMap {
	optionMap := FullOptionsMap{}
	for _, item := range equipped {
		baseItem := loadItem(item, model)
		optionList := Reforger_allOptions(baseItem, &model.ReforgeRules)
		addToSlot(&optionMap, baseItem.Slot, optionList)
	}
	return optionMap
}

func addToSlot(optionMap *FullOptionsMap, slotItem SlotItem, optionList []FullItem) {
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
		if optionMap[Equip_Ring1] == nil {
			slotEquip = Equip_Ring1
		} else {
			slotEquip = Equip_Ring2
		}

	case Item_Trinket:
		if optionMap[Equip_Trinket1] == nil {
			slotEquip = Equip_Trinket1
		} else {
			slotEquip = Equip_Trinket2
		}

	default:
		panic("unexpected SlotItem")
	}

	if optionMap[slotEquip] == nil {
		optionMap[slotEquip] = optionList
	} else {
		panic("duplicate item")
	}
}

func loadItem(item EquippedItem, model *Model) *FullItem {
	storedItem := WowSimDB_ByIdAndUpgrade(item.ItemId, item.UpgradeStep)
	if storedItem == nil && item.UpgradeStep > 0 {
		storedItem = WowSimDB_ByIdAndUpgrade(item.ItemId, 0)
	}

	panic("unimplemented")

	// TODO apply gems
	// TODO enchant validation
}
