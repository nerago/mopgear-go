package solve_highs

import (
	"github.com/nerago/mopgear-go/items"
)

func validateNewSet(itemSet items.SolvableItemSet, itemOptions *items.SolvableOptionsMap, checkSet func(itemSet *items.SolvableItemSet) (bool, string)) {
	itemSet.DebugValidate()
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		if itemOptions.Has(slot) != itemSet.Items().Has(slot) {
			panic("expected slots not filled")
		}
	}

	if isOk, message := checkSet(&itemSet); !isOk {
		panic("set fails CheckSet " + message)
	}
}
