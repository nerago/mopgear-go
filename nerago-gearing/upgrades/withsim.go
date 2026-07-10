package upgrades

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
)

// possible entry point
func FindUpgrades_Sim_Run(input *FindUpgrades_SimInputs, goal stats.OptimiseGoal, model *gear_model.SpecModel, gearFile string, upgradeItems []loaders.ItemFoundRef, substituteItems []items.ItemId, printer *util.PrintRecorder) {
	optionsMap := setup.OptionsSetup_FromGearFile(gearFile, model, setup.MissingEnchant_Panic, printer)

	tracker := util.TrackProgress_Start()

	findUpgradeAndSim(input, &optionsMap, upgradeItems, model, printer, tracker, goal, substituteItems, nil)
}

func findUpgradeAndSim(input *FindUpgrades_SimInputs, baseItems *items.FullOptionsMap, extraItems []loaders.ItemFoundRef, model *gear_model.SpecModel, printer *util.PrintRecorder, tracker *util.TrackProgress,
	goal stats.OptimiseGoal, substituteItems []items.ItemId, substituteEmptySlotOnly map[items.SlotItem]items.ItemId) []upgradeItemResultWithSim {

	extraTasks := prepareUpgradeInfo(extraItems, input.TargetUpgradeLevel, printer, &input.FindUpgrades_BasicInputs, baseItems, goal, substituteItems, model)

	tracker.RunOuterTracking((len(extraTasks) + 1) * 2)
	defer tracker.SetDone()

	baseSim, baseRating := runBaselineAndSim(printer, baseItems, model, tracker, input, goal)

	simResults := runEachUpgradeTaskAndSim(printer, extraTasks, baseItems, baseRating, model, tracker, substituteEmptySlotOnly, input, goal, baseSim)

	reportBasicResultsSim(simResults, printer, input.PositiveResultsOnly)
	return simResults
}

func runBaselineAndSim(printer *util.PrintRecorder, baseItems *items.FullOptionsMap, model *gear_model.SpecModel, tracker *util.TrackProgress, input *FindUpgrades_SimInputs, goal stats.OptimiseGoal) (stats.SimData, float64) {
	baseRating, baseSet := findBaseLine(printer, baseItems, model, tracker.NewChild())
	baseSim := simulate.WowSim_Execute_SpecifyAll(input.SimSize, model.SimSpeedUp, model.Spec, goal, model.SimulateAs, model.Professions, baseSet.Items(), nil, tracker.NewChild())
	printer.Println("SIM *BASELINE*")
	baseSim.Print(printer)
	return baseSim, baseRating
}

func runEachUpgradeTaskAndSim(printer *util.PrintRecorder, extraTasks []upgradeItemTask, baseItems *items.FullOptionsMap, baseRating float64, model *gear_model.SpecModel, tracker *util.TrackProgress, substituteEmptySlotOnly map[items.SlotItem]items.ItemId, input *FindUpgrades_SimInputs, goal stats.OptimiseGoal, baseSim stats.SimData) []upgradeItemResultWithSim {
	printer.Println("TRYING ITEMS")
	simResults := util_async.Map_SliceToSlice(c_upgradeEachThreads+c_simThreads, extraTasks, func(task *upgradeItemTask) upgradeItemResultWithSim {
		itemName := db.LookupItemNameByItemId(task.item.ItemId)
		initial := performUpgradeTask(task, baseItems, baseRating, model, printer, tracker.NewChild(), true, substituteEmptySlotOnly)
		if initial.success {
			simResult := simulate.WowSim_Execute_SpecifyAll(input.SimSize, model.SimSpeedUp, model.Spec, goal, model.SimulateAs, model.Professions, initial.itemSet.Items(), nil, tracker.NewChild())

			printer.Println("SIM " + itemName)
			simResult.Print(printer)
			return upgradeItemResultWithSim{initial, baseSim, simResult}
		} else {
			return upgradeItemResultWithSim{upgradeItemResult: initial}
		}
	})
	return simResults
}
