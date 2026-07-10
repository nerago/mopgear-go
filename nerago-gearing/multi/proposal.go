package multi

import (
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/solver/solve_highs"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"

	"github.com/google/uuid"
)

func (job *MultiSetJob) proposalsAllCommonAlternates(cancelGenerate util_async.CancelSignal, extendedAlternates bool) (<-chan multi_types.MultiProposedOutput, *util_async.Future[int]) {
	highProcess := job.highProcessSetup()

	multiSolveChannel, expectedCountFuture := highProcess.RunForSeveral_CommonDifferent(job.printer, util.Optional_Empty[int](), cancelGenerate, extendedAlternates)

	proposalChannel := util_async.Map_ChannelToChannel(4, multiSolveChannel, func(setResult solve_highs.HighsMultiResult) multi_types.MultiProposedOutput {
		return job.makeOutputFromHighs(setResult, job.printer)
	})

	existingProposal := job.existingGearAsProposal()
	proposalChannel = util_async.ChannelWithPrependedValues(proposalChannel, existingProposal)
	expectedCountFuture = expectedCountFuture.MapSameType(func(count int) (int, bool) { return count + 1, true })

	return proposalChannel, expectedCountFuture
}

func (job *MultiSetJob) proposalsUnderPermutation(solutionsPerPermute int, cancel util_async.CancelSignal) (<-chan multi_types.MultiProposedOutput, *util_async.Future[int]) {
	estimate := job.estimateFixedPermutations()
	job.printer.Printf("PERMUTE SET COUNT %d\n", estimate)

	permuteChannel := job.preparePermutations()

	proposalChannel := util_async.MapMulti_ChannelToChannel_Cancellable(highsThreadCount, permuteChannel, cancel,
		func(permuteSet permuteSet, resultChannel chan<- multi_types.MultiProposedOutput) {
			// TODO don't ignore updated count
			expectCount := util_async.Future_Make[int]()
			job.runPermute(permuteSet, solutionsPerPermute, expectCount, resultChannel, cancel)
		},
	)

	expectedTotalCount := util_async.Future_Make[int]()
	expectedTotalCount.SetResult(estimate * solutionsPerPermute)

	existingProposal := job.existingGearAsProposal()
	proposalChannel = util_async.ChannelWithPrependedValues(proposalChannel, existingProposal)

	return proposalChannel, expectedTotalCount
}

func (job *MultiSetJob) runPermute(permuteSet permuteSet, solutionsPerPermute int, expectedCount *util_async.Future[int], resultChannel chan<- multi_types.MultiProposedOutput, cancel util_async.CancelSignal) {
	printer := util.PrintRecorder_HoldAll()

	highProcess := job.highProcessSetupForPermute(permuteSet, printer)

	if solutionsPerPermute == 1 {
		future := highProcess.RunInterruptable(printer)
		util_async.ChainCancel(cancel, future)
		result, hasResult := future.WaitForResult()
		if hasResult {
			resultChannel <- job.makeOutputFromHighs(result, printer)
		}
	} else {
		nextChan, expectedSubCount := highProcess.RunForSeveral_CommonDifferent(printer, util.Optional_OfValue(solutionsPerPermute), cancel, false)
		expectedSubCount.ForwardResultToOtherFuture(expectedCount)
		for result := range nextChan {
			resultChannel <- job.makeOutputFromHighs(result, printer)
		}
	}

	job.printer.AppendOther(printer)
}

func (job *MultiSetJob) paramFromLabel(paramLabel string) *multiSetParamInternal {
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		if param.Label == paramLabel {
			return param
		}
	}
	panic("param not found")
}

func (job *MultiSetJob) checkNoPermutations() {
	if len(job.distinctUsageGroups) > 0 {
		panic("usage groups will be ignored, may lead to confusing results")
	}

	if len(job.alternateUpgradeChoices) > 0 {
		panic("alternate upgrades will be ignored, may lead to confusing results")
	}

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		for _, itemArray := range param.SemiFixedSlots {
			if len(itemArray) > 1 {
				panic("TryAll slots will be ignored, may lead to confusing results")
			}
		}
	}
}

func (job *MultiSetJob) highProcessSetup() solve_highs.SolverHighsMultiProcess {
	highProcess := solve_highs.SolverHighsMultiProcess{}

	optionsInputList := util.MapSliceAsNew(job.params, func(param *multiSetParamInternal) commonOptionsInput {
		return commonOptionsInput{param.Label, &param.itemOptions}
	})
	commonOptions := job.determineCommon(optionsInputList)
	highProcess.SetCommon(commonOptions)

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		highProcess.AddSetParam(solve_highs.SolverHighsMultiParam{
			Label:          param.Label,
			ItemOptions:    param.itemOptions,
			Gear_model:     &param.Model,
			RatingMultiply: param.ratingMultiply,
		})
	}
	return highProcess
}

func (job *MultiSetJob) makeOutputFromHighs(multiResult solve_highs.HighsMultiResult, printer *util.PrintRecorder) multi_types.MultiProposedOutput {
	var totalRatingSum float64
	outputs := make([]multi_types.SingleProposedOutput, len(job.params))

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		itemSet := multiResult.ItemSets[paramIndex]
		outputId := multiResult.OutputId[paramIndex]
		single := multi_types.SingleProposed_FromItemSet(itemSet, outputId, &param.Model)
		printer.Printf("LABEL %s\n", param.Label)
		single.Report(printer)
		param.seenInSolutions.Add(&itemSet)
		outputs[paramIndex] = single
		totalRatingSum += single.ResultRating * param.ratingMultiply
	}

	if checkNoConflicts(outputs, job.printer) {
		combo := multi_types.CommonCombo_FromProposed(outputs)
		proposed := multi_types.MultiProposedOutput{Id: uuid.NewString(), TotalRatingSum: totalRatingSum, Parts: outputs, Combo: combo, PermuteLabel: multiResult.PermuteLabel}
		return proposed
	} else {
		panic("conflicted items")
	}
}

func (job *MultiSetJob) listInitialOutputs(bestOutputs <-chan multi_types.MultiProposedOutput) <-chan multi_types.MultiProposedOutput {
	return util_async.PeekChannel(bestOutputs, func(best *multi_types.MultiProposedOutput) {
		job.printer.Printf("::::::::: MULTI RATING %.0f :::::::: %s ::::::::\n", best.TotalRatingSum, best.Id)
		for i, out := range best.Parts {
			job.printer.Println(job.params[i].Label)
			out.Report(job.printer)
		}
	})
}
