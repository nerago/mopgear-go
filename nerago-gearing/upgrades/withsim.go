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
	"paladin_gearing_go/util/channel_op"
	"slices"
)

const (
	simThreads  = 4
	runSize     = simulate.RunSize_QuickDirty
	runSizeBase = simulate.RunSize_Medium
)

func FindUpgrades_Sim_PaladinMiti_Run() {
	mode := Upgrade_Miti_Heroic
	model := model.Model_PallyProtMitigation()
	gearFile := files.GearFileProtMitigation
	upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	// upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Heroic)
	substituteItems := []items.ItemId{
		95291, // prot tier15 hand normal
		95290, // prot tier15 chest normal
		95292, // prot tier15 head normal
		96667, // prot tier15 leg heroic
		96668, // prot tier15 shoulder heroic
		96657, // ret tier15 legs heroic
		96769, // doomcloak
		96394, // frozen warlord bracer heroic
		96373, // cloudbreaker belt heroic
		96478, // treads of the blind heroic
		95142, // striker's battletags
		95205, // terra-cotta neck
		95178, // lootraptor amulet
		96533, // rein-binders fists heroic
		86957, // heroic bladed tempest ring
		86955, // heroic overwhelm assault belt
		95535, // normal lightning legs
		87015, // heroic clawfeet
		96481, // durumu tentacle heroic
		95513, // scaled tyrant normal
		95140, // shado assault band
	}
	FindUpgrades_Sim_Run(mode, &model, gearFile, upgradeItems, substituteItems)
}

func FindUpgrades_Sim_Run(mode upgradeMode, model *model.Model, gearFile string, upgradeItems []*items.FullItem, substituteItems []items.ItemId) {
	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(2)
	defer tracker.Stop()

	printer := util.PrintRecorder_CreateLogFile()
	optionsMap := setup.OptionsSetup_FromGearFile(gearFile, model, printer)

	addSubstituteItems(&optionsMap, substituteItems, model, printer)

	initialResult, baseSet := findUpgrade(&optionsMap, upgradeItems, model, printer, tracker.MakeNested(), mode)

	baseSim := simulate.WowSim_Execute(runSize, model.Spec, baseSet.Items(), nil, nil)
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
