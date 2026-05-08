package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"

	"github.com/google/uuid"
)

func (job *MultiSetJob) FindHighsResult() util.Optional[multiProposedOutput] {
	job.prepareInitial()
	highProcess := job.highProcessSetup()

	// TODO allow variants

	best := util_rank.BestCollector1[multiProposedOutput]{}

	setResults := highProcess.Run(job.printer)
	if setResults != nil {
		proposedOutput := job.makeOutputFromHighs(setResults)
		job.listInitialOutputs([]multiProposedOutput{proposedOutput})
		best.Offer(&proposedOutput, uint64(proposedOutput.totalRatingSum)) // TODO float
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

func (job *MultiSetJob) makeOutputFromHighs(setResults []items.FullItemSet) multiProposedOutput {
	var totalRatingSum float64
	outputs := make([]singleProposed, len(job.params))

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		itemSet := setResults[paramIndex]
		single := SingleProposed_FromItemSet(itemSet, &param.Model)
		outputs[paramIndex] = single
		totalRatingSum += single.resultRating * param.ratingMultiply
	}

	if checkNoConflicts(outputs) {
		combo := job.determineComboFromScratch(outputs)
		proposed := multiProposedOutput{uuid.NewString(), totalRatingSum, outputs, combo}
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
		proposalList := make([]multiProposedOutput, 0, len(setResultList))
		for _, setResult := range setResultList {
			proposedOutput := job.makeOutputFromHighs(setResult)
			proposalList = append(proposalList, proposedOutput)
		}

		proposalList = append(proposalList, job.existingGearAsProposal())

		util.RemoveDuplicatesFuncNotify(proposalList, func(a, b *multiProposedOutput) bool {
			return a.Equals(b)
		}, func(x *multiProposedOutput) {
			job.printer.Printf("Remove Duplicate %s\n", x.id)
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
