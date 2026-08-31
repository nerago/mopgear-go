package util_highs

import (
	"math"
	"sync"
	"time"

	"github.com/nerago/mopgear-go/util"

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
	Solver_Force_HIPO     SolverMode = iota
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
	elapsed  time.Duration
	err      error
}

func (lr *LinearResult) GetSolutionAndDiscardLog() (*highs.Solution, error) {
	if lr.solution == nil && lr.err == nil {
		return nil, util.ErrorTracedNew("missing solution")
	}
	return lr.solution, lr.err
}

func (lr *LinearResult) GetSolutionAndSaveLog(printer *util.PrintRecorder) (*highs.Solution, error) {
	printer.AppendOther(lr.log)
	if lr.solution == nil && lr.err == nil {
		return nil, util.ErrorTracedNew("missing solution")
	}
	return lr.solution, lr.err
}

func (lr *LinearResult) GetSolution2AndSaveLog(printer *util.PrintRecorder) (*Solution2, error) {
	printer.AppendOther(lr.log)
	if lr.solution == nil && lr.err == nil {
		return nil, util.ErrorTracedNew("missing solution")
	}
	return &Solution2{lr.solution, lr.build}, lr.err
}

func (lr *LinearResult) Elapsed() time.Duration {
	return lr.elapsed
}

type TimeLimitToken struct {
	mutex            sync.Mutex
	initialTimeGiven time.Duration
	solverElapsed    time.Duration
}

func (t *TimeLimitToken) linearBuildSetTimeout(build *LinearBuilder) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	remain := t.initialTimeGiven - t.solverElapsed
	build.TimeLimitSeconds = int(remain.Seconds())
}

func (t *TimeLimitToken) addSolveRunTime(runTime time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.solverElapsed += runTime
}

func TimeLimitTokenMake(seconds time.Duration) *TimeLimitToken {
	return &TimeLimitToken{initialTimeGiven: seconds}
}
