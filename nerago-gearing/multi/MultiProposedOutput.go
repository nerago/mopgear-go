package multi

import (
	"maps"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"

	"github.com/google/uuid"
)

type MultiProposedOutput struct {
	Id             string
	TotalRatingSum uint64
	Outputs        []solver.SolveOutput
	Combo          CommonCombo
}

func (proposed *MultiProposedOutput) Equals(other *MultiProposedOutput) bool {
	if proposed.TotalRatingSum != other.TotalRatingSum {
		return false
	}

	for i := range proposed.Outputs {
		if !proposed.Outputs[i].Equals(&other.Outputs[i]) {
			return false
		}
	}
	return true
}

func makeUUID() string {
	return uuid.NewString()
}

func (job *MultiSetJob) makeProposedChannel(comboChannel <-chan CommonCombo) <-chan MultiProposedOutput {
	return util.Channel_TransformEach_Multi(solveThreadCount, comboChannel, func(combo CommonCombo, outputChannel chan<- MultiProposedOutput) {
		proposed := job.subSolveCombo(combo)
		if proposed != nil {
			outputChannel <- *proposed
		}
	})
}

func (job *MultiSetJob) subSolveCombo(combo CommonCombo) *MultiProposedOutput {
	var totalRatingSum uint64
	output := make([]solver.SolveOutput, len(job.params))

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		if param.IncludeInFirstPass {
			result := job.firstPassSolveCombo(combo, param)
			if !result.Success {
				job.printer.Println("UNEXPECTED SOLVE FAILURE FOR " + param.Label)
				// TODO make sure to update tracker
				return nil
			}
			totalRatingSum += result.ResultRating * param.ratingMultiply
			output[paramIndex] = result
		}
	}

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		if !param.IncludeInFirstPass {
			result := job.secondPassSolveCombo(combo, output, param)
			if !result.Success {
				job.printer.Println("UNEXPECTED SOLVE FAILURE FOR " + param.Label)
				// TODO make sure to update tracker
				return nil
			}
			totalRatingSum += result.ResultRating * param.ratingMultiply
			output[paramIndex] = result
		}
	}

	proposed := MultiProposedOutput{makeUUID(), totalRatingSum, output, combo}
	if job.multiSetFilter != nil && !job.multiSetFilter(proposed) {
		return nil
	}
	return &proposed

}

func (job *MultiSetJob) firstPassSolveCombo(combo CommonCombo, param *MultiSetParam) solver.SolveOutput {
	options := buildOptionsGivenCombo(param.itemOptions, combo)
	return solver.Solver(solver.SolveInput{
		ItemOptions:         &options,
		Model:               &param.Model,
		PhasedAcceptable:    param.PhasedAcceptable,
		EnableTrackProgress: false,
		SolveSize:           solver.SolveSize_PerItem})
}

func (job *MultiSetJob) secondPassSolveCombo(baseCombo CommonCombo, otherOutputList []solver.SolveOutput, param *MultiSetParam) solver.SolveOutput {
	// extend combo limitations further based on items chosen for other sets
	restrictedCombo := maps.Clone(baseCombo)
	for _, otherOutput := range otherOutputList {
		if otherOutput.Success {
			for item := range otherOutput.FullSet.Items().AllItemSeq() {
				restrictedCombo[item.ItemId()] = *item
			}
		}
	}

	options := buildOptionsGivenCombo(param.itemOptions, restrictedCombo)
	return solver.Solver(solver.SolveInput{
		ItemOptions:         &options,
		Model:               &param.Model,
		PhasedAcceptable:    param.PhasedAcceptable,
		EnableTrackProgress: false,
		SolveSize:           solver.SolveSize_PerItem})
}

func buildOptionsGivenCombo(allOptions items.FullOptionsMap, combo CommonCombo) items.FullOptionsMap {
	selectedOptions := items.FullOptionsMap{}
	choicesAdded := make(map[items.ItemId]bool)
	for slot, slotOptions := range allOptions {
		selectedOptions[slot] = buildOptionsGivenCombo_Slot(slotOptions, combo, choicesAdded)
	}
	return selectedOptions
}

func buildOptionsGivenCombo_Slot(slotOptions []items.FullItem, combo CommonCombo, choicesAdded map[items.ItemId]bool) []items.FullItem {
	selectedItems := make([]items.FullItem, 0, len(slotOptions))
	clear(choicesAdded)
	for i := range slotOptions {
		item := &slotOptions[i]
		itemId := item.ItemId()
		chosenVersion, hasChoice := combo[itemId]
		if !hasChoice {
			selectedItems = append(selectedItems, *item)
		} else if !choicesAdded[itemId] {
			selectedItems = append(selectedItems, chosenVersion)
			choicesAdded[itemId] = true
		}
	}
	return selectedItems
}
