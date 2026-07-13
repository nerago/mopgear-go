package weightfind

import (
	"encoding/json"
	"os"
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/gear_model/ratings"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_rank"
	"paladin_gearing_go/weightfind/weight_highs"
	"slices"
	"sync"
	"time"
)

const c_timeoutSolvers = 500

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

type weightOption struct {
	weight       weight_highs.WeightResult
	accuracy     float64
	accuracyStat float64
	pawnString   string
}

func statWeightsGrid_updateOne(label string, gearModel *gear_model.SpecModel, gearFile string, ratios stats.SimData, weightFileOut string, substituteItems []items.ItemId, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, tracker *util.TrackProgress) {
	tracker.RunOuterTracking(5)
	defer tracker.SetDone()

	// LOAD GEAR TO CHECK
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

	// START BUILDING REPORT
	summary := util.StringBuild2{}
	summary.WriteString("Weights Accuracy Summary ::::: ")
	summary.WriteString(label)
	summary.WriteString(" ::::: ")

	// LOAD OLD WEIGHT VALUES
	best := util_rank.BestCollector1[weightOption]{}
	oldOption := loadOldWeights(label, weightFileOut, simTypes, ratios, mixedInputData, printer)
	if oldOption != nil {
		best.Offer(oldOption, oldOption.accuracy)
		addToSummary(&summary, oldOption, "OLD")
	}

	// SOLVE FOR STAT WEIGHTS
	gridOption := solveGridWeights(label, gearModel, printer, ratios, inputDataGrid, simTypes, inputDataReal)
	if gridOption != nil {
		best.Offer(gridOption, gridOption.accuracy)
		addToSummary(&summary, gridOption, "GRID")

		tweakOption := tweakedWeight(label, gearModel, gridOption.weight, ratios, mixedInputData, simTypes, printer)
		best.Offer(tweakOption, tweakOption.accuracy)
		addToSummary(&summary, tweakOption, "TWEAK")
	}
	tracker.NewChild().SetDone()

	// RANKING WEIGHTS
	rankingOption := solveRankingWeight(label, gearModel, printer, ratios, simTypes, mixedInputData, best.GetBestOptional())
	if rankingOption != nil {
		best.Offer(rankingOption, rankingOption.accuracy)
		addToSummary(&summary, rankingOption, "RANK")

		tweakOption := tweakedWeight(label, gearModel, rankingOption.weight, ratios, mixedInputData, simTypes, printer)
		best.Offer(tweakOption, tweakOption.accuracy)
		addToSummary(&summary, tweakOption, "TWEAK")
	}
	tracker.NewChild().SetDone()

	// SEARCH weights
	searchOption := solveSearchWeights(label, gearModel, ratios, printer, mixedInputData, simTypes, inputDataGrid)
	if searchOption != nil {
		best.Offer(searchOption, searchOption.accuracy)
		addToSummary(&summary, searchOption, "SEARCH")
	}
	tracker.NewChild().SetDone()

	// FINISH SUMMARY REPORT
	printer.Println(summary.String())

	// OVERWRITE WEIGHT FILE
	if bestOption, hasBest := best.GetBestOptional().GetWithFlag(); hasBest {
		util.WriteStringToFile(weightFileOut, bestOption.pawnString)
	}
}

func addToSummary(summary *util.StringBuild2, option *weightOption, prefix string) {
	summary.WriteString(prefix)
	summary.WriteString("=")
	summary.WriteFloat64(option.accuracy, 4)
	summary.WriteString(" (")
	summary.WriteFloat64(option.accuracyStat, 4)
	summary.WriteString(") ")
}

func loadOldWeights(label string, weightFileOut string, simTypes []stats.SimType, ratios stats.SimData, inputDataGrid []weight_highs.WeightInput, printer *util.PrintRecorder) *weightOption {
	oldWeight, oldWeightString, oldWeightExists := ratings.StatRatingsWeights_ReadFile_IfExists(weightFileOut, true, true, true)
	if oldWeightExists {
		oldWeightAsResult := weight_highs.WeightResult_FromRatingsWeight(oldWeight)
		accuracy := EvaluateAccuracyRanged(oldWeightAsResult, simTypes, ratios, inputDataGrid)
		accuracyStat := EvaluateAccuracyStatisticalDeviations(oldWeightAsResult, simTypes, ratios, inputDataGrid)
		printer.Printf("Old Weights accuracy %s normal=%f stat=%f\n", label, accuracy, accuracyStat)
		return &weightOption{oldWeightAsResult, accuracy, accuracyStat, oldWeightString}
	} else {
		return nil
	}
}

func solveGridWeights(label string, gearModel *gear_model.SpecModel, printer *util.PrintRecorder, ratios stats.SimData, inputDataGrid []weight_highs.WeightInput, simTypes []stats.SimType, inputDataReal []weight_highs.WeightInput) *weightOption {
	grid := weight_highs.GridStatWeightProcess1B{}
	grid.OUTLIER = 3
	grid.ROUNDMODE = 2
	grid.SCALEMODE = 1
	grid.CALCMODE = 0
	grid.Init(printer, c_timeoutSolvers)
	grid.SetTargetRatios(ratios)
	grid.SetRequiredStats(gearModel.StatsForWeighting)
	grid.SupplyData(inputDataGrid)
	weightsGridFuture := grid.Run(nil)

	weightsGridOptional := weightsGridFuture.WaitForResultAsOptional()
	if weightsGridOptional.HasValue() {
		weightsGrid := weightsGridOptional.GetOrPanic()
		printer.Println("Grid Weights >>>>> " + label)
		pawnGrid := tools.WritePawnString(weightsGrid, printer)
		accGridOnGridInput := EvaluateAccuracyRanged(weightsGrid, simTypes, ratios, inputDataGrid)
		accGridOnMixedInput := EvaluateAccuracyRanged(weightsGrid, simTypes, ratios, slices.Concat(inputDataReal, inputDataGrid))
		accStats := EvaluateAccuracyStatisticalDeviations(weightsGrid, simTypes, ratios, slices.Concat(inputDataReal, inputDataGrid))
		printer.Printf("Grid Weights accuracy %s gridInput=%f mixInput=%f\n", label, accGridOnGridInput, accGridOnMixedInput)
		return &weightOption{weightsGrid, accGridOnMixedInput, accStats, pawnGrid}
	} else {
		return nil
	}
}

func solveRankingWeight(label string, gearModel *gear_model.SpecModel, printer *util.PrintRecorder, ratios stats.SimData, simTypes []stats.SimType, mixedInputData []weight_highs.WeightInput, bestWeightsSoFar util.Optional[weightOption]) *weightOption {
	ranking := weight_highs.RankingStatWeightProcess3b{}
	ranking.TOTALWEIGHT = 2
	ranking.ALGO = 0
	ranking.Init(printer, c_timeoutSolvers)
	ranking.SetRequiredStats(gearModel.StatsForWeighting)
	ranking.SetTargetRatios(ratios)
	ranking.SupplyData(mixedInputData)

	var weightsRankingFuture *util_async.FutureCancellable[weight_highs.WeightResult]
	if !bestWeightsSoFar.IsEmpty() {
		weightsRankingFuture = ranking.RunSinglePassFromExternal(bestWeightsSoFar.GetOrPanic().weight, nil)
	} else {
		weightsRankingFuture = ranking.RunMultiRound(nil)
	}

	weightsRankingOptional := weightsRankingFuture.WaitForResultAsOptional()
	if weightsRankingOptional.HasValue() {
		weightsRanking := weightsRankingOptional.GetOrPanic()
		printer.Println("Ranking Weights >>>>> " + label)
		pawnRanking := tools.WritePawnString(weightsRanking, printer)
		accuracy := EvaluateAccuracyRanged(weightsRanking, simTypes, ratios, mixedInputData)
		accuracyStats := EvaluateAccuracyStatisticalDeviations(weightsRanking, simTypes, ratios, mixedInputData)
		return &weightOption{weightsRanking, accuracy, accuracyStats, pawnRanking}
	} else {
		return nil
	}
}

func solveSearchWeights(label string, gearModel *gear_model.SpecModel, ratios stats.SimData, printer *util.PrintRecorder, mixedInputData []weight_highs.WeightInput, simTypes []stats.SimType, inputDataGrid []weight_highs.WeightInput) *weightOption {
	search := WeightSearcher2{}
	search.AccuracyMode = 1
	search.Init(gearModel.StatsForWeighting, ratios, printer)
	search.SupplyData(mixedInputData)
	search.SetRanges(-1.0, 10.0)

	cancel := util_async.CancelSignal_Make()
	util_async.CancelAfterTimeout(cancel, time.Second*c_timeoutSolvers, printer)
	weightsSearch := search.Run(cancel)

	printer.Println("Search Weights >>>>> " + label)
	pawnSearch := tools.WritePawnString(weightsSearch, printer)
	accuracySearch := EvaluateAccuracyRanged(weightsSearch, simTypes, ratios, inputDataGrid)
	accuracySearchStat := EvaluateAccuracyStatisticalDeviations(weightsSearch, simTypes, ratios, inputDataGrid)
	return &weightOption{weightsSearch, accuracySearch, accuracySearchStat, pawnSearch}
}

func tweakedWeight(label string, gearModel *gear_model.SpecModel, startWeight weight_highs.WeightResult, ratios stats.SimData, mixedInputData []weight_highs.WeightInput, simTypes []stats.SimType, printer *util.PrintRecorder) *weightOption {
	weightsTweaked, accuracyTweaked := WeightTweakerWithLogging(startWeight, gearModel.StatsForWeighting, ratios, mixedInputData, printer)
	printer.Println("Tweaked Weights >>>>> " + label)
	pawnTweak := tools.WritePawnString(weightsTweaked, printer)
	accuracyStat := EvaluateAccuracyStatisticalDeviations(weightsTweaked, simTypes, ratios, mixedInputData)
	return &weightOption{weightsTweaked, accuracyTweaked, accuracyStat, pawnTweak}
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
