package multi

import (
	"maps"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"

	"github.com/google/uuid"
)

type MultiProposedOutput struct {
	id      string
	outputs []solver.SolveOutput
	combo   commonCombo
}

func makeUUID() string {
	return uuid.NewString()
}

func (job *MultiSetJob) makeProposedChannel(comboChannel <-chan commonCombo) <-chan MultiProposedOutput {
	return util.Channel_TransformEach_Multi(solveThreadCount, comboChannel, job.subSolveCombo)
}

func (job *MultiSetJob) subSolveCombo(combo commonCombo, outputChannel chan<- MultiProposedOutput) {
	allGood := true
	output := make([]solver.SolveOutput, len(job.params))
	for paramIndex, param := range job.params {
		if param.IncludeInFirstPass {
			output[paramIndex] = job.firstPassSolveCombo(combo, param, &allGood)
		}
	}
	for paramIndex, param := range job.params {
		if !param.IncludeInFirstPass {
			output[paramIndex] = job.secondPassSolveCombo(combo, output, param, &allGood)
		}
	}
	if allGood {
		outputChannel <- MultiProposedOutput{makeUUID(), output, combo}
	}
}

func (job *MultiSetJob) firstPassSolveCombo(combo commonCombo, param MultiSetParam, allGood *bool) solver.SolveOutput {
	options := buildOptionsGivenCombo(param.itemOptions, combo)
	result := solver.Solver(&options, &param.Model, param.PhasedAcceptable)
	if !result.Success {
		job.printer.Println("UNEXPECTED SOLVE FAILURE FOR " + param.Label)
		*allGood = false
	}
	return result
	// param.seenInSolutions.Add(&param.baselineResult.FullSet) or only on best ones?
}

func (job *MultiSetJob) secondPassSolveCombo(baseCombo commonCombo, otherOutputList []solver.SolveOutput, param MultiSetParam, allGood *bool) solver.SolveOutput {
	// extend combo limitations further based on items chosen for other sets
	restrictedCombo := maps.Clone(baseCombo)
	for _, otherOutput := range otherOutputList {
		if otherOutput.Success {
			for item := range otherOutput.FullSet.Items.AllItemSeq() {
				restrictedCombo[item.ItemId()] = *item
			}
		}
	}

	options := buildOptionsGivenCombo(param.itemOptions, restrictedCombo)
	result := solver.Solver(&options, &param.Model, param.PhasedAcceptable)
	if !result.Success {
		job.printer.Println("UNEXPECTED SOLVE FAILURE FOR " + param.Label)
		*allGood = false
	}
	return result
}

func buildOptionsGivenCombo(allOptions items.FullOptionsMap, combo commonCombo) items.FullOptionsMap {
	selectedOptions := items.FullOptionsMap{}
	for slot, slotOptions := range allOptions {
		selectedOptions[slot] = buildOptionsGivenCombo_Slot(slotOptions, combo)
	}
	return selectedOptions
}

func buildOptionsGivenCombo_Slot(slotOptions []items.FullItem, combo commonCombo) []items.FullItem {
	selectedItems := make([]items.FullItem, 0, len(slotOptions))
	choicesAdded := make(map[uint32]bool)
	for _, item := range slotOptions {
		itemId := item.ItemId()
		chosenVersion, hasChoice := combo[itemId]
		if !hasChoice {
			selectedItems = append(selectedItems, item)
		} else if !choicesAdded[itemId] {
			selectedItems = append(selectedItems, chosenVersion)
			choicesAdded[itemId] = true
		}
	}
	return selectedItems
}
