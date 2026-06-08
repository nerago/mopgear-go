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

var G_seigeTankTrinkets = []items.ItemId{
	// 102316, special PTR item?
	102307, //curse-of-hubris
	102297, //juggernauts-focusing-crystal
	102296, //rooks-unlucky-talisman
	102306, //vial-of-living-corruption
	103990, //resolve-of-niuzao
}
var G_siegeStrengthTrinkets = []items.ItemId{
	// 102315, special PTR item?
	102298, //evil-eye-of-galakras
	102295, //fusion-fire-core
	102308, //skeers-bloodsoaked-talisman
	102305, //thoks-tail-tip
	103989, //alacrity-of-xuen
}

var legendCloaks = []items.ItemId{102249, 102250}

var ordosItems = []items.ItemId{105804, 105766, 105758, 105782, 105776, 105784, 105789, 105765, 105795, 105792, 105793, 105791, 105810, 105787, 105774, 105771, 105806, 105809, 105808, 105778, 105754, 105805, 105800, 105790, 105798, 105799, 105775, 105783, 105760, 105767, 105779, 105807, 105759, 105772, 105755, 105811, 105769, 105786, 105768, 105761, 105788, 105763, 105756, 105777, 105764, 105796, 105797, 105757, 105762, 105801, 105794, 105803, 105773, 105785, 105781, 105780, 105802, 105770}

func ItemFinder_TimelessPlate(_ stats.Difficulty) []*items.FullItem {
	// var targetLevel uint16 = 496
	var targetLevel uint16 = 535
	result := make([]*items.FullItem, 0)
	for item := range db.WowSimDB_AllItems() {
		if (strings.Contains(item.BaseName(), "Cliffbreaker") || strings.Contains(item.BaseName(), "Elder Tortoiseshell")) && item.ItemLevel() == targetLevel {
			onlyCoreStats(item)

			var randomId int32
			if item.SlotItem() == items.Item_Head || item.SlotItem() == items.Item_Leg || item.SlotItem() == items.Item_Chest || item.SlotItem() == items.Item_Ring {
				randomId = -454 // haste/mastery
			} else if item.SlotItem() == items.Item_Wrist || item.SlotItem() == items.Item_Belt {
				randomId = -340 // haste
			}
			item = item.MakeItemWithRandomSuffix(randomId) // give it haste

			result = append(result, item)
		}
	}
	return result
}

func onlyCoreStats(item *items.FullItem) {
	for statIdx, value := range item.Total() {
		stat := stats.StatType(statIdx)
		if stat == stats.Stat_Strength || stat == stats.Stat_Stamina {
			if value == 0 {
				panic("expected value")
			}
		} else {
			if value != 0 {
				panic("expected zero")
			}
		}
	}
}

func ItemFinder_Ordos(_ stats.Difficulty) []*items.FullItem {
	result := make([]*items.FullItem, 0)
	for _, itemId := range ordosItems {
		item := db.WowSimDB_ByIdAndUpgrade(itemId, 0)
		if item.ArmorType() == stats.Armor_None || item.ArmorType() == stats.Armor_Plate || item.SlotItem() == items.Item_Back {
			if item.PrimaryStat() == stats.PrimaryStat_None || item.PrimaryStat() == stats.PrimaryStat_Strength {
				result = append(result, item)
			}
		}
	}
	return result
}

func ItemFinder_SiegeStrengthPlateTank(difficulty stats.Difficulty) []*items.FullItem {
	return slices.Concat(
		seigeClassGearSet(stats.Spec_PaladinProt, difficulty),
		seigeClassGearSet(stats.Spec_PaladinRet, difficulty),
		seigeGearGeneric(stats.Armor_Plate, stats.PrimaryStat_Strength, difficulty),
		trinketsForDifficulty(G_seigeTankTrinkets, difficulty, stats.Difficulty.ExpectedItemLevelSiege),
		trinketsForDifficulty(G_siegeStrengthTrinkets, difficulty, stats.Difficulty.ExpectedItemLevelSiege),
		// []*items.FullItem{db.WowSimDB_ByIdAndUpgrade(102249, 0), db.WowSimDB_ByIdAndUpgrade(102250, 0)},
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
		if item != nil {
			result = append(result, item)
		}
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
		if item != nil {
			result = append(result, item)
		}
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
	slices.SortFunc(itemList, func(a, b *items.FullItem) int { return cmp.Compare(a.ItemLevel(), b.ItemLevel()) })
	if difficulty == stats.Difficulty_Heroic {
		return itemList[len(itemList)-1]
	}

	// timeless pieces
	if len(itemList) == 2 && itemList[0].ItemLevel() == 496 && itemList[1].ItemLevel() == 535 {
		return itemList[1]
	}

	// garrosh weapons
	if len(itemList) == 3 && itemList[0].ItemLevel() == 556 {
		switch difficulty {
		case stats.Difficulty_Celestial:
			return nil
		case stats.Difficulty_Normal:
			return itemList[1]
		case stats.Difficulty_Heroic:
			return itemList[2]
		}
	}

	return nil
	// panic("unknown item choice")
}

func matchesSeigeGearCriteria(item *items.FullItem, armor stats.ArmorType, primary stats.PrimaryStatType) bool {
	return item.Phase() == 5 &&
		item.UpgradeLevel() == 0 &&
		!strings.Contains(item.BaseName(), "Gladiator") &&
		// TODO how do weapons work here?
		(item.ArmorType().Matches(armor) || item.SlotItem() == items.Item_Back) &&
		item.SlotItem() != items.Item_Trinket &&
		item.PrimaryStat() == primary &&
		item.ItemLevel() > 500 &&
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
		if item != nil {
			result = append(result, item)
		}
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
