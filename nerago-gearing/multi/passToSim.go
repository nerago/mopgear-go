package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"

	"github.com/google/uuid"
)

func (job *MultiSetJob) FindTopAndPassToSim(targetCount uint64, topCapture uint64, addRevisions bool, runSize simulate.WowSim_RunSize) {
	topTracker := util.TrackProgress_Start()
	topTracker.RunOuterTracking(3)
	defer topTracker.Stop()

	job.printer.Printf("@@@@@@@@@@ FIND TOP %d %d @@@@@@@@@@\n", targetCount, topCapture)
	bestOutputs := job.runForTopN(targetCount, topCapture, topTracker.MakeNested())
	job.listInitialOutputs(bestOutputs)

	var proposalList []multiProposedOutput
	if addRevisions {
		proposalList = job.prepareRevisionsForSim(bestOutputs, topTracker.MakeNested())
	} else {
		proposalList = bestOutputs
	}

	simList := job.prepareSimList(proposalList)
	job.runSims(simList, runSize, topTracker.MakeNested())

	simResult := job.linkSimResults(proposalList, simList)
	job.reportSimResults(simResult)
	job.reportAsCsv(simResult)
}

func (job *MultiSetJob) prepareRevisionsForSim(proposedList []multiProposedOutput, trackProgress *util.TrackProgress) []multiProposedOutput {
	job.printer.Printf("@@@@@@@@@@ MAKE REVISIONS FOR %d @@@@@@@@@@\n", len(proposedList))

	expectedSets := len(proposedList) * len(job.params) * revisedExtraSetsExpectedEach
	trackProgress.RunOuterTracking(expectedSets)
	defer trackProgress.Stop()

	allProposals := channel_op.Map_SliceToSlice(generateThreadCount, proposedList, func(prior *multiProposedOutput, downstream chan<- multiProposedOutput) {
		job.prepareOneRevisionForSim(prior, trackProgress, downstream)
	})

	allProposals = append(allProposals, job.existingGearAsProposal())
	return allProposals
}

func (job *MultiSetJob) prepareOneRevisionForSim(prior *multiProposedOutput, trackProgress *util.TrackProgress, downstream chan<- multiProposedOutput) {
	printer := util.PrintRecorder_HoldAll()
	printer.Printf(">>> PREP REVISIONS %s\n", prior.id)

	revisedCommon := job.revisedComboActuallyUsed(prior.parts, &prior.combo, printer)

	revisedOptionArrays := make([][]singleProposed, len(prior.parts))

	for i := range prior.parts {
		draft := &prior.parts[i]
		param := &job.params[i]

		// TODO refactor common parts with cullingMakeRevisions better
		printer.Printf("== %s\n", param.Label)
		printer.Println("DRAFT")
		draft.Report(printer)

		specOptions := job.makeRevised(param, &revisedCommon, trackProgress, printer)
		for _, newOutput := range specOptions {
			param.seenInSolutions.Add(&newOutput.fullSet)
		}

		param.seenInSolutions.Add(&draft.fullSet)
		specOptions = append(specOptions, *draft)

		specOptions = util.RemoveDuplicatesFuncNotify(specOptions, func(a, b *singleProposed) bool {
			return a.fullSet.Equals(&b.fullSet)
		}, func(removed *singleProposed) {
			printer.Printf("removed duplicate output %s\n", removed.outputId)
		})

		revisedOptionArrays[i] = specOptions
	}

	for outputSet := range util.PermuteAll(revisedOptionArrays) {
		var totalRatingSum float64
		for _, output := range outputSet {
			totalRatingSum += output.resultRating
		}
		if checkNoConflicts(outputSet) {
			proposed := multiProposedOutput{uuid.NewString(), totalRatingSum, outputSet, revisedCommon}
			componentIds := ""
			for _, set := range outputSet {
				componentIds = componentIds + set.outputId + " "
			}
			printer.Printf("&&& NEW PROPOSAL %s => %s\n", proposed.id, componentIds)
			downstream <- proposed
		}
	}

	printer.Println0()
	printer.Println0()

	job.printer.AppendOther(printer)
}

func checkNoConflicts(outputSet []singleProposed) bool {
	itemById := make(map[items.ItemId]*items.FullItem)
	for outputIndex := range outputSet {
		for item := range outputSet[outputIndex].fullSet.Items().AllItemSeq() {
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

func (job *MultiSetJob) existingGearAsProposal() multiProposedOutput {
	proposal := multiProposedOutput{id: uuid.NewString()}
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		single := SingleProposed_FromEquip(param.exactEquippedGear, param)
		proposal.parts = append(proposal.parts, single)
		proposal.totalRatingSum += single.resultRating
	}
	proposal.combo = job.determineComboFromScratch(proposal.parts, comboType_equippedExact)
	return proposal
}

type simulateJob struct {
	spec        stats.SpecType
	fight       stats.WowSim_Fight
	equip       items.FullEquipMap
	professions model.ProfessionInfo
	result      *simulate.SimResultStats
}

func (simJob *simulateJob) Equals(other *simulateJob) bool {
	return simJob.spec == other.spec && simJob.fight == other.fight && simJob.equip.Equals(&other.equip) && simJob.professions == other.professions
}

type simulateResult struct {
	proposed multiProposedOutput
	result   []simulate.SimResultStats
}

func (job *MultiSetJob) prepareSimList(proposalList []multiProposedOutput) []simulateJob {
	jobList := make([]simulateJob, 0)
	for _, proposal := range proposalList {
		for _, output := range proposal.parts {
			job := simulateJob{output.spec, output.model.SimulateAs, *output.fullSet.Items(), output.model.Professions, nil}
			jobList = append(jobList, job)
		}
	}

	jobList = util.RemoveDuplicatesFunc(jobList, (*simulateJob).Equals)

	return jobList
}

func (job *MultiSetJob) runSims(jobList []simulateJob, runSize simulate.WowSim_RunSize, trackProgress *util.TrackProgress) {
	job.printer.Printf("@@@@@@@@@@ RUN SIM JOBS %d @@@@@@@@@@\n", len(jobList))
	trackProgress.RunOuterTracking(len(jobList))
	defer trackProgress.Stop()

	channel_op.ForEach_Blocking_Void(evaluateThreadCount, jobList, func(sim *simulateJob) {
		result := simulate.WowSim_Execute_SelectFight(runSize, sim.spec, sim.fight, &sim.equip, sim.professions, nil, trackProgress.MakeNested())
		job.printer.Printf("sim %22s fight=%d %s\n", sim.spec.Name(), sim.fight, result.CompactStringGeneral())
		sim.result = &result
	})
}

func (job *MultiSetJob) linkSimResults(proposalList []multiProposedOutput, jobList []simulateJob) []simulateResult {
	resultList := make([]simulateResult, 0, len(proposalList))
	for _, proposal := range proposalList {
		result := linkSimResult(proposal, jobList)
		resultList = append(resultList, result)
	}
	return resultList
}

func linkSimResult(proposal multiProposedOutput, jobList []simulateJob) simulateResult {
	result := simulateResult{proposal, make([]simulate.SimResultStats, len(proposal.parts))}
	for outIndex := range proposal.parts {
		output := &proposal.parts[outIndex]
		for jobIndex := range jobList {
			job := &jobList[jobIndex]
			if output.fullSet.Items().Equals(&job.equip) && output.spec == job.spec && output.model.SimulateAs == job.fight {
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
		job.printer.Printf("&&&&&&&&&&&&& %s\n", result.proposed.id)
		printChosenCombo(&result.proposed.combo, job.printer)
		for specIndex, specResult := range result.result {
			param := &job.params[specIndex]
			job.printer.Printf("---------------- %s ----------------\n", param.Label)
			output := result.proposed.parts[specIndex]
			output.Report(job.printer)
			specResult.Print(job.printer)

			for slot, itemId := range param.reportVariant {
				variantEquip := *output.fullSet.Items()
				variantItem := job.findVariantItem(result, itemId, param)
				variantEquip[slot] = variantItem
				job.printer.Printf("---------------- %s %s ----------------\n", param.Label, variantItem.BaseName)
				simulate.WowSimJson_Write(&variantEquip, &param.Model, job.printer)
			}
		}
		job.printer.Println0()
		job.printer.Println0()
	}
}

func (job *MultiSetJob) findVariantItem(result simulateResult, itemId items.ItemId, param *multiSetParamInternal) *items.FullItem {
	variantItem := result.proposed.findItemById(itemId)
	if variantItem != nil {
		return variantItem
	}

	if item, found := param.itemOptions.FindItemIdFirstOptional(itemId); found {
		return item
	}

	for paramIndex := range job.params {
		otherParam := &job.params[paramIndex]
		if item, found := otherParam.itemOptions.FindItemIdFirstOptional(itemId); found {
			return item
		}
	}

	_, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, 2, &param.Model, job.printer)
	return example
}

func (job *MultiSetJob) reportAsCsv(simResultList []simulateResult) {
	job.printer.Println("@@@@@@@@@@@@@@@@ SPREADSHEET COPY @@@@@@@@@@@@@@@@")

	outputTypes := []simulate.SimResultType{simulate.Result_DPS, simulate.Result_DTPS, simulate.Result_TMI, simulate.Result_DEATH}

	csv := util.CSVOutputByColumn{}
	csv.InitRows(len(job.params)*len(outputTypes) + 1)
	csv.AddString("id")
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		for _, resultType := range outputTypes {
			csv.AddToBuilder(func(b *util.StringBuild2) {
				b.WriteString(resultType.String())
				b.WriteString(" (")
				b.WriteString(param.Label)
				b.WriteRune(')')
			})
		}
	}
	csv.FinishColumn()

	for _, simResult := range simResultList {
		csv.AddString(simResult.proposed.id)

		for _, resultStat := range simResult.result {
			for _, resultType := range outputTypes {
				value := resultStat.Get(resultType)
				csv.AddFloat64(value, -1)
			}
		}

		csv.FinishColumn()
	}

	csv.Write(job.printer)
}
