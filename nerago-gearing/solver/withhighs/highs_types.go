package withhighs

import (
	"cmp"
	"math"
	"os"
	"paladin_gearing_go/util"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_maxSetItems    = 5 // fundamental in MoP gear sets
	c_setItemsCounts = c_maxSetItems + 1

	c_debugHighs = false
	c_threads    = 6
)

// not really consts but would be nice
var (
	c_minusInf = math.Inf(-1)
	c_plusInf  = math.Inf(1)
)

func floatEqualsOne(value float64) bool {
	return 0.999999 <= value && value <= 1.000001
}

func floatEqualsZero(value float64) bool {
	return -0.000001 <= value && value <= 0.000001
}

func floatsApproxEquals(a, b float64) bool {
	ratio := a / b
	return 0.99999 <= ratio && ratio <= 1.00001
}

type columnIndex int32
type rowIndex int32

type inputBuilder struct {
	vars                 variableArrayBuilder
	mat                  constraintMatrixBuilder
	minimise             bool
	blendMultiObjectives bool
}

func (input *inputBuilder) createColumnBool() columnIndex {
	return input.vars.create(highs.Integer, 0, 1, 0)
}

func (input *inputBuilder) createColumnGeneral(varType highs.VariableType, lower, upper float64) columnIndex {
	return input.vars.create(varType, lower, upper, 0)
}

func (input *inputBuilder) createColumnWithOutput(varType highs.VariableType, lower, upper, cost float64) columnIndex {
	return input.vars.create(varType, lower, upper, cost)
}

func (input *inputBuilder) createColumnForLinearObjective(varType highs.VariableType, lower, upper, cost float64, linearObjectiveIndex int) columnIndex {
	return input.vars.createForLinear(varType, lower, upper, cost, linearObjectiveIndex)
}

func (input *inputBuilder) addRow(entries []indexAndValue, lowerBound float64, upperBound float64) rowIndex {
	return input.mat.addRow(entries, lowerBound, upperBound)
}

func (input *inputBuilder) clone() *inputBuilder {
	return &inputBuilder{
		vars: input.vars.clone(),
		mat:  input.mat.clone(),
	}
}

func (input *inputBuilder) runHighs() (*highs.Solution, *util.PrintRecorder) {
	solver, logFilename := input.preHighsRun()

	solution, err := highsPool.RunSolverUnderMutex(solver)
	verifyNoError(err)

	printer := input.postHighsRun(solver, logFilename)
	return solution, printer
}

func (input *inputBuilder) preHighsRun() (*highs.Solver, string) {
	logFilename := makeTempFilename()

	solver := highsPool.Get()
	input.configureHighsModel_internal(solver, logFilename)
	return solver, logFilename
}

func (*inputBuilder) postHighsRun(solver *highs.Solver, logFilename string) *util.PrintRecorder {
	verifyNoError(solver.SetStringOption("log_file", "")) // flush log
	printer := readLogfile(logFilename)

	highsPool.Put(solver)

	return printer
}

func readLogfile(tempFilename string) *util.PrintRecorder {
	bytes, err := os.ReadFile(tempFilename)
	verifyNoError(err)

	printer := util.PrintRecorder_HoldAll()
	printer.PrintBytes(bytes)
	return printer
}

func (input *inputBuilder) configureHighsModel_internal(solver *highs.Solver, logfile string) {
	// checkError(solver.SetStringOption("presolve", "off"))
	// checkError(solver.SetStringOption("parallel", "on"))
	// checkError(solver.SetIntOption("threads", c_threads))
	// verifyNoError(solver.SetFloatOption("time_limit", 300))

	verifyNoError(solver.SetStringOption("log_file", logfile))
	verifyNoError(solver.SetBoolOption("log_to_console", c_debugHighs))
	if c_debugHighs {
		verifyNoError(solver.SetIntOption("log_dev_level", 3))
	} else {
		verifyNoError(solver.SetIntOption("log_dev_level", 0))
	}

	verifyNoError(solver.SetMaximize(!input.minimise))
	verifyNoError(solver.SetBoolOption("blend_multi_objectives", input.blendMultiObjectives))

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

func (vars *variableArrayBuilder) create(varType highs.VariableType, lower, upper, cost float64) columnIndex {
	if len(vars.linearObjectives) > 0 && cost != 0.0 {
		panic("unexpected column cost while using linear objectives")
	}

	index := len(vars.colTypes)
	vars.colTypes = append(vars.colTypes, varType)
	vars.colLower = append(vars.colLower, lower)
	vars.colUpper = append(vars.colUpper, upper)
	vars.colCosts = append(vars.colCosts, cost)
	return columnIndex(index)
}

func (vars *variableArrayBuilder) createForLinear(varType highs.VariableType, lower float64, upper float64, cost float64, linearObjectiveIndex int) columnIndex {
	columnIndex := vars.create(varType, lower, upper, 0)

	linearObjective := &vars.linearObjectives[linearObjectiveIndex]
	linearObjective.coefficientEntries = append(linearObjective.coefficientEntries, indexAndValue{columnIndex, cost})

	return columnIndex
}

func (vars *variableArrayBuilder) clone() variableArrayBuilder {
	return variableArrayBuilder{
		colTypes: slices.Clone(vars.colTypes),
		colCosts: slices.Clone(vars.colCosts),
		colLower: slices.Clone(vars.colLower),
		colUpper: slices.Clone(vars.colUpper),
	}
}

func (vars *variableArrayBuilder) changeColumnCost(columnIndex int, cost float64) {
	vars.colCosts[columnIndex] = cost
}

type indexAndValue struct {
	columnNumber columnIndex
	value        float64
}

type constraintRowBuild struct {
	entries  []indexAndValue
	isAdded  bool
	rowIndex rowIndex
}

func (row *constraintRowBuild) isEmpty() bool {
	return len(row.entries) == 0
}

func (row *constraintRowBuild) hasValues() bool {
	return len(row.entries) > 0
}

func (row *constraintRowBuild) add(columnIndex columnIndex, value float64) {
	if value != 0.0 {
		row.entries = append(row.entries, indexAndValue{columnIndex, value})
	}
}

func (row *constraintRowBuild) change(columnIndex columnIndex, value float64) {
	for i := range row.entries {
		if row.entries[i].columnNumber == columnIndex {
			row.entries[i].value = value
			return
		}
	}

	panic("column didn't exist")
}

func (row *constraintRowBuild) finish(input *inputBuilder, lowerBound float64, upperBound float64) {
	// couldn't find reference for sure that indexes need to be sorted but probably best
	slices.SortFunc(row.entries, func(a, b indexAndValue) int { return cmp.Compare(a.columnNumber, b.columnNumber) })

	if row.isAdded {
		input.mat.changeRow(row.rowIndex, row.entries, lowerBound, upperBound)
	} else {
		row.rowIndex = input.addRow(row.entries, lowerBound, upperBound)
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

func (mat *constraintMatrixBuilder) addRow(entries []indexAndValue, lowerBound float64, upperBound float64) rowIndex {
	index := len(mat.entries)
	mat.entries = append(mat.entries, entries)
	mat.lowerBound = append(mat.lowerBound, lowerBound)
	mat.upperBound = append(mat.upperBound, upperBound)
	return rowIndex(index)
}

func (mat constraintMatrixBuilder) changeRow(rowIndex rowIndex, entries []indexAndValue, lowerBound float64, upperBound float64) {
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
