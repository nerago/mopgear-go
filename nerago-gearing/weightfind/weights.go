package weightfind

import (
	"encoding/json"
	"os"
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/weightfind/weight_highs"
	"slices"
	"sync"
)

const c_timeoutSolvers = 2000

type WeightOptions struct {
	Label           string
	WeightFileOut   string
	GearFile        string
	Model           gear_model.SpecModel
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

func statWeightsGrid_updateOne(label string, gearModel *gear_model.SpecModel, gearFile string, ratios stats.SimData, weightFileOut string, substituteItems []items.ItemId, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, tracker *util.TrackProgress) {
	tracker.RunOuterTracking(3)
	defer tracker.SetDone()

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(gearFile), gearModel, setup.MissingEnchant_Panic, printer)
	currentItemSet := items.FullItemSet_FromMap(currentEquip)
	simTypes := ratios.NonZeroTypes()

	// SIMULATE STAT CHANGES
	inputDataGrid := SimulateSteppedStatChangesForGrid(currentItemSet, printer, simSpeed, gearModel.SimSpeedUp, gearModel.StatsForWeighting, gearModel.Spec, gearModel.Goal, gearModel.SimulateAs, gearModel.Professions, tracker.NewChild())
	inputDataReal := SimulateRealRandomSets(gearFile, substituteItems, gearModel, len(inputDataGrid), simSpeed, false, printer, tracker.NewChild())
	mixedInputData := slices.Concat(inputDataGrid, inputDataReal)

	// SAVE SIM DATA IN CASE WE NEED TO RESTART
	writeWeightInputsToFile(inputDataGrid, files.TempPath+"weightfind-sim-grid-"+label+".json")
	writeWeightInputsToFile(inputDataReal, files.TempPath+"weightfind-sim-real-"+label+".json")

	// TODO report on accuracy rating of grid vs just inputDataGrid

	// SOLVE FOR STAT WEIGHTS
	grid := weight_highs.GridStatWeightProcess1B{}
	grid.OUTLIER = 3
	grid.ROUNDMODE = 2
	grid.SCALEMODE = 1
	grid.Init(printer, c_timeoutSolvers)
	grid.SetTargetRatios(ratios)
	grid.SetRequiredStats(gearModel.StatsForWeighting)
	grid.SetTestMode(simSpeed == simulate.RunSize_TestOnly)
	grid.SupplyData(inputDataGrid)
	weightsGridFuture := grid.Run(nil)
	weightsGrid := weightsGridFuture.WaitForResultOrNilValue()
	printer.Println("Grid Weights >>>>> " + label)
	pawnGrid := tools.WritePawnString(weightsGrid, printer)
	accGridOnGridInput := EvaluateAccuracyRanged(weightsGrid, simTypes, ratios, inputDataGrid)
	accGridOnRealInput := EvaluateAccuracyRanged(weightsGrid, simTypes, ratios, inputDataReal)
	printer.Printf("Grid Weights accuracy %s gridInput=%f realInput=%f\n", label, accGridOnGridInput, accGridOnRealInput)

	ranking := weight_highs.RankingStatWeightProcess3b{}
	ranking.Init(printer, c_timeoutSolvers)
	ranking.SetRequiredStats(gearModel.StatsForWeighting)
	ranking.SetTargetRatios(ratios)
	ranking.SupplyData(mixedInputData)
	var weightsRankingFuture *util_async.FutureCancellable[weight_highs.WeightResult]
	if !weightsGrid.IsEmpty() {
		weightsRankingFuture = ranking.RunSinglePassFromExternal(weightsGrid, nil)
	} else {
		weightsRankingFuture = ranking.RunMultiRound(nil)
	}
	weightsRanking := weightsRankingFuture.WaitForResultOrNilValue()
	printer.Println("Ranking Weights >>>>> " + label)
	pawnRanking := tools.WritePawnString(weightsRanking, printer)

	// TWEAK weights see if dumb changes can do better than grid
	weightsGrid, accuracyGrid := WeightTweakerWithLogging(weightsGrid, gearModel.StatsForWeighting, ratios, mixedInputData, printer)
	printer.Println("Tweaked Grid Weights >>>>> " + label)
	pawnGrid = tools.WritePawnString(weightsGrid, printer)

	// TWEAK ranking weights
	weightsRanking, accuracyRanking := WeightTweakerWithLogging(weightsRanking, gearModel.StatsForWeighting, ratios, mixedInputData, printer)
	printer.Println("Tweaked Ranking Weights >>>>> " + label)
	pawnRanking = tools.WritePawnString(weightsRanking, printer)

	// SEARCH weights
	search := WeightSearcher2{}
	search.Init(gearModel.StatsForWeighting, ratios, printer)
	search.SupplyData(mixedInputData)
	search.SetRanges(-1.0, 10.0)
	weightsSearch := search.Run(util_async.CancelSignal_Make())
	printer.Println("Search Weights >>>>> " + label)
	pawnSearch := tools.WritePawnString(weightsSearch, printer)
	accuracySearch := EvaluateAccuracyRanged(weightsSearch, simTypes, ratios, inputDataGrid)

	accuracyGridStat := EvaluateAccuracyStatisticalDeviations(weightsGrid, simTypes, ratios, inputDataGrid)
	accuracyRankingStat := EvaluateAccuracyStatisticalDeviations(weightsRanking, simTypes, ratios, inputDataGrid)
	accuracySearchStat := EvaluateAccuracyStatisticalDeviations(weightsSearch, simTypes, ratios, inputDataGrid)

	printer.Printf("Weights Accuracy Summary ::::: "+label+" ::::: g=%f r=%f s=%f ::::: g=%f r=%f s=%f\n", accuracyGrid, accuracyRanking, accuracySearch, accuracyGridStat, accuracyRankingStat, accuracySearchStat)

	// OVERWRITE WEIGHT FILE
	if accuracySearch > accuracyGrid && accuracySearch > accuracyRanking {
		util.WriteStringToFile(weightFileOut, pawnSearch)
	} else if accuracyGrid > accuracyRanking {
		util.WriteStringToFile(weightFileOut, pawnGrid)
	} else {
		util.WriteStringToFile(weightFileOut, pawnRanking)
	}
}

func writeWeightInputsToFile(weightInputs []weight_highs.WeightInput, filename string) {
	bytes, err := json.Marshal(weightInputs)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filename, bytes, 0666)
	if err != nil {
		panic(err)
	}
}

func readWeightInputFile(filename string) []weight_highs.WeightInput {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var weightInputs []weight_highs.WeightInput
	err = json.Unmarshal(bytes, &weightInputs)
	if err != nil {
		panic(err)
	}
	return weightInputs
}
