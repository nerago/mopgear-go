package multi

import (
	"fmt"
	"maps"
	"time"

	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/solver/solve_highs"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"

	"github.com/google/uuid"
)

func (group *workingGroup) groupProposals(proposalMix *util_async.FutureChannelMixerContinuing[*multi_types.MultiProposedOutput], expectedCountAdder *util_async.FutureValueAdderInt, cancel *util_async.CancelSignalBasic) {
	if group.hasPermutes() {
		group.proposalsUnderPermutation(&group.task.Permute, proposalMix, expectedCountAdder, cancel)
	} else {
		group.proposalsGeneral(nil, proposalMix, expectedCountAdder, cancel)
	}

	if group.task.AlsoExistingEquipped {
		proposalMix.AddValue(group.existingGearAsProposal())
	}
	if group.task.AlsoSpecOptimums {
		proposalMix.AddChannel(group.additionalProposalsFromSpecOptimalBaseline(cancel))
	}
}

func (group *workingGroup) proposalsUnderPermutation(inputPermute *multi_types.InputPermute, proposalMix *util_async.FutureChannelMixerContinuing[*multi_types.MultiProposedOutput], expectedCountAdder *util_async.FutureValueAdderInt, cancel util_async.CancelSignal) {
	estimate := group.job.estimateFixedPermutations(inputPermute)
	group.job.printer.Printf("PERMUTE SET COUNT %d\n", estimate)

	optionEntriesList := group.job.buildPermutationOptions(inputPermute)
	permuteChannel := permuteAsChannel(optionEntriesList, cancel)

	util_async.ForEach_Channel(c_permuteThreadCount, permuteChannel,
		func(permuteSet permuteSet) {
			group.proposalsGeneral(&permuteSet, proposalMix, expectedCountAdder, cancel)
		},
	)
}

func (group *workingGroup) proposalsGeneral(permuteSet *permuteSet, proposalMix *util_async.FutureChannelMixerContinuing[*multi_types.MultiProposedOutput], expectedCountAdder *util_async.FutureValueAdderInt, cancel util_async.CancelSignal) {
	includeInterimResults := group.task.IncludeInterimResults

	var highProcess *solve_highs.SolverHighsMultiProcess
	if permuteSet != nil {
		highProcess = group.highProcessSetupForPermute(permuteSet, group.job.printer)
	} else {
		highProcess = group.highProcessSetupSingle()
	}

	multiSolveChannel, expectedCountFuture := highProcess.Run(
		group.job.input.TimeLimitEachSolve, group.job.printer,
		group.task.Alternates, group.task.AlternatesLimit,
		cancel, includeInterimResults)

	nextChannel := util_async.Map_ChannelToChannel(1, multiSolveChannel, func(setResult solve_highs.HighsMultiResult) *multi_types.MultiProposedOutput {
		return group.makeProposalFromHighs(setResult, group.job.printer, uuid.NewString())
	})

	proposalMix.AddChannel(nextChannel)
	expectedCountAdder.AddFuture(expectedCountFuture)
}

func (group *workingGroup) hasPermutes() bool {
	task := group.task

	if len(task.Permute.DistinctUsageGroups) > 0 {
		return true
	}

	if len(task.Permute.AlternateUpgradeChoices) > 0 {
		return true
	}

	if task.Permute.AlternateGemsEnableAsPermute {
		return true
	}

	if task.Permute.PermuteOnItemCountOptions {
		return true
	}

	for param := range util_collection.ForPointer(group.job.input.Param) {
		for _, itemArray := range param.ItemInputs.SemiFixedSlots {
			if len(itemArray) > 1 {
				return true
			}
		}
	}

	return false
}

func (group *workingGroup) highProcessSetupSingle() *solve_highs.SolverHighsMultiProcess {
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

func (group *workingGroup) makeProposalFromHighs(multiResult solve_highs.HighsMultiResult, printer *util.PrintRecorder, proposalId string) *multi_types.MultiProposedOutput {
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
		proposed := &multi_types.MultiProposedOutput{
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

func (job *MultiSetJob) listInitialOutputs(bestOutputs <-chan *multi_types.MultiProposedOutput) <-chan *multi_types.MultiProposedOutput {
	return util_async.PeekChannel_NoPointer(bestOutputs, func(prop *multi_types.MultiProposedOutput) {
		job.printer.Printf("::::::::: PROPOSED %f :::::::: %s ::::::::\n", prop.TotalRatingSum, prop.Id)
		job.printer.Printf("Weight Type %d\n", prop.WeightType)
		for label, out := range prop.Parts {
			prep := job.itemPrep[label]
			job.printer.Println(label)
			out.Report(prep.model, job.printer)
		}
	})
}

func (group *workingGroup) existingGearAsProposal() *multi_types.MultiProposedOutput {
	proposal := &multi_types.MultiProposedOutput{
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

func (group *workingGroup) additionalProposalsFromSpecOptimalBaseline(cancel util_async.CancelSignal) <-chan *multi_types.MultiProposedOutput {
	allWorking := util_async.SeqToChannel(maps.Values(group.workers))
	return util_async.MapMulti_ChannelToChannel_Cancellable(c_additionalProposal_threadCount, allWorking, cancel, func(work *specWorker, downstream chan<- *multi_types.MultiProposedOutput) {
		printer := util.PrintRecorder_HoldAll()

		highProcess := group.highProcessSetupRestrictedOnBaseline(work)

		resultChannel, _ := highProcess.Run(
			group.job.input.TimeLimitEachSolve, printer,
			multi_types.AlternateModeNone, util_collection.Optional_Empty[int](),
			cancel, false)

		for result := range resultChannel {
			proposalId := fmt.Sprintf("With-Best-%d-%s-%s", work.weightType, work.Label(), time.Now().Format("2006-01-02-15-04-05"))
			downstream <- group.makeProposalFromHighs(result, printer, proposalId)
		}

		group.job.printer.AppendOther(printer)
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
