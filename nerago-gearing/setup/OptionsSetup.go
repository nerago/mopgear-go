package setup

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/stats/extern_stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"

	"github.com/wowsims/mop/sim/core"
)

type MissingEnchantMode int8

const (
	MissingEnchant_Ignore MissingEnchantMode = iota
	MissingEnchant_Fix    MissingEnchantMode = iota
	MissingEnchant_Panic  MissingEnchantMode = iota
)

func OptionsSetup_FromGearFile(filename string, model *model.Model, missingEnchant MissingEnchantMode, printer *util.PrintRecorder) items.FullOptionsMap {
	equipped := loaders.GearFileReader_Read(filename)
	return OptionsSetup_FromEquipped(equipped, model, missingEnchant, printer)
}

func OptionsSetup_FromEquipped(equipped []loaders.EquippedItem, model *model.Model, missingEnchant MissingEnchantMode, printer *util.PrintRecorder) items.FullOptionsMap {
	optionMap := items.FullOptionsMap{}
	for _, equipItem := range equipped {
		optionList, baseItem := OptionsSetup_Single_FromEquipped(equipItem, model, missingEnchant, printer)
		optionMap.FillSlot_ExpectedEmpty(baseItem.Slot, optionList)
	}
	return optionMap
}

func OptionsSetup_Single_FromEquipped(equipItem loaders.EquippedItem, model *model.Model, missingEnchant MissingEnchantMode, printer *util.PrintRecorder) ([]items.FullItem, *items.FullItem) {
	item := loadItemBasic(equipItem.ItemId, equipItem.UpgradeStep, printer)
	addDetailFromEquip(&item, equipItem, model, missingEnchant, printer)
	return tools.Reforger_AllOptions(&item, &model.ReforgeRules), &item
}

func OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId items.ItemId, upgradeLevel int8, model *model.Model, printer *util.PrintRecorder) ([]items.FullItem, *items.FullItem) {
	item := loadItemBasic(itemId, upgradeLevel, printer)
	addDetailUsingDefaults(&item, model)
	return tools.Reforger_AllOptions(&item, &model.ReforgeRules), &item
}

func OptionsSetup_ExactEquippedOnly(equipped []loaders.EquippedItem, model *model.Model, printer *util.PrintRecorder) items.FullEquipMap {
	resultMap := items.FullEquipMap{}
	for _, equipItem := range equipped {
		item := loadItemBasic(equipItem.ItemId, equipItem.UpgradeStep, printer)
		addDetailFromEquip(&item, equipItem, model, MissingEnchant_Ignore, printer)

		if equipItem.Reforging != 0 {
			reforge := db.WowSimDB_ReforgeById(equipItem.Reforging)
			item = *tools.Reforger_SinglePreset(&item, &reforge)
		}

		printer.Println(item.CreateString())
		resultMap.FillSlot_ExpectedEmpty(item.Slot, &item)
	}
	return resultMap
}

func loadItemBasic(itemId items.ItemId, upgradeLevel int8, printer *util.PrintRecorder) items.FullItem {
	return *db.WowSimDB_ByIdAndUpgrade_AllowFallback(itemId, upgradeLevel, printer)
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

func addDetailFromEquip(item *items.FullItem, equipItem loaders.EquippedItem, model *model.Model, missingEnchant MissingEnchantMode, printer *util.PrintRecorder) {

	if equipItem.RandomSuffix == -336 {
		item.RandomSuffix = equipItem.RandomSuffix

		// TODO finish
		suffixInfo := core.RandomSuffixesByID[equipItem.RandomSuffix]
		extern_stats.SimStatsToGearStatBlock(suffixInfo.Stats)
		// simulate.RunSize_SlowAccurate
		// suffixInfo.Stats

		stat := stats.Stat_Crit
		amount := itemLevelToRandomAmount[item.Ref.ItemLevel]

		item.StatBase[stat] = amount
		item.ChangeDerivedStatFields()
	} else if equipItem.RandomSuffix != 0 {
		panic("unknown random suffix")
	}

	processGemsAndEnchants(item, equipItem, model, missingEnchant, printer)
}

func addDetailUsingDefaults(item *items.FullItem, model *model.Model) {
	// TODO known random suffixes?

	// TODO trinket modelling
	if item.Slot == items.Item_Trinket {
		return
	}

	item.StatEnchant = stats.StatBlock{}

	enchantInfo := model.EnchantChoice.GetChoice(item.Slot)
	if enchantInfo != nil {
		stats.StatBlock_Increment_Mutating(&item.StatEnchant, &enchantInfo.Stats)
		item.EnchantChoice = enchantInfo.Id
	}

	addBlacksmithSocket(item, model.Professions)

	addDefaultGems(item, model)

	item.ChangeDerivedStatFields()
}

func processGemsAndEnchants(item *items.FullItem, equipItem loaders.EquippedItem, model *model.Model, missingEnchant MissingEnchantMode, printer *util.PrintRecorder) {
	if item.Slot == items.Item_Trinket {
		return
	}

	item.StatEnchant = stats.StatBlock{}

	if equipItem.EnchantChoice != 0 {
		enchantInfo := db.EnchantData_ById(equipItem.EnchantChoice)
		stats.StatBlock_Increment_Mutating(&item.StatEnchant, &enchantInfo.Stats)
		item.EnchantChoice = equipItem.EnchantChoice
	} else if item.Slot.CanEnchant() && missingEnchant == MissingEnchant_Panic {
		panic("missing enchant on " + item.CreateString())
	} else if enchantInfo := model.EnchantChoice.GetChoice(item.Slot); enchantInfo != nil && missingEnchant == MissingEnchant_Fix {
		before := item.CreateString()
		stats.StatBlock_Increment_Mutating(&item.StatEnchant, &enchantInfo.Stats)
		item.EnchantChoice = enchantInfo.Id
		printer.Println("Add missing enchant " + before + " ==> " + item.CreateString())
	} else if item.Slot.CanEnchant() && missingEnchant == MissingEnchant_Fix {
		panic("no enchant choice for " + item.Slot.Name())
	}

	addBlacksmithSocket(item, model.Professions)

	if len(equipItem.GemChoice) == len(item.SocketSlots) {
		addSpecifiedGems(item, equipItem)
	} else if len(equipItem.GemChoice) == 0 {
		switch missingEnchant {
		case MissingEnchant_Fix:
			before := item.CreateString()
			addDefaultGems(item, model)
			printer.Println("Add missing gems " + before + " ==> " + item.CreateString())
		case MissingEnchant_Panic:
			panic("missing gems on " + item.CreateString())
		}
	} else {
		panic("mismatch in gem array lengths on " + item.CreateString())
	}

	item.ChangeDerivedStatFields()
}

func addDefaultGems(item *items.FullItem, model *model.Model) {
	socketBonusMet := true
	item.GemChoice = make([]stats.GemInfo, 0, len(item.SocketSlots))

	for _, socketType := range item.SocketSlots {
		// NOTE unique engineering gems not checked
		gemInfo := model.GemChoice.GetChoice(socketType)
		item.GemChoice = append(item.GemChoice, *gemInfo)
		stats.StatBlock_Increment_Mutating(&item.StatEnchant, &gemInfo.Stats)

		if !socketType.SocketMatch(&gemInfo.Stats) {
			socketBonusMet = false
		}
	}

	if socketBonusMet {
		stats.StatBlock_Increment_Mutating(&item.StatEnchant, &item.SocketBonus)
	}
}

func addSpecifiedGems(item *items.FullItem, equipItem loaders.EquippedItem) {
	socketBonusMet := true
	item.GemChoice = make([]stats.GemInfo, 0, len(item.SocketSlots))

	for index, socketType := range item.SocketSlots {
		gemId := equipItem.GemChoice[index]
		gemInfo := db.GemData_ById(gemId)
		item.GemChoice = append(item.GemChoice, gemInfo)
		stats.StatBlock_Increment_Mutating(&item.StatEnchant, &gemInfo.Stats)

		if !socketType.SocketMatch(&gemInfo.Stats) {
			socketBonusMet = false
		}
	}

	if socketBonusMet {
		stats.StatBlock_Increment_Mutating(&item.StatEnchant, &item.SocketBonus)
	}
}

func addBlacksmithSocket(item *items.FullItem, professions model.ProfessionInfo) {
	if item.Slot.AlwaysBlacksmith() || (item.Slot.PossibleBlacksmith() && professions.IsBlacksmith) {
		if len(item.SocketSlots) == 0 || item.SocketSlots[len(item.SocketSlots)-1] != stats.Socket_General {
			item.SocketSlots = append(item.SocketSlots, stats.Socket_General)
		}
	}
}

func UpgradeExistingToLevel2(optionsMap *items.FullOptionsMap, targetUpgrade int8, model *model.Model, printer *util.PrintRecorder) {
	printer.Println("$$$$ UPGRADE EXISTING ITEMS $$$$")
	optionsMap.MapEachItem(func(currItem *items.FullItem) items.FullItem {
		if currItem.Ref.UpgradeLevel >= targetUpgrade {
			return *currItem
		} else {
			return upgradeItemTo2(currItem, targetUpgrade, model, printer)
		}
	})
	printer.Println("$$$$")
}

func upgradeItemTo2(currItem *items.FullItem, targetUpgrade int8, model *model.Model, printer *util.PrintRecorder) items.FullItem {
	upgradeItem := db.WowSimDB_ByIdAndUpgrade(currItem.ItemId(), targetUpgrade)
	if upgradeItem == nil {
		printer.Println("$ CAN'T UPGRADE " + currItem.CreateString())
		return *currItem
	} else {
		copyDetails(currItem, upgradeItem, model.Professions)
		printer.Println("$ UPGRADE IN  << " + currItem.CreateString())
		printer.Println("$ UPGRADE OUT >> " + upgradeItem.CreateString())
		// panic("upgrade TODO")
		return *upgradeItem
	}
}

func copyDetails(currItem *items.FullItem, upgradeItem *items.FullItem, professions model.ProfessionInfo) {
	upgradeItem.Reforge = currItem.Reforge
	upgradeItem.GemChoice = currItem.GemChoice
	upgradeItem.EnchantChoice = currItem.EnchantChoice
	upgradeItem.RandomSuffix = currItem.RandomSuffix
	upgradeItem.StatEnchant = currItem.StatEnchant
	addBlacksmithSocket(upgradeItem, professions)
	upgradeItem.ChangeDerivedStatFields()
}
