package withhighs

import (
	"fmt"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"

	"github.com/lanl/highs"
)

// see wowsim-external\ui\core\components\suggest_reforges_action.tsx
// wowsim-external/ui/worker/lp_format.ts

func run(itemOptions *items.SolvableOptionsMap, model *model.Model) {
	constraints := buildConstraint{}
	constraints.init(itemOptions)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		constraints.addItem(slot, item, model)
	}

	constraints.finishItems()

	m := constraints.model
	sol, err := m.Solve()
	if err != nil {
		panic(err)
	}
	fmt.Println(sol.Status.String())
	for _, x := range sol.ColumnPrimal {
		fmt.Println(x)
	}
}

type entryType int8

const (
	entry_item           entryType = iota
	entry_item_extracopy entryType = iota
	entry_set            entryType = iota
)

type lookupEntry struct {
	purpose  entryType
	itemSlot items.SlotEquip
	item     *items.SolvableItem
	set      *model.ActiveSet
}

type buildConstraint struct {
	model highs.Model

	hitValue    []float64 // values for the hits of each item
	expertValue []float64 // values for the expertise of each item

	slotsOneEach [16][]float64 // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot

	// something setbonus membership.
	// we could use a second copy of the item with inflated weight but constrainted around count of set items
	// another variable is integer count of items used from set. needs to both allow the higher-weight variants, and also check we have enough set items to meet it

	variableLookup []lookupEntry
}

func (cons *buildConstraint) init(itemOptions *items.SolvableOptionsMap) {
	cons.model.Maximize = false

	// itemCount := itemOptions.TotalItemCount()

	// for slot := range cons.slotsOneEach {
	// 	cons.slotsOneEach[slot] = make([]float64, 0, itemCount)
	// }

	// cons.hitValue = make([]float64, 0, itemCount)
	// cons.expertValue = make([]float64, 0, itemCount)
	// cons.variableLookup = make([]lookupEntry, 0, itemCount)
}

func (cons *buildConstraint) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, model *model.Model) {
	// item version "boolean" (0 or 1)
	cons.model.VarTypes = append(cons.model.VarTypes, highs.IntegerType)
	cons.model.ColLower = append(cons.model.ColLower, 0)
	cons.model.ColUpper = append(cons.model.ColUpper, 1)

	// the "objective function" value, or weighted stat value in our terms
	rating := model.CalcRatingSolveItem(item)
	cons.model.ColCosts = append(cons.model.ColCosts, float64(rating))

	cons.hitValue = append(cons.hitValue, float64(item.TotalCap().Hit()))
	cons.expertValue = append(cons.expertValue, float64(item.TotalCap().Expertise()))

	// model.SetBonus.ActiveSetForItem(item.ItemId())

	// 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		addValue := 0.0
		if slot == itemSlot {
			addValue = 1.0
		}
		cons.slotsOneEach[slot] = append(cons.slotsOneEach[slot], addValue)
	}

	entry := lookupEntry{purpose: entry_item, itemSlot: itemSlot, item: item}
	cons.variableLookup = append(cons.variableLookup, entry)
}

func (cons *buildConstraint) finishItems(model *model.Model) {
	cons.model.AddDenseRow(float64(model.StatRequirements.HitMin()), cons.hitValue, float64(model.StatRequirements.HitMax()))
	cons.model.AddDenseRow(float64(model.StatRequirements.ExpertMin()), cons.expertValue, float64(model.StatRequirements.ExpertMax()))

	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		cons.model.AddDenseRow(1, cons.slotsOneEach[slot], 1)
	}
}
