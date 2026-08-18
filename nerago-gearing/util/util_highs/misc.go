package util_highs

import (
	"math"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	C_HighsToConsole     = true
	C_DebugHighs         = false
	C_DiagnoseInfeasible = false
)

func InfPos() float64 {
	return math.Inf(1)
}
func InfNeg() float64 {
	return math.Inf(-1)
}

type SolverMode int8

const (
	Solver_NotSet         SolverMode = iota
	Solver_LP_USE_GPU     SolverMode = iota
	Solver_LP_NO_GPU      SolverMode = iota
	Solver_LP_GPU_IF_FREE SolverMode = iota
	Solver_MIP_Interior   SolverMode = iota
	Solver_MIP_Vertex     SolverMode = iota
	Solver_Force_Simplex  SolverMode = iota
	Solver_Force_IPX      SolverMode = iota
	Solver_Flexible       SolverMode = iota
)

type ColumnIndex int32
type ObjectiveIndex int32
type RowIndex int32

type LinearResult struct {
	solution *highs.Solution
	build    *LinearBuilder
	log      *util.PrintRecorder
}

func (lr *LinearResult) GetSolutionAndDiscardLog() *highs.Solution {
	return lr.solution
}

func (lr *LinearResult) GetSolutionAndSaveLog(printer *util.PrintRecorder) *highs.Solution {
	printer.AppendOther(lr.log)
	return lr.solution
}

func (lr *LinearResult) GetSolution2AndSaveLog(printer *util.PrintRecorder) *Solution2 {
	printer.AppendOther(lr.log)
	return &Solution2{lr.solution, lr.build}
}

func verifyNoError(err error) {
	if err != nil {
		panic(err)
	}
}
