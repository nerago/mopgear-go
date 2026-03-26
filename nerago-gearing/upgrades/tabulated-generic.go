package upgrades

import (
	"cmp"
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
	"strconv"
)

func FindUpgrades_AllRaid_Run(input *FindUpgrades_MultiSpec) {
	find := func(baseItems *items.FullOptionsMap, extraItems []*items.FullItem, model *model.Model, printer *util.PrintRecorder, tracker *util.TrackProgress, goal UpgradeGoal) []upgradeItemResult {
		result, _ := findUpgrade(&input.Basic, baseItems, extraItems, model, printer, tracker, goal)
		return result
	}

	outputMap, groups := findUpgrades_AllRaid(input.Specs, find)

	printer := util.PrintRecorder_CreateLogFile()
	reportTabulatedResults(outputMap, groups, printer)
	printer.Close()
}

func FindUpgrades_Sim_AllRaid_Run(input *FindUpgrades_MultiSpec_Sim) {
	find := func(baseItems *items.FullOptionsMap, extraItems []*items.FullItem, model *model.Model, printer *util.PrintRecorder, tracker *util.TrackProgress, goal UpgradeGoal) []upgradeItemResultWithSim {
		return findUpgradeAndSim(&input.Sim, baseItems, extraItems, model, printer, tracker, goal)
	}

	outputMap, groups := findUpgrades_AllRaid(input.Specs, find)

	printer := util.PrintRecorder_CreateLogFile()
	reportTabulatedSimResults(outputMap, groups, printer)
	printer.Close()
}

func findUpgrades_AllRaid[T any](specs []FindUpgrades_Spec, find func(baseItems *items.FullOptionsMap, extraItems []*items.FullItem, model *model.Model, printer *util.PrintRecorder, tracker *util.TrackProgress, goal UpgradeGoal) []T) (map[reportGroup][]T, []reportGroup) {
	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(2 * len(specs))
	defer tracker.Stop()

	outputMap := make(map[reportGroup][]T)
	groups := make([]reportGroup, 0, len(specs)*2)

	for _, spec := range specs {
		processSpec(find, &spec, stats.Difficulty_Normal, outputMap, &groups, tracker)
		processSpec(find, &spec, stats.Difficulty_Heroic, outputMap, &groups, tracker)
	}

	return outputMap, groups
}

func processSpec[T any](find func(*items.FullOptionsMap, []*items.FullItem, *model.Model, *util.PrintRecorder, *util.TrackProgress, UpgradeGoal) []T,
	spec *FindUpgrades_Spec, difficulty stats.Difficulty, outputMap map[reportGroup][]T, groups *[]reportGroup, tracker *util.TrackProgress) {
	printer := util.PrintRecorder_CreateLogFile()
	printer.Println("[[[[[[[[[[[[[[[[[[[[ " + spec.Label + " " + difficulty.Name() + " UPGRADES ]]]]]]]]]]]]]]]]]]]]")

	options := setup.OptionsSetup_FromGearFile(spec.GearFile, &spec.Model, printer)
	upgradeItems := spec.ItemFinder(difficulty)

	group := reportGroup{spec.Label, difficulty}
	*groups = append(*groups, group)

	outputMap[group] = find(&options, upgradeItems, &spec.Model, printer, tracker.MakeNested(), spec.Goal)

	printer.Close()
}

func groupByBossAndItem[T reportable, R reportForItemAddable[T]](outputMap map[reportGroup][]T, makeGroup func(*items.FullItem) R) map[string]map[items.ItemId]R {
	byBossThenItem := make(map[string]map[items.ItemId]R)
	for mode, resultList := range outputMap {
		for _, result := range resultList {
			itemMap := byBossThenItem[result.Boss()]
			if itemMap == nil {
				itemMap = make(map[items.ItemId]R)
				byBossThenItem[result.Boss()] = itemMap
			}

			itemId := result.Item().ItemId()
			report, exists := itemMap[itemId]
			if !exists {
				report = makeGroup(result.Item())
				itemMap[itemId] = report
			}

			report.Add(mode, result)
		}
	}
	return byBossThenItem
}

func reportTabulatedResults(outputMap map[reportGroup][]upgradeItemResult, groups []reportGroup, printer *util.PrintRecorder) {
	byBossThenItem := groupByBossAndItem(outputMap, makeReportForItem(len(groups)))

	printer.Println("MULTISPEC RANKING BY BOSS")
	for _, bossName := range db.BossItemData_NamesInOrder {
		itemMap := byBossThenItem[bossName]
		if itemMap != nil {
			printer.Println(bossName)

			reportList := make([]*reportForItem, 0, len(itemMap))
			for _, report := range itemMap {
				reportList = append(reportList, report)
			}
			slices.SortFunc(reportList, func(a, b *reportForItem) int { return cmp.Compare(a.BestRating(), b.BestRating()) })

			var tab util.TabulateOutput
			tab.SetColumnSpacing(2)
			tab.AddColumnHeader("slot", true)
			tab.AddColumnHeader("ilvl", false)
			tab.AddColumnHeader("name", true)
			for _, group := range groups {
				tab.AddColumnHeader(group.String(), true)
			}

			for _, report := range reportList {
				row := make([]string, 0, 3+len(outputMap))
				row = append(row, report.item.Slot.Name())
				row = append(row, strconv.FormatUint(uint64(report.item.Ref.ItemLevel), 10))
				row = append(row, report.item.BaseName)
				for _, group := range groups {
					value := report.grouped[group].increaseStr()
					row = append(row, value)
				}
				tab.AddRow(row)
			}

			tab.Write(printer)
			printer.Println0()
		}
	}
}

func reportTabulatedSimResults(outputMap map[reportGroup][]upgradeItemResultWithSim, groups []reportGroup, printer *util.PrintRecorder) {
	byBossThenItem := groupByBossAndItem(outputMap, makeReportSimForItem(len(groups)))

	printer.Println("MULTISPEC RANKING BY BOSS")
	for _, bossName := range db.BossItemData_NamesInOrder {
		itemMap := byBossThenItem[bossName]
		if itemMap != nil {
			printer.Println(bossName)

			reportList := make([]*reportForItemWithSim, 0, len(itemMap))
			for _, report := range itemMap {
				reportList = append(reportList, report)
			}
			slices.SortFunc(reportList, func(a, b *reportForItemWithSim) int { return cmp.Compare(a.BestRating(), b.BestRating()) })

			var tab util.TabulateOutput
			tab.SetColumnSpacing(2)
			tab.AddColumnHeader("slot", true)
			tab.AddColumnHeader("ilvl", false)
			tab.AddColumnHeader("name", true)
			for _, group := range groups {
				tab.AddColumnHeader(group.String(), true)
				tab.AddColumnHeader("sim", true)
			}
			tab.AddColumnHeader("best_sim", false)
			tab.AddColumnHeader("sim_detailed", false)

			for _, report := range reportList {
				row := make([]string, 0, len(groups)*2+5)
				row = append(row, report.item.Slot.Name())
				row = append(row, strconv.FormatUint(uint64(report.item.Ref.ItemLevel), 10))
				row = append(row, report.item.BaseName)

				for _, group := range groups {
					groupContent := report.grouped[group]
					row = append(row, groupContent.increaseStr())
					row = append(row, groupContent.percentStrSim())
				}

				bestSimIncrease, bestGroup := bestSimOf(report)
				if !bestSimIncrease.IsEmpty() {
					row = append(row, bestGroup.String())
					row = append(row, bestSimIncrease.CompactStringSignedPercent())
				}

				tab.AddRow(row)
			}

			tab.Write(printer)
			printer.Println0()
		}
	}
}

func bestSimOf(report *reportForItemWithSim) (simulate.SimResultStats, reportGroup) {
	var bestIncrease simulate.SimResultStats
	var bestGroup reportGroup
	best := -100.0
	for group, result := range report.grouped {
		value := result.percentSim()
		if value > best {
			best = value
			bestGroup = group
			bestIncrease = result.increaseSimBreakdown()
		}
	}
	return bestIncrease, bestGroup
}
