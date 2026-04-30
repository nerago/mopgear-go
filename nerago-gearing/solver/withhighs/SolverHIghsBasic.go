package withhighs

import (
	"fmt"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/util"

	"github.com/lanl/highs"
)

func RunBasic(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requiredSet gear_model.ActiveSet, requireSetCount util.Optional[int]) util.Optional[items.SolvableItemSet] {
	constraints := buildConstraintForBasic{}
	constraints.init()

	for slot, item := range itemOptions.AllItemSlotSeq() {
		constraints.addItem(slot, item, gear_model, requiredSet)
	}

	constraints.vars.apply(constraints.param)
	constraints.finishItems(itemOptions, gear_model, requireSetCount)

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

type buildConstraintForBasic struct {
	param *highs.RawModel

	vars variableArrayBuilder

	slotsOneEachRow     [16]constraintRowSequential // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	requireSetCountsRow constraintRowSequential     // if requireSetCount then constrain set count to match
	hitValueRow         constraintRowSequential     // values for the hits of each item
	expertValueRow      constraintRowSequential     // values for the expertise of each item

	itemLookup []lookupEntryBasic
}

func (cons *buildConstraintForBasic) init() {
	cons.param = highs.NewRawModel()
	err := cons.param.SetMaximization(true)
	if err != nil {
		panic(err)
	}
}

func (cons *buildConstraintForBasic) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, gear_model *gear_model.Model, requiredSet gear_model.ActiveSet) {
	rating := float64(gear_model.CalcRatingSolveItemAsFloat(item))

	// item version "boolean" (0 or 1)
	cons.vars.add(highs.IntegerType, 0, 1, rating)

	// specific hit/expertise values for hi/lo limits
	cons.hitValueRow.add(float64(item.TotalCap().Hit()))
	cons.expertValueRow.add(float64(item.TotalCap().Expertise()))

	// 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		addValue := 0.0
		if slot == itemSlot {
			addValue = 1.0
		}
		cons.slotsOneEachRow[slot].add(addValue)
	}

	// if this item belongs to target item set then flag with a 1
	itemSet, _ := gear_model.SetBonus.ActiveSetForItem(item.ItemId())
	if itemSet != nil && requiredSet != nil && itemSet.Equals(requiredSet) {
		cons.requireSetCountsRow.add(1)
	} else {
		cons.requireSetCountsRow.add(0)
	}

	entry := lookupEntryBasic{itemSlot, item}
	cons.itemLookup = append(cons.itemLookup, entry)
}

func (cons buildConstraintForBasic) finishItems(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requireSetCount util.Optional[int]) {
	for slot, row := range cons.slotsOneEachRow {
		if len(itemOptions[slot]) > 0 {
			row.finish(cons.param, 1, 1)
		}
	}

	cons.hitValueRow.finish(cons.param, float64(gear_model.StatRequirements.HitMin()), float64(gear_model.StatRequirements.HitMax()))
	cons.expertValueRow.finish(cons.param, float64(gear_model.StatRequirements.ExpertMin()), float64(gear_model.StatRequirements.ExpertMax()))

	if require, hasRequire := requireSetCount.GetWithFlag(); hasRequire {
		cons.requireSetCountsRow.finish(cons.param, float64(require), float64(require))
	}
}

func (cons *buildConstraintForBasic) buildResultSet(solution *highs.Solution) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for colIndex, variableResult := range solution.ColumnPrimal {
		if variableResult == 1.0 {
			entry := cons.itemLookup[colIndex]
			itemSet.AddItem_DeferCalc_ExpectEmpty(entry.itemSlot, entry.item)
		}
	}
	items.SolvableItemSet_RecalculateTotal(&itemSet)
	return itemSet
}

type lookupEntryBasic struct {
	itemSlot items.SlotEquip
	item     *items.SolvableItem
}

type constraintRowSequential struct {
	insertColumn  int
	columnNumbers []int
	values        []float64
}

func (row *constraintRowSequential) add(value float64) {
	if value != 0.0 {
		row.columnNumbers = append(row.columnNumbers, row.insertColumn)
		row.values = append(row.values, value)
	}
	row.insertColumn++
}

func (row *constraintRowSequential) finish(param *highs.RawModel, lowerBound float64, upperBound float64) {
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
