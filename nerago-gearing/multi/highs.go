package multi

import (
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"sync/atomic"

	"github.com/google/uuid"
)

func (job *MultiSetJob) proposalsUnderPermutation(tracker *util.TrackProgress, solutionsPerPermute int) <-chan multi_types.MultiProposedOutput {
	estimate := job.estimateFixedPermutations()
	job.printer.Printf("PERMUTE SET COUNT %d\n", estimate)
	currentProgress := atomic.Uint64{}
	tracker.RunFromAtomicInt(&currentProgress, estimate)
	defer tracker.SetDone()

	permuteChannel := job.preparePermutations()

	setResultChannel := channel_op.MapMulti_ChannelToChannel(highsThreadCount, permuteChannel,
		func(permuteSet permuteSet, resultChannel chan<- multi_types.MultiProposedOutput) {
			printer := util.PrintRecorder_HoldAll()

			highProcess := job.highProcessSetupForPermute(permuteSet, printer)

			var nextChan <-chan withhighs.HighsMultiResult

			if solutionsPerPermute == 1 {
				nextChan = highProcess.RunInterruptable(printer, tracker)
			} else {
				nextChan = highProcess.RunForSeveral_CommonDifferent(printer, util.Optional_OfValue(solutionsPerPermute))
			}

			for res := range nextChan {
				resultChannel <- job.makeOutputFromHighs(res, printer)
			}

			job.printer.AppendOther(printer)
			currentProgress.Add(1)
		},
	)

	existingProposal := job.existingGearAsProposal()
	return channel_op.ChannelWithPrependedValues(setResultChannel, existingProposal)
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

func (job *MultiSetJob) proposalsToSimAndOutput(proposalChannel <-chan multi_types.MultiProposedOutput, tracker *util.TrackProgress) {
	proposalChannel = channel_op.Channel_RemoveDuplicatesFuncNotify(proposalChannel, func(a, b *multi_types.MultiProposedOutput) bool {
		return a.Equals(b)
	}, func(x *multi_types.MultiProposedOutput) {
		job.printer.Printf("Remove Duplicate %s\n", x.Id)
	})

	proposalChannel = job.listInitialOutputs(proposalChannel)
	proposalList := make([]multi_types.MultiProposedOutput, 0)
	proposalChannel = channel_op.TeeChannelToSlice(proposalChannel, &proposalList)

	simChannel := job.prepareSimList(proposalChannel)
	futureSimResultList := job.runSims(simChannel, tracker, -1)

	simResultList, gotResult := futureSimResultList.WaitForResultOrKeyPress()

	if gotResult {
		simMultiResults := job.linkSimResults(proposalList, simResultList)
		job.reportSimResults(simMultiResults)
		job.reportAsCsv(simMultiResults)

		job.suggestResultFromRankings(simMultiResults)
	} else {
		job.printer.Println("cancelled without result")
	}
}

func (job *MultiSetJob) makeOutputFromHighs(multiResult withhighs.HighsMultiResult, printer *util.PrintRecorder) multi_types.MultiProposedOutput {
	var totalRatingSum float64
	outputs := make([]multi_types.SingleProposedOutput, len(job.params))

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		itemSet := multiResult.ItemSets[paramIndex]
		outputId := multiResult.OutputId[paramIndex]
		single := multi_types.SingleProposed_FromItemSet(itemSet, outputId, &param.Model)
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
