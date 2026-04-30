package withhighs

import (
	"github.com/lanl/highs"
)

const c_maxSetItems = 5 // fundamental in MoP gear sets

type variableArrayBuilder struct {
	ColTypes []highs.VariableType // Type of each model variable
	ColCosts []float64            // Column costs (i.e., the objective function itself)
	ColLower []float64            // Column lower bounds
	ColUpper []float64            // Column upper bounds
}

func (vars *variableArrayBuilder) add(varType highs.VariableType, lower, upper, cost float64) int {
	index := len(vars.ColTypes)
	vars.ColTypes = append(vars.ColTypes, highs.IntegerType)
	vars.ColLower = append(vars.ColLower, 0)
	vars.ColUpper = append(vars.ColUpper, 1)
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
