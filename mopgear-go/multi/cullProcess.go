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
		defer tracker.SetDone()

		solveOptions := items.SolvableOptionsMap_of(work.ItemOptions())
		solverModel, err := solve_highs_types.SolverModelBuild(work.Model(), work.weightType, nil)
		if err != nil {
			util.GlobalWarnHandler(err)
			return // bail out, don't really have any good error handling in here
		}

		highCull := solve_highs.OptionsCulling{}
		highCull.Init(work.Label(), targetNum, solveOptions, solverModel, printer)

		resultChannel := highCull.Run(cancel)

		for solvedSet := range resultChannel {
			fullSet, err := items.FullItemSet_FromSolved(solvedSet, work.ItemOptions())
			if err == nil {
				work.AddSeenScaled(fullSet.Items(), addScale)
			} else {
				util.GlobalWarnHandler(err)
			}
			currentNum.Add(1)
		}
	})
}
