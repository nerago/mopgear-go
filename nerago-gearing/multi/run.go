package multi

import (
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"sync"
	"time"
)

func (job *MultiSetJob) RunNoPermutations_BestOnly(alsoExistingEquipped bool, alsoSpecOptimums bool) {
	job.checkNoPermutations()
	job.prepareItems()
	job.prepareWorking()

	cancelGenerate := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancelGenerate)

	proposalMix := util_async.FutureChannelMixerMake[multi_types.MultiProposedOutput]()
	expectedCount := 0

	for _, weightType := range job.input.WeightTypeList {
		proposalFuture := job.proposalSingleBest(weightType)
		util_async.ChainCancel(cancelGenerate, proposalFuture)

		proposalMix.AddFuture(proposalFuture)
		expectedCount++
	}

	if alsoExistingEquipped {
		for _, weightType := range job.input.WeightTypeList {
			existingProposal := job.existingGearAsProposal(weightType)
			proposalMix.AddValue(existingProposal)
			expectedCount++
		}
	}

	if alsoSpecOptimums {
		additionalChannel, additionalCount := job.additionalProposalsFromSpecOptimalBaseline(cancelGenerate)
		proposalMix.AddChannel(additionalChannel)
		expectedCount += additionalCount
	}

	expectedCountChannel := make(chan int, 1)
	expectedCountChannel <- expectedCount

	proposalChannel := proposalMix.ReadyUpAndPrepareChannel()
	futureSimResultList, futureProposalList := job.proposalsToSimResult(proposalChannel, util.TrackProgress_Start(), expectedCountChannel)

	proposalList, gotResult2 := futureProposalList.WaitForResult()
	simResultList, gotResult1 := futureSimResultList.WaitForResult()
	job.generalMultiReport(gotResult1 && gotResult2, proposalList, simResultList)
}

func (job *MultiSetJob) RunNoPermutations_AllCommonAlternates(extendedAlternates bool, includeInterimResults bool) {
	job.checkNoPermutations()
	job.prepareItems()
	job.prepareWorking()

	cancelGenerate := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancelGenerate)

	proposalMixer, futureCountAdder := job.proposalsAllCommonAlternates(cancelGenerate, extendedAlternates, includeInterimResults)

	additionalChannel, additionalCount := job.additionalProposalsFromSpecOptimalBaseline(cancelGenerate)
	proposalMixer.AddChannel(additionalChannel)
	futureCountAdder.AddValueImmediate(func(x int) int { return x + additionalCount })

	proposalChannel := proposalMixer.ReadyUpAndPrepareChannel()
	expectedCount := futureCountAdder.ReadyUpAndPrepareChannel()
	futureSimResultList, futureProposalList := job.proposalsToSimResult(proposalChannel, util.TrackProgress_Start(), expectedCount)

	proposalList, gotResult2 := futureProposalList.WaitForResult()
	simResultList, gotResult1 := futureSimResultList.WaitForResult()
	job.generalMultiReport(gotResult1 && gotResult2, proposalList, simResultList)
}

func (job *MultiSetJob) RunForSolutionsPerPermute(solutionsPerPermute int, includeInterimResults bool) {
	job.prepareItems()
	job.prepareWorking()

	cancelGenerate := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancelGenerate)

	tracker := util.TrackProgress_Start()
	defer tracker.SetDone()

	proposalChannel, expectedCountAdder := job.proposalsUnderPermutation(solutionsPerPermute, includeInterimResults, cancelGenerate)

	additionalChannel, additionalCount := job.additionalProposalsFromSpecOptimalBaseline(cancelGenerate)
	proposalChannel = util_async.MixChannels(proposalChannel, additionalChannel)
	expectedCountAdder.AddValueImmediate(additionalCount)

	expectedCountChannel := expectedCountAdder.ReadyUpAndPrepareChannel()
	futureSimResultList, futureProposalList := job.proposalsToSimResult(proposalChannel, tracker, expectedCountChannel)

	proposalList, gotResult2 := futureProposalList.WaitForResult()
	simResultList, gotResult1 := futureSimResultList.WaitForResult()
	job.generalMultiReport(gotResult1 && gotResult2, proposalList, simResultList)
}

func (job *MultiSetJob) RunCullingSets(targetSolutionCount int64, timeLimit time.Duration) {
	job.prepareItems()
	job.prepareWorking()

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(job.working.Size())
	defer tracker.SetDone()

	cancel := util_async.CancelSignal_Make()
	timer := util_async.CancelAfterTimeout(cancel, timeLimit, job.printer)
	defer timer.Stop()

	waitGroup := sync.WaitGroup{}
	for work := range job.working.SeqValues() {
		work.runCullingProcess(targetSolutionCount, &waitGroup, cancel, tracker.NewChild(), job.printer)
	}

	waitGroup.Wait()

	job.CullingReport()
}

func (job *MultiSetJob) generalMultiReport(gotResult bool, proposalList []multi_types.MultiProposedOutput, simResultList []simulateJobResult) {
	if gotResult {
		simMultiResults := job.linkSimResults(proposalList, simResultList)
		job.reportSimResults(simMultiResults)

		best, rankedData := job.rankAllResults(simMultiResults)
		job.reportAsCsv(rankedData)

		job.handleBestRankedResult(best)
	} else {
		job.printer.Println("cancelled without result")
	}
}
