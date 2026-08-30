package util_highs

import (
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/bartolsthoorn/gohighs/highs"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
)

var g_gpuMutex sync.Mutex

type linearRun struct {
	build        *LinearBuilder
	solver       *highs.Solver
	logFilename  string
	requireGpu   bool
	optionalGpu  bool
	log          *util.PrintRecorder
	stopwatch    *util.Stopwatch
	timeoutToken *TimeLimitToken
	future       *util_async.FutureCancellable[LinearResult]
}

func (build *LinearBuilder) RunHighsFuture(stopwatch *util.Stopwatch) *util_async.FutureCancellable[LinearResult] {
	build.vars.validate()

	run, err := build.prepareHighsRun()
	if err != nil {
		return run.makeImmediateError(err)
	}

	bailout := run.makeFuture()
	if bailout {
		return run.future
	}

	if stopwatch == nil {
		stopwatch = util.StopwatchMakeStopped()
	}
	run.stopwatch = stopwatch

	go run.runSolveThread()

	return run.future
}

func (build *LinearBuilder) RunHighsFuture2(timeout *TimeLimitToken) *util_async.FutureCancellable[LinearResult] {
	build.vars.validate()
	timeout.linearBuildSetTimeout(build)

	run, err := build.prepareHighsRun()
	if err != nil {
		return run.makeImmediateError(err)
	}

	bailout := run.makeFuture()
	if bailout {
		return run.future
	}

	run.timeoutToken = timeout

	go run.runSolveThread()

	return run.future
}

func (build *LinearBuilder) prepareHighsRun() (*linearRun, error) {
	run := &linearRun{
		build: build,
		log:   util.PrintRecorder_HoldAll(),
	}

	if solver, err := g_HighsPool.Get(); err == nil {
		run.solver = solver
	} else {
		return run, err
	}

	if err := run.configureHighsMatrix(); err != nil {
		return run, err
	}

	if logFilename, err := makeTempFilename(); err == nil {
		run.logFilename = logFilename
	} else {
		return run, err
	}

	if err := run.configureHighsUtil(); err != nil {
		return run, err
	}

	if err := run.configureHighsSolver(); err != nil {
		return run, err
	}

	return run, nil
}

func (run *linearRun) runSolveThread() {
	solution, errSolve := run.runInner()

	runTime := run.solver.GetRunTime()
	if run.stopwatch != nil {
		run.stopwatch.AddElapsedDuration(runTime)
	}
	if run.timeoutToken != nil {
		run.timeoutToken.addSolveRunTime(runTime)
	}

	if errSolve == nil {
		run.onGoodRun(solution, runTime)
	} else {
		run.onBadRun(solution, errSolve, runTime)
	}
}

func (run *linearRun) runInner() (*highs.Solution, error) {
	if run.requireGpu {
		g_gpuMutex.Lock()
		defer g_gpuMutex.Unlock()
	} else if run.optionalGpu {
		if g_gpuMutex.TryLock() {
			defer g_gpuMutex.Unlock()
		} else {
			err := run.solver.SetStringOption("solver", "choose")
			if err != nil {
				return nil, err
			}
		}
	}

	solution, err := run.solver.Run()
	return solution, err
}

//goland:noinspection GoBoolExpressions
func (run *linearRun) onGoodRun(solution *highs.Solution, runTime time.Duration) {
	if C_DebugHighs && C_DiagnoseInfeasible && solution.Status == highs.ModelStatusInfeasible {
		diagnoseInfeasible(run.build, run.log)
	}

	err := run.postRunCleanup()

	run.sendResult(solution, runTime, err)

	g_HighsPool.Put(run.solver)
}

func (run *linearRun) onBadRun(solution *highs.Solution, errSolve error, runTime time.Duration) {
	err := errors.Join(errSolve, run.postRunCleanup())

	run.sendResult(solution, runTime, err)

	// don't return solver to pool since in unknown state after the errors
}

func (run *linearRun) sendResult(solution *highs.Solution, runTime time.Duration, err error) {
	errHandling := run.future.SetResult(LinearResult{
		solution: solution,
		build:    run.build,
		log:      run.log,
		elapsed:  runTime,
		err:      err,
	})

	if errHandling != nil {
		util.GlobalErrorHandler(errors.Join(err, errHandling))
	}
}

func (run *linearRun) makeFuture() bool {
	run.future = util_async.FutureCancellable_Make[LinearResult]()

	err := run.future.AddCancelHandler(func() error {
		return run.solver.InterruptSetFlag(true)
	})

	if err != nil {
		run.sendEarlyError(err, run.future)
		return true
	}

	return false
}

func (run *linearRun) makeImmediateError(err error) *util_async.FutureCancellable[LinearResult] {
	future := util_async.FutureCancellable_Make[LinearResult]()
	run.sendEarlyError(err, future)
	return future
}

func (run *linearRun) sendEarlyError(err error, future *util_async.FutureCancellable[LinearResult]) {
	if run.solver != nil {
		err = errors.Join(err, run.postRunCleanup())
		g_HighsPool.Put(run.solver)
	}

	errInErrorHandling := future.SetResult(LinearResult{build: run.build, err: err})
	if errInErrorHandling != nil {
		util.GlobalErrorHandler(errors.Join(err, errInErrorHandling))
	}
}

func (run *linearRun) postRunCleanup() error {
	var err error

	if run.logFilename != "" {
		err = errors.Join(
			run.solver.SetStringOption("log_file", ""), // flush log
			readLogfile(run.logFilename, run.log),
		)
	}

	if run.build.Callback != nil {
		err = errors.Join(err, run.solver.ClearCallback())
	} else {
		err = errors.Join(err, run.solver.InterruptSupportDisable())
	}

	return errors.Join(
		err,
		run.solver.ClearLinearObjectives(),
		run.solver.ClearModel(),
		run.solver.ClearSolver(),
		run.solver.ClearClock(),
		run.solver.Clear(),
	)
}

func (run *linearRun) configureHighsMatrix() error {
	build := run.build
	solver := run.solver

	numRows, lowerBound, upperBound, startArray, indexArray, valuesArray := build.mat.createSolverInputArrays()

	var colTypes []highs.VariableType
	if build.isMIP() {
		colTypes = run.build.vars.colTypes
	}

	err := solver.PassModel(
		len(build.vars.colCosts),
		numRows,
		build.vars.colCosts, build.vars.colLower, build.vars.colUpper,
		lowerBound, upperBound,
		startArray, indexArray, valuesArray,
		colTypes, !build.Minimise, 0)

	err = errors.Join(err, solver.ClearLinearObjectives())
	for linearObjectiveIndex := range build.vars.objectives {
		objective := &build.vars.objectives[linearObjectiveIndex]
		coefficientArray := make([]float64, len(build.vars.colTypes))
		for _, entry := range objective.coefficientEntries {
			coefficientArray[entry.columnNumber] = entry.value
		}
		err = errors.Join(err, solver.AddLinearObjective(objective.weight, objective.offset, coefficientArray, objective.abs_tolerance, objective.rel_tolerance, objective.priority))
	}
	err = errors.Join(err, solver.SetBoolOption("blend_multi_objectives", build.BlendMultiObjectives))

	if len(build.vars.partialSolution) > 0 {
		indexArray := make([]int32, len(build.vars.partialSolution))
		valueArray := make([]float64, len(build.vars.partialSolution))
		index := 0
		for columnNumber, value := range build.vars.partialSolution {
			indexArray[index] = int32(columnNumber)
			valueArray[index] = value
			index++
		}
		err = errors.Join(err, solver.SetSparseSolution(indexArray, valueArray))
	}

	return err
}

func (run *linearRun) configureHighsUtil() (err error) {
	build := run.build
	solver := run.solver

	for name, value := range build.FloatOptions {
		err = errors.Join(err, solver.SetFloatOption(name, value))
	}
	for name, value := range build.StringOptions {
		err = errors.Join(err, solver.SetStringOption(name, value))
	}
	for name, value := range build.IntOptions {
		err = errors.Join(err, solver.SetIntOption(name, value))
	}
	for name, value := range build.BoolOptions {
		err = errors.Join(err, solver.SetBoolOption(name, value))
	}

	if build.TimeLimitSeconds != 0 {
		err = errors.Join(err, solver.SetFloatOption("time_limit", float64(build.TimeLimitSeconds)))
	} else {
		err = errors.Join(err, solver.SetFloatOption("time_limit", InfPos()))
	}

	if build.Callback != nil {
		// TODO move this into gohighs
		interruptTypes := []highs.CallbackType{
			highs.CallbackTypeSimplexInterrupt,
			highs.CallbackTypeIpmInterrupt,
			highs.CallbackTypeMipInterrupt,
		}
		callbackTypes := slices.Concat(build.CallbackTypes, interruptTypes)
		util_collection.RemoveDuplicatesComparable_InPlace(&callbackTypes)
		err = errors.Join(err, solver.SetCallback(func(callbackType highs.CallbackType, str string, out highs.CallbackData) highs.CallbackResult {
			result := build.Callback(callbackType, str, out)
			if slices.Contains(interruptTypes, callbackType) {
				result.UserInterrupt = solver.Interrupted
			}
			return result
		}, build.CallbackTypes))
	} else {
		err = errors.Join(err, solver.InterruptSupportEnable())
	}

	err = errors.Join(err, solver.SetStringOption("log_file", run.logFilename))
	err = errors.Join(err, solver.SetBoolOption("log_to_console", (C_DebugHighs || C_HighsToConsole) && !build.NoOutput))
	if C_DebugHighs {
		err = errors.Join(err, solver.SetIntOption("log_dev_level", 2))
	} else {
		err = errors.Join(err, solver.SetIntOption("log_dev_level", 0))
	}

	if build.DisablePreSolve {
		err = errors.Join(err, solver.SetStringOption("presolve", "off"))
	} else {
		err = errors.Join(err, solver.SetStringOption("presolve", "on"))
	}

	return err
}

func (run *linearRun) configureHighsSolver() (err error) {
	solver := run.solver
	expectMip := false

	switch run.build.Solver {
	case Solver_LP_USE_GPU:
		err = errors.Join(err, solver.SetStringOption("solver", "hipdlp"))
		run.requireGpu = true
	case Solver_LP_GPU_IF_FREE:
		err = errors.Join(err, solver.SetStringOption("solver", "hipdlp"))
		run.optionalGpu = true
	case Solver_LP_NO_GPU:
		if run.build.isLargeModel() {
			err = errors.Join(err, solver.SetStringOption("solver", "hipo"))
		} else {
			err = errors.Join(err, solver.SetStringOption("solver", "choose"))
		}
	case Solver_Force_HIPO:
		err = errors.Join(err, solver.SetStringOption("solver", "hipo"))
	case Solver_Force_Simplex:
		err = errors.Join(err, solver.SetStringOption("solver", "simplex"))
	case Solver_Force_IPX:
		err = errors.Join(err, solver.SetStringOption("solver", "ipx"))
	case Solver_MIP_Interior:
		if run.build.isLargeModel() {
			err = errors.Join(
				err,
				solver.SetStringOption("solver", "choose"),
				solver.SetStringOption("mip_lp_solver", "hipo"),
				solver.SetStringOption("mip_ipm_solver", "hipo"),
			)
		} else {
			err = errors.Join(
				err,
				solver.SetStringOption("solver", "choose"),
				solver.SetStringOption("mip_lp_solver", "choose"),
				solver.SetStringOption("mip_ipm_solver", "choose"),
			)
		}
		expectMip = true
	case Solver_MIP_Vertex:
		err = errors.Join(
			err,
			solver.SetStringOption("solver", "choose"),
			solver.SetStringOption("mip_lp_solver", "choose"),
			solver.SetStringOption("mip_ipm_solver", "choose"),
		)
		expectMip = true
	case Solver_Flexible:
		err = errors.Join(
			err,
			solver.SetStringOption("solver", "choose"),
			solver.SetStringOption("mip_lp_solver", "choose"),
			solver.SetStringOption("mip_ipm_solver", "choose"),
		)
		return err
	default:
		err = errors.Join(err, errors.New("solver not specified"))
	}

	if expectMip != run.build.isMIP() {
		err = errors.Join(err, errors.New("solver wrong for MIP/non-MIP model"))
	}

	return err
}
