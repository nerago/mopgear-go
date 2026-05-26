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

var g_throneTankTrinkets = []items.ItemId{
	96523, // Delicate Vial of the Sanguinaire
	96421, // Fortitude of the Zandalari
	96471, // Ji-Kun's Rising Winds
	96555, // Soul Barrier
}
var g_throneStrengthTrinkets = []items.ItemId{
	96470, // Fabled Feather of Ji-Kun
	96543, // Gaze of the Twins
	96501, // Primordius' Talisman of Rage
	96398, // Spark of Zandalar
}

var g_seigeTankTrinkets = []items.ItemId{
	// 102316, special PTR item?
	102307, //curse-of-hubris
	102297, //juggernauts-focusing-crystal
	102296, //rooks-unlucky-talisman
	102306, //vial-of-living-corruption
}
var g_siegeStrengthTrinkets = []items.ItemId{
	// 102315, special PTR item?
	102298, //evil-eye-of-galakras
	102295, //fusion-fire-core
	102308, //skeers-bloodsoaked-talisman
	102305, //thoks-tail-tip
}

func ItemFinder_SiegeStrengthPlateTank(difficulty stats.Difficulty) []*items.FullItem {
	return slices.Concat(
		throneClassGearSet(stats.Spec_PaladinProt, difficulty),
		throneClassGearSet(stats.Spec_PaladinRet, difficulty),
		seigeGearGeneric(stats.Armor_Plate, stats.PrimaryStat_Strength, difficulty),
		trinketsForDifficulty(g_seigeTankTrinkets, difficulty, stats.Difficulty.ExpectedItemLevelSiege),
		trinketsForDifficulty(g_siegeStrengthTrinkets, difficulty, stats.Difficulty.ExpectedItemLevelSiege),
	)
}

func ItemFinder_ThroneStrengthPlateTank(difficulty stats.Difficulty) []*items.FullItem {
	return slices.Concat(
		throneClassGearSet(stats.Spec_PaladinProt, difficulty),
		throneClassGearSet(stats.Spec_PaladinRet, difficulty),
		throneGearGeneric(stats.Armor_Plate, stats.PrimaryStat_Strength, difficulty),
		trinketsForDifficulty(g_throneTankTrinkets, difficulty, stats.Difficulty.ExpectedItemLevelThrone),
		trinketsForDifficulty(g_throneStrengthTrinkets, difficulty, stats.Difficulty.ExpectedItemLevelThrone),
		[]*items.FullItem{db.WowSimDB_ByIdAndUpgrade(96436, 0)}, // Tortos' Discarded Shell; int/haste shield
	)
}

func ItemFinder_ThroneStrengthPlateTank_MinusConflictStuff(difficulty stats.Difficulty) []*items.FullItem {
	exclude := []items.ItemId{95513}
	return util.FilterSliceAsNew(ItemFinder_ThroneStrengthPlateTank(difficulty), func(item **items.FullItem) bool {
		return !slices.Contains(exclude, (*item).ItemId())
	})
}

func ItemFinder_ThroneStrengthPlateTank_RadenOnly(difficulty stats.Difficulty) []*items.FullItem {
	return util.FilterSliceAsNew(ItemFinder_ThroneStrengthPlateTank(difficulty), func(item **items.FullItem) bool {
		return isRadenItem((*item).ItemId())
	})
}

func ItemFinder_CelestialCloak(difficulty stats.Difficulty) []*items.FullItem {
	return []*items.FullItem{
		db.WowSimDB_ByIdAndUpgrade(98147, 0),
		db.WowSimDB_ByIdAndUpgrade(98335, 0),
		db.WowSimDB_ByIdAndUpgrade(98146, 0),
	}
}

func seigeClassGearSet(specType stats.SpecType, difficulty stats.Difficulty) []*items.FullItem {
	result := make([]*items.FullItem, 0)
	specBonus := model.SetBonus_ForSpec_AllowFallback(specType, stats.OptimiseGoal_Unknown, true)
	targetLevel := difficulty.ExpectedItemLevelSiege()
	for itemId := range specBonus.AllSetItemIds() {
		item := db.WowSimDB_ByIdAndUpgrade(itemId, 0)
		if item.ItemLevel() == targetLevel {
			result = append(result, item)
		}
	}
	if len(result) != 5 {
		panic("should be 5 items")
	}
	return result
}

func throneClassGearSet(specType stats.SpecType, difficulty stats.Difficulty) []*items.FullItem {
	result := make([]*items.FullItem, 0)
	specBonus := model.SetBonus_ForSpec_AllowFallback(specType, stats.OptimiseGoal_Unknown, true)
	targetLevel := difficulty.ExpectedItemLevelThrone()
	for itemId := range specBonus.AllSetItemIds() {
		item := db.WowSimDB_ByIdAndUpgrade(itemId, 0)
		if item.ItemLevel() == targetLevel {
			result = append(result, item)
		}
	}
	if len(result) != 5 {
		panic("should be 5 items")
	}
	return result
}

func seigeGearGeneric(armor stats.ArmorType, primary stats.PrimaryStatType, difficulty stats.Difficulty) []*items.FullItem {
	groupByName := make(map[string][]*items.FullItem)
	for item := range db.WowSimDB_AllItems() {
		name := item.BaseName()
		if matchesSeigeGearCriteria(item, armor, primary) {
			groupByName[name] = append(groupByName[name], item)
		}
	}

	result := make([]*items.FullItem, 0)
	for _, groupList := range groupByName {
		item := selectAppropriateDifficultyItem(groupList, difficulty, stats.Difficulty.ExpectedItemLevelSiege)
		result = append(result, item)
	}
	return result
}

func throneGearGeneric(armor stats.ArmorType, primary stats.PrimaryStatType, difficulty stats.Difficulty) []*items.FullItem {
	groupByName := make(map[string][]*items.FullItem)
	for item := range db.WowSimDB_AllItems() {
		name := item.BaseName()
		if matchesThroneGearCriteria(item, armor, primary) {
			groupByName[name] = append(groupByName[name], item)
		}
	}

	result := make([]*items.FullItem, 0)
	for _, groupList := range groupByName {
		item := selectAppropriateDifficultyItem(groupList, difficulty, stats.Difficulty.ExpectedItemLevelThrone)
		result = append(result, item)
	}
	return result
}

func selectAppropriateDifficultyItem(itemList []*items.FullItem, difficulty stats.Difficulty, expectedItemLevelFunc func(stats.Difficulty) uint16) *items.FullItem {
	if len(itemList) == 1 {
		return itemList[0]
	}

	targetLevel := expectedItemLevelFunc(difficulty)
	for _, item := range itemList {
		if item.ItemLevel() == targetLevel {
			return item
		}
	}

	if len(itemList) == 2 && itemList[0].ItemLevel() == itemList[1].ItemLevel() {
		return itemList[0]
	}

	// some items don't have heroic version trash drops etc
	if difficulty == stats.Difficulty_Heroic {
		slices.SortFunc(itemList, func(a, b *items.FullItem) int { return cmp.Compare(a.ItemLevel(), b.ItemLevel()) })
		return itemList[len(itemList)-1]
	}

	panic("unknown item choice")
}

func matchesSeigeGearCriteria(item *items.FullItem, armor stats.ArmorType, primary stats.PrimaryStatType) bool {
	return item.Phase() == 5 &&
		item.UpgradeLevel() == 0 &&
		!strings.Contains(item.BaseName(), "Gladiator") &&
		(item.ArmorType().Matches(armor) || item.SlotItem() == items.Item_Back) &&
		item.SlotItem() != items.Item_Trinket &&
		item.PrimaryStat() == primary &&
		!model.SetBonus_IsAnyKnownItem(item.ItemId())
}

func matchesThroneGearCriteria(item *items.FullItem, armor stats.ArmorType, primary stats.PrimaryStatType) bool {
	return item.Phase() == 3 &&
		item.UpgradeLevel() == 0 &&
		!strings.Contains(item.BaseName(), "Gladiator") &&
		(item.ArmorType().Matches(armor) || item.SlotItem() == items.Item_Back) &&
		item.SlotItem() != items.Item_Trinket &&
		item.PrimaryStat() == primary &&
		!model.SetBonus_IsAnyKnownItem(item.ItemId())
}

var g_radenItems = []items.ItemId{95025, 95013, 95001, 95038, 95035, 95033, 95028, 95002, 94995, 95003, 95015, 95010, 95000, 95029, 95030, 95027, 95031, 95023, 95011, 94999, 95036, 95037, 95020, 95018, 95022, 95019, 95021, 95014, 95032, 95040, 95006, 95012, 95034, 95026, 95039, 95004, 94998, 95024, 95005, 95009, 95007, 94996, 95016, 95008, 94997, 95017}

func isRadenItem(itemId items.ItemId) bool {
	return slices.Contains(g_radenItems, itemId)
}

func ItemFinder_FilterOutRadenItems(upgradeItems []*items.FullItem) []*items.FullItem {
	return util.FilterSliceAsNew(upgradeItems, func(item **items.FullItem) bool {
		return !isRadenItem((*item).ItemId())
	})
}

func trinketsForDifficulty(trinketIds []items.ItemId, difficulty stats.Difficulty, expectedItemLevelFunc func(stats.Difficulty) uint16) []*items.FullItem {
	result := make([]*items.FullItem, 0)
	for _, id := range trinketIds {
		item := trinketForDifficulty(id, difficulty, expectedItemLevelFunc)
		result = append(result, item)
	}
	return result
}

func trinketForDifficulty(exampleItemId items.ItemId, difficulty stats.Difficulty, expectedItemLevelFunc func(stats.Difficulty) uint16) *items.FullItem {
	itemName := db.WowSimDB_ByIdAndUpgrade(exampleItemId, 0).BaseName()
	candidates := make([]*items.FullItem, 0)
	for item := range db.WowSimDB_AllItems() {
		if item.BaseName() == itemName {
			candidates = append(candidates, item)
		}
	}
	return selectAppropriateDifficultyItem(candidates, difficulty, expectedItemLevelFunc)
}
