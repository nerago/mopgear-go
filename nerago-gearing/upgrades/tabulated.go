package upgrades

import (
	"cmp"
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"slices"
	"strconv"
)

// possible entry point
func FindUpgrades_Sim_AllRaid_Run(input *FindUpgrades_MultiSpec_Sim, cancel util_async.CancelSignal) {
	outputMap := findUpgrades_AllRaid(&input.FindUpgrades_SimInputs, input.Specs, cancel)

	printer := util.PrintRecorder_CreateLogFileNamed(files.LogOutputPath, "upgrade-report")
	reportTabulatedSimResults_All(outputMap, input.Specs, printer, input.PositiveResultsOnly)
	reportTabulatedSimResults_Boss(outputMap, input.Specs, printer, input.PositiveResultsOnly)
}

func findUpgrades_AllRaid(input *FindUpgrades_SimInputs, specs []FindUpgrades_Spec, cancel util_async.CancelSignal) []upgradeGroupResult {
	groupTasks := make([]upgradeGroupTask, 0)
	for _, spec := range specs {
		if input.IncludeCelestial {
			groupTasks = append(groupTasks, upgradeGroupTask{&spec, stats.Difficulty_Celestial})
		}
		if input.IncludeNormal {
			groupTasks = append(groupTasks, upgradeGroupTask{&spec, stats.Difficulty_Normal})
		}
		if input.IncludeHeroic {
			groupTasks = append(groupTasks, upgradeGroupTask{&spec, stats.Difficulty_Heroic})
		}
	}

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(groupTasks))
	defer tracker.SetDone()

	return util_async.Map_SliceToSlice_Cancellable(4, groupTasks, cancel, func(task *upgradeGroupTask) upgradeGroupResult {
		resultList := processSpecUpgradeGroupTask(input, task.spec, task.difficulty, tracker.NewChild())
		return upgradeGroupResult{
			task: *task,
			group: reportGroup{
				specLabel:  task.spec.Label,
				difficulty: task.difficulty,
			},
			resultList: resultList,
		}
	})
}

func processSpecUpgradeGroupTask(input *FindUpgrades_SimInputs, spec *FindUpgrades_Spec, difficulty stats.Difficulty, tracker *util.TrackProgress) []upgradeItemResultWithSim {
	printer := util.PrintRecorder_CreateLogFileNamed(files.LogOutputPath, "upgrade-"+spec.Label+"-"+difficulty.Name())
	printer.Println("[[[[[[[[[[[[[[[[[[[[ " + spec.Label + " " + difficulty.Name() + " UPGRADES ]]]]]]]]]]]]]]]]]]]]")
	defer printer.Close()

	options := setup.OptionsSetup_FromGearFile(spec.GearFile, &spec.Model, setup.MissingEnchant_Panic, printer)

	upgradeItems := spec.ItemFinder(difficulty)
	upgradeItems = util.RemoveDuplicatesComparable(upgradeItems)

	return findUpgradeAndSim(input, &options, upgradeItems, &spec.Model, printer, tracker, spec.Model.Goal, spec.SubstituteItems, spec.SubstituteEmptySlotOnly)
}

func groupByBossAndItem(outputMap []upgradeGroupResult) map[string]map[reportItemRef]*reportForItemWithSim {
	byBossThenItem := make(map[string]map[reportItemRef]*reportForItemWithSim)
	for _, groupResultEntry := range outputMap {
		group := groupResultEntry.group
		for _, result := range groupResultEntry.resultList {
			itemMap := byBossThenItem[result.boss]
			if itemMap == nil {
				itemMap = make(map[reportItemRef]*reportForItemWithSim)
				byBossThenItem[result.boss] = itemMap
			}

			ref := reportItemRef{result.item.ItemId, result.slot}
			report, exists := itemMap[ref]
			if !exists {
				report = &reportForItemWithSim{
					result.ItemName(),
					result.ItemLevel(),
					result.boss,
					result.slot,
					make(map[string]upgradeItemResultWithSim),
				}
				itemMap[ref] = report
			}

			report.Add(group, result)
		}
	}
	return byBossThenItem
}

func groupByItem(outputMap []upgradeGroupResult) map[reportItemRef]*reportForItemWithSim {
	byItem := make(map[reportItemRef]*reportForItemWithSim)
	for _, groupResultEntry := range outputMap {
		group := groupResultEntry.group
		for _, result := range groupResultEntry.resultList {
			ref := reportItemRef{result.item.ItemId, result.slot}
			report, exists := byItem[ref]
			if !exists {
				report = &reportForItemWithSim{
					result.ItemName(),
					result.ItemLevel(),
					result.boss,
					result.slot,
					make(map[string]upgradeItemResultWithSim),
				}
				byItem[ref] = report
			}

			report.Add(group, result)
		}
	}
	return byItem
}

func reportTabulatedSimResults_All(outputMap []upgradeGroupResult, specs []FindUpgrades_Spec, printer *util.PrintRecorder, positiveResultsOnly bool) {
	itemMap := groupByItem(outputMap)
	printer.Println("MULTISPEC RANKING ALL")
	reportTabledSimResultItemMap(itemMap, positiveResultsOnly, specs, printer)
	printer.Println("MULTISPEC RANKING ALL - sim percents only")
	reportTabledSimResultItemMap_NoWeight(itemMap, positiveResultsOnly, specs, printer)
}

func reportTabulatedSimResults_Boss(outputMap []upgradeGroupResult, specs []FindUpgrades_Spec, printer *util.PrintRecorder, positiveResultsOnly bool) {
	byBossThenItem := groupByBossAndItem(outputMap)

	printer.Println("MULTISPEC RANKING BY BOSS")
	for _, bossName := range db.BossItemData_NamesInOrder {
		itemMap := byBossThenItem[bossName]
		if itemMap != nil {
			printer.Println(bossName)
			reportTabledSimResultItemMap(itemMap, positiveResultsOnly, specs, printer)
		}
	}
}

func reportTabledSimResultItemMap(itemMap map[reportItemRef]*reportForItemWithSim, positiveResultsOnly bool, specs []FindUpgrades_Spec, printer *util.PrintRecorder) {
	reportList := make([]*reportForItemWithSim, 0, len(itemMap))
	for _, report := range itemMap {
		if !positiveResultsOnly || report.BestRating() > 0 {
			reportList = append(reportList, report)
		}
	}
	slices.SortFunc(reportList, func(a, b *reportForItemWithSim) int {
		return cmp.Compare(a.BestRating(), b.BestRating())
	})

	var tab util.TabulateOutput
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("slot", true)
	tab.AddColumnHeader("ilvl", false)
	tab.AddColumnHeader("name", false)
	for _, spec := range specs {
		tab.AddColumnHeader(spec.Label, true)
		tab.AddColumnHeader("sim", true)
	}
	tab.AddColumnHeader("best_sim", true)
	tab.AddColumnHeader("sim_detailed", false)

	for _, report := range reportList {
		row := make([]string, 0, tab.ColumnCount())
		row = append(row, report.slot.Name())
		row = append(row, strconv.FormatUint(uint64(report.itemLevel), 10))
		row = append(row, report.itemName)

		for _, spec := range specs {
			groupContent := report.grouped[spec.Label]
			row = append(row, groupContent.increaseWeightsStr(true))
			row = append(row, groupContent.increaseSimStr(false))
		}

		bestSimIncrease, bestGroup := bestSimOf(report)
		if !bestSimIncrease.IsEmpty() {
			row = append(row, bestGroup)
			row = append(row, bestSimIncrease.CompactStringSignedPercent())
		}

		tab.AddRow(row)
	}

	tab.Write(printer)
	printer.Println0()
}

func reportTabledSimResultItemMap_NoWeight(itemMap map[reportItemRef]*reportForItemWithSim, positiveResultsOnly bool, specs []FindUpgrades_Spec, printer *util.PrintRecorder) {
	reportList := make([]*reportForItemWithSim, 0, len(itemMap))
	for _, report := range itemMap {
		if !positiveResultsOnly || report.BestRating() > 0 {
			reportList = append(reportList, report)
		}
	}
	slices.SortFunc(reportList, func(a, b *reportForItemWithSim) int {
		return cmp.Compare(a.BestRating_NoWeight(), b.BestRating_NoWeight())
	})

	var tab util.TabulateOutput
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("slot", true)
	tab.AddColumnHeader("ilvl", false)
	tab.AddColumnHeader("name", false)
	tab.AddColumnHeader("boss", false)
	for _, spec := range specs {
		tab.AddColumnHeader(spec.Label, true)
	}
	tab.AddColumnHeader("best_sim", true)
	tab.AddColumnHeader("sim_detailed", false)

	for _, report := range reportList {
		row := make([]string, 0, tab.ColumnCount())
		row = append(row, report.slot.Name())
		row = append(row, strconv.FormatUint(uint64(report.itemLevel), 10))
		row = append(row, report.itemName)
		row = append(row, report.boss)

		for _, spec := range specs {
			groupContent := report.grouped[spec.Label]
			row = append(row, groupContent.increaseSimStr(true))
		}

		bestSimIncrease, bestGroup := bestSimOf(report)
		if !bestSimIncrease.IsEmpty() {
			row = append(row, bestGroup)
			row = append(row, bestSimIncrease.CompactStringSignedPercent())
		}

		tab.AddRow(row)
	}

	tab.Write(printer)
	printer.Println0()
}

func bestSimOf(report *reportForItemWithSim) (stats.SimData, string) {
	var bestIncrease stats.SimData
	var bestLabel string
	best := c_nullIncrease
	for label, result := range report.grouped {
		value := result.increaseSim()
		if value > best {
			best = value
			bestLabel = label
			bestIncrease = *result.increaseSimBreakdown()
		}
	}
	return bestIncrease, bestLabel
}
