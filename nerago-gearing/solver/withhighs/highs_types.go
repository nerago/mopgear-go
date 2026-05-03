package withhighs

import (
	"cmp"
	"math"
	"slices"

	"github.com/lanl/highs"
)

const (
	c_maxSetItems    = 5 // fundamental in MoP gear sets
	c_setItemsCounts = c_maxSetItems + 1
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

type inputBuilder struct {
	vars variableArrayBuilder
	mat  constraintMatrixBuilder
}

func (input *inputBuilder) createColumnBool() int {
	return input.vars.create(highs.IntegerType, 0, 1, 0)
}

func (input *inputBuilder) createColumnGeneral(varType highs.VariableType, lower, upper float64) int {
	return input.vars.create(varType, lower, upper, 0)
}

func (input *inputBuilder) createColumnWithOutput(varType highs.VariableType, lower, upper, cost float64) int {
	return input.vars.create(varType, lower, upper, cost)
}

func (input *inputBuilder) addRow(entries []indexAndValue, lowerBound float64, upperBound float64) {
	input.mat.addRow(entries, lowerBound, upperBound)
}

func (input *inputBuilder) toHighsModel() *highs.RawModel {
	model := highs.NewRawModel()

	model.SetStringOption("presolve", "off")
	model.SetBoolOption("log_to_console", true)
	model.SetIntOption("log_dev_level", 3)

	// tempFile, err := os.CreateTemp("", "highslog")
	// if err != nil {
	// 	panic(err)
	// }
	// tempFile.Close()
	// model.SetStringOption("log_file", tempFile.Name())

	err := model.SetMaximization(true)
	if err != nil {
		panic(err)
	}

	input.vars.applyToModel(model)
	input.mat.finishAndApplyToModelEfficient(model)
	// input.mat.finishAndApplyToModelSlow(model)

	return model
}

type variableArrayBuilder struct {
	ColTypes []highs.VariableType // Type of each model variable
	ColCosts []float64            // Column costs (i.e., the objective function itself)
	ColLower []float64            // Column lower bounds
	ColUpper []float64            // Column upper bounds
}

func (vars *variableArrayBuilder) create(varType highs.VariableType, lower, upper, cost float64) int {
	index := len(vars.ColTypes)
	vars.ColTypes = append(vars.ColTypes, varType)
	vars.ColLower = append(vars.ColLower, lower)
	vars.ColUpper = append(vars.ColUpper, upper)
	vars.ColCosts = append(vars.ColCosts, cost)
	return index
}

func (vars variableArrayBuilder) applyToModel(param *highs.RawModel) {
	err := param.AddColumnBounds(vars.ColLower, vars.ColUpper)
	if err != nil {
		panic(err)
	}

	err = param.SetColumnCosts(vars.ColCosts)
	if err != nil {
		panic(err)
	}

	err = param.SetIntegrality(vars.ColTypes)
	if err != nil {
		panic(err)
	}
}

type indexAndValue struct {
	columnNumber int
	value        float64
}

type constraintRowBuild struct {
	entries []indexAndValue
}

func (row *constraintRowBuild) isEmpty() bool {
	return len(row.entries) == 0
}

func (row *constraintRowBuild) hasValues() bool {
	return len(row.entries) > 0
}

func (row *constraintRowBuild) add(columnIndex int, value float64) {
	if value != 0.0 {
		row.entries = append(row.entries, indexAndValue{columnIndex, value})
	}
}

func (row *constraintRowBuild) finish(input *inputBuilder, lowerBound float64, upperBound float64) {
	// couldn't find reference for sure that indexes need to be sorted but probably best
	slices.SortFunc(row.entries, func(a, b indexAndValue) int { return cmp.Compare(a.columnNumber, b.columnNumber) })

	input.addRow(row.entries, lowerBound, upperBound)
}

type constraintMatrixBuilder struct {
	entries    [][]indexAndValue
	lowerBound []float64
	upperBound []float64
}

func (mat *constraintMatrixBuilder) addRow(entries []indexAndValue, lowerBound float64, upperBound float64) {
	mat.entries = append(mat.entries, entries)
	mat.lowerBound = append(mat.lowerBound, lowerBound)
	mat.upperBound = append(mat.upperBound, upperBound)
}

func (mat *constraintMatrixBuilder) finishAndApplyToModelSlow(param *highs.RawModel) {
	for i := range mat.entries {
		applyRowEntriesForRow(param, mat.entries[i], mat.lowerBound[i], mat.upperBound[i])
	}
}

func (mat *constraintMatrixBuilder) finishAndApplyToModelEfficient(param *highs.RawModel) {
	numRows := len(mat.entries)
	if len(mat.lowerBound) != numRows || len(mat.upperBound) != numRows {
		panic("inconsistent row count")
	}

	valueCount := 0
	for i := range numRows {
		valueCount += len(mat.entries[i])
	}
	if valueCount == 0 {
		panic("completely empty model")
	}

	startArray := make([]int, numRows)
	indexArray := make([]int, valueCount)
	valuesArray := make([]float64, valueCount)

	insertIndex := 0

	for rowNum, rowEntries := range mat.entries {
		startArray[rowNum] = insertIndex

		for _, entry := range rowEntries {
			indexArray[insertIndex] = entry.columnNumber
			valuesArray[insertIndex] = entry.value
			insertIndex++
		}
	}

	err := param.AddCompSparseRows(
		mat.lowerBound,
		startArray,
		indexArray,
		valuesArray,
		mat.upperBound,
	)
	if err != nil {
		panic(err)
	}
}

func applyRowEntriesForRow(param *highs.RawModel, entries []indexAndValue, lowerBound float64, upperBound float64) {
	values := make([]float64, len(entries))
	columnNumbers := make([]int, len(entries))
	for i := range entries {
		values[i] = entries[i].value
		columnNumbers[i] = entries[i].columnNumber
	}

	applyRowData(param, values, columnNumbers, lowerBound, upperBound)
}

func applyRowData(param *highs.RawModel, values []float64, columnNumbers []int, lowerBound float64, upperBound float64) {
	var err error
	if len(values) > 0 {
		err = param.AddCompSparseRows(
			[]float64{lowerBound},
			[]int{0},
			columnNumbers,
			values,
			[]float64{upperBound},
		)
	} else {
		// need to set an explicit zero value so array isn't empty
		// i'd argue this is a bug in go/highs binding library,
		// empty array should be acceptable to lower level code
		// maybe these don't need to be added in many use cases though, automatically skipping seems risky
		err = param.AddCompSparseRows(
			[]float64{lowerBound},
			[]int{0},
			[]int{0},
			[]float64{0.0},
			[]float64{upperBound},
		)
	}

	if err != nil {
		panic(err)
	}
}
