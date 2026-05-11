package tools

import (
	"log"
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/stats"
)

func Reforger_AllOptions(baseItem *FullItem, rules *ReforgeRules) []FullItem {
	outputItems := []FullItem{*baseItem}

	targetArray := rules.Target()
	sourceArray := rules.Source()

	for _, source := range sourceArray {
		originalValue := baseItem.StatBase().Get(source)
		if originalValue != 0 {
			reforgeQuantity := (originalValue * 4) / 10
			remainQuantity := originalValue - reforgeQuantity

			for _, target := range targetArray {
				if baseItem.StatBase().Get(target) == 0 {
					modified := makeModified(baseItem, source, target, reforgeQuantity, remainQuantity)
					outputItems = append(outputItems, *modified)
				}
			}
		}
	}

	return outputItems
}

func Reforger_SinglePreset(baseItem *FullItem, recipe *ReforgeRecipe) *FullItem {
	if recipe.IsEmpty() {
		return baseItem
	}

	source := recipe.From
	target := recipe.To
	if source == target {
		log.Panic("expected different stats")
	}

	originalValue := baseItem.StatBase().Get(source)
	if originalValue == 0 {
		log.Panicf("expected item to have source stat %s on %s", source.Name(), baseItem.CreateString())
	}

	if baseItem.StatBase().Get(target) != 0 {
		log.Panicf("expected item to have zero target stat %s on %s", target.Name(), baseItem.CreateString())
	}

	reforgeQuantity := (originalValue * 4) / 10
	remainQuantity := originalValue - reforgeQuantity
	return makeModified(baseItem, source, target, reforgeQuantity, remainQuantity)
}

func makeModified(baseItem *FullItem, source, target StatType, reforgeQuantity, remainQuantity uint32) *FullItem {
	var newStats StatBlock = *baseItem.StatBase()
	newStats[source] = remainQuantity
	newStats[target] = reforgeQuantity

	reforge := ReforgeRecipe_of(source, target)
	return baseItem.NewWithChangedStatsReforge(newStats, reforge)
}
