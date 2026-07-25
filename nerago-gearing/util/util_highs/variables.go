package util_highs

import (
	"paladin_gearing_go/util/util_collection"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

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
		objectives: util_collection.MapSliceAsNew(vars.objectives, (*objectiveFields).clone),
	}
}

func (vars *variableArrayBuilder) changeColumnCost(columnIndex int, cost float64) {
	vars.colCosts[columnIndex] = cost
}
