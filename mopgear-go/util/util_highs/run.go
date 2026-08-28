package util_highs

import (
	"sync"

	"github.com/bartolsthoorn/gohighs/highs"
	"github.com/nerago/mopgear-go/util"
)

var g_gpuMutex sync.Mutex

func (build *LinearBuilder) runInner(solver *highs.Solver, requireGpu bool, optionalGpu bool, stopwatch *util.Stopwatch) (*highs.Solution, error) {
	if requireGpu {
		g_gpuMutex.Lock()
		defer g_gpuMutex.Unlock()
	} else if optionalGpu {
		if g_gpuMutex.TryLock() {
			defer g_gpuMutex.Unlock()
		} else {
			err := solver.SetStringOption("solver", "choose")
			if err != nil {
				return nil, err
			}
		}
	}

	solution, err := solver.Run()
	stopwatch.AddElapsedDuration(solver.GetRunTime())
	return solution, err
}
