package util_highs

import (
	"github.com/nerago/mopgear-go/util"
	"sync"

	"github.com/bartolsthoorn/gohighs/highs"
)

var G_HighsPool = highsPoolType{}

type highsPoolType struct {
	instances      []*highs.Solver
	instancesMutex sync.Mutex

	// runMutex sync.RWMutex // running solve instances are considered "reads" while cleanup is a "write"
	gpuMutex sync.Mutex
}

func (pool *highsPoolType) Get() *highs.Solver {
	pool.instancesMutex.Lock()
	defer pool.instancesMutex.Unlock()

	if len(pool.instances) > 0 {
		solver := pool.instances[len(pool.instances)-1]
		pool.instances = pool.instances[:len(pool.instances)-1]
		return solver
	} else {
		solver, err := highs.NewSolver()
		verifyNoError(err)
		return solver
	}
}

func (pool *highsPoolType) Put(solver *highs.Solver) {
	pool.instancesMutex.Lock()
	pool.instances = append(pool.instances, solver)
	pool.instancesMutex.Unlock()

	// pool.tryClosePending()
}

func (pool *highsPoolType) RunSolverUnderMutex(solver *highs.Solver, requireGpu, optionalGpu bool, stopwatch *util.Stopwatch) (solution *highs.Solution, err error) {
	// pool.runMutex.RLock()
	// defer pool.runMutex.RUnlock()

	if requireGpu {
		//stopwatch.Start()
		//verifyNoError(solver.Presolve())
		//stopwatch.Stop()

		pool.gpuMutex.Lock()
		stopwatch.Start()
		solution, err = solver.Run()
		stopwatch.Stop()
		pool.gpuMutex.Unlock()
	} else if optionalGpu {
		if pool.gpuMutex.TryLock() {
			defer pool.gpuMutex.Unlock()
		} else {
			verifyNoError(solver.SetStringOption("solver", "choose"))
		}

		stopwatch.Start()
		solution, err = solver.Run()
		stopwatch.Stop()
	} else {
		stopwatch.Start()
		solution, err = solver.Run()
		stopwatch.Stop()
	}

	return solution, err
}

// solver.Close is not thread safe so we pool instances instead
// see comments in /highs/Highs.h:resetGlobalScheduler
// func (pool *highsPoolType) tryClosePending() {
// 	if pool.runMutex.TryLock() {
// 		pool.instancesMutex.Lock()

// 		for _, solver := range pool.instances {
// 			fmt.Println("Closing high")
// 			solver.Close()
// 			fmt.Println("Close ok")
// 		}

// 		clear(pool.instances)
// 		pool.instances = pool.instances[:0]

// 		pool.instancesMutex.Unlock()
// 		pool.runMutex.Unlock()
// 	}
// }
