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

	setResultChan := highProcess.RunForSeveral_CommonDifferent(job.printer, util.Optional_Empty[int]())

	tracker := util.TrackProgress_Start()
	defer tracker.Stop()

	proposalChannel := make(chan multi_types.MultiProposedOutput, 8)
	proposalChannel <- job.existingGearAsProposal()
	channel_op.Map_ChannelToChannel_Provided(4, setResultChan, proposalChannel, func(setResult withhighs.HighsMultiResult, next chan<- multi_types.MultiProposedOutput) {
		next <- job.makeOutputFromHighs(setResult, job.printer)
	})

	// TODO tracker covers highs part too
	job.proposalsToSimAndOutput(proposalChannel, tracker)

	job.CullingReport()
}

func (job *MultiSetJob) RunForSolutionsPerPerumte(solutionsPerPermute int) {
	job.prepareInitial()

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(2)
	defer tracker.Stop()

	setResultChannel := job.proposalsUnderPermutation(tracker.MakeNested(), solutionsPerPermute)

	job.proposalsToSimAndOutput(setResultChannel, tracker.MakeNested())

	job.CullingReport()
}

func (job *MultiSetJob) RunCullingSets(targetSolutionCount int64, timeLimit time.Duration) {
	job.prepareInitial()

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(job.params))
	defer tracker.Stop()

	timer := time.AfterFunc(timeLimit, func() {
		job.printer.Println("###################### TIME LIMIT EXPIRED ######################")
		tracker.Stop()
	})
	defer timer.Stop()

	waitGroup := sync.WaitGroup{}
	for param := range util.ForPointer(job.params) {
		param.runCullingProcess(targetSolutionCount, &waitGroup, tracker.MakeNested())
	}

	waitGroup.Wait()

	job.CullingReport()
}
