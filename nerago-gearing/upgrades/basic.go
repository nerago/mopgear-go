package upgrades

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"slices"
)

func prepareUpgradeInfo(extraItems []loaders.ItemFoundRef, upgradeLevel items.UpgradeLevel, printer *util.PrintRecorder, input *FindUpgrades_BasicInputs, baseItems *items.FullOptionsMap, goal stats.OptimiseGoal, substituteItems []items.ItemId, model *gear_model.SpecModel) []upgradeItemTask {
	extraItems = changeUpgradeLevels(extraItems, upgradeLevel)
	checkDuplicates(extraItems)
	extraTasks := makeExtraTasks(input, extraItems, baseItems, printer, goal)
	addSubstituteItems(baseItems, substituteItems, model, printer)
	return extraTasks
}

func findBaseLine(printer *util.PrintRecorder, baseItems *items.FullOptionsMap, model *gear_model.SpecModel, tracker *util.TrackProgress) (float64, *items.FullItemSet) {
	printer.Println("FINDING BASELINE")
	baseRating, baseSet := findBase(baseItems, model, printer, tracker)
	tools.ReportSetFewerParams(model, baseSet, printer)
	return baseRating, baseSet
}

func changeUpgradeLevels(extraItems []loaders.ItemFoundRef, upgradeLevel items.UpgradeLevel) []loaders.ItemFoundRef {
	return util.MapSliceAsNew(extraItems, func(ref *loaders.ItemFoundRef) loaders.ItemFoundRef {
		if upgradeLevel > ref.UpgradeLevel {
			return loaders.ItemFoundRef{
				ItemId:       ref.ItemId,
				UpgradeLevel: upgradeLevel,
				RandomSuffix: ref.RandomSuffix,
			}
		} else {
			return *ref
		}
	})
}

func checkDuplicates(extraItems []loaders.ItemFoundRef) {
	byName := make(map[string]loaders.ItemFoundRef)
	for _, itemRef := range extraItems {
		itemName := db.LookupItemNameByItemId(itemRef.ItemId)
		_, alreadySeen := byName[itemName]
		if alreadySeen {
			panic("duplicate item for " + itemName)
		} else {
			byName[itemName] = itemRef
		}
	}
}

func makeExtraTasks(input *FindUpgrades_BasicInputs, extraItems []loaders.ItemFoundRef, baseItems *items.FullOptionsMap, printer *util.PrintRecorder, goal stats.OptimiseGoal) []upgradeItemTask {
	bagsFile := loaders.BagsFile_PlusPaladinGear_Read()

	taskList := make([]upgradeItemTask, 0, len(extraItems))
	for _, extra := range extraItems {
		boss := db.BossItemData_BossForItemId(extra.ItemId)
		tempItem := db.WowSimDB_LoadItemById(extra.ItemId, 0)
		for _, slot := range tempItem.SlotItem().ToSlotEquipOptions() {
			canUpgrade := canPerformSpecifiedUpgrade(input, tempItem, slot, baseItems, bagsFile, printer)
			switch canUpgrade {
			case items.CanUpgrade_Yes, items.CanUpgrade_Equipped, items.CanUpgrade_Equipped_Similar, items.CanUpgrade_AvailableInBags:
				taskList = append(taskList, upgradeItemTask{item: extra, slot: slot, goal: goal, boss: boss, canUpgrade: canUpgrade})
			}
		}
	}
	return taskList
}

func addSubstituteItems(optionsMap *items.FullOptionsMap, substituteItems []items.ItemId, model *gear_model.SpecModel, printer *util.PrintRecorder) {
	for _, itemId := range substituteItems {
		if !optionsMap.IncludesItemId(itemId) {
			// TODO system for random suffixes
			options, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, model, printer)
			optionsMap.AddSeveralOptions(example.SlotItem(), options)
			printer.Println("SUBSTITUTE " + example.CreateString())
		}
	}
}

func canPerformSpecifiedUpgrade(input *FindUpgrades_BasicInputs, extra *items.FullItem, slot items.SlotEquip, baseItems *items.FullOptionsMap, bagsFile loaders.EquippedArray, printer *util.PrintRecorder) items.CanUpgradeResult {
	if slices.Contains(input.IgnoredItems, extra.ItemId()) {
		return items.CanUpgrade_InvalidAlways
	}

	if result := baseItems.CouldAddUpgrade_EquipSlot(slot, extra, printer); result != items.CanUpgrade_Yes {
		return result
	}

	if bagsFile.HasAnyWithItemId(extra.ItemId()) {
		printer.Println("ALREADY AVAILABLE IN BAG " + extra.CreateString())
		return items.CanUpgrade_AvailableInBags
	}

	return items.CanUpgrade_Yes
}

func findBase(baseItems *items.FullOptionsMap, model *gear_model.SpecModel, printer *util.PrintRecorder, tracker *util.TrackProgress) (float64, *items.FullItemSet) {
	output := solver.Solver(solver.SolveInput{
		ItemOptions:        baseItems,
		Model:              model,
		OuterTrackProgress: tracker,
		Printer:            printer,
	})

	if !output.Success {
		panic("couldn't find valid baseline set")
	}

	printer.Printf("\n%s\nBASE RATING    = %.0f\n\n", output.SolvedSet.Total().CreateString(), output.ResultRating)
	return float64(output.ResultRating), &output.FullSet
}

func performUpgradeTask(extraTask *upgradeItemTask, baseItems *items.FullOptionsMap, baseRating float64, model *gear_model.SpecModel, parentPrinter *util.PrintRecorder, outerTracker *util.TrackProgress, forceIncludeMost bool, substituteEmptySlotOnly map[items.SlotItem]items.ItemId) upgradeItemResult {
	slot := extraTask.slot
	incompleteItem := extraTask.item // this "item" is from ItemFinder and not a full item
	itemId := incompleteItem.ItemId
	upgradeLevel := incompleteItem.UpgradeLevel
	randomSuffix := incompleteItem.RandomSuffix

	if !extraTask.actuallyAttemptUpgrade(forceIncludeMost) {
		itemName := db.LookupItemNameByItemId(itemId)
		parentPrinter.Println("SKIPPING " + itemName)
		return upgradeItemResult_OfFailure(extraTask, nil)
	}

	printer := util.PrintRecorder_HoldAll()

	newOptions, exampleItem := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, upgradeLevel, randomSuffix, model, printer)
	jobItems := baseItems.Clone()
	jobItems[slot] = newOptions

	printer.Println("OFFER " + newOptions[0].CreateString())
	printer.Println("REPLACING " + baseItems.Get(slot)[0].CreateString())

	if extraTask.canUpgrade == items.CanUpgrade_Equipped || extraTask.canUpgrade == items.CanUpgrade_Equipped_Similar {
		removePairedSimilar(&jobItems, slot, exampleItem, substituteEmptySlotOnly, model, printer)
	}

	output := solver.Solver(solver.SolveInput{
		ItemOptions:        &jobItems,
		Model:              model,
		OuterTrackProgress: outerTracker,
		Printer:            printer,
	})

	var result upgradeItemResult
	if output.Success {
		printer.Printf("SET STATS %s\n", output.SolvedSet.Total().CreateString())
		output.Report(printer) // verbose

		factor := float64(output.ResultRating) / baseRating
		printer.Printf("UPGRADE RATING = %.0f FACTOR = %1.3f\n", output.ResultRating, factor)

		setBonus := model.SetBonus.CountInAnySet(output.FullSet.Items())

		result = upgradeItemResult{upgradeItemTask: *extraTask, success: true, itemSet: &output.FullSet, fullItem: exampleItem, factor: util.Optional_OfValue(factor), setBonus: setBonus}
	} else {
		printer.Println("UPGRADE SET NOT FOUND")
		return upgradeItemResult_OfFailure(extraTask, exampleItem)
	}

	printer.Println0()
	parentPrinter.AppendOther(printer)

	return result
}

func removePairedSimilar(jobItems *items.FullOptionsMap, testSlot items.SlotEquip, testItem *items.FullItem, substituteEmptySlotOnly map[items.SlotItem]items.ItemId, model *gear_model.SpecModel, printer *util.PrintRecorder) {
	pairedSlot := testSlot.PairedSlot()
	if pairedSlot != -1 {
		printer.Println("removePairedSimilar")
		for _, z := range jobItems[pairedSlot] {
			printer.Println("---" + z.CreateString())
		}
		jobItems.FilterSlotNoValidate(pairedSlot, func(x *items.FullItem) bool { return !items.UniqueEquipViolation(x, testItem) })

		if len(jobItems[pairedSlot]) == 0 {
			substituteId, hasSub := substituteEmptySlotOnly[testItem.SlotItem()]
			if hasSub {
				subOpts, _ := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(substituteId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, model, printer)
				jobItems[pairedSlot] = subOpts
			} else {
				panic("remove paired " + testItem.BaseName() + " left empty slot")
			}
		}
	}
}
