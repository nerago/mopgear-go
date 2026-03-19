package loaders

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/stats"
	"strings"
)

func ThroneProtLootMinusRaden(difficulty stats.Difficulty) {
	// return minusRadenLoot(slices.Concat(
	// 	throneClassGearSet(stats.Spec_PaladinProtMitigation, difficulty),
	// 	throneClassGearSet(stats.Spec_PaladinRet, difficulty),
	// 	strengthPlateThroneNormal(difficulty),
	// 	tankTrinketsThroneNormal(difficulty),
	// 	strengthDpsTrinketsThroneNormal(difficulty)))
}

func throneGearGeneric(armor stats.ArmorType, primary stats.PrimaryStatType, difficulty stats.Difficulty) []items.FullItem {
	groupByName := make(map[string][]*items.FullItem)
	for item := range db.WowSimDB_AllItems() {
		name := item.BaseName
		if item.Phase == 3 && item.Ref.UpgradeLevel() == 0 &&
			!strings.Contains(name, "Gladiator") && item.ArmorType.Matches(armor) && item.PrimaryStat == primary {

			groupByName[name] = append(groupByName[name], item)
		}
	}

	// result := make([]items.FullItem, 0)
	return nil
}

func throneClassGearSet(specType stats.SpecType, difficulty stats.Difficulty) []items.FullItem {
	result := make([]items.FullItem, 0)
	specBonus := model.SetBonus_ForSpec(specType)
	targetLevel := difficulty.ExpectedItemLevel()
	for itemId := range specBonus.AllSetItemIds() {
		item := db.WowSimDB_ByIdAndUpgrade(itemId, 0)
		if item.Ref.ItemLevelBase == targetLevel {
			result = append(result, *item)
		}
	}
	if len(result) != 5 {
		panic("should be 5 items")
	}
	return result
}
