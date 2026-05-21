package withhighs

import (
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/util"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_maxSetItems    = 5 // fundamental in MoP gear sets
	c_setItemsCounts = c_maxSetItems + 1
)

func RunAllActiveSets(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, printer *util.PrintRecorder) util.Optional[items.SolvableItemSet] {
	inputBuilder := utilhighs.InputBuilder{}
	setup := setupBonusedInputs(&inputBuilder, gear_model, itemOptions, 1)

	solution, log := inputBuilder.RunHighs()
	printer.AppendOther(log)
	printer.Printf("SOLUTION STATUS = %s\n", solution.Status.String())

	debugPrint(solution, setup, printer)

	if solution.HasSolution() {
		result := setup.buildResultSet(solution, itemOptions, gear_model)
		checkSetRatingIsObjective(solution, &result, gear_model)
		return util.Optional_OfValue(result)
	} else {
		return util.Optional_Empty[items.SolvableItemSet]()
	}
}

func setupBonusedInputs(inputBuilder *utilhighs.InputBuilder, gear_model *gear_model.Model, itemOptions *items.SolvableOptionsMap, scaleOutputRating float64) *setupInputsForSetBonus {
	setup := setupInputsForSetBonus{input: inputBuilder}

	setup.addMainOutputVariable(scaleOutputRating)
	setup.addSumRatingVariable()
	setup.prepareActiveSets(gear_model)
	setup.prepareUniqueEquipped(itemOptions)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		setup.addItem(slot, item, gear_model)
	}

	setup.finishItems(itemOptions, gear_model)

	if gear_model.SetBonusRequired > 0 {
		if len(setup.setData) != 1 {
			panic("set bonus required only available for single set")
		}
		setCountCol := setup.setData[0].setTotalCountVar
		rowSetCountRequired := utilhighs.ConstraintRowBuild{}
		rowSetCountRequired.Add(setCountCol.columnIndex, 1)
		rowSetCountRequired.Finish(setup.input, float64(gear_model.SetBonusRequired), utilhighs.C_PlusInf)
	}

	return &setup
}

func debugPrint(solution *highs.Solution, setup *setupInputsForSetBonus, printer *util.PrintRecorder) {
	if !utilhighs.C_DebugHighs {
		return
	}

	printer.Printf("OBJECTIVE VALUE = %f\n", solution.Objective*c_scaled_ratings)

	activeBonus := ""
	activeBonusWeight := 0.0

	for colIndex, outputValue := range solution.ColValues {
		if !debugPrintColumn(setup.allColumns, utilhighs.ColumnIndex(colIndex), outputValue, &activeBonus, &activeBonusWeight, printer) {
			printer.Printf("%d %f UNKNOWN\n", colIndex, outputValue)
		}
	}

	printer.Printf("ACTIVE highs Bonus = %s %f\n", activeBonus, activeBonusWeight)
}

func debugPrintColumn(allColumns []columnInfo, columnIndex utilhighs.ColumnIndex, outputValue float64, activeBonus *string, activeBonusWeight *float64, printer *util.PrintRecorder) bool {
	var colEntry columnInfo
	found := false
	for _, col := range allColumns {
		if col.columnIndex == columnIndex {
			colEntry = col
			found = true
			break
		}
	}

	if found {
		debugPrintColumnEntry(colEntry, columnIndex, outputValue, activeBonus, activeBonusWeight, printer)
	}
	return found
}

func debugPrintColumnEntry(colEntry columnInfo, columnIndex utilhighs.ColumnIndex, outputValue float64, activeBonus *string, activeBonusWeight *float64, printer *util.PrintRecorder) {
	switch colEntry.entryType {
	case entry_item:
		printer.Printf("%d %f %s %s %d\n", columnIndex, outputValue, "item", colEntry.itemSlot.Name(), colEntry.item.ItemId())
	case entry_set_total_count:
		printer.Printf("%d %f %s %s\n", columnIndex, outputValue, "set total count", colEntry.set.Name())
	case entry_set_exact_count:
		printer.Printf("%d %f %s %s %d\n", columnIndex, outputValue, "set exact count flag", colEntry.set.Name(), colEntry.itemCount)
	case entry_sum_rating:
		printer.Printf("%d %f %s\n", columnIndex, outputValue, "initial item rating sum")
	case entry_permutation_active:
		printer.Printf("%d %f %s %s\n", columnIndex, outputValue, "permutation active", colEntry.permutation.debugStr())
		if utilhighs.FloatEqualsOne(outputValue) && activeBonus != nil {
			*activeBonus += colEntry.permutation.debugStr()
		}
	case entry_permutation_output_weighted:
		printer.Printf("%d %f %s %s %f\n", columnIndex, outputValue, "permutation weighted output", colEntry.permutation.debugStr(), colEntry.weight)
		if !utilhighs.FloatEqualsZero(outputValue) && activeBonusWeight != nil {
			*activeBonusWeight += colEntry.weight
		}
	case entry_main_output:
		printer.Printf("%d %f %s\n", columnIndex, outputValue, "final value")
	case entry_multi_enable_forge:
		printer.Printf("%d %f %s %d\n", columnIndex, outputValue, "multi enable forge", colEntry.itemFull.ItemId())
	case entry_multi_output:
		printer.Printf("%d %f %s\n", columnIndex, outputValue, "multi output")
	default:
		panic("unknown column")
	}
}

type setupInputsForSetBonus struct {
	input *utilhighs.InputBuilder

	slotsOneEachRow [16]utilhighs.ConstraintRowBuild // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot

	hitValueRow     utilhighs.ConstraintRowBuild // constrains values for the hits of each item
	expertValueRow  utilhighs.ConstraintRowBuild // constrains values for the expertise of each item
	minimumValueRow utilhighs.ConstraintRowBuild // when an extra minimum is specified

	baseRatingSumRow utilhighs.ConstraintRowBuild // values for the ratings of each item
	baseRatingSumVar *columnInfo                  // sum of values for the ratings of selected items

	mainOutputRow utilhighs.ConstraintRowBuild // compute final output from set based alternatives
	mainOutputVar *columnInfo                  // output variable, to be used directly or scaled against other models

	uniqueEquipRows    map[items.ItemId]*utilhighs.ConstraintRowBuild // lookup by id, may have multiple mappings for an item so need pointers
	uniqueEquipRowsAll []*utilhighs.ConstraintRowBuild                // definitive copy of each unique equip row constraint

	setData           []setInfo
	allSetPermutation []setPermutation

	itemColumns util.MapSlice[items.ItemId, columnInfo]
	allColumns  []columnInfo
}

type setPermutation struct {
	content []setWithCount

	outputVar     utilhighs.ColumnIndex
	activatingVar utilhighs.ColumnIndex
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
	c_scaled_ratings = 10000000.0 // try to make highs happier

	// example rating      178237915
	//                     187513497
	c_ratings_low_range  = 10000000.0 / c_scaled_ratings
	c_ratings_high_range = 1000000000000.0 / c_scaled_ratings
)

func (setup *setupInputsForSetBonus) buildSimpleNoSetsOutput() {
	// just copy initial rating sum into final if no sets
	setup.mainOutputRow.Add(setup.baseRatingSumVar.columnIndex, 1)
}

func (setup *setupInputsForSetBonus) buildSetMultipliedOutput(permutation *setPermutation) {
	outputVar, weight := setup.buildSetWeightedOutputVar(permutation)
	activatingVar := setup.buildPermutationActivatingVar(permutation)

	// copy regular rating sum to column if flag is set
	utilhighs.ContraintIfBoolCopyValueElseZero(setup.input, activatingVar, setup.baseRatingSumVar.columnIndex, outputVar, c_ratings_low_range, c_ratings_high_range)

	// add scaled rating to final computation
	setup.mainOutputRow.Add(outputVar, weight)

	permutation.outputVar = outputVar
	permutation.activatingVar = activatingVar
	permutation.weight = weight
}

func (setup *setupInputsForSetBonus) buildSetWeightedOutputVar(permutation *setPermutation) (utilhighs.ColumnIndex, float64) {
	totalWeight := 1.0
	for _, setAndCount := range permutation.content {
		bonusForCount := setAndCount.setInfo.activeSet.BonusForCount(uint8(setAndCount.count))
		totalWeight *= float64(bonusForCount)
	}

	// the actual output variable from this permutation, applies relevant set related multipliers
	columnIndex := setup.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
	entry := columnInfo{entryType: entry_permutation_output_weighted, columnIndex: columnIndex, permutation: permutation, weight: totalWeight}
	setup.allColumns = append(setup.allColumns, entry)
	return columnIndex, totalWeight
}

func (setup *setupInputsForSetBonus) buildPermutationActivatingVar(permutation *setPermutation) utilhighs.ColumnIndex {
	// we are effecively building a logical AND between these vars

	permutationActiveBool := setup.input.CreateColumnBool()

	buildAnd := utilhighs.ContraintAndBuilder{}
	buildAnd.SetOutput(permutationActiveBool)

	for _, setAndCount := range permutation.content {
		setInfo := setAndCount.setInfo
		count := setAndCount.count
		specificExactBool := setInfo.setExactCountVars[count]
		buildAnd.AddInput(specificExactBool.columnIndex)
	}

	buildAnd.FinishAndApply(setup.input)

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
	columnIndex := setup.input.CreateColumnGeneral(highs.Integer, 0, c_maxSetItems)

	// add corresponding -1 to the set item count array, so we can compare value to the sum of items, relevant items will flag a 1
	info.countSetItemsRow.Add(columnIndex, -1)

	// save reference
	entry := columnInfo{entryType: entry_set_total_count, columnIndex: columnIndex, set: info.activeSet}
	info.setTotalCountVar = &entry
	setup.allColumns = append(setup.allColumns, entry)
}

func (setup *setupInputsForSetBonus) addSetItemsCountExactVariables(info *setInfo) {
	// compare total number of items previous computed into this constraint
	compareRow := utilhighs.ConstraintRowBuild{}
	compareRow.Add(info.setTotalCountVar.columnIndex, -1)

	// constraint so only one of these flags gets set
	singleFlagOnly := utilhighs.ConstraintRowBuild{}

	// make a bool for each possible count in range 0..5
	for itemCount := 0; itemCount <= c_maxSetItems; itemCount++ {
		boolColumn := setup.input.CreateColumnBool()

		// should activate this flag which will match the total count
		compareRow.Add(boolColumn, float64(itemCount))

		// but only one flag at a time
		singleFlagOnly.Add(boolColumn, 1)

		entry := columnInfo{entryType: entry_set_exact_count, columnIndex: boolColumn, set: info.activeSet, itemCount: itemCount}
		info.setExactCountVars[itemCount] = entry
		setup.allColumns = append(setup.allColumns, entry)
	}

	compareRow.Finish(setup.input, 0, 0)     // equal
	singleFlagOnly.Finish(setup.input, 1, 1) // sum of flags should be just one, should pull the zero flag up if no other set
}

func (setup *setupInputsForSetBonus) addMainOutputVariable(scaleOutputRating float64) {
	// goes directly into overall rating, but could have an external scale applied
	columnIndex := setup.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, scaleOutputRating)

	// derive value based on whichever setup bonus permutation is active
	setup.mainOutputRow.Add(columnIndex, -1)

	// save reference
	entry := columnInfo{entryType: entry_main_output, columnIndex: columnIndex}
	setup.mainOutputVar = &entry
	setup.allColumns = append(setup.allColumns, entry)
}

func (setup *setupInputsForSetBonus) addSumRatingVariable() {
	// sum of individual selected item ratings
	// doesen't go directly into output rating
	columnIndex := setup.input.CreateColumnGeneral(highs.Continuous, 0, utilhighs.C_PlusInf)

	// main action of this variable: derive value to match rest of rest of row sum
	setup.baseRatingSumRow.Add(columnIndex, -1)

	// save reference
	entry := columnInfo{entryType: entry_sum_rating, columnIndex: columnIndex}
	setup.baseRatingSumVar = &entry
	setup.allColumns = append(setup.allColumns, entry)
}

func (setup *setupInputsForSetBonus) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, gear_model *gear_model.Model) utilhighs.ColumnIndex {
	// boolean value to flag use of specific item
	// contributes 0 to final rating itself, but via additional summation and calcs
	columnIndex := setup.input.CreateColumnBool()

	// add rating via a summation condition
	// scale down ratings to keep numbers small for solver stability
	rating := float64(gear_model.CalcRatingSolveItem(item)) / c_scaled_ratings
	setup.baseRatingSumRow.Add(columnIndex, rating)

	// specific hit/expertise values for hi/lo limits
	setup.hitValueRow.Add(columnIndex, float64(item.Total().Hit()))
	setup.expertValueRow.Add(columnIndex, float64(item.Total().Expertise()))

	// additional minimum value (e.g. haste)
	additionalMinimum := gear_model.StatRequirements.AdditionalMinimumRequirement
	if additionalMinimum != nil {
		setup.minimumValueRow.Add(columnIndex, float64(item.Total().Get(additionalMinimum.StatType)))
	}

	// 1 for that slot that matches the item, so we can tell solver only one item per slot
	setup.slotsOneEachRow[itemSlot].Add(columnIndex, 1.0)

	// if this item belongs to any item set then flag with a 1
	activeSetIndex, hasSet := gear_model.SetBonus.ActiveSetIndexForItem(item.ItemId())
	if hasSet {
		setup.setData[activeSetIndex].countSetItemsRow.Add(columnIndex, 1)
	}

	// if this item is unique equipped (mostly checked for ring/trinket)
	uniqueRow := setup.uniqueEquipRows[item.ItemId()]
	if uniqueRow != nil {
		uniqueRow.Add(columnIndex, 1)
	}

	entry := columnInfo{entryType: entry_item, columnIndex: columnIndex, itemSlot: itemSlot, item: item}
	setup.allColumns = append(setup.allColumns, entry)
	setup.itemColumns.Add(item.ItemId(), entry)

	return columnIndex
}

func (setup *setupInputsForSetBonus) finishItems(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model) {
	// constrain: exactly one item for each slot
	for slot, row := range setup.slotsOneEachRow {
		if itemOptions.Has(items.SlotEquip(slot)) {
			row.Finish(setup.input, 1, 1)
		} else {
			row.Finish(setup.input, 0, 0)
		}
	}

	// constrain: total sum of hit/exp are within requested limits
	setup.hitValueRow.Finish(setup.input, float64(gear_model.StatRequirements.HitMin()), float64(gear_model.StatRequirements.HitMax()))
	setup.expertValueRow.Finish(setup.input, float64(gear_model.StatRequirements.ExpertMin()), float64(gear_model.StatRequirements.ExpertMax()))

	// constrain: additional minimum value if specified has required minimum
	additionalMinimum := gear_model.StatRequirements.AdditionalMinimumRequirement
	if additionalMinimum != nil {
		setup.minimumValueRow.Finish(setup.input, float64(additionalMinimum.Value), utilhighs.C_PlusInf)
	}

	// constrain: matching sum to individual ratings
	setup.baseRatingSumRow.Finish(setup.input, 0, 0)

	// constrain: matching number of items from each given set
	for _, setInfo := range setup.setData {
		setInfo.countSetItemsRow.Finish(setup.input, 0, 0)
	}

	// constrain: unique item by itemid/unique set
	for _, row := range setup.uniqueEquipRowsAll {
		row.Finish(setup.input, 0, 1)
	}

	// constrain: whichever alternate set output into final
	setup.mainOutputRow.Finish(setup.input, 0, 0)
}

func (setup *setupInputsForSetBonus) prepareUniqueEquipped(itemOptions *items.SolvableOptionsMap) {
	setup.uniqueEquipRows = make(map[items.ItemId]*utilhighs.ConstraintRowBuild)
	setup.uniqueEquipRowsAll = make([]*utilhighs.ConstraintRowBuild, 0)
	seen := make(map[items.ItemId]bool)

	// add items from predefined unique equipped sets
	for _, set := range items.UniqueItemIdSets {
		row := new(utilhighs.ConstraintRowBuild)
		setup.uniqueEquipRowsAll = append(setup.uniqueEquipRowsAll, row)

		for _, itemId := range set {
			if seen[itemId] {
				panic("unique equipped data has duplicate")
			}
			setup.uniqueEquipRows[itemId] = row
			seen[itemId] = true
		}
	}

	// add all remaining Ring/Trinket items
	pairedSlots := []items.SlotEquip{items.Equip_Ring1, items.Equip_Ring2, items.Equip_Trinket1, items.Equip_Trinket2}
	for _, slotEquip := range pairedSlots {
		for item := range itemOptions.SlotItemSeq(slotEquip) {
			itemId := item.ItemId()
			if !seen[itemId] {
				row := new(utilhighs.ConstraintRowBuild)
				setup.uniqueEquipRowsAll = append(setup.uniqueEquipRowsAll, row)
				setup.uniqueEquipRows[itemId] = row
				seen[itemId] = true
			}
		}
	}
}

func (setup *setupInputsForSetBonus) buildResultSet(solution *highs.Solution, itemOptions *items.SolvableOptionsMap, model *gear_model.Model) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for columnEntry := range setup.itemColumns.SeqValues() {
		variableResult := solution.ColValues[columnEntry.columnIndex]
		if columnEntry.entryType == entry_item && utilhighs.FloatEqualsOne(variableResult) {
			itemSet.AddItem_DeferCalc_ExpectEmpty(columnEntry.itemSlot, columnEntry.item)
		}
	}
	items.SolvableItemSet_RecalculateTotal(&itemSet)

	validateNewSet(itemSet, itemOptions, model)
	setup.checkActivePermutation(solution, &itemSet)

	return itemSet
}

func validateNewSet(itemSet items.SolvableItemSet, itemOptions *items.SolvableOptionsMap, model *gear_model.Model) {
	itemSet.DebugValidate()
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		if itemOptions.Has(slot) != itemSet.Items().Has(slot) {
			panic("expected slots not filled")
		}
	}

	if !model.CheckSet(&itemSet) {
		sb := util.StringBuild2{}
		sb.WriteString("set fails standard CheckSet ")
		sb.WriteUint32(itemSet.Total().Hit())
		sb.WriteRune(' ')
		sb.WriteUint32(itemSet.Total().Expertise())
		panic(sb.String())
	}
}

func checkSetRatingIsObjective(solution *highs.Solution, itemSet *items.SolvableItemSet, gear_model *gear_model.Model) {
	checkRating := gear_model.CalcRatingSolve(itemSet)
	if !utilhighs.FloatsApproxEquals(solution.Objective*c_scaled_ratings, float64(checkRating)) {
		panic("rating inconsistent " + strconv.FormatFloat(solution.Objective, 'f', 0, 64) + " " + strconv.FormatFloat(float64(checkRating), 'f', 0, 32))
	}
}

func (setup *setupInputsForSetBonus) checkActivePermutation(solution *highs.Solution, solvableItemSet *items.SolvableItemSet) {
	if len(setup.allSetPermutation) > 0 {
		var activePermutation *setPermutation

		for _, permutation := range setup.allSetPermutation {
			variableResult := solution.ColValues[permutation.activatingVar]
			if utilhighs.FloatEqualsOne(variableResult) {
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
	entry_main_output                 entryType = iota
	entry_multi_enable_forge          entryType = iota
	entry_multi_output                entryType = iota
)

type columnInfo struct {
	entryType   entryType
	columnIndex utilhighs.ColumnIndex

	itemSlot items.SlotEquip
	item     *items.SolvableItem
	itemFull *items.FullItem

	set         gear_model.ActiveSet
	itemCount   int
	permutation *setPermutation
	weight      float64
}

type setInfo struct {
	activeSet gear_model.ActiveSet
	setIndex  int

	countSetItemsRow  utilhighs.ConstraintRowBuild // use to count items used from this set, has 1 or 0 flags
	setTotalCountVar  *columnInfo                  // total count of items used
	setExactCountVars [c_setItemsCounts]columnInfo // specific bools for different counts
}
