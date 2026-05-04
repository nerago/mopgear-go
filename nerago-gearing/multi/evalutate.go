package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"paladin_gearing_go/util/util_rank"
)

func (job *MultiSetJob) runForTopN(targetCount uint64, topCapture uint64, trackProgress *util.TrackProgress) []multiProposedOutput {
	job.prepareInitial()
	commonOptions := job.determineCommon()

	comboChannel, comboCount := job.makeCommonChannel(commonOptions, targetCount)
	proposedChannel := job.makeProposedChannel(comboChannel, comboCount, trackProgress)
	bestOutputs := job.evalutateTopN(proposedChannel, topCapture)

	trackProgress.Stop()
	return bestOutputs
}

func (job *MultiSetJob) evalutateTopN(proposedChannel <-chan multiProposedOutput, topCount uint64) []multiProposedOutput {
	// TODO make sure we have some of each combotype

	if len(job.specificAllowRates) == 0 {
		job.printer.Printf("COLLECTING TOP %d\n", topCount)
		bestChannel := channel_op.TransformAll_ChannelToChannel(evaluateThreadCount, proposedChannel, func(_ int, proposedChannel <-chan multiProposedOutput, bestChannel chan<- util_rank.HighestCollectorFloatN[multiProposedOutput]) {
			evalutateTopNWorker(proposedChannel, bestChannel, topCount)
		})
		return util_rank.HighestCollectorFloatN_OfChannel(bestChannel, evaluateThreadCount)
	} else {
		job.printer.Printf("COLLECTING TOP %d WITH SPLIT\n", topCount)
		bestChannel := channel_op.TransformAll_ChannelToChannel(evaluateThreadCount, proposedChannel, func(_ int, proposedChannel <-chan multiProposedOutput, bestChannel chan<- splitHighCollector) {
			evalutateTopNSplitWorker(proposedChannel, bestChannel, topCount, job.specificAllowRates)
		})
		return splitHighCollector_OfChannel(bestChannel, evaluateThreadCount)
	}
}

func evalutateTopNWorker(proposedChannel <-chan multiProposedOutput, bestChannel chan<- util_rank.HighestCollectorFloatN[multiProposedOutput], topCount uint64) {
	best := util_rank.HighestCollectorFloat_ForN(topCount, (*multiProposedOutput).Equals)
	for proposed := range proposedChannel {
		best.Offer(&proposed, proposed.totalRatingSum)
	}
	bestChannel <- best
}

func evalutateTopNSplitWorker(proposedChannel <-chan multiProposedOutput, bestChannel chan<- splitHighCollector, topCount uint64, specificAllowRates map[items.ItemId]specificAllowEntry) {
	best := splitHighCollector_make(specificAllowRates, topCount)
	for proposed := range proposedChannel {
		best.Offer(&proposed)
	}
	bestChannel <- best
}

type splitHighCollector struct {
	allowIds       []items.ItemId
	highCollectors []util_rank.HighestCollectorFloatN[multiProposedOutput]
}

func splitHighCollector_make(specificAllowRates map[items.ItemId]specificAllowEntry, topCount uint64) splitHighCollector {
	collector := splitHighCollector{}

	splitCount := []uint64{topCount}
	for itemId, entry := range specificAllowRates {
		percent := entry.proportion
		nextSplitCount := make([]uint64, 0, len(splitCount)*2)
		for _, count := range splitCount {
			trueCount := uint64(float32(count) * percent)
			falseCount := count - trueCount
			nextSplitCount = append(nextSplitCount, falseCount, trueCount)
		}
		splitCount = nextSplitCount

		collector.allowIds = append(collector.allowIds, itemId)
	}

	for _, count := range splitCount {
		if count <= 0 {
			panic("unexpected zero")
		}
		collector.highCollectors = append(collector.highCollectors, util_rank.HighestCollectorFloat_ForN(count, (*multiProposedOutput).Equals))
	}

	return collector
}

func splitHighCollector_OfChannel(channel <-chan splitHighCollector, expectNum int) []multiProposedOutput {
	var collector *splitHighCollector = nil
	for range expectNum {
		threadResult := <-channel
		if collector == nil {
			collector = &threadResult
		} else {
			collector.Merge_Mutating(&threadResult)
		}
	}
	return collector.ResultsFlat()
}

func (collector *splitHighCollector) Offer(output *multiProposedOutput) {
	choices := output.combo.allowChoices

	index := 0
	blockSize := len(collector.highCollectors) / 2

	for _, itemId := range collector.allowIds {
		itemChoice := choices[itemId]
		if itemChoice == true {
			index += blockSize
		}
		blockSize /= 2
	}

	collector.highCollectors[index].Offer(output, output.totalRatingSum)
}

func (collector *splitHighCollector) Merge_Mutating(other *splitHighCollector) {
	for i := range collector.highCollectors {
		collector.highCollectors[i].Merge_Mutating(&other.highCollectors[i])
	}
}

func (collector *splitHighCollector) ResultsFlat() []multiProposedOutput {
	result := []multiProposedOutput{}
	for i := range collector.highCollectors {
		subList := collector.highCollectors[i].ResultsFlat()
		result = append(result, subList...)
	}
	return result
}
