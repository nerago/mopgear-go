package upgrades

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"slices"
)

func findUpgrade(baseItems *items.FullOptionsMap, extraItems []*items.FullItem, model *model.Model, printer *util.PrintRecorder, tracker *util.TrackProgress, mode upgradeMode) ([]upgradeItemResult, *items.FullItemSet) {
	extraItems = upgradeExtraItems(extraItems, printer)
	checkDuplicates(extraItems)
	extraTasks := makeExtraTasks(extraItems, baseItems, printer, mode)

	tracker.RunOuterTracking(len(extraTasks) + 1)
	defer tracker.Stop()

	printer.Println("FINDING BASELINE")
	baseRating, baseSet := findBase(baseItems, model, printer, tracker)

	printer.Println("TRYING ITEMS")
	resultList := channel_op.IterateEach_SliceToSlice(upgradeEachThreads, extraTasks,
		func(task *upgradeItemTask, resultChannel chan<- upgradeItemResult) {
			resultChannel <- performUpgradeTask(task, baseItems, baseRating, model, printer, tracker)
		})
	reportBasicResults(resultList, printer)
	return resultList, baseSet
}

func upgradeExtraItems(extraItems []*items.FullItem, printer *util.PrintRecorder) []*items.FullItem {
	result := make([]*items.FullItem, 0, len(extraItems))
	for _, item := range extraItems {
		replace := db.WowSimDB_ByIdAndUpgrade_AllowFallback(item.ItemId(), targetUpgradeLevel, printer)
		result = append(result, replace)
	}
	return result
}

func checkDuplicates(extraItems []*items.FullItem) {
	byName := make(map[string]*items.FullItem)
	for _, item := range extraItems {
		_, alreadySeen := byName[item.BaseName]
		if alreadySeen {
			panic("duplicate item for " + item.BaseName)
		} else {
			byName[item.BaseName] = item
		}
	}
}

func makeExtraTasks(extraItems []*items.FullItem, baseItems *items.FullOptionsMap, printer *util.PrintRecorder, mode upgradeMode) []upgradeItemTask {
	taskList := make([]upgradeItemTask, 0, len(extraItems))
	for _, extra := range extraItems {
		boss := db.BossItemData_BossForItem(extra)
		for _, slot := range extra.Slot.ToSlotEquipOptions() {
			if canPerformSpecifiedUpgrade(extra, slot, baseItems, printer) {
				taskList = append(taskList, upgradeItemTask{extra, slot, mode, boss})
			}
		}
	}
	return taskList
}

func canPerformSpecifiedUpgrade(extra *items.FullItem, slot items.SlotEquip, baseItems *items.FullOptionsMap, printer *util.PrintRecorder) bool {
	if slices.Contains(ignoredItems, extra.ItemId()) {
		return false
	}

	if !baseItems.Has(slot) {
		printer.Println("SLOT NOT USED IN CURRENT SET " + extra.CreateString())
		return false
	}

	if slot == items.Equip_Weapon {
		currentWeapon := baseItems.Get(items.Equip_Weapon)[0]
		if extra.Slot != currentWeapon.Slot {
			printer.Println("WRONG WEAPON TYPE " + extra.CreateString())
			return false
		}
	}

	if baseItems.IncludesItemIdInSlot(extra.ItemId(), slot) {
		printer.Println("SAME ITEM " + extra.CreateString())
		return false
	}

	paired := slot.PairedSlot()
	if paired != -1 && baseItems.IncludesItemIdInSlot(extra.ItemId(), paired) {
		printer.Println("SAME ITEM ID IN OTHER SLOT " + extra.CreateString())
		return false
	} else if paired != -1 && baseItems.IncludesItemNameInSlot(extra.BaseName, paired) {
		printer.Println("SAME ITEM NAME IN OTHER SLOT (unique equipped) " + extra.CreateString())
		return false
	}

	return true
}

func findBase(baseItems *items.FullOptionsMap, model *model.Model, printer *util.PrintRecorder, tracker *util.TrackProgress) (float64, *items.FullItemSet) {
	output := solver.Solver(solver.SolveInput{
		ItemOptions:        baseItems,
		Model:              model,
		PhasedAcceptable:   false,
		OuterTrackProgress: tracker,
		Printer:            printer,
		SolveSize:          baseSolveSize})

	if !output.Success {
		panic("couldn't find valid baseline set")
	}

	printer.Printf("\n%s\nBASE RATING    = %d\n\n", output.SolvedSet.TotalRated().CreateString(), output.ResultRating)
	return float64(output.ResultRating), &output.FullSet
}

func performUpgradeTask(extraTask *upgradeItemTask, baseItems *items.FullOptionsMap, baseRating float64, model *model.Model, parentPrinter *util.PrintRecorder, outerTracker *util.TrackProgress) upgradeItemResult {
	printer := util.PrintRecorder_HoldAll()

	item := extraTask.item // this "item" is from ItemFinder and is just a basic DB object
	slot := extraTask.slot
	printer.Println("OFFER " + item.CreateString())
	printer.Println("REPLACING " + baseItems.Get(slot)[0].CreateString())

	// TODO consider loading from bags etc
	newOptions, _ := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(item.ItemId(), item.Ref.UpgradeLevel(), model, printer)
	jobItems := baseItems.Clone()
	jobItems[slot] = newOptions

	output := solver.Solver(solver.SolveInput{
		ItemOptions:        &jobItems,
		Model:              model,
		PhasedAcceptable:   false,
		OuterTrackProgress: outerTracker,
		Printer:            printer,
		SolveSize:          itemSolveSize})

	var result upgradeItemResult
	if output.Success {
		printer.Printf("SET STATS %s\n", output.SolvedSet.TotalRated().CreateString())
		output.Report(printer) // verbose

		factor := float64(output.ResultRating) / baseRating
		printer.Printf("UPGRADE RATING = %d FACTOR = %1.3f\n", output.ResultRating, factor)

		setBonus := model.SetBonus.CountInAnySet(output.FullSet.Items())

		result = upgradeItemResult{upgradeItemTask: *extraTask, success: true, itemSet: &output.FullSet, factor: factor, setBonus: setBonus}
	} else {
		printer.Println("UPGRADE SET NOT FOUND")
		result = upgradeItemResult{upgradeItemTask: *extraTask, success: false, factor: -1.0}
	}

	printer.Println0()
	parentPrinter.AppendOther(printer)

	return result
}
