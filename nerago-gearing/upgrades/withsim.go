package upgrades

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
)

// possible entry point
func FindUpgrades_Sim_Run(input *FindUpgrades_SimInputs, goal UpgradeGoal, model *model.Model, gearFile string, upgradeItems []*items.FullItem, substituteItems []items.ItemId, printer *util.PrintRecorder) {
	optionsMap := setup.OptionsSetup_FromGearFile(gearFile, model, setup.MissingEnchant_Panic, printer)

	tracker := util.TrackProgress_Start()

	findUpgradeAndSim(input, &optionsMap, upgradeItems, model, printer, tracker, goal, substituteItems, nil)
}

func findUpgradeAndSim(input *FindUpgrades_SimInputs, baseItems *items.FullOptionsMap, extraItems []*items.FullItem, model *model.Model, printer *util.PrintRecorder, tracker *util.TrackProgress,
	goal UpgradeGoal, substituteItems []items.ItemId, substituteEmptySlotOnly map[items.SlotItem]items.ItemId) []upgradeItemResultWithSim {

	tracker.RunOuterTracking(3)
	defer tracker.Stop()

	initialList, baseSet := findUpgrade(&input.FindUpgrades_BasicInputs, baseItems, extraItems, model, printer, tracker.MakeNested(), goal, true, substituteItems, substituteEmptySlotOnly)

	baseSim := simulate.WowSim_Execute(input.SimSize, model.Spec, baseSet.Items(), model.Professions, nil, tracker.MakeNested())
	printer.Println("SIM *BASELINE*")
	baseSim.Print(printer)

	simResults := simEachInitialResult(input, initialList, model, &baseSim, tracker.MakeNested(), printer)
	reportBasicResultsSim(simResults, printer, input.PositiveResultsOnly)
	return simResults
}

func simEachInitialResult(input *FindUpgrades_SimInputs, inputList []upgradeItemResult, model *model.Model, baseSim *simulate.SimResultStats, tracker *util.TrackProgress, printer *util.PrintRecorder) []upgradeItemResultWithSim {
	tracker.RunOuterTracking(len(inputList))
	defer tracker.Stop()

	return channel_op.IterateEach_SliceToSlice(c_simThreads, inputList, func(initial *upgradeItemResult, resultChannel chan<- upgradeItemResultWithSim) {
		if initial.success {
			simResult := simulate.WowSim_Execute(input.SimSize, model.Spec, initial.itemSet.Items(), model.Professions, nil, tracker.MakeNested())

			printer.Println("SIM " + initial.item.BaseName)
			simResult.Print(printer)

			resultChannel <- upgradeItemResultWithSim{*initial, *baseSim, simResult}
		} else {
			resultChannel <- upgradeItemResultWithSim{upgradeItemResult: *initial}
		}
	})
}
