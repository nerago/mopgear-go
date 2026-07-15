package solve_highs

import (
	gear_model "paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

type WeightExtended util.MapMap[stats.StatType, stats.SimType, float64]
type StatRequiredExtended map[stats.StatType]util.HiLoInt

type ExtendedModel struct {
	weight    WeightExtended
	require   StatRequiredExtended
	gearModel *gear_model.SpecModel
}

func SingleGearSetExtendedMain(itemOptions *items.SolvableOptionsMap, model *ExtendedModel, printer *util.PrintRecorder) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior

	setup := setupGearSetExtended(&build, model, itemOptions, 1)

	solutionFuture := build.RunHighsFuture(nil)

	return util_async.FutureCancellable_MapValue(solutionFuture, func(result util_highs.LinearResult) (items.SolvableItemSet, bool) {
		solution := result.GetSolutionAndSaveLog(printer)

		printer.Printf("SOLUTION STATUS = %s\n", solution.Status.String())
		debugPrint(solution, setup.build, setup.allColumns, printer)

		if solution.HasSolution() {
			itemSet := setup.buildResultSet(solution, itemOptions, model)
			//checkSetRatingIsObjective(solution, &itemSet, model) // TODO extended version
			return itemSet, true
		} else {
			return items.SolvableItemSet{}, false
		}
	})
}

func setupGearSetExtended(build *util_highs.LinearBuilder, model *ExtendedModel, itemOptions *items.SolvableOptionsMap, scaleOutputRating float64) *singleGearSetExtendedInputs {
	setup := singleGearSetExtendedInputs{build: build}

	setup.addMainOutputVariable(scaleOutputRating)
	setup.addSumRatingVariable()
	setup.prepareRequire(&model.require)
	setup.prepareActiveSets(&model.gearModel.SetBonus)
	setup.prepareUniqueEquipped(itemOptions)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		setup.addItem(slot, item, &model.weight, &model.require, &model.gearModel.SetBonus)
	}

	setup.finishItems(itemOptions, &model.require)

	setup.addSetNeededCounts(model.gearModel.SetBonusRequired)

	return &setup
}

func (setup *singleGearSetExtendedInputs) addSetNeededCounts(setBonusRequired []gear_model.ActiveSetCountsRequired) {
	if len(setBonusRequired) > 0 {
		if len(setup.setData) == 0 {
			panic("no setdata to use from SetBonusRequired")
		} else if len(setup.setData) == 1 && len(setBonusRequired) == 1 && setBonusRequired[0].Count() == 1 {
			setCountCol := setup.setData[0].setTotalCountVar
			_, needCount := setBonusRequired[0].PairsByIndex(0)

			rowSetCountRequired := util_highs.ConstraintRow{Debug: "rowSetCountRequired"}
			rowSetCountRequired.Add(setCountCol.columnIndex, 1)
			rowSetCountRequired.Build(setup.build, float64(needCount), util_highs.C_PlusInf)
		} else {
			oneOfTheseOptions := util_highs.ConstraintRow{}

			for _, option := range setBonusRequired {
				optionParts := util_highs.ConstraintAndBuilder{}

				for activeSet, needCount := range option.Pairs() {
					setInfo := util.FindWith(setup.setData, func(x setInfo) bool { return x.activeSet.Equals(activeSet) })
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

type singleGearSetExtendedInputs struct {
	build *util_highs.LinearBuilder

	slotsOneEachRow [items.ITEM_SLOT_COUNT]util_highs.ConstraintRow // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot

	requireRows map[stats.StatType]*util_highs.ConstraintRow // constrains values for the hit/expertise/etc of each item
	statTotals  map[stats.StatType]*util_highs.ConstraintRow

	//baseRatingSumRow util_highs.ConstraintRow // values for the ratings of each item
	//baseRatingSumVar *columnInfo              // sum of values for the ratings of selected items

	mainOutputRow util_highs.ConstraintRow // compute final output from set based alternatives
	mainOutputVar *columnInfo              // output variable, to be used directly or scaled against other models

	uniqueEquipRowsById map[items.ItemId]*util_highs.ConstraintRow // lookup by id, may have multiple mappings for an item so need pointers
	uniqueEquipRowsAll  []*util_highs.ConstraintRow                // definitive copy of each unique equip row constraint

	setData           []setInfo
	allSetPermutation []setPermutation

	itemColumns util.MapSlice[items.ItemId, *columnInfo]
	allColumns  []*columnInfo
}

func (setup *singleGearSetExtendedInputs) prepareActiveSets(setBonus *gear_model.SetBonus) {
	// constrain: exact item count in each active set
	activeSets := setBonus.ActiveSets()
	if len(activeSets) > 0 {
		setup.setData = make([]setInfo, len(activeSets))
		for setIndex, set := range activeSets {
			info := setInfo{activeSet: set, setIndex: setIndex}
			info.countSetItemsRow.Debug = "countSetItemsRow" + strconv.Itoa(setIndex)
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

func (setup *singleGearSetExtendedInputs) buildSimpleNoSetsOutput() {
	// just copy initial rating sum into final if no sets
	setup.mainOutputRow.Add(setup.baseRatingSumVar.columnIndex, 1)
}

func (setup *singleGearSetExtendedInputs) buildSetMultipliedOutput(permutation *setPermutation) {
	outputVar, weight := setup.buildSetWeightedOutputVar(permutation)
	activatingVar := setup.buildPermutationActivatingVar(permutation)

	// copy regular rating sum to column if flag is set
	setup.build.ContraintIfBoolCopyValueElseZero(activatingVar.columnIndex, setup.baseRatingSumVar.columnIndex, outputVar.columnIndex, c_ratings_low_range, c_ratings_high_range)

	// add scaled rating to final computation
	setup.mainOutputRow.Add(outputVar.columnIndex, weight)

	permutation.outputVar = outputVar
	permutation.activatingVar = activatingVar
	permutation.weight = weight
}

func (setup *singleGearSetExtendedInputs) buildSetWeightedOutputVar(permutation *setPermutation) (*columnInfo, float64) {
	totalWeight := 1.0
	for _, setAndCount := range permutation.content {
		bonusForCount := setAndCount.setInfo.activeSet.BonusForCount(uint8(setAndCount.count))
		totalWeight *= float64(bonusForCount)
	}

	// the actual output variable from this permutation, applies relevant set related multipliers
	entry := columnInfo{entryType: entry_permutation_output_weighted, permutation: permutation, weight: totalWeight}
	entry.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, &entry)

	setup.allColumns = append(setup.allColumns, &entry)
	return &entry, totalWeight
}

func (setup *singleGearSetExtendedInputs) buildPermutationActivatingVar(permutation *setPermutation) *columnInfo {
	// we are effecively building a logical AND between these vars

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

func (setup *singleGearSetExtendedInputs) addSetItemCountVariable(info *setInfo) {
	entry := columnInfo{entryType: entry_set_total_count, set: info.activeSet}

	// set item actual count
	entry.columnIndex = setup.build.CreateColumnGeneral(highs.Integer, 0, c_maxSetItems, &entry)

	// add corresponding -1 to the set item count array, so we can compare value to the sum of items, relevant items will flag a 1
	info.countSetItemsRow.Add(entry.columnIndex, -1)

	// save reference

	info.setTotalCountVar = &entry
	setup.allColumns = append(setup.allColumns, &entry)
}

func (setup *singleGearSetExtendedInputs) addSetItemsCountExactVariables(info *setInfo) {
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

func (setup *singleGearSetExtendedInputs) addMainOutputVariable(scaleOutputRating float64) {
	entry := columnInfo{entryType: entry_main_output}

	// goes directly into overall rating, but could have an external scale applied
	entry.columnIndex = setup.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, scaleOutputRating, &entry)

	// derive value based on whichever setup bonus permutation is active
	setup.mainOutputRow.Add(entry.columnIndex, -1)

	// save reference
	setup.mainOutputVar = &entry
	setup.allColumns = append(setup.allColumns, &entry)
}

func (setup *singleGearSetExtendedInputs) addSumRatingVariable() {
	entry := columnInfo{entryType: entry_sum_rating}

	// sum of individual selected item ratings
	// doesen't go directly into output rating
	entry.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.C_PlusInf, &entry)

	// main action of this variable: derive value to match rest of rest of row sum
	setup.baseRatingSumRow.Add(entry.columnIndex, -1)

	// save reference
	setup.baseRatingSumVar = &entry
	setup.allColumns = append(setup.allColumns, &entry)
}

func (setup *singleGearSetExtendedInputs) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, weight *WeightExtended, require *StatRequiredExtended, setBonus *gear_model.SetBonus) util_highs.ColumnIndex {
	entry := columnInfo{entryType: entry_item, itemSlot: itemSlot, item: item}

	// boolean value to flag use of specific item
	// contributes 0 to final rating itself, but via additional summation and calcs
	columnIndex := setup.build.CreateColumnBool(&entry)
	entry.columnIndex = columnIndex
	setup.allColumns = append(setup.allColumns, &entry)
	setup.itemColumns.Add(item.ItemId(), &entry)

	// add rating via a summation condition
	// scale down ratings to keep numbers small for solver stability
	rating := float64(gear_model.CalcRatingSolveItem(item)) / c_scaled_ratings
	setup.baseRatingSumRow.Add(columnIndex, rating)

	// specific hit/expertise values for hi/lo limits
	for statType := range *require {
		setup.requireRows[statType].Add(columnIndex, item.Total().GetFloat(statType))
	}

	// 1 for that slot that matches the item, so we can tell solver only one item per slot
	setup.slotsOneEachRow[itemSlot].Add(columnIndex, 1.0)

	// if this item belongs to any item set then flag with a 1
	activeSetIndex, hasSet := setBonus.ActiveSetIndexForItem(item.ItemId())
	if hasSet {
		setup.setData[activeSetIndex].countSetItemsRow.Add(columnIndex, 1)
	}

	// if this item is unique equipped (mostly checked for ring/trinket)
	uniqueRow := setup.uniqueEquipRowsById[item.ItemId()]
	if uniqueRow != nil {
		uniqueRow.Add(columnIndex, 1)
	}

	return columnIndex
}

func (setup *singleGearSetExtendedInputs) prepareRequire(require *StatRequiredExtended) {
	setup.requireRows = make(map[stats.StatType]*util_highs.ConstraintRow, len(*require))
	for statType := range *require {
		setup.requireRows[statType] = &util_highs.ConstraintRow{Debug: "require " + statType.Name()}
	}
}

func (setup *singleGearSetExtendedInputs) finishItems(itemOptions *items.SolvableOptionsMap, require *StatRequiredExtended) {
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

	// constrain: total sum of hit/exp are within requested limits
	for statType, hilo := range *require {
		row := setup.requireRows[statType]
		row.Build(setup.build, float64(hilo.Lo), float64(hilo.Hi))
	}

	// constrain: matching sum to individual ratings
	setup.baseRatingSumRow.Debug = "baseRatingSumRow"
	setup.baseRatingSumRow.Build(setup.build, 0, 0)

	// constrain: matching number of items from each given set
	for _, setInfo := range setup.setData {
		setInfo.countSetItemsRow.Build(setup.build, 0, 0)
	}

	// constrain: unique item by itemid/unique set
	for _, row := range setup.uniqueEquipRowsAll {
		if !row.IsEmpty() {
			row.Build(setup.build, 0, 1)
		}
	}

	// constrain: whichever alternate set output into final
	setup.mainOutputRow.Debug = "mainOutputRow"
	setup.mainOutputRow.Build(setup.build, 0, 0)
}

func (setup *singleGearSetExtendedInputs) prepareUniqueEquipped(itemOptions *items.SolvableOptionsMap) {
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

func (setup *singleGearSetExtendedInputs) buildResultSet(solution *highs.Solution, itemOptions *items.SolvableOptionsMap, model *ExtendedModel) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for columnEntry := range setup.itemColumns.SeqValues() {
		variableResult := solution.ColValues[columnEntry.columnIndex]
		if columnEntry.entryType == entry_item && util.FloatEqualsOne(variableResult) {
			itemSet.AddItem_DeferCalc_ExpectEmpty(columnEntry.itemSlot, columnEntry.item)
		}
	}
	items.SolvableItemSet_RecalculateTotal(&itemSet)

	//validateNewSet(itemSet, itemOptions, model) TODO extended version
	setup.checkActivePermutation(solution, &itemSet)

	return itemSet
}

func (setup *singleGearSetExtendedInputs) checkActivePermutation(solution *highs.Solution, solvableItemSet *items.SolvableItemSet) {
	if len(setup.allSetPermutation) > 0 {
		var activePermutation *setPermutation

		for _, permutation := range setup.allSetPermutation {
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
