package multi

import (
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"sync"
	"time"
)

func (job *MultiSetJob) RunNoPermutations_AllCommonAlternates() {
	job.checkNoPermutations()
	job.prepareInitial()
	highProcess := job.highProcessSetup()

	cancel := channel_op.CancelSignal_Make()

	setResultChan := highProcess.RunForSeveral_CommonDifferent(job.printer, util.Optional_Empty[int](), cancel)

	tracker := util.TrackProgress_Start()
	defer tracker.SetDone()

	proposalChannel := channel_op.Map_ChannelToChannel(4, setResultChan, func(setResult withhighs.HighsMultiResult) multi_types.MultiProposedOutput {
		return job.makeOutputFromHighs(setResult, job.printer)
	})

	existingProposal := job.existingGearAsProposal()
	combinedProposalChannel := channel_op.ChannelWithPrependedValues(proposalChannel, existingProposal)

	// TODO tracker covers highs part too
	job.proposalsToSimAndOutput(combinedProposalChannel, tracker, cancel)

	// job.CullingReport()
}

func (job *MultiSetJob) RunForSolutionsPerPerumte(solutionsPerPermute int) {
	job.prepareInitial()

	cancel := channel_op.CancelSignal_Make()
	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(2)
	defer tracker.SetDone()

	setResultChannel := job.proposalsUnderPermutation(tracker.NewChild(), solutionsPerPermute, cancel)

	job.proposalsToSimAndOutput(setResultChannel, tracker.NewChild(), cancel)

	// job.CullingReport()
}

func (job *MultiSetJob) RunCullingSets(targetSolutionCount int64, timeLimit time.Duration) {
	job.prepareInitial()

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(job.params))
	defer tracker.SetDone()

	cancel := channel_op.CancelSignal_Make()

	timer := time.AfterFunc(timeLimit, func() {
		job.printer.Println("###################### TIME LIMIT EXPIRED ######################")
		tracker.SetDone()
	})
	defer timer.Stop()

	waitGroup := sync.WaitGroup{}
	for param := range util.ForPointer(job.params) {
		param.runCullingProcess(targetSolutionCount, &waitGroup, cancel, tracker.NewChild())
	}

	waitGroup.Wait()

	job.CullingReport()
}
