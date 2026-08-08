package multi_types

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
)

type CommonCombo map[items.ItemRef]*items.FullItem

func CommonCombo_FromProposed(outputs map[string]SingleProposedOutput) CommonCombo {
	combo := make(CommonCombo)
	itemSeen := make(map[items.ItemRef]*items.FullItem)

	for _, single := range outputs {
		for item := range single.FullSet.Items().AllItemSeq() {
			ref := items.ItemRef_Of(item)
			previousVersion, hasPrevious := itemSeen[ref]
			if hasPrevious && previousVersion.Equals(item) {
				combo.AddItem(ref, item)
			} else if hasPrevious {
				panic("inconsistent version of item " + item.CreateString())
			} else {
				itemSeen[ref] = item
			}
		}
	}

	return combo
}

func (combo *CommonCombo) AddItem(itemRef items.ItemRef, item *items.FullItem) {
	(*combo)[itemRef] = item
}

func (combo *CommonCombo) HasItem(itemRef items.ItemRef) bool {
	_, has := (*combo)[itemRef]
	return has
}

func (combo *CommonCombo) Print(printer *util.PrintRecorder) {
	printer.Println("COMMON_COMBO")
	for _, item := range *combo {
		printer.Printf("COMMON %s\n", item.CreateString())
	}
	for _, item := range *combo {
		if item.Reforge().IsEmpty() {
			printer.Printf("common[%d] = stats.ReforgeRecipe_empty\n", item.ItemId())
		} else {
			printer.Printf("common[%d] = stats.ReforgeRecipe_of(stats.%s, stats.%s)\n", item.ItemId(), item.Reforge().From.EnumName(), item.Reforge().To.EnumName())
		}
	}
}
