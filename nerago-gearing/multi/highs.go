package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/withhighs"

	"github.com/google/uuid"
)

func (job *MultiSetJob) FindHighsResult() {
	highProcess := job.highProcessSetup()

	setResults := highProcess.Run(job.printer)
	if setResults != nil {
		proposedOutput := job.makeOutputFromHighs(setResults)
		job.listInitialOutputs([]multiProposedOutput{proposedOutput})
	} else {
		job.printer.Println("FAILED")
	}
}

func (job *MultiSetJob) highProcessSetup() withhighs.SolverHighsMultiProcess {
	highProcess := withhighs.SolverHighsMultiProcess{}

	job.prepareInitial()
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
	var totalRatingSum uint64
	outputs := make([]singleProposed, len(job.params))

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		itemSet := setResults[paramIndex]
		single := SingleProposed_FromItemSet(itemSet, &param.Model)
		outputs[paramIndex] = single
		totalRatingSum += single.resultRating * param.ratingMultiply
	}

	if checkNoConflicts(outputs) {
		combo := job.determineComboFromScratch(outputs, comboType_highs)
		proposed := multiProposedOutput{uuid.NewString(), totalRatingSum, outputs, combo}
		return proposed
	} else {
		panic("conflicted items")
	}
}

func (job *MultiSetJob) FindSeveralHighsAndSim(runSize simulate.WowSim_RunSize) {
	highProcess := job.highProcessSetup()

	// var setResults [][]items.FullItemSet = highProcess.RunForSeveral(job.printer, 10)
	_ = highProcess.RunForSeveral(job.printer, 10)
	// if setResults != nil {
	// 	proposedOutput := job.makeOutputFromHighs(setResults)
	// 	job.listInitialOutputs([]multiProposedOutput{proposedOutput})
	// } else {
	// 	job.printer.Println("FAILED")
	// }
}
