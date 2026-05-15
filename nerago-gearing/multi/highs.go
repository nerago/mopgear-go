package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"paladin_gearing_go/util/util_rank"
	"sync/atomic"

	"github.com/google/uuid"
)

func (job *MultiSetJob) FindHighsResult_Sample() util.Optional[multi_types.MultiProposedOutput] {
	job.prepareInitial()
	highProcess := job.highProcessSetup()

	best := util_rank.BestCollector1[multi_types.MultiProposedOutput]{}

	setResults := highProcess.RunForSeveral_NextObjective(job.printer, 6)
	// setResults := highProcess.RunForSeveral_CommonDifferent_Sampling(job.printer, 6)
	if setResults != nil {
		proposedOutput := util.CastSliceAsNew(setResults, func(x *[]items.FullItemSet) multi_types.MultiProposedOutput {
			return job.makeOutputFromHighs(*x, job.printer)
		})
		job.listInitialOutputs(proposedOutput)
		for _, x := range proposedOutput {
			best.Offer(&x, x.TotalRatingSum)
		}
	} else {
		job.printer.Println("FAILED")
	}

	return best.GetBestOptional()
}

func (job *MultiSetJob) FindHighsResultPerPermute(solutionsPerPermute int) {
	job.prepareInitial()

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(2)
	defer tracker.Stop()

	setResultList := job.proposalsUnderPermutation(tracker.MakeNested(), solutionsPerPermute)

	job.proposalsToSimAndOutput(setResultList, tracker.MakeNested())
}

func (job *MultiSetJob) proposalsUnderPermutation(tracker *util.TrackProgress, solutionsPerPermute int) []multi_types.MultiProposedOutput {
	commonOptions := job.determineCommon() // TODO common after resolving options might be good

	estimate := job.estimateFixedPermutations()
	job.printer.Printf("PERMUTE SET COUNT %d\n", estimate)
	currentProgress := atomic.Uint64{}
	tracker.RunFromAtomicInt(&currentProgress, estimate)
	defer tracker.Stop()

	permuteChannel := job.preparePermutations()
	setResultList := channel_op.Map_ChannelToSlice(highsThreadCount, permuteChannel,
		func(permuteSet permuteSet, resultChannel chan<- multi_types.MultiProposedOutput) {
			printer := util.PrintRecorder_HoldAll()

			highProcess := job.highProcessSetupForPermute(permuteSet, commonOptions, printer)

			if solutionsPerPermute == 1 {
				setResults := highProcess.Run(printer)
				if setResults != nil {
					resultChannel <- job.makeOutputFromHighs(setResults, printer)
				}
			} else {
				setResultsList := highProcess.RunForSeveral_CommonDifferent_Sampling(printer, solutionsPerPermute)
				for _, setResults := range setResultsList {
					resultChannel <- job.makeOutputFromHighs(setResults, printer)
				}
			}

			job.printer.AppendOther(printer)
			currentProgress.Add(1)
		},
	)
	setResultList = append(setResultList, job.existingGearAsProposal())
	return setResultList
}

func (job *MultiSetJob) FindSeveralHighsAndSim() {
	job.prepareInitial()
	highProcess := job.highProcessSetup()

	setResultList := highProcess.RunForSeveral_CommonDifferent_WithParallel(job.printer)

	if setResultList != nil {
		proposalList := make([]multi_types.MultiProposedOutput, 0, len(setResultList))
		for _, setResult := range setResultList {
			proposedOutput := job.makeOutputFromHighs(setResult, job.printer)
			proposalList = append(proposalList, proposedOutput)
		}
		proposalList = append(proposalList, job.existingGearAsProposal())

		// TODO tracker covers highs part too
		tracker := util.TrackProgress_Start()
		defer tracker.Stop()
		job.proposalsToSimAndOutput(proposalList, tracker)
	} else {
		job.printer.Println("FAILED")
	}
}

func (job *MultiSetJob) highProcessSetup() withhighs.SolverHighsMultiProcess {
	highProcess := withhighs.SolverHighsMultiProcess{}

	commonOptions := job.determineCommon()
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

func (job *MultiSetJob) proposalsToSimAndOutput(proposalList []multi_types.MultiProposedOutput, tracker *util.TrackProgress) {
	util.RemoveDuplicatesFuncNotify(proposalList, func(a, b *multi_types.MultiProposedOutput) bool {
		return a.Equals(b)
	}, func(x *multi_types.MultiProposedOutput) {
		job.printer.Printf("Remove Duplicate %s\n", x.Id)
	})

	job.listInitialOutputs(proposalList)

	simList := job.prepareSimList(proposalList)
	simResultList := job.runSims(simList, tracker)

	simMultiResults := job.linkSimResults(proposalList, simResultList)
	job.reportSimResults(simMultiResults)
	job.reportAsCsv(simMultiResults)
}

func (job *MultiSetJob) makeOutputFromHighs(setResults []items.FullItemSet, printer *util.PrintRecorder) multi_types.MultiProposedOutput {
	var totalRatingSum float64
	outputs := make([]multi_types.SingleProposedOutput, len(job.params))

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		itemSet := setResults[paramIndex]
		single := multi_types.SingleProposed_FromItemSet(itemSet, &param.Model)
		single.Report(printer)
		outputs[paramIndex] = single
		totalRatingSum += single.ResultRating * param.ratingMultiply
	}

	if checkNoConflicts(outputs) {
		combo := multi_types.CommonCombo_FromProposed(outputs)
		proposed := multi_types.MultiProposedOutput{Id: uuid.NewString(), TotalRatingSum: totalRatingSum, Parts: outputs, Combo: combo}
		return proposed
	} else {
		panic("conflicted items")
	}
}

func (job *MultiSetJob) listInitialOutputs(bestOutputs []multi_types.MultiProposedOutput) {
	for _, best := range bestOutputs {
		job.printer.Printf("::::::::: MULTI RATING %.0f :::::::: %s ::::::::\n", best.TotalRatingSum, best.Id)
		for i, out := range best.Parts {
			job.printer.Println(job.params[i].Label)
			out.Report(job.printer)
		}
	}
}
