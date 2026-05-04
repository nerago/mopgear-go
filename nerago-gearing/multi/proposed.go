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
	totalRatingSum float64
	parts          []singleProposed
	combo          commonCombo
}

type singleProposed struct {
	fullSet      items.FullItemSet
	exists       bool
	spec         stats.SpecType
	outputId     string
	resultRating float64
	model        *model.Model
}

func SingleProposed_FromOutput(out *solver.SolveOutput) singleProposed {
	if !out.Success {
		panic("should filter before putting as proposal")
	}
	return singleProposed{exists: true, spec: out.Input.Model.Spec, outputId: out.OutputId, resultRating: float64(out.ResultRating), fullSet: out.FullSet, model: out.Input.Model}
}

func SingleProposed_FromEquip(equipMap items.FullEquipMap, param *multiSetParamInternal) singleProposed {
	set := items.FullItemSet_FromMap(equipMap)
	return singleProposed{exists: true, spec: param.Model.Spec, outputId: uuid.NewString(), resultRating: float64(param.Model.CalcRatingFullAsFloat(&set)), fullSet: set, model: &param.Model}
}

func SingleProposed_FromItemSet(itemSet items.FullItemSet, model *model.Model) singleProposed {
	return singleProposed{exists: true, spec: model.Spec, outputId: uuid.NewString(), resultRating: float64(model.CalcRatingFullAsFloat(&itemSet)), fullSet: itemSet, model: model}
}

func (single *singleProposed) Equals(b *singleProposed) bool {
	return single.exists == b.exists && single.resultRating == b.resultRating && single.fullSet.Equals(&b.fullSet)
}

func (single *singleProposed) Report(printer *util.PrintRecorder) {
	printer.Println(single.outputId)
	solver.ReportSet(printer, single.fullSet, uint64(single.resultRating), single.model)
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
		job.subSolveCombo(&combo, trackProgress.MakeNested(), outputChannel)
	})
}

func (job *MultiSetJob) subSolveCombo(combo *commonCombo, trackProgress *util.TrackProgress, outputChannel chan<- multiProposedOutput) {
	var totalRatingSum float64
	output := make([]singleProposed, len(job.params))
	trackProgress.RunOuterTracking(len(job.params))
	defer trackProgress.Stop()

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		if param.IncludeInFirstPass {
			result := job.firstPassSolveCombo(combo, param, trackProgress.MakeNested())
			if solveFailure(result, param, job) {
				return
			}
			totalRatingSum += float64(result.ResultRating) * param.ratingMultiply
			output[paramIndex] = SingleProposed_FromOutput(&result)
		}
	}

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		if !param.IncludeInFirstPass {
			result := job.secondPassSolveCombo(combo, output, param, trackProgress.MakeNested())
			if solveFailure(result, param, job) {
				return
			}
			totalRatingSum += float64(result.ResultRating) * param.ratingMultiply
			output[paramIndex] = SingleProposed_FromOutput(&result)
		}
	}

	proposed := multiProposedOutput{uuid.NewString(), totalRatingSum, output, *combo}
	if job.multiSetFilter != nil && !job.multiSetFilter(proposed) {
		return
	}
	outputChannel <- proposed

}

func solveFailure(result solver.SolveOutput, param *multiSetParamInternal, job *MultiSetJob) bool {
	if result.Success {
		param.solveSuccessCount.Add(1)
		return false
	} else {
		good := param.solveSuccessCount.Load()
		bad := param.solveFailCount.Add(1)

		job.printer.Printf("UNEXPECTED SOLVE FAILURE FOR %s %s [good=%d bad=%d]\n", param.Label, result.FailureSummary, good, bad)
		return true
	}
}

func (job *MultiSetJob) firstPassSolveCombo(combo *commonCombo, param *multiSetParamInternal, tracker *util.TrackProgress) solver.SolveOutput {
	options := buildOptionsGivenCombo(param.itemOptions, combo)
	return solver.Solver(solver.SolveInput{
		ItemOptions:        &options,
		Model:              &param.Model,
		PhasedAcceptable:   param.PhasedAcceptable,
		OuterTrackProgress: tracker,
		SolveSize:          job.solveSizeProposal})
}

func (job *MultiSetJob) secondPassSolveCombo(baseCombo *commonCombo, otherOutputList []singleProposed, param *multiSetParamInternal, tracker *util.TrackProgress) solver.SolveOutput {
	// extend combo limitations further based on items chosen for other sets
	restrictedCombo := baseCombo.clone()
	for _, otherOutput := range otherOutputList {
		if otherOutput.exists {
			for item := range otherOutput.fullSet.Items().AllItemSeq() {
				restrictedCombo.addItem(item.ItemId(), item, Force_Optional)
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
		selectedOptions[slot] = buildOptionsGivenComboForSlot(slotOptions, combo, itemIdSeen)
	}
	return selectedOptions
}

func buildOptionsGivenComboForSlot(slotOptions []items.FullItem, combo *commonCombo, itemIdSeen map[items.ItemId]bool) []items.FullItem {
	clear(itemIdSeen)
	selectedItems := make([]items.FullItem, 0, len(slotOptions))

	for i := range slotOptions {
		item := &slotOptions[i]
		itemId := item.ItemId()
		forceMode, specifiedItem := combo.getAnySpecifications(itemId)
		if forceMode == Force_Unknown {
			selectedItems = append(selectedItems, *item)
		} else if !itemIdSeen[itemId] {
			switch forceMode {
			case Force_Optional:
				selectedItems = append(selectedItems, *specifiedItem)
			case Force_FixedWhereAvailable, Force_RequiredAlways:
				return []items.FullItem{*specifiedItem}
			case Force_Forbidden:
				// nothing
			default:
				panic("unexpected force value")
			}
			itemIdSeen[itemId] = true
		}
	}

	for itemId, entry := range combo.entryMap {
		if entry.forceMode == Force_RequiredAlways && !itemIdSeen[itemId] {
			panic("never saw forced item " + entry.Item.CreateString())
		}
	}

	return selectedItems
}
