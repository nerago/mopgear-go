package loaders

import (
	"cmp"
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
	"strings"
)

var g_tankTrinkets = []items.ItemId{
	96523, // Delicate Vial of the Sanguinaire
	96421, // Fortitude of the Zandalari
	96471, // Ji-Kun's Rising Winds
	96555} // Soul Barrier
var g_strengthTrinkets = []items.ItemId{
	96470, // Fabled Feather of Ji-Kun
	96543, // Gaze of the Twins
	96501, // Primordius' Talisman of Rage
	96398} // Spark of Zandalar

func ItemFinder_ThroneStrengthPlateTank(difficulty stats.Difficulty) []*items.FullItem {
	return slices.Concat(
		throneClassGearSet(stats.Spec_PaladinProtMitigation, difficulty),
		throneClassGearSet(stats.Spec_PaladinRet, difficulty),
		throneGearGeneric(stats.Armor_Plate, stats.PrimaryStat_Strength, difficulty),
		trinketsForDifficulty(g_tankTrinkets, difficulty),
		trinketsForDifficulty(g_strengthTrinkets, difficulty),
		[]*items.FullItem{db.WowSimDB_ByIdAndUpgrade(96436, 0)}, // Tortos' Discarded Shell; int/haste shield
	)
}

func ItemFinder_ThroneStrengthPlateTank_MinusConflictStuff(difficulty stats.Difficulty) []*items.FullItem {
	exclude := []items.ItemId{95513}
	return util.FilterSliceAsNew(ItemFinder_ThroneStrengthPlateTank(difficulty), func(item **items.FullItem) bool {
		return !slices.Contains(exclude, (*item).ItemId())
	})
}

func ItemFinder_CelestialCloak(difficulty stats.Difficulty) []*items.FullItem {
	return []*items.FullItem{
		db.WowSimDB_ByIdAndUpgrade(98147, 0),
		db.WowSimDB_ByIdAndUpgrade(98335, 0),
		db.WowSimDB_ByIdAndUpgrade(98146, 0),
	}
}

func throneClassGearSet(specType stats.SpecType, difficulty stats.Difficulty) []*items.FullItem {
	result := make([]*items.FullItem, 0)
	specBonus := model.SetBonus_ForSpec(specType)
	targetLevel := difficulty.ExpectedItemLevel()
	for itemId := range specBonus.AllSetItemIds() {
		item := db.WowSimDB_ByIdAndUpgrade(itemId, 0)
		if item.Ref.ItemLevel == targetLevel {
			result = append(result, item)
		}
	}
	if len(result) != 5 {
		panic("should be 5 items")
	}
	return result
}

func throneGearGeneric(armor stats.ArmorType, primary stats.PrimaryStatType, difficulty stats.Difficulty) []*items.FullItem {
	groupByName := make(map[string][]*items.FullItem)
	for item := range db.WowSimDB_AllItems() {
		name := item.BaseName
		if matchesGenericGearCriteria(item, armor, primary) {
			groupByName[name] = append(groupByName[name], item)
		}
	}

	result := make([]*items.FullItem, 0)
	for _, groupList := range groupByName {
		item := selectAppropriateDifficultyItem(groupList, difficulty)
		result = append(result, item)
	}
	return result
}

func selectAppropriateDifficultyItem(itemList []*items.FullItem, difficulty stats.Difficulty) *items.FullItem {
	if len(itemList) == 1 {
		return itemList[0]
	}

	targetLevel := difficulty.ExpectedItemLevel()
	for _, item := range itemList {
		if item.Ref.ItemLevel == targetLevel {
			return item
		}
	}

	if len(itemList) == 2 && itemList[0].Ref.ItemLevel == itemList[1].Ref.ItemLevel {
		return itemList[0]
	}

	// some items don't have heroic version trash drops etc
	if difficulty == stats.Difficulty_Heroic {
		slices.SortFunc(itemList, func(a, b *items.FullItem) int { return cmp.Compare(a.Ref.ItemLevel, b.Ref.ItemLevel) })
		return itemList[len(itemList)-1]
	}

	// slices.SortFunc(itemList, func(a, b *items.FullItem) int { return cmp.Compare(a.Ref.ItemLevel, b.Ref.ItemLevel) })
	// if difficulty == stats.Difficulty_Heroic {
	// 	return itemList[len(itemList) - 1]
	// } else {
	// 	return itemList[0]
	// }

	panic("unknown item choice")
}

func matchesGenericGearCriteria(item *items.FullItem, armor stats.ArmorType, primary stats.PrimaryStatType) bool {
	return item.Phase == 3 &&
		item.Ref.UpgradeLevel == 0 &&
		!strings.Contains(item.BaseName, "Gladiator") &&
		(item.ArmorType.Matches(armor) || item.Slot == items.Item_Back) &&
		item.Slot != items.Item_Trinket &&
		item.PrimaryStat == primary &&
		!model.SetBonus_IsAnyKnownItem(item.ItemId())
}

var g_radenItems = []items.ItemId{95025, 95013, 95001, 95038, 95035, 95033, 95028, 95002, 94995, 95003, 95015, 95010, 95000, 95029, 95030, 95027, 95031, 95023, 95011, 94999, 95036, 95037, 95020, 95018, 95022, 95019, 95021, 95014, 95032, 95040, 95006, 95012, 95034, 95026, 95039, 95004, 94998, 95024, 95005, 95009, 95007, 94996, 95016, 95008, 94997, 95017}

func isRadenItem(itemId items.ItemId) bool {
	return slices.Contains(g_radenItems, itemId)
}

func ItemFinder_FilterOutRadenItems(upgradeItems []*items.FullItem) []*items.FullItem {
	return util.FilterSliceInPlace(upgradeItems, func(item **items.FullItem) bool {
		return !isRadenItem((*item).ItemId())
	})
}

func trinketsForDifficulty(trinketIds []items.ItemId, difficulty stats.Difficulty) []*items.FullItem {
	result := make([]*items.FullItem, 0)
	for _, id := range trinketIds {
		item := trinketForDifficulty(id, difficulty)
		result = append(result, item)
	}
	return result
}

func trinketForDifficulty(exampleItemId items.ItemId, difficulty stats.Difficulty) *items.FullItem {
	itemName := db.WowSimDB_ByIdAndUpgrade(exampleItemId, 0).BaseName
	candidates := make([]*items.FullItem, 0)
	for item := range db.WowSimDB_AllItems() {
		if item.BaseName == itemName {
			candidates = append(candidates, item)
		}
	}
	return selectAppropriateDifficultyItem(candidates, difficulty)
}
