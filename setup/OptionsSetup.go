package setup

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
)

func OptionsSetup_FromGearFile(filename string, model *model.Model, printer *util.PrintRecorder) items.FullOptionsMap {
	equipped := loaders.GearFileReader_Read(filename)
	return OptionsSetup_FromEquipped(equipped, model, printer)
}

func OptionsSetup_FromEquipped(equipped []loaders.EquippedItem, model *model.Model, printer *util.PrintRecorder) items.FullOptionsMap {
	optionMap := items.FullOptionsMap{}
	for _, equipItem := range equipped {
		optionList, baseItem := OptionsSetup_FromEquipped_Single(equipItem, model, printer)
		optionMap.FillSlot_ExpectedEmpty(baseItem.Slot, optionList)
	}
	return optionMap
}

func OptionsSetup_FromEquipped_Single(equipItem loaders.EquippedItem, model *model.Model, printer *util.PrintRecorder) ([]items.FullItem, *items.FullItem) {
	item := loadItemBasic(equipItem.ItemId, equipItem.UpgradeStep, printer)
	addDetailFromEquip(&item, equipItem)
	printer.Println(item.String())
	return tools.Reforger_AllOptions(&item, &model.ReforgeRules), &item
}

func OptionsSetup_FromIdOnlyUseAllDefaults(itemId uint32, upgradeLevel int16, model *model.Model, printer *util.PrintRecorder) ([]items.FullItem, *items.FullItem) {
	item := loadItemBasic(itemId, upgradeLevel, printer)
	addDetailUsingDefaults(&item, model)
	printer.Println(item.String())
	return tools.Reforger_AllOptions(&item, &model.ReforgeRules), &item
}

func OptionsSetup_ExactEquippedOnly(equipped []loaders.EquippedItem, model *model.Model, printer *util.PrintRecorder) items.FullEquipMap {
	resultMap := items.FullEquipMap{}
	for _, equipItem := range equipped {
		item := loadItemBasic(equipItem.ItemId, equipItem.UpgradeStep, printer)
		addDetailFromEquip(&item, equipItem)

		if equipItem.Reforging != 0 {
			reforge := db.WowSimDB_ReforgeById(equipItem.Reforging)
			item = *tools.Reforger_SinglePreset(&item, &reforge)
		}

		printer.Println(item.String())
		resultMap.FillSlot_ExpectedEmpty(item.Slot, &item)
	}
	return resultMap
}

func loadItemBasic(itemId uint32, upgradeLevel int16, printer *util.PrintRecorder) items.FullItem {
	storedItem := db.WowSimDB_ByIdAndUpgrade(itemId, upgradeLevel)
	if storedItem == nil && upgradeLevel > 0 {
		storedItem = db.WowSimDB_ByIdAndUpgrade(itemId, 0)
		printer.Printf("NOT FOUND at specified upgrade %d = %s\n", upgradeLevel, storedItem)
	}
	return *storedItem
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

func addDetailFromEquip(item *items.FullItem, equipItem loaders.EquippedItem) {
	if equipItem.RandomSuffix == -336 {
		stat := stats.Stat_Crit
		amount := itemLevelToRandomAmount[item.Ref.ItemLevel]
		item.ChangeBaseStats(item.StatBase.WithChange(stat, amount))
		item.RandomSuffix = equipItem.RandomSuffix
	} else if equipItem.RandomSuffix != 0 {
		panic("unknown random suffix")
	}

	calcGemsAndEnchants(item, equipItem)
}

func calcGemsAndEnchants(item *items.FullItem, equipItem loaders.EquippedItem) {
	// TODO trinket modelling
	if item.Slot == items.Item_Trinket {
		return
	}

	total := stats.StatBlock{}

	if equipItem.EnchantChoice != 0 {
		// TODO enchant validation
		enchantValue := db.EnchantData_ById(equipItem.EnchantChoice)
		total.Increment_Mutating(&enchantValue.Stats)
		item.EnchantChoice = equipItem.EnchantChoice
	}

	// TODO blacksmith status in params
	if item.Slot.PossibleBlacksmith() {
		item.SocketSlots = append(item.SocketSlots, stats.Socket_General)
	}

	socketBonusMet := true
	gemChoice := make([]stats.GemInfo, 0)
	for index, gemId := range equipItem.GemChoice {
		gemInfo := db.GemData_ById(gemId)
		gemChoice = append(gemChoice, gemInfo)
		total.Increment_Mutating(&gemInfo.Stats)

		socket := item.SocketSlots[index]
		if !socket.SocketMatch(&gemInfo.Stats) {
			socketBonusMet = false
		}
	}
	item.GemChoice = gemChoice

	if socketBonusMet {
		total.Increment_Mutating(&item.SocketBonus)
	}

	item.ChangeEnchantStats(total)
}

func addDetailUsingDefaults(item *items.FullItem, model *model.Model) {
	// TODO known random suffixes?

	// TODO trinket modelling
	if item.Slot == items.Item_Trinket {
		return
	}

	total := stats.StatBlock{}

	enchantChoice := model.EnchantChoice.GetChoice(item.Slot)
	if enchantChoice != nil {
		total.Increment_Mutating(&enchantChoice.Stats)
		item.EnchantChoice = enchantChoice.Id
	}

	// TODO blacksmith status in params
	if item.Slot.PossibleBlacksmith() {
		item.SocketSlots = append(item.SocketSlots, stats.Socket_General)
	}

	socketBonusMet := true
	gemChoice := make([]stats.GemInfo, 0)
	for _, socketType := range item.SocketSlots {
		// NOTE unique engineering gems not checked
		gemInfo := model.GemChoice.GetChoice(socketType)
		gemChoice = append(gemChoice, *gemInfo)
		total.Increment_Mutating(&gemInfo.Stats)

		if !socketType.SocketMatch(&gemInfo.Stats) {
			socketBonusMet = false
		}
	}
	item.GemChoice = gemChoice

	if socketBonusMet {
		total.Increment_Mutating(&item.SocketBonus)
	}

	item.ChangeEnchantStats(total)
}
