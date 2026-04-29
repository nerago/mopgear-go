package withhighs

// see wowsim-external\ui\core\components\suggest_reforges_action.tsx
// wowsim-external/ui/worker/lp_format.ts

// type buildConstraintForSetBonus struct {
// 	param highs.Model

// 	hitValueRow    constraintRow // values for the hits of each item
// 	expertValueRow constraintRow // values for the expertise of each item

// 	slotsOneEachRow [16]constraintRow // 1 or 0 where the slot matches the item, so we can tell solver only one item per slot

// 	activeSets          []gear_model.ActiveSet
// 	setEachItemCountRow []constraintRow
// 	permuteCodesRow     constraintRow
// 	allPermutations     []multiSetPermutation

// 	variableLookup []lookupEntry
// }

// func RunSetBonus(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model) {
// 	constraints := buildConstraintForSetBonus{}
// 	constraints.init(itemOptions, gear_model)

// 	for slot, item := range itemOptions.AllItemSlotSeq() {
// 		constraints.addItem(slot, item, gear_model)
// 	}

// 	constraints.finish(gear_model)

// 	m := constraints.param
// 	sol, err := m.Solve()
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(sol.Status.String())
// 	for _, x := range sol.ColumnPrimal {
// 		fmt.Println(x)
// 	}
// }

// func (cons *buildConstraintForSetBonus) init(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.Model) {
// 	cons.param.Maximize = false

// 	activeSets := gear_model.SetBonus.ActiveSets()
// 	cons.setEachItemCountRow = make([]constraintRow, len(activeSets))
// 	for setIndex := range len(activeSets) {
// 		cons.addSetVariable(setIndex, activeSets[setIndex])
// 	}

// 	cons.allPermutations = []multiSetPermutation{{totalBonus: 1, permutationCode: 0}}

// 	// SUDDENLY NOT CONVINCED, DUE TO NOT BEING ABLE TO <= THINGS?
// }

// func (cons *buildConstraintForSetBonus) addSetVariable(activeSetIndex int, activeSet gear_model.ActiveSet) {
// 	// set item target count
// 	cons.param.VarTypes = append(cons.param.VarTypes, highs.IntegerType)
// 	cons.param.ColLower = append(cons.param.ColLower, 0)
// 	cons.param.ColUpper = append(cons.param.ColUpper, 5)

// 	// the direct value of the variable towards rating is zero, but constraints enable upscaled item versions
// 	cons.param.ColCosts = append(cons.param.ColCosts, 0)

// 	// not an item, no effect on caps
// 	cons.hitValueRow.add(0)
// 	cons.expertValueRow.add(0)

// 	// not an item, no effect on slots
// 	for slot := range cons.slotsOneEachRow {
// 		cons.slotsOneEachRow[slot].add(0)
// 	}

// 	// add corresponding -1 to the item count array, so we can compare variable value to the sum of items
// 	for index := range len(cons.setEachItemCountRow) {
// 		addValue := 0.0
// 		if index == activeSetIndex {
// 			addValue = -1.0
// 		}
// 		cons.slotsOneEachRow[index].add(addValue)
// 	}

// 	// add permute code negative
// 	codeMultiply := 1 << (c_permuteCodeShift * activeSetIndex)
// 	cons.permuteCodesRow.add(float64(-codeMultiply))

// 	// save reference
// 	entry := lookupEntry{purpose: entry_set_item_count, set: activeSet}
// 	cons.variableLookup = append(cons.variableLookup, entry)
// }

// func (cons *buildConstraintForSetBonus) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, gear_model *gear_model.Model) {
// 	rating := float64(gear_model.CalcRatingSolveItemAsFloat(item))
// 	set, setIndex := gear_model.SetBonus.ActiveSetForItem(item.ItemId())

// 	if set == nil {
// 		cons.addItemWithVariant(itemSlot, item, rating, -1, 0)
// 	} else {
// 		cons.addItemWithVariant(itemSlot, item, rating, setIndex, 0)

// 		mult2 := set.BonusForCount(2)
// 		cons.addItemWithVariant(itemSlot, item, rating*float64(mult2), setIndex, 2)

// 		mult4 := set.BonusForCount(4)
// 		cons.addItemWithVariant(itemSlot, item, rating*float64(mult4), setIndex, 4)

// 		// so only one item for each slot will light up
// 		// if we sum them for a 0,1 should be baseline items
// 		// if we sum them for a 2 should be either 2 or 3 enhanced items
// 		// alt: if we sum them for a 2 should be exactly 2 enhanced items
// 		// if we sum four of them we want ones that meet more of the product/square?

// 		// or we have separate constraints by bonus slot, then since only one will light up we could kinda do a ==count compare

// 		// if we do every slot then would need separate constrains for each permutation of set counts. every item exists in [012345]*[012345] variants
// 	}
// }

// func (cons *buildConstraintForSetBonus) addItemWithVariant(itemSlot items.SlotEquip, item *items.SolvableItem, rating float64, activeSetIndex int, permuteCode float64) {
// 	// item version "boolean" (0 or 1)
// 	cons.param.VarTypes = append(cons.param.VarTypes, highs.IntegerType)
// 	cons.param.ColLower = append(cons.param.ColLower, 0)
// 	cons.param.ColUpper = append(cons.param.ColUpper, 1)

// 	// the "objective function" value, or weighted total stat rating in our terms
// 	cons.param.ColCosts = append(cons.param.ColCosts, rating)

// 	// specific hit/expertise values for hi/lo limits
// 	cons.hitValueRow.add(float64(item.TotalCap().Hit()))
// 	cons.expertValueRow.add(float64(item.TotalCap().Expertise()))

// 	// 1 or 0 where the slot matches the item, so we can tell solver only one item per slot
// 	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
// 		addValue := 0.0
// 		if slot == itemSlot {
// 			addValue = 1.0
// 		}
// 		cons.slotsOneEachRow[slot].add(addValue)
// 	}

// 	// add corresponding 1 to the item count array, count this item in that set if included
// 	for index := range len(cons.setEachItemCountRow) {
// 		addValue := 0.0
// 		if index == activeSetIndex {
// 			addValue = 1.0
// 		}
// 		cons.setEachItemCountRow[index].add(addValue)
// 	}

// 	entry := lookupEntry{purpose: entry_item, itemSlot: itemSlot, item: item}
// 	cons.variableLookup = append(cons.variableLookup, entry)
// }

// func (cons *buildConstraintForSetBonus) finish(gear_model *gear_model.Model) {
// 	cons.finishItems(gear_model)
// 	cons.finishSets()
// }

// func (cons *buildConstraintForSetBonus) finishSets() {
// 	// constrain us to have exactly that many items from the set
// 	for setIndex := range cons.setEachItemCountRow {
// 		cons.param.AddDenseRow(0, cons.setEachItemCountRow[setIndex].getDataChecked(), 0)
// 	}
// }

// func (cons *buildConstraintForSetBonus) finishItems(model *gear_model.Model) {
// 	cons.param.AddDenseRow(float64(model.StatRequirements.HitMin()), cons.hitValueRow.getDataChecked(), float64(model.StatRequirements.HitMax()))
// 	cons.param.AddDenseRow(float64(model.StatRequirements.ExpertMin()), cons.expertValueRow.getDataChecked(), float64(model.StatRequirements.ExpertMax()))

// 	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
// 		cons.param.AddDenseRow(1, cons.slotsOneEachRow[slot].getDataChecked(), 1)
// 	}
// }
