package solve_highs

import (
	"math"

	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_highs"
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

func assertColumnRangeSmallerThanBigM(build *util_highs.LinearBuilder, columnIndex util_highs.ColumnIndex, bigM float64) {
	lo, hi := build.GetColumnMinMax(columnIndex)
	if math.IsInf(lo, 0) || math.IsInf(hi, 0) {
		panic("column is inf, need to set better range")
	}

	diff := hi - lo
	if diff > bigM {
		panic("value range is bigger than the bigM we're using, will be inefficient")
	}
}

func assertColumnBoolOrLess(build *util_highs.LinearBuilder, columnIndex util_highs.ColumnIndex) {
	lo, hi := build.GetColumnMinMax(columnIndex)
	if (lo == 0.0 && hi == 1.0) || (lo == 0.0 && hi == 0.0) || (lo == 1.0 && hi == 1.0) {
		// ok
	} else {
		panic("not a bool column")
	}
}
