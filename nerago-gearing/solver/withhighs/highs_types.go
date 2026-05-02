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

// not sure if necessary but I don't trust floats
func floatEqualsOne(value float64) bool {
	return 0.999999 <= value && value <= 1.000001
}

func floatsApproxEquals(a, b float64) bool {
	ratio := a / b
	return 0.99999 <= ratio && ratio <= 1.00001
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

func (row *constraintRowBuild) add(columnIndex int, value float64) {
	if value != 0.0 {
		row.entries = append(row.entries, indexAndValue{columnIndex, value})
	}
}

func (row *constraintRowBuild) finish(build *constraintMatrixBuilder, lowerBound float64, upperBound float64) {
	// couldn't find reference for sure that indexes need to be sorted but probably best
	slices.SortFunc(row.entries, func(a, b indexAndValue) int { return cmp.Compare(a.columnNumber, b.columnNumber) })

	build.addRow(row.entries, lowerBound, upperBound)
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

	valueCount := 0
	for i := range numRows {
		valueCount += len(mat.entries[i])
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
