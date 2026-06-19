package upgrades

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
)

// possible entry point
func FindUpgrades_Sim_Run(input *FindUpgrades_SimInputs, goal stats.OptimiseGoal, model *model.Model, gearFile string, upgradeItems []*items.FullItem, substituteItems []items.ItemId, printer *util.PrintRecorder) {
	optionsMap := setup.OptionsSetup_FromGearFile(gearFile, model, setup.MissingEnchant_Panic, printer)

	tracker := util.TrackProgress_Start()

	findUpgradeAndSim(input, &optionsMap, upgradeItems, model, printer, tracker, goal, substituteItems, nil)
}

func findUpgradeAndSim(input *FindUpgrades_SimInputs, baseItems *items.FullOptionsMap, extraItems []*items.FullItem, model *model.Model, printer *util.PrintRecorder, tracker *util.TrackProgress,
	goal stats.OptimiseGoal, substituteItems []items.ItemId, substituteEmptySlotOnly map[items.SlotItem]items.ItemId) []upgradeItemResultWithSim {

	tracker.RunOuterTracking(3)
	defer tracker.Stop()

	initialList, baseSet := findUpgrade(&input.FindUpgrades_BasicInputs, baseItems, extraItems, model, printer, tracker.MakeNested(), goal, input.TargetUpgradeLevel, true, substituteItems, substituteEmptySlotOnly)

	baseSim := simulate.WowSim_Execute_SpecifyAll(input.SimSize, model.Spec, goal, model.SimulateAs, model.Professions, baseSet.Items(), nil, tracker.MakeNested())
	printer.Println("SIM *BASELINE*")
	baseSim.Print(printer)

	simResults := simEachInitialResult(input, initialList, goal, model, &baseSim, tracker.MakeNested(), printer)
	reportBasicResultsSim(simResults, printer, input.PositiveResultsOnly)
	return simResults
}

func simEachInitialResult(input *FindUpgrades_SimInputs, inputList []upgradeItemResult, goal stats.OptimiseGoal, model *model.Model, baseSim *simulate.SimData, tracker *util.TrackProgress, printer *util.PrintRecorder) []upgradeItemResultWithSim {
	tracker.RunOuterTracking(len(inputList))
	defer tracker.Stop()

	return channel_op.Map_SliceToSlice(c_simThreads, inputList, func(initial *upgradeItemResult, resultChannel chan<- upgradeItemResultWithSim) {
		if initial.success {
			simResult := simulate.WowSim_Execute_SpecifyAll(input.SimSize, model.Spec, goal, model.SimulateAs, model.Professions, initial.itemSet.Items(), nil, tracker.MakeNested())

			printer.Println("SIM " + initial.item.BaseName())
			simResult.Print(printer)

			resultChannel <- upgradeItemResultWithSim{*initial, *baseSim, simResult}
		} else {
			resultChannel <- upgradeItemResultWithSim{upgradeItemResult: *initial}
		}
	})
}
