package multi

import (
	"sync"
	"time"

	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
)

func (job *MainJob) Run() {
	err := job.prepareItems()
	if err != nil {
		util.GlobalFatalErrorHandler(err)
	}

	cancelGenerate := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancelGenerate)

	groupChannel := job.prepareWorkingGroups(cancelGenerate)

	proposalChannel, expectedCountChannel, possibleProposalErrors := job.makeProposalChannel(groupChannel, cancelGenerate)

	simJobChannel1, resultPendingChan := job.prepareSimList(proposalChannel)
	simJobChannel2 := job.mergeDuplicateJobs(simJobChannel1)

	simsDone := job.runSims(simJobChannel2, expectedCountChannel)

	simMultiResults := job.incrementalReporting(resultPendingChan, simsDone)
	job.finalMultiReport(simMultiResults, possibleProposalErrors)
}

func (job *MainJob) mergeDuplicateJobs(simJobChannel1 <-chan *simulateJobPending) <-chan *simulateJobPending {
	return util_async.Channel_RemoveDuplicatesFuncNotify(
		simJobChannel1,
		func(a, b *simulateJobPending) bool {
			return (*a).simJob.Equals(&(*b).simJob)
		},
		func(keep, drop *simulateJobPending) {
			keep.MergePendingFrom(drop)
		},
	)
}

func (job *MainJob) makeProposalChannel(groupChannel <-chan *workingGroup, cancelGenerate *util_async.CancelSignalBasic) (<-chan *multi_types.MultiProposedOutput, <-chan int, *util_async.PossibleFutureError) {
	futureCount := util_async.FutureValueAdderIntMake(0)
	proposalMixer := util_async.FutureChannelMixerContinuing[*multi_types.MultiProposedOutput]{}
	possibleError := util_async.PossibleFutureErrorMake()

	util_async.ForEach_Channel_NonBlocking(c_mainProposal_threadCount, groupChannel, func(group *workingGroup) {
		group.groupProposals(&proposalMixer, futureCount, possibleError, cancelGenerate)
	}, func() {
		proposalMixer.ShutdownAsync(func() {
			job.printer.Println("<<< PROPOSALS ALL SUBMITTED >>>")
		})
	})

	expectedCountChannel := futureCount.ReadyUpAndPrepareChannel()
	proposalChannel := proposalMixer.ReadyUpAndPrepareChannel()

	proposalChannel = util_async.Channel_RemoveDuplicatesFuncNotify(proposalChannel, func(a, b *multi_types.MultiProposedOutput) bool {
		return a.Equals(b)
	}, func(keep, drop *multi_types.MultiProposedOutput) {
		job.printer.Printf("Remove Duplicate %s for %s\n", drop.Id, keep.Id)
	})

	proposalChannel = job.listInitialOutputs(proposalChannel)
	return proposalChannel, expectedCountChannel, possibleError
}

func (job *MainJob) RunCullingSets(targetSolutionCount int64, timeLimit time.Duration) {
	err := job.prepareItems()
	if err != nil {
		util.GlobalFatalErrorHandler(err)
	}

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(job.workGroups) * len(job.itemPrep))
	defer tracker.SetDone()

	cancel := util_async.CancelSignal_Make()
	timer := util_async.CancelAfterTimeout(cancel, timeLimit, job.printer)
	defer timer.Stop()

	groupChannel := job.prepareWorkingGroups(cancel)

	waitGroup := sync.WaitGroup{}
	util_async.ForEach_Channel(6, groupChannel, func(group *workingGroup) {
		for _, work := range group.workers {
			work.runCullingProcess(targetSolutionCount, &waitGroup, cancel, tracker.NewChild(), job.printer)
		}
	})

	waitGroup.Wait()

	job.CullingReport()
}

func (job *MainJob) finalMultiReport(simMultiResults []simulateMultiResult, possibleErrors *util_async.PossibleFutureError) {
	if errors, hasErrors := possibleErrors.GetResultNoWait(); hasErrors {
		job.printer.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		job.printer.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		job.printer.Println(errors.Error())
		job.printer.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		job.printer.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}

	if len(simMultiResults) > 0 {
		job.reportSimResults(simMultiResults, true)

		best, rankedData := job.rankAllResults(simMultiResults)
		job.reportAsCsv(rankedData)

		job.handleBestRankedResult(best)
	} else {
		job.printer.Println("cancelled without result")
	}
}
