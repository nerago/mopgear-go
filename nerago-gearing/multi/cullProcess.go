package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/util"
	"sync"
	"sync/atomic"
)

func (param *multiSetParamInternal) runCullingProcess(waitGroup *sync.WaitGroup, tracker *util.TrackProgress) {
	var targetNum int64 = 2000
	currentNum := atomic.Uint64{}
	tracker.RunFromAtomicInt(&currentNum, uint64(targetNum))

	waitGroup.Go(func() {
		highCull := withhighs.OptionsCulling{}
		solveOptions := items.SolvableOptionsMap_of(&param.itemOptions)
		switch param.Label {
		case "Ret":
			highCull.Init(param.Label, targetNum/16, solveOptions, &param.Model, param.job.printer)
		case "Prot-Mitigation-WithSet":
			highCull.Init(param.Label, targetNum/8, solveOptions, &param.Model, param.job.printer)
		default:
			highCull.Init(param.Label, targetNum, solveOptions, &param.Model, param.job.printer)
		}

		resultChannel := highCull.Run()

		for solvedSet := range resultChannel {
			fullSet := items.FullItemSet_FromSolved(solvedSet, &param.itemOptions)
			param.seenInSolutions.Add(&fullSet)
			currentNum.Add(1)
		}

		tracker.Stop()
	})
}
