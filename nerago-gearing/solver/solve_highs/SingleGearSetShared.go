package solve_highs

import (
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

type singleGearSetShared struct {
	build *util_highs.LinearBuilder

	slotsOneEachRow [items.ITEM_SLOT_COUNT]util_highs.ConstraintRow // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot

	mainOutputRow util_highs.ConstraintRow // compute final output from set based alternatives
	mainOutputVar *columnInfo              // output variable, to be used directly or scaled against other models

	uniqueEquipRowsById map[items.ItemId]*util_highs.ConstraintRow // lookup by id, may have multiple mappings for an item so need pointers
	uniqueEquipRowsAll  []*util_highs.ConstraintRow                // definitive copy of each unique equip row constraint

	bonusData   []bonusInfo
	bonusCombos []bonusCombo

	itemColumns util_collection.MapSlice[items.ItemId, *columnInfo]
	allColumns  []*columnInfo
}

func (setup *singleGearSetShared) addItemCommon(itemSlot items.SlotEquip, item *items.SolvableItem, setBonus *gear_model.SetBonus) util_highs.ColumnIndex {
	entry := columnInfo{entryType: entry_item, itemSlot: itemSlot, item: item}

	// boolean value to flag use of specific item, in exact reforge/gem state
	columnIndex := setup.build.CreateColumnBool(&entry)
	entry.columnIndex = columnIndex
	setup.allColumns = append(setup.allColumns, &entry)
	setup.itemColumns.Add(item.ItemId(), &entry)

	// 1 for that slot that matches the item, so we can tell solver only one item per slot
	setup.slotsOneEachRow[itemSlot].Add(columnIndex, 1.0)

	// if this item belongs to any item set then flag with a 1
	activeSetIndex, hasSet := setBonus.ActiveSetIndexForItem(item.ItemId())
	if hasSet {
		setup.bonusData[activeSetIndex].countSetItemsRow.Add(columnIndex, 1)
	}

	// if this item is unique equipped (mostly checked for ring/trinket)
	uniqueRow := setup.uniqueEquipRowsById[item.ItemId()]
	if uniqueRow != nil {
		uniqueRow.Add(columnIndex, 1)
	}

	return columnIndex
}

func (setup *singleGearSetShared) finishItemsCommon(itemOptions *items.SolvableOptionsMap) {
	// constrain: exactly one item for each slot
	for slot, row := range setup.slotsOneEachRow {
		slotEquip := items.SlotEquip(slot)
		row.Debug = "slotsOneEachRow_" + slotEquip.Name()
		if itemOptions.Has(slotEquip) {
			row.Build(setup.build, 1, 1)
		} else {
			row.Build(setup.build, 0, 0)
		}
	}

	// constrain: matching number of items from each given set
	for _, setInfo := range setup.bonusData {
		setInfo.countSetItemsRow.Build(setup.build, 0, 0)
	}

	// constrain: unique item by itemid/unique set
	for _, row := range setup.uniqueEquipRowsAll {
		if !row.IsEmpty() {
			row.Build(setup.build, 0, 1)
		}
	}
}

func (setup *singleGearSetShared) prepareActiveSetCombos(setBonus *gear_model.SetBonus) {
	// constrain: exact item count in each active set
	activeSets := setBonus.ActiveSets()
	if len(activeSets) > 0 {
		setup.bonusData = make([]bonusInfo, len(activeSets))
		for setIndex, set := range activeSets {
			info := bonusInfo{activeSet: set, setIndex: setIndex}
			info.countSetItemsRow.Debug = "countSetItemsRow" + strconv.Itoa(setIndex)
			setup.addSetItemCountVariable(&info)
			setup.addSetItemsCountExactVariables(&info)
			setup.bonusData[setIndex] = info
		}
	}
}

func (setup *singleGearSetShared) multiplyRatingsByActiveSetCombo(setBonus *gear_model.SetBonus, combinedRatingVar *columnInfo) {
	activeSets := setBonus.ActiveSets()
	if len(activeSets) > 0 {
		setup.bonusCombos = makeSetPermutations(setup.bonusData)
		for permIndex := range setup.bonusCombos {
			setup.buildSetMultipliedOutput(&setup.bonusCombos[permIndex], combinedRatingVar)
		}
	} else {
		setup.buildSimpleNoSetsOutput(combinedRatingVar)
	}

	// constrain: whichever alternate set output into final
	setup.mainOutputRow.Debug = "mainOutputRow"
	setup.mainOutputRow.Build(setup.build, 0, 0)
}

func (setup *singleGearSetShared) buildSimpleNoSetsOutput(combinedRatingVar *columnInfo) {
	// just copy initial rating sum into final if no sets
	setup.mainOutputRow.Add(combinedRatingVar.columnIndex, 1)
}

func (setup *singleGearSetShared) buildSetMultipliedOutput(permutation *bonusCombo, combinedRatingVar *columnInfo) {
	activatingVar := setup.buildPermutationActivatingVar(permutation)

	totalWeight := 1.0
	for _, setAndCount := range permutation.content {
		bonusForCount := setAndCount.setInfo.activeSet.BonusForCount(uint8(setAndCount.count))
		totalWeight *= bonusForCount
	}

	// the actual output variable from this permutation, applies relevant set related multipliers
	permutationOutput := columnInfo{entryType: entry_permutation_output_weighted, permutation: permutation, weight: totalWeight}
	permutationOutput.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, &permutationOutput)
	setup.allColumns = append(setup.allColumns, &permutationOutput)

	// copy regular rating sum to column if flag is set
	setup.build.ConstraintIfBoolCopyValueElseZero(activatingVar.columnIndex, combinedRatingVar.columnIndex, permutationOutput.columnIndex, c_ratings_low_range, c_ratings_high_range)

	// add scaled rating to final computation
	setup.mainOutputRow.Add(permutationOutput.columnIndex, totalWeight)

	permutation.outputVar = &permutationOutput
	permutation.activatingVar = activatingVar
	permutation.weight = totalWeight
}

func (setup *singleGearSetShared) addMainOutputVariable(scaleOutputRating float64) {
	entry := columnInfo{entryType: entry_main_output}

	// goes directly into overall rating, but could have an external scale applied
	entry.columnIndex = setup.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, scaleOutputRating, &entry)

	// derive value based on whichever setup bonus permutation is active
	setup.mainOutputRow.Add(entry.columnIndex, -1)

	// save reference
	setup.mainOutputVar = &entry
	setup.allColumns = append(setup.allColumns, &entry)
}

func (setup *singleGearSetShared) addSetNeededCounts(setBonusRequired []gear_model.ActiveSetCountsRequired) {
	if len(setBonusRequired) > 0 {
		if len(setup.bonusData) == 0 {
			panic("no setdata to use from SetBonusRequired")
		} else if len(setup.bonusData) == 1 && len(setBonusRequired) == 1 && setBonusRequired[0].Count() == 1 {
			setCountCol := setup.bonusData[0].setTotalCountVar
			_, needCount := setBonusRequired[0].PairsByIndex(0)

			rowSetCountRequired := util_highs.ConstraintRow{Debug: "rowSetCountRequired"}
			rowSetCountRequired.Add(setCountCol.columnIndex, 1)
			rowSetCountRequired.Build(setup.build, float64(needCount), util_highs.C_PlusInf)
		} else {
			oneOfTheseOptions := util_highs.ConstraintRow{}

			for _, option := range setBonusRequired {
				optionParts := util_highs.ConstraintAndBuilder{}

				for activeSet, needCount := range option.Pairs() {
					setInfo := util_collection.FindWith(setup.bonusData, func(x bonusInfo) bool { return x.activeSet.Equals(activeSet) })
					setCountCol := setInfo.setTotalCountVar

					inRange := setup.build.ColumnIsGreaterOrEqualThanConstant(setCountCol.columnIndex, float64(needCount), 10, 1.0)
					optionParts.AddInput(inRange)
				}

				optionActive := setup.build.CreateColumnBool(util_highs.DebugText("SetBonusRequired option"))
				optionParts.SetOutput(optionActive)
				optionParts.Build(setup.build)

				oneOfTheseOptions.Add(optionActive, 1)
			}

			oneOfTheseOptions.Build(setup.build, 1, util_highs.C_PlusInf)
		}
	}
}

func (setup *singleGearSetShared) buildPermutationActivatingVar(permutation *bonusCombo) *columnInfo {
	// we are effectively building a logical AND between these vars

	permutationActiveBool := columnInfo{entryType: entry_permutation_active, permutation: permutation}
	permutationActiveBool.columnIndex = setup.build.CreateColumnBool(&permutationActiveBool)

	buildAnd := util_highs.ConstraintAndBuilder{}
	buildAnd.SetOutput(permutationActiveBool.columnIndex)

	for _, setAndCount := range permutation.content {
		setInfo := setAndCount.setInfo
		count := setAndCount.count
		specificExactBool := setInfo.setExactCountVars[count]
		buildAnd.AddInput(specificExactBool.columnIndex)
	}

	buildAnd.Build(setup.build)

	setup.allColumns = append(setup.allColumns, &permutationActiveBool)

	return &permutationActiveBool
}

func (setup *singleGearSetShared) addSetItemCountVariable(info *bonusInfo) {
	entry := columnInfo{entryType: entry_set_total_count, set: info.activeSet}

	// set item actual count
	entry.columnIndex = setup.build.CreateColumnGeneral(highs.Integer, 0, c_maxSetItems, &entry)

	// add corresponding -1 to the set item count array, so we can compare value to the sum of items, relevant items will flag a 1
	info.countSetItemsRow.Add(entry.columnIndex, -1)

	// save reference

	info.setTotalCountVar = &entry
	setup.allColumns = append(setup.allColumns, &entry)
}

func (setup *singleGearSetShared) addSetItemsCountExactVariables(info *bonusInfo) {
	// compare total number of items previous computed into this constraint
	compareRow := util_highs.ConstraintRow{Debug: "setItemsCompareRow"}
	compareRow.Add(info.setTotalCountVar.columnIndex, -1)

	// constraint so only one of these flags gets set
	singleFlagOnly := util_highs.ConstraintRow{Debug: "setItemsSingleFlagOnly"}

	// make a bool for each possible count in range 0..5
	for itemCount := 0; itemCount <= c_maxSetItems; itemCount++ {
		boolColumn := columnInfo{entryType: entry_set_exact_count, set: info.activeSet, itemCount: itemCount}
		boolColumn.columnIndex = setup.build.CreateColumnBool(&boolColumn)

		// should activate this flag which will match the total count
		compareRow.Add(boolColumn.columnIndex, float64(itemCount))

		// but only one flag at a time
		singleFlagOnly.Add(boolColumn.columnIndex, 1)

		info.setExactCountVars[itemCount] = &boolColumn
		setup.allColumns = append(setup.allColumns, &boolColumn)
	}

	compareRow.Build(setup.build, 0, 0)     // equal
	singleFlagOnly.Build(setup.build, 1, 1) // sum of flags should be just one, should pull the zero flag up if no other set
}

func (setup *singleGearSetShared) prepareUniqueEquipped(itemOptions *items.SolvableOptionsMap) {
	setup.uniqueEquipRowsById = make(map[items.ItemId]*util_highs.ConstraintRow)
	setup.uniqueEquipRowsAll = make([]*util_highs.ConstraintRow, 0)
	seen := make(map[items.ItemId]bool)

	// add items from predefined unique equipped sets
	for _, set := range itemOptions.UniqueEquippedSets() {
		row := new(util_highs.ConstraintRow)
		row.Debug = "uniqueEquipped" + set[0].String()
		setup.uniqueEquipRowsAll = append(setup.uniqueEquipRowsAll, row)

		for _, itemId := range set {
			if seen[itemId] {
				panic("unique equipped data has duplicate")
			}
			setup.uniqueEquipRowsById[itemId] = row
			seen[itemId] = true
		}
	}
}

func (setup *singleGearSetShared) checkActivePermutation(solution *highs.Solution, solvableItemSet *items.SolvableItemSet) {
	if len(setup.bonusCombos) > 0 {
		var activePermutation *bonusCombo

		for _, permutation := range setup.bonusCombos {
			variableResult := solution.ColValues[permutation.activatingVar.columnIndex]
			if util.FloatEqualsOne(variableResult) {
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

func (setup *singleGearSetShared) buildResultSet(solution *highs.Solution) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for columnEntry := range setup.itemColumns.SeqValues() {
		variableResult := solution.ColValues[columnEntry.columnIndex]
		if columnEntry.entryType == entry_item && util.FloatEqualsOne(variableResult) {
			itemSet.AddItem_DeferCalc_ExpectEmpty(columnEntry.itemSlot, columnEntry.item)
		}
	}
	items.SolvableItemSet_RecalculateTotal(&itemSet)

	setup.checkActivePermutation(solution, &itemSet)
	return itemSet
}
