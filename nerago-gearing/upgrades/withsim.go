package upgrades

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

const (
	simThreads = 4
	runSize    = simulate.RunSize_QuickDirty
)

func SomethingWithSim() {
	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(4)
	defer tracker.Stop()

	//modelMitigation := model.Model_PallyProtMitigation()
	modelDps := model.Model_PallyProtMitigation()

	//upgradeNormal := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	upgradeHeroic := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Heroic)

	printer := util.PrintRecorder_CreateLogFile()
	printer.Println("[[[[[[[[[[[[[[[[[[[[ PALLY PROT DPS heroic UPGRADES SIMULATE ]]]]]]]]]]]]]]]]]]]]")
	optionsDps := setup.OptionsSetup_FromGearFile(files.GearFileProtDps, &modelDps, printer)

	initialResult, baseSet := findUpgrade(&optionsDps, upgradeHeroic, &modelDps, printer, tracker.MakeNested(), Upgrade_Dps_Normal)
	simResult := simEachOutput(initialResult, &modelDps, baseSet)
	reportSimUpgrades(simResult, baseSet)
}

func simEachOutput(inputList []upgradeItemResult, model *model.Model, baseSet *items.FullItemSet) []upgradeItemResultWithSim {
	baseSim := simulate.WowSim_Execute(runSize, model.Spec, baseSet.Items(), nil)

	return util.Channel_IterateEach_Multi_AsSlice(simThreads, inputList, func(input *upgradeItemResult, resultChannel chan<- upgradeItemResultWithSim) {
		simResult := simulate.WowSim_Execute(runSize, model.Spec, input.itemSet.Items(), nil)
		resultChannel <- upgradeItemResultWithSim{*input, baseSim, simResult}
	})
}

func reportSimUpgrades(simResult []upgradeItemResultWithSim, baseSet *items.FullItemSet) {
	panic("unimplemented")
}
