package multi

import "paladin_gearing_go/util"

func (job *MultiSetJob) evalutateTopN(proposedChannel <-chan MultiProposedOutput, topCount int) []MultiProposedOutput {
	job.printer.Printf("COLLECTING TOP %d\n", topCount)
	bestChannel := util.Channel_TransformAll_Multi(evaluateThreadCount, proposedChannel, func(proposedChannel <-chan MultiProposedOutput, bestChannel chan<- util.HighestCollectorN[MultiProposedOutput]) {
		evalutateTopNWorker(proposedChannel, bestChannel, topCount)
	})
	return util.HighestCollectorN_OfChannel(bestChannel, evaluateThreadCount)
}

func evalutateTopNWorker(proposedChannel <-chan MultiProposedOutput, bestChannel chan<- util.HighestCollectorN[MultiProposedOutput], topCount int) {
	best := util.HighestCollector_ForN[MultiProposedOutput](topCount, func(a, b *MultiProposedOutput) bool {return a.Equals(b)})
	for proposed := range proposedChannel {
		best.Offer(&proposed, proposed.TotalRatingSum)
	}
	bestChannel <- best
}
