package weightfind

import (
	"os"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"slices"
	"sync"
)

type WeightOptions struct {
	WeightFileOut   string
	GearFile        string
	Model           model.Model
	SubstituteItems []items.ItemId
}

func StatWeights_updateAll(simSpeed simulate.WowSim_RunSize, printer *util.PrintRecorder, options []WeightOptions) {
	waitGroup := sync.WaitGroup{}
	progress := util.TrackProgress_Start()
	progress.RunOuterTracking(len(options))

	for _, option := range options {
		waitGroup.Go(func() {
			statWeightsGrid_updateOne(&option.Model, option.GearFile, option.Model.SimRatioWeighting, option.WeightFileOut,
				option.SubstituteItems, printer, simSpeed, progress.NewChild())
		})
	}

	waitGroup.Wait()
	progress.SetDone()
}

func statWeightsGrid_updateOne(gearModel *model.Model, gearFile string, ratios stats.SimData, weightFileOut string, substituteItems []items.ItemId, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, tracker *util.TrackProgress) {
	tracker.RunOuterTracking(4)
	defer tracker.SetDone()

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(gearFile), gearModel, setup.MissingEnchant_Panic, printer)
	currentItemSet := items.FullItemSet_FromMap(currentEquip)

	// SIMULATE STAT CHANGES
	inputDataGrid := SimulateSteppedStatChangesForGrid(currentItemSet, printer, simSpeed, gearModel.SimSpeedUp, gearModel.Spec, gearModel.Goal, gearModel.SimulateAs, gearModel.Professions, tracker.NewChild())
	inputDataReal := SimulateRealRandomSets(gearFile, substituteItems, gearModel, 200, simSpeed, false, printer, tracker.NewChild())

	// SOLVE FOR STAT WEIGHTS
	process := stathighs.GridStatWeightProcess{}
	process.Init(printer)
	process.SetTargetRatios(ratios)
	process.SetTestMode(simSpeed == simulate.RunSize_TestOnly)
	process.SupplyData(inputDataGrid)
	weights := process.Run()
	printer.Println(">>>>> Grid Weights:")
	pawn := tools.WritePawnString(weights, printer)

	tracker.NewChild().SetDone() // pretend we were tracking the linear process and mark done

	// TWEAK weights see if dumb changes can do better than grid
	// TODO look into ranking stats solver
	mixedInputData := slices.Concat(inputDataGrid, inputDataReal)
	weights = WeightTweaker(weights, TweakerChangeStats, ratios, mixedInputData, printer)
	printer.Println(">>>>> Tweaked Weights:")
	pawn = tools.WritePawnString(weights, printer)

	// OVERWRITE WEIGHT FILE
	writeFile(weightFileOut, pawn)
}

func writeFile(filename, content string) {
	bytes := []byte(content)
	err := os.WriteFile(filename, bytes, 0)
	if err != nil {
		panic(err)
	}
}
