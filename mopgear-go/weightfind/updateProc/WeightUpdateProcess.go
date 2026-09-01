package updateProc

import (
	"time"

	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
)

const (
	c_simDataAgeMax     = 48 * time.Hour
	c_updateThreadCount = 6
	c_ratioThreadCount  = 12

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
		process: wup,
	}
	wup.specs = append(wup.specs, spec)
}

func (wup *WeightUpdateProcess) Run(cancel util_async.CancelSignal, outerThreads int) {
	progress := util.TrackProgress_Start()
	progress.RunOuterTracking(len(wup.specs))
	defer progress.SetDone()

	taskPool := util_async.NestedTaskPoolParent{}
	taskPool.Start(c_updateThreadCount)
	defer taskPool.Stop()

	summaries, err := util_async.Map_SliceToSlice_Cancellable_PassError(outerThreads, wup.specs, cancel, func(spec **weightSpecInternal) (string, error) {
		return (*spec).updateSpec(taskPool.NewChild(), taskPool.NewChild(), progress.NewChild(), cancel)
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

func (spec *weightSpecInternal) updateSpec(taskPoolSim, taskPoolSolve *util_async.NestedTaskPoolChild, tracker *util.TrackProgress, cancel util_async.CancelSignal) (string, error) {
	// each simulator process is considered 1/4, fitting is 1/4, then remaining solving is remaining.
	tracker.RunOuterTracking(5)
	defer tracker.SetDone()

	// READ OLD DATA AND/OR RUN SIM
	if err := spec.prepareSimData(taskPoolSim, tracker, cancel); err != nil {
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
			taskPool:    taskPoolSolve,
			timeoutEach: spec.process.timeoutEach,
			printer:     spec.process.printer,
			input:       &spec.inputs,
			output:      &spec.out,
		}
		solve.startSolvers()
	}

	if err := taskPoolSolve.WaitAllComplete(); err != nil {
		return "", err
	}

	// REPORTING
	summary := spec.reportAndWriteWeights()

	spec.inputs.freeData()
	return summary, nil
}
