package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"

	"github.com/google/uuid"
)

func (job *MultiSetJob) FindHighsResult() util.Optional[multi_types.MultiProposedOutput] {
	job.prepareInitial()
	highProcess := job.highProcessSetup()

	// TODO allow variants

	best := util_rank.BestCollector1[multi_types.MultiProposedOutput]{}

	setResults := highProcess.Run(job.printer)
	if setResults != nil {
		proposedOutput := job.makeOutputFromHighs(setResults)
		job.listInitialOutputs([]multi_types.MultiProposedOutput{proposedOutput})
		best.Offer(&proposedOutput, proposedOutput.TotalRatingSum)
	} else {
		job.printer.Println("FAILED")
	}

	return best.GetBestOptional()
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

func (job *MultiSetJob) makeOutputFromHighs(setResults []items.FullItemSet) multi_types.MultiProposedOutput {
	var totalRatingSum float64
	outputs := make([]multi_types.SingleProposedOutput, len(job.params))

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		itemSet := setResults[paramIndex]
		single := multi_types.SingleProposed_FromItemSet(itemSet, &param.Model)
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

func (job *MultiSetJob) FindSeveralHighsAndSim(runSize simulate.WowSim_RunSize) {
	job.prepareInitial()
	highProcess := job.highProcessSetup()

	// TODO allow variants

	// setResultList := highProcess.RunForSeveral_CommonDifferent(job.printer)
	setResultList := highProcess.RunForSeveral_CommonDifferent_WithParallel(job.printer)

	if setResultList != nil {
		proposalList := make([]multi_types.MultiProposedOutput, 0, len(setResultList))
		for _, setResult := range setResultList {
			proposedOutput := job.makeOutputFromHighs(setResult)
			proposalList = append(proposalList, proposedOutput)
		}

		proposalList = append(proposalList, job.existingGearAsProposal())

		util.RemoveDuplicatesFuncNotify(proposalList, func(a, b *multi_types.MultiProposedOutput) bool {
			return a.Equals(b)
		}, func(x *multi_types.MultiProposedOutput) {
			job.printer.Printf("Remove Duplicate %s\n", x.Id)
		})

		job.listInitialOutputs(proposalList)

		tracker := util.TrackProgress_Start()
		defer tracker.Stop()

		simList := job.prepareSimList(proposalList)
		job.runSims(simList, runSize, tracker)

		simResult := job.linkSimResults(proposalList, simList)
		job.reportSimResults(simResult)
		job.reportAsCsv(simResult)
	} else {
		job.printer.Println("FAILED")
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
