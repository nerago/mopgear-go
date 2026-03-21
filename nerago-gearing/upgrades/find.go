package upgrades

import (
	"cmp"
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
	"strconv"
)

type upgradeMode int8

const (
	Upgrade_Miti_Normal upgradeMode = iota
	Upgrade_Miti_Heroic             = iota
	Upgrade_Dps_Normal              = iota
	Upgrade_Dps_Heroic              = iota
)

type upgradeItemTask struct {
	item *items.FullItem
	slot items.SlotEquip
	mode upgradeMode
}

type upgradeItemResult struct {
	upgradeItemTask
	factor float64
	boss   string
}

func (result upgradeItemResult) percent() float64 {
	return (result.factor - 1.0) * 100
}

func (result upgradeItemResult) percentStr() string {
	if result.factor == 0 {
		return ""
	}
	percent := result.percent()
	str := strconv.FormatFloat(percent, 'f', 2, 64)
	if percent < 0 {
		return str + "%"
	} else {
		return "+" + str + "%"
	}
}

const (
	upgradeEachThreads = 4
	targetUpgradeLevel = 2
	baseSolveSize      = solver.SolveSize_Long
	itemSolveSize      = solver.SolveSize_Medium
	// baseSolveSize = solver.SolveSize_Medium
	// itemSolveSize = solver.SolveSize_PerItem
)

var ignoredItems = []items.ItemId{
	63207, // org port cloak
	84661, // fishing pole
	90042} // straw hat

func FindUpgrades_AllRaid_Run() {
	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(4)
	defer tracker.Stop()

	modelMitigation := model.Model_PallyProtMitigation()
	modelDps := model.Model_PallyProtMitigation()

	upgradeNormal := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	upgradeHeroic := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Heroic)

	outputMap := make(map[upgradeMode][]upgradeItemResult)

	printer := util.PrintRecorder_CreateLogFile()
	printer.Println("[[[[[[[[[[[[[[[[[[[[ PALLY PROT DPS normal UPGRADES ]]]]]]]]]]]]]]]]]]]]")
	optionsDps := setup.OptionsSetup_FromGearFile(files.GearFileProtDps, &modelDps, printer)
	outputMap[Upgrade_Dps_Normal] = findUpgrade(&optionsDps, upgradeNormal, &modelDps, printer, tracker.MakeNested(), Upgrade_Dps_Normal)
	printer.Close()

	printer = util.PrintRecorder_CreateLogFile()
	printer.Println("[[[[[[[[[[[[[[[[[[[[ PALLY PROT DPS heroic UPGRADES ]]]]]]]]]]]]]]]]]]]]")
	outputMap[Upgrade_Dps_Heroic] = findUpgrade(&optionsDps, upgradeHeroic, &modelDps, printer, tracker.MakeNested(), Upgrade_Dps_Heroic)
	printer.Close()

	printer = util.PrintRecorder_CreateLogFile()
	printer.Println("[[[[[[[[[[[[[[[[[[[[ PALLY PROT MITIGATION normal UPGRADES ]]]]]]]]]]]]]]]]]]]]")
	optionsMitigation := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigation, &modelMitigation, printer)
	outputMap[Upgrade_Miti_Normal] = findUpgrade(&optionsMitigation, upgradeNormal, &modelDps, printer, tracker.MakeNested(), Upgrade_Miti_Normal)
	printer.Close()

	printer = util.PrintRecorder_CreateLogFile()
	printer.Println("[[[[[[[[[[[[[[[[[[[[ PALLY PROT MITIGATION heroic UPGRADES ]]]]]]]]]]]]]]]]]]]]")
	outputMap[Upgrade_Miti_Heroic] = findUpgrade(&optionsMitigation, upgradeHeroic, &modelDps, printer, tracker.MakeNested(), Upgrade_Miti_Heroic)
	printer.Close()

	printer = util.PrintRecorder_CreateLogFile()
	reportTabulatedResults(outputMap, printer)
	printer.Close()

}

func findExtraResults(extraTask *upgradeItemTask, baseItems *items.FullOptionsMap, baseRating float64, model *model.Model, parentPrinter *util.PrintRecorder, outerTracker *util.TrackProgress) upgradeItemResult {
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

	var factor float64
	if output.Success {
		printer.Printf("SET STATS %s\n", output.SolvedSet.TotalRated().CreateString())

		factor = float64(output.ResultRating) / baseRating
		printer.Printf("UPGRADE RATING = %d FACTOR = %1.3f\n", output.ResultRating, factor)
	} else {
		factor = 0
		printer.Println("UPGRADE SET NOT FOUND")
	}

	printer.Println0()
	parentPrinter.AppendOther(printer)

	boss := db.BossItemData_BossForItem(extraTask.item)
	return upgradeItemResult{*extraTask, factor, boss}
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


type reportForItem struct {
	item   *items.FullItem
	byMode map[upgradeMode]upgradeItemResult
}

func (report *reportForItem) BestRating() float64 {
	var best float64
	for _, item := range report.byMode {
		if item.factor > best {
			best = item.factor
		}
	}
	return best
}

func reportTabulatedResults(outputMap map[upgradeMode][]upgradeItemResult, printer *util.PrintRecorder) {
	byBossThenItem := make(map[string]map[items.ItemId]*reportForItem)
	for mode, resultList := range outputMap {
		for _, result := range resultList {
			itemMap := byBossThenItem[result.boss]
			if itemMap == nil {
				itemMap = make(map[items.ItemId]*reportForItem)
				byBossThenItem[result.boss] = itemMap
			}

			itemId := result.item.ItemId()
			report := itemMap[itemId]
			if report == nil {
				report = new(reportForItem)
				report.item = result.item
				report.byMode = make(map[upgradeMode]upgradeItemResult)
				itemMap[itemId] = report
			}

			report.byMode[mode] = result
		}
	}

	printer.Println("MULTISPEC RANKING BY BOSS")
	for _, bossName := range db.BossItemData_NamesInOrder {
		itemMap := byBossThenItem[bossName]
		if itemMap != nil {
			printer.Println(bossName)

			reportList := make([]reportForItem, 0, len(itemMap))
			for _, report := range itemMap {
				reportList = append(reportList, *report)
			}
			slices.SortFunc(reportList, func(a, b reportForItem) int { return cmp.Compare(a.BestRating(), b.BestRating()) })

			printer.Printf("%10s%5s%45s%10s%10s%10s%10s\n", "slot", "ilvl", "name", "DPS_norm", "MIT_norm", "DPS_hero", "MIT_hero")

			for _, report := range reportList {
				printer.Printf("%10s%5d%45s%10s%10s%10s%10s\n", report.item.Slot.Name(), report.item.Ref.ItemLevel, report.item.BaseName,
					report.byMode[Upgrade_Dps_Normal].percentStr(), report.byMode[Upgrade_Miti_Normal].percentStr(),
					report.byMode[Upgrade_Dps_Heroic].percentStr(), report.byMode[Upgrade_Miti_Heroic].percentStr())
			}

			printer.Println0()
		}
	}
}
