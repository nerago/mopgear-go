package items

import (
	"paladin_gearing_go/util"
)

type CanUpgradeResult int8

const (
	CanUpgrade_Yes              CanUpgradeResult = iota
	CanUpgrade_Equipped         CanUpgradeResult = iota
	CanUpgrade_Equipped_Similar CanUpgradeResult = iota
	CanUpgrade_AvailableInBags  CanUpgradeResult = iota // bags result cannot be checked in this file but other callers may use. TODO do better
	CanUpgrade_InvalidAlways    CanUpgradeResult = iota
)

func (can CanUpgradeResult) TextLong() string {
	switch can {
	case CanUpgrade_Equipped:
		return "equipped"
	case CanUpgrade_Equipped_Similar:
		return "equipped similar"
	case CanUpgrade_AvailableInBags:
		return "available in bags"
	case CanUpgrade_InvalidAlways:
		return "invalid"
	default:
		return ""
	}
}

func (can CanUpgradeResult) TextAbbrev() string {
	switch can {
	case CanUpgrade_Equipped:
		return "E"
	case CanUpgrade_Equipped_Similar:
		return "e"
	case CanUpgrade_AvailableInBags:
		return "b"
	case CanUpgrade_InvalidAlways:
		return "X"
	default:
		return ""
	}
}

func (optionsMap *FullOptionsMap) CouldAddUpgrade_ItemSlot(slot SlotItem, extra *FullItem, printer *util.PrintRecorder) CanUpgradeResult {
	result := CanUpgrade_InvalidAlways
	for _, slotEquip := range slot.ToSlotEquipOptions() {
		result = optionsMap.CouldAddUpgrade_EquipSlot(slotEquip, extra, printer)
		if result == CanUpgrade_Yes {
			return result
		}
	}
	return result
}

func (optionsMap *FullOptionsMap) CouldAddUpgrade_EquipSlot(slot SlotEquip, extra *FullItem, printer *util.PrintRecorder) CanUpgradeResult {
	if !optionsMap.Has(slot) {
		printer.Println("SLOT NOT USED IN CURRENT SET " + extra.CreateString())
		return CanUpgrade_InvalidAlways
	}

	if slot == Equip_Weapon {
		currentWeapon := optionsMap.Get(Equip_Weapon)[0]
		if extra.Slot != currentWeapon.Slot {
			printer.Println("WRONG WEAPON TYPE " + extra.CreateString())
			return CanUpgrade_InvalidAlways
		}
	}

	if slot == Equip_Offhand {
		currentWeapon := optionsMap.Get(Equip_Weapon)[0]
		if currentWeapon.Slot == Item_Weapon2H {
			printer.Println("INVALID OFFHAND WITH 2H WEAPON " + extra.CreateString())
			return CanUpgrade_InvalidAlways
		}
	}

	if optionsMap.IncludesItemIdInSlot(extra.ItemId(), slot) {
		printer.Println("SAME ITEM " + extra.CreateString())
		return CanUpgrade_Equipped
	}

	paired := slot.PairedSlot()
	if paired != -1 && optionsMap.IncludesItemIdInSlot(extra.ItemId(), paired) {
		printer.Println("SAME ITEM ID IN OTHER SLOT " + extra.CreateString())
		return CanUpgrade_Equipped
	} else if paired != -1 && optionsMap.IncludesUniqueEquippedViolationInSlot(extra.BaseName, paired) {
		printer.Println("RELATED ITEM NAME IN OTHER SLOT (unique equipped) " + extra.CreateString())
		return CanUpgrade_Equipped_Similar
	}

	return CanUpgrade_Yes
}
