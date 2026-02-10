package setup

import (
	. "paladin_gearing_go/db"
	. "paladin_gearing_go/loaders"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/process"
	. "paladin_gearing_go/types/common"
	. "paladin_gearing_go/types/items"
	. "paladin_gearing_go/types/stats"
	. "paladin_gearing_go/util"
)

func OptionsSetup_FromGearFile(filename string, model *Model, printer *PrintRecorder) FullOptionsMap {
	equipped := GearFileReader_Read(filename)
	return convertToOptions(equipped, model, printer)
}

func convertToOptions(equipped []EquippedItem, model *Model, printer *PrintRecorder) FullOptionsMap {
	optionMap := FullOptionsMap{}
	for _, item := range equipped {
		baseItem := loadItem(item)
		printer.Println(baseItem.String())
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

func loadItem(equipItem EquippedItem) *FullItem {
	storedItem := WowSimDB_ByIdAndUpgrade(equipItem.ItemId, equipItem.UpgradeStep)
	if storedItem == nil && equipItem.UpgradeStep > 0 {
		storedItem = WowSimDB_ByIdAndUpgrade(equipItem.ItemId, 0)
	}

	item := addDetails(storedItem, equipItem)
	return item
}

var itemLevelToRandomAmount = makeItemLevelToRandomAmount()

func makeItemLevelToRandomAmount() map[uint16]uint32 {
	lookup := make(map[uint16]uint32)
	lookup[502] = 712
	lookup[522] = 858
	lookup[528] = 907
	lookup[535] = 968
	lookup[541] = 1019
	return lookup
}

func addDetails(item *FullItem, equipItem EquippedItem) *FullItem {
	if equipItem.RandomSuffix == -336 {
		stat := Stat_Crit
		amount := itemLevelToRandomAmount[item.Ref.ItemLevel]
		item = item.ChangedBaseStats(item.StatBase.WithChange(stat, amount))
		item.RandomSuffix = equipItem.RandomSuffix
	} else if equipItem.RandomSuffix != 0 {
		panic("unknown random suffix")
	}

	if item.Slot != Item_Trinket {
		enchantStats := calcGemsAndEnchants(item, equipItem)
		item = item.ChangedEnchantStats(enchantStats)
	}
	item.GemChoice = equipItem.GemChoice
	item.EnchantChoice = equipItem.EnchantChoice

	// TODO trinket model

	return item
}

// TODO enchant validation
func calcGemsAndEnchants(item *FullItem, equipItem EquippedItem) StatBlock {
	stats := StatBlock{}

	if equipItem.EnchantChoice != 0 {
		enchantValue := EnchantData_ById(equipItem.EnchantChoice)
		stats.Increment_Mutating(&enchantValue.Stats)
	}

	if item.Slot.PossibleBlacksmith() {
		item.SocketSlots = append(item.SocketSlots, Socket_General)
	}

	socketBonusMet := true
	for index, gemId := range equipItem.GemChoice {
		gemInfo := GemData_ById(gemId)
		stats.Increment_Mutating(&gemInfo.Stats)

		socket := item.SocketSlots[index]
		if !socketMatch(socket, &gemInfo.Stats) {
			socketBonusMet = false
		}
	}

	if socketBonusMet {
		stats.Increment_Mutating(&item.SocketBonus)
	}

	return stats
}

func socketMatch(socket SocketType, gemStat *StatBlock) bool {
	switch socket {
	case Socket_Red:
		return gemStat[Stat_Agility] != 0 || gemStat[Stat_Strength] != 0 || gemStat[Stat_Intellect] != 0 || gemStat[Stat_Expertise] != 0
	case Socket_Yellow:
		return gemStat[Stat_Crit] != 0 || gemStat[Stat_Haste] != 0 || gemStat[Stat_Mastery] != 0
	case Socket_Blue:
		return gemStat[Stat_Hit] != 0 || gemStat[Stat_Spirit] != 0 || gemStat[Stat_Stamina] != 0
	case Socket_General, Socket_Meta, Socket_Engineering, Socket_Sha:
		return true
	default:
		panic("unexpected common.SocketType")
	}
}
