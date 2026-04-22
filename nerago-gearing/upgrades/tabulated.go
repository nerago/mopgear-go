package upgrades

import (
	"cmp"
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
	"strconv"
)

func FindUpgrades_AllRaid_Run(input *FindUpgrades_MultiSpec, printer *util.PrintRecorder) {
	find := func(baseItems *items.FullOptionsMap, extraItems []*items.FullItem, model *model.Model, printer *util.PrintRecorder, tracker *util.TrackProgress, goal UpgradeGoal, substituteEmptySlotOnly map[items.SlotItem]items.ItemId) []upgradeItemResult {
		result, _ := findUpgrade(&input.FindUpgrades_BasicInputs, baseItems, extraItems, model, printer, tracker, goal, false, substituteEmptySlotOnly)
		return result
	}

	outputMap, _ := findUpgrades_AllRaid(&input.FindUpgrades_BasicInputs, input.Specs, find)

	reportTabulatedResults(outputMap, input.Specs, printer, input.PositiveResultsOnly)
}

func FindUpgrades_Sim_AllRaid_Run(input *FindUpgrades_MultiSpec_Sim, printer *util.PrintRecorder) {
	find := func(baseItems *items.FullOptionsMap, extraItems []*items.FullItem, model *model.Model, printer *util.PrintRecorder, tracker *util.TrackProgress, goal UpgradeGoal, substituteEmptySlotOnly map[items.SlotItem]items.ItemId) []upgradeItemResultWithSim {
		return findUpgradeAndSim(&input.FindUpgrades_SimInputs, baseItems, extraItems, model, printer, tracker, goal, substituteEmptySlotOnly)
	}

	outputMap, _ := findUpgrades_AllRaid(&input.FindUpgrades_BasicInputs, input.Specs, find)

	reportTabulatedSimResults_All(outputMap, input.Specs, printer, input.PositiveResultsOnly)
	reportTabulatedSimResults_Boss(outputMap, input.Specs, printer, input.PositiveResultsOnly)
}

func findUpgrades_AllRaid[T any](input *FindUpgrades_BasicInputs, specs []FindUpgrades_Spec, find func(baseItems *items.FullOptionsMap, extraItems []*items.FullItem, model *model.Model, printer *util.PrintRecorder, tracker *util.TrackProgress, goal UpgradeGoal, substituteEmptySlotOnly map[items.SlotItem]items.ItemId) []T) (map[reportGroup][]T, []reportGroup) {
	outerCount := 0
	if input.IncludeNormal {
		outerCount += len(specs)
	}
	if input.IncludeHeroic {
		outerCount += len(specs)
	}

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(outerCount)
	defer tracker.Stop()

	outputMap := make(map[reportGroup][]T)
	groups := make([]reportGroup, 0, len(specs)*2)

	for _, spec := range specs {
		if input.IncludeNormal {
			processSpec(find, input, &spec, stats.Difficulty_Normal, outputMap, &groups, tracker)
		}
		if input.IncludeHeroic {
			processSpec(find, input, &spec, stats.Difficulty_Heroic, outputMap, &groups, tracker)
		}
	}

	return outputMap, groups
}

func processSpec[T any](find func(*items.FullOptionsMap, []*items.FullItem, *model.Model, *util.PrintRecorder, *util.TrackProgress, UpgradeGoal, map[items.SlotItem]items.ItemId) []T,
	input *FindUpgrades_BasicInputs, spec *FindUpgrades_Spec, difficulty stats.Difficulty, outputMap map[reportGroup][]T, groups *[]reportGroup, tracker *util.TrackProgress) {
	printer := util.PrintRecorder_CreateLogFile(files.LogOutputPath)
	printer.Println("[[[[[[[[[[[[[[[[[[[[ " + spec.Label + " " + difficulty.Name() + " UPGRADES ]]]]]]]]]]]]]]]]]]]]")

	options := setup.OptionsSetup_FromGearFile(spec.GearFile, &spec.Model, setup.MissingEnchant_Panic, printer)
	addSubstituteItems(&options, spec.SubstituteItems, &spec.Model, printer)

	upgradeItems := spec.ItemFinder(difficulty)
	if !input.IncludeRaden {
		upgradeItems = loaders.ItemFinder_FilterOutRadenItems(upgradeItems)
	}

	group := reportGroup{spec.Label, difficulty}
	*groups = append(*groups, group)

	outputMap[group] = find(&options, upgradeItems, &spec.Model, printer, tracker.MakeNested(), spec.Goal, spec.SubstituteEmptySlotOnly)

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

func groupByItem[T reportable, R reportForItemAddable[T]](outputMap map[reportGroup][]T, makeGroup func(*items.FullItem, items.SlotEquip) R) map[reportItemRef]R {
	byItem := make(map[reportItemRef]R)
	for mode, resultList := range outputMap {
		for _, result := range resultList {
			ref := reportItemRef{result.Item().ItemId(), result.Slot()}
			report, exists := byItem[ref]
			if !exists {
				report = makeGroup(result.Item(), result.Slot())
				byItem[ref] = report
			}

			report.Add(mode, result)
		}
	}
	return byItem
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

func reportTabulatedSimResults_All(outputMap map[reportGroup][]upgradeItemResultWithSim, specs []FindUpgrades_Spec, printer *util.PrintRecorder, positiveResultsOnly bool) {
	itemMap := groupByItem(outputMap, makeReportSimForItem(len(specs)))
	printer.Println("MULTISPEC RANKING ALL")
	reportTabledSimResultItemMap(itemMap, positiveResultsOnly, specs, printer)
	printer.Println("MULTISPEC RANKING ALL - sim percents only")
	reportTabledSimResultItemMap_NoWeight(itemMap, positiveResultsOnly, specs, printer)
}

func reportTabulatedSimResults_Boss(outputMap map[reportGroup][]upgradeItemResultWithSim, specs []FindUpgrades_Spec, printer *util.PrintRecorder, positiveResultsOnly bool) {
	byBossThenItem := groupByBossAndItem(outputMap, makeReportSimForItem(len(specs)))

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
		row = append(row, strconv.FormatUint(uint64(report.item.Ref.ItemLevel), 10))
		row = append(row, report.item.BaseName)
		row = append(row, db.BossItemData_BossForItem(report.item))

		for _, spec := range specs {
			groupContent := report.grouped[spec.Label]
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
