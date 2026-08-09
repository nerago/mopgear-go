package solve_highs

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_highs"
	"strconv"
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

func checkSetRatingIsObjective(solution *util_highs.Solution2, itemSet *items.SolvableItemSet, calcRating func(item *items.SolvableItemSet) float64, ratingScale float64) {
	checkRating := calcRating(itemSet)
	if !util.FloatsApproxEquals(solution.Objective()*ratingScale, checkRating) {
		panic("rating inconsistent " + strconv.FormatFloat(solution.Objective(), 'f', 0, 64) + " " + strconv.FormatFloat(checkRating, 'f', 0, 32))
	}
}
