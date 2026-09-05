package util_highs

import (
	"fmt"
	"math"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
	"github.com/nerago/mopgear-go/util"
)

type LinearBuilder struct {
	vars                 variableArrayBuilder
	mat                  constraintMatrixBuilder
	NoOutput             bool
	SkipDiagnose         bool
	Minimise             bool
	BlendMultiObjectives bool
	Solver               SolverMode
	DisablePreSolve      bool
	TimeLimitSeconds     int // *TimeLimitToken
	Callback             highs.Callback
	CallbackTypes        []highs.CallbackType
	FloatOptions         map[string]float64
	StringOptions        map[string]string
	IntOptions           map[string]int
	BoolOptions          map[string]bool
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

func (build *LinearBuilder) AddOptionFloat(name string, value float64) {
	if build.FloatOptions == nil {
		build.FloatOptions = make(map[string]float64)
	}
	build.FloatOptions[name] = value
}

func (build *LinearBuilder) AddOptionString(name string, value string) {
	if build.StringOptions == nil {
		build.StringOptions = make(map[string]string)
	}
	build.StringOptions[name] = value
}

func (build *LinearBuilder) AddOptionInt(name string, value int) {
	if build.IntOptions == nil {
		build.IntOptions = make(map[string]int)
	}
	build.IntOptions[name] = value
}

func (build *LinearBuilder) AddOptionBool(name string, value bool) {
	if build.BoolOptions == nil {
		build.BoolOptions = make(map[string]bool)
	}
	build.BoolOptions[name] = value
}

func (build *LinearBuilder) SetEachTolerance(value float64) {
	if build.FloatOptions == nil {
		build.FloatOptions = make(map[string]float64)
	}
	build.FloatOptions["primal_feasibility_tolerance"] = value
	build.FloatOptions["dual_feasibility_tolerance"] = value
	build.FloatOptions["primal_residual_tolerance"] = value
	build.FloatOptions["dual_residual_tolerance"] = value
	build.FloatOptions["optimality_tolerance"] = value
	build.FloatOptions["kkt_tolerance"] = value
	build.FloatOptions["mip_feasibility_tolerance"] = value
	build.FloatOptions["mip_abs_gap"] = value
	build.FloatOptions["ipm_optimality_tolerance"] = value
	build.FloatOptions["pdlp_optimality_tolerance"] = value
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

func (build *LinearBuilder) ChangeColumnOutputWeight(columnIndex ColumnIndex, cost float64) {
	build.vars.changeColumnCost(columnIndex, cost)
}

func (build *LinearBuilder) GetColumnMinMax(columnIndex ColumnIndex) (lower, upper float64) {
	return build.vars.colLower[columnIndex], build.vars.colUpper[columnIndex]
}

func (build *LinearBuilder) ChangeColumnMinMax(columnIndex ColumnIndex, lower, upper float64) {
	build.vars.colLower[columnIndex] = lower
	build.vars.colUpper[columnIndex] = upper
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

func (build *LinearBuilder) isLargeModel() bool {
	return len(build.vars.colTypes) > 500 || len(build.mat.entries) > 500
}

func (build *LinearBuilder) isMIP() bool {
	for _, colType := range build.vars.colTypes {
		if colType != highs.Continuous {
			return true
		}
	}
	return false
}

func (build *LinearBuilder) ValidateInitialSolutionState() {
	colValues := make(map[ColumnIndex]float64)
	for colNum, value := range build.vars.partialSolution {
		colValues[colNum] = value
		if !util.FloatsBetween(build.vars.colLower[colNum], value, build.vars.colUpper[colNum]) {
			panic(fmt.Sprintf("initial value out of bounds index=%d debug=%s values=%e<%e<%e", colNum, build.vars.debug[colNum],
				build.vars.colLower[colNum], value, build.vars.colUpper[colNum]))
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

func (build *LinearBuilder) SetCallback(callbackTypes []highs.CallbackType, callback highs.Callback) {
	build.Callback = callback
	build.CallbackTypes = callbackTypes
}
