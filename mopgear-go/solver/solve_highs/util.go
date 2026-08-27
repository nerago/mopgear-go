package solve_highs

import (
	"errors"

	"github.com/nerago/mopgear-go/items"
)

func validateNewSet(itemSet items.SolvableItemSet, itemOptions *items.SolvableOptionsMap, checkSet func(itemSet *items.SolvableItemSet) (bool, string)) error {
	itemSet.DebugValidate()
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		if itemOptions.Has(slot) != itemSet.Items().Has(slot) {
			return errors.New("expected slots not filled")
		}
	}

	if isOk, message := checkSet(&itemSet); !isOk {
		return errors.New("set fails CheckSet " + message)
	}
}
