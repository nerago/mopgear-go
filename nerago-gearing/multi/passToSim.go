package multi

import (
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
)

type simulateJob struct {
	spec        stats.SpecType
	goal        stats.OptimiseGoal
	fight       stats.WowSim_Fight
	simSpeedUp  int
	equip       items.FullEquipMap
	professions gear_model.ProfessionInfo
}

type simulateJobResult struct {
	job    simulateJob
	result stats.SimData
}

type simulateMultiResult struct {
	proposed multi_types.MultiProposedOutput
	simMap   map[string]stats.SimData
}

func (job *MultiSetJob) runSimsForProposalChannel(proposalChannel <-chan multi_types.MultiProposedOutput, tracker *util.TrackProgress, expectedCount <-chan int) (*util_async.FutureCancellable[[]simulateJobResult], *util_async.Future[[]multi_types.MultiProposedOutput]) {
	proposalChannel = util_async.Channel_RemoveDuplicatesFuncNotify(proposalChannel, func(a, b *multi_types.MultiProposedOutput) bool {
		return a.Equals(b)
	}, func(x *multi_types.MultiProposedOutput) {
		job.printer.Printf("Remove Duplicate %s\n", x.Id)
	})

	proposalChannel = job.listInitialOutputs(proposalChannel)
	proposalChannel, futureProposalList := util_async.TeeChannelToSlice(proposalChannel)
	simChannel := job.prepareSimList(proposalChannel)

	simsPerProposal := len(job.itemPrep)
	expectedCount = util_async.Map_ChannelToChannel(1, expectedCount, func(x int) int { return x * simsPerProposal })

	futureSimResultList := job.runSims(simChannel, tracker, expectedCount)
	return futureSimResultList, futureProposalList
}

func checkNoConflicts(outputSet map[string]multi_types.SingleProposedOutput, printer *util.PrintRecorder) bool {
	itemByRef := make(map[items.ItemRef]*items.FullItem)
	for _, single := range outputSet {
		for item := range single.FullSet.Items().AllItemSeq() {
			ref := items.ItemRef_Of(item)
			existing, found := itemByRef[ref]
			if !found {
				itemByRef[ref] = item
			} else if !existing.Equals(item) {
				printer.Printf("!! CONFLICT %s\n!!          %s\n", item.CreateString(), existing.CreateString())
				return false
			}
		}
	}
	return true
}

func (simJob *simulateJob) equalForLink(part *multi_types.SingleProposedOutput, prep *specItemPrep) bool {
	return simJob.spec == part.Spec &&
		simJob.goal == prep.model.Goal &&
		simJob.fight == prep.model.SimulateAs &&
		simJob.professions == prep.model.Professions &&
		simJob.equip.Equals(part.FullSet.Items())
}

func (simJob *simulateJob) Equals(other *simulateJob) bool {
	return simJob.spec == other.spec &&
		simJob.goal == other.goal &&
		simJob.fight == other.fight &&
		simJob.professions == other.professions &&
		simJob.equip.Equals(&other.equip)
}

func (job *MultiSetJob) prepareSimList(proposalList <-chan multi_types.MultiProposedOutput) <-chan simulateJob {
	jobChannel := util_async.MapMulti_ChannelToChannel(1, proposalList, func(proposal multi_types.MultiProposedOutput, nextChan chan<- simulateJob) {
		for _, single := range proposal.Parts {
			label := single.SpecLabel
			prep := job.itemPrep[label]
			nextChan <- simulateJob{
				spec:        single.Spec,
				goal:        prep.model.Goal,
				fight:       prep.model.SimulateAs,
				simSpeedUp:  prep.model.SimSpeedUp,
				professions: prep.model.Professions,
				equip:       *single.FullSet.Items()}
		}
	})

	return util_async.Channel_RemoveDuplicatesFunc(jobChannel, (*simulateJob).Equals)
}

func (job *MultiSetJob) runSims(jobChan <-chan simulateJob, trackProgress *util.TrackProgress, expectedCount <-chan int) *util_async.FutureCancellable[[]simulateJobResult] {
	trackProgress.RunOuterTracking(0)

	go func() {
		for newCount := range expectedCount {
			trackProgress.UpdateExpectedChildCount(newCount)
		}
	}()

	return util_async.Map_ChannelToSlice_FutureCancellable(c_simThreadCount, jobChan, trackProgress.SetDone, func(sim simulateJob) simulateJobResult {
		result := simulate.WowSim_Execute_SpecifyAll(job.input.SimRunSize, sim.simSpeedUp, sim.spec, sim.goal, sim.fight, sim.professions, &sim.equip, nil, trackProgress.NewChild())
		job.printer.Printf("sim %22s fight=%d %s\n", sim.spec.Name(), sim.fight, result.CompactStringGeneral())
		return simulateJobResult{sim, result}
	})
}

func (job *MultiSetJob) linkSimResults(proposalList []multi_types.MultiProposedOutput, jobList []simulateJobResult) []simulateMultiResult {
	resultList := make([]simulateMultiResult, 0, len(proposalList))
	for _, proposal := range proposalList {
		result := job.linkSimResult(proposal, jobList)
		resultList = append(resultList, result)
	}
	return resultList
}

func (job *MultiSetJob) linkSimResult(proposal multi_types.MultiProposedOutput, resultList []simulateJobResult) simulateMultiResult {
	multiResult := simulateMultiResult{proposal, make(map[string]stats.SimData, len(proposal.Parts))}
	for label, part := range proposal.Parts {
		prep := job.itemPrep[label]
		for simResult := range util_collection.ForPointer(resultList) {
			if simResult.job.equalForLink(&part, prep) {
				multiResult.simMap[label] = simResult.result
				break
			}
		}
	}
	return multiResult
}

func (job *MultiSetJob) writeToGearFiles(result *simulateMultiResult) {
	for label, prep := range job.itemPrep {
		itemSet := result.proposed.Parts[label].FullSet
		gearJson := tools.WowSimJson_Write(itemSet.Items(), &prep.model, util.PrintRecorder_Nop())

		gearFile := prep.inputs.GearFile
		util.WriteStringToFile(gearFile, gearJson)
	}
}
