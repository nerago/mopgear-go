package multi

import (
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"sync"
	"time"
)

func (job *MultiSetJob) RunNoPermutations_AllCommonAlternates(extendedAlternates bool) {
	job.checkNoPermutations()
	job.prepareInitial()

	cancelGenerate := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancelGenerate)

	proposalChannel, expectedCount := job.proposalsAllCommonAlternates(cancelGenerate, extendedAlternates)

	additional := job.additionalProposalsFromSpecOptimum(cancelGenerate)
	proposalChannel = util_async.MixChannels(proposalChannel, additional)

	futureSimResultList, futureProposalList := job.proposalsToSimResult(proposalChannel, util.TrackProgress_Start(), expectedCount)

	proposalList, gotResult2 := futureProposalList.WaitForResult()
	simResultList, gotResult1 := futureSimResultList.WaitForResult()
	job.generalMultiReport(gotResult1 && gotResult2, proposalList, simResultList)
}

func (job *MultiSetJob) RunForSolutionsPerPermute(solutionsPerPermute int) {
	job.prepareInitial()

	cancelGenerate := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancelGenerate)

	tracker := util.TrackProgress_Start()
	defer tracker.SetDone()

	proposalChannel, expectedCount := job.proposalsUnderPermutation(solutionsPerPermute, cancelGenerate)

	additional := job.additionalProposalsFromSpecOptimum(cancelGenerate)
	proposalChannel = util_async.MixChannels(proposalChannel, additional)

	futureSimResultList, futureProposalList := job.proposalsToSimResult(proposalChannel, tracker, expectedCount)

	proposalList, gotResult2 := futureProposalList.WaitForResult()
	simResultList, gotResult1 := futureSimResultList.WaitForResult()
	job.generalMultiReport(gotResult1 && gotResult2, proposalList, simResultList)
}

func (job *MultiSetJob) RunCullingSets(targetSolutionCount int64, timeLimit time.Duration) {
	job.prepareInitial()

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(job.params))
	defer tracker.SetDone()

	cancel := util_async.CancelSignal_Make()
	timer := util_async.CancelAfterTimeout(cancel, timeLimit, job.printer)
	defer timer.Stop()

	waitGroup := sync.WaitGroup{}
	for param := range util.ForPointer(job.params) {
		param.runCullingProcess(targetSolutionCount, &waitGroup, cancel, tracker.NewChild())
	}

	waitGroup.Wait()

	job.CullingReport()
}

func (job *MultiSetJob) proposalsToSimResult(proposalChannel <-chan multi_types.MultiProposedOutput, tracker *util.TrackProgress, expectedCount *util_async.Future[int]) (*util_async.FutureCancellable[[]simulateJobResult], *util_async.Future[[]multi_types.MultiProposedOutput]) {
	proposalChannel = util_async.Channel_RemoveDuplicatesFuncNotify(proposalChannel, func(a, b *multi_types.MultiProposedOutput) bool {
		return a.Equals(b)
	}, func(x *multi_types.MultiProposedOutput) {
		job.printer.Printf("Remove Duplicate %s\n", x.Id)
	})

	proposalChannel = job.listInitialOutputs(proposalChannel)
	proposalChannel, futureProposalList := util_async.TeeChannelToSlice(proposalChannel)
	expectedCount = expectedCount.MapSameType(func(multiSetCount int) (int, bool) { return multiSetCount * len(job.params), true })

	simChannel := job.prepareSimList(proposalChannel)
	futureSimResultList := job.runSims(simChannel, tracker, expectedCount)
	return futureSimResultList, futureProposalList
}

func (job *MultiSetJob) generalMultiReport(gotResult bool, proposalList []multi_types.MultiProposedOutput, simResultList []simulateJobResult) {
	if gotResult {
		simMultiResults := job.linkSimResults(proposalList, simResultList)
		job.reportSimResults(simMultiResults)
		job.reportAsCsv(simMultiResults)

		job.suggestResultFromRankings(simMultiResults)
	} else {
		job.printer.Println("cancelled without result")
	}
}
