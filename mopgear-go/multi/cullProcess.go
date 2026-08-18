package multi

import (
	"sync"
	"sync/atomic"

	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/solver/solve_highs"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
)

func (work *specWorker) runCullingProcess(targetNum int64, waitGroup *sync.WaitGroup, cancel util_async.CancelSignal, tracker *util.TrackProgress, printer *util.PrintRecorder) {
	addScale := uint32(1)
	if work.weightType == 3 {
		targetNum /= 10
		addScale = 10
	}

	currentNum := atomic.Uint64{}
	tracker.RunFromAtomicInt(&currentNum, uint64(targetNum))

	waitGroup.Go(func() {
		highCull := solve_highs.OptionsCulling{}
		solveOptions := items.SolvableOptionsMap_of(work.ItemOptions())
		solverModel := solve_highs_types.SolverModelBuild(work.Model(), work.weightType, nil)
		highCull.Init(work.Label(), targetNum, solveOptions, solverModel, printer)

		resultChannel := highCull.Run(cancel)

		for solvedSet := range resultChannel {
			fullSet := items.FullItemSet_FromSolved(solvedSet, work.ItemOptions())
			work.AddSeenScaled(fullSet.Items(), addScale)
			currentNum.Add(1)
		}

		tracker.SetDone()
	})
}
