package withhighs

import (
	"fmt"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/util"

	"github.com/lanl/highs"
)

// see wowsim-external\ui\core\components\suggest_reforges_action.tsx
// wowsim-external/ui/worker/lp_format.ts

type buildConstraintForBasic struct {
	param *highs.RawModel

	ColTypes []highs.VariableType // Type of each model variable
	ColCosts []float64            // Column costs (i.e., the objective function itself)
	ColLower []float64            // Column lower bounds
	ColUpper []float64            // Column upper bounds

	slotsOneEachRow     [16]constraintRow // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	requireSetCountsRow constraintRow     // if requireSetCount then constrain set count to match
	hitValueRow         constraintRow     // values for the hits of each item
	expertValueRow      constraintRow     // values for the expertise of each item

	variableLookup []lookupEntry
}

func RunBasic(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requiredSet gear_model.ActiveSet, requireSetCount util.Optional[int]) util.Optional[items.SolvableItemSet] {
	constraints := buildConstraintForBasic{}
	constraints.init(itemOptions, gear_model, requireSetCount)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		constraints.addItem(slot, item, gear_model, requiredSet)
	}

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

	result := buildResultSet(&solution.Solution, &constraints)
	return util.Optional_OfValue(result)
}

func buildResultSet(solution *highs.Solution, constraints *buildConstraintForBasic) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for colIndex, variableResult := range solution.ColumnPrimal {
		if variableResult == 1.0 {
			entry := constraints.variableLookup[colIndex]
			itemSet.AddItem_DeferCalc_ExpectEmpty(entry.itemSlot, entry.item)
		}
	}
	items.SolvableItemSet_RecalculateTotal(&itemSet)
	return itemSet
}

func (cons *buildConstraintForBasic) init(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requireSetCount util.Optional[int]) {
	cons.param = highs.NewRawModel()
	err := cons.param.SetMaximization(true)
	if err != nil {
		panic(err)
	}

	cons.hitValueRow = constraintRow_make(cons.param, float64(gear_model.StatRequirements.HitMin()), float64(gear_model.StatRequirements.HitMax()))
	cons.expertValueRow = constraintRow_make(cons.param, float64(gear_model.StatRequirements.ExpertMin()), float64(gear_model.StatRequirements.ExpertMax()))

	// constraint:
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		if itemOptions.Has(slot) {
			cons.slotsOneEachRow[slot] = constraintRow_make(cons.param, 1, 1)
		} else {
			cons.slotsOneEachRow[slot] = constraintRow_make_nil()
		}
	}

	// constrain: exact item count in specific gear set
	if require, hasRequire := requireSetCount.GetWithFlag(); hasRequire {
		cons.requireSetCountsRow = constraintRow_make(cons.param, float64(require), float64(require))
	} else {
		cons.requireSetCountsRow = constraintRow_make_nil()
	}
}

func (cons *buildConstraintForBasic) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, gear_model *gear_model.Model, requiredSet gear_model.ActiveSet) {
	rating := float64(gear_model.CalcRatingSolveItemAsFloat(item))

	// item version "boolean" (0 or 1)
	cons.ColTypes = append(cons.ColTypes, highs.IntegerType)
	cons.ColLower = append(cons.ColLower, 0)
	cons.ColUpper = append(cons.ColUpper, 1)

	// the "objective function" value, or weighted total stat rating in our terms
	cons.ColCosts = append(cons.ColCosts, rating)

	// specific hit/expertise values for hi/lo limits
	cons.hitValueRow.add(float64(item.TotalCap().Hit()))
	cons.expertValueRow.add(float64(item.TotalCap().Expertise()))
	// cons.hitValueRow.add(2)
	// cons.expertValueRow.add(3)

	// 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		addValue := 0.0
		if slot == itemSlot {
			addValue = 1.0
		}
		cons.slotsOneEachRow[slot].add(addValue)
	}

	itemSet, _ := gear_model.SetBonus.ActiveSetForItem(item.ItemId())
	if itemSet != nil && requiredSet != nil && itemSet.Equals(requiredSet) {
		cons.requireSetCountsRow.add(1)
	} else {
		cons.requireSetCountsRow.add(0)
	}

	entry := lookupEntry{purpose: entry_item, itemSlot: itemSlot, item: item}
	cons.variableLookup = append(cons.variableLookup, entry)
}

func (cons buildConstraintForBasic) finishColumns() {
	err := cons.param.AddColumnBounds(cons.ColLower, cons.ColUpper)
	if err != nil {
		panic(err)
	}

	err = cons.param.SetColumnCosts(cons.ColCosts)
	if err != nil {
		panic(err)
	}

	err = cons.param.SetIntegrality(cons.ColTypes)
	if err != nil {
		panic(err)
	}
}

func (cons buildConstraintForBasic) finishRows() {
	for _, row := range cons.slotsOneEachRow {
		row.finish()
	}

	cons.requireSetCountsRow.finish()
	cons.hitValueRow.finish()
	cons.expertValueRow.finish()
}
