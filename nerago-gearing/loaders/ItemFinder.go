package loaders

import (
	"cmp"
	"paladin_gearing_go/db"
	"paladin_gearing_go/gear_model/bonus_set"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_collection"
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
	103678, //time-lost artifact
}

var legendCloaks = []items.ItemId{102249, 102250}

var ordosItems = []items.ItemId{105804, 105766, 105758, 105782, 105776, 105784, 105789, 105765, 105795, 105792, 105793, 105791, 105810, 105787, 105774, 105771, 105806, 105809, 105808, 105778, 105754, 105805, 105800, 105790, 105798, 105799, 105775, 105783, 105760, 105767, 105779, 105807, 105759, 105772, 105755, 105811, 105769, 105786, 105768, 105761, 105788, 105763, 105756, 105777, 105764, 105796, 105797, 105757, 105762, 105801, 105794, 105803, 105773, 105785, 105781, 105780, 105802, 105770}

type ItemFoundRef struct {
	ItemId       items.ItemId
	RandomSuffix items.RandomSuffix
	UpgradeLevel items.UpgradeLevel
}

func ItemFoundRef_Of(item *items.FullItem) ItemFoundRef {
	return ItemFoundRef{
		item.ItemId(), item.RandomSuffix(), item.UpgradeLevel(),
	}
}

func (ref ItemFoundRef) Equals(other ItemFoundRef) bool {
	return ref == other
}

func ItemFinder_TimelessPlate(_ stats.Difficulty) []ItemFoundRef {
	var targetLevel uint16 = 535
	result := make([]ItemFoundRef, 0)
	for item := range db.WowSimDB_AllItems() {
		if (strings.Contains(item.BaseName(), "Cliffbreaker") || strings.Contains(item.BaseName(), "Elder Tortoiseshell")) && item.ItemLevel() == targetLevel {
			onlyCoreStats(item)

			//var randomId items.RandomSuffix = -454         // haste/mastery
			//item = item.MakeItemWithRandomSuffix(randomId) // give it haste
			//result = append(result, item)

			var randomId items.RandomSuffix = -340 // haste
			result = append(result, ItemFoundRef{ItemId: item.ItemId(), UpgradeLevel: item.UpgradeLevel(), RandomSuffix: randomId})

			for r := -460; r <= -441; r++ {
				result = append(result, ItemFoundRef{ItemId: item.ItemId(), UpgradeLevel: item.UpgradeLevel(), RandomSuffix: items.RandomSuffix(r)})
			}
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

func ItemFinder_Ordos(_ stats.Difficulty) []ItemFoundRef {
	result := make([]ItemFoundRef, 0)
	for _, itemId := range ordosItems {
		item := db.WowSimDB_LoadItemById(itemId, 0)
		if item.ArmorType() == stats.Armor_None || item.ArmorType() == stats.Armor_Plate || item.SlotItem() == items.Item_Back {
			if item.PrimaryStat() == stats.PrimaryStat_None || item.PrimaryStat() == stats.PrimaryStat_Strength {
				result = append(result, ItemFoundRef_Of(item))
			}
		}
	}
	return result
}

func ItemFinder_NormalHeroicBossFiltered(innerFinder func(stats.Difficulty) []ItemFoundRef, normalBossNames []string, heroicBossNames []string) func(stats.Difficulty) []ItemFoundRef {
	return func(difficulty stats.Difficulty) []ItemFoundRef {
		regularResult := innerFinder(difficulty)
		if difficulty == stats.Difficulty_Celestial {
			return regularResult
		} else if difficulty == stats.Difficulty_Normal {
			return util_collection.FilterSliceAsNew_NoPointer(regularResult, func(item ItemFoundRef) bool {
				bossName := db.BossItemData_BossForItemId(item.ItemId)
				return slices.Contains(normalBossNames, bossName)
			})
		} else {
			return util_collection.FilterSliceAsNew_NoPointer(regularResult, func(item ItemFoundRef) bool {
				bossName := db.BossItemData_BossForItemId(item.ItemId)
				return slices.Contains(heroicBossNames, bossName)
			})
		}
	}
}

func ItemFinder_HeroicBossFiltered(innerFinder func(stats.Difficulty) []ItemFoundRef, heroicBossNames []string) func(stats.Difficulty) []ItemFoundRef {
	return func(difficulty stats.Difficulty) []ItemFoundRef {
		regularResult := innerFinder(difficulty)
		if difficulty == stats.Difficulty_Celestial || difficulty == stats.Difficulty_Normal {
			return regularResult
		}

		return util_collection.FilterSliceAsNew_NoPointer(regularResult, func(item ItemFoundRef) bool {
			bossName := db.BossItemData_BossForItemId(item.ItemId)
			return slices.Contains(heroicBossNames, bossName)
		})
	}
}

func ItemFinder_SiegeStrengthPlateTank(difficulty stats.Difficulty) []ItemFoundRef {
	initial := slices.Concat(
		siegeClassGearSet(stats.Spec_PaladinProt, difficulty),
		siegeClassGearSet(stats.Spec_PaladinRet, difficulty),
		siegeGearGeneric(stats.Armor_Plate, stats.PrimaryStat_Strength, difficulty),
		trinketsForDifficulty(G_seigeTankTrinkets, difficulty, stats.Difficulty.ExpectedItemLevelSiege),
		trinketsForDifficulty(G_siegeStrengthTrinkets, difficulty, stats.Difficulty.ExpectedItemLevelSiege),
	)
	// Visage of the Monstrous spirit/haste shield
	if difficulty == stats.Difficulty_Normal {
		initial = append(initial, ItemFoundRef{ItemId: 103848})
	} else if difficulty == stats.Difficulty_Heroic {
		initial = append(initial, ItemFoundRef{ItemId: 104579})
	}
	return initial
}

func ItemFinder_ThroneStrengthPlateTank(difficulty stats.Difficulty) []ItemFoundRef {
	return slices.Concat(
		throneClassGearSet(stats.Spec_PaladinProt, difficulty),
		throneClassGearSet(stats.Spec_PaladinRet, difficulty),
		throneGearGeneric(stats.Armor_Plate, stats.PrimaryStat_Strength, difficulty),
		trinketsForDifficulty(g_throneTankTrinkets, difficulty, stats.Difficulty.ExpectedItemLevelThrone),
		trinketsForDifficulty(g_throneStrengthTrinkets, difficulty, stats.Difficulty.ExpectedItemLevelThrone),
		[]ItemFoundRef{{ItemId: 96436}}, // Tortos' Discarded Shell; int/haste shield
	)
}

func ItemFinder_ThroneStrengthPlateTank_RadenOnly(difficulty stats.Difficulty) []ItemFoundRef {
	result := util_collection.FilterSliceAsNew(ItemFinder_ThroneStrengthPlateTank(difficulty), func(item *ItemFoundRef) bool {
		return isRadenItem(item.ItemId)
	})
	util_collection.RemoveDuplicatesComparable_InPlace(&result)
	return result
}

func SiegeClassGearSetMultiple(specType ...stats.SpecType) func(stats.Difficulty) []ItemFoundRef {
	return func(difficulty stats.Difficulty) []ItemFoundRef {
		result := make([]ItemFoundRef, 0)
		for _, spec := range specType {
			result = append(result, siegeClassGearSet(spec, difficulty)...)
		}
		return result
	}
}

func siegeClassGearSet(specType stats.SpecType, difficulty stats.Difficulty) []ItemFoundRef {
	result := make([]ItemFoundRef, 0)
	specBonus := bonus_set.SpecSetsEnableForSpec_AllowFallback(specType, stats.OptimiseGoal_Unknown, true, nil)
	targetLevel := difficulty.ExpectedItemLevelSiege()
	for itemId := range specBonus.AllItemIds() {
		item := db.WowSimDB_LoadItemById(itemId, 0)
		if item.ItemLevel() == targetLevel {
			result = append(result, ItemFoundRef_Of(item))
		}
	}
	if len(result) != 5 {
		panic("should be 5 items")
	}
	return result
}

func throneClassGearSet(specType stats.SpecType, difficulty stats.Difficulty) []ItemFoundRef {
	result := make([]ItemFoundRef, 0)
	specBonus := bonus_set.SpecSetsEnableForSpec_AllowFallback(specType, stats.OptimiseGoal_Unknown, true, nil)
	targetLevel := difficulty.ExpectedItemLevelThrone()
	for itemId := range specBonus.AllItemIds() {
		item := db.WowSimDB_LoadItemById(itemId, 0)
		if item.ItemLevel() == targetLevel {
			result = append(result, ItemFoundRef_Of(item))
		}
	}
	if len(result) != 5 {
		panic("should be 5 items")
	}
	return result
}

func siegeGearGeneric(armor stats.ArmorType, primary stats.PrimaryStatType, difficulty stats.Difficulty) []ItemFoundRef {
	groupByName := make(map[string][]*items.FullItem)
	for item := range db.WowSimDB_AllItems() {
		name := item.BaseName()
		if difficulty == stats.Difficulty_Celestial && item.ItemLevel() >= 555 {
			continue
		}
		if matchesSiegeGearCriteria(item, armor, primary) {
			groupByName[name] = append(groupByName[name], item)
		}
	}

	result := make([]ItemFoundRef, 0)
	for _, groupList := range groupByName {
		item := selectAppropriateDifficultyItem(groupList, difficulty, stats.Difficulty.ExpectedItemLevelSiege)
		if item != nil {
			result = append(result, *item)
		}
	}

	return result
}

func throneGearGeneric(armor stats.ArmorType, primary stats.PrimaryStatType, difficulty stats.Difficulty) []ItemFoundRef {
	groupByName := make(map[string][]*items.FullItem)
	for item := range db.WowSimDB_AllItems() {
		name := item.BaseName()
		if matchesThroneGearCriteria(item, armor, primary) {
			groupByName[name] = append(groupByName[name], item)
		}
	}

	result := make([]ItemFoundRef, 0)
	for _, groupList := range groupByName {
		item := selectAppropriateDifficultyItem(groupList, difficulty, stats.Difficulty.ExpectedItemLevelThrone)
		if item != nil {
			result = append(result, *item)
		}
	}
	return result
}

func selectAppropriateDifficultyItem(itemList []*items.FullItem, difficulty stats.Difficulty, expectedItemLevelFunc func(stats.Difficulty) uint16) *ItemFoundRef {
	item := selectAppropriateDifficultyItemFull(itemList, difficulty, expectedItemLevelFunc)
	if item == nil {
		return nil
	} else {
		ref := ItemFoundRef_Of(item)
		return &ref
	}
}

func selectAppropriateDifficultyItemFull(itemList []*items.FullItem, difficulty stats.Difficulty, expectedItemLevelFunc func(stats.Difficulty) uint16) *items.FullItem {
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

func matchesSiegeGearCriteria(item *items.FullItem, armor stats.ArmorType, primary stats.PrimaryStatType) bool {
	return item.Phase() == 5 &&
		item.UpgradeLevel() == 0 &&
		!strings.Contains(item.BaseName(), "Gladiator") &&
		(item.ArmorType().Matches(armor) || item.SlotItem() == items.Item_Back) &&
		item.SlotItem() != items.Item_Trinket &&
		item.PrimaryStat() == primary &&
		item.ItemLevel() > 500 &&
		!bonus_set.IsAnyKnownItem(item.ItemId())
}

func matchesThroneGearCriteria(item *items.FullItem, armor stats.ArmorType, primary stats.PrimaryStatType) bool {
	return item.Phase() == 3 &&
		item.UpgradeLevel() == 0 &&
		!strings.Contains(item.BaseName(), "Gladiator") &&
		(item.ArmorType().Matches(armor) || item.SlotItem() == items.Item_Back) &&
		item.SlotItem() != items.Item_Trinket &&
		item.PrimaryStat() == primary &&
		!bonus_set.IsAnyKnownItem(item.ItemId())
}

var g_radenItems = []items.ItemId{95025, 95013, 95001, 95038, 95035, 95033, 95028, 95002, 94995, 95003, 95015, 95010, 95000, 95029, 95030, 95027, 95031, 95023, 95011, 94999, 95036, 95037, 95020, 95018, 95022, 95019, 95021, 95014, 95032, 95040, 95006, 95012, 95034, 95026, 95039, 95004, 94998, 95024, 95005, 95009, 95007, 94996, 95016, 95008, 94997, 95017}

func isRadenItem(itemId items.ItemId) bool {
	return slices.Contains(g_radenItems, itemId)
}

func trinketsForDifficulty(trinketIds []items.ItemId, difficulty stats.Difficulty, expectedItemLevelFunc func(stats.Difficulty) uint16) []ItemFoundRef {
	result := make([]ItemFoundRef, 0)
	for _, id := range trinketIds {
		item := trinketForDifficulty(id, difficulty, expectedItemLevelFunc)
		if item != nil {
			result = append(result, *item)
		}
	}
	return result
}

func trinketForDifficulty(exampleItemId items.ItemId, difficulty stats.Difficulty, expectedItemLevelFunc func(stats.Difficulty) uint16) *ItemFoundRef {
	itemName := db.LookupItemNameByItemId(exampleItemId)
	candidates := make([]*items.FullItem, 0)
	for item := range db.WowSimDB_AllItems() {
		if item.BaseName() == itemName {
			candidates = append(candidates, item)
		}
	}
	return selectAppropriateDifficultyItem(candidates, difficulty, expectedItemLevelFunc)
}

func ItemFinder_BagsUpgraded(_ stats.Difficulty) []ItemFoundRef {
	bagsItems := BagsFile_PlusPaladinGear_Read()

	result := make([]ItemFoundRef, 0, len(bagsItems))
	for _, equip := range bagsItems {
		item := db.WowSimDB_LoadItemById(equip.ItemId, equip.UpgradeStepOrItemLevel)
		if item.UpgradeLevel() < items.MAX_UPGRADE_LEVEL {
			upgraded := db.WowSimDB_LoadItemById(equip.ItemId, int32(items.MAX_UPGRADE_LEVEL))
			if upgraded != nil && upgraded.ItemLevel() > item.ItemLevel() {
				result = append(result, ItemFoundRef_Of(upgraded))
			}
		}
	}

	util_collection.RemoveDuplicatesComparable_InPlace(&result)
	return result
}
