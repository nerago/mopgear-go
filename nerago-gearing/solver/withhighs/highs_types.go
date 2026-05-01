package withhighs

import (
	"github.com/lanl/highs"
)

const (
	c_maxSetItems    = 5 // fundamental in MoP gear sets
	c_setItemsCounts = c_maxSetItems + 1
)

type variableArrayBuilder struct {
	ColTypes []highs.VariableType // Type of each model variable
	ColCosts []float64            // Column costs (i.e., the objective function itself)
	ColLower []float64            // Column lower bounds
	ColUpper []float64            // Column upper bounds
}

func (vars *variableArrayBuilder) add(varType highs.VariableType, lower, upper, cost float64) int {
	index := len(vars.ColTypes)
	vars.ColTypes = append(vars.ColTypes, varType)
	vars.ColLower = append(vars.ColLower, lower)
	vars.ColUpper = append(vars.ColUpper, upper)
	vars.ColCosts = append(vars.ColCosts, cost)
	return index
}

func (vars variableArrayBuilder) apply(param *highs.RawModel) {
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

type constraintRowSequential struct {
	insertColumn  int
	columnNumbers []int
	values        []float64
}

func (row *constraintRowSequential) add(value float64) {
	if value != 0.0 {
		row.columnNumbers = append(row.columnNumbers, row.insertColumn)
		row.values = append(row.values, value)
	}
	row.insertColumn++
}

func (row *constraintRowSequential) apply(param *highs.RawModel, lowerBound float64, upperBound float64) {
	applyRowData(param, row.values, row.columnNumbers, lowerBound, upperBound)
}

type constraintRowSparse struct {
	columnNumbers []int
	values        []float64
}

func (row *constraintRowSparse) add(columnIndex int, value float64) {
	if value != 0.0 {
		row.columnNumbers = append(row.columnNumbers, columnIndex)
		row.values = append(row.values, value)
	}
}

func (row *constraintRowSparse) apply(param *highs.RawModel, lowerBound float64, upperBound float64) {
	applyRowData(param, row.values, row.columnNumbers, lowerBound, upperBound)
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
