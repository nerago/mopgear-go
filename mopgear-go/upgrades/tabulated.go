package upgrades

import (
	"cmp"
	"slices"
	"strconv"

	"github.com/nerago/mopgear-go/db"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
)

// possible entry point
func FindUpgrades_Run(input *FindUpgradesMultiSpec, cancel util_async.CancelSignal) {
	printer := util.PrintRecorder_CreateLogFileNamed(files.LogOutputPath, "upgrade-report")
	defer printer.Close()

	sw := util.StopwatchNoisyStart(printer)
	defer sw.Stop()

	outputMap := findUpgrades_AllRaid(input, cancel)

	reportTabulatedSimResults_All(outputMap, input.Specs, printer, input.Settings.PositiveResultsOnly)
	reportTabulatedSimResults_Boss(outputMap, input.Specs, printer, input.Settings.PositiveResultsOnly)
}

func findUpgrades_AllRaid(input *FindUpgradesMultiSpec, cancel util_async.CancelSignal) []upgradeGroupResult {
	groupTasks := make([]upgradeGroupTask, 0)
	for _, spec := range input.Specs {
		if input.Settings.IncludeCelestial {
			groupTasks = append(groupTasks, upgradeGroupTask{&spec, stats.Difficulty_Celestial})
		}
		if input.Settings.IncludeNormal {
			groupTasks = append(groupTasks, upgradeGroupTask{&spec, stats.Difficulty_Normal})
		}
		if input.Settings.IncludeHeroic {
			groupTasks = append(groupTasks, upgradeGroupTask{&spec, stats.Difficulty_Heroic})
		}
	}

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(groupTasks))
	defer tracker.SetDone()

	return util_async.Map_SliceToSlice_Cancellable(4, groupTasks, cancel, func(task *upgradeGroupTask) upgradeGroupResult {
		resultList := processSpecUpgradeGroupTask(&input.Settings, task.spec, task.difficulty, tracker.NewChild())
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

func processSpecUpgradeGroupTask(input *InputSettings, spec *SpecInput, difficulty stats.Difficulty, tracker *util.TrackProgress) []upgradeItemResult {
	printer := util.PrintRecorder_CreateLogFileNamed(files.LogOutputPath, "upgrade-"+spec.Label+"-"+difficulty.Name())
	printer.Println("[[[[[[[[[[[[[[[[[[[[ " + spec.Label + " " + difficulty.Name() + " UPGRADES ]]]]]]]]]]]]]]]]]]]]")
	defer printer.Close()

	upgradeResults, err := processSpecUpgrade(input, spec, difficulty, tracker, printer)

	if err != nil {
		printer.Println("ERROR: " + err.Error())
	}

	return upgradeResults
}

func groupByBossAndItem(outputMap []upgradeGroupResult) map[string]map[reportItemRef]*reportForItem {
	byBossThenItem := make(map[string]map[reportItemRef]*reportForItem)
	for _, groupResultEntry := range outputMap {
		group := groupResultEntry.group
		for _, result := range groupResultEntry.resultList {
			itemMap := byBossThenItem[result.boss]
			if itemMap == nil {
				itemMap = make(map[reportItemRef]*reportForItem)
				byBossThenItem[result.boss] = itemMap
			}

			ref := reportItemRef{result.itemRef.ItemId, result.slot}
			report, exists := itemMap[ref]
			if !exists {
				report = &reportForItem{
					itemName:    result.ItemName(),
					itemLevel:   result.ItemLevel(),
					boss:        result.boss,
					statSummary: result.ItemStatSummary(),
					slot:        result.slot,
					grouped:     make(map[string]upgradeItemResult),
				}
				itemMap[ref] = report
			}

			report.Add(group, result)
		}
	}
	return byBossThenItem
}

func groupByItem(outputMap []upgradeGroupResult) map[reportItemRef]*reportForItem {
	byItem := make(map[reportItemRef]*reportForItem)
	for _, groupResultEntry := range outputMap {
		group := groupResultEntry.group
		for _, result := range groupResultEntry.resultList {
			ref := reportItemRef{result.itemRef.ItemId, result.slot}
			report, exists := byItem[ref]
			if !exists {
				report = &reportForItem{
					itemName:    result.ItemName(),
					itemLevel:   result.ItemLevel(),
					boss:        result.boss,
					statSummary: result.ItemStatSummary(),
					slot:        result.slot,
					grouped:     make(map[string]upgradeItemResult),
				}
				byItem[ref] = report
			}

			report.Add(group, result)
		}
	}
	return byItem
}

func reportTabulatedSimResults_All(outputMap []upgradeGroupResult, specs []SpecInput, printer *util.PrintRecorder, positiveResultsOnly bool) {
	itemMap := groupByItem(outputMap)
	printer.Println("MULTISPEC RANKING ALL")
	reportTabledSimResultItemMap(itemMap, positiveResultsOnly, specs, printer)
	printer.Println("MULTISPEC RANKING ALL - sim percents only")
	reportTabledSimResultItemMap_NoWeight(itemMap, positiveResultsOnly, specs, printer)
}

func reportTabulatedSimResults_Boss(outputMap []upgradeGroupResult, specs []SpecInput, printer *util.PrintRecorder, positiveResultsOnly bool) {
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

func reportTabledSimResultItemMap(itemMap map[reportItemRef]*reportForItem, positiveResultsOnly bool, specs []SpecInput, printer *util.PrintRecorder) {
	reportList := make([]*reportForItem, 0, len(itemMap))
	for _, report := range itemMap {
		if !positiveResultsOnly || report.BestRating() > 0 {
			reportList = append(reportList, report)
		}
	}
	slices.SortFunc(reportList, func(a, b *reportForItem) int {
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
	tab.AddColumnHeader("secondary", true)

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
		} else {
			row = append(row, "", "")
		}

		row = append(row, report.statSummary)

		tab.AddRow(row)
	}

	tab.Write(printer)
	printer.Println0()
}

func reportTabledSimResultItemMap_NoWeight(itemMap map[reportItemRef]*reportForItem, positiveResultsOnly bool, specs []SpecInput, printer *util.PrintRecorder) {
	reportList := make([]*reportForItem, 0, len(itemMap))
	for _, report := range itemMap {
		if !positiveResultsOnly || report.BestRating() > 0 {
			reportList = append(reportList, report)
		}
	}
	slices.SortFunc(reportList, func(a, b *reportForItem) int {
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
	tab.AddColumnHeader("secondary", true)

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
		} else {
			row = append(row, "", "")
		}

		row = append(row, report.statSummary)

		tab.AddRow(row)
	}

	tab.Write(printer)
	printer.Println0()
}

func bestSimOf(report *reportForItem) (stats.SimData, string) {
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
