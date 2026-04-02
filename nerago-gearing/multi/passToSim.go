package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"strconv"
	"strings"
)

func (job *MultiSetJob) FindTopAndPassToSim(targetCount uint64, topCapture int, runSize simulate.WowSim_RunSize) {
	topTracker := util.TrackProgress_Start()
	topTracker.RunOuterTracking(3)
	defer topTracker.Stop()

	job.printer.Printf("@@@@@@@@@@ FIND TOP %d %d @@@@@@@@@@\n", targetCount, topCapture)
	bestOutputs := job.runForTopN(targetCount, topCapture, topTracker.MakeNested())
	job.listInitialOutputs(bestOutputs)

	proposalList := job.prepareRevisionsForSim(bestOutputs, topTracker.MakeNested())

	simList := job.prepareSimList(proposalList)
	job.runSims(simList, runSize, topTracker.MakeNested())

	simResult := job.linkSimResults(proposalList, simList)
	job.reportSimResults(simResult)
	job.reportAsCsv(simResult)
}

func (job *MultiSetJob) prepareRevisionsForSim(proposedList []MultiProposedOutput, trackProgress *util.TrackProgress) []MultiProposedOutput {
	job.printer.Printf("@@@@@@@@@@ MAKE REVISIONS FOR %d @@@@@@@@@@\n", len(proposedList))

	expectedSets := len(proposedList) * len(job.params) * revisedExtraSetsExpectedEach
	trackProgress.RunOuterTracking(expectedSets)
	defer trackProgress.Stop()

	allProposals := channel_op.IterateEach_SliceToSlice(generateThreadCount, proposedList, func(prior *MultiProposedOutput, downstream chan<- MultiProposedOutput) {
		printer := util.PrintRecorder_HoldAll()
		printer.Printf(">>> PREP REVISIONS %s\n", prior.Id)

		revisedCommon := job.revisedComboActuallyUsed(prior.Outputs, prior.Combo, printer)

		revisedOptionArrays := make([][]solver.SolveOutput, len(prior.Outputs))

		for i := range prior.Outputs {
			draft := &prior.Outputs[i]
			param := &job.params[i]

			printer.Println("DRAFT")
			draft.Report(printer)

			specOptions := job.makeRevised(param, revisedCommon, trackProgress, printer)
			for _, newOutput := range specOptions {
				param.seenInSolutions.Add(&newOutput.FullSet)
			}

			param.seenInSolutions.Add(&draft.FullSet)
			specOptions = append(specOptions, *draft)

			specOptions = util.RemoveDuplicatesFuncNotify(specOptions, func(a, b *solver.SolveOutput) bool {
				return a.FullSet.Equals(&b.FullSet)
			}, func(removed *solver.SolveOutput) {
				printer.Printf("removed duplicate output %s\n", removed.OutputId)
			})

			revisedOptionArrays[i] = specOptions
		}

		for outputSet := range util.PermuteAll(revisedOptionArrays) {
			var totalRatingSum uint64
			for _, output := range outputSet {
				totalRatingSum += output.ResultRating
			}
			if checkNoConflicts(outputSet, printer) {
				proposed := MultiProposedOutput{makeUUID(), totalRatingSum, outputSet, revisedCommon}
				componentIds := ""
				for _, set := range outputSet {
					componentIds = componentIds + set.OutputId + " "
				}
				printer.Printf("&&& NEW PROPOSAL %s => %s\n", proposed.Id, componentIds)
				downstream <- proposed
			}
		}

		job.printer.AppendOther(printer)
	})

	allProposals = append(allProposals, job.existingGearAsProposal())
	return allProposals
}

func checkNoConflicts(outputSet []solver.SolveOutput, printer *util.PrintRecorder) bool {
	itemById := make(map[items.ItemId]*items.FullItem)
	for outputIndex := range outputSet {
		for item := range outputSet[outputIndex].FullSet.Items().AllItemSeq() {
			existing, found := itemById[item.ItemId()]
			if !found {
				itemById[item.ItemId()] = item
			} else if !existing.Equals(item) {
				// printer.Printf("!! CONFLICT %s\n!!          %s\n", item.CreateString(), existing.CreateString())
				return false
			}
		}
	}
	return true
}

func (job *MultiSetJob) existingGearAsProposal() MultiProposedOutput {
	proposal := MultiProposedOutput{Id: makeUUID()}
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		proposal.Outputs = append(proposal.Outputs, param.baselineResult)
		proposal.TotalRatingSum += param.baselineResult.ResultRating
		proposal.Combo = job.revisedComboActuallyUsed(proposal.Outputs, make(CommonCombo), util.PrintRecorder_HoldAll())
	}
	return proposal
}

type simulateJob struct {
	spec   stats.SpecType
	equip  items.FullEquipMap
	model  *model.Model
	result *simulate.SimResultStats
}

type simulateResult struct {
	proposed MultiProposedOutput
	result   []simulate.SimResultStats
}

func (job *MultiSetJob) prepareSimList(proposalList []MultiProposedOutput) []simulateJob {
	jobList := make([]simulateJob, 0)
	for _, proposal := range proposalList {
		for _, output := range proposal.Outputs {
			job := simulateJob{output.Input.Model.Spec, *output.FullSet.Items(), output.Input.Model, nil}
			jobList = append(jobList, job)
		}
	}

	jobList = util.RemoveDuplicatesComparable(jobList)

	return jobList
}

func (job *MultiSetJob) runSims(jobList []simulateJob, runSize simulate.WowSim_RunSize, trackProgress *util.TrackProgress) {
	job.printer.Printf("@@@@@@@@@@ RUN SIM JOBS %d @@@@@@@@@@\n", len(jobList))
	trackProgress.RunOuterTracking(len(jobList))
	defer trackProgress.Stop()

	channel_op.IterateEach_Blocking_Void(evaluateThreadCount, jobList, func(sim *simulateJob) {
		result := simulate.WowSim_Execute(runSize, sim.spec, &sim.equip, sim.model, nil, trackProgress.MakeNested())
		sim.result = &result
	})
}

func (job *MultiSetJob) linkSimResults(proposalList []MultiProposedOutput, jobList []simulateJob) []simulateResult {
	resultList := make([]simulateResult, 0, len(proposalList))
	for _, proposal := range proposalList {
		result := linkSimResult(proposal, jobList)
		resultList = append(resultList, result)
	}
	return resultList
}

func linkSimResult(proposal MultiProposedOutput, jobList []simulateJob) simulateResult {
	result := simulateResult{proposal, make([]simulate.SimResultStats, len(proposal.Outputs))}
	for outIndex := range proposal.Outputs {
		output := &proposal.Outputs[outIndex]
		for jobIndex := range jobList {
			job := &jobList[jobIndex]
			if output.FullSet.Items().Equals(&job.equip) && output.Input.Model.Spec == job.spec {
				result.result[outIndex] = *job.result
				break
			}
		}
	}
	return result
}

func (job *MultiSetJob) reportSimResults(resultList []simulateResult) {
	job.printer.Println("@@@@@@@@@@@@@@@@ RESULTS @@@@@@@@@@@@@@@@")
	for _, result := range resultList {
		job.printer.Printf("&&&&&&&&&&&&& %s\n", result.proposed.Id)
		for specIndex, specResult := range result.result {
			param := job.params[specIndex]
			job.printer.Printf("--- %s\n", param.Label)

			output := result.proposed.Outputs[specIndex]
			output.Report(job.printer)
			job.printer.Println0()
			specResult.Print(job.printer)
			// TODO common stuff
		}
	}
}

func (job *MultiSetJob) reportAsCsv(simResultList []simulateResult) {
	job.printer.Println("@@@@@@@@@@@@@@@@ SPREADSHEET COPY @@@@@@@@@@@@@@@@")

	const linesPerSpec = 7
	lines := make([]strings.Builder, 1+len(job.params)*linesPerSpec)

	for _, simResult := range simResultList {
		lineIndex := 0
		lines[lineIndex].WriteString(simResult.proposed.Id + ",")
		lineIndex++

		for _, resultStat := range simResult.result {
			for _, resultType := range simulate.SimResultTypeList {
				value := resultStat.Get(resultType)
				valueStr := strconv.FormatFloat(value, 'f', -1, 64)
				lines[lineIndex].WriteString(valueStr + ",")
				lineIndex++
			}
		}
	}

	for _, line := range lines {
		job.printer.Println(line.String())
	}
}
