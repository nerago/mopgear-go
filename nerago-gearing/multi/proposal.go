package multi

import (
	"fmt"
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/solver/solve_highs"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/weightfind/weight_types"
	"time"

	"github.com/google/uuid"
)

func (job *MultiSetJob) proposalSingleBest(weightType weight_types.WeightType) *util_async.FutureCancellable[multi_types.MultiProposedOutput] {
	highProcess := job.highProcessSetup_SingleOrInitial(weightType)
	futureResult := highProcess.RunInterruptable(job.input.TimeLimitEachSolve, job.printer)

	return util_async.FutureCancellable_MapValue(futureResult, func(result solve_highs.HighsMultiResult) (multi_types.MultiProposedOutput, bool) {
		return job.makeProposalFromHighs(result, job.printer, uuid.NewString(), weightType), true
	})
}

func (job *MultiSetJob) proposalsAllCommonAlternates(cancelGenerate util_async.CancelSignal, extendedAlternates bool, includeInterimResults bool) (*util_async.FutureChannelMixer[multi_types.MultiProposedOutput], *util_async.FutureValueAdder[int]) {
	proposalMix := &util_async.FutureChannelMixer[multi_types.MultiProposedOutput]{}
	futureAdder := util_async.FutureValueAdderMake(0, func(a int, b int) int { return a + b })

	for _, weightType := range job.input.WeightTypeList {
		highProcess := job.highProcessSetup_SingleOrInitial(weightType)

		multiSolveChannel, expectedCountFuture := highProcess.RunForSeveral_CommonDifferent(job.input.TimeLimitEachSolve, job.printer, util_collection.Optional_Empty[int](), cancelGenerate, extendedAlternates, includeInterimResults)

		nextChannel := util_async.Map_ChannelToChannel(1, multiSolveChannel, func(setResult solve_highs.HighsMultiResult) multi_types.MultiProposedOutput {
			return job.makeProposalFromHighs(setResult, job.printer, uuid.NewString(), weightType)
		})

		proposalMix.AddChannel(nextChannel)
		futureAdder.AddFuture(expectedCountFuture)
	}

	for _, weightType := range job.input.WeightTypeList {
		existingProposal := job.existingGearAsProposal(weightType)
		proposalMix.AddValue(existingProposal)
		futureAdder.AddValueImmediate(func(x int) int { return x + 1 })
	}

	return proposalMix, futureAdder
}

func (job *MultiSetJob) proposalsUnderPermutation(solutionsPerPermute int, includeInterimResults bool, cancel util_async.CancelSignal) (<-chan multi_types.MultiProposedOutput, *util_async.FutureValueAdderInt) {
	estimate := job.estimateFixedPermutations()
	job.printer.Printf("PERMUTE SET COUNT %d\n", estimate)
	futureAdder := util_async.FutureValueAdderIntMake(0)

	permuteChannel := job.buildPermutations()

	proposalChannel := util_async.MapMulti_ChannelToChannel_Cancellable(c_permuteThreadCount, permuteChannel, cancel,
		func(permuteSet permuteSet, resultChannel chan<- multi_types.MultiProposedOutput) {
			for _, weightType := range job.input.WeightTypeList {
				job.runPermute(permuteSet, weightType, solutionsPerPermute, futureAdder, resultChannel, cancel, includeInterimResults)
			}
		},
	)

	for _, weightType := range job.input.WeightTypeList {
		existingProposal := job.existingGearAsProposal(weightType)
		proposalChannel = util_async.ChannelWithPrependedValues(proposalChannel, existingProposal)
		futureAdder.AddValueImmediate(1)
	}

	return proposalChannel, futureAdder
}

func (job *MultiSetJob) runPermute(permuteSet permuteSet, weightType weight_types.WeightType, solutionsPerPermute int, expectedCount *util_async.FutureValueAdderInt, resultChannel chan<- multi_types.MultiProposedOutput, cancel util_async.CancelSignal, includeInterimResults bool) {
	printer := util.PrintRecorder_HoldAll()

	highProcessList := job.highProcessSetupForPermute(permuteSet, printer)

	if solutionsPerPermute == 1 && !includeInterimResults {
		for _, highProcess := range highProcessList {
			future := highProcess.RunInterruptable(job.input.TimeLimitEachSolve, printer)
			util_async.ChainCancel(cancel, future)
			result, hasResult := future.WaitForResult()
			if hasResult {
				resultChannel <- job.makeProposalFromHighs(result, printer, uuid.NewString(), weightType)
			}
		}
		expectedCount.AddValueImmediate(len(highProcessList))
	} else {
		for _, highProcess := range highProcessList {
			nextChan, expectedSubCount := highProcess.RunForSeveral_CommonDifferent(
				job.input.TimeLimitEachSolve, printer,
				util_collection.Optional_OfValue(solutionsPerPermute), cancel,
				false, includeInterimResults)
			expectedCount.AddFuture(expectedSubCount)
			for result := range nextChan {
				resultChannel <- job.makeProposalFromHighs(result, printer, uuid.NewString(), weightType)
			}
		}
	}

	job.printer.AppendOther(printer)
}

func (job *MultiSetJob) checkNoPermutations() {
	if len(job.input.ItemInput.DistinctUsageGroups) > 0 {
		panic("usage groups will be ignored, may lead to confusing results")
	}

	if len(job.input.ItemInput.AlternateUpgradeChoices) > 0 {
		panic("alternate upgrades will be ignored, may lead to confusing results")
	}

	if job.input.ItemInput.AlternateGemsEnableAsPermute {
		panic("alternate gems always applied, may lead to confusing results")
	}

	// EnablePermuteOnItemCountOptions don't care, will be checked as normal

	for paramIndex := range job.input.Param {
		param := &job.input.Param[paramIndex]
		for _, itemArray := range param.ItemInputs.SemiFixedSlots {
			if len(itemArray) > 1 {
				panic("TryAll slots will be ignored, may lead to confusing results")
			}
		}
	}
}

func (job *MultiSetJob) highProcessSetup_SingleOrInitial(weightType weight_types.WeightType) *solve_highs.SolverHighsMultiProcess {
	highProcess := new(solve_highs.SolverHighsMultiProcess)

	itemOptionsEach := make(map[string]*items.FullOptionsMap)
	for _, work := range job.working.SeqKey1ValueWithKey2(weightType) {
		itemOptionsEach[work.label()] = work.itemOptions()
	}

	commonOptions := job.determineCommon(itemOptionsEach, job.input.ItemInput.ReforgingAllowNonCommon)
	highProcess.SetCommon(commonOptions)

	for label, work := range job.working.SeqKey1ValueWithKey2(weightType) {
		highProcess.AddSetParam(solve_highs.SolverHighsMultiParam{
			Label:          label,
			ItemOptions:    *itemOptionsEach[label],
			SolverModel:    *solve_highs_types.SolverModelBuild(work.model(), work.weightType, nil),
			RatingMultiply: work.ratingMultiply,
		})
	}

	return highProcess
}

func (job *MultiSetJob) highProcessSetupRestrictedOnBaseline(baselineWork *specWorking) *solve_highs.SolverHighsMultiProcess {
	highProcess := new(solve_highs.SolverHighsMultiProcess)

	itemOptionsEach := make(map[string]*items.FullOptionsMap)
	for _, work := range job.working.SeqKey1ValueWithKey2(baselineWork.weightType) {
		label := work.label()
		if label == baselineWork.label() {
			itemOptionsEach[label] = new(setup.OptionsSetup_FromItemSet(&baselineWork.baselineResult.FullSet))
		} else {
			itemOptions := new(work.itemOptions().Clone())
			job.restrictOptionsToVersionsInSet(itemOptions, &baselineWork.baselineResult.FullSet)
			itemOptionsEach[label] = itemOptions
		}
	}

	commonOptions := job.determineCommon(itemOptionsEach, job.input.ItemInput.ReforgingAllowNonCommon)
	highProcess.SetCommon(commonOptions)

	for label, work := range job.working.SeqKey1ValueWithKey2(baselineWork.weightType) {
		highProcess.AddSetParam(solve_highs.SolverHighsMultiParam{
			Label:          label,
			ItemOptions:    *itemOptionsEach[label],
			SolverModel:    *solve_highs_types.SolverModelBuild(work.model(), work.weightType, nil),
			RatingMultiply: work.ratingMultiply,
		})
	}
	return highProcess
}

func (job *MultiSetJob) restrictOptionsToVersionsInSet(itemOptions *items.FullOptionsMap, baselineSet *items.FullItemSet) {
	for slot := range itemOptions {
		if len(itemOptions[slot]) > 0 {
			itemOptions[slot] = util_collection.FilterSliceAsNew(itemOptions[slot], func(check *items.FullItem) bool {
				baseItem := baselineSet.Items().FindItemId(check.ItemId())
				if baseItem != nil {
					return baseItem.Equals(check)
				} else {
					return true
				}
			})
			if len(itemOptions[slot]) == 0 {
				itemOptions[slot] = []items.FullItem{*baselineSet.Items()[slot]}
			}
		}
	}
}

func (job *MultiSetJob) makeProposalFromHighs(multiResult solve_highs.HighsMultiResult, printer *util.PrintRecorder, proposalId string, weightType weight_types.WeightType) multi_types.MultiProposedOutput {
	totalRatingSum := 0.0
	outputs := make(map[string]multi_types.SingleProposedOutput, len(multiResult.Entries))

	for label, work := range job.working.SeqKey1ValueWithKey2(weightType) {
		resultEntry := multiResult.Entries[label]

		rating := work.model().CalcRatingFull(&resultEntry.ItemSet, weightType)
		totalRatingSum += rating * work.ratingMultiply

		single := multi_types.SingleProposed_FromItemSet(resultEntry.ItemSet, resultEntry.OutputId, work.model().Spec, label, rating)
		outputs[label] = single

		work.addSeen(resultEntry.ItemSet.Items())

		printer.Printf("LABEL %s\n", label)
		single.Report(work.model(), printer)
	}

	if multiResult.InterimResult {
		proposalId += "-interim"
	}

	if checkNoConflicts(outputs, job.printer) {
		combo := multi_types.CommonCombo_FromProposed(outputs)
		proposed := multi_types.MultiProposedOutput{
			Id:             proposalId,
			TotalRatingSum: totalRatingSum,
			Parts:          outputs,
			Combo:          combo,
			PermuteLabel:   multiResult.PermuteLabel,
			WeightType:     weightType,
		}
		return proposed
	} else {
		panic("conflicted items")
	}
}

func (job *MultiSetJob) listInitialOutputs(bestOutputs <-chan multi_types.MultiProposedOutput) <-chan multi_types.MultiProposedOutput {
	return util_async.PeekChannel(bestOutputs, func(prop *multi_types.MultiProposedOutput) {
		job.printer.Printf("::::::::: PROPOSED %f :::::::: %s ::::::::\n", prop.TotalRatingSum, prop.Id)
		job.printer.Printf("Weight Type %d\n", prop.WeightType)
		for label, out := range prop.Parts {
			prep := job.itemPrep[label]
			job.printer.Println(label)
			out.Report(&prep.model, job.printer)
		}
	})
}

func (job *MultiSetJob) existingGearAsProposal(weightType weight_types.WeightType) multi_types.MultiProposedOutput {
	proposal := multi_types.MultiProposedOutput{
		Id:           fmt.Sprintf("Existing-Gear-%d", weightType),
		Parts:        make(map[string]multi_types.SingleProposedOutput),
		PermuteLabel: "",
		WeightType:   weightType,
	}
	for label, work := range job.working.SeqKey1ValueWithKey2(weightType) {
		prep := job.itemPrep[label]
		set := items.FullItemSet_FromMap(prep.exactEquippedGear)

		rating := work.model().CalcRatingFull(&set, weightType)
		proposal.TotalRatingSum += rating

		single := multi_types.SingleProposed_FromItemSet(set, uuid.NewString(), work.model().Spec, label, rating)
		proposal.Parts[label] = single
	}
	proposal.Combo = multi_types.CommonCombo_FromProposed(proposal.Parts)
	return proposal
}

func (job *MultiSetJob) additionalProposalsFromSpecOptimalBaseline(cancel util_async.CancelSignal) (<-chan multi_types.MultiProposedOutput, int) {
	allWorking := util_async.SeqToChannel(job.working.SeqValues())
	outputChannel := util_async.MapMulti_ChannelToChannel_Cancellable(c_additionalProposal_threadCount, allWorking, cancel, func(work *specWorking, downstream chan<- multi_types.MultiProposedOutput) {
		printer := util.PrintRecorder_HoldAll()

		highProcess := job.highProcessSetupRestrictedOnBaseline(work)

		future := highProcess.RunInterruptable(job.input.TimeLimitEachSolve, printer)
		util_async.ChainCancel(cancel, future)
		result, hasResult := future.WaitForResult()

		if hasResult {
			proposalId := fmt.Sprintf("With-Best-%d-%s-%s", work.weightType, work.label(), time.Now().Format("2006-01-02-15-04-05"))
			downstream <- job.makeProposalFromHighs(result, printer, proposalId, work.weightType)
		}

		job.printer.AppendOther(printer)
	})
	estimate := job.working.Size()
	return outputChannel, estimate
}

func fakeIdNumber(index int) string {
	build := util.StringBuild2{}
	if index >= 0 && index <= 9 {
		for range 8 {
			build.WriteInt(index)
		}
		build.WriteRune('-')
		for range 4 {
			build.WriteInt(index)
		}
		build.WriteRune('-')
		for range 4 {
			build.WriteInt(index)
		}
		build.WriteRune('-')
		for range 4 {
			build.WriteInt(index)
		}
		build.WriteRune('-')
		for range 12 {
			build.WriteInt(index)
		}
	} else {
		build.WriteString("PROPOSAL-ID-")
		build.WriteInt(index)
	}
	return build.String()
}
