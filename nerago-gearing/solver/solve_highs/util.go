package solve_highs

import (
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

func validateNewSet(itemSet items.SolvableItemSet, itemOptions *items.SolvableOptionsMap, model *gear_model.SpecModel) {
	itemSet.DebugValidate()
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		if itemOptions.Has(slot) != itemSet.Items().Has(slot) {
			panic("expected slots not filled")
		}
	}

	if !model.CheckSet(&itemSet) {
		sb := util.StringBuild2{}
		sb.WriteString("set fails CheckSet ")
		sb.WriteUint32(itemSet.Total().Hit())
		sb.WriteRune(' ')
		sb.WriteUint32(itemSet.Total().Expertise())
		panic(sb.String())
	}
}

func checkSetRatingIsObjective(solution *highs.Solution, itemSet *items.SolvableItemSet, gear_model *gear_model.SpecModel) {
	checkRating := gear_model.CalcRatingSolve(itemSet)
	if !util.FloatsApproxEquals(solution.Objective*c_scaled_ratings, float64(checkRating)) {
		panic("rating inconsistent " + strconv.FormatFloat(solution.Objective, 'f', 0, 64) + " " + strconv.FormatFloat(float64(checkRating), 'f', 0, 32))
	}
}

func makeSetPermutations(setData []setInfo) []setPermutation {
	allSetPermutation := make([]setPermutation, 0, len(setData)*c_maxSetItems) // lower bound on size, not often right
	return makeSetPermutationsRecur(allSetPermutation, setData, 0, 0, []setWithCount{})
}

func makeSetPermutationsRecur(allSetPermutation []setPermutation, setData []setInfo, setIndex int, totalCount int, built []setWithCount) []setPermutation {
	if setIndex == len(setData) {
		return append(allSetPermutation, setPermutation{content: built})
	}

	addSet := setData[setIndex]
	for itemCount := 0; itemCount <= c_maxSetItems; itemCount++ {
		if totalCount+itemCount <= c_maxSetItems {
			next := setWithCount{addSet, itemCount}
			progress := util.CopyAndAppend(built, next)
			allSetPermutation = makeSetPermutationsRecur(allSetPermutation, setData, setIndex+1, totalCount+itemCount, progress)
		}
	}

	return allSetPermutation
}
