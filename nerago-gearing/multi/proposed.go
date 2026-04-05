package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"

	"github.com/google/uuid"
)

type multiProposedOutput struct {
	id             string
	totalRatingSum uint64
	parts          []singleProposed
	combo          commonCombo
}

type singleProposed struct {
	fullSet      items.FullItemSet
	exists       bool
	spec         stats.SpecType
	outputId     string
	resultRating uint64
	model        *model.Model
}

func SingleProposed_FromOutput(out *solver.SolveOutput) singleProposed {
	if !out.Success {
		panic("should filter before putting as proposal")
	}
	return singleProposed{exists: true, spec: out.Input.Model.Spec, outputId: out.OutputId, resultRating: out.ResultRating, fullSet: out.FullSet, model: out.Input.Model}
}

func SingleProposed_FromEquip(equipMap items.FullEquipMap, param *MultiSetParam) singleProposed {
	set := items.FullItemSet_FromMap(equipMap)
	return singleProposed{exists: true, spec: param.Model.Spec, outputId: uuid.NewString(), resultRating: param.Model.CalcRatingFull(&set), fullSet: set, model: &param.Model}
}

func (single *singleProposed) Equals(b *singleProposed) bool {
	return single.exists == b.exists && single.resultRating == b.resultRating && single.fullSet.Equals(&b.fullSet)
}

func (single *singleProposed) Report(printer *util.PrintRecorder) {
	printer.Println(single.outputId)
	solver.ReportSet(printer, single.fullSet, single.resultRating, single.model)
}

func (proposed *multiProposedOutput) Equals(other *multiProposedOutput) bool {
	if proposed.totalRatingSum != other.totalRatingSum {
		return false
	}

	for i := range proposed.parts {
		if !proposed.parts[i].Equals(&other.parts[i]) {
			return false
		}
	}
	return true
}

func (job *MultiSetJob) makeProposedChannel(comboChannel <-chan commonCombo, comboCount uint64, trackProgress *util.TrackProgress) <-chan multiProposedOutput {
	trackProgress.RunOuterTracking(int(comboCount))
	return channel_op.TransformEach_ChannelToChannel(solveThreadCount, comboChannel, func(combo commonCombo, outputChannel chan<- multiProposedOutput) {
		proposed := job.subSolveCombo(&combo, trackProgress.MakeNested())
		if proposed != nil {
			outputChannel <- *proposed
		}
	})
}

func (job *MultiSetJob) subSolveCombo(combo *commonCombo, trackProgress *util.TrackProgress) *multiProposedOutput {
	var totalRatingSum uint64
	output := make([]singleProposed, len(job.params))
	trackProgress.RunOuterTracking(len(job.params))
	defer trackProgress.Stop()

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		if param.IncludeInFirstPass {
			result := job.firstPassSolveCombo(combo, param, trackProgress.MakeNested())
			if !result.Success {
				job.printer.Println("UNEXPECTED SOLVE FAILURE FOR " + param.Label + " " + result.FailureSummary)
				return nil
			}
			totalRatingSum += result.ResultRating * param.ratingMultiply
			output[paramIndex] = SingleProposed_FromOutput(&result)
		}
	}

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		if !param.IncludeInFirstPass {
			result := job.secondPassSolveCombo(combo, output, param, trackProgress.MakeNested())
			if !result.Success {
				job.printer.Println("UNEXPECTED SOLVE FAILURE FOR " + param.Label + " " + result.FailureSummary)
				return nil
			}
			totalRatingSum += result.ResultRating * param.ratingMultiply
			output[paramIndex] = SingleProposed_FromOutput(&result)
		}
	}

	proposed := multiProposedOutput{uuid.NewString(), totalRatingSum, output, *combo}
	if job.multiSetFilter != nil && !job.multiSetFilter(proposed) {
		return nil
	}
	return &proposed

}

func (job *MultiSetJob) firstPassSolveCombo(combo *commonCombo, param *MultiSetParam, tracker *util.TrackProgress) solver.SolveOutput {
	options := buildOptionsGivenCombo(param.itemOptions, combo)
	return solver.Solver(solver.SolveInput{
		ItemOptions:        &options,
		Model:              &param.Model,
		PhasedAcceptable:   param.PhasedAcceptable,
		OuterTrackProgress: tracker,
		SolveSize:          job.solveSizeProposal})
}

func (job *MultiSetJob) secondPassSolveCombo(baseCombo *commonCombo, otherOutputList []singleProposed, param *MultiSetParam, tracker *util.TrackProgress) solver.SolveOutput {
	// extend combo limitations further based on items chosen for other sets
	restrictedCombo := baseCombo.clone()
	for _, otherOutput := range otherOutputList {
		if otherOutput.exists {
			for item := range otherOutput.fullSet.Items().AllItemSeq() {
				restrictedCombo.addItem(item.ItemId(), item)
			}
		}
	}

	options := buildOptionsGivenCombo(param.itemOptions, &restrictedCombo)
	return solver.Solver(solver.SolveInput{
		ItemOptions:        &options,
		Model:              &param.Model,
		PhasedAcceptable:   param.PhasedAcceptable,
		OuterTrackProgress: tracker,
		SolveSize:          job.solveSizeProposal})
}

func buildOptionsGivenCombo(allOptions items.FullOptionsMap, combo *commonCombo) items.FullOptionsMap {
	selectedOptions := items.FullOptionsMap{}
	itemIdSeen := make(map[items.ItemId]bool)
	for slot, slotOptions := range allOptions {
		selectedOptions[slot] = buildOptionsGivenCombo_Slot(slotOptions, combo, itemIdSeen)
	}
	return selectedOptions
}

func buildOptionsGivenCombo_Slot(slotOptions []items.FullItem, combo *commonCombo, itemIdSeen map[items.ItemId]bool) []items.FullItem {
	selectedItems := make([]items.FullItem, 0, len(slotOptions))
	clear(itemIdSeen)
	for i := range slotOptions {
		item := &slotOptions[i]
		itemId := item.ItemId()
		hasRestriction, isAllowed, specifiedItem := combo.getValues(itemId)
		if !hasRestriction {
			selectedItems = append(selectedItems, *item)
		} else if !itemIdSeen[itemId] {
			if isAllowed {
				selectedItems = append(selectedItems, *specifiedItem)
			}
			itemIdSeen[itemId] = true
		}
	}
	return selectedItems
}
