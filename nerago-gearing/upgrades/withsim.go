package upgrades

import (
	"cmp"
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"slices"
)

// possible entry point
func FindUpgrades_Sim_Run(mode upgradeMode, model *model.Model, gearFile string, upgradeItems []*items.FullItem, substituteItems []items.ItemId) {
	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(2)
	defer tracker.Stop()

	printer := util.PrintRecorder_CreateLogFile()
	optionsMap := setup.OptionsSetup_FromGearFile(gearFile, model, printer)

	addSubstituteItems(&optionsMap, substituteItems, model, printer)

	initialResult, baseSet := findUpgrade(&optionsMap, upgradeItems, model, printer, tracker.MakeNested(), mode)

	baseSim := simulate.WowSim_Execute(simRunSize, model.Spec, baseSet.Items(), nil, nil)
	printer.Println("SIM *BASELINE*")
	baseSim.Print(printer)

	simResult := simEachInitialResult(initialResult, model, &baseSim, tracker.MakeNested(), printer)

	reportBasicResultsSim(simResult, printer)
}

func addSubstituteItems(optionsMap *items.FullOptionsMap, substituteItems []items.ItemId, model *model.Model, printer *util.PrintRecorder) {
	for _, itemId := range substituteItems {
		if !optionsMap.IncludesItemId(itemId) {
			options, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, 2, model, printer)
			optionsMap.AddSeveralOptions(example.Slot, options)
			printer.Println("SUBSTITUTE " + example.CreateString())
		}
	}
}

func simEachInitialResult(inputList []upgradeItemResult, model *model.Model, baseSim *simulate.SimResultStats, tracker *util.TrackProgress, printer *util.PrintRecorder) []upgradeItemResultWithSim {
	tracker.RunOuterTracking(len(inputList))
	defer tracker.Stop()

	return channel_op.IterateEach_SliceToSlice(simThreads, inputList, func(input *upgradeItemResult, resultChannel chan<- upgradeItemResultWithSim) {
		if input.success {
			simResult := simulate.WowSim_Execute(simRunSize, model.Spec, input.itemSet.Items(), nil, tracker.MakeNested())

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

			printer.Printf("%10s%5s%45s%10s%10s%10s%10s%10s%10s%10s%10s  %10s  %s\n",
				"slot", "ilvl", "name",
				"DPS_norm", "sim_d_n",
				"DPS_hero", "sim_d_h",
				"MIT_norm", "sim_m_n",
				"MIT_hero", "sim_m_h",
				"best_sim",
				"sim_detailed")

			for _, report := range reportList {
				_, bestSimIncrease, bestMode := bestSimOf(report)

				var modeName, simDetail string = "", ""
				if !bestSimIncrease.IsEmpty() {
					modeName = bestMode.Name()
					simDetail = bestSimIncrease.CompactStringSignedPercent()
				}

				printer.Printf("%10s%5d%45s%10s%10s%10s%10s%10s%10s%10s%10s  %10s  %s\n",
					report.item.Slot.Name(), report.item.Ref.ItemLevel, report.item.BaseName,
					report.byMode[Upgrade_Dps_Normal].increaseStr(), report.byMode[Upgrade_Dps_Normal].percentStrSim(),
					report.byMode[Upgrade_Dps_Heroic].increaseStr(), report.byMode[Upgrade_Dps_Heroic].percentStrSim(),
					report.byMode[Upgrade_Miti_Normal].increaseStr(), report.byMode[Upgrade_Miti_Normal].percentStrSim(),
					report.byMode[Upgrade_Miti_Heroic].increaseStr(), report.byMode[Upgrade_Miti_Heroic].percentStrSim(),
					modeName, simDetail)
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
	for mode, result := range report.byMode {
		value := result.bestOfSimResults()
		if value > best {
			best = value
			bestMode = mode
			bestSim = result.sim
			bestIncrease = result.increaseSimBreakdown()
		}
	}
	return bestSim, bestIncrease, bestMode
}
