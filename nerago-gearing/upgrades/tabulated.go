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
		result, _ := findUpgrade(&input.FindUpgrades_BasicInputs, baseItems, extraItems, model, printer, tracker, goal)
		return result
	}

	outputMap, _ := findUpgrades_AllRaid(input.Specs, find)

	printer := util.PrintRecorder_CreateLogFile()
	reportTabulatedResults(outputMap, input.Specs, printer, input.PositiveResultsOnly)
	printer.Close()
}

func FindUpgrades_Sim_AllRaid_Run(input *FindUpgrades_MultiSpec_Sim) {
	find := func(baseItems *items.FullOptionsMap, extraItems []*items.FullItem, model *model.Model, printer *util.PrintRecorder, tracker *util.TrackProgress, goal UpgradeGoal) []upgradeItemResultWithSim {
		return findUpgradeAndSim(&input.FindUpgrades_SimInputs, baseItems, extraItems, model, printer, tracker, goal)
	}

	outputMap, _ := findUpgrades_AllRaid(input.Specs, find)

	printer := util.PrintRecorder_CreateLogFile()
	reportTabulatedSimResults(outputMap, input.Specs, printer, input.PositiveResultsOnly)
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

	options := setup.OptionsSetup_FromGearFile(spec.GearFile, &spec.Model, setup.MissingEnchant_Panic, printer)
	upgradeItems := spec.ItemFinder(difficulty)

	group := reportGroup{spec.Label, difficulty}
	*groups = append(*groups, group)

	outputMap[group] = find(&options, upgradeItems, &spec.Model, printer, tracker.MakeNested(), spec.Goal)

	printer.Close()
}

func groupByBossAndItem[T reportable, R reportForItemAddable[T]](outputMap map[reportGroup][]T, makeGroup func(*items.FullItem, items.SlotEquip) R) map[string]map[reportItemRef]R {
	byBossThenItem := make(map[string]map[reportItemRef]R)
	for mode, resultList := range outputMap {
		for _, result := range resultList {
			itemMap := byBossThenItem[result.Boss()]
			if itemMap == nil {
				itemMap = make(map[reportItemRef]R)
				byBossThenItem[result.Boss()] = itemMap
			}

			ref := reportItemRef{result.Item().ItemId(), result.Slot()}
			report, exists := itemMap[ref]
			if !exists {
				report = makeGroup(result.Item(), result.Slot())
				itemMap[ref] = report
			}

			report.Add(mode, result)
		}
	}
	return byBossThenItem
}

func reportTabulatedResults(outputMap map[reportGroup][]upgradeItemResult, specs []FindUpgrades_Spec, printer *util.PrintRecorder, positiveResultsOnly bool) {
	byBossThenItem := groupByBossAndItem(outputMap, makeReportForItem(len(specs)))

	printer.Println("MULTISPEC RANKING BY BOSS")
	for _, bossName := range db.BossItemData_NamesInOrder {
		itemMap := byBossThenItem[bossName]
		if itemMap != nil {
			printer.Println(bossName)

			reportList := make([]*reportForItemBasic, 0, len(itemMap))
			for _, report := range itemMap {
				if !positiveResultsOnly || report.BestRating() > 0 {
					reportList = append(reportList, report)
				}
			}
			slices.SortFunc(reportList, func(a, b *reportForItemBasic) int { return cmp.Compare(a.BestRating(), b.BestRating()) })

			var tab util.TabulateOutput
			tab.SetColumnSpacing(2)
			tab.AddColumnHeader("slot", true)
			tab.AddColumnHeader("ilvl", false)
			tab.AddColumnHeader("name", false)
			for _, spec := range specs {
				tab.AddColumnHeader(spec.Label, true)
			}

			for _, report := range reportList {
				row := make([]string, 0, tab.ColumnCount())
				row = append(row, report.slot.Name())
				row = append(row, strconv.FormatUint(uint64(report.item.Ref.ItemLevel), 10))
				row = append(row, report.item.BaseName)
				for _, spec := range specs {
					value := report.grouped[spec.Label].increaseStr()
					row = append(row, value)
				}
				tab.AddRow(row)
			}

			tab.Write(printer)
			printer.Println0()
		}
	}
}

func reportTabulatedSimResults(outputMap map[reportGroup][]upgradeItemResultWithSim, specs []FindUpgrades_Spec, printer *util.PrintRecorder, positiveResultsOnly bool) {
	byBossThenItem := groupByBossAndItem(outputMap, makeReportSimForItem(len(specs)))

	printer.Println("MULTISPEC RANKING BY BOSS")
	for _, bossName := range db.BossItemData_NamesInOrder {
		itemMap := byBossThenItem[bossName]
		if itemMap != nil {
			printer.Println(bossName)

			reportList := make([]*reportForItemWithSim, 0, len(itemMap))
			for _, report := range itemMap {
				if !positiveResultsOnly || report.BestRating() > 0 {
					reportList = append(reportList, report)
				}
			}
			slices.SortFunc(reportList, func(a, b *reportForItemWithSim) int { return cmp.Compare(a.BestRating(), b.BestRating()) })

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
				row = append(row, strconv.FormatUint(uint64(report.item.Ref.ItemLevel), 10))
				row = append(row, report.item.BaseName)

				for _, spec := range specs {
					groupContent := report.grouped[spec.Label]
					row = append(row, groupContent.increaseStr())
					row = append(row, groupContent.percentStrSim())
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
	}
}

func bestSimOf(report *reportForItemWithSim) (simulate.SimResultStats, string) {
	var bestIncrease simulate.SimResultStats
	var bestLabel string
	best := -100.0
	for label, result := range report.grouped {
		value := result.percentSim()
		if value > best {
			best = value
			bestLabel = label
			bestIncrease = result.increaseSimBreakdown()
		}
	}
	return bestIncrease, bestLabel
}
