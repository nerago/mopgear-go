package withhighs

import (
	"fmt"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/util"

	"github.com/lanl/highs"
)

func RunAllActiveSets(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model) util.Optional[items.SolvableItemSet] {
	constraints := buildInputsForSetBonus{}
	constraints.prepareActiveSets(gear_model)
	constraints.addSumRatingVariable()

	for slot, item := range itemOptions.AllItemSlotSeq() {
		constraints.addItem(slot, item, gear_model)
	}

	constraints.finishItems(itemOptions, gear_model)

	highs_model := constraints.toHighsModel()
	solution, err := highs_model.Solve()
	fmt.Println(solution.Status.String())
	if err != nil {
		panic(err)
	}

	fmt.Println(solution.Status.String())
	for i, x := range solution.ColumnPrimal {
		fmt.Println(i, x)
	}

	if solution.Status != highs.Optimal && solution.Status != highs.ObjectiveBound && solution.Status != highs.ObjectiveTarget {
		return util.Optional_Empty[items.SolvableItemSet]()
	}

	result := constraints.buildResultSet(&solution.Solution)
	checkSetRating(&solution.Solution, &result, gear_model)
	return util.Optional_OfValue(result)
}

type buildInputsForSetBonus struct {
	mat  constraintMatrixBuilder
	vars variableArrayBuilder

	slotsOneEachRow [16]constraintRowSparse // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot

	hitValueRow    constraintRowSparse // constrains values for the hits of each item
	expertValueRow constraintRowSparse // constrains values for the expertise of each item

	baseRatingSumRow constraintRowSparse // values for the ratings of each item
	baseRatingSumVar columnInfo          // sum of values for the ratings of selected items

	setData []setInfo

	itemColumns []columnInfo
}

func (cons *buildInputsForSetBonus) toHighsModel() *highs.RawModel {
	model := highs.NewRawModel()

	model.SetBoolOption("log_to_console", true)
	model.SetIntOption("log_dev_level", 3)

	err := model.SetMaximization(true)
	if err != nil {
		panic(err)
	}

	cons.vars.applyToModel(model)
	cons.mat.finishAndApplyToModelEfficient(model)

	return model
}

type setPermutation struct {
	content []setWithCount
}

type setWithCount struct {
	setInfo setInfo
	count   int
}

func (cons *buildInputsForSetBonus) prepareActiveSets(gear_model *gear_model.Model) {
	// constrain: exact item count in each active set
	activeSets := gear_model.SetBonus.ActiveSets()
	if len(activeSets) == 0 {
		panic("shouldn't be in this solver without active sets")
	}

	cons.setData = make([]setInfo, len(activeSets))
	for setIndex, set := range activeSets {
		info := setInfo{activeSet: set, setIndex: setIndex}
		cons.addSetItemCountVariable(&info)
		cons.addSetItemsCountExactVariables(&info)
		cons.setData[setIndex] = info
	}

	allSetPermutation := makeSetPermutations(cons.setData)
	for _, permutation := range allSetPermutation {
		cons.buildSetMultipliedOutput(permutation)
	}

}

const (
	// TODO better values
	c_ratings_low_range  = 1000.0
	c_ratings_high_range = 100000000000.0
)

func (cons *buildInputsForSetBonus) buildSetMultipliedOutput(permutation setPermutation) {
	outputVar := cons.buildSetMultipliedOutputVar(permutation)
	activatingBool := cons.buildPermutationActivatingVar(permutation)

	contraintIfBoolCopyValueElseZero(&cons.mat, activatingBool, cons.baseRatingSumVar.columnIndex, outputVar, c_ratings_low_range, c_ratings_high_range)
}

func (cons *buildInputsForSetBonus) buildSetMultipliedOutputVar(permutation setPermutation) int {
	totalMultiplier := 1.0
	for _, setAndCount := range permutation.content {
		bonusForCount := setAndCount.setInfo.activeSet.BonusForCount(uint8(setAndCount.count))
		totalMultiplier *= float64(bonusForCount)
	}

	// the actual output variable from this permutation, applies relevant set related multipliers
	return cons.vars.add(highs.ContinuousType, 0, c_ratings_high_range, totalMultiplier)
}

func (cons *buildInputsForSetBonus) buildPermutationActivatingVar(permutation setPermutation) int {
	// we are effecively building a logical AND between these vars

	permutationActiveBool := cons.vars.add(highs.IntegerType, 0, 1, 0)

	buildAnd := contraintAndBuilder{}
	buildAnd.setOutput(permutationActiveBool)

	for _, setAndCount := range permutation.content {
		setInfo := setAndCount.setInfo
		count := setAndCount.count
		specificExactBool := setInfo.countExactItemsVars[count]
		buildAnd.addInput(specificExactBool.columnIndex)
	}

	buildAnd.finishAndApply(&cons.mat)

	return permutationActiveBool
}

func makeSetPermutations(setData []setInfo) []setPermutation {
	allSetPermutation := make([]setPermutation, 0, len(setData)*c_maxSetItems) // lower bound on size, not often right
	return makeSetPermutationsRecur(allSetPermutation, setData, 0, 0, []setWithCount{})
}

func makeSetPermutationsRecur(allSetPermutation []setPermutation, setData []setInfo, setIndex int, totalCount int, built []setWithCount) []setPermutation {
	if setIndex == len(setData) {
		return append(allSetPermutation, setPermutation{built})
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

func (cons *buildInputsForSetBonus) addSetItemCountVariable(info *setInfo) {
	// set item actual count
	columnIndex := cons.vars.add(highs.IntegerType, 0, c_maxSetItems, 0)

	// add corresponding -1 to the set item count array, so we can compare value to the sum of items, relevant items will flag a 1
	info.countSetItemsRow.add(columnIndex, -1)

	// save reference
	entry := columnInfo{entryType: entry_set_items_count, columnIndex: columnIndex, set: info.activeSet}
	info.countTotalItemsVar = entry
}

func (cons *buildInputsForSetBonus) addSetItemsCountExactVariables(info *setInfo) {
	// compare total number of items previous computed into this constraint
	compareRow := constraintRowSparse{}
	compareRow.add(info.countTotalItemsVar.columnIndex, -1)

	// constraint so only one of these flags gets set
	singleFlagOnly := constraintRowSparse{}

	// make a bool for each possible count in range 0..5
	for itemCount := 0; itemCount <= c_maxSetItems; itemCount++ {
		boolColumn := cons.vars.add(highs.IntegerType, 0, 1, 0)

		// should activate this flag which will match the total count
		compareRow.add(boolColumn, float64(itemCount))

		// but only one flag at a time
		singleFlagOnly.add(boolColumn, 1)

		entry := columnInfo{entryType: entry_set_items_exact, columnIndex: boolColumn, set: info.activeSet, itemCount: itemCount}
		info.countExactItemsVars[itemCount] = entry
	}

	compareRow.finish(&cons.mat, 0, 0)     // equal
	singleFlagOnly.finish(&cons.mat, 1, 1) // sum of flags should be just one, should pull the zero flag up if no other set
}

func (cons *buildInputsForSetBonus) addSumRatingVariable() {
	// sum of individual selected item ratings
	// doesen't go directly into output rating
	columnIndex := cons.vars.add(highs.ContinuousType, 0, c_plusInf, 0)

	// main action of this variable: derive value to match rest of rest of row sum
	cons.baseRatingSumRow.add(columnIndex, -1)

	// save reference
	entry := columnInfo{entryType: entry_sum_rating, columnIndex: columnIndex}
	cons.baseRatingSumVar = entry
}

func (cons *buildInputsForSetBonus) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, gear_model *gear_model.Model) {
	// boolean value to flag use of specific item
	// contributes 0 to final rating itself, but via additional summation and calcs
	columnIndex := cons.vars.add(highs.IntegerType, 0, 1, 0)

	// add rating via a summation condition
	rating := float64(gear_model.CalcRatingSolveItemAsFloat(item))
	cons.baseRatingSumRow.add(columnIndex, rating)

	// specific hit/expertise values for hi/lo limits
	cons.hitValueRow.add(columnIndex, float64(item.TotalCap().Hit()))
	cons.expertValueRow.add(columnIndex, float64(item.TotalCap().Expertise()))

	// 1 for that slot that matches the item, so we can tell solver only one item per slot
	cons.slotsOneEachRow[itemSlot].add(columnIndex, 1.0)

	// if this item belongs to any item set then flag with a 1
	activeSetIndex, hasSet := gear_model.SetBonus.ActiveSetIndexForItem(item.ItemId())
	if hasSet {
		cons.setData[activeSetIndex].countSetItemsRow.add(columnIndex, 1)
	}

	entry := columnInfo{entryType: entry_item, columnIndex: columnIndex, itemSlot: itemSlot, item: item}
	cons.itemColumns = append(cons.itemColumns, entry)
}

func (cons *buildInputsForSetBonus) finishItems(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model) {
	// constrain: exactly one item for each slot
	for slot, row := range cons.slotsOneEachRow {
		if itemOptions.Has(items.SlotEquip(slot)) {
			row.finish(&cons.mat, 1, 1)
		} else {
			row.finish(&cons.mat, 0, 0)
		}
	}

	// constrain: total sum of hit/exp are within requested limits
	cons.hitValueRow.finish(&cons.mat, float64(gear_model.StatRequirements.HitMin()), float64(gear_model.StatRequirements.HitMax()))
	cons.expertValueRow.finish(&cons.mat, float64(gear_model.StatRequirements.ExpertMin()), float64(gear_model.StatRequirements.ExpertMax()))

	// constrain: matching sum to individual ratings
	cons.baseRatingSumRow.finish(&cons.mat, 0, 0)

	// constrain: matching number of items from each given set
	for _, setInfo := range cons.setData {
		setInfo.countSetItemsRow.finish(&cons.mat, 0, 0)
	}
}

func (cons *buildInputsForSetBonus) buildResultSet(solution *highs.Solution) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for _, columnEntry := range cons.itemColumns {
		variableResult := solution.ColumnPrimal[columnEntry.columnIndex]
		if columnEntry.entryType == entry_item && floatEqualsOne(variableResult) {
			itemSet.AddItem_DeferCalc_ExpectEmpty(columnEntry.itemSlot, columnEntry.item)
		}
	}
	items.SolvableItemSet_RecalculateTotal(&itemSet)

	return itemSet
}

func checkSetRating(solution *highs.Solution, itemSet *items.SolvableItemSet, gear_model *gear_model.Model) {
	checkRating := gear_model.CalcRatingSolveAsFloat(itemSet)
	if solution.Objective != float64(checkRating) {
		panic("rating inconsistent")
	}
}

type entryType int8

const (
	entry_item            entryType = iota
	entry_set_items_count entryType = iota
	entry_set_items_exact entryType = iota
	entry_sum_rating      entryType = iota
)

type columnInfo struct {
	entryType   entryType
	columnIndex int

	itemSlot  items.SlotEquip
	item      *items.SolvableItem
	set       gear_model.ActiveSet
	itemCount int
}

type setInfo struct {
	activeSet gear_model.ActiveSet
	setIndex  int

	countSetItemsRow    constraintRowSparse          // use to count items used from this set, has 1 or 0 flags
	countTotalItemsVar  columnInfo                   // total count of items used
	countExactItemsVars [c_setItemsCounts]columnInfo // specific bools for different counts
}
