package withhighs

import (
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

type RequiredSetCounts map[gear_model.ActiveSet]int

func RunSingle(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requiredSet RequiredSetCounts, printer *util.PrintRecorder) util.Optional[items.SolvableItemSet] {
	inputBuilder := utilhighs.InputBuilder{}

	inputs := setupBasicConstraint(&inputBuilder, itemOptions, gear_model, requiredSet)
	if inputs == nil {
		return util.Optional_Empty[items.SolvableItemSet]()
	}

	solution, log := inputBuilder.RunHighs()
	printer.AppendOther(log)
	printer.Println(solution.Status.String())

	// for i, x := range solution.ColumnPrimal {
	// 	printer.Println(i, x)
	// }

	if solution.HasSolution() {
		result := inputs.buildResultSet(solution, itemOptions, gear_model)
		checkSetBonusMet(&result, requiredSet)
		return util.Optional_OfValue(result)
	} else {
		return util.Optional_Empty[items.SolvableItemSet]()
	}
}

func setupBasicConstraint(inputBuilder *utilhighs.InputBuilder, itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requiredSet RequiredSetCounts) *setupInputForBasic {
	inputs := setupInputForBasic{input: inputBuilder}
	inputs.prepareRequiredSets(requiredSet)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		inputs.addItem(slot, item, gear_model)
	}

	if inputs.finishItems(itemOptions, gear_model, requiredSet) {
		return &inputs
	} else {
		return nil
	}
}

type setupInputForBasic struct {
	input *utilhighs.InputBuilder

	slotsOneEachRow [16]utilhighs.ConstraintRowBuild                       // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	requireSetRows  map[gear_model.ActiveSet]*utilhighs.ConstraintRowBuild // if requireSetCount then constrain set count to match
	hitValueRow     utilhighs.ConstraintRowBuild                           // values for the hits of each item
	expertValueRow  utilhighs.ConstraintRowBuild                           // values for the expertise of each item

	itemLookup []lookupEntryBasic
}

func (setup *setupInputForBasic) prepareRequiredSets(requiredSet RequiredSetCounts) {
	if requiredSet != nil {
		setup.requireSetRows = make(map[gear_model.ActiveSet]*utilhighs.ConstraintRowBuild)
		for set := range requiredSet {
			setup.requireSetRows[set] = &utilhighs.ConstraintRowBuild{}
		}
	}
}

func (setup *setupInputForBasic) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, model *gear_model.Model) {
	rating := float64(model.CalcRatingSolveItem(item))

	// item version "boolean" (0 or 1)
	columnIndex := setup.input.CreateColumnWithOutput(highs.Integer, 0, 1, rating)

	// specific hit/expertise values for hi/lo limits
	setup.hitValueRow.Add(columnIndex, float64(item.Total().Hit()))
	setup.expertValueRow.Add(columnIndex, float64(item.Total().Expertise()))

	// 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		if slot == itemSlot {
			setup.slotsOneEachRow[slot].Add(columnIndex, 1)
		}
	}

	// if this item belongs to target item set then flag with a 1
	for set, row := range setup.requireSetRows {
		if set.ContainsItem(item.ItemId()) {
			row.Add(columnIndex, 1)
		}
	}

	entry := lookupEntryBasic{itemSlot, item}
	setup.itemLookup = append(setup.itemLookup, entry)
}

func (setup *setupInputForBasic) finishItems(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requiredSet RequiredSetCounts) bool {
	for slot, row := range setup.slotsOneEachRow {
		if itemOptions.Has(items.SlotEquip(slot)) {
			row.Finish(setup.input, 1, 1)
		} else if row.IsEmpty() {
			row.Finish(setup.input, 0, 0)
		} else {
			panic("unexpected items added for supposedly empty slot")
		}
	}

	setup.hitValueRow.Finish(setup.input, float64(gear_model.StatRequirements.HitMin()), float64(gear_model.StatRequirements.HitMax()))
	setup.expertValueRow.Finish(setup.input, float64(gear_model.StatRequirements.ExpertMin()), float64(gear_model.StatRequirements.ExpertMax()))

	for set, requireCount := range requiredSet {
		row := setup.requireSetRows[set]
		row.Finish(setup.input, float64(requireCount), float64(requireCount))
	}

	return true
}

func (setup *setupInputForBasic) buildResultSet(solution *highs.Solution, itemOptions *items.SolvableOptionsMap, model *gear_model.Model) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for colIndex, variableResult := range solution.ColValues {
		if utilhighs.FloatEqualsOne(variableResult) {
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
		sb.WriteUint32(itemSet.Total().Hit())
		sb.WriteRune(' ')
		sb.WriteUint32(itemSet.Total().Expertise())
		panic(sb.String())
	}
}

func checkSetBonusMet(solvableItemSet *items.SolvableItemSet, requiredSet RequiredSetCounts) {
	for set, requireCount := range requiredSet {
		actualCount := 0
		for item := range solvableItemSet.Items().AllItemSeq() {
			if set.ContainsItem(item.ItemId()) {
				actualCount++
			}
		}

		if actualCount != requireCount {
			panic("didn't add correct number of set items")
		}
	}
}

type lookupEntryBasic struct {
	itemSlot items.SlotEquip
	item     *items.SolvableItem
}
