package multi

import (
	"paladin_gearing_go/util/channel_op"
	"paladin_gearing_go/util/util_rank"
)

func (job *MultiSetJob) evalutateTopN(proposedChannel <-chan MultiProposedOutput, topCount int) []MultiProposedOutput {
	job.printer.Printf("COLLECTING TOP %d\n", topCount)
	bestChannel := channel_op.TransformAll_ChannelToChannel(evaluateThreadCount, proposedChannel, func(proposedChannel <-chan MultiProposedOutput, bestChannel chan<- util_rank.HighestCollectorN[MultiProposedOutput]) {
		evalutateTopNWorker(proposedChannel, bestChannel, topCount)
	})
	return util_rank.HighestCollectorN_OfChannel(bestChannel, evaluateThreadCount)
}

func evalutateTopNWorker(proposedChannel <-chan MultiProposedOutput, bestChannel chan<- util_rank.HighestCollectorN[MultiProposedOutput], topCount int) {
	best := util_rank.HighestCollector_ForN(topCount, func(a, b *MultiProposedOutput) bool { return a.Equals(b) })
	for proposed := range proposedChannel {
		best.Offer(&proposed, proposed.TotalRatingSum)
	}
	bestChannel <- best
}
