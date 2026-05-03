package withhighs

import (
	"fmt"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/util"

	"github.com/lanl/highs"
)

func RunBasic(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requiredSet gear_model.ActiveSet, requireSetCount util.Optional[int]) util.Optional[items.SolvableItemSet] {
	inputBuilder := inputBuilder{}
	constraints := setupBasicConstraint(&inputBuilder, itemOptions, gear_model, requiredSet, requireSetCount)

	highs_model := inputBuilder.toHighsModel()
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

func setupBasicConstraint(inputBuilder *inputBuilder, itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requiredSet gear_model.ActiveSet, requireSetCount util.Optional[int]) setupInputForBasic {
	constraints := setupInputForBasic{input: inputBuilder}

	for slot, item := range itemOptions.AllItemSlotSeq() {
		constraints.addItem(slot, item, gear_model, requiredSet)
	}

	constraints.finishItems(itemOptions, gear_model, requireSetCount)
	return constraints
}

type setupInputForBasic struct {
	input *inputBuilder

	slotsOneEachRow     [16]constraintRowBuild // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	requireSetCountsRow constraintRowBuild     // if requireSetCount then constrain set count to match
	hitValueRow         constraintRowBuild     // values for the hits of each item
	expertValueRow      constraintRowBuild     // values for the expertise of each item

	itemLookup []lookupEntryBasic
}

func (setup *setupInputForBasic) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, gear_model *gear_model.Model, requiredSet gear_model.ActiveSet) {
	rating := float64(gear_model.CalcRatingSolveItemAsFloat(item))

	// item version "boolean" (0 or 1)
	columnIndex := setup.input.createColumnWithOutput(highs.IntegerType, 0, 1, rating)

	// specific hit/expertise values for hi/lo limits
	setup.hitValueRow.add(columnIndex, float64(item.TotalCap().Hit()))
	setup.expertValueRow.add(columnIndex, float64(item.TotalCap().Expertise()))

	// 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		if slot == itemSlot {
			setup.slotsOneEachRow[slot].add(columnIndex, 1)
		}
	}

	// if this item belongs to target item set then flag with a 1
	itemSet := gear_model.SetBonus.ActiveSetForItem(item.ItemId())
	if itemSet != nil && requiredSet != nil && itemSet.Equals(requiredSet) {
		setup.requireSetCountsRow.add(columnIndex, 1)
	}

	entry := lookupEntryBasic{itemSlot, item}
	setup.itemLookup = append(setup.itemLookup, entry)
}

func (setup *setupInputForBasic) finishItems(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requireSetCount util.Optional[int]) {
	for slot, row := range setup.slotsOneEachRow {
		if itemOptions.Has(items.SlotEquip(slot)) {
			row.finish(setup.input, 1, 1)
		} else {
			row.finish(setup.input, 0, 0)
		}
	}

	setup.hitValueRow.finish(setup.input, float64(gear_model.StatRequirements.HitMin()), float64(gear_model.StatRequirements.HitMax()))
	setup.expertValueRow.finish(setup.input, float64(gear_model.StatRequirements.ExpertMin()), float64(gear_model.StatRequirements.ExpertMax()))

	if require, hasRequire := requireSetCount.GetWithFlag(); hasRequire {
		setup.requireSetCountsRow.finish(setup.input, float64(require), float64(require))
	}
}

func (setup *setupInputForBasic) buildResultSet(solution *highs.Solution) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for colIndex, variableResult := range solution.ColumnPrimal {
		if variableResult == 1.0 {
			entry := setup.itemLookup[colIndex]
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
