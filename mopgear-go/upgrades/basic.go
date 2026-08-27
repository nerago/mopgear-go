package upgrades

import (
	"fmt"
	"slices"

	"github.com/nerago/mopgear-go/db"
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/solver"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func prepareUpgradeTasks(extraItems []loaders.ItemFoundRef, baseItems *items.FullOptionsMap, spec *SpecInput, settings *InputSettings, printer *util.PrintRecorder) []upgradeItemTask {
	extraItems = changeUpgradeLevels(extraItems, settings.TargetUpgradeLevel)
	checkDuplicates(extraItems)
	extraTasks := makeExtraTasks(settings, extraItems, baseItems, printer, &spec.Model)
	addSubstituteItems(baseItems, spec.SubstituteItems, &spec.Model, printer)
	return extraTasks
}

func solveBaseLine(printer *util.PrintRecorder, baseItems *items.FullOptionsMap, spec *SpecInput, input *InputSettings) (float64, *items.FullItemSet, error) {
	printer.Println("FINDING BASELINE")

	output := solver.Solver(
		baseItems, &spec.Model, printer,
		input.WeightType, input.SolverTimeout, nil,
	)

	if !output.Success {
		return 0, nil, fmt.Errorf("couldn't find valid baseline set for %s: %w", spec.Label, output.Error)
	}

	printer.Printf("\nBASE RATING %s   = %.0f\n\n", output.SolvedSet.Total().CreateString(), output.ResultRating)

	tools.ReportSet(&spec.Model, &output.FullSet, printer)
	return output.ResultRating, &output.FullSet, nil
}

func changeUpgradeLevels(extraItems []loaders.ItemFoundRef, upgradeLevel items.UpgradeLevel) []loaders.ItemFoundRef {
	return util_collection.MapSliceAsNew(extraItems, func(ref *loaders.ItemFoundRef) loaders.ItemFoundRef {
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
	//byName := make(map[string]loaders.ItemFoundRef)
	//for _, itemRef := range extraItems {
	//	itemName := db.LookupItemNameByItemId(itemRef.ItemId)
	//	_, alreadySeen := byName[itemName]
	//	if alreadySeen {
	//		panic("duplicate item for " + itemName)
	//	} else {
	//		byName[itemName] = itemRef
	//	}
	//}
}

func makeExtraTasks(input *InputSettings, extraItems []loaders.ItemFoundRef, baseItems *items.FullOptionsMap, printer *util.PrintRecorder, model *gear_model.SpecModel) []upgradeItemTask {
	bagsFile := loaders.BagsFile_PlusPaladinGear_Read()

	taskList := make([]upgradeItemTask, 0, len(extraItems))
	for _, extra := range extraItems {
		boss := db.BossItemData_BossForItemId(extra.ItemId)
		tempItem := db.WowSimDB_LoadItemById(extra.ItemId, 0)
		tempItem.SlotItem().ForEachEquip(func(slot items.SlotEquip) {
			canUpgrade := canPerformSpecifiedUpgrade(input, tempItem, slot, baseItems, bagsFile, printer, model)
			switch canUpgrade {
			case items.CanUpgrade_Yes, items.CanUpgrade_Equipped, items.CanUpgrade_Equipped_Similar, items.CanUpgrade_AvailableInBags:
				taskList = append(taskList, upgradeItemTask{itemRef: extra, slot: slot, goal: model.Goal, boss: boss, canUpgrade: canUpgrade})
			case items.CanUpgrade_InvalidAlways:
				// ignore
			}
		})
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

func canPerformSpecifiedUpgrade(settings *InputSettings, extra *items.FullItem, slot items.SlotEquip, baseItems *items.FullOptionsMap, bagsFile loaders.EquippedArray, printer *util.PrintRecorder, model *gear_model.SpecModel) items.CanUpgradeResult {
	if slices.Contains(settings.IgnoredItems, extra.ItemId()) {
		return items.CanUpgrade_InvalidAlways
	}

	if result := baseItems.CouldAddUpgrade_EquipSlot(slot, extra, printer, model.SpecificIncompatibleList); result != items.CanUpgrade_Yes {
		return result
	}

	if bagsFile.HasAnyWithItemId(extra.ItemId()) {
		printer.Println("ALREADY AVAILABLE IN BAG " + extra.CreateString())
		return items.CanUpgrade_AvailableInBags
	}

	return items.CanUpgrade_Yes
}

func performUpgradeTask(task *upgradeItemTask, baseItems *items.FullOptionsMap, baseRating float64, model *gear_model.SpecModel, parentPrinter *util.PrintRecorder, substituteEmptySlotOnly map[items.SlotItem]items.ItemId, weightType weight_types.WeightType, timeout int) upgradeItemResult {
	if task.canUpgrade == items.CanUpgrade_InvalidAlways {
		itemName := db.LookupItemNameByItemId(task.itemRef.ItemId)
		parentPrinter.Println("SKIPPING " + itemName)
		return upgradeItemResult_OfFailure(task, nil)
	}

	innerPrint := util.PrintRecorder_HoldAll()

	itemRef := task.itemRef
	newOptions, exampleItem := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemRef.ItemId, itemRef.UpgradeLevel, itemRef.RandomSuffix, model, innerPrint)
	jobItems := baseItems.Clone()
	jobItems[task.slot] = newOptions

	innerPrint.Println("OFFER " + exampleItem.CreateString())
	innerPrint.Println("REPLACING " + baseItems.Get(task.slot)[0].CreateString())

	if task.canUpgrade == items.CanUpgrade_Equipped || task.canUpgrade == items.CanUpgrade_Equipped_Similar {
		removePairedSimilar(&jobItems, task.slot, exampleItem, substituteEmptySlotOnly, model, innerPrint)
	}

	output := solver.Solver(
		&jobItems, model, innerPrint,
		weightType, timeout, nil,
	)

	var result upgradeItemResult
	if output.Success {
		innerPrint.Printf("SET STATS %s\n", output.SolvedSet.Total().CreateString())
		output.Report(innerPrint) // verbose

		factor := output.ResultRating / baseRating
		innerPrint.Printf("UPGRADE RATING = %.0f FACTOR = %1.3f\n", output.ResultRating, factor)

		setBonus := model.BonusEnabled.CountInAnySet(output.FullSet.Items())

		result = upgradeItemResult{upgradeItemTask: *task, success: true, itemSet: &output.FullSet, fullItem: exampleItem, factor: util_collection.Optional_OfValue(factor), setBonus: setBonus}
	} else {
		innerPrint.Println("UPGRADE SET NOT FOUND")
		return upgradeItemResult_OfFailure(task, exampleItem)
	}

	innerPrint.Println0()
	parentPrinter.AppendOther(innerPrint)

	return result
}

func removePairedSimilar(jobItems *items.FullOptionsMap, testSlot items.SlotEquip, testItem *items.FullItem, substituteEmptySlotOnly map[items.SlotItem]items.ItemId, model *gear_model.SpecModel, printer *util.PrintRecorder) {
	pairedSlot := testSlot.PairedSlot()
	if pairedSlot != items.SlotEquip_Invalid {
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
