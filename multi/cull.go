package multi

import (
	"math/rand"
	"paladin_gearing_go/util"
	"sync"
)

func (job *MultiSetJob) SuggestCulls(targetCount uint64, topCapture int) {
	job.prepareInitial()
	commonOptions := job.determineCommon()
	comboChannel := job.makeCommonChannel(commonOptions, targetCount)
	proposedChannel := job.makeProposedChannel(comboChannel)
	bestOutputs := job.evalutateTopN(proposedChannel, topCapture)
	// bestOutputs := job.combinedPipeline(commonOptions, targetCount, topCapture)
	for _, best := range bestOutputs {
		job.printer.Printf("::::::::: MULTI RATING %d ::::::::\n", best.TotalRatingSum)
		for i, out := range best.Outputs {
			job.printer.Println(job.params[i].Label)
			out.Report(&job.printer)
		}
	}
}

func (job *MultiSetJob) combinedPipeline(commonOptions commonComboOptions, targetCount uint64, topCapture int) []MultiProposedOutput {
	eachThreadCount := max(targetCount/generateThreadCount, 1)
	bestChannel := make(chan util.HighestCollectorN[MultiProposedOutput])

	var waitGroup sync.WaitGroup
	for threadNum := range generateThreadCount {
		waitGroup.Go(func() {
			best := util.HighestCollector_ForN(topCapture, func(a, b *MultiProposedOutput) bool { return a.Equals(b) })
			rng := rand.New(rand.NewSource(int64(threadNum)))
			for range eachThreadCount {
				combo := makeRandomCombo(commonOptions, rng)
				proposed := job.subSolveCombo(combo)
				if proposed != nil {
					best.Offer(proposed, proposed.TotalRatingSum)
				}
			}
			bestChannel <- best
		})
	}
	return util.HighestCollectorN_OfChannel(bestChannel, generateThreadCount)
}
