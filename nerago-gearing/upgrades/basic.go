package upgrades

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"slices"
)

func findUpgrade(input *FindUpgrades_BasicInputs, baseItems *items.FullOptionsMap, extraItems []*items.FullItem, model *model.Model, printer *util.PrintRecorder, tracker *util.TrackProgress,
	goal stats.OptimiseGoal, upgradeLevel items.UpgradeLevel, forceIncludeMost bool, substituteItems []items.ItemId, substituteEmptySlotOnly map[items.SlotItem]items.ItemId) ([]upgradeItemResult, *items.FullItemSet) {

	extraItems = setupUpgradeLevel(extraItems, upgradeLevel, printer)
	checkDuplicates(extraItems)
	extraTasks := makeExtraTasks(input, extraItems, baseItems, printer, goal)
	addSubstituteItems(baseItems, substituteItems, model, printer)

	tracker.RunOuterTracking(len(extraTasks) + 1)
	defer tracker.SetDone()

	printer.Println("FINDING BASELINE")
	baseRating, baseSet := findBase(baseItems, model, printer, tracker)
	tools.ReportSetFewerParams(model, baseSet, printer)

	printer.Println("TRYING ITEMS")
	resultList := channel_op.Map_SliceToSlice(c_upgradeEachThreads, extraTasks,
		func(task *upgradeItemTask) upgradeItemResult {
			return performUpgradeTask(task, baseItems, baseRating, model, printer, tracker, forceIncludeMost, substituteEmptySlotOnly)
		})
	return resultList, baseSet
}

func setupUpgradeLevel(extraItems []*items.FullItem, upgradeLevel items.UpgradeLevel, printer *util.PrintRecorder) []*items.FullItem {
	result := make([]*items.FullItem, 0, len(extraItems))
	for _, item := range extraItems {
		replace := db.WowSimDB_ByIdAndUpgrade_AllowFallback(item.ItemId(), int32(upgradeLevel), printer)
		replace = replace.MakeItemWithRandomSuffix(item.RandomSuffix())
		result = append(result, replace)
	}
	return result
}

func checkDuplicates(extraItems []*items.FullItem) {
	byName := make(map[string]*items.FullItem)
	for _, item := range extraItems {
		_, alreadySeen := byName[item.BaseName()]
		if alreadySeen {
			panic("duplicate item for " + item.BaseName())
		} else {
			byName[item.BaseName()] = item
		}
	}
}

func makeExtraTasks(input *FindUpgrades_BasicInputs, extraItems []*items.FullItem, baseItems *items.FullOptionsMap, printer *util.PrintRecorder, goal stats.OptimiseGoal) []upgradeItemTask {
	bagsFile := loaders.BagsFile_PlusPaladinGear_Read()

	taskList := make([]upgradeItemTask, 0, len(extraItems))
	for _, extra := range extraItems {
		boss := db.BossItemData_BossForItem(extra)
		for _, slot := range extra.SlotItem().ToSlotEquipOptions() {
			canUpgrade := canPerformSpecifiedUpgrade(input, extra, slot, baseItems, bagsFile, printer)
			switch canUpgrade {
			case items.CanUpgrade_Yes, items.CanUpgrade_Equipped, items.CanUpgrade_Equipped_Similar, items.CanUpgrade_AvailableInBags:
				taskList = append(taskList, upgradeItemTask{item: extra, slot: slot, goal: goal, boss: boss, canUpgrade: canUpgrade})
			}
		}
	}
	return taskList
}

func addSubstituteItems(optionsMap *items.FullOptionsMap, substituteItems []items.ItemId, model *model.Model, printer *util.PrintRecorder) {
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

func findBase(baseItems *items.FullOptionsMap, model *model.Model, printer *util.PrintRecorder, tracker *util.TrackProgress) (float64, *items.FullItemSet) {
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

func performUpgradeTask(extraTask *upgradeItemTask, baseItems *items.FullOptionsMap, baseRating float64, model *model.Model, parentPrinter *util.PrintRecorder, outerTracker *util.TrackProgress, forceIncludeMost bool, substituteEmptySlotOnly map[items.SlotItem]items.ItemId) upgradeItemResult {
	if !extraTask.actuallyAttemptUpgrade(forceIncludeMost) {
		parentPrinter.Println("SKIPPING " + extraTask.item.BaseName())
		return upgradeItemResult_OfFailure(extraTask)
	}

	printer := util.PrintRecorder_HoldAll()

	incompleteItem := extraTask.item // this "item" is from ItemFinder and is just a basic DB object
	itemId := incompleteItem.ItemId()
	upgradeLevel := incompleteItem.UpgradeLevel()
	randomSuffix := incompleteItem.RandomSuffix()
	slot := extraTask.slot

	newOptions, _ := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, upgradeLevel, randomSuffix, model, printer)
	jobItems := baseItems.Clone()
	jobItems[slot] = newOptions

	printer.Println("OFFER " + newOptions[0].CreateString())
	printer.Println("REPLACING " + baseItems.Get(slot)[0].CreateString())

	if extraTask.canUpgrade == items.CanUpgrade_Equipped || extraTask.canUpgrade == items.CanUpgrade_Equipped_Similar {
		removePairedSimilar(&jobItems, slot, incompleteItem, substituteEmptySlotOnly, model, printer)
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

		result = upgradeItemResult{upgradeItemTask: *extraTask, success: true, itemSet: &output.FullSet, factor: util.Optional_OfValue(factor), setBonus: setBonus}
	} else {
		printer.Println("UPGRADE SET NOT FOUND")
		return upgradeItemResult_OfFailure(extraTask)
	}

	printer.Println0()
	parentPrinter.AppendOther(printer)

	return result
}

func removePairedSimilar(jobItems *items.FullOptionsMap, testSlot items.SlotEquip, testItem *items.FullItem, substituteEmptySlotOnly map[items.SlotItem]items.ItemId, model *model.Model, printer *util.PrintRecorder) {
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
