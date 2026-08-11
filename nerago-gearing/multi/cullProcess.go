package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"sync"
	"sync/atomic"
)

func (work *specWorking) runCullingProcess(targetNum int64, waitGroup *sync.WaitGroup, cancel util_async.CancelSignal, tracker *util.TrackProgress, printer *util.PrintRecorder) {
	currentNum := atomic.Uint64{}
	tracker.RunFromAtomicInt(&currentNum, uint64(targetNum))

	waitGroup.Go(func() {
		highCull := solve_highs.OptionsCulling{}
		solveOptions := items.SolvableOptionsMap_of(&work.itemPrep.itemOptions)
		solverModel := solve_highs_types.SolverModelBuild(&work.itemPrep.model, work.weightType, nil)
		highCull.Init(work.itemPrep.label, targetNum, solveOptions, solverModel, printer)

		resultChannel := highCull.Run(cancel)

		for solvedSet := range resultChannel {
			fullSet := items.FullItemSet_FromSolved(solvedSet, &work.itemPrep.itemOptions)
			work.itemPrep.seenInSolutions.Add(new(fullSet))
			currentNum.Add(1)
		}

		tracker.SetDone()
	})
}
