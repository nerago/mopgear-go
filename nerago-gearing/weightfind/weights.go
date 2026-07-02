package weightfind

import (
	"encoding/json"
	"os"
	"paladin_gearing_go/files"
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

const c_timeoutSolvers = 2000

type WeightOptions struct {
	Label           string
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
			statWeightsGrid_updateOne(option.Label, &option.Model, option.GearFile, option.Model.SimRatioWeighting, option.WeightFileOut,
				option.SubstituteItems, printer, simSpeed, progress.NewChild())
		})
	}

	waitGroup.Wait()
	progress.SetDone()
}

func statWeightsGrid_updateOne(label string, gearModel *model.Model, gearFile string, ratios stats.SimData, weightFileOut string, substituteItems []items.ItemId, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, tracker *util.TrackProgress) {
	tracker.RunOuterTracking(3)
	defer tracker.SetDone()

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(gearFile), gearModel, setup.MissingEnchant_Panic, printer)
	currentItemSet := items.FullItemSet_FromMap(currentEquip)

	// SIMULATE STAT CHANGES
	inputDataGrid := SimulateSteppedStatChangesForGrid(currentItemSet, printer, simSpeed, gearModel.SimSpeedUp, gearModel.StatsForWeighting, gearModel.Spec, gearModel.Goal, gearModel.SimulateAs, gearModel.Professions, tracker.NewChild())
	inputDataReal := SimulateRealRandomSets(gearFile, substituteItems, gearModel, len(inputDataGrid), simSpeed, false, printer, tracker.NewChild())
	mixedInputData := slices.Concat(inputDataGrid, inputDataReal)

	// SAVE SIM DATA IN CASE WE NEED TO RESTART
	writeWeightInputsToFile(inputDataGrid, files.TempPath+"weightfind-sim-grid-"+label+".json")
	writeWeightInputsToFile(inputDataReal, files.TempPath+"weightfind-sim-real-"+label+".json")

	// TODO avoid changing set bonuses
	// TODO report on accuracy rating of grid vs just inputDataGrid

	// SOLVE FOR STAT WEIGHTS
	grid := stathighs.GridStatWeightProcess1B{}
	grid.OUTLIER = 3
	grid.ROUNDMODE = 2
	grid.SCALEMODE = 1
	grid.Init(printer, c_timeoutSolvers)
	grid.SetTargetRatios(ratios)
	grid.SetRequiredStats(gearModel.StatsForWeighting)
	grid.SetTestMode(simSpeed == simulate.RunSize_TestOnly)
	grid.SupplyData(inputDataGrid)
	weightsGrid := grid.Run(nil)
	printer.Println("Grid Weights >>>>> " + label)
	pawnGrid := tools.WritePawnString(weightsGrid, printer)
	accGridOnGridInput := EvaluateAccuracy(weightsGrid, inputDataGrid, ratios)
	accGridOnRealInput := EvaluateAccuracy(weightsGrid, inputDataReal, ratios)
	printer.Printf("Grid Weights accuracy %s gridInput=%f realInput=%f\n", label, accGridOnGridInput, accGridOnRealInput)

	ranking := stathighs.RankingStatWeightProcess3b{}
	ranking.SCALE1 = false
	ranking.FINAL = 0
	ranking.Init(printer, c_timeoutSolvers)
	ranking.SetRequiredStats(gearModel.StatsForWeighting)
	ranking.SetTargetRatios(ratios)
	ranking.SupplyData(mixedInputData)
	var weightsRanking stathighs.WeightResult
	if !weightsGrid.IsEmpty() {
		weightsRanking = ranking.RunSinglePassFromExternal(weightsGrid, nil)
	} else {
		weightsRanking = ranking.RunMultiRound(nil)
	}
	printer.Println("Ranking Weights >>>>> " + label)
	pawnRanking := tools.WritePawnString(weightsRanking, printer)

	// TWEAK weights see if dumb changes can do better than grid
	weightsGrid, accuracyGrid := WeightTweaker(weightsGrid, gearModel.StatsForWeighting, ratios, mixedInputData, printer)
	printer.Println("Tweaked Grid Weights >>>>> " + label)
	pawnGrid = tools.WritePawnString(weightsGrid, printer)

	// TWEAK ranking weights
	weightsRanking, accuracyRanking := WeightTweaker(weightsRanking, gearModel.StatsForWeighting, ratios, mixedInputData, printer)
	printer.Println("Tweaked Ranking Weights >>>>> " + label)
	pawnRanking = tools.WritePawnString(weightsRanking, printer)

	// OVERWRITE WEIGHT FILE
	if accuracyGrid > accuracyRanking {
		writeFile(weightFileOut, pawnGrid)
	} else {
		writeFile(weightFileOut, pawnRanking)
	}
}

func writeFile(filename, content string) {
	bytes := []byte(content)
	err := os.WriteFile(filename, bytes, 0666)
	if err != nil {
		panic(err)
	}
}

func writeWeightInputsToFile(weightInputs []stathighs.WeightInput, filename string) {
	bytes, err := json.Marshal(weightInputs)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filename, bytes, 0666)
	if err != nil {
		panic(err)
	}
}

func readWeightInputFile(filename string) []stathighs.WeightInput {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var weightInputs []stathighs.WeightInput
	err = json.Unmarshal(bytes, &weightInputs)
	if err != nil {
		panic(err)
	}
	return weightInputs
}
