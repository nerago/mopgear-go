package multi

import (
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"sync/atomic"

	"github.com/google/uuid"
)

func (job *MultiSetJob) proposalsAllCommonAlternates(cancelGenerate channel_op.CancelSignal) <-chan multi_types.MultiProposedOutput {
	highProcess := job.highProcessSetup()

	multiSolveChannel := highProcess.RunForSeveral_CommonDifferent(job.printer, util.Optional_Empty[int](), cancelGenerate)

	proposalChannel := channel_op.Map_ChannelToChannel(4, multiSolveChannel, func(setResult withhighs.HighsMultiResult) multi_types.MultiProposedOutput {
		return job.makeOutputFromHighs(setResult, job.printer)
	})

	existingProposal := job.existingGearAsProposal()
	return channel_op.ChannelWithPrependedValues(proposalChannel, existingProposal)
}

func (job *MultiSetJob) proposalsUnderPermutation(tracker *util.TrackProgress, solutionsPerPermute int, cancel channel_op.CancelSignal) <-chan multi_types.MultiProposedOutput {
	estimate := job.estimateFixedPermutations()
	job.printer.Printf("PERMUTE SET COUNT %d\n", estimate)

	currentProgress := atomic.Uint64{}
	tracker.RunFromAtomicInt(&currentProgress, estimate)
	defer tracker.SetDone()

	permuteChannel := job.preparePermutations()

	proposalChannel := channel_op.MapMulti_ChannelToChannel_Cancellable(highsThreadCount, permuteChannel, cancel,
		func(permuteSet permuteSet, resultChannel chan<- multi_types.MultiProposedOutput) {
			job.runPermute(permuteSet, solutionsPerPermute, resultChannel, cancel)
			currentProgress.Add(1)
		},
	)

	existingProposal := job.existingGearAsProposal()
	return channel_op.ChannelWithPrependedValues(proposalChannel, existingProposal)
}

func (job *MultiSetJob) runPermute(permuteSet permuteSet, solutionsPerPermute int, resultChannel chan<- multi_types.MultiProposedOutput, cancel channel_op.CancelSignal) {
	printer := util.PrintRecorder_HoldAll()

	highProcess := job.highProcessSetupForPermute(permuteSet, printer)

	if solutionsPerPermute == 1 {
		future := highProcess.RunInterruptable(printer)
		channel_op.ChainCancel(cancel, future)
		result, hasResult := future.WaitForResult()
		if hasResult {
			resultChannel <- job.makeOutputFromHighs(result, printer)
		}
	} else {
		nextChan := highProcess.RunForSeveral_CommonDifferent(printer, util.Optional_OfValue(solutionsPerPermute), cancel)
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

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		for _, itemArray := range param.SemiFixedSlots {
			if len(itemArray) > 1 {
				panic("TryAll slots will be ignored, may lead to confusing results")
			}
		}
	}
}

func (job *MultiSetJob) highProcessSetup() withhighs.SolverHighsMultiProcess {
	highProcess := withhighs.SolverHighsMultiProcess{}

	optionsInputList := util.MapSliceAsNew(job.params, func(param *multiSetParamInternal) commonOptionsInput {
		return commonOptionsInput{param.Label, &param.itemOptions}
	})
	commonOptions := job.determineCommon(optionsInputList)
	highProcess.SetCommon(commonOptions)

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		highProcess.AddSetParam(withhighs.SolverHighsMultiParam{
			Label:          param.Label,
			ItemOptions:    param.itemOptions,
			Gear_model:     &param.Model,
			RatingMultiply: param.ratingMultiply,
		})
	}
	return highProcess
}

func (job *MultiSetJob) makeOutputFromHighs(multiResult withhighs.HighsMultiResult, printer *util.PrintRecorder) multi_types.MultiProposedOutput {
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
		proposed := multi_types.MultiProposedOutput{Id: uuid.NewString(), TotalRatingSum: totalRatingSum, Parts: outputs, Combo: combo}
		return proposed
	} else {
		panic("conflicted items")
	}
}

func (job *MultiSetJob) listInitialOutputs(bestOutputs <-chan multi_types.MultiProposedOutput) <-chan multi_types.MultiProposedOutput {
	return channel_op.PeekChannel(bestOutputs, func(best *multi_types.MultiProposedOutput) {
		job.printer.Printf("::::::::: MULTI RATING %.0f :::::::: %s ::::::::\n", best.TotalRatingSum, best.Id)
		for i, out := range best.Parts {
			job.printer.Println(job.params[i].Label)
			out.Report(job.printer)
		}
	})
}
