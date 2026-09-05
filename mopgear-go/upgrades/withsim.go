package upgrades

import (
	"github.com/nerago/mopgear-go/db"
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/gear_model/model_factory"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_rank"
)

func processSpecUpgrade(settings *InputSettings, spec *SpecInput, difficulty stats.Difficulty, tracker *util.TrackProgress, printer *util.PrintRecorder) ([]upgradeItemResult, error) {
	if err := model_factory.SetupModelWeights(&spec.Model); err != nil {
		return nil, err
	}

	itemOptions, err := setup.OptionsSetup_FromGearFile(spec.Model.GearFile, &spec.Model, setup.MissingEnchant_Panic, printer)
	if err != nil {
		return nil, err
	}

	upgradeItems := spec.ItemFinder(difficulty)
	upgradeItems = util_collection.RemoveDuplicatesComparable_NewIfChanged(upgradeItems)

	extraTasks := prepareUpgradeTasks(upgradeItems, &itemOptions, spec, settings, printer)

	tracker.RunOuterTracking(1 + len(extraTasks) + settings.ExtraSimForTopResultsCount)
	defer tracker.SetDone()

	baseSim, baseRating, err := runBaseline(spec, settings, &itemOptions, tracker.NewChild(), printer)
	if err != nil {
		return nil, err
	}

	simResults := runEachUpgradeTaskAndSim(extraTasks, &itemOptions, baseRating, baseSim, spec, settings, tracker, printer)

	if settings.ExtraSimForTopResultsCount > 0 {
		runMoreDetailedSimForBestN(simResults, settings.ExtraSimForTopResultsCount, settings.ExtraSimForTopResultsSimSize, &spec.Model, tracker)
	}

	reportBasicResultsSim(simResults, printer, settings.PositiveResultsOnly)
	return simResults, nil
}

func runBaseline(spec *SpecInput, input *InputSettings, baseItems *items.FullOptionsMap, tracker *util.TrackProgress, printer *util.PrintRecorder) (stats.SimData, float64, error) {
	baseRating, baseSet, err := solveBaseLine(printer, baseItems, spec, input)
	if err != nil {
		return stats.SimData{}, 0, err
	}

	baseSim, err2 := simulate.ExecuteUseModel(input.SimSizeBaseline, &spec.Model, baseSet.Items(), nil, tracker)

	return baseSim, baseRating, err2
}

func runEachUpgradeTaskAndSim(extraTasks []upgradeItemTask, baseItems *items.FullOptionsMap, baseRating float64, baseSim stats.SimData, spec *SpecInput, settings *InputSettings, tracker *util.TrackProgress, printer *util.PrintRecorder) []upgradeItemResult {
	printer.Println("TRYING ITEMS")
	simResults := util_async.Map_SliceToSlice(c_upgradeEachThreads, extraTasks, func(task *upgradeItemTask) upgradeItemResult {
		upgradeResult := performUpgradeTask(task, baseItems, baseRating, &spec.Model, printer, spec.SubstituteEmptySlotOnly,
			settings.WeightType, settings.SolverTimeout)

		if upgradeResult.success {
			simResult, err := simulate.ExecuteSpecifyAll(settings.SimSizeItemInitial, spec.Model.SimSpeedUp, spec.Model.Spec, spec.Model.Goal,
				spec.Model.SimulateAs, spec.Model.Professions, upgradeResult.itemSet.Items(), nil, tracker.NewChild())

			if err == nil {
				itemName := db.LookupItemNameByItemId(task.itemRef.ItemId)
				printer.Println("SIM " + itemName)
				simResult.Print(printer)

				upgradeResult.baseSim = baseSim
				upgradeResult.simResult = simResult
			} else {
				util.GlobalWarnHandler(err)
			}
		}

		return upgradeResult
	})
	return simResults
}

func runMoreDetailedSimForBestN(resultList []upgradeItemResult, topResultsCount int, extraSimSize simulate.WowSim_RunSize, model *gear_model.SpecModel, tracker *util.TrackProgress) {
	topN := util_rank.HighestCollector_ForN[*upgradeItemResult](
		uint64(topResultsCount),
		func(a, b **upgradeItemResult) bool { return (*a).Equals(**b) },
	)
	for result := range util_collection.ForPointer(resultList) {
		topN.Offer(&result, result.increaseSim())
	}

	for result := range topN.ResultsSeq() {
		runMoreDetailedSimFor(*result, extraSimSize, model, model.Goal, tracker.NewChild())
	}
}

func runMoreDetailedSimFor(result *upgradeItemResult, extraSimSize simulate.WowSim_RunSize, model *gear_model.SpecModel, goal stats.OptimiseGoal, tracker *util.TrackProgress) {
	updatedSimResult, err := simulate.ExecuteSpecifyAll(extraSimSize, model.SimSpeedUp, model.Spec, goal, model.SimulateAs, model.Professions, result.itemSet.Items(), nil, tracker)
	if err == nil {
		result.simResult = updatedSimResult
	} else {
		util.GlobalWarnHandler(err)
	}
}
