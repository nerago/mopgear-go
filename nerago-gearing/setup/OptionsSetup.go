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

type MissingEnchantMode int8

const (
	MissingEnchant_Panic  MissingEnchantMode = iota
	MissingEnchant_Fix    MissingEnchantMode = iota
	MissingEnchant_Ignore MissingEnchantMode = iota
)

func OptionsSetup_FromGearFile(filename string, model *model.Model, missingEnchant MissingEnchantMode, printer *util.PrintRecorder) items.FullOptionsMap {
	equipped := loaders.GearFileReader_Read(filename)
	return OptionsSetup_FromEquipped(equipped, model, missingEnchant, printer)
}

func OptionsSetup_FromEquipped(equipped []loaders.EquippedItem, model *model.Model, missingEnchant MissingEnchantMode, printer *util.PrintRecorder) items.FullOptionsMap {
	optionMap := items.FullOptionsMap{}
	for _, equipItem := range equipped {
		optionList, baseItem := OptionsSetup_Single_FromEquipped(equipItem, model, missingEnchant, printer)
		optionMap.FillSlot_ExpectedEmpty(baseItem.SlotItem(), optionList)
	}
	return optionMap
}

func OptionsSetup_Single_FromEquipped(equipItem loaders.EquippedItem, model *model.Model, missingEnchant MissingEnchantMode, printer *util.PrintRecorder) ([]items.FullItem, *items.FullItem) {
	item := *db.WowSimDB_ByIdAndUpgrade_AllowFallback(equipItem.ItemId, equipItem.UpgradeStepOrItemLevel, printer)
	item = addDetailFromEquip(item, equipItem, model, missingEnchant, printer)
	return tools.Reforger_AllOptions(&item, &model.ReforgeRules), &item
}

func OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId items.ItemId, upgradeLevel items.UpgradeLevel, randomSuffix items.RandomSuffix, model *model.Model, printer *util.PrintRecorder) ([]items.FullItem, *items.FullItem) {
	item := *db.WowSimDB_ByIdAndUpgrade_AllowFallback(itemId, int32(upgradeLevel), printer)
	item = addDetailUsingDefaults(item, randomSuffix, model)
	return tools.Reforger_AllOptions(&item, &model.ReforgeRules), &item
}

func OptionsSetup_ExactEquippedOnly(equipped []loaders.EquippedItem, model *model.Model, missingEnchant MissingEnchantMode, printer *util.PrintRecorder) items.FullEquipMap {
	resultMap := items.FullEquipMap{}
	for _, equipItem := range equipped {
		item := OptionsSetup_ExactEquippedOnly_Item(equipItem, missingEnchant, model, printer)

		printer.Println(item.CreateString())
		resultMap.FillSlot_ExpectedEmpty(item.SlotItem(), &item)
	}
	return resultMap
}

func OptionsSetup_ExactEquippedOnly_Item(equipItem loaders.EquippedItem, missingEnchant MissingEnchantMode, model *model.Model, printer *util.PrintRecorder) items.FullItem {
	item := *db.WowSimDB_ByIdAndUpgrade_AllowFallback(equipItem.ItemId, equipItem.UpgradeStepOrItemLevel, printer)
	item = addDetailFromEquip(item, equipItem, model, missingEnchant, printer)

	if equipItem.Reforging != 0 {
		reforge := db.WowSimDB_ReforgeById(equipItem.Reforging)
		item = *tools.Reforger_SinglePreset(&item, &reforge)
	}

	return item
}

func addDetailUsingDefaults(item items.FullItem, randomSuffix items.RandomSuffix, model *model.Model) items.FullItem {
	item = *item.MakeItemWithRandomSuffix(randomSuffix)

	if item.SlotItem() == items.Item_Trinket {
		return item
	}

	var enchantChoice uint32 = 0
	statEnchant := stats.StatBlock{}

	enchantInfo := model.EnchantChoice.GetChoice(item.SlotItem())
	if enchantInfo != nil {
		stats.StatBlock_Increment_Mutating(&statEnchant, &enchantInfo.Stats)
		enchantChoice = enchantInfo.Id
	}

	socketSlots := confirmSocketSlots(item, model.Professions)

	gemChoice := addDefaultGems(&statEnchant, socketSlots, item.SocketBonus(), model)

	return *item.NewWithEnchantDetails(socketSlots, gemChoice, enchantChoice, statEnchant)
}

func addDetailFromEquip(item items.FullItem, equipItem loaders.EquippedItem, model *model.Model, missingEnchant MissingEnchantMode, printer *util.PrintRecorder) items.FullItem {
	item = *item.MakeItemWithRandomSuffix(equipItem.RandomSuffix)

	if item.SlotItem() == items.Item_Trinket {
		return item
	}

	var enchantChoice uint32 = 0
	statEnchant := stats.StatBlock{}

	if equipItem.EnchantChoice != 0 {
		enchantInfo := db.EnchantData_ById(equipItem.EnchantChoice)
		stats.StatBlock_Increment_Mutating(&statEnchant, &enchantInfo.Stats)
		enchantChoice = equipItem.EnchantChoice
	} else if item.SlotItem().CanEnchant() && missingEnchant == MissingEnchant_Panic {
		panic("missing enchant on " + item.CreateString())
	} else if enchantInfo := model.EnchantChoice.GetChoice(item.SlotItem()); enchantInfo != nil && missingEnchant == MissingEnchant_Fix {
		printer.Println("Add missing enchant " + item.CreateString())
		stats.StatBlock_Increment_Mutating(&statEnchant, &enchantInfo.Stats)
		enchantChoice = enchantInfo.Id
	} else if item.SlotItem().CanEnchant() && missingEnchant == MissingEnchant_Fix {
		panic("no enchant choice for " + item.SlotItem().Name())
	}

	socketSlots := confirmSocketSlots(item, model.Professions)

	var gemChoice []stats.GemInfo
	if len(equipItem.GemChoice) == len(socketSlots) {
		gemChoice = addSpecifiedGems(&statEnchant, socketSlots, item.SocketBonus(), equipItem)
	} else if len(equipItem.GemChoice) == 0 {
		switch missingEnchant {
		case MissingEnchant_Fix:
			printer.Println("Add missing gems " + item.CreateString())
			gemChoice = addDefaultGems(&statEnchant, socketSlots, item.SocketBonus(), model)
		case MissingEnchant_Panic:
			panic("missing gems on " + item.CreateString())
		}
	} else {
		panic("mismatch in gem array lengths on " + item.CreateString())
	}

	return *item.NewWithEnchantDetails(socketSlots, gemChoice, enchantChoice, statEnchant)
}

func addDefaultGems(statEnchant *stats.StatBlock, socketSlots []stats.SocketType, socketBonus *stats.StatBlock, model *model.Model) []stats.GemInfo {
	socketBonusMet := true
	gemChoice := make([]stats.GemInfo, 0, len(socketSlots))

	for _, socketType := range socketSlots {
		// NOTE unique engineering gems not checked
		gemInfo := model.GemChoice.GetChoice(socketType)
		gemChoice = append(gemChoice, *gemInfo)
		stats.StatBlock_Increment_Mutating(statEnchant, &gemInfo.Stats)

		if !socketType.SocketMatch(&gemInfo.Stats) {
			socketBonusMet = false
		}
	}

	if socketBonusMet {
		stats.StatBlock_Increment_Mutating(statEnchant, socketBonus)
	}

	return gemChoice
}

func addSpecifiedGems(statEnchant *stats.StatBlock, socketSlots []stats.SocketType, socketBonus *stats.StatBlock, equipItem loaders.EquippedItem) []stats.GemInfo {
	socketBonusMet := true
	gemChoice := make([]stats.GemInfo, 0, len(socketSlots))

	for index, socketType := range socketSlots {
		gemId := equipItem.GemChoice[index]
		gemInfo := db.GemData_ById(gemId)
		gemChoice = append(gemChoice, gemInfo)
		stats.StatBlock_Increment_Mutating(statEnchant, &gemInfo.Stats)

		if !socketType.SocketMatch(&gemInfo.Stats) {
			socketBonusMet = false
		}
	}

	if socketBonusMet {
		stats.StatBlock_Increment_Mutating(statEnchant, socketBonus)
	}

	return gemChoice
}

func confirmSocketSlots(item items.FullItem, professions model.ProfessionInfo) []stats.SocketType {
	existingSockets := item.SocketSlots()
	if item.SlotItem().AlwaysBlacksmith() || (item.SlotItem().PossibleBlacksmith() && professions.IsBlacksmith) {
		if len(existingSockets) == 0 || existingSockets[len(existingSockets)-1] != stats.Socket_General {
			return util.CopyAndAppend(existingSockets, stats.Socket_General)
		}
	}
	return existingSockets
}

func UpgradeExistingToLevel2(optionsMap *items.FullOptionsMap, targetUpgrade items.UpgradeLevel, model *model.Model, printer *util.PrintRecorder) {
	printer.Println("$$$$ UPGRADE EXISTING ITEMS $$$$")
	optionsMap.MapEachItem(func(currItem *items.FullItem) items.FullItem {
		if currItem.UpgradeLevel() >= targetUpgrade {
			return *currItem
		} else {
			return upgradeItemTo2(currItem, targetUpgrade, model, printer)
		}
	})
}

func upgradeItemTo2(currItem *items.FullItem, targetUpgrade items.UpgradeLevel, model *model.Model, printer *util.PrintRecorder) items.FullItem {
	upgradeItem := db.WowSimDB_ByIdAndUpgrade(currItem.ItemId(), int32(targetUpgrade))
	if upgradeItem == nil {
		printer.Println("$ CAN'T UPGRADE " + currItem.CreateString())
		return *currItem
	} else {
		upgradeItem = copyDetails(currItem, upgradeItem, model.Professions)
		printer.Println("$ UPGRADE IN  << " + currItem.CreateString())
		printer.Println("$ UPGRADE OUT >> " + upgradeItem.CreateString())
		return *upgradeItem
	}
}

func copyDetails(oldItem *items.FullItem, upgradeItem *items.FullItem, professions model.ProfessionInfo) *items.FullItem {
	socketSlots := confirmSocketSlots(*upgradeItem, professions)
	return upgradeItem.NewWithInstanceDetails(socketSlots, oldItem.Reforge(), oldItem.GemChoice(), oldItem.EnchantChoice(),
		oldItem.RandomSuffix(), *oldItem.StatEnchant())
}
