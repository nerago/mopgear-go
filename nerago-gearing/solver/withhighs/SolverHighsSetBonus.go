package withhighs

import (
	"fmt"
	"math"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/util"

	"github.com/lanl/highs"
)

func RunAllActiveSets(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model) util.Optional[items.SolvableItemSet] {
	constraints := buildConstraintForSetBonus{}
	constraints.init(itemOptions, gear_model)
	constraints.prepareActiveSets(gear_model)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		constraints.addItem(slot, item, gear_model)
	}

	constraints.vars.apply(constraints.param)
	constraints.finishColumns()
	constraints.finishRows()

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
	return util.Optional_OfValue(result)
}

type buildConstraintForSetBonus struct {
	param *highs.RawModel

	vars variableArrayBuilder

	slotsOneEachRow      [16]constraintRowSparse // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	countEachSetItemsRow []constraintRowSparse   // count items used from each set, more 1 or 0 flags
	baseRatingSumRow     constraintRowSparse     // values for the hits of each item
	hitValueRow          constraintRowSparse     // constrains values for the hits of each item
	expertValueRow       constraintRowSparse     // constrains values for the expertise of each item

	variableLookup []lookupEntryWithSetBonuses
}

func (cons *buildConstraintForSetBonus) init(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model) {
	cons.param = highs.NewRawModel()
	err := cons.param.SetMaximization(true)
	if err != nil {
		panic(err)
	}

	activeSets := gear_model.SetBonus.ActiveSets()
	cons.countEachSetItemsRow = make([]constraintRowSparse, len(activeSets))
}

func (cons *buildConstraintForSetBonus) prepareActiveSets(gear_model *gear_model.Model) {
	// constrain: exact item count in each active set
	activeSets := gear_model.SetBonus.ActiveSets()
	if len(activeSets) == 0 {
		panic("shouldn't be in this solver without active sets")
	}
	for range activeSets {
		cons.countEachSetItemsRow = append(cons.countEachSetItemsRow, constraintRow_make(cons.param, 0, 0))
	}
	for setIndex, setInfo := range activeSets {
		cons.addSetVariable(setIndex, setInfo)
	}
}

func (cons *buildConstraintForSetBonus) addSumRatingVariable(activeSetIndex int, activeSet gear_model.ActiveSet) {
	// set item target/derived actual count, solver might approach from either direction
	cons.ColTypes = append(cons.ColTypes, highs.ContinuousType)
	cons.ColLower = append(cons.ColLower, 0)
	cons.ColUpper = append(cons.ColUpper, math.Inf(1))
	cons.ColCosts = append(cons.ColCosts, 0) // contributes 0 to rating itself

	// not an item, no effect on caps
	cons.hitValueRow.add(0)
	cons.expertValueRow.add(0)

	// not an item, no effect on slots
	for slot := range cons.slotsOneEachRow {
		cons.slotsOneEachRow[slot].add(0)
	}

	// main action of this variable: derive value to match rest of rest of row sum
	cons.baseRatingSumRow.add(-1)

	// add corresponding -1 to the item count array, so we can compare variable value to the sum of items
	for index := range len(cons.countEachSetItemsRow) {
		cons.slotsOneEachRow[index].add(0)
	}

	// save reference
	entry := lookupEntryBasic{purpose: entry_sum_rating}
	cons.variableLookup = append(cons.variableLookup, entry)
}

func (cons *buildConstraintForSetBonus) addSetVariable(activeSetIndex int, activeSet gear_model.ActiveSet) {
	// set item target/derived actual count, solver might approach from either direction
	cons.ColTypes = append(cons.ColTypes, highs.IntegerType)
	cons.ColLower = append(cons.ColLower, 0)
	cons.ColUpper = append(cons.ColUpper, c_maxSetItems)
	cons.ColCosts = append(cons.ColCosts, 0) // contributes 0 to rating itself

	// not an item, no effect on caps
	cons.hitValueRow.add(0)
	cons.expertValueRow.add(0)

	// not an item, no effect on slots
	for slot := range cons.slotsOneEachRow {
		cons.slotsOneEachRow[slot].add(0)
	}

	// not an item, no rating value
	cons.baseRatingSumRow.add(0)

	// add corresponding -1 to the item count array, so we can compare variable value to the sum of items
	cons.countEachSetItemsRow
	for index := range len(cons.countEachSetItemsRow) {
		addValue := 0.0
		if index == activeSetIndex {
			addValue = -1.0
		}
		cons.slotsOneEachRow[index].add(addValue)
	}

	// save reference
	entry := lookupEntryBasic{purpose: entry_set_item_count, set: activeSet}
	cons.variableLookup = append(cons.variableLookup, entry)
}

func (cons *buildConstraintForSetBonus) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, gear_model *gear_model.Model) {
	rating := float64(gear_model.CalcRatingSolveItemAsFloat(item))

	// boolean value to flag use of specific item
	// contributes 0 to final rating itself, but via additional summation and calcs
	columnIndex := cons.vars.add(highs.IntegerType, 0, 1, 0)

	// add rating via a summation condition
	cons.baseRatingSumRow.add(columnIndex, rating)

	// specific hit/expertise values for hi/lo limits
	cons.hitValueRow.add(columnIndex, float64(item.TotalCap().Hit()))
	cons.expertValueRow.add(columnIndex, float64(item.TotalCap().Expertise()))

	// 1 for that slot that matches the item, so we can tell solver only one item per slot
	cons.slotsOneEachRow[itemSlot].add(columnIndex, 1.0)

	// if this item belongs to any item set then flag with a 1
	activeSetIndex, hasSet := gear_model.SetBonus.ActiveSetIndexForItem(item.ItemId())
	if hasSet {
		cons.countEachSetItemsRow[activeSetIndex].add(columnIndex, 1.0)
	}

	entry := lookupEntryWithSetBonuses{entryType: entry_item, itemSlot: itemSlot, item: item}
	cons.variableLookup = append(cons.variableLookup, entry)
}

func (cons buildConstraintForSetBonus) finishRows(itemOptions *items.SolvableOptionsMap) {

	// constrain: exactly one item for each slot
	for slot, row := range cons.slotsOneEachRow {
		if len(itemOptions[slot]) > 0 {
			row.finish(cons.param, 1, 1)
		}
	}
	// slotsOneEachRow      [16]constraintRowSparse // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	// countEachSetItemsRow []constraintRowSparse   // count items used from each set, more 1 or 0 flags
	// baseRatingSumRow     constraintRowSparse     // values for the hits of each item
	// hitValueRow          constraintRowSparse     // constrains values for the hits of each item
	// expertValueRow       constraintRowSparse     // constrains values for the expertise of each item

	// constrain: total sum of hit/exp are within requested limits
	// cons.hitValueRow = constraintRow_make(cons.param, float64(gear_model.StatRequirements.HitMin()), float64(gear_model.StatRequirements.HitMax()))
	// cons.expertValueRow = constraintRow_make(cons.param, float64(gear_model.StatRequirements.ExpertMin()), float64(gear_model.StatRequirements.ExpertMax()))

	//
	// for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
	// 	if itemOptions.Has(slot) {
	// 		cons.slotsOneEachRow[slot] = constraintRow_make(cons.param, 1, 1)
	// 	} else {
	// 		cons.slotsOneEachRow[slot] = constraintRow_make_nil()
	// 	}
	// }

	// // calculate: sum the basic unscaled item ratings
	// cons.baseRatingSumRow = constraintRow_make(cons.param, 0, 0)

	// cons.requireSetCountsRow.finish()
	cons.hitValueRow.finish()
	cons.expertValueRow.finish()
}

func (cons *buildConstraintForSetBonus) buildResultSet(solution *highs.Solution) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for colIndex, variableResult := range solution.ColumnPrimal {
		if variableResult == 1.0 {
			entry := cons.variableLookup[colIndex]
			itemSet.AddItem_DeferCalc_ExpectEmpty(entry.itemSlot, entry.item)
		}
	}
	items.SolvableItemSet_RecalculateTotal(&itemSet)
	return itemSet
}

type constraintRowSparse struct {
	columnNumbers []int
	values        []float64
}

func (row *constraintRowSparse) add(columnIndex int, value float64) {
	if value != 0.0 {
		row.columnNumbers = append(row.columnNumbers, columnIndex)
		row.values = append(row.values, value)
	}
}

func (row *constraintRowSparse) finish(param *highs.RawModel, lowerBound float64, upperBound float64) {
	var err error
	if len(row.values) > 0 {
		err = param.AddCompSparseRows(
			[]float64{lowerBound},
			[]int{0},
			row.columnNumbers,
			row.values,
			[]float64{upperBound},
		)
	} else {
		// need to set an explicit zero value so array isn't empty
		// i'd argue this is a bug in go/highs binding library,
		// empty array should be acceptable to lower level code
		// maybe these don't need to be added in many use cases though, automatically skipping seems risky
		err = param.AddCompSparseRows(
			[]float64{lowerBound},
			[]int{0},
			[]int{0},
			[]float64{0.0},
			[]float64{upperBound},
		)
	}

	if err != nil {
		panic(err)
	}
}

type entryType int8

const (
	entry_item           entryType = iota
	entry_set_item_count entryType = iota
	entry_sum_rating     entryType = iota
)

type lookupEntryWithSetBonuses struct {
	entryType entryType
	itemSlot  items.SlotEquip
	item      *items.SolvableItem
	set       gear_model.ActiveSet
}
