package weightfind

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
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
	"paladin_gearing_go/util/util_rank"
	"paladin_gearing_go/weightfind/weight_highs"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
	"time"
)

const c_timeoutSolvers = 3000
const c_simDataAgeMax = 48 * time.Hour
const c_updateThreadCount = 8

type WeightUpdateProcess struct {
	simSpeed     simulate.WowSim_RunSize
	forceSkipSim bool
	printer      *util.PrintRecorder
	specs        []*WeightSpec
}

type WeightSpec struct {
	Label           string
	WeightFileOut   string
	GearFile        string
	Model           gear_model.SpecModel
	SubstituteItems []items.ItemId

	process *WeightUpdateProcess

	simTypes    []stats.SimType
	statTypes   []stats.StatType
	targetRatio weight_types.SimPriorityBasic

	choices []weightChoice
	summary util.StringBuild2

	dataGrid, dataRand, dataAll []weight_types.WeightInput
}

type weightChoice struct {
	choiceName   string
	weight       weight_types.Weight1Basic
	accuracy     float64
	accuracyStat float64
	pawnString   string
}

func (wup *WeightUpdateProcess) Init(simSpeed simulate.WowSim_RunSize, forceSkipSim bool, printer *util.PrintRecorder) {
	wup.simSpeed = simSpeed
	wup.forceSkipSim = forceSkipSim
	wup.printer = printer
}

func (wup *WeightUpdateProcess) AddSpec(spec *WeightSpec) {
	wup.specs = append(wup.specs, spec)
}

func (wup *WeightUpdateProcess) Run(cancel util_async.CancelSignal) {
	progress := util.TrackProgress_Start()
	progress.RunOuterTracking(len(wup.specs))
	defer progress.SetDone()

	summaries := util_async.Map_SliceToSlice_Cancellable(c_updateThreadCount, wup.specs, cancel, func(spec **WeightSpec) string {
		return (*spec).updateOne(progress.NewChild())
	})

	for _, summary := range summaries {
		wup.printer.Println(summary)
	}
}

func (spec *WeightSpec) updateOne(tracker *util.TrackProgress) string {
	// each simulator process is considered 1/3, then remaining solving is remaining third.
	// only very small sim runs should be overpowered by solvers
	tracker.RunOuterTracking(3)
	defer tracker.SetDone()

	spec.statTypes = spec.Model.StatsForWeighting
	spec.simTypes = spec.Model.SimPriority.SimTypes()
	spec.targetRatio = spec.Model.SimPriority

	// READ OLD DATA AND/OR RUN SIM
	spec.prepareSimData(tracker)

	// START BUILDING REPORT
	spec.summary.WriteString("Weights Accuracy Summary ::::: ")
	spec.summary.WriteString(spec.Label)
	spec.summary.WriteString(" ::::: ")

	// LOAD OLD WEIGHT VALUES
	spec.loadOldWeights()

	// GRID WEIGHTS
	for gridMode := range 2 {
		spec.solveGridWeights(gridMode)
	}

	// FORMULA2 WEIGHTS
	spec.solveFormulaWeight()

	// SEARCH weights
	for searchMode := range 2 {
		spec.solveSearchWeights(searchMode)
	}

	// RANKING WEIGHTS - late in order so can use a good start point
	for rankMode := range 2 {
		spec.solveRankingWeight(rankMode)
	}

	// TWEAK EACH
	spec.tweakEachWeight()

	// OVERWRITE WEIGHT FILE
	if bestChoice, hasBest := spec.bestWeightChoice(); hasBest {
		util.WriteStringToFile(spec.WeightFileOut, bestChoice.pawnString)

		spec.summary.WriteString(" ::::: ")
		spec.summary.WriteString(bestChoice.choiceName)
		spec.summary.WriteRune(' ')
		spec.summary.WriteFloat64(bestChoice.accuracy, 4)
	}

	// FINISH SUMMARY REPORT
	spec.process.printer.PrintlnFromBuild(spec.summary)
	return spec.summary.String()
}

func (spec *WeightSpec) prepareSimData(tracker *util.TrackProgress) {
	// READ IN ANY RECENT DATA
	tempPathGrid := files.TempPath + "weightfind-sim-grid-" + spec.Label + ".json"
	tempPathReal := files.TempPath + "weightfind-sim-real-" + spec.Label + ".json"
	inputDataGrid, dataAgeGrid := readWeightInputFile(tempPathGrid)
	inputDataReal, dataAgeReal := readWeightInputFile(tempPathReal)

	// DO WE ACCEPT THE OLD DATA
	if dataAgeGrid > c_simDataAgeMax && !spec.process.forceSkipSim {
		inputDataGrid = nil
	}
	if dataAgeReal > c_simDataAgeMax && !spec.process.forceSkipSim {
		inputDataReal = nil
	}

	// SIMULATE STAT CHANGES, SAVE SIM DATA IN CASE WE NEED TO RESTART
	if inputDataGrid == nil {
		currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(spec.GearFile), &spec.Model, setup.MissingEnchant_Panic, spec.process.printer)
		currentItemSet := items.FullItemSet_FromMap(currentEquip)
		inputDataGrid = SimulateSteppedStatChangesForGrid(currentItemSet, spec.process.printer, spec.process.simSpeed,
			spec.Model.SimSpeedUp, spec.Model.StatsForWeighting, spec.Model.Spec, spec.Model.Goal, spec.Model.SimulateAs,
			spec.Model.Professions, tracker.NewChild(), spec.Label)
		writeWeightInputsToFile(inputDataGrid, tempPathGrid)
	} else {
		tracker.NewChild().SetDone()
	}
	if inputDataReal == nil {
		inputDataReal = SimulateRealRandomSets(spec.GearFile, spec.SubstituteItems, &spec.Model, grid_sim_max_run_count,
			spec.process.simSpeed, false, spec.process.printer, tracker.NewChild(), spec.Label)
		writeWeightInputsToFile(inputDataReal, tempPathReal)
	} else {
		tracker.NewChild().SetDone()
	}

	spec.dataGrid = inputDataGrid
	spec.dataRand = inputDataReal
	spec.dataAll = slices.Concat(inputDataGrid, inputDataReal)
}

func (spec *WeightSpec) addChoice(choice weightChoice) {
	spec.choices = append(spec.choices, choice)
	spec.addToSummary(choice)
}

func (spec *WeightSpec) addToSummary(option weightChoice) {
	spec.summary.WriteString(option.choiceName)
	spec.summary.WriteString("=")
	spec.summary.WriteFloat64(option.accuracy, 4)
	spec.summary.WriteString(" (")
	spec.summary.WriteFloat64(option.accuracyStat, 4)
	spec.summary.WriteString(") ")
}

func (spec *WeightSpec) loadOldWeights() {
	oldWeightBlock, _, oldWeightExists := tools.PawnWeightReadFile(spec.WeightFileOut)
	if oldWeightExists {
		oldWeight := weight_types.Weight1Basic_FromBlock(oldWeightBlock)
		spec.evaluateWeight("OLD", &oldWeight, &oldWeight)
	}
}

func (spec *WeightSpec) solveGridWeights(gridOutlierSetting int) {
	grid := weight_highs.GridStatWeightProcess1B{}
	grid.OUTLIER = gridOutlierSetting
	grid.SCALEMODE = 1
	grid.ROUNDMODE = 2
	grid.CALCMODE = 2
	grid.Init(spec.process.printer, c_timeoutSolvers)
	grid.SetTargetRatios(spec.targetRatio)
	grid.SetRequiredStats(spec.statTypes)
	grid.SupplyData(spec.dataGrid)
	weightsFuture := grid.Run()

	choiceName := fmt.Sprintf("grid %d", gridOutlierSetting)
	spec.evaluateWeightFuture(choiceName, weightsFuture)
}

func (spec *WeightSpec) evaluateWeightFuture(choiceName string, weightResultFuture *util_async.FutureCancellable[weight_types.WeightResult]) {
	if weightResult, hasResult := weightResultFuture.WaitForResult(); hasResult {
		weightOrig := weightResult.Weight
		weight1 := weightResult.AsWeight1()
		if weightOrig != nil && weight1 != nil {
			spec.evaluateWeight(choiceName, weight1, weightOrig)
		}
	}
}

func (spec *WeightSpec) evaluateWeight(choiceName string, weight1 *weight_types.Weight1Basic, weightOrig weight_types.IWeight) {
	accuracyOrig := EvaluateAccuracy(weightOrig, spec.simTypes, &spec.targetRatio, spec.dataAll)
	accuracy1 := EvaluateAccuracy(weight1, spec.simTypes, &spec.targetRatio, spec.dataAll)
	accuracy1Stat := EvaluateAccuracyStatistical(weight1, spec.simTypes, &spec.targetRatio, spec.dataAll)

	pawnString := tools.WritePawnString(*weight1, spec.process.printer)
	spec.process.printer.Println(weightOrig.String())
	tools.WriteWeightString(weightOrig, spec.process.printer)

	if weight1.IsEmpty() || weightOrig.IsEmpty() {
		spec.process.printer.Printf("Weights accuracy %s %s EMPTY normal=%f pref=%f stat=%f\n", spec.Label, choiceName, accuracy1, accuracyOrig, accuracy1Stat)
		spec.addChoice(weightChoice{choiceName, *weight1, 0, 0, pawnString})
	} else {
		spec.process.printer.Printf("Weights accuracy %s %s normal=%f pref=%f stat=%f\n", spec.Label, choiceName, accuracy1, accuracyOrig, accuracy1Stat)
		spec.addChoice(weightChoice{choiceName, *weight1, accuracy1, accuracy1Stat, pawnString})
	}
}

func (spec *WeightSpec) bestWeightChoice() (weightChoice, bool) {
	best := util_rank.BestCollector1[weightChoice]{}
	for _, choice := range spec.choices {
		best.Offer(&choice, choice.accuracyStat)
	}
	return best.GetBestOptional().GetWithFlag()
}

func (spec *WeightSpec) solveRankingWeight(rankMode int) {
	if rankMode == 0 {
		ranking := weight_highs.RankingStatWeightProcess3b{}
		ranking.TOTALWEIGHT = 2
		ranking.ALGO = 0
		ranking.Init(spec.process.printer, c_timeoutSolvers)
		ranking.SetRequiredStats(spec.statTypes)
		ranking.SetTargetRatios(spec.targetRatio)
		ranking.SupplyData(spec.dataAll)
		var weightsFuture *util_async.FutureCancellable[weight_types.WeightResult]
		if bestWeightsSoFar, hasBest := spec.bestWeightChoice(); hasBest {
			weightsFuture = ranking.RunSinglePassFromExternal(bestWeightsSoFar.weight)
		} else {
			weightsFuture = ranking.RunMultiRound()
		}
		spec.evaluateWeightFuture("RANK3", weightsFuture)
	} else {
		ranking := weight_highs.RankingStatWeightProcess{}
		ranking.RANKMODE = 0
		ranking.WEIGHTSUM = 0
		ranking.Init(spec.process.printer)
		ranking.SetRequiredStats(spec.statTypes)
		ranking.SetTargetRatios(spec.targetRatio)
		ranking.SupplyData(spec.dataAll)
		weightsFuture := ranking.Run(c_timeoutSolvers)
		spec.evaluateWeightFuture("RANK1", weightsFuture)
	}
}

func (spec *WeightSpec) solveFormulaWeight() {
	comp := weight_highs.FormulaStatWeightProcess2{}
	comp.Init(spec.process.printer)
	comp.SetRequiredStats(spec.statTypes)
	comp.SetTargetRatios(spec.targetRatio)
	comp.SetMinimumIncludeRate(1.0) // TODO can we experiment
	comp.SupplyData(spec.dataAll)
	weights2Future := comp.Run(c_timeoutSolvers)
	spec.evaluateWeightFuture("FORM2", weights2Future)
}

func (spec *WeightSpec) solveSearchWeights(searchMode int) {
	cancel := util_async.CancelSignal_Make()
	timer := util_async.CancelAfterTimeout(cancel, time.Second*c_timeoutSolvers, spec.process.printer)
	defer timer.Stop()

	var weightResult weight_types.WeightResult
	if searchMode == 0 {
		search := WeightSearcher2{}
		search.AccuracyStatistical = false
		search.Init(spec.statTypes, spec.targetRatio, spec.process.printer)
		search.SupplyData(spec.dataAll)
		search.SetRanges(-1.0, 10.0)
		weightResult = search.Run(cancel)
		spec.evaluateWeight("SEARCH2", weightResult.AsWeight1(), weightResult.Weight)
	} else {
		search := WeightSearcher3{}
		search.AccuracyStatistical = true
		search.Init(spec.statTypes, spec.targetRatio)
		search.SupplyData(spec.dataAll)
		search.SetRanges(-1.0, 10.0)
		weightResult = search.Run(cancel)
		spec.evaluateWeight("SEARCH3", weightResult.AsWeight1(), weightResult.Weight)
	}
}

func (spec *WeightSpec) tweakEachWeight() {
	currentChoiceSize := len(spec.choices)
	for i := range currentChoiceSize {
		choice := spec.choices[i]
		if !choice.weight.IsEmpty() {
			weightsTweaked, _ := WeightTweakerWithLogging(choice.weight, spec.statTypes, &spec.targetRatio, spec.dataAll, spec.process.printer)
			spec.evaluateWeight(choice.choiceName+"_TWEAK", &weightsTweaked, &weightsTweaked)
		}
	}
}

func writeWeightInputsToFile(weightInputs []weight_types.WeightInput, filename string) {
	bytes, err := json.Marshal(weightInputs)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filename, bytes, 0666)
	if err != nil {
		panic(err)
	}
}

func readWeightInputFile(filename string) ([]weight_types.WeightInput, time.Duration) {
	statInfo, err := os.Stat(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, 0
	} else if err != nil {
		panic(err)
	}

	// only use data from "today"
	dataAge := time.Since(statInfo.ModTime())

	bytes, err := os.ReadFile(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, 0
	} else if err != nil {
		panic(err)
	}

	var weightInputs []weight_types.WeightInput
	err = json.Unmarshal(bytes, &weightInputs)
	if err != nil {
		panic(err)
	}

	return weightInputs, dataAge
}
