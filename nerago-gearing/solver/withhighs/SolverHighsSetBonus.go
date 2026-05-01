package withhighs

import (
	"fmt"
	"math"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/util"

	"github.com/lanl/highs"
)

var g_minusInf = math.Inf(-1)

func RunAllActiveSets(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model) util.Optional[items.SolvableItemSet] {
	constraints := buildConstraintForSetBonus{}
	constraints.init(gear_model)
	constraints.prepareActiveSets(gear_model)
	constraints.addSumRatingVariable()

	for slot, item := range itemOptions.AllItemSlotSeq() {
		constraints.addItem(slot, item, gear_model)
	}

	constraints.finishRows(itemOptions, gear_model)

	constraints.vars.apply(constraints.param)

	highs_model := constraints.param
	solution, err := highs_model.Solve()
	fmt.Println(solution.Status.String())
	if err != nil {
		panic(err)
	}
	fmt.Println(solution.Status.String())
	for i, x := range solution.ColumnPrimal {
		fmt.Println(i, x)
	}

	result := constraints.buildResultSet(&solution.Solution)
	checkSetRating(&solution.Solution, &result, gear_model)
	return util.Optional_OfValue(result)
}

type buildConstraintForSetBonus struct {
	param *highs.RawModel

	vars variableArrayBuilder

	slotsOneEachRow [16]constraintRowSparse // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot

	hitValueRow    constraintRowSparse // constrains values for the hits of each item
	expertValueRow constraintRowSparse // constrains values for the expertise of each item

	baseRatingSumRow constraintRowSparse // values for the ratings of each item
	baseRatingSumVar columnInfo          // sum of values for the ratings of selected items

	setData []setInfo

	itemColumns []columnInfo
}

func (cons *buildConstraintForSetBonus) init(gear_model *gear_model.Model) {
	cons.param = highs.NewRawModel()
	err := cons.param.SetMaximization(true)
	if err != nil {
		panic(err)
	}
}

type setPermutation struct {
	content []setWithCount
}

type setWithCount struct {
	setInfo setInfo
	count   int
}

func (cons *buildConstraintForSetBonus) prepareActiveSets(gear_model *gear_model.Model) {
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
		cons.buildSetMultipliedOutputVar(permutation)
	}

}

const (
	// TODO better values
	c_ratings_low_range  = 1000.0
	c_ratings_high_range = 100000000000.0
	c_positionalShift    = 3
)

func (cons *buildConstraintForSetBonus) buildSetMultipliedOutputs(permutation setPermutation) {
	outputVar := cons.buildSetMultipliedOutputVar(permutation)

}

func (cons *buildConstraintForSetBonus) buildSetMultipliedOutputVar(permutation setPermutation) int {
	totalMultiplier := 1.0
	for _, setAndCount := range permutation.content {
		bonusForCount := setAndCount.setInfo.activeSet.BonusForCount(uint8(setAndCount.count))
		totalMultiplier *= float64(bonusForCount)
	}
	// the actual output variable from this permutation, applies relevant set related multipliers
	return cons.vars.add(highs.ContinuousType, 0, c_ratings_high_range, totalMultiplier)
}

func (cons *buildConstraintForSetBonus) buildPermutationActivatingVar(permutation setPermutation) int {
	activeBool := cons.vars.add(highs.IntegerType, 0, 1, 0)
	activatingRow := constraintRowSparse{}
	activatingRow.add(activeBool, -1)

	// we setup a positional code with shifted components, in such a way set counts can't be added to each other and active reversed combinations
	positionalMultiplier := 1
	positionalFullCode := 0
	for _, setAndCount := range permutation.content {
		setItemCountVar := setAndCount.setInfo.countSetItemsVar.columnIndex

		activatingRow.add(setItemCountVar, float64(positionalMultiplier))

		positionalMultiplier <<= c_positionalShift
		positionalFullCode = (positionalFullCode << c_positionalShift) | setAndCount.count
	}

	activatingRow.apply(cons.param, positionalFullCode, positionalFullCode+1)
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

func (cons *buildConstraintForSetBonus) addSetItemCountVariable(info *setInfo) {
	// set item actual count
	// contributes no rating itself
	columnIndex := cons.vars.add(highs.IntegerType, 0, c_maxSetItems, 0)

	// add corresponding -1 to the set item count array, so we can compare value to the sum of items
	info.countSetItemsRow.add(columnIndex, -1)

	// save reference
	entry := columnInfo{entryType: entry_set_items_count, columnIndex: columnIndex, set: info.activeSet}
	info.countSetItemsVar = entry
}

func (cons *buildConstraintForSetBonus) addSetItemsCountExactVariables(info *setInfo) {
	// grab total number of items into constraint
	settingRow := constraintRowSparse{}
	settingRow.add(info.countTotalItemsVar.columnIndex, -1)

	// make a bool for each possible count in range 0..5
	for itemCount := 0; itemCount <= c_maxSetItems; itemCount++ {
		boolColumn := cons.vars.add(highs.IntegerType, 0, 1, 0)

		// plus one on each count so that zero has a value too, will balance out in constraint target
		settingRow.add(boolColumn, float64(itemCount+1))

		entry := columnInfo{entryType: entry_set_items_exact, columnIndex: boolColumn, set: info.activeSet, itemCount: itemCount}
		info.countExactItemsVars[itemCount] = entry
	}

	settingRow.apply(cons.param, 1, 1)
}

func (cons *buildConstraintForSetBonus) addSumRatingVariable() {
	// sum of individual selected item ratings
	// doesen't go directly into output rating
	columnIndex := cons.vars.add(highs.ContinuousType, 0, math.Inf(1), 0)

	// main action of this variable: derive value to match rest of rest of row sum
	cons.baseRatingSumRow.add(columnIndex, -1)

	// save reference
	entry := columnInfo{entryType: entry_sum_rating, columnIndex: columnIndex}
	cons.baseRatingSumVar = entry
}

func (cons *buildConstraintForSetBonus) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, gear_model *gear_model.Model) {
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

func (cons *buildConstraintForSetBonus) finishItems(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model) {
	// constrain: exactly one item for each slot
	for slot, row := range cons.slotsOneEachRow {
		if itemOptions.Has(items.SlotEquip(slot)) {
			row.apply(cons.param, 1, 1)
		} else {
			row.apply(cons.param, 0, 0)
		}
	}

	// constrain: total sum of hit/exp are within requested limits
	cons.hitValueRow.apply(cons.param, float64(gear_model.StatRequirements.HitMin()), float64(gear_model.StatRequirements.HitMax()))
	cons.expertValueRow.apply(cons.param, float64(gear_model.StatRequirements.ExpertMin()), float64(gear_model.StatRequirements.ExpertMax()))

	// constrain: matching sum to individual ratings
	cons.baseRatingSumRow.apply(cons.param, 0, 0)

	// constrain: matching number of items from each given set
	for _, setInfo := range cons.setData {
		setInfo.countSetItemsRow.apply(cons.param, 0, 0)
	}
}

func (cons *buildConstraintForSetBonus) buildResultSet(solution *highs.Solution) items.SolvableItemSet {
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

// not sure if necessary but I don't trust floats
func floatEqualsOne(value float64) bool {
	return 0.999999 <= value && value <= 1.000001
}

func checkSetRating(solution *highs.Solution, itemSet *items.SolvableItemSet, gear_model *gear_model.Model) {
	checkRating := gear_model.CalcRatingSolveAsFloat(itemSet)
	if solution.Objective != float64(checkRating) {
		panic("rating inconsistent")
	}
}

func contraintIfBoolCopyValueOrZero(param *highs.RawModel, boolSwitchVar, sourceVar, targetVar int, rangeLow, rangeHigh float64) {
	// based on https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d

	// m.AddDenseRow(minusInf, []float64{-1, 0, 1, high, 0}, high)
	// m.AddDenseRow(minusInf, []float64{1, 0, -1, -low, 0}, -low)
	// m.AddDenseRow(minusInf, []float64{1, 0, 0, -high, 0}, 0)
	// m.AddDenseRow(minusInf, []float64{-1, 0, 0, low, 0}, 0)

	valueHigh := constraintRowSparse{}
	valueHigh.add(targetVar, -1)
	valueHigh.add(sourceVar, 1)
	valueHigh.add(boolSwitchVar, rangeHigh)
	valueHigh.apply(param, g_minusInf, rangeHigh)

	valueLow := constraintRowSparse{}
	valueLow.add(targetVar, 1)
	valueLow.add(sourceVar, -1)
	valueLow.add(boolSwitchVar, -rangeLow)
	valueLow.apply(param, g_minusInf, -rangeLow)

	zeroHigh := constraintRowSparse{}
	zeroHigh.add(targetVar, 1)
	zeroHigh.add(boolSwitchVar, -rangeHigh)
	zeroHigh.apply(param, g_minusInf, 0)

	zeroLow := constraintRowSparse{}
	zeroLow.add(targetVar, -1)
	zeroLow.add(boolSwitchVar, rangeLow)
	zeroLow.apply(param, g_minusInf, 0)
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
