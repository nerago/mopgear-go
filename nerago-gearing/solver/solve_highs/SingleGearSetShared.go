package solve_highs

import (
	"fmt"
	"iter"
	"paladin_gearing_go/gear_model/bonus_set"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

type singleGearSetShared struct {
	build *util_highs.LinearBuilder

	slotsOneEachRow [items.ITEM_SLOT_COUNT]util_highs.ConstraintRow // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot

	ratingPreScale float64

	uniqueEquipRowsById map[items.ItemId]*util_highs.ConstraintRow // lookup by id, may have multiple mappings for an item so need pointers
	uniqueEquipRowsAll  []*util_highs.ConstraintRow                // definitive copy of each unique equip row constraint

	bonusData   []bonusInfo
	bonusCombos []bonusCombo

	itemColumns util_collection.MapSlice[items.ItemId, *columnInfo]
	allColumns  []*columnInfo
}

func (sc *singleGearSetShared) runForFutureResult(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder) *util_async.FutureCancellable[items.SolvableItemSet] {
	solutionFuture := sc.build.RunHighsFuture(nil)

	return util_async.FutureCancellable_MapValue(solutionFuture, func(result util_highs.LinearResult) (items.SolvableItemSet, bool) {
		solution := result.GetSolution2AndSaveLog(printer)

		printer.Printf("SOLUTION STATUS = %s\n", solution.Status().String())
		debugPrint(solution, sc.build, sc.allColumns, printer)

		if solution.HasSolution() {
			itemSet := sc.buildResultSet(solution)
			validateNewSet(itemSet, itemOptions, model.CheckSet)
			sc.checkSetRatingIsObjective(solution, &itemSet, model.CalcRatingSet)
			return itemSet, true
		} else {
			return items.SolvableItemSet{}, false
		}
	})
}

func (sc *singleGearSetShared) checkSetRatingIsObjective(solution *util_highs.Solution2, itemSet *items.SolvableItemSet, calcRating func(item *items.SolvableItemSet) float64) {
	checkRating := calcRating(itemSet)
	if !util.FloatsApproxEquals(solution.Objective()/sc.ratingPreScale, checkRating) {
		panic(fmt.Sprintf("rating inconsistent %e %e ", solution.Objective(), checkRating))
	}
}

func (sc *singleGearSetShared) ColumnsForItemId(itemId items.ItemId) iter.Seq[*columnInfo] {
	return sc.itemColumns.ValuesForKeyAsSeq(itemId)
}

func (sc *singleGearSetShared) AllColumns() []*columnInfo {
	return sc.allColumns
}

func (sc *singleGearSetShared) RatingPreScale() float64 {
	return sc.ratingPreScale
}

func (sc *singleGearSetShared) prepareCommon(model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap) {
	sc.prepareEnabledSets(model)
	sc.prepareSetCombos()
	sc.prepareUniqueEquipped(itemOptions)
	sc.addSetNeededCounts(model.SetBonusRequiredCounts, model.SetBonusCountMode)
}

func (sc *singleGearSetShared) addItemCommon(itemSlot items.SlotEquip, item *items.SolvableItem, activeSet func(id items.ItemId) (int, bool)) util_highs.ColumnIndex {
	entry := columnInfo{entryType: entry_item, itemSlot: itemSlot, item: item}

	// boolean value to flag use of specific item, in exact reforge/gem state
	columnIndex := sc.build.CreateColumnBool(&entry)
	entry.columnIndex = columnIndex
	sc.allColumns = append(sc.allColumns, &entry)
	sc.itemColumns.Add(item.ItemId(), &entry)

	// 1 for that slot that matches the item, so we can tell solver only one item per slot
	sc.slotsOneEachRow[itemSlot].Add(columnIndex, 1.0)

	// if this item belongs to any item set then flag with a 1
	activeSetIndex, hasSet := activeSet(item.ItemId())
	if hasSet {
		sc.bonusData[activeSetIndex].countSetItemsRow.Add(columnIndex, 1)
	}

	// if this item is unique equipped (mostly checked for ring/trinket)
	uniqueRow := sc.uniqueEquipRowsById[item.ItemId()]
	if uniqueRow != nil {
		uniqueRow.Add(columnIndex, 1)
	}

	return columnIndex
}

func (sc *singleGearSetShared) finishItemsCommon(itemOptions *items.SolvableOptionsMap) {
	// constrain: exactly one item for each slot
	for slot, row := range sc.slotsOneEachRow {
		slotEquip := items.SlotEquip(slot)
		row.Debug = "slotsOneEachRow_" + slotEquip.Name()
		if itemOptions.Has(slotEquip) {
			row.Build(sc.build, 1, 1)
		} else {
			row.Build(sc.build, 0, 0)
		}
	}

	// constrain: matching number of items from each given set
	for _, setInfo := range sc.bonusData {
		setInfo.countSetItemsRow.Build(sc.build, 0, 0)
	}

	// constrain: unique item by itemid/unique set
	for _, row := range sc.uniqueEquipRowsAll {
		if !row.IsEmpty() {
			row.Build(sc.build, 0, 1)
		}
	}
}

func (sc *singleGearSetShared) prepareEnabledSets(model *solve_highs_types.SolverModel) {
	// constrain: exact item count in each active set
	if model.SetBonusTotalCount > 0 {
		sc.bonusData = make([]bonusInfo, model.SetBonusTotalCount)
		for setIndex := range model.SetBonusTotalCount {
			info := bonusInfo{
				setIndex:            solve_highs_types.SetBonusIndex(setIndex),
				setCountItems:       model.SetBonusCountItems[setIndex],
				setMultipliers:      model.SetBonusMultipliersFlat[setIndex],
				setMultipliersBySim: model.SetBonusMultipliersBySim[setIndex],
			}
			info.countSetItemsRow.Debug = "countSetItemsRow" + strconv.Itoa(setIndex)
			sc.addSetItemCountVariable(&info)
			sc.addSetItemsCountExactVariables(&info)
			sc.bonusData[setIndex] = info
		}
	}
}

func (sc *singleGearSetShared) multiplyByActiveCombo(combinedRatingVar *columnInfo, outputVar *columnInfo, rangeHigh float64, getMultiplier func(*bonusCombo) float64) {
	if len(sc.bonusData) > 0 {
		for combo := range util_collection.ForPointer(sc.bonusCombos) {
			sc.buildSetMultipliedOutput(combo, combinedRatingVar, outputVar, rangeHigh, getMultiplier)
		}
	} else {
		sc.buildSimpleNoSetsOutput(combinedRatingVar, outputVar)
	}
}

func (sc *singleGearSetShared) prepareSetCombos() {
	if len(sc.bonusData) == 0 {
		return
	}

	sc.makeSetCombosRecur(sc.bonusData, nil, 0)

	for combo := range util_collection.ForPointer(sc.bonusCombos) {
		sc.buildCombinationActivatingVar(combo)
	}

	checkSingleCombo := util_highs.ConstraintRow{}
	for combo := range util_collection.ForPointer(sc.bonusCombos) {
		checkSingleCombo.Add(combo.activatingVar.columnIndex, 1)
	}
	checkSingleCombo.Build(sc.build, 1, 1)
}

func (sc *singleGearSetShared) makeSetCombosRecur(setData []bonusInfo, built []bonusWithCount, builtComboItemCount int) {
	if len(setData) == 0 || builtComboItemCount == c_maxSetItems {
		sc.bonusCombos = append(sc.bonusCombos, bonusCombo{condition: built})
	} else {
		addSet := &setData[0]
		for itemCount := 0; itemCount <= c_maxSetItems-builtComboItemCount; itemCount++ {
			next := bonusWithCount{addSet, itemCount}
			progress := util_collection.CopyAndAppend(built, next)
			sc.makeSetCombosRecur(setData[1:], progress, builtComboItemCount+itemCount)
		}
	}
}

func (sc *singleGearSetShared) buildSimpleNoSetsOutput(inputVar *columnInfo, outputVar *columnInfo) {
	// just copy initial rating sum into final if no sets
	sc.build.ConstraintCopy(
		inputVar.columnIndex, 1,
		outputVar.columnIndex,
		"nullComboCopy",
	)
}

func (sc *singleGearSetShared) buildSetMultipliedOutput(combo *bonusCombo, inputVar *columnInfo, outputVar *columnInfo, rangeHigh float64, getMultiplier func(*bonusCombo) float64) {
	activatingVar := combo.activatingVar
	bonusMultiplier := getMultiplier(combo)

	sc.build.ConstraintCopyIfBool(
		activatingVar.columnIndex,
		inputVar.columnIndex, bonusMultiplier,
		outputVar.columnIndex,
		rangeHigh,
	)
}

func (sc *singleGearSetShared) createOutputVariableForSeparateRun() *columnInfo {
	entry := &columnInfo{entryType: entry_main_output}

	// goes directly into overall rating
	entry.columnIndex = sc.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), 1, entry)

	// save reference
	sc.allColumns = append(sc.allColumns, entry)

	return entry
}

func (sc *singleGearSetShared) addSetNeededCounts(setBonusRequired []solve_highs_types.SetBonusRequiredCounts, countMode bonus_set.ItemCountsRequiredMode) {
	if len(setBonusRequired) > 0 {
		if len(sc.bonusData) == 0 {
			panic("no bonusData to use for addSetNeededCounts")
		} else if len(sc.bonusData) == 1 && len(setBonusRequired) == 1 && len(setBonusRequired[0]) == 1 {
			setCountCol := sc.bonusData[0].setTotalCountVar
			needCount := setBonusRequired[0][0]

			rowSetCountRequired := util_highs.ConstraintRow{Debug: "rowSetCountRequired"}
			rowSetCountRequired.Add(setCountCol.columnIndex, 1)
			switch countMode {
			case bonus_set.CountMode_Exact:
				rowSetCountRequired.Build(sc.build, float64(needCount), float64(needCount))
			case bonus_set.CountMode_Minimum:
				rowSetCountRequired.Build(sc.build, float64(needCount), util_highs.InfPos())
			case bonus_set.CountMode_AllowPlusOne:
				rowSetCountRequired.Build(sc.build, float64(needCount), float64(needCount+1))
			default:
				panic("unknown type")
			}
		} else {
			oneOfTheseOptions := util_highs.ConstraintRow{}

			for _, option := range setBonusRequired {
				optionParts := util_highs.ConstraintAndBuilder{}

				for setIndex, needCount := range option {
					setInfo := sc.bonusData[setIndex]
					setCountCol := setInfo.setTotalCountVar

					var inRange util_highs.ColumnIndex
					switch countMode {
					case bonus_set.CountMode_Exact:
						inRange = sc.build.CreateColumnBool(nil)
						sc.build.ColumnIsEqualConstant(setCountCol.columnIndex, inRange, float64(needCount), 10, 1.0)
					case bonus_set.CountMode_Minimum:
						inRange = sc.build.ColumnIsGreaterOrEqualThanConstant(setCountCol.columnIndex, float64(needCount), 10, 1.0)
					case bonus_set.CountMode_AllowPlusOne:
						inRange = sc.build.ColumnIsBetweenConstants(setCountCol.columnIndex, float64(needCount), float64(needCount+1), 10, 1.0)
					default:
						panic("unknown type")
					}
					optionParts.AddInput(inRange)
				}

				optionActive := sc.build.CreateColumnBool(util_highs.DebugText("SetBonusRequired option"))
				optionParts.SetOutput(optionActive)
				optionParts.Build(sc.build)

				oneOfTheseOptions.Add(optionActive, 1)
			}

			oneOfTheseOptions.Build(sc.build, 1, util_highs.InfPos())
		}
	}
}

// logical AND between exact count vars
func (sc *singleGearSetShared) buildCombinationActivatingVar(combo *bonusCombo) {
	comboActiveBool := &columnInfo{entryType: entry_combo_active, combo: combo}
	comboActiveBool.columnIndex = sc.build.CreateColumnBool(comboActiveBool)
	combo.activatingVar = comboActiveBool
	sc.allColumns = append(sc.allColumns, comboActiveBool)

	buildAnd := util_highs.ConstraintAndBuilder{}
	buildAnd.SetOutput(comboActiveBool.columnIndex)
	for _, setAndCount := range combo.condition {
		setInfo := setAndCount.setInfo
		count := setAndCount.count
		specificExactBool := setInfo.setExactCountVars[count]
		buildAnd.AddInput(specificExactBool.columnIndex)
	}
	buildAnd.Build(sc.build)
}

func (sc *singleGearSetShared) addSetItemCountVariable(info *bonusInfo) {
	entry := columnInfo{entryType: entry_set_total_count, setIndex: info.setIndex}

	// set item actual count
	entry.columnIndex = sc.build.CreateColumnGeneral(highs.Integer, 0, c_maxSetItems, &entry)

	// add corresponding -1 to the set item count array, so we can compare value to the sum of items, relevant items will flag a 1
	info.countSetItemsRow.Add(entry.columnIndex, -1)

	// save reference
	info.setTotalCountVar = &entry
	sc.allColumns = append(sc.allColumns, &entry)
}

func (sc *singleGearSetShared) addSetItemsCountExactVariables(info *bonusInfo) {
	// compare total number of items previous computed into this constraint
	compareRow := util_highs.ConstraintRow{Debug: "setItemsCompareRow"}
	compareRow.Add(info.setTotalCountVar.columnIndex, -1)

	// constraint so only one of these flags gets set
	singleFlagOnly := util_highs.ConstraintRow{Debug: "setItemsSingleFlagOnly"}

	// make a bool for each possible count in range 0..5
	for itemCount := 0; itemCount <= c_maxSetItems; itemCount++ {
		boolColumn := columnInfo{entryType: entry_set_exact_count, setIndex: info.setIndex, itemCount: itemCount}
		boolColumn.columnIndex = sc.build.CreateColumnBool(&boolColumn)

		// should activate this flag which will match the total count
		compareRow.Add(boolColumn.columnIndex, float64(itemCount))

		// but only one flag at a time
		singleFlagOnly.Add(boolColumn.columnIndex, 1)

		info.setExactCountVars[itemCount] = &boolColumn
		sc.allColumns = append(sc.allColumns, &boolColumn)
	}

	compareRow.Build(sc.build, 0, 0)     // equal
	singleFlagOnly.Build(sc.build, 1, 1) // sum of flags should be just one, should pull the zero flag up if no other set
}

func (sc *singleGearSetShared) prepareUniqueEquipped(itemOptions *items.SolvableOptionsMap) {
	sc.uniqueEquipRowsById = make(map[items.ItemId]*util_highs.ConstraintRow)
	sc.uniqueEquipRowsAll = make([]*util_highs.ConstraintRow, 0)
	seen := make(map[items.ItemId]bool)

	// add items from predefined unique equipped sets
	for _, set := range itemOptions.UniqueEquippedSets() {
		row := new(util_highs.ConstraintRow)
		row.Debug = "uniqueEquipped" + set[0].String()
		sc.uniqueEquipRowsAll = append(sc.uniqueEquipRowsAll, row)

		for _, itemId := range set {
			if seen[itemId] {
				panic("unique equipped data has duplicate")
			}
			sc.uniqueEquipRowsById[itemId] = row
			seen[itemId] = true
		}
	}
}

func (sc *singleGearSetShared) checkActiveCombo(solution util_highs.ISolution, solvableItemSet *items.SolvableItemSet) {
	if len(sc.bonusCombos) > 0 {
		var activeCombo *bonusCombo

		for _, combo := range sc.bonusCombos {
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

func (sc *singleGearSetShared) buildResultSet(solution util_highs.ISolution) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for columnEntry := range sc.itemColumns.SeqValues() {
		isTrue := solution.ValueIsOne(columnEntry.columnIndex)
		if columnEntry.entryType == entry_item && isTrue {
			itemSet.AddItem_DeferCalc_ExpectEmpty(columnEntry.itemSlot, columnEntry.item)
		}
	}
	items.SolvableItemSet_RecalculateTotal(&itemSet)

	sc.checkActiveCombo(solution, &itemSet)
	return itemSet
}
