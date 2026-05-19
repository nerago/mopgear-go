package utilhighs

import (
	"cmp"
	"math"
	"os"
	"paladin_gearing_go/util"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	C_DebugHighs = true
	c_threads    = 6
)

// not really consts but would be nice
var (
	C_MinusInf = math.Inf(-1)
	C_PlusInf  = math.Inf(1)
)

func FloatEqualsOne(value float64) bool {
	return 0.999999 <= value && value <= 1.000001
}

func FloatEqualsZero(value float64) bool {
	return -0.000001 <= value && value <= 0.000001
}

func FloatsApproxEquals(a, b float64) bool {
	ratio := a / b
	return 0.99999 <= ratio && ratio <= 1.00001
}

type ColumnIndex int32
type RowIndex int32

type InputBuilder struct {
	vars                 variableArrayBuilder
	mat                  constraintMatrixBuilder
	Minimise             bool
	BlendMultiObjectives bool
	Solver               string
}

func (input *InputBuilder) AddLinearObjective(weight float64, offset float64, abs_tolerance float64, rel_tolerance float64, priority int) int {
	return input.vars.addLinearObjective(weight, offset, abs_tolerance, rel_tolerance, priority)
}

func (input *InputBuilder) CreateColumnBool() ColumnIndex {
	return input.vars.create(highs.Integer, 0, 1, 0)
}

func (input *InputBuilder) CreateColumnGeneral(varType highs.VariableType, lower, upper float64) ColumnIndex {
	return input.vars.create(varType, lower, upper, 0)
}

func (input *InputBuilder) CreateColumnWithOutput(varType highs.VariableType, lower, upper, cost float64) ColumnIndex {
	return input.vars.create(varType, lower, upper, cost)
}

func (input *InputBuilder) CreateColumnForLinearObjective(varType highs.VariableType, lower, upper, cost float64, linearObjectiveIndex int) ColumnIndex {
	return input.vars.createForLinear(varType, lower, upper, cost, linearObjectiveIndex)
}

func (input *InputBuilder) AddRow(entries []indexAndValue, lowerBound float64, upperBound float64) RowIndex {
	return input.mat.addRow(entries, lowerBound, upperBound)
}

func (input *InputBuilder) Clone() *InputBuilder {
	return &InputBuilder{
		vars:                 input.vars.clone(),
		mat:                  input.mat.clone(),
		Minimise:             input.Minimise,
		BlendMultiObjectives: input.BlendMultiObjectives,
	}
}

func (input *InputBuilder) RunHighs() (*highs.Solution, *util.PrintRecorder) {
	solver, logFilename := input.preHighsRun()

	solution, err := G_HighsPool.RunSolverUnderMutex(solver)
	verifyNoError(err)

	printer := input.postHighsRun(solver, logFilename)
	return solution, printer
}

func (input *InputBuilder) preHighsRun() (*highs.Solver, string) {
	logFilename := makeTempFilename()

	solver := G_HighsPool.Get()
	input.configureHighsModel_internal(solver, logFilename)
	return solver, logFilename
}

func (*InputBuilder) postHighsRun(solver *highs.Solver, logFilename string) *util.PrintRecorder {
	verifyNoError(solver.SetStringOption("log_file", "")) // flush log
	printer := readLogfile(logFilename)

	G_HighsPool.Put(solver)

	return printer
}

func readLogfile(tempFilename string) *util.PrintRecorder {
	bytes, err := os.ReadFile(tempFilename)
	verifyNoError(err)

	printer := util.PrintRecorder_HoldAll()
	printer.PrintBytes(bytes)
	return printer
}

func (input *InputBuilder) configureHighsModel_internal(solver *highs.Solver, logfile string) {
	// checkError(solver.SetStringOption("presolve", "off"))
	// checkError(solver.SetStringOption("parallel", "on"))
	// checkError(solver.SetIntOption("threads", c_threads))
	// verifyNoError(solver.SetFloatOption("time_limit", 300))

	verifyNoError(solver.SetStringOption("log_file", logfile))
	verifyNoError(solver.SetBoolOption("log_to_console", C_DebugHighs))
	if C_DebugHighs {
		verifyNoError(solver.SetIntOption("log_dev_level", 3))
	} else {
		verifyNoError(solver.SetIntOption("log_dev_level", 0))
	}

	verifyNoError(solver.SetMaximize(!input.Minimise))
	verifyNoError(solver.SetBoolOption("blend_multi_objectives", input.BlendMultiObjectives))

	if input.Solver != "" {
		verifyNoError(solver.SetStringOption("solver", input.Solver))
	} else {
		verifyNoError(solver.SetStringOption("solver", "choose"))
	}

	numRows, lowerBound, upperBound, startArray, indexArray, valuesArray := input.mat.createSolverInputArrays()
	verifyNoError(solver.PassModel2(
		int32(len(input.vars.colTypes)),
		numRows,
		input.vars.colCosts, input.vars.colLower, input.vars.colUpper,
		lowerBound, upperBound,
		startArray, indexArray, valuesArray,
		input.vars.colTypes, true, 0))

	for linearObjectiveIndex := range input.vars.linearObjectives {
		objective := &input.vars.linearObjectives[linearObjectiveIndex]
		coefficientArray := make([]float64, len(input.vars.colTypes))
		for _, entry := range objective.coefficientEntries {
			coefficientArray[entry.columnNumber] = entry.value
		}
		verifyNoError(solver.AddLinearObjective(objective.weight, objective.offset, coefficientArray, objective.abs_tolerance, objective.rel_tolerance, objective.priority))
	}
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

type variableArrayBuilder struct {
	colTypes         []highs.VariableType // Type of each model variable
	colCosts         []float64            // Column costs (i.e., the objective function itself)
	colLower         []float64            // Column lower bounds
	colUpper         []float64            // Column upper bounds
	linearObjectives []linearObjectiveFields
}

type linearObjectiveFields struct {
	coefficientEntries []indexAndValue
	weight             float64
	offset             float64
	abs_tolerance      float64
	rel_tolerance      float64
	priority           int
}

func (lin *linearObjectiveFields) clone() linearObjectiveFields {
	return linearObjectiveFields{
		coefficientEntries: slices.Clone(lin.coefficientEntries),
		weight:             lin.weight,
		offset:             lin.offset,
		abs_tolerance:      lin.abs_tolerance,
		rel_tolerance:      lin.rel_tolerance,
		priority:           lin.priority,
	}
}

func (vars *variableArrayBuilder) addLinearObjective(weight float64, offset float64, abs_tolerance float64, rel_tolerance float64, priority int) int {
	index := len(vars.linearObjectives)
	vars.linearObjectives = append(vars.linearObjectives, linearObjectiveFields{
		weight:        weight,
		offset:        offset,
		abs_tolerance: abs_tolerance,
		rel_tolerance: rel_tolerance,
		priority:      priority,
	})
	return index
}

func (vars *variableArrayBuilder) create(varType highs.VariableType, lower, upper, cost float64) ColumnIndex {
	if len(vars.linearObjectives) > 0 && cost != 0.0 {
		panic("unexpected column cost while using linear objectives")
	}

	index := len(vars.colTypes)
	vars.colTypes = append(vars.colTypes, varType)
	vars.colLower = append(vars.colLower, lower)
	vars.colUpper = append(vars.colUpper, upper)
	vars.colCosts = append(vars.colCosts, cost)
	return ColumnIndex(index)
}

func (vars *variableArrayBuilder) createForLinear(varType highs.VariableType, lower float64, upper float64, cost float64, linearObjectiveIndex int) ColumnIndex {
	columnIndex := vars.create(varType, lower, upper, 0)

	linearObjective := &vars.linearObjectives[linearObjectiveIndex]
	linearObjective.coefficientEntries = append(linearObjective.coefficientEntries, indexAndValue{columnIndex, cost})

	return columnIndex
}

func (vars *variableArrayBuilder) clone() variableArrayBuilder {
	return variableArrayBuilder{
		colTypes:         slices.Clone(vars.colTypes),
		colCosts:         slices.Clone(vars.colCosts),
		colLower:         slices.Clone(vars.colLower),
		colUpper:         slices.Clone(vars.colUpper),
		linearObjectives: util.MapSliceAsNew(vars.linearObjectives, (*linearObjectiveFields).clone),
	}
}

func (vars *variableArrayBuilder) changeColumnCost(columnIndex int, cost float64) {
	vars.colCosts[columnIndex] = cost
}

type indexAndValue struct {
	columnNumber ColumnIndex
	value        float64
}

type ConstraintRowBuild struct {
	entries  []indexAndValue
	isAdded  bool
	rowIndex RowIndex
}

func (row *ConstraintRowBuild) IsEmpty() bool {
	return len(row.entries) == 0
}

func (row *ConstraintRowBuild) HasValues() bool {
	return len(row.entries) > 0
}

func (row *ConstraintRowBuild) Add(columnIndex ColumnIndex, value float64) {
	if value != 0.0 {
		row.entries = append(row.entries, indexAndValue{columnIndex, value})
	}
}

func (row *ConstraintRowBuild) Change(columnIndex ColumnIndex, value float64) {
	for i := range row.entries {
		if row.entries[i].columnNumber == columnIndex {
			row.entries[i].value = value
			return
		}
	}

	panic("column didn't exist")
}

func (row *ConstraintRowBuild) Finish(input *InputBuilder, lowerBound float64, upperBound float64) {
	// couldn't find reference for sure that indexes need to be sorted but probably best
	slices.SortFunc(row.entries, func(a, b indexAndValue) int { return cmp.Compare(a.columnNumber, b.columnNumber) })

	if row.isAdded {
		input.mat.changeRow(row.rowIndex, row.entries, lowerBound, upperBound)
	} else {
		row.rowIndex = input.AddRow(row.entries, lowerBound, upperBound)
		row.isAdded = true
	}
}

type constraintMatrixBuilder struct {
	entries    [][]indexAndValue
	lowerBound []float64
	upperBound []float64
}

func (mat *constraintMatrixBuilder) clone() constraintMatrixBuilder {
	return constraintMatrixBuilder{
		entries: util.MapSliceAsNew(mat.entries, func(subSlice *[]indexAndValue) []indexAndValue {
			return slices.Clone(*subSlice)
		}),
		lowerBound: slices.Clone(mat.lowerBound),
		upperBound: slices.Clone(mat.upperBound),
	}
}

func (mat *constraintMatrixBuilder) addRow(entries []indexAndValue, lowerBound float64, upperBound float64) RowIndex {
	index := len(mat.entries)
	mat.entries = append(mat.entries, entries)
	mat.lowerBound = append(mat.lowerBound, lowerBound)
	mat.upperBound = append(mat.upperBound, upperBound)
	return RowIndex(index)
}

func (mat constraintMatrixBuilder) changeRow(rowIndex RowIndex, entries []indexAndValue, lowerBound float64, upperBound float64) {
	mat.entries[rowIndex] = entries
	mat.lowerBound[rowIndex] = lowerBound
	mat.upperBound[rowIndex] = upperBound
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
