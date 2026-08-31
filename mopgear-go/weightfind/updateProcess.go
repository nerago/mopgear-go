package weightfind

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/bartolsthoorn/gohighs/highs"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_rank"
	"github.com/nerago/mopgear-go/weightfind/weight_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting3"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting4"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

const (
	c_simDataAgeMax = 48 * time.Hour
	//c_updateThreadCount = 3
	c_ratioThreadCount = 12

	c_eachSimTargetGenerateDataCount = 600

	c_dataSampleFitRank = 300
	c_dataSampleGrid    = 96
	c_useSamplingFit    = true
	c_useSamplingRank4  = true
	c_useSamplingGrid   = false
)

type WeightUpdateProcess struct {
	simSpeed     simulate.WowSim_RunSize
	forceSkipSim bool
	skipSolve    bool
	timeoutEach  int
	printer      *util.PrintRecorder
	specs        []*WeightSpec
}

type WeightSpec struct {
	Label           string
	WeightFile1     string
	WeightFile2     string
	WeightFile3     string
	GearFile        string
	Model           gear_model.SpecModel
	FixStatsMode    weight_types.FixStatsRangeMode
	SubstituteItems []items.ItemId

	process *WeightUpdateProcess

	simTypes    []stats.SimType
	statTypes   []stats.StatType
	targetRatio weight_types.SimPriorityBasic

	choices []weightChoice
	summary util.StringBuild2

	dataGrid, dataRand, dataFit, dataAll []weight_types.WeightInput
}

type weightChoice struct {
	choiceName    string
	weight        weight_types.Weight1Basic
	weight2       *weight_types.Weight2Extended
	weight3       *weight_types.Weight3ExtendedRanged
	hadExtended   bool
	accuracy1     float64
	accuracy1Stat float64
	accuracyX     float64
	accuracyXStat float64
	pawnString    string
	weightResult  *weight_types.WeightResultCommon
	weightOrig    weight_types.IWeight
}

func (wup *WeightUpdateProcess) Init(simSpeed simulate.WowSim_RunSize, forceSkipSim bool, skipSolve bool, timeoutEach int, printer *util.PrintRecorder) {
	wup.simSpeed = simSpeed
	wup.forceSkipSim = forceSkipSim
	wup.skipSolve = skipSolve
	wup.timeoutEach = timeoutEach
	wup.printer = printer
}

func (wup *WeightUpdateProcess) AddSpec(spec *WeightSpec) {
	spec.process = wup
	spec.statTypes = spec.Model.StatsForWeighting
	spec.simTypes = spec.Model.SimPriority.SimTypes()
	spec.targetRatio = spec.Model.SimPriority
	wup.specs = append(wup.specs, spec)
}

func (wup *WeightUpdateProcess) Run(cancel util_async.CancelSignal, outerThreads int) {
	progress := util.TrackProgress_Start()
	progress.RunOuterTracking(len(wup.specs))
	defer progress.SetDone()

	summaries, err := util_async.Map_SliceToSlice_Cancellable_PassError(outerThreads, wup.specs, cancel, func(spec **WeightSpec) (string, error) {
		return (*spec).updateSpec(progress.NewChild(), cancel)
	})

	if err != nil {
		panic(err)
	}

	for _, summary := range summaries {
		wup.printer.Println(summary)
	}

	for _, spec := range wup.specs {
		spec.tabularReport(spec.process.printer)
	}
}

func (spec *WeightSpec) tabularReportWriteFile(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	printer := util.FilePrintableMake(file)
	printer.Printf("Weight detail output at %s\n", time.Now().Format(time.DateTime))
	spec.tabularReport(printer)

	err = file.Close()
	if err != nil {
		panic(err)
	}
}

func (spec *WeightSpec) tabularReport(print util.Printable) {
	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("algo", false)
	tab.AddColumnHeader("acc1", false)
	tab.AddColumnHeader("acc1_stat", false)
	tab.AddColumnHeader("accX", false)
	tab.AddColumnHeader("accX_stat", false)
	tab.AddColumnHeader("time", false)
	tab.AddColumnHeader("status", false)
	tab.AddColumnHeader("pawn", false)

	slices.SortFunc(spec.choices, func(a, b weightChoice) int {
		return cmp.Compare(max(a.accuracy1, a.accuracy1Stat), max(b.accuracy1, b.accuracy1Stat))
	})
	for choice := range util_collection.ForPointer(spec.choices) {
		row := make([]string, 0, tab.ColumnCount())
		row = append(row, choice.choiceName)
		row = append(row, strconv.FormatFloat(choice.accuracy1, 'f', 4, 64))
		row = append(row, strconv.FormatFloat(choice.accuracy1Stat, 'f', 4, 64))
		if choice.hadExtended {
			row = append(row, strconv.FormatFloat(choice.accuracyX, 'f', 4, 64))
			row = append(row, strconv.FormatFloat(choice.accuracyXStat, 'f', 4, 64))
		} else {
			row = append(row, "", "")
		}
		if choice.weightResult != nil {
			row = append(row, choice.weightResult.SolveTime.String())
			row = append(row, choice.weightResult.Status)
		} else {
			row = append(row, "", "")
		}
		row = append(row, choice.pawnString)
		tab.AddRow(row)
	}

	print.Printf("TABLE %s\n", spec.Label)
	tab.Write(print)
}

func (spec *WeightSpec) updateSpec(tracker *util.TrackProgress, cancel util_async.CancelSignal) (string, error) {
	// each simulator process is considered 1/4, fitting is 1/4, then remaining solving is remaining.
	tracker.RunOuterTracking(5)
	defer tracker.SetDone()

	// READ OLD DATA AND/OR RUN SIM
	err := spec.prepareSimData(tracker, cancel)
	if err != nil {
		return "", err
	}

	// START BUILDING REPORT
	spec.summary.WriteString("Weights Accuracy Summary ::::: ")
	spec.summary.WriteString(spec.Label)
	spec.summary.WriteString(" ::::: ")

	spec.loadOldWeights()

	if !spec.process.skipSolve {
		spec.runSolvers(tracker, cancel)
		spec.tweakEachWeight()
	}

	spec.reportAndWriteWeights()

	spec.dataGrid = nil
	spec.dataRand = nil
	spec.dataFit = nil
	spec.dataAll = nil

	return spec.summary.String(), nil
}

func (spec *WeightSpec) runSolvers(tracker *util.TrackProgress, cancel util_async.CancelSignal) {
	// FORMULA2 WEIGHTS - MIP
	spec.solveFormulaWeight(cancel)

	// FITTING - Slow MIP
	spec.solveFittingWeight(cancel, tracker.NewChild())
	spec.solveFittingFast(cancel)

	// GRID WEIGHTS - GPU*2 - later for less contention
	for gridMode := range 2 {
		spec.solveGridWeights(gridMode, cancel)
	}

	// SEARCH weights - Non-Highs
	for searchMode := range 3 {
		spec.solveSearchWeights(searchMode, cancel)
	}

	// RANKING WEIGHTS - simplex*2, IPX*2
	for rankMode := range 6 {
		spec.solveRankingWeight(rankMode, cancel)
	}
}

func (spec *WeightSpec) reportAndWriteWeights() {
	// OVERWRITE WEIGHT FILE
	if bestChoice, hasBest := spec.bestWeightChoice1(); hasBest {
		spec.summary.WriteString(" ::::: ")
		spec.summary.WriteString(" W1(")
		spec.summary.WriteString(bestChoice.choiceName)
		spec.summary.WriteString(" ")
		spec.summary.WriteFloat64(bestChoice.accuracy1, 4)
		spec.summary.WriteString(" ")
		spec.summary.WriteFloat64(bestChoice.accuracy1Stat, 4)
		spec.summary.WriteString(") ")

		util.WriteStringToFile(spec.WeightFile1, bestChoice.pawnString)
	}

	weight2Opt, weight3Opt := spec.bestWeightChoiceExtended()

	if weight2Choice, hasWeight2 := weight2Opt.GetWithFlag(); hasWeight2 {
		str := tools.FormatWeight2String(weight2Choice.weight2)
		util.WriteStringToFile(spec.WeightFile2, str)

		spec.summary.WriteString("W2(")
		spec.summary.WriteString(weight2Choice.choiceName)
		spec.summary.WriteRune(' ')
		spec.summary.WriteFloat64(weight2Choice.accuracyXStat, 4)
		spec.summary.WriteString(") ")
	}

	if weight3Choice, hasWeight3 := weight3Opt.GetWithFlag(); hasWeight3 {
		str := tools.FormatWeight3String(weight3Choice.weight3)
		util.WriteStringToFile(spec.WeightFile3, str)

		spec.summary.WriteString("W3(")
		spec.summary.WriteString(weight3Choice.choiceName)
		spec.summary.WriteRune(' ')
		spec.summary.WriteFloat64(weight3Choice.accuracyXStat, 4)
		spec.summary.WriteString(") ")
	}

	spec.process.printer.PrintlnFromBuild(spec.summary)

	logText := spec.summary.Clone()
	logText.WriteRune('\n')
	logText.WriteString(time.Now().Format(time.DateTime))
	util.WriteStringToFile(spec.WeightFile1+"-accuracy.log", logText.String())

	spec.tabularReportWriteFile(spec.WeightFile1 + "-detail.log")
}

func (spec *WeightSpec) prepareSimData(tracker *util.TrackProgress, cancel util_async.CancelSignal) error {
	inputDataGrid, err := spec.prepareDataGrid(tracker, cancel)
	if err != nil {
		return err
	}

	inputDataReal, err2 := spec.prepareDataRandom(tracker, cancel)
	if err2 != nil {
		return err2
	}

	inputDataFit, err3 := spec.prepareDataFit(tracker, cancel)
	if err3 != nil {
		return err3
	}

	spec.dataAll = slices.Concat(inputDataGrid, inputDataReal, inputDataFit)
	return nil
}

func (spec *WeightSpec) prepareDataGrid(tracker *util.TrackProgress, cancel util_async.CancelSignal) ([]weight_types.WeightInput, error) {
	// READ IN ANY RECENT DATA
	tempPathGrid := files.TempData + "weightfind-sim-grid-" + spec.Label + ".json"
	inputDataGrid, dataAgeGrid := readWeightInputFile(tempPathGrid)
	// DO WE ACCEPT THE OLD DATA
	if dataAgeGrid > c_simDataAgeMax && !spec.process.forceSkipSim {
		inputDataGrid = nil
	}
	// SIMULATE STAT CHANGES, SAVE SIM DATA IN CASE WE NEED TO RESTART
	if inputDataGrid == nil {
		currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(spec.GearFile), &spec.Model, setup.MissingEnchant_Panic, spec.process.printer)
		currentItemSet := items.FullItemSet_FromMap(currentEquip)
		data, err := SimulateSteppedStatChangesForGrid(currentItemSet, spec.process.printer, spec.process.simSpeed,
			spec.Model.SimSpeedUp, spec.Model.StatsForWeighting, spec.Model.Spec, spec.Model.Goal, spec.Model.SimulateAs,
			spec.Model.Professions, tracker.NewChild(), spec.Label, cancel, spec.FixStatsMode)
		if err != nil {
			return nil, err
		}
		inputDataGrid = data
		weight_types.WeightInputWriteFile(inputDataGrid, tempPathGrid)
	} else {
		tracker.NewChild().SetDone()
	}
	spec.dataGrid = inputDataGrid
	return inputDataGrid, nil
}

func (spec *WeightSpec) prepareDataRandom(tracker *util.TrackProgress, cancel util_async.CancelSignal) ([]weight_types.WeightInput, error) {
	// READ IN ANY RECENT DATA
	tempPathReal := files.TempData + "weightfind-sim-real-" + spec.Label + ".json"
	inputDataReal, dataAgeReal := readWeightInputFile(tempPathReal)
	// DO WE ACCEPT THE OLD DATA
	if dataAgeReal > c_simDataAgeMax && !spec.process.forceSkipSim {
		inputDataReal = nil
	}
	// SIMULATE STAT CHANGES, SAVE SIM DATA IN CASE WE NEED TO RESTART
	if inputDataReal == nil {
		data, err := SimulateRealRandomSets(spec.GearFile, spec.SubstituteItems, &spec.Model, c_eachSimTargetGenerateDataCount,
			spec.process.simSpeed, spec.FixStatsMode, spec.process.printer, tracker.NewChild(), spec.Label, cancel)
		if err != nil {
			return nil, err
		}
		inputDataReal = data
		weight_types.WeightInputWriteFile(inputDataReal, tempPathReal)
	} else {
		tracker.NewChild().SetDone()
	}
	spec.dataRand = inputDataReal
	return inputDataReal, nil
}

func (spec *WeightSpec) prepareDataFit(tracker *util.TrackProgress, cancel util_async.CancelSignal) ([]weight_types.WeightInput, error) {
	// READ IN ANY RECENT DATA
	tempPathFit := files.TempData + "weightfind-sim-fit-" + spec.Label + ".json"
	inputDataFit, dataAgeFit := readWeightInputFile(tempPathFit)
	// DO WE ACCEPT THE OLD DATA
	if dataAgeFit > c_simDataAgeMax && !spec.process.forceSkipSim {
		inputDataFit = nil
	}
	// SIMULATE STAT CHANGES, SAVE SIM DATA IN CASE WE NEED TO RESTART
	if inputDataFit == nil {
		currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(spec.GearFile), &spec.Model, setup.MissingEnchant_Panic, spec.process.printer)
		currentItemSet := items.FullItemSet_FromMap(currentEquip)
		data, err := SimulateSteppedStatChangesForFitting(currentItemSet, spec.process.printer, spec.process.simSpeed,
			spec.Model.SimSpeedUp, spec.Model.StatsForWeighting, spec.Model.Spec, spec.Model.Goal, spec.Model.SimulateAs,
			spec.Model.Professions, tracker.NewChild(), spec.Label, cancel)
		if err != nil {
			return nil, err
		}
		inputDataFit = data
		weight_types.WeightInputWriteFile(inputDataFit, tempPathFit)
	} else {
		tracker.NewChild().SetDone()
	}
	spec.dataFit = inputDataFit
	return inputDataFit, nil
}

func (spec *WeightSpec) addChoice(choice weightChoice) {
	spec.choices = append(spec.choices, choice)
	spec.addToSummary(choice)
}

func (spec *WeightSpec) addToSummary(option weightChoice) {
	spec.summary.WriteString(option.choiceName)
	spec.summary.WriteString("=")
	spec.summary.WriteFloat64(option.accuracy1, 4)
	spec.summary.WriteString(" (")
	spec.summary.WriteFloat64(option.accuracy1Stat, 4)
	spec.summary.WriteString(") ")
}

func (spec *WeightSpec) loadOldWeights() {
	oldWeightBlock, _, oldWeightExists := tools.PawnWeightReadFile(spec.WeightFile1)
	if oldWeightExists {
		oldWeight := weight_types.Weight1Basic_FromBlock(oldWeightBlock)
		spec.evaluateWeight1("OLD", &oldWeight)
	}

	if spec.WeightFile2 == "" {
		spec.WeightFile2 = files.ToWeight2(spec.WeightFile1)
	}
	if weight2, weight2Found := tools.ReadWeight2File(spec.WeightFile2); weight2Found {
		spec.evaluateWeight2("OLD2", weight2)
	}

	if spec.WeightFile3 == "" {
		spec.WeightFile3 = files.ToWeight3(spec.WeightFile1)
	}
	if weight3, weight3Found := tools.ReadWeight3File(spec.WeightFile3); weight3Found {
		spec.evaluateWeight3("OLD3", weight3)
	}
}

func (spec *WeightSpec) solveGridWeights(gridOutlierSetting int, cancel util_async.CancelSignal) {
	gridData := spec.dataGrid
	if c_useSamplingGrid {
		gridData = util_collection.SliceSampleFromStart(gridData, c_dataSampleGrid)
	}

	choiceName := fmt.Sprintf("GRID%d", gridOutlierSetting)

	grid := weight_highs.GridStatWeightProcess1B{}
	grid.OUTLIER = gridOutlierSetting
	grid.SCALEMODE = 1
	grid.ROUNDMODE = 2
	grid.CALCMODE = 2
	grid.Init(spec.process.printer, spec.process.timeoutEach)
	grid.SetTargetRatios(spec.targetRatio)
	grid.SetRequiredStats(spec.statTypes)
	grid.SupplyData(gridData)
	weightsFuture, err := grid.Run()
	spec.handleFuture1OrError(choiceName, weightsFuture, cancel, err)
}

func (spec *WeightSpec) bestWeightChoice1() (weightChoice, bool) {
	best := util_rank.BestCollector1[weightChoice]{}
	for _, choice := range spec.choices {
		best.Offer(&choice, choice.accuracy1Stat)
		best.Offer(&choice, choice.accuracy1)
	}
	return best.GetBestOptional().GetWithFlag()
}

func (spec *WeightSpec) bestWeightChoiceExtended() (util_collection.Optional[weightChoice], util_collection.Optional[weightChoice]) {
	best2 := util_rank.BestCollector1[weightChoice]{}
	best3 := util_rank.BestCollector1[weightChoice]{}
	for _, choice := range spec.choices {
		weightOrig := choice.weightOrig
		switch weightCast := weightOrig.(type) {
		case *weight_types.Weight2Extended:
			choice.weight2 = weightCast
			acc2 := EvaluateAccuracyBasic(weightCast, spec.simTypes, &spec.targetRatio, spec.dataAll)
			acc2St := EvaluateAccuracyStatisticalExtended(weightCast, spec.simTypes, &spec.targetRatio, spec.dataAll)
			best2.Offer(&choice, acc2)
			best2.Offer(&choice, acc2St)
		case *weight_types.Weight3ExtendedRanged:
			choice.weight3 = weightCast
			weightConvert2 := weightCast.ConvertToWeight2(spec.dataAll)
			choice.weight2 = weightConvert2

			acc3 := EvaluateAccuracyBasic(weightCast, spec.simTypes, &spec.targetRatio, spec.dataAll)
			acc3St := EvaluateAccuracyStatisticalExtended(weightCast, spec.simTypes, &spec.targetRatio, spec.dataAll)
			best3.Offer(&choice, acc3)
			best3.Offer(&choice, acc3St)

			acc2 := EvaluateAccuracyBasic(weightConvert2, spec.simTypes, &spec.targetRatio, spec.dataAll)
			acc2St := EvaluateAccuracyStatisticalExtended(weightConvert2, spec.simTypes, &spec.targetRatio, spec.dataAll)
			best2.Offer(&choice, acc2)
			best2.Offer(&choice, acc2St)
		}
	}
	return best2.GetBestOptional(), best3.GetBestOptional()
}

func (spec *WeightSpec) solveRankingWeight(rankMode int, cancel util_async.CancelSignal) {
	rankData := spec.dataAll

	if rankMode == 0 {
		ranking := weight_highs.RankingStatWeightProcess3c{}
		ranking.Init(spec.process.printer, spec.process.timeoutEach)
		ranking.SetRequiredStats(spec.statTypes)
		ranking.SetTargetRatios(spec.targetRatio)
		ranking.SupplyData(rankData)
		weightsFuture, err := ranking.RunMultiRound()
		spec.handleFuture1OrError("RANK3C", weightsFuture, cancel, err)
	} else if rankMode == 1 {
		ranking := weight_highs.RankingStatWeightProcess3b{}
		ranking.TOTALWEIGHT = 2
		ranking.ALGO = 0
		ranking.Init(spec.process.printer, spec.process.timeoutEach)
		ranking.SetRequiredStats(spec.statTypes)
		ranking.SetTargetRatios(spec.targetRatio)
		ranking.SupplyData(rankData)
		var weightsFuture *util_async.FutureCancellable[weight_types.WeightResult1]
		var err error
		if bestWeightsSoFar, hasBest := spec.bestWeightChoice1(); hasBest {
			weightsFuture, err = ranking.RunSinglePassFromExternal(bestWeightsSoFar.weight)
		} else {
			weightsFuture, err = ranking.RunMultiRound()
		}
		spec.handleFuture1OrError("RANK3-2-0", weightsFuture, cancel, err)
	} else if rankMode == 2 {
		ranking := weight_highs.RankingStatWeightProcess3b{}
		ranking.TOTALWEIGHT = 2
		ranking.ALGO = 1
		ranking.Init(spec.process.printer, spec.process.timeoutEach)
		ranking.SetRequiredStats(spec.statTypes)
		ranking.SetTargetRatios(spec.targetRatio)
		ranking.SupplyData(rankData)
		var weightsFuture *util_async.FutureCancellable[weight_types.WeightResult1]
		var err error
		if bestWeightsSoFar, hasBest := spec.bestWeightChoice1(); hasBest {
			weightsFuture, err = ranking.RunSinglePassFromExternal(bestWeightsSoFar.weight)
		} else {
			weightsFuture, err = ranking.RunMultiRound()
		}
		spec.handleFuture1OrError("RANK3-2-1", weightsFuture, cancel, err)
	} else if rankMode == 3 {
		ranking := weight_highs.RankingStatWeightProcess{}
		ranking.RANKMODE = 0
		ranking.WEIGHTSUM = 0
		ranking.Init(spec.process.printer)
		ranking.SetRequiredStats(spec.statTypes)
		ranking.SetTargetRatios(spec.targetRatio)
		ranking.SupplyData(rankData)
		weightsFuture, err := ranking.Run(spec.process.timeoutEach)
		spec.handleFuture1OrError("RANK1-0", weightsFuture, cancel, err)
	} else if rankMode == 4 {
		ranking := weight_highs.RankingStatWeightProcess{}
		ranking.RANKMODE = 0
		ranking.WEIGHTSUM = 1
		ranking.Init(spec.process.printer)
		ranking.SetRequiredStats(spec.statTypes)
		ranking.SetTargetRatios(spec.targetRatio)
		ranking.SupplyData(rankData)
		weightsFuture, err := ranking.Run(spec.process.timeoutEach)
		spec.handleFuture1OrError("RANK1-1", weightsFuture, cancel, err)
	} else {
		if c_useSamplingRank4 {
			rankData = util_collection.SliceSampleRandom(rankData, c_dataSampleFitRank)
		}
		ranking := weight_highs.RankingStatWeightProcess4{}
		ranking.MULTIPLY = 0
		ranking.Init(spec.process.printer)
		ranking.SetRequiredStats(spec.statTypes)
		ranking.SetTargetRatios(spec.targetRatio)
		ranking.SupplyData(rankData)
		if bestWeightSoFar, hasBest := spec.bestWeightChoice1(); hasBest {
			existing1 := bestWeightSoFar.weight
			weightsFuture, err := ranking.RunUsingExternalStart(existing1, spec.process.timeoutEach)
			spec.handleFuture1OrError("RANK4", weightsFuture, cancel, err)
		}
	}
}

func (spec *WeightSpec) solveFormulaWeight(cancel util_async.CancelSignal) {
	comp := weight_highs.FormulaStatWeightProcess2{}
	comp.Init(spec.process.printer)
	comp.SetRequiredStats(spec.statTypes)
	comp.SetTargetRatios(spec.targetRatio)
	comp.SetMinimumIncludeRate(1.0)
	comp.SupplyData(spec.dataAll)
	weights2Future, err := comp.Run(spec.process.timeoutEach)
	spec.handleFuture2OrError("FORM2", weights2Future, cancel, err)

	compB := weight_highs.FormulaStatWeightProcess2{}
	compB.BLEND = 3
	compB.Init(spec.process.printer)
	compB.SetRequiredStats(spec.statTypes)
	compB.SetTargetRatios(spec.targetRatio)
	compB.SetMinimumIncludeRate(0.7)
	compB.SupplyData(spec.dataAll)
	weights2FutureB, err := compB.Run(spec.process.timeoutEach)
	spec.handleFuture2OrError("FORM2-70", weights2FutureB, cancel, err)
}

func (spec *WeightSpec) solveFittingWeight(cancel util_async.CancelSignal, tracker *util.TrackProgress) {
	comp := fitting3.FittingEachStatWeightProcess3{}
	comp.Init(3, spec.process.printer, spec.process.timeoutEach)
	comp.SetRequiredStats(spec.statTypes, spec.simTypes)
	comp.SetTargetRatios(spec.targetRatio)
	comp.SupplyData(spec.dataFit)
	weights3 := comp.Run(cancel, tracker)
	spec.evaluateWeightResult3("FITTING3", &weights3)
}

func (spec *WeightSpec) solveFittingFast(cancel util_async.CancelSignal) {
	for segments := 2; segments <= 4; segments++ {
		{
			fit4data := fitting4.FittingEachStatWeightProcess4{}
			fit4data.SegmentOnData = true
			fit4data.Init(segments, spec.process.printer, spec.process.timeoutEach)
			fit4data.SetRequiredStats(spec.statTypes, spec.simTypes)
			fit4data.SetTargetRatios(spec.targetRatio)
			fit4data.SupplyData(spec.dataAll)
			weights3data := fit4data.Run(cancel)
			label := fmt.Sprintf("FITTING4-data-%d", segments)
			spec.evaluateWeightResult3(label, &weights3data)
		}

		{
			fit4stat := fitting4.FittingEachStatWeightProcess4{}
			fit4stat.SegmentOnData = false
			fit4stat.Init(segments, spec.process.printer, spec.process.timeoutEach)
			fit4stat.SetRequiredStats(spec.statTypes, spec.simTypes)
			fit4stat.SetTargetRatios(spec.targetRatio)
			fit4stat.SupplyData(spec.dataAll)
			weights3stat := fit4stat.Run(cancel)
			label := fmt.Sprintf("FITTING4-stat-%d", segments)
			spec.evaluateWeightResult3(label, &weights3stat)
		}
	}
}

func (spec *WeightSpec) solveSearchWeights(searchMode int, cancel util_async.CancelSignal) {
	innerCancel := util_async.CancelSignal_Make()
	timer := util_async.CancelAfterTimeout(innerCancel, time.Second*time.Duration(spec.process.timeoutEach), spec.process.printer)
	_ = util_async.ChainCancel(cancel, innerCancel)
	defer timer.Stop()

	if searchMode == 0 {
		search := WeightSearcher2{}
		search.AccuracyStatistical = false
		search.Init(spec.statTypes, spec.targetRatio, spec.process.printer)
		search.SupplyData(spec.dataAll)
		search.SetRanges(-1.0, 10.0)
		weightResult := search.Run(innerCancel)
		spec.evaluateWeightResult1("SEARCH2", &weightResult)
	} else if searchMode == 1 {
		search := WeightSearcher3{}
		search.AccuracyStatistical = true
		search.Init(spec.statTypes, spec.targetRatio)
		search.SupplyData(spec.dataAll)
		search.SetRanges(-1.0, 10.0)
		weightResult := search.Run(innerCancel)
		spec.evaluateWeightResult1("SEARCH3", &weightResult)
	} else {
		sw := util.StopwatchMakeStarted()
		search := WeightSearcherExtended1{}
		search.Init(spec.statTypes, spec.targetRatio)
		search.SupplyData(spec.dataAll)
		search.SetRanges(-1.0, 5.0)
		weight2 := new(search.Run(innerCancel))

		weightResult := weight_types.WeightResult2Make(weight2, sw.Elapsed(), highs.ModelStatusOptimal)
		spec.evaluateWeightResult2("SEARCH-EX1", &weightResult)
	}
}

func (spec *WeightSpec) tweakEachWeight() {
	currentChoiceSize := len(spec.choices)
	for i := range currentChoiceSize {
		choice := spec.choices[i]
		if !choice.weight.IsEmpty() {
			weightsTweaked, _ := WeightTweakerWithLogging(choice.weight, spec.statTypes, &spec.targetRatio, spec.dataAll, spec.process.printer)
			spec.evaluateWeight1(choice.choiceName+"_TWEAK", &weightsTweaked)
		}
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

func (spec *WeightSpec) handleWeightError(choiceName string, err error) {
	spec.process.printer.Printf("Weights error %s %s NULL\n", spec.Label, err)
	spec.addChoice(weightChoice{choiceName: choiceName})
}

func (spec *WeightSpec) evaluateWeight1(choiceName string, weight1 *weight_types.Weight1Basic) {
	spec.evaluateWeightGeneric(choiceName, weight1, weight1, nil)
}

func (spec *WeightSpec) evaluateWeight2(choiceName string, weight2 *weight_types.Weight2Extended) {
	spec.evaluateWeightGeneric(choiceName, weight2.ConvertToWeight1(), weight2, nil)
}

func (spec *WeightSpec) evaluateWeight3(choiceName string, weight3 *weight_types.Weight3ExtendedRanged) {
	spec.evaluateWeightGeneric(choiceName, weight3.ConvertToWeight2(spec.dataAll).ConvertToWeight1(), weight3, nil)
}

func (spec *WeightSpec) evaluateWeightResult1(choiceName string, weightResult *weight_types.WeightResult1) {
	spec.evaluateWeightGeneric(choiceName, weightResult.Weight, weightResult.WeightInterface, &weightResult.WeightResultCommon)
}

func (spec *WeightSpec) evaluateWeightResult2(choiceName string, weightResult *weight_types.WeightResult2) {
	spec.evaluateWeightGeneric(choiceName, weightResult.AsWeight1(spec.dataAll), weightResult.WeightInterface, &weightResult.WeightResultCommon)
}

func (spec *WeightSpec) evaluateWeightResult3(choiceName string, weightResult *weight_types.WeightResult3) {
	spec.evaluateWeightGeneric(choiceName, weightResult.AsWeight1(spec.dataAll), weightResult.WeightInterface, &weightResult.WeightResultCommon)
}

func (spec *WeightSpec) evaluateWeightResultFuture1(choiceName string, futureResult *util_async.FutureCancellable[weight_types.WeightResult1]) {
	if weightResult, hasResult := futureResult.WaitForResult(); hasResult {
		spec.evaluateWeightResult1(choiceName, &weightResult)
	}
}

func (spec *WeightSpec) evaluateWeightResultFuture2(choiceName string, futureResult *util_async.FutureCancellable[weight_types.WeightResult2]) {
	if weightResult, hasResult := futureResult.WaitForResult(); hasResult {
		spec.evaluateWeightResult2(choiceName, &weightResult)
	}
}

//func (spec *WeightSpec) evaluateWeightResultFuture3(choiceName string, futureResult *util_async.FutureCancellable[weight_types.WeightResult3]) {
//	if weightResult, hasResult := futureResult.WaitForResult(); hasResult {
//		spec.evaluateWeightResult3(choiceName, &weightResult)
//	}
//}

func (spec *WeightSpec) evaluateWeightGeneric(choiceName string, weight1 *weight_types.Weight1Basic, weightOrig weight_types.IWeight, weightResult *weight_types.WeightResultCommon) {
	if weight1 == nil || weightOrig == nil {
		spec.process.printer.Printf("Weights accuracy %s %s NULL\n", spec.Label, choiceName)
		spec.addChoice(weightChoice{choiceName: choiceName, weightResult: weightResult})
		return
	}

	var accuracyX, accuracyXStat float64
	var hadExtended bool
	if _, isOne := weightOrig.(*weight_types.Weight1Basic); isOne {
		hadExtended = false
	} else {
		accuracyX = EvaluateAccuracyBasic(weightOrig, spec.simTypes, &spec.targetRatio, spec.dataAll)
		accuracyXStat = EvaluateAccuracyStatisticalExtended(weightOrig, spec.simTypes, &spec.targetRatio, spec.dataAll)
		hadExtended = true
	}
	accuracy1 := EvaluateAccuracyBasic(weight1, spec.simTypes, &spec.targetRatio, spec.dataAll)
	accuracy1Stat := EvaluateAccuracyStatisticalExtended(weight1, spec.simTypes, &spec.targetRatio, spec.dataAll)

	pawnString := tools.WritePawnString(*weight1, spec.process.printer)
	spec.process.printer.Println(weightOrig.String())
	tools.WriteWeightString(weightOrig, spec.process.printer)

	if weight1.IsEmpty() || weightOrig.IsEmpty() {
		spec.process.printer.Printf("Weights accuracy %s %s EMPTY a1=%f a1s=%f aX=%f aXs=%f\n", spec.Label, choiceName, accuracy1, accuracy1Stat, accuracyX, accuracyXStat)
		spec.addChoice(weightChoice{choiceName: choiceName, weight: *weight1, pawnString: pawnString, weightResult: weightResult})
	} else if weight1.IsOverlySimple() {
		spec.process.printer.Printf("Weights accuracy %s %s OVERLY SIMPLE a1=%f a1s=%f aX=%f aXs=%f\n", spec.Label, choiceName, accuracy1, accuracy1Stat, accuracyX, accuracyXStat)
		spec.addChoice(weightChoice{choiceName: choiceName, weight: *weight1, pawnString: pawnString, weightResult: weightResult})
	} else {
		spec.process.printer.Printf("Weights accuracy %s %s a1=%f a1s=%f aX=%f aXs=%f\n", spec.Label, choiceName, accuracy1, accuracy1Stat, accuracyX, accuracyXStat)
		spec.addChoice(weightChoice{choiceName, *weight1, nil, nil, hadExtended,
			accuracy1, accuracy1Stat,
			accuracyX, accuracyXStat,
			pawnString, weightResult, weightOrig})
	}
}

func (spec *WeightSpec) handleFuture1OrError(choiceName string, weightsFuture *util_async.FutureCancellable[weight_types.WeightResult1], cancel util_async.CancelSignal, err error) {
	if err != nil {
		spec.handleWeightError(choiceName, err)
		return
	}

	err = util_async.ChainCancel(cancel, weightsFuture)
	if err != nil {
		spec.handleWeightError(choiceName, err)
		return
	}

	spec.evaluateWeightResultFuture1(choiceName, weightsFuture)
}

func (spec *WeightSpec) handleFuture2OrError(choiceName string, weightsFuture *util_async.FutureCancellable[weight_types.WeightResult2], cancel util_async.CancelSignal, err error) {
	if err != nil {
		spec.handleWeightError(choiceName, err)
		return
	}

	err = util_async.ChainCancel(cancel, weightsFuture)
	if err != nil {
		spec.handleWeightError(choiceName, err)
		return
	}

	spec.evaluateWeightResultFuture2(choiceName, weightsFuture)
}
