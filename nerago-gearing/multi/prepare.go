package multi

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"
	"slices"
)

func (job *MultiSetJob) prepareInitial() {
	job.printer.Println("LOADING BAGS")
	job.bagsGear = loaders.BagsFileReader_Read()

	job.printer.Println("PREPARING STARTING GEAR")
	for i := range job.params {
		job.params[i].prepareStartingGear()
	}

	job.printer.Println("PREPARING EXTRA ITEMS")
	for i := range job.params {
		job.params[i].prepareExtraItems()
	}

	job.printer.Println("RESTRICTING ANY BLOCKED ITEMS")
	for i := range job.params {
		job.params[i].removeBlocked()
	}

	job.printer.Println("RESTRICTING ANY FIXED SLOTS")
	for i := range job.params {
		job.params[i].restrictFixed()
	}

	job.validateMultiSetAlignItemSlots()

	for i := range job.params {
		job.params[i].runBaseline()
	}

	job.prepareRatingMultipliers()
}

func (param *multiSetParamInternal) prepareStartingGear() {
	param.job.printer.Println(param.Label)

	equipped := loaders.GearFileReader_Read(param.GearFile)
	param.exactEquippedGear = setup.OptionsSetup_ExactEquippedOnly(equipped, &param.Model, param.job.printer)
	param.itemOptions = setup.OptionsSetup_FromEquipped(equipped, &param.Model, setup.MissingEnchant_Panic, param.job.printer)

	setup.UpgradeExistingToLevel2(&param.itemOptions, param.ForceUpgradeExistingItems, &param.Model, param.job.printer)
}

func (param *multiSetParamInternal) prepareExtraItems() {
	param.job.printer.Println(param.Label)

	for _, itemId := range param.ExtraItems {
		param.includeExtra(itemId)
	}

	if param.ExtraFromBags {
		for _, item := range param.job.bagsGear {
			param.tryAddExtraFromBags(&item)
		}
	}
}

func (param *multiSetParamInternal) includeExtra(itemId items.ItemId) {
	if param.itemOptions.IncludesItemId(itemId) {
		param.job.printer.Printf("EXTRA already included %d\n", itemId)
		return
	}

	if param.copyExtraFromOtherSpec(itemId) {
		return
	}

	if param.copyExtraFromBags(itemId) {
		return
	}

	param.extraLoadAndGenerate(itemId)
}

func (param *multiSetParamInternal) copyExtraFromOtherSpec(itemId items.ItemId) bool {
	options := make([]items.FullItem, 0)
	for otherIndex := range param.job.params {
		more := param.job.params[otherIndex].itemOptions.FindItemId(itemId)
		options = slices.AppendSeq(options, more)
	}

	options = util.RemoveDuplicatesFunc(options, (*items.FullItem).Equals)

	// NOTE these may not copy with the model's reforge preferences etc

	if len(options) > 0 {
		param.itemOptions.AddSeveralOptions(options[0].SlotItem(), options)
		param.job.printer.Printf("OPTION from other spec %s\n", options[0].CreateString())
		return true
	} else {
		return false
	}
}

func (param *multiSetParamInternal) copyExtraFromBags(itemId items.ItemId) bool {
	equipped := param.job.bagsGear.GetWithItemId(itemId)
	if equipped != nil {
		// bags file doesn't have upgrade steps
		equipped.UpgradeStep = param.ExtraUpgradeLevel

		options, example := setup.OptionsSetup_Single_FromEquipped(*equipped, &param.Model, setup.MissingEnchant_Fix, param.job.printer)
		param.itemOptions.AddSeveralOptions(example.SlotItem(), options)
		param.job.printer.Printf("OPTION from bags %s\n", example.CreateString())
		return true
	}
	return false
}

func (param *multiSetParamInternal) tryAddExtraFromBags(equipped *loaders.EquippedItem) {
	if db.WowSimDB_HasItemId(equipped.ItemId) {
		// bail early before considering full item stats/enchants/etc that might not fit spec
		basicVersion := db.WowSimDB_ByIdAndUpgrade(equipped.ItemId, 0)
		if param.itemOptions.CouldAddUpgrade_ItemSlot(basicVersion.SlotItem(), basicVersion, param.job.printer) != items.CanUpgrade_Yes {
			return
		}

		if param.copyExtraFromOtherSpec(equipped.ItemId) {
			return
		}

		// bags file doesn't have upgrade steps
		equipped.UpgradeStep = param.ExtraUpgradeLevel

		options, example := setup.OptionsSetup_Single_FromEquipped(*equipped, &param.Model, setup.MissingEnchant_Fix, param.job.printer)

		added := false
		for _, slot := range example.SlotItem().ToSlotEquipOptions() {
			if param.itemOptions.CouldAddUpgrade_EquipSlot(slot, example, param.job.printer) == items.CanUpgrade_Yes {
				param.job.printer.Printf("ADDITIONAL EXTRA OPTION from bags %s\n", example.CreateString())
				param.itemOptions.AddSeveralOptionsSpecific(slot, options)
				added = true
			}
		}

		if added {
			param.addedFromBags = append(param.addedFromBags, equipped.ItemId)
		}
	} else {
		param.job.printer.Printf("UNKNOWN itemid IN bags %d\n", equipped.ItemId)
	}
}

func (param *multiSetParamInternal) extraLoadAndGenerate(itemId items.ItemId) {
	options, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, param.ExtraUpgradeLevel, &param.Model, param.job.printer)
	param.itemOptions.AddSeveralOptions(example.SlotItem(), options)
	param.job.printer.Printf("OPTION %s\n", example.CreateString())
}

func (param *multiSetParamInternal) restrictFixed() {
	param.job.printer.Println(param.Label)

	for slot, itemIdList := range param.SemiFixedSlots {
		if !param.itemOptions.Has(slot) {
			panic("restricting slot but already empty")
		} else if len(itemIdList) == 0 {
			panic("empty restrict list")
		}

		if len(itemIdList) == 1 {
			param.itemOptions.ForceSlotOnlySpecifiedItemId(slot, itemIdList[0])
		} else {
			for _, itemId := range itemIdList {
				if !param.itemOptions.IncludesItemIdInSlot(itemId, slot) {
					panic("item included in slot restrictions but not actually available option " + itemId.String())
				}
			}

			param.itemOptions.FilterSlot(slot, func(x *items.FullItem) bool { return slices.Contains(itemIdList, x.ItemId()) })
		}

		paired := slot.PairedSlot()
		if paired != -1 {
			for _, itemId := range itemIdList {
				if param.itemOptions.IncludesItemIdInSlot(itemId, slot) {
					panic("item is fixed in one slot but also available in paired slot " + itemId.String())
				}
			}
		}
	}
}

func (param *multiSetParamInternal) removeBlocked() {
	// remove blocked items
	for _, itemId := range param.BlockedItems {
		param.job.printer.Printf("BLOCKING ITEM %d\n", itemId)
		param.itemOptions.RemoveItemIdFromAll(itemId)
	}
}

func (job *MultiSetJob) validateMultiSetAlignItemSlots() {
	// TODO can't remember what this was in aid of, not needed anymore?
	// maybe was in order of a warning that things might be bad downstream or explode permutations

	// NEW VERSION check within set, make things easier for solvers
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		seen := make(map[items.ItemId]items.SlotEquip)
		for slot, item := range param.itemOptions.AllItemsWithSlot() {
			seenSlot, found := seen[item.ItemId()]
			if found && seenSlot != slot {
				panic("duplicate item in different slots " + item.CreateString() + " in set " + param.Label)
			} else if !found {
				seen[item.ItemId()] = slot
			}
		}
	}

	// OLD VERSION checked across all sets
	// seen := make(map[items.ItemId]items.SlotEquip)
	// for paramIndex := range job.params {
	// 	for slot, item := range job.params[paramIndex].itemOptions.AllItemsWithSlot() {
	// 		seenSlot, found := seen[item.ItemId()]
	// 		if found && seenSlot != slot && !slices.Contains(job.suppressSlotCheck, item.ItemId()) {
	// 			panic("duplicate in non-matching slot " + item.CreateString())
	// 		} else if !found {
	// 			seen[item.ItemId()] = slot
	// 		}
	// 	}
	// }
}

func (param *multiSetParamInternal) runBaseline() {
	param.job.printer.Printf("BASELINE for %s\n", param.Label)
	param.baselineResult = solver.Solver(solver.SolveInput{
		ItemOptions:         &param.itemOptions,
		Model:               &param.Model,
		EnableTrackProgress: true,
		Printer:             param.job.printer})

	if !param.baselineResult.Success {
		panic("failed to find baseline for " + param.Label)
	}
	param.baselineResult.Report(param.job.printer)
	param.seenInSolutions.Add(&param.baselineResult.FullSet)
}

func (job *MultiSetJob) prepareRatingMultipliers() {
	var totalPercent float64
	for i := range job.params {
		job.params[i].prepareRatingMultiplier()
		totalPercent += job.params[i].RequestRatingPercent
	}

	if totalPercent < 0.99 || totalPercent > 1.01 {
		panic("percents don't add to one")
	}
}

func (param *multiSetParamInternal) prepareRatingMultiplier() {
	var targetCombined float64 = 100000000.0
	baselineRating := float64(param.baselineResult.ResultRating)

	targetForThis := targetCombined * param.RequestRatingPercent
	multiplyRatingsBy := targetForThis / baselineRating
	param.ratingMultiply = multiplyRatingsBy

	param.job.printer.Printf("MULTIPLIERS %s base=%.0f mult=%f value=%0.f percent=%.2f\n",
		param.Label, param.baselineResult.ResultRating, param.ratingMultiply,
		baselineRating*param.ratingMultiply,
		baselineRating*param.ratingMultiply/targetCombined*100,
	)
}
