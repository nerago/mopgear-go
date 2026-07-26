package util_highs

import (
	"fmt"
	"math"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

type LinearBuilder struct {
	vars                 variableArrayBuilder
	mat                  constraintMatrixBuilder
	NoOutput             bool
	Minimise             bool
	BlendMultiObjectives bool
	Solver               SolverMode
	DisablePreSolve      bool
	TimeLimitSeconds     int
}

func (build *LinearBuilder) Clone() *LinearBuilder {
	return &LinearBuilder{
		vars:                 build.vars.clone(),
		mat:                  build.mat.clone(),
		NoOutput:             build.NoOutput,
		Minimise:             build.Minimise,
		BlendMultiObjectives: build.BlendMultiObjectives,
		Solver:               build.Solver,
		DisablePreSolve:      build.DisablePreSolve,
		TimeLimitSeconds:     build.TimeLimitSeconds,
	}
}

func (build *LinearBuilder) AddObjectiveBlended(weight float64, offset float64) ObjectiveIndex {
	if !build.BlendMultiObjectives {
		panic("wrong linear type")
	}
	return build.vars.addObjective(weight, offset, -1, -1, -1)
}

func (build *LinearBuilder) AddObjectivePrioritised(maximise bool, abs_tolerance float64, rel_tolerance float64, priority int) ObjectiveIndex {
	if build.BlendMultiObjectives {
		panic("wrong linear type")
	}
	// lp.sense_ = linear_objective.weight > 0 ? ObjSense::kMinimize : ObjSense::kMaximize;
	if maximise {
		return build.vars.addObjective(-1, 0, abs_tolerance, rel_tolerance, priority)
	} else {
		return build.vars.addObjective(1, 0, abs_tolerance, rel_tolerance, priority)
	}
}

func (build *LinearBuilder) CreateColumnBool(debug DebugContext) ColumnIndex {
	return build.vars.create(highs.Integer, 0, 1, 0, debug)
}

func (build *LinearBuilder) CreateColumnBoolWithOutput(cost float64, debug DebugContext) ColumnIndex {
	return build.vars.create(highs.Integer, 0, 1, cost, debug)
}

func (build *LinearBuilder) CreateColumnBoolWithObjective(cost float64, objectiveIndex ObjectiveIndex, debug DebugContext) ColumnIndex {
	return build.vars.createForLinear(highs.Integer, 0, 1, cost, objectiveIndex, debug)
}

func (build *LinearBuilder) CreateColumnGeneral(varType highs.VariableType, lower, upper float64, debug DebugContext) ColumnIndex {
	return build.vars.create(varType, lower, upper, 0, debug)
}

func (build *LinearBuilder) CreateColumnWithOutput(varType highs.VariableType, lower, upper, cost float64, debug DebugContext) ColumnIndex {
	return build.vars.create(varType, lower, upper, cost, debug)
}

func (build *LinearBuilder) CreateColumnWithObjective(varType highs.VariableType, lower, upper, cost float64, objectiveIndex ObjectiveIndex, debug DebugContext) ColumnIndex {
	return build.vars.createForLinear(varType, lower, upper, cost, objectiveIndex, debug)
}

func (build *LinearBuilder) ClearInitialSolutionValue() {
	clear(build.vars.partialSolution)
}

func (build *LinearBuilder) SetInitialSolutionValue(columnNumber ColumnIndex, value float64) {
	if build.vars.partialSolution == nil {
		build.vars.partialSolution = make(map[ColumnIndex]float64)
	}
	if 0 <= columnNumber && int(columnNumber) < len(build.vars.colLower) {
		build.vars.partialSolution[columnNumber] = value
	} else {
		panic("invalid column number")
	}
}

func (build *LinearBuilder) GetInitialSolutionValue(columnNumber ColumnIndex) float64 {
	value, hasValue := build.vars.partialSolution[columnNumber]
	if !hasValue {
		panic("initial not set")
	}
	return value
}

func (build *LinearBuilder) RunHighsFuture(stopwatch *util.Stopwatch) *util_async.FutureCancellable[LinearResult] {
	solver, logFilename, requestGpu := build.prepareHighsRun(true)

	future := util_async.FutureCancellable_Make[LinearResult]()
	future.AddCancelHandler(func() {
		verifyNoError(solver.InterruptSetFlag(1))
	})

	if stopwatch == nil {
		stopwatch = util.StopwatchMakeStopped()
	}

	go func() {
		log := util.PrintRecorder_HoldAll()

		solution, err := G_HighsPool.RunSolverUnderMutex(solver, requestGpu, stopwatch)
		verifyNoError(err)

		if C_DebugHighs && C_DiagnoseInfeasible && solution.Status == highs.ModelStatusInfeasible {
			diagnoseInfeasible(build, log)
		}

		build.postHighsRun(solver, logFilename, log)
		future.SetResult(LinearResult{solution, build, log})

		G_HighsPool.Put(solver)
	}()

	return future
}

func (build *LinearBuilder) prepareHighsRun(needLog bool) (*highs.Solver, string, bool) {
	solver := G_HighsPool.Get()

	build.configureHighsMatrix(solver)

	logFilename := ""
	if needLog {
		logFilename = makeTempFilename()
	}
	build.configureHighsUtil(solver, logFilename)

	requestGpu := build.configureHighsSolver(solver)

	return solver, logFilename, requestGpu
}

func (*LinearBuilder) postHighsRun(solver *highs.Solver, logFilename string, printer *util.PrintRecorder) {
	if logFilename != "" {
		verifyNoError(solver.SetStringOption("log_file", "")) // flush log
		readLogfile(logFilename, printer)
	}

	verifyNoError(solver.InterruptSupportDisable())
	verifyNoError(solver.Clear())
}

func (build *LinearBuilder) configureHighsMatrix(solver *highs.Solver) {
	numRows, lowerBound, upperBound, startArray, indexArray, valuesArray := build.mat.createSolverInputArrays()
	verifyNoError(solver.PassModel2(
		int32(len(build.vars.colTypes)),
		numRows,
		build.vars.colCosts, build.vars.colLower, build.vars.colUpper,
		lowerBound, upperBound,
		startArray, indexArray, valuesArray,
		build.vars.colTypes, !build.Minimise, 0))

	verifyNoError(solver.ClearLinearObjectives())
	for linearObjectiveIndex := range build.vars.objectives {
		objective := &build.vars.objectives[linearObjectiveIndex]
		coefficientArray := make([]float64, len(build.vars.colTypes))
		for _, entry := range objective.coefficientEntries {
			coefficientArray[entry.columnNumber] = entry.value
		}
		verifyNoError(solver.AddLinearObjective(objective.weight, objective.offset, coefficientArray, objective.abs_tolerance, objective.rel_tolerance, objective.priority))
	}
	verifyNoError(solver.SetBoolOption("blend_multi_objectives", build.BlendMultiObjectives))

	if len(build.vars.partialSolution) > 0 {
		indexArray := make([]int32, len(build.vars.partialSolution))
		valueArray := make([]float64, len(build.vars.partialSolution))
		index := 0
		for columnNumber, value := range build.vars.partialSolution {
			indexArray[index] = int32(columnNumber)
			valueArray[index] = value
			index++
		}
		verifyNoError(solver.SetSparseSolution(indexArray, valueArray))
	}
}

func (build *LinearBuilder) configureHighsUtil(solver *highs.Solver, logfile string) {
	//verifyNoError(solver.SetStringOption("parallel", "on"))
	//verifyNoError(solver.SetIntOption("threads", c_threads))

	if build.TimeLimitSeconds != 0 {
		verifyNoError(solver.SetFloatOption("time_limit", float64(build.TimeLimitSeconds)))
	} else {
		verifyNoError(solver.SetFloatOption("time_limit", C_PlusInf))
	}
	verifyNoError(solver.InterruptSupportEnable())

	verifyNoError(solver.SetStringOption("log_file", logfile))
	verifyNoError(solver.SetBoolOption("log_to_console", (C_DebugHighs || C_HighsToConsole) && !build.NoOutput))
	if C_DebugHighs {
		verifyNoError(solver.SetIntOption("log_dev_level", 2))
	} else {
		verifyNoError(solver.SetIntOption("log_dev_level", 0))
	}

	if build.DisablePreSolve {
		verifyNoError(solver.SetStringOption("presolve", "off"))
	} else {
		verifyNoError(solver.SetStringOption("presolve", "on"))
	}
}

func (build *LinearBuilder) configureHighsSolver(solver *highs.Solver) bool {
	requestGpu := false
	expectMip := false

	switch build.Solver {
	case Solver_LP_USE_GPU:
		verifyNoError(solver.SetStringOption("solver", "hipdlp"))
		verifyNoError(solver.SetStringOption("presolve", "on"))
		requestGpu = true
	case Solver_LP_NO_GPU:
		if build.isLargeModel() {
			verifyNoError(solver.SetStringOption("solver", "hipo"))
		} else {
			verifyNoError(solver.SetStringOption("solver", "choose"))
		}
	case Solver_Force_Simplex:
		verifyNoError(solver.SetStringOption("solver", "simplex"))
	case Solver_Force_IPX:
		verifyNoError(solver.SetStringOption("solver", "ipx"))
	case Solver_MIP_Interior:
		if build.isLargeModel() {
			verifyNoError(solver.SetStringOption("solver", "choose"))
			verifyNoError(solver.SetStringOption("mip_lp_solver", "hipo"))
			verifyNoError(solver.SetStringOption("mip_ipm_solver", "hipo"))
		} else {
			verifyNoError(solver.SetStringOption("solver", "choose"))
			verifyNoError(solver.SetStringOption("mip_lp_solver", "choose"))
			verifyNoError(solver.SetStringOption("mip_ipm_solver", "choose"))
		}
		expectMip = true
	case Solver_MIP_Vertex:
		verifyNoError(solver.SetStringOption("solver", "choose"))
		verifyNoError(solver.SetStringOption("mip_lp_solver", "choose"))
		verifyNoError(solver.SetStringOption("mip_ipm_solver", "choose"))
		expectMip = true
	default:
		panic("solver not specified")
	}

	if expectMip != build.isMIP() {
		panic("solver wrong for MIP/non-MIP model")
	}

	return requestGpu
}

func (build *LinearBuilder) isLargeModel() bool {
	return len(build.vars.colTypes) > 500 || len(build.mat.entries) > 500
}

func (build *LinearBuilder) isMIP() bool {
	hasInt := false
	for _, colType := range build.vars.colTypes {
		if colType != highs.Continuous {
			hasInt = true
		}
	}
	return hasInt
}

func (build *LinearBuilder) ValidateInitialSolutionState() {
	colValues := make(map[ColumnIndex]float64)
	for colNum, value := range build.vars.partialSolution {
		colValues[colNum] = value
		if value < build.vars.colLower[colNum] || value > build.vars.colUpper[colNum] {
			panic(fmt.Sprintf("initial value out of bounds index=%d debug=%s", colNum, build.vars.debug[colNum]))
		}
		if build.vars.colTypes[colNum] == highs.Integer && math.Round(value) != value {
			panic("initial value not integer")
		}
	}

	for rowIndex := range build.mat.entries {
		lowerBound := build.mat.lowerBound[rowIndex]
		upperBound := build.mat.upperBound[rowIndex]
		entries := build.mat.entries[rowIndex]
		debug := build.mat.debug[rowIndex]

		anyKnown := false
		anyUnknown := false
		sum := 0.0
		for _, entry := range entries {
			value, haveValue := colValues[entry.columnNumber]
			if haveValue {
				sum += value * entry.value
				anyKnown = true
			} else {
				anyUnknown = true
			}
		}

		if anyKnown && !anyUnknown {
			if !util.FloatsBetween(lowerBound, sum, upperBound) {
				panic("initial values don't fit row " + strconv.FormatInt(int64(rowIndex), 10) + " " + debug)
			}
		}
	}
}
