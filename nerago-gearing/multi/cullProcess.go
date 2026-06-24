package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"sync"
	"sync/atomic"
)

func (param *multiSetParamInternal) runCullingProcess(targetNum int64, waitGroup *sync.WaitGroup, cancel channel_op.CancelSignal, tracker *util.TrackProgress) {
	currentNum := atomic.Uint64{}
	tracker.RunFromAtomicInt(&currentNum, uint64(targetNum))

	waitGroup.Go(func() {
		highCull := withhighs.OptionsCulling{}
		solveOptions := items.SolvableOptionsMap_of(&param.itemOptions)
		highCull.Init(param.Label, targetNum, solveOptions, &param.Model, param.job.printer)

		resultChannel := highCull.Run(cancel)

		for solvedSet := range resultChannel {
			fullSet := items.FullItemSet_FromSolved(solvedSet, &param.itemOptions)
			param.seenInSolutions.Add(&fullSet)
			currentNum.Add(1)
		}

		tracker.SetDone()
	})
}
