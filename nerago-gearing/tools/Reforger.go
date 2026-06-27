package tools

import (
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/stats"
)

func Reforger_AllOptions(baseItem *FullItem, rules *ReforgeRules) []FullItem {
	outputItems := []FullItem{*baseItem}

	targetArray := rules.Target()
	sourceArray := rules.Source()

	for _, source := range sourceArray {
		originalValue := baseItem.StatBase().GetUInt(source)
		if originalValue != 0 {
			reforgeQuantity := (originalValue * 4) / 10
			remainQuantity := originalValue - reforgeQuantity

			for _, target := range targetArray {
				if baseItem.StatBase().GetUInt(target) == 0 {
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
		panic("expected different stats")
	}

	originalValue := baseItem.StatBase().GetUInt(source)
	if originalValue == 0 {
		panic("expected item to have source stat " + source.Name() + " on " + baseItem.CreateString())
	}

	if baseItem.StatBase().GetUInt(target) != 0 {
		panic("expected item to have zero target stat " + target.Name() + " on " + baseItem.CreateString())
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
