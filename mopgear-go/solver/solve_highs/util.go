package solve_highs

import (
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/util"
)

func validateNewSet(itemSet items.SolvableItemSet, itemOptions *items.SolvableOptionsMap, checkSet func(itemSet *items.SolvableItemSet) (bool, string)) error {
	itemSet.DebugValidate()
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		if itemOptions.Has(slot) != itemSet.Items().Has(slot) {
			return util.ErrorTracedNew("expected slots not filled")
		}
	}

	if isOk, message := checkSet(&itemSet); !isOk {
		return util.ErrorTracedNew("set fails CheckSet " + message)
	}

	return nil
}
