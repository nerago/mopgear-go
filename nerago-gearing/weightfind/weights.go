package weightfind

import (
	"os"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"slices"
	"sync"
)

type WeightOptions struct {
	WeightFileOut   string
	GearFile        string
	Model           model.Model
	Ratios          simulate.SimData
	SubstituteItems []items.ItemId
}

func StatWeights_updateAll(simSpeed simulate.WowSim_RunSize, printer *util.PrintRecorder, options []WeightOptions) {
	waitGrounp := sync.WaitGroup{}
	progress := util.TrackProgress_Start()
	progress.RunOuterTracking(len(options))

	for _, option := range options {
		waitGrounp.Go(func() {
			statWeightsGrid_updateOne(&option.Model, option.GearFile, option.Ratios, option.WeightFileOut,
				option.SubstituteItems, printer, simSpeed, progress.MakeNested())
		})
	}

	waitGrounp.Wait()
	progress.Stop()
}

func statWeightsGrid_updateOne(gearModel *model.Model, gearFile string, ratios simulate.SimData, weightFileOut string, substituteItems []items.ItemId, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, tracker *util.TrackProgress) {
	tracker.RunOuterTracking(4)

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(gearFile), gearModel, printer)
	currentItemSet := items.FullItemSet_FromMap(currentEquip)

	// SIMULATE STAT CHANGES
	inputDataGrid := GenerateRatingsInputFromArtificalStatOverrides_ForGrid(currentItemSet, printer, simSpeed, gearModel.Spec, gearModel.Goal, gearModel.SimulateAs, gearModel.Professions, tracker.MakeNested())
	inputDataReal := GenerateRatingsInputFromRealRandomSetsGeneral(gearFile, substituteItems, gearModel, 200, simSpeed, false, printer, tracker.MakeNested())

	// SOLVE FOR STAT WEIGHTS
	process := stathighs.GridStatWeightProcess{}
	process.Init(printer)
	process.SetTargetRatios(ratios)
	process.SupplyData(inputDataGrid)
	weights := process.Run()
	printer.Println(">>>>> Grid Weights:")
	pawn := tools.WritePawnString(weights, printer)
	printer.Println(pawn)

	tracker.MakeNested().Stop() // pretend we were tracking the highs and mark done

	// TWEAK weights see if dumb changes can do better than grid
	// TODO look into ranking stats solver
	mixedInputData := slices.Concat(inputDataGrid, inputDataReal)
	WeightTweaker(weights, TweakerChangeStats, ratios, mixedInputData, printer)
	printer.Println(">>>>> Tweaked Weights:")
	pawn = tools.WritePawnString(weights, printer)
	printer.Println(pawn)

	// OVERWRITE WEIGHT FILE
	writeFile(weightFileOut, pawn)

	tracker.Stop()
}

func writeFile(filename, content string) {
	bytes := []byte(content)
	err := os.WriteFile(filename, bytes, 0)
	if err != nil {
		panic(err)
	}
}
