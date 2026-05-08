package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/google/uuid"
)

type multiProposedOutput struct {
	id             string
	totalRatingSum float64
	parts          []singleProposed
	combo          commonCombo
}

func (proposed *multiProposedOutput) findItemById(itemId items.ItemId) *items.FullItem {
	for _, part := range proposed.parts {
		item := part.fullSet.Items().FindItemId(itemId)
		if item != nil {
			return item
		}
	}
	return nil
}

type singleProposed struct {
	fullSet      items.FullItemSet
	exists       bool
	spec         stats.SpecType
	outputId     string
	resultRating float64
	model        *model.Model
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
