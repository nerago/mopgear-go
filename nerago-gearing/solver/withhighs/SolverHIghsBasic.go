package withhighs

import (
	"fmt"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/util"

	"github.com/lanl/highs"
)

func RunBasic(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requiredSet gear_model.ActiveSet, requireSetCount util.Optional[int]) util.Optional[items.SolvableItemSet] {
	constraints := buildInputForBasic{}

	for slot, item := range itemOptions.AllItemSlotSeq() {
		constraints.addItem(slot, item, gear_model, requiredSet)
	}

	constraints.finishItems(itemOptions, gear_model, requireSetCount)

	highs_model := constraints.toHighsModel()
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

type buildInputForBasic struct {
	mat  constraintMatrixBuilder
	vars variableArrayBuilder

	slotsOneEachRow     [16]constraintRowBuild // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	requireSetCountsRow constraintRowBuild     // if requireSetCount then constrain set count to match
	hitValueRow         constraintRowBuild     // values for the hits of each item
	expertValueRow      constraintRowBuild     // values for the expertise of each item

	itemLookup []lookupEntryBasic
}

func (cons *buildInputForBasic) toHighsModel() *highs.RawModel {
	model := highs.NewRawModel()

	// model.SetStringOption("presolve", "off")
	model.SetBoolOption("log_to_console", true)
	model.SetIntOption("log_dev_level", 3)

	err := model.SetMaximization(true)
	if err != nil {
		panic(err)
	}

	cons.vars.applyToModel(model)
	cons.mat.finishAndApplyToModelEfficient(model)

	return model
}

func (cons *buildInputForBasic) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, gear_model *gear_model.Model, requiredSet gear_model.ActiveSet) {
	rating := float64(gear_model.CalcRatingSolveItemAsFloat(item))

	// item version "boolean" (0 or 1)
	columnIndex := cons.vars.create(highs.IntegerType, 0, 1, rating)

	// specific hit/expertise values for hi/lo limits
	cons.hitValueRow.add(columnIndex, float64(item.TotalCap().Hit()))
	cons.expertValueRow.add(columnIndex, float64(item.TotalCap().Expertise()))

	// 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		if slot == itemSlot {
			cons.slotsOneEachRow[slot].add(columnIndex, 1)
		}
	}

	// if this item belongs to target item set then flag with a 1
	itemSet := gear_model.SetBonus.ActiveSetForItem(item.ItemId())
	if itemSet != nil && requiredSet != nil && itemSet.Equals(requiredSet) {
		cons.requireSetCountsRow.add(columnIndex, 1)
	}

	entry := lookupEntryBasic{itemSlot, item}
	cons.itemLookup = append(cons.itemLookup, entry)
}

func (cons *buildInputForBasic) finishItems(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model, requireSetCount util.Optional[int]) {
	for slot, row := range cons.slotsOneEachRow {
		if itemOptions.Has(items.SlotEquip(slot)) {
			row.finish(&cons.mat, 1, 1)
		} else {
			row.finish(&cons.mat, 0, 0)
		}
	}

	cons.hitValueRow.finish(&cons.mat, float64(gear_model.StatRequirements.HitMin()), float64(gear_model.StatRequirements.HitMax()))
	cons.expertValueRow.finish(&cons.mat, float64(gear_model.StatRequirements.ExpertMin()), float64(gear_model.StatRequirements.ExpertMax()))

	if require, hasRequire := requireSetCount.GetWithFlag(); hasRequire {
		cons.requireSetCountsRow.finish(&cons.mat, float64(require), float64(require))
	}
}

func (cons *buildInputForBasic) buildResultSet(solution *highs.Solution) items.SolvableItemSet {
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
