package multi

import (
	"maps"
	"sync"

	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
)

type simulateJob struct {
	spec        stats.SpecType
	goal        stats.OptimiseGoal
	fight       stats.WowSim_Fight
	simSpeedUp  int
	equip       items.FullEquipMap
	professions gear_model.ProfessionInfo
}

type simulateJobPending struct {
	simJob        simulateJob
	resultPending *stats.SimData
	multiPending  []*simulateMultiResultPending
	mutex         sync.Mutex
	merged        bool
}

func (sjp *simulateJobPending) SetResult(result *stats.SimData) {
	sjp.mutex.Lock()
	defer sjp.mutex.Unlock()

	if sjp.merged {
		panic("don't expect result for on merged result")
	}

	sjp.resultPending = result
	for _, r := range sjp.multiPending {
		r.notifyOnResult()
	}
	sjp.multiPending = nil
}

func (sjp *simulateJobPending) MergePendingFrom(other *simulateJobPending) {
	other.mutex.Lock()
	defer other.mutex.Unlock()
	sjp.mutex.Lock()
	defer sjp.mutex.Unlock()

	for _, om := range other.multiPending {
		om.mutex.Lock()
		for k, oj := range om.simPending {
			if oj == other {
				om.simPending[k] = sjp
			}
		}
		om.mutex.Unlock()
	}

	if sjp.resultPending != nil {
		for _, r := range other.multiPending {
			r.notifyOnResult()
		}
	} else {
		sjp.multiPending = append(sjp.multiPending, other.multiPending...)
	}

	other.multiPending = nil
	other.merged = true
}

type simulateMultiResultPending struct {
	proposed   *multi_types.MultiProposedOutput
	simPending map[string]*simulateJobPending
	mutex      sync.Mutex
	ready      bool
}

func (m *simulateMultiResultPending) notifyOnResult() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	allReady := true
	for _, p := range m.simPending {
		if p.resultPending == nil {
			allReady = false
		}
	}
	if allReady {
		m.ready = true
	}
}

type simulateMultiResult struct {
	proposed *multi_types.MultiProposedOutput
	simMap   map[string]*stats.SimData
}

func (r simulateMultiResult) Equals(b *simulateMultiResult) bool {
	return r.proposed.Equals(b.proposed) &&
		maps.EqualFunc(r.simMap, b.simMap, func(x stats.SimData, y stats.SimData) bool {
			return x.Equals(&y)
		})
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

func (job *MultiSetJob) prepareSimList(proposalList <-chan *multi_types.MultiProposedOutput) (<-chan *simulateJobPending, <-chan *simulateMultiResultPending) {
	simJobChan := make(chan *simulateJobPending)
	resultPendingChan := make(chan *simulateMultiResultPending)
	util_async.ForEach_Channel_NonBlocking(1, proposalList, func(proposal *multi_types.MultiProposedOutput) {
		proposalPending := &simulateMultiResultPending{
			proposed:   proposal,
			simPending: make(map[string]*simulateJobPending),
		}

		for label, single := range proposal.Parts {
			simJob := job.createPendingJob(label, single, proposalPending)
			simJobChan <- simJob
		}

		resultPendingChan <- proposalPending
	}, func() {
		close(simJobChan)
		close(resultPendingChan)
	})
	return simJobChan, resultPendingChan
}

func (job *MultiSetJob) createPendingJob(label string, single multi_types.SingleProposedOutput, proposalPending *simulateMultiResultPending) *simulateJobPending {
	prep := job.itemPrep[label]
	simJob := &simulateJobPending{
		simJob: job.createSimJob(single, prep),
	}
	proposalPending.simPending[label] = simJob
	simJob.multiPending = append(simJob.multiPending, proposalPending)
	return simJob
}

func (job *MultiSetJob) createSimJob(single multi_types.SingleProposedOutput, prep *specItemPrep) simulateJob {
	return simulateJob{
		spec:        single.Spec,
		goal:        prep.model.Goal,
		fight:       prep.model.SimulateAs,
		simSpeedUp:  prep.model.SimSpeedUp,
		professions: prep.model.Professions,
		equip:       *single.FullSet.Items(),
	}
}

func (job *MultiSetJob) runSims(jobChan <-chan *simulateJobPending, expectedCount <-chan int) *util_async.FutureVoid {
	simFinished := util_async.FutureVoid_Make()

	simsPerProposal := len(job.itemPrep)
	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(0)
	go func() {
		for newCount := range expectedCount {
			tracker.UpdateExpectedChildCount(newCount * simsPerProposal)
		}
	}()

	util_async.ForEach_Channel_NonBlocking(c_simThreadCount, jobChan, func(pending *simulateJobPending) {
		sim := pending.simJob
		result := simulate.WowSim_Execute_SpecifyAll(job.input.SimRunSize, sim.simSpeedUp, sim.spec, sim.goal, sim.fight,
			sim.professions, &sim.equip, nil, tracker.NewChild())
		job.printer.Printf("sim %s fight=%d %s\n", sim.spec.Name(), sim.fight, result.CompactStringGeneral())
		pending.SetResult(new(result))
	}, func() {
		tracker.SetDone()
		simFinished.SetResultEmpty()
	})

	return simFinished
}

func (job *MultiSetJob) writeToGearFiles(result *simulateMultiResult) {
	for label, prep := range job.itemPrep {
		itemSet := result.proposed.Parts[label].FullSet
		gearJson := tools.WowSimJson_Write(itemSet.Items(), prep.model, util.PrintRecorder_Nop())

		gearFile := prep.inputs.GearFile
		util.WriteStringToFile(gearFile, gearJson)
	}
}
