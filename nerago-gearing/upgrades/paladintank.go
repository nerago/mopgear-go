package upgrades

import (
	"cmp"
	"math"
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
)

const (
	upgradeEachThreads = 4
	targetUpgradeLevel = 2
	baseSolveSize      = solver.SolveSize_Long
	itemSolveSize      = solver.SolveSize_Medium
	// baseSolveSize = solver.SolveSize_Medium
	// itemSolveSize = solver.SolveSize_PerItem
)

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
	outputMap[Upgrade_Dps_Normal], _ = findUpgrade(&optionsDps, upgradeNormal, &modelDps, printer, tracker.MakeNested(), Upgrade_Dps_Normal)
	printer.Close()

	printer = util.PrintRecorder_CreateLogFile()
	printer.Println("[[[[[[[[[[[[[[[[[[[[ PALLY PROT DPS heroic UPGRADES ]]]]]]]]]]]]]]]]]]]]")
	outputMap[Upgrade_Dps_Heroic], _ = findUpgrade(&optionsDps, upgradeHeroic, &modelDps, printer, tracker.MakeNested(), Upgrade_Dps_Heroic)
	printer.Close()

	printer = util.PrintRecorder_CreateLogFile()
	printer.Println("[[[[[[[[[[[[[[[[[[[[ PALLY PROT MITIGATION normal UPGRADES ]]]]]]]]]]]]]]]]]]]]")
	optionsMitigation := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigation, &modelMitigation, printer)
	outputMap[Upgrade_Miti_Normal], _ = findUpgrade(&optionsMitigation, upgradeNormal, &modelDps, printer, tracker.MakeNested(), Upgrade_Miti_Normal)
	printer.Close()

	printer = util.PrintRecorder_CreateLogFile()
	printer.Println("[[[[[[[[[[[[[[[[[[[[ PALLY PROT MITIGATION heroic UPGRADES ]]]]]]]]]]]]]]]]]]]]")
	outputMap[Upgrade_Miti_Heroic], _ = findUpgrade(&optionsMitigation, upgradeHeroic, &modelDps, printer, tracker.MakeNested(), Upgrade_Miti_Heroic)
	printer.Close()

	printer = util.PrintRecorder_CreateLogFile()
	reportTabulatedResults(outputMap, printer)
	printer.Close()

}

type reportForItem struct {
	item   *items.FullItem
	byMode map[upgradeMode]upgradeItemResult
}

func makeReportForItem(item *items.FullItem) *reportForItem {
	report := new(reportForItem)
	report.item = item
	report.byMode = make(map[upgradeMode]upgradeItemResult)
	return report
}

func (report *reportForItem) Add(mode upgradeMode, result upgradeItemResult) {
	report.byMode[mode] = result
}

func (report *reportForItem) BestRating() float64 {
	var best float64 = -1.0
	for _, item := range report.byMode {
		best = math.Max(best, item.factor)
	}
	return best
}

type reportForItemWithSim struct {
	item   *items.FullItem
	byMode map[upgradeMode]upgradeItemResultWithSim
}

func makeReportSimForItem(item *items.FullItem) *reportForItemWithSim {
	report := new(reportForItemWithSim)
	report.item = item
	report.byMode = make(map[upgradeMode]upgradeItemResultWithSim)
	return report
}

func (report *reportForItemWithSim) Add(mode upgradeMode, result upgradeItemResultWithSim) {
	report.byMode[mode] = result
}

func (report *reportForItemWithSim) BestRating() float64 {
	var best float64 = -1.0
	for _, item := range report.byMode {
		best = math.Max(best, math.Max(item.percentSim(), item.increase()))
	}
	return best
}

type reportable interface {
	Boss() string
	Item() *items.FullItem
}

type reportGroup[T reportable] interface {
	Add(mode upgradeMode, result T)
}

func groupByBossAndItem[T reportable, R reportGroup[T]](outputMap map[upgradeMode][]T, makeGroup func(*items.FullItem) R) map[string]map[items.ItemId]R {
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

func reportTabulatedResults(outputMap map[upgradeMode][]upgradeItemResult, printer *util.PrintRecorder) {
	byBossThenItem := groupByBossAndItem(outputMap, makeReportForItem)

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

			printer.Printf("%10s%5s%45s%10s%10s%10s%10s\n", "slot", "ilvl", "name", "DPS_norm", "MIT_norm", "DPS_hero", "MIT_hero")

			for _, report := range reportList {
				printer.Printf("%10s%5d%45s%10s%10s%10s%10s\n", report.item.Slot.Name(), report.item.Ref.ItemLevel, report.item.BaseName,
					report.byMode[Upgrade_Dps_Normal].increaseStr(), report.byMode[Upgrade_Miti_Normal].increaseStr(),
					report.byMode[Upgrade_Dps_Heroic].increaseStr(), report.byMode[Upgrade_Miti_Heroic].increaseStr())
			}

			printer.Println0()
		}
	}
}
