package multi

import (
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"sync"
	"time"
)

func (job *MultiSetJob) RunNoPermutations_BestOnly(alsoExistingEquipped bool, alsoSpecOptimums bool) {
	job.checkNoPermutations()
	job.prepareItems()
	groupChannel := job.prepareWorkingGroups()

	cancelGenerate := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancelGenerate)

	proposalChannel := util_async.MapToChannel_ChannelToChannel_Cancellable(c_mainProposal_threadCount, groupChannel, cancelGenerate, func(group *workingGroup) <-chan multi_types.MultiProposedOutput {
		mixer := util_async.FutureChannelMixerStatic[multi_types.MultiProposedOutput]{}
		mixer.AddFutureCancellable(group.proposalSingleBest())
		if alsoExistingEquipped {
			mixer.AddValue(group.existingGearAsProposal())
		}
		if alsoSpecOptimums {
			mixer.AddChannel(group.additionalProposalsFromSpecOptimalBaseline(cancelGenerate))
		}
		return mixer.ReadyUpAndPrepareChannel()
	})

	expectedCount := len(job.input.WeightTypeList)
	if alsoExistingEquipped {
		expectedCount += len(job.input.WeightTypeList)
	}
	if alsoSpecOptimums {
		expectedCount += len(job.input.Param) * len(job.input.WeightTypeList)
	}
	expectedCountChannel := make(chan int, 1)
	expectedCountChannel <- expectedCount

	tracker := util.TrackProgress_Start()
	defer tracker.SetDone()

	futureSimResultList, futureProposalList := job.runSimsForProposalChannel(proposalChannel, tracker, expectedCountChannel)

	proposalList, gotResult2 := futureProposalList.WaitForResult()
	simResultList, gotResult1 := futureSimResultList.WaitForResult()
	job.generalMultiReport(gotResult1 && gotResult2, proposalList, simResultList)
}

func (job *MultiSetJob) RunNoPermutations_AllCommonAlternates(extendedAlternates bool, includeInterimResults bool) {
	job.checkNoPermutations()
	job.prepareItems()
	groupChannel := job.prepareWorkingGroups()

	cancelGenerate := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancelGenerate)

	futureCount := util_async.FutureValueAdderIntMake(0)

	proposalChannel := util_async.MapToChannel_ChannelToChannel_Cancellable(c_mainProposal_threadCount, groupChannel, cancelGenerate, func(group *workingGroup) <-chan multi_types.MultiProposedOutput {
		proposalMixer := group.proposalsAllCommonAlternates(cancelGenerate, futureCount, extendedAlternates, includeInterimResults)

		additionalChannel := group.additionalProposalsFromSpecOptimalBaseline(cancelGenerate)
		futureCount.AddValueImmediate(len(group.workers))
		proposalMixer.AddChannel(additionalChannel)

		return proposalMixer.ReadyUpAndPrepareChannel()
	})

	tracker := util.TrackProgress_Start()
	defer tracker.SetDone()

	expectedCount := futureCount.ReadyUpAndPrepareChannel()
	futureSimResultList, futureProposalList := job.runSimsForProposalChannel(proposalChannel, tracker, expectedCount)

	proposalList, gotResult2 := futureProposalList.WaitForResult()
	simResultList, gotResult1 := futureSimResultList.WaitForResult()
	job.generalMultiReport(gotResult1 && gotResult2, proposalList, simResultList)
}

func (job *MultiSetJob) RunForSolutionsPerPermute(solutionsPerPermute int, includeInterimResults bool) {
	job.prepareItems()
	groupChannel := job.prepareWorkingGroups()

	cancelGenerate := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancelGenerate)

	futureCount := util_async.FutureValueAdderIntMake(0)

	proposalChannel := util_async.MapToChannel_ChannelToChannel_Cancellable(c_mainProposal_threadCount, groupChannel, cancelGenerate, func(group *workingGroup) <-chan multi_types.MultiProposedOutput {
		nestedMixer := group.proposalsUnderPermutation(solutionsPerPermute, includeInterimResults, futureCount, cancelGenerate)

		additionalChannel := group.additionalProposalsFromSpecOptimalBaseline(cancelGenerate)
		nestedMixer.AddChannel(additionalChannel)

		return nestedMixer.ReadyUpAndPrepareChannel()
	})

	tracker := util.TrackProgress_Start()
	defer tracker.SetDone()

	expectedCountChannel := futureCount.ReadyUpAndPrepareChannel()
	futureSimResultList, futureProposalList := job.runSimsForProposalChannel(proposalChannel, tracker, expectedCountChannel)

	proposalList, gotResult2 := futureProposalList.WaitForResult()
	simResultList, gotResult1 := futureSimResultList.WaitForResult()
	job.generalMultiReport(gotResult1 && gotResult2, proposalList, simResultList)
}

func (job *MultiSetJob) RunCullingSets(targetSolutionCount int64, timeLimit time.Duration) {
	job.prepareItems()
	groupChannel := job.prepareWorkingGroups()

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(job.workGroups) * len(job.itemPrep))
	defer tracker.SetDone()

	cancel := util_async.CancelSignal_Make()
	timer := util_async.CancelAfterTimeout(cancel, timeLimit, job.printer)
	defer timer.Stop()

	waitGroup := sync.WaitGroup{}
	util_async.ForEach_Channel(2, groupChannel, func(group *workingGroup) {
		for _, work := range group.workers {
			work.runCullingProcess(targetSolutionCount, &waitGroup, cancel, tracker.NewChild(), job.printer)
		}
	})

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
