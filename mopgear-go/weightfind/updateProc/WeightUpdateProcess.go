package updateProc

import (
	"fmt"
	"time"

	"github.com/nerago/mopgear-go/gear_model/model_factory"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
)

const (
	c_simDataAgeMax     = 48 * time.Hour
	c_updateThreadCount = 6
	c_ratioThreadCount  = 12

	c_eachSimTargetGenerateDataCount = 600
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
		label: param.Model.Label,
		inputs: updateInputs{
			simTypes:    param.Model.SimPriority.SimTypes(),
			statTypes:   param.Model.StatsForWeighting,
			targetRatio: param.Model.SimPriority,
		},
		process: wup,
	}
	if param.LabelOverride != "" {
		spec.label = param.LabelOverride
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
		wup.printer.Println("UPDATE PROC ERROR: " + err.Error())
	}

	for _, summary := range summaries {
		wup.printer.Println(summary)
	}

	for _, spec := range wup.specs {
		spec.tabularReport(wup.printer)
	}
}

func (spec *weightSpecInternal) updateSpec(taskPoolSim, taskPoolSolve *util_async.NestedTaskPoolChild, tracker *util.TrackProgress, cancel util_async.CancelSignal) (string, error) {
	expectedCount := 0
	if !spec.process.forceSkipSim {
		expectedCount += 3
	}
	if !spec.process.skipSolve {
		expectedCount++
	}
	tracker.RunOuterTracking(expectedCount)
	defer tracker.SetDone()

	// READ OLD DATA AND/OR RUN SIM
	if err := spec.prepareSimData(taskPoolSim, tracker, cancel); err != nil {
		return "", err
	}

	// INIT
	if err := model_factory.SetupModelWeightsHaveSamples(&spec.param.Model, spec.inputs.dataAll); err != nil {
		return "", err
	}

	// START BUILDING REPORT
	spec.out = choiceOutput{
		label:   spec.label,
		input:   &spec.inputs,
		printer: spec.process.printer,
	}
	spec.out.startReport()
	spec.out.initAccuracy()

	// LOAD OLD WEIGHT FILES
	loadOldWeights(&spec.param, &spec.out)

	// RUN RATING SOLVERS
	if !spec.process.skipSolve {
		solve := solves{
			cancel:      cancel,
			tracker:     tracker.NewChild(),
			taskPool:    taskPoolSolve,
			timeoutEach: spec.process.timeoutEach,
			printer:     spec.process.printer,
			input:       &spec.inputs,
			output:      &spec.out,
		}
		err := solve.startSolvers()
		if err != nil {
			util.GlobalWarnHandler(fmt.Errorf("ERROR IN SOLVERS: %w", err))
		}
	}

	// REPORTING
	summary := spec.reportAndWriteWeights()

	spec.inputs.freeData()
	return summary, nil
}
