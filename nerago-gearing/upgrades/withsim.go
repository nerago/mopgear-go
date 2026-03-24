package upgrades

import (
	"cmp"
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"slices"
)

const (
	simThreads  = 4
	runSize     = simulate.RunSize_QuickDirty
	runSizeBase = simulate.RunSize_Medium
)

func FindUpgrades_Sim_Run() {
	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(2)
	defer tracker.Stop()

	mode := Upgrade_Miti_Heroic
	//modelMitigation := model.Model_PallyProtMitigation()
	modelDps := model.Model_PallyProtMitigation()

	//upgradeNormal := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	upgradeHeroic := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Heroic)

	printer := util.PrintRecorder_CreateLogFile()
	printer.Println("[[[[[[[[[[[[[[[[[[[[ PALLY PROT DPS heroic UPGRADES SIMULATE ]]]]]]]]]]]]]]]]]]]]")
	optionsDps := setup.OptionsSetup_FromGearFile(files.GearFileProtDps, &modelDps, printer)

	initialResult, baseSet := findUpgrade(&optionsDps, upgradeHeroic, &modelDps, printer, tracker.MakeNested(), mode)

	baseSim := simulate.WowSim_Execute(runSize, modelDps.Spec, baseSet.Items(), nil, nil)
	printer.Println("SIM *BASELINE*")
	baseSim.Print(printer)

	// baseSim := simulate.WowSim_Execute(runSize, modelDps.Spec, baseSet.Items(), nil, tracker.MakeNested())
	simResult := simEachInitialResult(initialResult, &modelDps, &baseSim, tracker.MakeNested(), printer)

	reportBasicResultsSim(simResult, printer)
}

func simEachInitialResult(inputList []upgradeItemResult, model *model.Model, baseSim *simulate.SimResultStats, tracker *util.TrackProgress, printer *util.PrintRecorder) []upgradeItemResultWithSim {
	tracker.RunOuterTracking(len(inputList))
	return channel_op.IterateEach_SliceToSlice(simThreads, inputList, func(input *upgradeItemResult, resultChannel chan<- upgradeItemResultWithSim) {
		if input.success {
			simResult := simulate.WowSim_Execute(runSize, model.Spec, input.itemSet.Items(), nil, tracker.MakeNested())

			printer.Println("SIM " + input.item.BaseName)
			simResult.Print(printer)

			resultChannel <- upgradeItemResultWithSim{*input, *baseSim, simResult}
		}
	})
}

func reportTabulatedSimResults(outputMap map[upgradeMode][]upgradeItemResultWithSim, printer *util.PrintRecorder) {
	byBossThenItem := groupByBossAndItem(outputMap, makeReportSimForItem)

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

			printer.Printf("%10s%5s%45s%10s%10s%10s%10s%10s%10s%10s%10s%s\n",
				"slot", "ilvl", "name",
				"DPS_norm", "sim_d_n",
				"DPS_hero", "sim_d_h",
				"MIT_norm", "sim_m_n",
				"MIT_hero", "sim_m_h",
				"sim_detailed")

			for _, report := range reportList {
				_, bestSimIncrease, mode := bestSimOf(report)

				printer.Printf("%10s%5d%45s%10s%10s%10s%10s%10s%10s%10s%10s\t\t%8s%s\n",
					report.item.Slot.Name(), report.item.Ref.ItemLevel, report.item.BaseName,
					report.byMode[Upgrade_Dps_Normal].increaseStr(), report.byMode[Upgrade_Dps_Normal].percentStrSim(),
					report.byMode[Upgrade_Dps_Heroic].increaseStr(), report.byMode[Upgrade_Dps_Heroic].percentStrSim(),
					report.byMode[Upgrade_Miti_Normal].increaseStr(), report.byMode[Upgrade_Miti_Normal].percentStrSim(),
					report.byMode[Upgrade_Miti_Heroic].increaseStr(), report.byMode[Upgrade_Miti_Heroic].percentStrSim(),
					mode.Name(), bestSimIncrease.CompactStringSignedPercent())
			}

			printer.Println0()
		}
	}
}

func bestSimOf(report *reportForItemWithSim) (simulate.SimResultStats, simulate.SimResultStats, upgradeMode) {
	var bestSim simulate.SimResultStats
	var bestIncrease simulate.SimResultStats
	var bestMode upgradeMode
	best := -1.0
	for _, report := range report.byMode {
		value := report.bestOfSimResults()
		if value > best {
			best = value
			bestSim = report.sim
			bestIncrease = report.increaseSimBreakdown()
		}
	}
	return bestSim, bestIncrease, bestMode
}
