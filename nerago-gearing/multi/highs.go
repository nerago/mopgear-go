package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/google/uuid"
)

func (job *MultiSetJob) FindHighsResult() {
	job.prepareInitial()
	highProcess := job.highProcessSetup()

	setResults := highProcess.Run(job.printer)
	if setResults != nil {
		proposedOutput := job.makeOutputFromHighs(setResults)
		job.listInitialOutputs([]multiProposedOutput{proposedOutput})
	} else {
		job.printer.Println("FAILED")
	}
}

func (job *MultiSetJob) DetermineWhatRatingsLeadToResult(commonChoices map[items.ItemId]stats.ReforgeRecipe) {
	job.prepareInitial()

	job.printer.Println("{{ FIND OPTIMUM ACCORDING TO HIGHS }}")
	highProcess := job.highProcessSetup()
	optimumMultiSet := highProcess.Run(job.printer)

	setBonus := model.SetBonus_Named("Plate of the Lightning Emperor")

	oldMultipliers := []float64{}
	equippedRatings := []float64{}
	bestUnderOldMultsRatings := []float64{}
	optimumIndependentRatings := []float64{}

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		equippedSet := items.FullItemSet_FromMap(param.exactEquippedGear)
		equippedRating := param.Model.CalcRatingFull(&equippedSet)
		bestIndependentRating := param.baselineResultHighs.ResultRating
		bestMultiRating := param.Model.CalcRatingFull(&optimumMultiSet[paramIndex])
		job.printer.Printf("MULTI PENALTY %30s equip=%10d multi=%10d indep=%10d bestRatio=%.6f equipRatio=%.6f setitems=%d,%d,%d mult_calc=%f\n", param.Label,
			equippedRating,
			bestMultiRating,
			bestIndependentRating,
			float64(bestMultiRating)/float64(bestIndependentRating),
			float64(equippedRating)/float64(bestIndependentRating),
			setBonus.CountInAnySet(equippedSet.Items()),
			setBonus.CountInAnySet(optimumMultiSet[paramIndex].Items()),
			setBonus.CountInAnySet(param.baselineResultHighs.FullSet.Items()),
			param.ratingMultiply,
		)

		oldMultipliers = append(oldMultipliers, param.RequestRatingPercent)
		equippedRatings = append(equippedRatings, float64(equippedRating))
		bestUnderOldMultsRatings = append(bestUnderOldMultsRatings, float64(bestMultiRating))
		optimumIndependentRatings = append(optimumIndependentRatings, float64(bestIndependentRating))
	}

	withhighs.FindSuggestedRatingMultipliers(oldMultipliers, equippedRatings, bestUnderOldMultsRatings, optimumIndependentRatings, job.printer)

	// so if we imagine we're multiplying these to

	// what if baseline is good, vs highs version

	// setResults := highProcess.DetermineWhatRatingsLeadToResult(job.printer, commonChoices)
	// if setResults != nil {
	// 	proposedOutput := job.makeOutputFromHighs(setResults)
	// 	job.listInitialOutputs([]multiProposedOutput{proposedOutput})
	// } else {
	// 	job.printer.Println("FAILED")
	// }
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
		combo := job.determineComboFromScratch(outputs, comboType_highs)
		proposed := multiProposedOutput{uuid.NewString(), totalRatingSum, outputs, combo}
		return proposed
	} else {
		panic("conflicted items")
	}
}

func (job *MultiSetJob) FindSeveralHighsAndSim(runSize simulate.WowSim_RunSize) {
	job.prepareInitial()
	highProcess := job.highProcessSetup()

	setResultList := highProcess.RunForSeveral_CommonDifferent(job.printer)

	if setResultList != nil {
		proposalList := make([]multiProposedOutput, 0, len(setResultList))
		for _, setResult := range setResultList {
			proposedOutput := job.makeOutputFromHighs(setResult)
			proposalList = append(proposalList, proposedOutput)
		}

		util.RemoveDuplicatesFunc(proposalList, func(a, b *multiProposedOutput) bool {
			return a.Equals(b)
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
