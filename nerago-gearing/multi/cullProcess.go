package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/util"
	"sync"
	"sync/atomic"
)

func (param *multiSetParamInternal) runCullingProcess(waitGroup *sync.WaitGroup, tracker *util.TrackProgress) {
	var targetNum int32 = 1000
	currentNum := atomic.Uint64{}
	tracker.RunFromAtomicInt(&currentNum, uint64(targetNum))

	waitGroup.Go(func() {
		highCull := withhighs.OptionsCulling{}
		solveOptions := items.SolvableOptionsMap_of(&param.itemOptions)
		if param.Label == "Ret" {
			highCull.Init(param.Label, targetNum/2, solveOptions, &param.Model, param.job.printer)
		} else {
			highCull.Init(param.Label, targetNum, solveOptions, &param.Model, param.job.printer)
		}

		resultChannel := highCull.Run()

		for solvedSet := range resultChannel {
			fullSet := items.FullItemSet_FromSolved(solvedSet, &param.itemOptions)
			param.seenInSolutions.Add(&fullSet)
			currentNum.Add(1)
		}
	})
}
