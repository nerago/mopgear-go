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

	inputs := setupBasicConstraint(&inputBuilder, itemOptions, gear_model, requiredSet, requireSetCount)
	if inputs == nil {
		return util.Optional_Empty[items.SolvableItemSet]()
	}

	highs_model := inputBuilder.toHighsModel()
	solution, err := highs_model.Solve()
	fmt.Println(solution.Status.String())
	if err != nil {
		panic(err)
	}

	for i, x := range solution.ColumnPrimal {
		fmt.Println(i, x)
	}

	if solution.Status != highs.Optimal && solution.Status != highs.ObjectiveBound && solution.Status != highs.ObjectiveTarget {
		return util.Optional_Empty[items.SolvableItemSet]()
	}

	result := inputs.buildResultSet(&solution.Solution, itemOptions, gear_model)
	checkSetBonusMet(&result, requiredSet, requireSetCount)
	return util.Optional_OfValue(result)
}

func setupBasicConstraint(inputBuilder *inputBuilder, itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requiredSet gear_model.ActiveSet, requireSetCount util.Optional[int]) *setupInputForBasic {
	inputs := setupInputForBasic{input: inputBuilder}

	for slot, item := range itemOptions.AllItemSlotSeq() {
		inputs.addItem(slot, item, gear_model, requiredSet)
	}

	if inputs.finishItems(itemOptions, gear_model, requireSetCount) {
		return &inputs
	} else {
		return nil
	}
}

type setupInputForBasic struct {
	input *inputBuilder

	slotsOneEachRow     [16]constraintRowBuild // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	requireSetCountsRow constraintRowBuild     // if requireSetCount then constrain set count to match
	hitValueRow         constraintRowBuild     // values for the hits of each item
	expertValueRow      constraintRowBuild     // values for the expertise of each item

	itemLookup []lookupEntryBasic
}

func (setup *setupInputForBasic) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, model *gear_model.Model, requiredSet gear_model.ActiveSet) {
	rating := float64(model.CalcRatingSolveItemAsFloat(item))

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
	if requiredSet != nil && requiredSet.ContainsItem(item.ItemId()) {
		setup.requireSetCountsRow.add(columnIndex, 1)
	}

	entry := lookupEntryBasic{itemSlot, item}
	setup.itemLookup = append(setup.itemLookup, entry)
}

func (setup *setupInputForBasic) finishItems(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requireSetCount util.Optional[int]) bool {
	for slot, row := range setup.slotsOneEachRow {
		if itemOptions.Has(items.SlotEquip(slot)) {
			row.finish(setup.input, 1, 1)
		} else if row.isEmpty() {
			row.finish(setup.input, 0, 0)
		} else {
			panic("unexpected items added for supposedly empty slot")
		}
	}

	setup.hitValueRow.finish(setup.input, float64(gear_model.StatRequirements.HitMin()), float64(gear_model.StatRequirements.HitMax()))
	setup.expertValueRow.finish(setup.input, float64(gear_model.StatRequirements.ExpertMin()), float64(gear_model.StatRequirements.ExpertMax()))

	if require, hasRequire := requireSetCount.GetWithFlag(); hasRequire {
		if setup.requireSetCountsRow.isEmpty() && require > 0 {
			return false
		}
		setup.requireSetCountsRow.finish(setup.input, float64(require), float64(require))
	}

	return true
}

func (setup *setupInputForBasic) buildResultSet(solution *highs.Solution, itemOptions *items.SolvableOptionsMap, model *gear_model.Model) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for colIndex, variableResult := range solution.ColumnPrimal {
		if floatEqualsOne(variableResult) {
			entry := setup.itemLookup[colIndex]
			itemSet.AddItem_DeferCalc_ExpectEmpty(entry.itemSlot, entry.item)
		}
	}
	items.SolvableItemSet_RecalculateTotal(&itemSet)

	validateNewSet(itemSet, itemOptions, model)
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
		sb.WriteUint32(itemSet.TotalCap().Hit())
		sb.WriteRune(' ')
		sb.WriteUint32(itemSet.TotalCap().Expertise())
		panic(sb.String())
	}
}

func checkSetBonusMet(solvableItemSet *items.SolvableItemSet, requiredSet gear_model.ActiveSet, requireSetCount util.Optional[int]) {
	if require, hasRequire := requireSetCount.GetWithFlag(); hasRequire {
		actualCount := 0
		for item := range solvableItemSet.Items().AllItemSeq() {
			if requiredSet.ContainsItem(item.ItemId()) {
				actualCount++
			}
		}

		if actualCount != require {
			panic("didn't add correct number of set items")
		}
	}
}

type lookupEntryBasic struct {
	itemSlot items.SlotEquip
	item     *items.SolvableItem
}
