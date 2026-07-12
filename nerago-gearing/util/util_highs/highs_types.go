package util_highs

import (
	"cmp"
	"fmt"
	"math"
	"os"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"slices"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	C_HighsToConsole     = true
	C_DebugHighs         = false
	C_DiagnoseInfeasible = false
	c_threads            = 6
)

// not really consts but would be nice
var (
	C_MinusInf = math.Inf(-1)
	C_PlusInf  = math.Inf(1)
)

type SolverMode int8

const (
	Solver_NotSet        SolverMode = iota
	Solver_LP_USE_GPU    SolverMode = iota
	Solver_LP_NO_GPU     SolverMode = iota
	Solver_MIP_Interior  SolverMode = iota
	Solver_MIP_Vertex    SolverMode = iota
	Solver_Force_Simplex SolverMode = iota
	Solver_Force_IPX     SolverMode = iota
)

type ColumnIndex int32
type ObjectiveIndex int32
type RowIndex int32

type LinearResult struct {
	solution *highs.Solution
	log      *util.PrintRecorder
}

func (lr *LinearResult) GetSolutionAndDiscardLog() *highs.Solution {
	return lr.solution
}

func (lr *LinearResult) GetSolutionAndSaveLog(printer *util.PrintRecorder) *highs.Solution {
	printer.AppendOther(lr.log)
	return lr.solution
}

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
		future.SetResult(LinearResult{solution, log})
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

	// verifyNoError(solver.InterruptSupportDisable())

	verifyNoError(solver.Clear())
	G_HighsPool.Put(solver)
}

func readLogfile(tempFilename string, printer *util.PrintRecorder) {
	if tempFilename != "" && printer != nil {
		file, err := os.Open(tempFilename)
		verifyNoError(err)
		defer file.Close()

		_, err = file.WriteTo(printer)
		verifyNoError(err)
	}
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

func verifyNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func makeTempFilename() string {
	tempFile, err := os.CreateTemp("", "highslog")
	if err != nil {
		panic(err)
	}
	verifyNoError(tempFile.Close())
	return tempFile.Name()
}

type DebugContext interface {
	DebugText() string
}

type DebugString struct {
	Text string
}

func (debugString DebugString) DebugText() string {
	return debugString.Text
}

func debugText(debug DebugContext) string {
	debugText := ""
	if debug != nil {
		debugText = debug.DebugText()
	}
	return debugText
}

func DebugText(text string) DebugString {
	return DebugString{Text: text}
}

func (build *LinearBuilder) DebugPrintColumns(solution *highs.Solution, printer *util.PrintRecorder) {
	if C_DebugHighs {
		build.debugPrintColumnsForce(solution, printer)
	}
}
func (build *LinearBuilder) debugPrintColumnsForce(solution *highs.Solution, printer *util.PrintRecorder) {
	for i, x := range solution.ColValues {
		printer.Printf("%3d %14f %s\n", i, x, debugText(build.vars.debug[i]))
	}
}

func (build *LinearBuilder) DebugTextFor(columnIndex ColumnIndex) string {
	return debugText(build.vars.debug[columnIndex])
}

type variableArrayBuilder struct {
	colTypes []highs.VariableType // Type of each model variable
	colCosts []float64            // Column costs (i.e., the objective function itself)
	colLower []float64            // Column lower bounds
	colUpper []float64            // Column upper bounds
	debug    []DebugContext

	partialSolution map[ColumnIndex]float64
	objectives      []objectiveFields
}

type objectiveFields struct {
	coefficientEntries []indexAndValue
	weight             float64
	offset             float64
	abs_tolerance      float64
	rel_tolerance      float64
	priority           int
}

func (lin *objectiveFields) clone() objectiveFields {
	return objectiveFields{
		coefficientEntries: slices.Clone(lin.coefficientEntries),
		weight:             lin.weight,
		offset:             lin.offset,
		abs_tolerance:      lin.abs_tolerance,
		rel_tolerance:      lin.rel_tolerance,
		priority:           lin.priority,
	}
}

func (vars *variableArrayBuilder) addObjective(weight float64, offset float64, abs_tolerance float64, rel_tolerance float64, priority int) ObjectiveIndex {
	index := len(vars.objectives)
	vars.objectives = append(vars.objectives, objectiveFields{
		weight:        weight,
		offset:        offset,
		abs_tolerance: abs_tolerance,
		rel_tolerance: rel_tolerance,
		priority:      priority,
	})
	return ObjectiveIndex(index)
}

func (vars *variableArrayBuilder) create(varType highs.VariableType, lower, upper, cost float64, debug DebugContext) ColumnIndex {
	if len(vars.objectives) > 0 && cost != 0.0 {
		panic("unexpected column cost while using linear objectives")
	}

	return vars.create_inner(varType, lower, upper, cost, debug)
}

func (vars *variableArrayBuilder) create_inner(varType highs.VariableType, lower float64, upper float64, cost float64, debug DebugContext) ColumnIndex {
	index := len(vars.colTypes)
	vars.colTypes = append(vars.colTypes, varType)
	vars.colLower = append(vars.colLower, lower)
	vars.colUpper = append(vars.colUpper, upper)
	vars.colCosts = append(vars.colCosts, cost)
	vars.debug = append(vars.debug, debug)
	return ColumnIndex(index)
}

func (vars *variableArrayBuilder) createForLinear(varType highs.VariableType, lower float64, upper float64, cost float64, linearObjectiveIndex ObjectiveIndex, debug DebugContext) ColumnIndex {
	if linearObjectiveIndex == -1 && len(vars.objectives) == 0 {
		return vars.create_inner(varType, lower, upper, cost, debug)
	} else if int(linearObjectiveIndex) < len(vars.objectives) {
		columnIndex := vars.create_inner(varType, lower, upper, 0, debug)

		linearObjective := &vars.objectives[linearObjectiveIndex]
		linearObjective.coefficientEntries = append(linearObjective.coefficientEntries, indexAndValue{columnIndex, cost})

		return columnIndex
	} else {
		panic("no such linear objective")
	}
}

func (vars *variableArrayBuilder) clone() variableArrayBuilder {
	return variableArrayBuilder{
		colTypes:   slices.Clone(vars.colTypes),
		colCosts:   slices.Clone(vars.colCosts),
		colLower:   slices.Clone(vars.colLower),
		colUpper:   slices.Clone(vars.colUpper),
		debug:      slices.Clone(vars.debug),
		objectives: util.MapSliceAsNew(vars.objectives, (*objectiveFields).clone),
	}
}

func (vars *variableArrayBuilder) changeColumnCost(columnIndex int, cost float64) {
	vars.colCosts[columnIndex] = cost
}

type indexAndValue struct {
	columnNumber ColumnIndex
	value        float64
}

type ConstraintRow struct {
	entries []indexAndValue
	Debug   string
}

func (row *ConstraintRow) IsEmpty() bool {
	return len(row.entries) == 0
}

func (row *ConstraintRow) HasValues() bool {
	return len(row.entries) > 0
}

func (row *ConstraintRow) Add(columnIndex ColumnIndex, value float64) {
	if value != 0.0 {
		row.entries = append(row.entries, indexAndValue{columnIndex, value})
	}
}

func (row *ConstraintRow) Change(columnIndex ColumnIndex, value float64) {
	for i := range row.entries {
		if row.entries[i].columnNumber == columnIndex {
			row.entries[i].value = value
			return
		}
	}

	panic("column didn't exist")
}

func (row *ConstraintRow) Build(build *LinearBuilder, lowerBound float64, upperBound float64) {
	// couldn't find reference for sure that indexes need to be sorted but probably best
	slices.SortFunc(row.entries, func(a, b indexAndValue) int { return cmp.Compare(a.columnNumber, b.columnNumber) })

	if C_DebugHighs {
		if len(row.entries) == 0 && lowerBound != 0 && upperBound != 0 {
			panic("empty row makes infeasible")
		} else if len(row.entries) == 0 {
			fmt.Printf("warn empty row\n")
		}

		if (0 < math.Abs(lowerBound) && math.Abs(lowerBound) < 1e-4) || (0 < math.Abs(upperBound) && math.Abs(upperBound) < 1e-4) {
			panic("row bounds very small")
		}
		for i := range row.entries {
			value := row.entries[i].value
			if value >= 1e+15 {
				panic("matrix entry very large")
			}
		}
	}

	build.mat.addRow(row.entries, lowerBound, upperBound, row.Debug)
}

type constraintMatrixBuilder struct {
	entries    [][]indexAndValue
	lowerBound []float64
	upperBound []float64
	debug      []string
}

func (mat *constraintMatrixBuilder) clone() constraintMatrixBuilder {
	return constraintMatrixBuilder{
		entries: util.MapSliceAsNew(mat.entries, func(subSlice *[]indexAndValue) []indexAndValue {
			return slices.Clone(*subSlice)
		}),
		lowerBound: slices.Clone(mat.lowerBound),
		upperBound: slices.Clone(mat.upperBound),
		debug:      slices.Clone(mat.debug),
	}
}

func (mat *constraintMatrixBuilder) addRow(entries []indexAndValue, lowerBound float64, upperBound float64, debug string) RowIndex {
	index := len(mat.entries)
	mat.entries = append(mat.entries, entries)
	mat.lowerBound = append(mat.lowerBound, lowerBound)
	mat.upperBound = append(mat.upperBound, upperBound)
	mat.debug = append(mat.debug, debug)
	return RowIndex(index)
}

func (mat *constraintMatrixBuilder) deleteRow(rowIndex int) {
	mat.entries = util.DeleteIndexInPlace(mat.entries, rowIndex)
	mat.lowerBound = util.DeleteIndexInPlace(mat.lowerBound, rowIndex)
	mat.upperBound = util.DeleteIndexInPlace(mat.upperBound, rowIndex)
	mat.debug = util.DeleteIndexInPlace(mat.debug, rowIndex)
}

func (mat *constraintMatrixBuilder) deleteRowRange(firstDelete, lastDelete int) {
	mat.entries = slices.Delete(mat.entries, firstDelete, lastDelete+1)
	mat.lowerBound = slices.Delete(mat.lowerBound, firstDelete, lastDelete+1)
	mat.upperBound = slices.Delete(mat.upperBound, firstDelete, lastDelete+1)
	mat.debug = slices.Delete(mat.debug, firstDelete, lastDelete+1)
}

func (mat *constraintMatrixBuilder) createSolverInputArrays() (numRows int32, lowerBound []float64, upperBound []float64, startArray []int32, indexArray []int32, valuesArray []float64) {
	numRows = int32(len(mat.entries))
	if int32(len(mat.lowerBound)) != numRows || int32(len(mat.upperBound)) != numRows {
		panic("inconsistent row count")
	}

	valueCount := 0
	for i := range numRows {
		valueCount += len(mat.entries[i])
	}
	if valueCount == 0 {
		panic("completely empty model")
	}

	startArray = make([]int32, numRows)
	indexArray = make([]int32, valueCount)
	valuesArray = make([]float64, valueCount)

	var insertIndex int32 = 0
	for rowNum, rowEntries := range mat.entries {
		startArray[rowNum] = insertIndex

		for _, entry := range rowEntries {
			indexArray[insertIndex] = int32(entry.columnNumber)
			valuesArray[insertIndex] = entry.value
			insertIndex++
		}
	}

	return numRows, mat.lowerBound, mat.upperBound, startArray, indexArray, valuesArray
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
