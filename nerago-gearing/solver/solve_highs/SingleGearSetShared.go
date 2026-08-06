package solve_highs

import (
	"iter"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

type singleGearSetShared struct {
	build *util_highs.LinearBuilder

	slotsOneEachRow [items.ITEM_SLOT_COUNT]util_highs.ConstraintRow // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot

	//mainOutputRow util_highs.ConstraintRow // compute final output from set based alternatives
	mainOutputVar *columnInfo // output variable, to be used directly or scaled against other models

	uniqueEquipRowsById map[items.ItemId]*util_highs.ConstraintRow // lookup by id, may have multiple mappings for an item so need pointers
	uniqueEquipRowsAll  []*util_highs.ConstraintRow                // definitive copy of each unique equip row constraint

	bonusData   []bonusInfo
	bonusCombos []bonusCombo

	itemColumns util_collection.MapSlice[items.ItemId, *columnInfo]
	allColumns  []*columnInfo
}

func (setup *singleGearSetShared) ColumnsForItemId(itemId items.ItemId) iter.Seq[*columnInfo] {
	return setup.itemColumns.ValuesForKeyAsSeq(itemId)
}

func (setup *singleGearSetShared) AllColumns() []*columnInfo {
	return setup.allColumns
}

func (setup *singleGearSetShared) MainOutputVar() *columnInfo {
	return setup.mainOutputVar
}

func (setup *singleGearSetShared) addItemCommon(itemSlot items.SlotEquip, item *items.SolvableItem, activeSet func(id items.ItemId) (int, bool)) util_highs.ColumnIndex {
	entry := columnInfo{entryType: entry_item, itemSlot: itemSlot, item: item}

	// boolean value to flag use of specific item, in exact reforge/gem state
	columnIndex := setup.build.CreateColumnBool(&entry)
	entry.columnIndex = columnIndex
	setup.allColumns = append(setup.allColumns, &entry)
	setup.itemColumns.Add(item.ItemId(), &entry)

	// 1 for that slot that matches the item, so we can tell solver only one item per slot
	setup.slotsOneEachRow[itemSlot].Add(columnIndex, 1.0)

	// if this item belongs to any item set then flag with a 1
	activeSetIndex, hasSet := activeSet(item.ItemId())
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

func (setup *singleGearSetShared) prepareActiveSetCombos(activeSetCount int) {
	// constrain: exact item count in each active set
	if activeSetCount > 0 {
		setup.bonusData = make([]bonusInfo, activeSetCount)
		for setIndex := range activeSetCount {
			info := bonusInfo{setIndex: setBonusIndex(setIndex)}
			info.countSetItemsRow.Debug = "countSetItemsRow" + strconv.Itoa(setIndex)
			setup.addSetItemCountVariable(&info)
			setup.addSetItemsCountExactVariables(&info)
			setup.bonusData[setIndex] = info
		}
	}
}

func (setup *singleGearSetShared) multiplyRatingsByActiveSetCombo(combinedRatingVar *columnInfo, rangeHigh float64) {
	if len(setup.bonusData) > 0 {
		setup.makeSetCombos()

		checkSingleCombo := util_highs.ConstraintRow{}
		for i := range setup.bonusCombos {
			comboActiveVar := setup.buildSetMultipliedOutput(&setup.bonusCombos[i], combinedRatingVar, rangeHigh)
			checkSingleCombo.Add(comboActiveVar.columnIndex, 1)
		}
		checkSingleCombo.Build(setup.build, 1, 1)
	} else {
		setup.buildSimpleNoSetsOutput(combinedRatingVar)
	}
}

func (setup *singleGearSetShared) makeSetCombos() {
	setup.makeSetCombosRecur(setup.bonusData, nil, 0)
}

func (setup *singleGearSetShared) makeSetCombosRecur(setData []bonusInfo, built []bonusWithCount, builtComboItemCount int) {
	if len(setData) == 0 || builtComboItemCount == c_maxSetItems {
		setup.bonusCombos = append(setup.bonusCombos, bonusCombo{condition: built})
	} else {
		addSet := &setData[0]
		for itemCount := 0; itemCount <= c_maxSetItems-builtComboItemCount; itemCount++ {
			next := bonusWithCount{addSet, itemCount}
			progress := util_collection.CopyAndAppend(built, next)
			setup.makeSetCombosRecur(setData[1:], progress, builtComboItemCount+itemCount)
		}
	}
}

func (setup *singleGearSetShared) buildSimpleNoSetsOutput(combinedRatingVar *columnInfo) {
	// just copy initial rating sum into final if no sets
	setup.build.ConstraintCopy(
		combinedRatingVar.columnIndex, 1,
		setup.mainOutputVar.columnIndex,
		"nullComboCopy",
	)
}

func (setup *singleGearSetShared) buildSetMultipliedOutput(combo *bonusCombo, combinedRatingVar *columnInfo, rangeHigh float64) *columnInfo {
	activatingVar := setup.buildCombinationActivatingVar(combo)

	bonusMultiplier := 1.0
	for _, setAndCount := range combo.condition {
		bonusForCount := setAndCount.setInfo.setMultipliers[setAndCount.count]
		bonusMultiplier *= bonusForCount
	}

	setup.build.ConstraintCopyIfBool(
		activatingVar.columnIndex,
		combinedRatingVar.columnIndex, bonusMultiplier,
		setup.mainOutputVar.columnIndex,
		rangeHigh,
	)

	return activatingVar
}

func (setup *singleGearSetShared) addMainOutputVariable(scaleOutputRating float64) {
	entry := columnInfo{entryType: entry_main_output}

	// goes directly into overall rating, but could have an external scale applied
	entry.columnIndex = setup.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), scaleOutputRating, &entry)

	// save reference
	setup.mainOutputVar = &entry
	setup.allColumns = append(setup.allColumns, &entry)
}

func (setup *singleGearSetShared) addSetNeededCounts(setBonusRequired []setBonusRequiredCounts) {
	if len(setBonusRequired) > 0 {
		if len(setup.bonusData) == 0 {
			panic("no bonusData to use for addSetNeededCounts")
		} else if len(setup.bonusData) == 1 && len(setBonusRequired) == 1 && len(setBonusRequired[0]) == 1 {
			setCountCol := setup.bonusData[0].setTotalCountVar
			needCount := setBonusRequired[0][0]

			rowSetCountRequired := util_highs.ConstraintRow{Debug: "rowSetCountRequired"}
			rowSetCountRequired.Add(setCountCol.columnIndex, 1)
			rowSetCountRequired.Build(setup.build, float64(needCount), util_highs.InfPos())
		} else {
			oneOfTheseOptions := util_highs.ConstraintRow{}

			for _, option := range setBonusRequired {
				optionParts := util_highs.ConstraintAndBuilder{}

				for setIndex, needCount := range option {
					setInfo := setup.bonusData[setIndex]
					setCountCol := setInfo.setTotalCountVar

					inRange := setup.build.ColumnIsGreaterOrEqualThanConstant(setCountCol.columnIndex, float64(needCount), 10, 1.0)
					optionParts.AddInput(inRange)
				}

				optionActive := setup.build.CreateColumnBool(util_highs.DebugText("SetBonusRequired option"))
				optionParts.SetOutput(optionActive)
				optionParts.Build(setup.build)

				oneOfTheseOptions.Add(optionActive, 1)
			}

			oneOfTheseOptions.Build(setup.build, 1, util_highs.InfPos())
		}
	}
}

// logical AND between exact count vars
func (setup *singleGearSetShared) buildCombinationActivatingVar(combo *bonusCombo) *columnInfo {
	comboActiveBool := &columnInfo{entryType: entry_combo_active, combo: combo}
	comboActiveBool.columnIndex = setup.build.CreateColumnBool(comboActiveBool)
	combo.activatingVar = comboActiveBool
	setup.allColumns = append(setup.allColumns, comboActiveBool)

	buildAnd := util_highs.ConstraintAndBuilder{}
	buildAnd.SetOutput(comboActiveBool.columnIndex)
	for _, setAndCount := range combo.condition {
		setInfo := setAndCount.setInfo
		count := setAndCount.count
		specificExactBool := setInfo.setExactCountVars[count]
		buildAnd.AddInput(specificExactBool.columnIndex)
	}
	buildAnd.Build(setup.build)

	return comboActiveBool
}

func (setup *singleGearSetShared) addSetItemCountVariable(info *bonusInfo) {
	entry := columnInfo{entryType: entry_set_total_count, setIndex: info.setIndex}

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
		boolColumn := columnInfo{entryType: entry_set_exact_count, setIndex: info.setIndex, itemCount: itemCount}
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

func (setup *singleGearSetShared) checkActiveCombo(solution util_highs.ISolution, solvableItemSet *items.SolvableItemSet) {
	if len(setup.bonusCombos) > 0 {
		var activeCombo *bonusCombo

		for _, combo := range setup.bonusCombos {
			if solution.ValueIsOne(combo.activatingVar.columnIndex) {
				if activeCombo == nil {
					activeCombo = &combo
				} else {
					panic("multiple combos active")
				}
			}
		}

		if activeCombo != nil {
			for _, entry := range activeCombo.condition {
				actualCount := entry.setInfo.setCountItems(solvableItemSet.Items())
				if actualCount != uint8(entry.count) {
					panic("number of items not what solver returned")
				}
			}
		} else {
			panic("no combos active")
		}
	}
}

func (setup *singleGearSetShared) buildResultSet(solution util_highs.ISolution) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for columnEntry := range setup.itemColumns.SeqValues() {
		isTrue := solution.ValueIsOne(columnEntry.columnIndex)
		if columnEntry.entryType == entry_item && isTrue {
			itemSet.AddItem_DeferCalc_ExpectEmpty(columnEntry.itemSlot, columnEntry.item)
		}
	}
	items.SolvableItemSet_RecalculateTotal(&itemSet)

	setup.checkActiveCombo(solution, &itemSet)
	return itemSet
}
