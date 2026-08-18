package multi

import (
	"fmt"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/solver/solve_highs"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"maps"
	"time"

	"github.com/google/uuid"
)

func (group *workingGroup) proposalSingleBest() *util_async.FutureCancellable[multi_types.MultiProposedOutput] {
	highProcess := group.highProcessSetup_SingleOrInitial()
	futureResult := highProcess.RunInterruptable(group.job.input.TimeLimitEachSolve, group.job.printer)

	return util_async.FutureCancellable_MapValue(futureResult, func(result solve_highs.HighsMultiResult) (multi_types.MultiProposedOutput, bool) {
		return group.makeProposalFromHighs(result, group.job.printer, uuid.NewString()), true
	})
}

func (group *workingGroup) proposalsAllCommonAlternates(cancelGenerate util_async.CancelSignal, futureCount *util_async.FutureValueAdderInt, extendedAlternates bool, includeInterimResults bool) *util_async.FutureChannelMixerStatic[multi_types.MultiProposedOutput] {
	proposalMix := &util_async.FutureChannelMixerStatic[multi_types.MultiProposedOutput]{}

	highProcess := group.highProcessSetup_SingleOrInitial()

	multiSolveChannel, expectedCountFuture := highProcess.RunForSeveral_CommonDifferent(
		group.job.input.TimeLimitEachSolve, group.job.printer, util_collection.Optional_Empty[int](), cancelGenerate,
		extendedAlternates, includeInterimResults)

	nextChannel := util_async.Map_ChannelToChannel(1, multiSolveChannel, func(setResult solve_highs.HighsMultiResult) multi_types.MultiProposedOutput {
		return group.makeProposalFromHighs(setResult, group.job.printer, uuid.NewString())
	})

	proposalMix.AddChannel(nextChannel)
	futureCount.AddFuture(expectedCountFuture)

	existingProposal := group.existingGearAsProposal()
	proposalMix.AddValue(existingProposal)
	futureCount.AddValueImmediate(1)

	return proposalMix
}

func (group *workingGroup) proposalsUnderPermutation(solutionsPerPermute int, includeInterimResults bool, futureCount *util_async.FutureValueAdderInt, cancel util_async.CancelSignal) *util_async.FutureChannelMixerStatic[multi_types.MultiProposedOutput] {
	estimate := group.job.estimateFixedPermutations()
	group.job.printer.Printf("PERMUTE SET COUNT %d\n", estimate)

	permuteChannel := group.job.buildPermutations()

	proposalChannel := util_async.MapMulti_ChannelToChannel_Cancellable(c_permuteThreadCount, permuteChannel, cancel,
		func(permuteSet permuteSet, resultChannel chan<- multi_types.MultiProposedOutput) {
			group.runPermute(permuteSet, solutionsPerPermute, futureCount, resultChannel, cancel, includeInterimResults)
		},
	)

	mixer := util_async.FutureChannelMixerStatic[multi_types.MultiProposedOutput]{}
	mixer.AddChannel(proposalChannel)

	existingProposal := group.existingGearAsProposal()
	mixer.AddValue(existingProposal)
	futureCount.AddValueImmediate(1)

	return &mixer
}

func (group *workingGroup) runPermute(permuteSet permuteSet, solutionsPerPermute int, expectedCount *util_async.FutureValueAdderInt, resultChannel chan<- multi_types.MultiProposedOutput, cancel util_async.CancelSignal, includeInterimResults bool) {
	printer := util.PrintRecorder_HoldAll()

	highProcess := group.highProcessSetupForPermute(permuteSet, printer)

	if solutionsPerPermute == 1 && !includeInterimResults {
		future := highProcess.RunInterruptable(group.job.input.TimeLimitEachSolve, printer)
		util_async.ChainCancel(cancel, future)
		result, hasResult := future.WaitForResult()
		if hasResult {
			resultChannel <- group.makeProposalFromHighs(result, printer, uuid.NewString())
		}

		expectedCount.AddValueImmediate(1)
	} else {
		nextChan, expectedSubCount := highProcess.RunForSeveral_CommonDifferent(
			group.job.input.TimeLimitEachSolve, printer,
			util_collection.Optional_OfValue(solutionsPerPermute), cancel,
			false, includeInterimResults)
		expectedCount.AddFuture(expectedSubCount)
		for result := range nextChan {
			resultChannel <- group.makeProposalFromHighs(result, printer, uuid.NewString())
		}
	}

	group.job.printer.AppendOther(printer)
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

func (group *workingGroup) highProcessSetup_SingleOrInitial() *solve_highs.SolverHighsMultiProcess {
	highProcess := new(solve_highs.SolverHighsMultiProcess)

	itemOptionsEach := make(map[string]*items.FullOptionsMap)
	for _, work := range group.workers {
		itemOptionsEach[work.Label()] = work.ItemOptions()
	}

	commonOptions := group.determineCommon(itemOptionsEach)
	highProcess.SetCommon(commonOptions)

	for label, work := range group.workers {
		highProcess.AddSetParam(solve_highs.SolverHighsMultiParam{
			Label:          label,
			ItemOptions:    *itemOptionsEach[label],
			SolverModel:    *solve_highs_types.SolverModelBuild(work.Model(), work.weightType, nil),
			RatingMultiply: work.ratingMultiply,
		})
	}

	return highProcess
}

func (group *workingGroup) highProcessSetupRestrictedOnBaseline(baselineWork *specWorker) *solve_highs.SolverHighsMultiProcess {
	highProcess := new(solve_highs.SolverHighsMultiProcess)

	itemOptionsEach := make(map[string]*items.FullOptionsMap)
	for _, work := range group.workers {
		label := work.Label()
		if label == baselineWork.Label() {
			itemOptionsEach[label] = new(setup.OptionsSetup_FromItemSet(&baselineWork.baselineResult.FullSet))
		} else {
			itemOptions := new(work.ItemOptions().Clone())
			group.restrictOptionsToVersionsInSet(itemOptions, &baselineWork.baselineResult.FullSet)
			itemOptionsEach[label] = itemOptions
		}
	}

	commonOptions := group.determineCommon(itemOptionsEach)
	highProcess.SetCommon(commonOptions)

	for label, work := range group.workers {
		highProcess.AddSetParam(solve_highs.SolverHighsMultiParam{
			Label:          label,
			ItemOptions:    *itemOptionsEach[label],
			SolverModel:    *solve_highs_types.SolverModelBuild(work.Model(), work.weightType, nil),
			RatingMultiply: work.ratingMultiply,
		})
	}
	return highProcess
}

func (group *workingGroup) restrictOptionsToVersionsInSet(itemOptions *items.FullOptionsMap, baselineSet *items.FullItemSet) {
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

func (group *workingGroup) makeProposalFromHighs(multiResult solve_highs.HighsMultiResult, printer *util.PrintRecorder, proposalId string) multi_types.MultiProposedOutput {
	totalRatingSum := 0.0
	outputs := make(map[string]multi_types.SingleProposedOutput, len(multiResult.Entries))

	for label, work := range group.workers {
		resultEntry := multiResult.Entries[label]

		rating := work.Model().CalcRatingFull(&resultEntry.ItemSet, group.weightType)
		totalRatingSum += rating * work.ratingMultiply

		single := multi_types.SingleProposed_FromItemSet(resultEntry.ItemSet, resultEntry.OutputId, work.Model().Spec, label, rating)
		outputs[label] = single

		work.AddSeen(resultEntry.ItemSet.Items())

		printer.Printf("LABEL %s\n", label)
		single.Report(work.Model(), printer)
	}

	if multiResult.InterimResult {
		proposalId += "-interim"
	}

	if checkNoConflicts(outputs, group.job.printer) {
		combo := multi_types.CommonCombo_FromProposed(outputs)
		proposed := multi_types.MultiProposedOutput{
			Id:             proposalId,
			TotalRatingSum: totalRatingSum,
			Parts:          outputs,
			Combo:          combo,
			PermuteLabel:   multiResult.PermuteLabel,
			WeightType:     group.weightType,
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

func (group *workingGroup) existingGearAsProposal() multi_types.MultiProposedOutput {
	proposal := multi_types.MultiProposedOutput{
		Id:           fmt.Sprintf("Existing-Gear-%d", group.weightType),
		Parts:        make(map[string]multi_types.SingleProposedOutput),
		PermuteLabel: "",
		WeightType:   group.weightType,
	}
	for label, work := range group.workers {
		prep := group.job.itemPrep[label]
		set := items.FullItemSet_FromMap(prep.exactEquippedGear)

		rating := work.Model().CalcRatingFull(&set, group.weightType)
		proposal.TotalRatingSum += rating

		single := multi_types.SingleProposed_FromItemSet(set, uuid.NewString(), work.Model().Spec, label, rating)
		proposal.Parts[label] = single
	}
	proposal.Combo = multi_types.CommonCombo_FromProposed(proposal.Parts)
	return proposal
}

func (group *workingGroup) additionalProposalsFromSpecOptimalBaseline(cancel util_async.CancelSignal) <-chan multi_types.MultiProposedOutput {
	allWorking := util_async.SeqToChannel(maps.Values(group.workers))
	outputChannel := util_async.MapMulti_ChannelToChannel_Cancellable(c_additionalProposal_threadCount, allWorking, cancel, func(work *specWorker, downstream chan<- multi_types.MultiProposedOutput) {
		printer := util.PrintRecorder_HoldAll()

		highProcess := group.highProcessSetupRestrictedOnBaseline(work)

		future := highProcess.RunInterruptable(group.job.input.TimeLimitEachSolve, printer)
		util_async.ChainCancel(cancel, future)
		result, hasResult := future.WaitForResult()

		if hasResult {
			proposalId := fmt.Sprintf("With-Best-%d-%s-%s", work.weightType, work.Label(), time.Now().Format("2006-01-02-15-04-05"))
			downstream <- group.makeProposalFromHighs(result, printer, proposalId)
		}

		group.job.printer.AppendOther(printer)
	})
	return outputChannel
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
