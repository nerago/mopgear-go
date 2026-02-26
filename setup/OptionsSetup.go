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
	return convertToOptions(equipped, model, printer)
}

func convertToOptions(equipped []loaders.EquippedItem, model *model.Model, printer *util.PrintRecorder) items.FullOptionsMap {
	optionMap := items.FullOptionsMap{}
	for _, item := range equipped {
		baseItem := loadItem(item)
		printer.Println(baseItem.String())
		optionList := tools.Reforger_AllOptions(baseItem, &model.ReforgeRules)
		optionMap.FillSlot_ExpectedEmpty(baseItem.Slot, optionList)
	}
	return optionMap
}

func loadItem(equipItem loaders.EquippedItem) *items.FullItem {
	storedItem := db.WowSimDB_ByIdAndUpgrade(equipItem.ItemId, equipItem.UpgradeStep)
	if storedItem == nil && equipItem.UpgradeStep > 0 {
		storedItem = db.WowSimDB_ByIdAndUpgrade(equipItem.ItemId, 0)
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

func addDetails(item *items.FullItem, equipItem loaders.EquippedItem) *items.FullItem {
	if equipItem.RandomSuffix == -336 {
		stat := stats.Stat_Crit
		amount := itemLevelToRandomAmount[item.Ref.ItemLevel]
		item = item.ChangedBaseStats(item.StatBase.WithChange(stat, amount))
		item.RandomSuffix = equipItem.RandomSuffix
	} else if equipItem.RandomSuffix != 0 {
		panic("unknown random suffix")
	}

	if item.Slot != stats.Item_Trinket {
		var enchantStats stats.StatBlock
		enchantStats, item.GemChoice = calcGemsAndEnchants(item, equipItem)
		item = item.ChangedEnchantStats(enchantStats)
	}
	item.EnchantChoice = equipItem.EnchantChoice

	// TODO trinket model

	return item
}

// TODO enchant validation
func calcGemsAndEnchants(item *items.FullItem, equipItem loaders.EquippedItem) (stats.StatBlock, []stats.GemInfo) {
	block := stats.StatBlock{}
	gemChoice := make([]stats.GemInfo, 0)

	if equipItem.EnchantChoice != 0 {
		enchantValue := db.EnchantData_ById(equipItem.EnchantChoice)
		block.Increment_Mutating(&enchantValue.Stats)
	}

	if item.Slot.PossibleBlacksmith() {
		item.SocketSlots = append(item.SocketSlots, stats.Socket_General)
	}

	socketBonusMet := true
	for index, gemId := range equipItem.GemChoice {
		gemInfo := db.GemData_ById(gemId)
		gemChoice = append(gemChoice, gemInfo)
		block.Increment_Mutating(&gemInfo.Stats)

		socket := item.SocketSlots[index]
		if !socket.SocketMatch(&gemInfo.Stats) {
			socketBonusMet = false
		}
	}

	if socketBonusMet {
		block.Increment_Mutating(&item.SocketBonus)
	}

	return block, gemChoice
}
