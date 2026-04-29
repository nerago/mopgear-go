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

	VarTypes []highs.VariableType // Type of each model variable
	ColCosts []float64            // Column costs (i.e., the objective function itself)
	ColLower []float64            // Column lower bounds
	ColUpper []float64

	slotsOneEachRow     [16]constraintRow // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	requireSetCountsRow constraintRow     // if requireSetCount then constrain set count to match
	hitValueRow         constraintRow     // values for the hits of each item
	expertValueRow      constraintRow     // values for the expertise of each item

	variableLookup []lookupEntry
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
	err = cons.param.SetIntegrality(cons.VarTypes)
	if err != nil {
		panic(err)
	}
}

func RunBasic(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requiredSet gear_model.ActiveSet, requireSetCount util.Optional[int]) util.Optional[items.SolvableItemSet] {
	constraints := buildConstraintForBasic{}
	constraints.init(itemOptions, gear_model)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		constraints.addItem(slot, item, gear_model, requiredSet)
	}

	constraints.finishColumns()
	constraints.finishItems(itemOptions, gear_model)
	// constraints.finishSet(requireSetCount)

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

func (cons *buildConstraintForBasic) init(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model) {
	// cons.param.Maximize = false
	cons.param = highs.NewRawModel()
	err := cons.param.SetMaximization(true)
	if err != nil {
		panic(err)
	}
}

func (cons *buildConstraintForBasic) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, gear_model *gear_model.Model, requiredSet gear_model.ActiveSet) {
	rating := float64(gear_model.CalcRatingSolveItemAsFloat(item))

	// item version "boolean" (0 or 1)
	cons.VarTypes = append(cons.VarTypes, highs.IntegerType)
	cons.ColLower = append(cons.ColLower, 0)
	cons.ColUpper = append(cons.ColUpper, 1)

	// the "objective function" value, or weighted total stat rating in our terms
	cons.ColCosts = append(cons.ColCosts, rating)

	// specific hit/expertise values for hi/lo limits
	// cons.hitValueRow.add(float64(item.TotalCap().Hit()))
	cons.hitValueRow.add(3)
	// cons.expertValueRow.add(float64(item.TotalCap().Expertise()))
	cons.expertValueRow.add(4)

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

func (cons *buildConstraintForBasic) finishSet(requireSetCount util.Optional[int]) {
	// constrain us to have exactly that many items from the set
	if require, hasRequire := requireSetCount.GetWithFlag(); hasRequire {
		requireFloat := float64(require)
		cons.param.AddDenseRow(requireFloat, cons.requireSetCountsRow.getDataChecked(), requireFloat)
	}
}

func (cons *buildConstraintForBasic) finishItems(itemOptions *items.SolvableOptionsMap, model *gear_model.Model) {
	err := cons.param.AddDenseRow(float64(model.StatRequirements.HitMin()), cons.hitValueRow.getDataChecked(), float64(model.StatRequirements.HitMax()))
	if err != nil {
		panic(err)
	}

	err = cons.param.AddDenseRow(float64(model.StatRequirements.ExpertMin()), cons.expertValueRow.getDataChecked(), float64(model.StatRequirements.ExpertMax()))
	if err != nil {
		panic(err)
	}

	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		if itemOptions.Has(slot) {
			err = cons.param.AddDenseRow(1, cons.slotsOneEachRow[slot].getDataChecked(), 1)
			if err != nil {
				panic(err)
			}
		}
	}
}
