package multi

import (
	"sync"
	"time"

	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
)

func (job *MultiSetJob) Run() {
	job.prepareItems()

	cancelGenerate := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancelGenerate)

	groupChannel := job.prepareWorkingGroups(cancelGenerate)

	proposalChannel, expectedCountChannel := job.makeProposalChannel(groupChannel, cancelGenerate)

	proposalChannel, futureProposalList := util_async.TeeChannelToSlice(proposalChannel)

	tracker := util.TrackProgress_Start()
	defer tracker.SetDone()

	futureSimResultList := job.runSimsForProposalChannel(proposalChannel, tracker, expectedCountChannel)

	proposalList, gotResult2 := futureProposalList.WaitForResult()
	simResultList, gotResult1 := futureSimResultList.WaitForResult()
	job.generalMultiReport(gotResult1 && gotResult2, proposalList, simResultList)
}

func (job *MultiSetJob) makeProposalChannel(groupChannel <-chan *workingGroup, cancelGenerate *util_async.CancelSignalBasic) (<-chan multi_types.MultiProposedOutput, <-chan int) {
	futureCount := util_async.FutureValueAdderIntMake(0)
	proposalMixer := util_async.FutureChannelMixerContinuing[multi_types.MultiProposedOutput]{}

	util_async.ForEach_Channel_NonBlocking(c_mainProposal_threadCount, groupChannel, func(group *workingGroup) {
		group.groupProposals(&proposalMixer, futureCount, cancelGenerate)
	}, func() {
		proposalMixer.ShutdownAsync()
	})

	expectedCountChannel := futureCount.ReadyUpAndPrepareChannel()
	proposalChannel := proposalMixer.ReadyUpAndPrepareChannel()

	proposalChannel = util_async.Channel_RemoveDuplicatesFuncNotify(proposalChannel, func(a, b *multi_types.MultiProposedOutput) bool {
		return a.Equals(b)
	}, func(x *multi_types.MultiProposedOutput) {
		job.printer.Printf("Remove Duplicate %s\n", x.Id)
	})

	proposalChannel = job.listInitialOutputs(proposalChannel)
	return proposalChannel, expectedCountChannel
}

func (job *MultiSetJob) RunCullingSets(targetSolutionCount int64, timeLimit time.Duration) {
	job.prepareItems()

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(job.workGroups) * len(job.itemPrep))
	defer tracker.SetDone()

	cancel := util_async.CancelSignal_Make()
	timer := util_async.CancelAfterTimeout(cancel, timeLimit, job.printer)
	defer timer.Stop()

	groupChannel := job.prepareWorkingGroups(cancel)

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
