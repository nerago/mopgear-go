package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/solver/solve_highs"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"time"

	"github.com/google/uuid"
)

func (job *MultiSetJob) proposalSingleBest() *util_async.FutureCancellable[multi_types.MultiProposedOutput] {
	highProcess := job.highProcessSetup()
	futureResult := highProcess.RunInterruptable(job.printer)

	return util_async.FutureCancellable_MapValue(futureResult, func(result solve_highs.HighsMultiResult) (multi_types.MultiProposedOutput, bool) {
		return job.makeOutputFromHighs(result, job.printer, uuid.NewString()), true
	})
}

func (job *MultiSetJob) proposalsAllCommonAlternates(cancelGenerate util_async.CancelSignal, extendedAlternates bool, includeInterimResults bool) (<-chan multi_types.MultiProposedOutput, *util_async.Future[int]) {
	highProcess := job.highProcessSetup()

	multiSolveChannel, expectedCountFuture := highProcess.RunForSeveral_CommonDifferent(job.printer, util_collection.Optional_Empty[int](), cancelGenerate, extendedAlternates, includeInterimResults)

	proposalChannel := util_async.Map_ChannelToChannel(4, multiSolveChannel, func(setResult solve_highs.HighsMultiResult) multi_types.MultiProposedOutput {
		return job.makeOutputFromHighs(setResult, job.printer, uuid.NewString())
	})

	existingProposal := job.existingGearAsProposal()
	proposalChannel = util_async.ChannelWithPrependedValues(proposalChannel, existingProposal)
	expectedCountFuture = expectedCountFuture.MapSameType(func(count int) (int, bool) { return count + 1, true })

	return proposalChannel, expectedCountFuture
}

func (job *MultiSetJob) proposalsUnderPermutation(solutionsPerPermute int, includeInterimResults bool, cancel util_async.CancelSignal) (<-chan multi_types.MultiProposedOutput, *util_async.Future[int]) {
	estimate := job.estimateFixedPermutations()
	job.printer.Printf("PERMUTE SET COUNT %d\n", estimate)

	permuteChannel := job.buildPermutations()

	proposalChannel := util_async.MapMulti_ChannelToChannel_Cancellable(highsThreadCount, permuteChannel, cancel,
		func(permuteSet permuteSet, resultChannel chan<- multi_types.MultiProposedOutput) {
			// TODO don't ignore updated count
			expectCount := util_async.Future_Make[int]()
			job.
				runPermute(permuteSet, solutionsPerPermute, expectCount, resultChannel, cancel, includeInterimResults)
		},
	)

	expectedTotalCount := util_async.Future_Make[int]()
	expectedTotalCount.SetResult(estimate * solutionsPerPermute)

	existingProposal := job.existingGearAsProposal()
	proposalChannel = util_async.ChannelWithPrependedValues(proposalChannel, existingProposal)

	return proposalChannel, expectedTotalCount
}

func (job *MultiSetJob) runPermute(permuteSet permuteSet, solutionsPerPermute int, expectedCount *util_async.Future[int], resultChannel chan<- multi_types.MultiProposedOutput, cancel util_async.CancelSignal, includeInterimResults bool) {
	printer := util.PrintRecorder_HoldAll()

	highProcess := job.highProcessSetupForPermute(permuteSet, printer)

	if solutionsPerPermute == 1 && !includeInterimResults {
		future := highProcess.RunInterruptable(printer)
		util_async.ChainCancel(cancel, future)
		result, hasResult := future.WaitForResult()
		if hasResult {
			resultChannel <- job.makeOutputFromHighs(result, printer, uuid.NewString())
		}
	} else {
		nextChan, expectedSubCount := highProcess.RunForSeveral_CommonDifferent(printer, util_collection.Optional_OfValue(solutionsPerPermute), cancel, false, includeInterimResults)
		expectedSubCount.ForwardResultToOtherFuture(expectedCount)
		for result := range nextChan {
			resultChannel <- job.makeOutputFromHighs(result, printer, uuid.NewString())
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

	if job.alternateGemmingAsPermute {
		panic("alternate gemming always applied, may lead to confusing results")
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

func (job *MultiSetJob) highProcessSetup() *solve_highs.SolverHighsMultiProcess {
	itemOptionsEach := util_collection.MapSliceAsNew(job.params, func(param *multiSetParamInternal) items.FullOptionsMap {
		return param.itemOptions.Clone()
	})

	highProcess := new(solve_highs.SolverHighsMultiProcess)
	job.highProcessSetup_addOptions(highProcess, itemOptionsEach)
	return highProcess
}

func (job *MultiSetJob) highProcessSetupRestrictedOnBaseline(baselineParam *multiSetParamInternal) *solve_highs.SolverHighsMultiProcess {
	itemOptionsEach := util_collection.MapSliceAsNew(job.params, func(checkParam *multiSetParamInternal) items.FullOptionsMap {
		if checkParam.paramIndex == baselineParam.paramIndex {
			return setup.OptionsSetup_FromItemSet(&baselineParam.baselineResult.FullSet)
		} else {
			itemOptions := checkParam.itemOptions.Clone()
			job.restrictOptionsToVersionsInSet(&itemOptions, &baselineParam.baselineResult.FullSet)
			return itemOptions
		}
	})

	highProcess := new(solve_highs.SolverHighsMultiProcess)
	job.highProcessSetup_addOptions(highProcess, itemOptionsEach)
	return highProcess
}

func (job *MultiSetJob) restrictOptionsToVersionsInSet(itemOptions *items.FullOptionsMap, baselineSet *items.FullItemSet) {
	itemOptions.FilterAllItems(func(check *items.FullItem) bool {
		// NOTE assumes unique equipped
		baseItem := baselineSet.Items().FindItemId(check.ItemId())
		if baseItem != nil {
			return baseItem.Equals(check)
		} else {
			return true
		}
	})
}

func (job *MultiSetJob) highProcessSetup_addOptions(highProcess *solve_highs.SolverHighsMultiProcess, itemOptionsEach []items.FullOptionsMap) {
	optionsInputList := make([]commonOptionsInput, len(job.params))
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		optionsInputList[paramIndex] = commonOptionsInput{param.Label, &itemOptionsEach[paramIndex]}
	}
	commonOptions := job.determineCommon(optionsInputList)

	highProcess.SetCommon(commonOptions)

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		highProcess.AddSetParam(solve_highs.SolverHighsMultiParam{
			Label:          param.Label,
			ItemOptions:    itemOptionsEach[paramIndex],
			SolverModel:    *solve_highs.SolverModelBuild(&param.Model, job.weightType),
			RatingMultiply: param.ratingMultiply,
		})
	}
}

func (job *MultiSetJob) makeOutputFromHighs(multiResult solve_highs.HighsMultiResult, printer *util.PrintRecorder, proposalId string) multi_types.MultiProposedOutput {
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

	if multiResult.InterimResult {
		proposalId += "-interim"
	}

	if checkNoConflicts(outputs, job.printer) {
		combo := multi_types.CommonCombo_FromProposed(outputs)
		proposed := multi_types.MultiProposedOutput{Id: proposalId, TotalRatingSum: totalRatingSum, Parts: outputs, Combo: combo, PermuteLabel: multiResult.PermuteLabel}
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

func (job *MultiSetJob) existingGearAsProposal() multi_types.MultiProposedOutput {
	proposal := multi_types.MultiProposedOutput{Id: "Existing-Gear"}
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		single := multi_types.SingleProposed_FromEquip(param.itemPrep.exactEquippedGear, &param.MultiSetParam)
		proposal.Parts = append(proposal.Parts, single)
		proposal.TotalRatingSum += single.ResultRating
	}
	proposal.Combo = multi_types.CommonCombo_FromProposed(proposal.Parts)
	return proposal
}

func (job *MultiSetJob) additionalProposalsFromSpecOptimum(cancel util_async.CancelSignal) <-chan multi_types.MultiProposedOutput {
	return util_async.MapOptional_SliceToChannel_Cancellable(2, job.params, cancel, func(param *multiSetParamInternal) (multi_types.MultiProposedOutput, bool) {
		printer := util.PrintRecorder_HoldAll()

		highProcess := job.highProcessSetupRestrictedOnBaseline(param)

		future := highProcess.RunInterruptable(printer)
		util_async.ChainCancel(cancel, future)
		result, hasResult := future.WaitForResult()

		var output multi_types.MultiProposedOutput
		if hasResult {
			proposalId := "With-Best-" + param.Label + "-" + time.Now().Format("2006-01-02-15-04-05")
			output = job.makeOutputFromHighs(result, printer, proposalId)
		}

		job.printer.AppendOther(printer)
		return output, hasResult
	})
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
