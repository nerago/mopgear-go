package multi_types

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
)

type CommonCombo map[items.ItemId]*items.FullItem

func CommonCombo_FromProposed(outputs []SingleProposedOutput) CommonCombo {
	combo := make(CommonCombo)
	itemSeen := make(map[items.ItemId]*items.FullItem)

	for index := range outputs {
		for item := range outputs[index].FullSet.Items().AllItemSeq() {
			previousVersion, hasPrevious := itemSeen[item.ItemId()]
			if hasPrevious && previousVersion.Equals(item) {
				combo.AddItem(item.ItemId(), item)
			} else if hasPrevious {
				panic("inconsisent version of item " + item.CreateString())
			} else {
				itemSeen[item.ItemId()] = item
			}
		}
	}

	return combo
}

func (combo *CommonCombo) AddItem(itemId items.ItemId, item *items.FullItem) {
	(*combo)[itemId] = item
}

func (combo *CommonCombo) HasItem(itemId items.ItemId) bool {
	_, has := (*combo)[itemId]
	return has
}

func (combo *CommonCombo) Print(printer *util.PrintRecorder) {
	printer.Println("COMMON_COMBO")
	for _, item := range *combo {
		printer.Printf("COMMON %s\n", item.CreateString())
	}
	for _, item := range *combo {
		if item.Reforge.IsEmpty() {
			printer.Printf("common[%d] = stats.ReforgeRecipe_empty\n", item.ItemId())
		} else {
			printer.Printf("common[%d] = stats.ReforgeRecipe_of(stats.%s, stats.%s)\n", item.ItemId(), item.Reforge.From.EnumName(), item.Reforge.To.EnumName())
		}
	}
}
