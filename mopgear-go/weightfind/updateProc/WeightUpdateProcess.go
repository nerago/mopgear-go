package updateProc

import (
	"cmp"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
)

const (
	c_simDataAgeMax = 48 * time.Hour
	//c_updateThreadCount = 3
	c_ratioThreadCount = 12

	c_eachSimTargetGenerateDataCount = 600

	c_dataSampleFitRank  = 300
	c_dataSampleGrid     = 96
	c_useSamplingRank4   = true
	c_useSamplingGrid1   = false
	c_useSamplingFormMIP = true
)

type WeightUpdateProcess struct {
	simSpeed     simulate.WowSim_RunSize
	forceSkipSim bool
	skipSolve    bool
	timeoutEach  int
	printer      *util.PrintRecorder
	specs        []*weightSpecInternal
}

func (wup *WeightUpdateProcess) Init(simSpeed simulate.WowSim_RunSize, forceSkipSim bool, skipSolve bool, timeoutEach int, printer *util.PrintRecorder) {
	wup.simSpeed = simSpeed
	wup.forceSkipSim = forceSkipSim
	wup.skipSolve = skipSolve
	wup.timeoutEach = timeoutEach
	wup.printer = printer
}

func (wup *WeightUpdateProcess) AddSpecParam(param SpecParam) {
	spec := &weightSpecInternal{
		param: param,
		inputs: updateInputs{
			simTypes:    param.Model.SimPriority.SimTypes(),
			statTypes:   param.Model.StatsForWeighting,
			targetRatio: param.Model.SimPriority,
		},
		out: choiceOutput{
			printer: wup.printer,
		},
		process: wup,
	}
	wup.specs = append(wup.specs, spec)
}

func (wup *WeightUpdateProcess) Run(cancel util_async.CancelSignal, outerThreads int) {
	progress := util.TrackProgress_Start()
	progress.RunOuterTracking(len(wup.specs))
	defer progress.SetDone()

	summaries, err := util_async.Map_SliceToSlice_Cancellable_PassError(outerThreads, wup.specs, cancel, func(spec **weightSpecInternal) (string, error) {
		return (*spec).updateSpec(progress.NewChild(), cancel)
	})

	if err != nil {
		panic(err)
	}

	for _, summary := range summaries {
		wup.printer.Println(summary)
	}

	for _, spec := range wup.specs {
		spec.tabularReport(wup.printer)
	}
}

func (spec *weightSpecInternal) tabularReportWriteFile(filename string) {
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

func (spec *weightSpecInternal) tabularReport(print util.Printable) {
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

	choices := spec.out.getChoices()
	slices.SortFunc(choices, func(a, b weightChoice) int {
		return cmp.Compare(max(a.accuracy1, a.accuracy1Stat), max(b.accuracy1, b.accuracy1Stat))
	})
	for choice := range util_collection.ForPointer(choices) {
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

	print.Printf("TABLE %s\n", spec.param.Label)
	tab.Write(print)
}

func (spec *weightSpecInternal) updateSpec(tracker *util.TrackProgress, cancel util_async.CancelSignal) (string, error) {
	// each simulator process is considered 1/4, fitting is 1/4, then remaining solving is remaining.
	tracker.RunOuterTracking(5)
	defer tracker.SetDone()

	// READ OLD DATA AND/OR RUN SIM
	err := spec.prepareSimData(tracker, cancel)
	if err != nil {
		return "", err
	}

	// START BUILDING REPORT
	spec.out = choiceOutput{
		label:   spec.param.Label,
		input:   &spec.inputs,
		printer: spec.process.printer,
	}
	spec.out.startReport()

	// LOAD OLD WEIGHT FILES
	loadOldWeights(&spec.param, &spec.out)

	// RUN RATING SOLVERS
	if !spec.process.skipSolve {
		solve := solves{
			cancel:      cancel,
			tracker:     tracker,
			timeoutEach: spec.process.timeoutEach,
			printer:     spec.process.printer,
			input:       &spec.inputs,
			output:      &spec.out,
		}
		solve.runSolvers()

		spec.tweakEachWeight()
	}

	summary := spec.reportAndWriteWeights()

	spec.inputs.freeData()

	return summary, nil
}

func (spec *weightSpecInternal) reportAndWriteWeights() string {
	summary := spec.out.getSummary()

	// OVERWRITE WEIGHT FILE
	if bestChoice, hasBest := spec.out.bestWeightChoice1(); hasBest {
		summary.WriteString(" ::::: ")
		summary.WriteString(" W1(")
		summary.WriteString(bestChoice.choiceName)
		summary.WriteString(" ")
		summary.WriteFloat64(bestChoice.accuracy1, 4)
		summary.WriteString(" ")
		summary.WriteFloat64(bestChoice.accuracy1Stat, 4)
		summary.WriteString(") ")

		util.WriteStringToFile(spec.param.WeightFile1, bestChoice.pawnString)
	}

	weight2Opt, weight3Opt := spec.out.bestWeightChoiceExtended()

	if weight2Choice, hasWeight2 := weight2Opt.GetWithFlag(); hasWeight2 {
		str := tools.FormatWeight2String(weight2Choice.weight2)
		util.WriteStringToFile(spec.param.WeightFile2, str)

		summary.WriteString("W2(")
		summary.WriteString(weight2Choice.choiceName)
		summary.WriteRune(' ')
		summary.WriteFloat64(weight2Choice.accuracyXStat, 4)
		summary.WriteString(") ")
	}

	if weight3Choice, hasWeight3 := weight3Opt.GetWithFlag(); hasWeight3 {
		str := tools.FormatWeight3String(weight3Choice.weight3)
		util.WriteStringToFile(spec.param.WeightFile3, str)

		summary.WriteString("W3(")
		summary.WriteString(weight3Choice.choiceName)
		summary.WriteRune(' ')
		summary.WriteFloat64(weight3Choice.accuracyXStat, 4)
		summary.WriteString(") ")
	}

	spec.process.printer.PrintlnFromBuild(summary)

	logText := summary.Clone()
	logText.WriteRune('\n')
	logText.WriteString(time.Now().Format(time.DateTime))
	util.WriteStringToFile(spec.param.WeightFile1+"-accuracy.log", logText.String())

	spec.tabularReportWriteFile(spec.param.WeightFile1 + "-detail.log")

	return summary.String()
}
