package solve_highs

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"strconv"
)

// TODO make extended version. e.g. checkSetWithMessage
func validateNewSet(itemSet items.SolvableItemSet, itemOptions *items.SolvableOptionsMap, checkSet func(itemSet *items.SolvableItemSet) bool) {
	itemSet.DebugValidate()
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		if itemOptions.Has(slot) != itemSet.Items().Has(slot) {
			panic("expected slots not filled")
		}
	}

	if !checkSet(&itemSet) {
		sb := util.StringBuild2{}
		sb.WriteString("set fails CheckSet ")
		sb.WriteUint32(itemSet.Total().Hit())
		sb.WriteRune(' ')
		sb.WriteUint32(itemSet.Total().Expertise())
		panic(sb.String())
	}
}

func checkSetRatingIsObjective(solution *util_highs.Solution2, itemSet *items.SolvableItemSet, calcRating func(item *items.SolvableItemSet) float64, ratingScale float64) {
	checkRating := calcRating(itemSet)
	if !util.FloatsApproxEquals(solution.Objective()*ratingScale, checkRating) {
		panic("rating inconsistent " + strconv.FormatFloat(solution.Objective(), 'f', 0, 64) + " " + strconv.FormatFloat(checkRating, 'f', 0, 32))
	}
}

func makeSetPermutations(setData []bonusInfo) []bonusCombo {
	allSetPermutation := make([]bonusCombo, 0, len(setData)*c_maxSetItems) // lower bound on size, not often right
	return makeSetPermutationsRecur(allSetPermutation, setData, 0, 0, []bonusWithCount{})
}

func makeSetPermutationsRecur(allSetPermutation []bonusCombo, setData []bonusInfo, setIndex int, totalCount int, built []bonusWithCount) []bonusCombo {
	if setIndex == len(setData) {
		return append(allSetPermutation, bonusCombo{content: built})
	}

	addSet := setData[setIndex]
	for itemCount := 0; itemCount <= c_maxSetItems; itemCount++ {
		if totalCount+itemCount <= c_maxSetItems {
			next := bonusWithCount{addSet, itemCount}
			progress := util_collection.CopyAndAppend(built, next)
			allSetPermutation = makeSetPermutationsRecur(allSetPermutation, setData, setIndex+1, totalCount+itemCount, progress)
		}
	}

	return allSetPermutation
}
