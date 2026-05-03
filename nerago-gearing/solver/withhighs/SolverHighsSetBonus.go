package withhighs

import (
	"fmt"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/util"
	"strconv"

	"github.com/lanl/highs"
)

func RunAllActiveSets(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model) util.Optional[items.SolvableItemSet] {
	inputBuilder := inputBuilder{}
	setup := setupBonusedInputs(&inputBuilder, gear_model, itemOptions, 1)

	highs_model := inputBuilder.toHighsModel()
	solution, err := highs_model.Solve()
	fmt.Println("SOLUTION STATUS = " + solution.Status.String())
	if err != nil {
		panic(err)
	}

	debugPrint(solution, setup)

	if solution.Status != highs.Optimal && solution.Status != highs.ObjectiveBound && solution.Status != highs.ObjectiveTarget {
		return util.Optional_Empty[items.SolvableItemSet]()
	}

	result := setup.buildResultSet(&solution.Solution, itemOptions, gear_model)
	checkSetRating(&solution.Solution, &result, gear_model)
	return util.Optional_OfValue(result)
}

func setupBonusedInputs(inputBuilder *inputBuilder, gear_model *gear_model.Model, itemOptions *items.SolvableOptionsMap, scaleOutputRating float64) *setupInputsForSetBonus {
	setup := setupInputsForSetBonus{input: inputBuilder}

	setup.addFinalOutputVariable(scaleOutputRating)
	setup.addSumRatingVariable()
	setup.prepareActiveSets(gear_model)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		setup.addItem(slot, item, gear_model)
	}

	setup.finishItems(itemOptions, gear_model)
	return &setup
}

func debugPrint(solution *highs.RawSolution, setup *setupInputsForSetBonus) {
	fmt.Println("OBJECTIVE VALUE = ", solution.Objective)

	activeBonus := ""
	activeBonusWeight := 0.0

	for columnIndex, outputValue := range solution.ColumnPrimal {
		var colEntry columnInfo
		found := false
		for _, col := range setup.allColumns {
			if col.columnIndex == columnIndex {
				colEntry = col
				found = true
			}
		}

		if found {
			switch colEntry.entryType {
			case entry_item:
				fmt.Println(columnIndex, outputValue, "item", colEntry.itemSlot.Name(), colEntry.item.ItemId())
			case entry_set_total_count:
				fmt.Println(columnIndex, outputValue, "set total count", colEntry.set.Name())
			case entry_set_exact_count:
				fmt.Println(columnIndex, outputValue, "set exact count flag", colEntry.set.Name(), colEntry.itemCount)
			case entry_sum_rating:
				fmt.Println(columnIndex, outputValue, "initial item rating sum")
			case entry_permutation_active:
				fmt.Println(columnIndex, outputValue, "permutation active", colEntry.permutation.debugStr())
				if floatEqualsOne(outputValue) {
					activeBonus += colEntry.permutation.debugStr()
				}
			case entry_permutation_output_weighted:
				fmt.Println(columnIndex, outputValue, "permutation weighted output", colEntry.permutation.debugStr(), colEntry.weight)
				if !floatEqualsZero(outputValue) {
					activeBonusWeight += colEntry.weight
				}
			case entry_final_output:
				fmt.Println(columnIndex, outputValue, "final value")
			default:
				panic("unknown column")
			}
		} else {
			fmt.Println(columnIndex, outputValue, "NOT FOUND???")
		}
	}

	fmt.Printf("ACTIVE highs Bonus = %s %f\n", activeBonus, activeBonusWeight)
}

type setupInputsForSetBonus struct {
	input *inputBuilder

	slotsOneEachRow [16]constraintRowBuild // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot

	hitValueRow    constraintRowBuild // constrains values for the hits of each item
	expertValueRow constraintRowBuild // constrains values for the expertise of each item

	baseRatingSumRow constraintRowBuild // values for the ratings of each item
	baseRatingSumVar *columnInfo        // sum of values for the ratings of selected items

	finalOutputRow constraintRowBuild // compute final output from set based alternatives
	finalOutputVar *columnInfo        // output variable, to be used directly or scaled against other models

	setData           []setInfo
	allSetPermutation []setPermutation

	itemColumns []columnInfo
	allColumns  []columnInfo
}

type setPermutation struct {
	content []setWithCount

	outputVar     int
	activatingVar int
	weight        float64
}

func (perm setPermutation) debugStr() string {
	build := util.StringBuild2{}
	for _, set := range perm.content {
		build.WriteString(set.setInfo.activeSet.Name())
		build.WriteRune('=')
		build.WriteInt64(int64(set.count))
		build.WriteRune(' ')
	}
	return build.String()
}

type setWithCount struct {
	setInfo setInfo
	count   int
}

func (setup *setupInputsForSetBonus) prepareActiveSets(gear_model *gear_model.Model) {
	// constrain: exact item count in each active set
	activeSets := gear_model.SetBonus.ActiveSets()
	if len(activeSets) > 0 {
		setup.setData = make([]setInfo, len(activeSets))
		for setIndex, set := range activeSets {
			info := setInfo{activeSet: set, setIndex: setIndex}
			setup.addSetItemCountVariable(&info)
			setup.addSetItemsCountExactVariables(&info)
			setup.setData[setIndex] = info
		}

		setup.allSetPermutation = makeSetPermutations(setup.setData)
		for permIndex := range setup.allSetPermutation {
			setup.buildSetMultipliedOutput(&setup.allSetPermutation[permIndex])
		}
	} else {
		setup.buildSimpleNoSetsOutput()
	}
}

const (
	// example rating      178237915
	//                     187513497
	c_ratings_low_range  = 10000000.0
	c_ratings_high_range = 100000000000.0
)

func (setup *setupInputsForSetBonus) buildSimpleNoSetsOutput() {
	// just copy initial rating sum into final if no sets
	setup.finalOutputRow.add(setup.baseRatingSumVar.columnIndex, 1)
}

func (setup *setupInputsForSetBonus) buildSetMultipliedOutput(permutation *setPermutation) {
	outputVar, weight := setup.buildSetWeightedOutputVar(permutation)
	activatingVar := setup.buildPermutationActivatingVar(permutation)

	// copy regular rating sum to column if flag is set
	contraintIfBoolCopyValueElseZero(setup.input, activatingVar, setup.baseRatingSumVar.columnIndex, outputVar, c_ratings_low_range, c_ratings_high_range)

	// add scaled rating to final computation
	setup.finalOutputRow.add(outputVar, weight)

	permutation.outputVar = outputVar
	permutation.activatingVar = activatingVar
	permutation.weight = weight
}

func (setup *setupInputsForSetBonus) buildSetWeightedOutputVar(permutation *setPermutation) (int, float64) {
	totalWeight := 1.0
	for _, setAndCount := range permutation.content {
		bonusForCount := setAndCount.setInfo.activeSet.BonusForCount(uint8(setAndCount.count))
		totalWeight *= float64(bonusForCount)
	}

	// the actual output variable from this permutation, applies relevant set related multipliers
	columnIndex := setup.input.createColumnGeneral(highs.ContinuousType, c_minusInf, c_plusInf)
	entry := columnInfo{entryType: entry_permutation_output_weighted, columnIndex: columnIndex, permutation: permutation, weight: totalWeight}
	setup.allColumns = append(setup.allColumns, entry)
	return columnIndex, totalWeight
}

func (setup *setupInputsForSetBonus) buildPermutationActivatingVar(permutation *setPermutation) int {
	// we are effecively building a logical AND between these vars

	permutationActiveBool := setup.input.createColumnBool()

	buildAnd := contraintAndBuilder{}
	buildAnd.setOutput(permutationActiveBool)

	for _, setAndCount := range permutation.content {
		setInfo := setAndCount.setInfo
		count := setAndCount.count
		specificExactBool := setInfo.setExactCountVars[count]
		buildAnd.addInput(specificExactBool.columnIndex)
	}

	buildAnd.finishAndApply(setup.input)

	entry := columnInfo{entryType: entry_permutation_active, columnIndex: permutationActiveBool, permutation: permutation}
	setup.allColumns = append(setup.allColumns, entry)

	return permutationActiveBool
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

func (setup *setupInputsForSetBonus) addSetItemCountVariable(info *setInfo) {
	// set item actual count
	columnIndex := setup.input.createColumnGeneral(highs.IntegerType, 0, c_maxSetItems)

	// add corresponding -1 to the set item count array, so we can compare value to the sum of items, relevant items will flag a 1
	info.countSetItemsRow.add(columnIndex, -1)

	// save reference
	entry := columnInfo{entryType: entry_set_total_count, columnIndex: columnIndex, set: info.activeSet}
	info.setTotalCountVar = &entry
	setup.allColumns = append(setup.allColumns, entry)
}

func (setup *setupInputsForSetBonus) addSetItemsCountExactVariables(info *setInfo) {
	// compare total number of items previous computed into this constraint
	compareRow := constraintRowBuild{}
	compareRow.add(info.setTotalCountVar.columnIndex, -1)

	// constraint so only one of these flags gets set
	singleFlagOnly := constraintRowBuild{}

	// make a bool for each possible count in range 0..5
	for itemCount := 0; itemCount <= c_maxSetItems; itemCount++ {
		boolColumn := setup.input.createColumnBool()

		// should activate this flag which will match the total count
		compareRow.add(boolColumn, float64(itemCount))

		// but only one flag at a time
		singleFlagOnly.add(boolColumn, 1)

		entry := columnInfo{entryType: entry_set_exact_count, columnIndex: boolColumn, set: info.activeSet, itemCount: itemCount}
		info.setExactCountVars[itemCount] = entry
		setup.allColumns = append(setup.allColumns, entry)
	}

	compareRow.finish(setup.input, 0, 0)     // equal
	singleFlagOnly.finish(setup.input, 1, 1) // sum of flags should be just one, should pull the zero flag up if no other set
}

func (setup *setupInputsForSetBonus) addFinalOutputVariable(scaleOutputRating float64) {
	// goes directly into overall rating, but could have an external scale applied
	columnIndex := setup.input.createColumnWithOutput(highs.ContinuousType, 0, c_plusInf, scaleOutputRating)

	// derive value based on whichever setup bonus permutation is active
	setup.finalOutputRow.add(columnIndex, -1)

	// save reference
	entry := columnInfo{entryType: entry_final_output, columnIndex: columnIndex}
	setup.finalOutputVar = &entry
	setup.allColumns = append(setup.allColumns, entry)
}

func (setup *setupInputsForSetBonus) addSumRatingVariable() {
	// sum of individual selected item ratings
	// doesen't go directly into output rating
	columnIndex := setup.input.createColumnGeneral(highs.ContinuousType, 0, c_plusInf)

	// main action of this variable: derive value to match rest of rest of row sum
	setup.baseRatingSumRow.add(columnIndex, -1)

	// save reference
	entry := columnInfo{entryType: entry_sum_rating, columnIndex: columnIndex}
	setup.baseRatingSumVar = &entry
	setup.allColumns = append(setup.allColumns, entry)
}

func (setup *setupInputsForSetBonus) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, gear_model *gear_model.Model) {
	// boolean value to flag use of specific item
	// contributes 0 to final rating itself, but via additional summation and calcs
	columnIndex := setup.input.createColumnBool()

	// add rating via a summation condition
	rating := float64(gear_model.CalcRatingSolveItemAsFloat(item))
	setup.baseRatingSumRow.add(columnIndex, rating)

	// specific hit/expertise values for hi/lo limits
	setup.hitValueRow.add(columnIndex, float64(item.TotalCap().Hit()))
	setup.expertValueRow.add(columnIndex, float64(item.TotalCap().Expertise()))

	// 1 for that slot that matches the item, so we can tell solver only one item per slot
	setup.slotsOneEachRow[itemSlot].add(columnIndex, 1.0)

	// if this item belongs to any item set then flag with a 1
	activeSetIndex, hasSet := gear_model.SetBonus.ActiveSetIndexForItem(item.ItemId())
	if hasSet {
		setup.setData[activeSetIndex].countSetItemsRow.add(columnIndex, 1)
	}

	entry := columnInfo{entryType: entry_item, columnIndex: columnIndex, itemSlot: itemSlot, item: item}
	setup.itemColumns = append(setup.itemColumns, entry)
	setup.allColumns = append(setup.allColumns, entry)
}

func (setup *setupInputsForSetBonus) finishItems(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model) {
	// constrain: exactly one item for each slot
	for slot, row := range setup.slotsOneEachRow {
		if itemOptions.Has(items.SlotEquip(slot)) {
			row.finish(setup.input, 1, 1)
		} else {
			row.finish(setup.input, 0, 0)
		}
	}

	// constrain: total sum of hit/exp are within requested limits
	setup.hitValueRow.finish(setup.input, float64(gear_model.StatRequirements.HitMin()), float64(gear_model.StatRequirements.HitMax()))
	setup.expertValueRow.finish(setup.input, float64(gear_model.StatRequirements.ExpertMin()), float64(gear_model.StatRequirements.ExpertMax()))

	// constrain: matching sum to individual ratings
	setup.baseRatingSumRow.finish(setup.input, 0, 0)

	// constrain: matching number of items from each given set
	for _, setInfo := range setup.setData {
		setInfo.countSetItemsRow.finish(setup.input, 0, 0)
	}

	// constrain: whichever alternate set output into final
	setup.finalOutputRow.finish(setup.input, 0, 0)
}

func (setup *setupInputsForSetBonus) buildResultSet(solution *highs.Solution, itemOptions *items.SolvableOptionsMap, model *gear_model.Model) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for _, columnEntry := range setup.itemColumns {
		variableResult := solution.ColumnPrimal[columnEntry.columnIndex]
		if columnEntry.entryType == entry_item && floatEqualsOne(variableResult) {
			itemSet.AddItem_DeferCalc_ExpectEmpty(columnEntry.itemSlot, columnEntry.item)
		}
	}
	items.SolvableItemSet_RecalculateTotal(&itemSet)

	validateNewSet(itemSet, itemOptions, model)

	setup.checkActivePermutation(solution, &itemSet)

	return itemSet
}

func checkSetRating(solution *highs.Solution, itemSet *items.SolvableItemSet, gear_model *gear_model.Model) {
	checkRating := gear_model.CalcRatingSolveAsFloat(itemSet)
	if !floatsApproxEquals(solution.Objective, float64(checkRating)) {
		panic("rating inconsistent " + strconv.FormatFloat(solution.Objective, 'f', 0, 64) + " " + strconv.FormatFloat(float64(checkRating), 'f', 0, 32))
	}
}

func (setup *setupInputsForSetBonus) checkActivePermutation(solution *highs.Solution, solvableItemSet *items.SolvableItemSet) {
	if len(setup.allSetPermutation) > 0 {
		var activePermutation *setPermutation

		for _, permutation := range setup.allSetPermutation {
			variableResult := solution.ColumnPrimal[permutation.activatingVar]
			if floatEqualsOne(variableResult) {
				if activePermutation == nil {
					activePermutation = &permutation
				} else {
					panic("multiple permutations active")
				}
			}
		}

		if activePermutation != nil {
			for _, entry := range activePermutation.content {
				activeSet := entry.setInfo.activeSet
				actualCount := activeSet.CountItems(solvableItemSet.Items())
				if actualCount != uint8(entry.count) {
					panic("number of items not what solver returned")
				}
			}
		} else {
			panic("no permutation active")
		}
	}
}

type entryType int8

const (
	entry_item                        entryType = iota
	entry_set_total_count             entryType = iota
	entry_set_exact_count             entryType = iota
	entry_sum_rating                  entryType = iota
	entry_permutation_active          entryType = iota
	entry_permutation_output_weighted entryType = iota
	entry_final_output                entryType = iota
)

type columnInfo struct {
	entryType   entryType
	columnIndex int

	itemSlot items.SlotEquip
	item     *items.SolvableItem

	set         gear_model.ActiveSet
	itemCount   int
	permutation *setPermutation
	weight      float64
}

type setInfo struct {
	activeSet gear_model.ActiveSet
	setIndex  int

	countSetItemsRow  constraintRowBuild           // use to count items used from this set, has 1 or 0 flags
	setTotalCountVar  *columnInfo                  // total count of items used
	setExactCountVars [c_setItemsCounts]columnInfo // specific bools for different counts
}
