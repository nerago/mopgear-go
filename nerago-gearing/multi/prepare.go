package multi

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
)

const ReGem_GemAlternate = "GemAlternate"
const ReGem_GemDefault = "GemDefault"

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

	for i := range job.params {
		job.params[i].runBaseline()
	}

	job.prepareRatingMultipliers()

	for i := range job.params {
		job.params[i].setupAlternateGemming(job.alternateGemming)
	}
}

func (param *multiSetParamInternal) prepareStartingGear() {
	param.job.printer.Println(param.Label)

	equipped := loaders.GearFileReader_Read(param.GearFile)
	param.exactEquippedGear = setup.OptionsSetup_ExactEquippedOnly(equipped, &param.Model, param.MissingEnchant, param.job.printer)
	param.itemOptions = setup.OptionsSetup_FromEquipped(equipped, &param.Model, param.MissingEnchant, param.job.printer)

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

	param.replicatePairedOptions()
}

func (param *multiSetParamInternal) replicatePairedOptions() {
	// risky compared to old algorithms, but now we have better unique equipped maybe ok?
	param.replicatePairedSlots(items.Equip_Ring1, items.Equip_Ring2)
	param.replicatePairedSlots(items.Equip_Trinket1, items.Equip_Trinket2)
	param.itemOptions.RemoveDuplicates()
}

func (param *multiSetParamInternal) replicatePairedSlots(slotA items.SlotEquip, slotB items.SlotEquip) {
	optA := param.itemOptions.Get(slotA)
	optB := param.itemOptions.Get(slotB)
	combined := slices.Concat(optA, optB)
	param.itemOptions[slotA] = combined
	param.itemOptions[slotB] = combined
}

func (param *multiSetParamInternal) includeExtra(itemId items.ItemId) {
	if param.itemOptions.IncludesItemId(itemId) {
		param.job.printer.Printf("EXTRA already included %d\n", itemId)
		return
	}

	basicVersion := db.WowSimDB_ByIdAndUpgrade(itemId, 0)
	if param.itemOptions.CouldAddUpgrade_ItemSlot(basicVersion.SlotItem(), basicVersion, param.job.printer) == items.CanUpgrade_InvalidAlways {
		return
	}

	if param.copyExtraFromOtherSpec(itemId) {
		return
	}

	if param.copyExtraFromBags(itemId) {
		return
	}

	param.extraLoadAndGenerate(itemId, items.NO_RANDOM_SUFFIX)
}

func (param *multiSetParamInternal) copyExtraFromOtherSpec(itemId items.ItemId) bool {
	usageGroups, hasUsageGroups := param.job.distinctUsageGroups[itemId]

	options := make([]items.FullItem, 0)
	if !hasUsageGroups {
		for otherIndex := range param.job.params {
			more := param.job.params[otherIndex].itemOptions.FindItemId(itemId)
			options = slices.AppendSeq(options, more)
		}
	} else {
		if slices.Contains(usageGroups.groupAIndexes, param.paramIndex) {
			for _, otherIndex := range usageGroups.groupAIndexes {
				more := param.job.params[otherIndex].itemOptions.FindItemId(itemId)
				options = slices.AppendSeq(options, more)
			}
		} else if slices.Contains(usageGroups.groupBIndexes, param.paramIndex) {
			for _, otherIndex := range usageGroups.groupBIndexes {
				more := param.job.params[otherIndex].itemOptions.FindItemId(itemId)
				options = slices.AppendSeq(options, more)
			}
		} else {
			panic("expected param index to be in one of the groups")
		}
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
		// bags file doesn't have proper upgrade steps
		loaders.BagsFileItemSetExtraDefaults(equipped, param.ExtraUpgradeLevel)

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
		loaders.BagsFileItemSetExtraDefaults(equipped, param.ExtraUpgradeLevel)

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

func (param *multiSetParamInternal) extraLoadAndGenerate(itemId items.ItemId, randomSuffix items.RandomSuffix) {
	options, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, param.ExtraUpgradeLevel, randomSuffix, &param.Model, param.job.printer)
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
					panic("item included in slot restrictions but not actually available option " + param.Label + " " + itemId.String())
				}
			}

			param.itemOptions.FilterSlot(slot, func(x *items.FullItem) bool { return slices.Contains(itemIdList, x.ItemId()) })
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

func (param *multiSetParamInternal) setupAlternateGemming(alternateGemList []stats.GemInfo) {
	if len(alternateGemList) > 0 {
		for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
			existing := param.itemOptions.Get(slot)
			for _, item := range existing {
				for _, alternateGem := range alternateGemList {
					alternateItem := param.regemAlternate(item, alternateGem)
					param.itemOptions.AddSeveralOptionsSpecific(slot, []items.FullItem{alternateItem})
				}

				defaultItem := param.regemDefault(item)
				param.itemOptions.AddSeveralOptionsSpecific(slot, []items.FullItem{defaultItem})
			}
		}
		param.itemOptions.RemoveDuplicates()
	}
}

func (param *multiSetParamInternal) regemAlternate(item items.FullItem, alternateGem stats.GemInfo) items.FullItem {
	alternateEquipItem := loaders.EquippedItem_FromFull(&item)
	for i := range alternateEquipItem.GemChoice {
		socket := item.SocketSlots()[i]
		if socket.IsStandard() {
			alternateEquipItem.GemChoice[i] = alternateGem.Id
		}
	}
	alternateItem := setup.OptionsSetup_ExactEquippedOnly_Item(alternateEquipItem, setup.MissingEnchant_Panic, &param.Model, param.job.printer)

	alternateItem.SetNameTag(ReGem_GemAlternate)
	return alternateItem
}

func (param *multiSetParamInternal) regemDefault(item items.FullItem) items.FullItem {
	defaultEquipItem := loaders.EquippedItem_FromFull(&item)
	defaultEquipItem.GemChoice = nil
	defaultItem := setup.OptionsSetup_ExactEquippedOnly_Item(defaultEquipItem, setup.MissingEnchant_Fix, &param.Model, param.job.printer)
	defaultItem.SetNameTag(ReGem_GemDefault)
	return defaultItem
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
