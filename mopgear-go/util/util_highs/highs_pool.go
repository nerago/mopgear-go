package util_highs

import (
	"sync"

	"github.com/bartolsthoorn/gohighs/highs"
)

var G_HighsPool = highsPoolType{}

type highsPoolType struct {
	instances      []*highs.Solver
	instancesMutex sync.Mutex
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
}
