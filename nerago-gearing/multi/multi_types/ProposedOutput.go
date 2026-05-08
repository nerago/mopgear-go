package multi_types

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"

	"github.com/google/uuid"
)

type MultiProposedOutput struct {
	Id             string
	TotalRatingSum float64
	Parts          []SingleProposedOutput
	Combo          CommonCombo
}

func (proposed *MultiProposedOutput) FindItemById(itemId items.ItemId) *items.FullItem {
	for _, part := range proposed.Parts {
		item := part.FullSet.Items().FindItemId(itemId)
		if item != nil {
			return item
		}
	}
	return nil
}

func (proposed *MultiProposedOutput) Equals(other *MultiProposedOutput) bool {
	if proposed.TotalRatingSum != other.TotalRatingSum {
		return false
	}

	for i := range proposed.Parts {
		if !proposed.Parts[i].Equals(&other.Parts[i]) {
			return false
		}
	}
	return true
}

type SingleProposedOutput struct {
	FullSet      items.FullItemSet
	Exists       bool
	Spec         stats.SpecType
	OutputId     string
	ResultRating float64
	Model        *model.Model
}

func SingleProposed_FromEquip(equipMap items.FullEquipMap, param *MultiSetParam) SingleProposedOutput {
	set := items.FullItemSet_FromMap(equipMap)
	return SingleProposedOutput{Exists: true, Spec: param.Model.Spec, OutputId: uuid.NewString(), ResultRating: float64(param.Model.CalcRatingFullAsFloat(&set)), FullSet: set, Model: &param.Model}
}

func SingleProposed_FromItemSet(itemSet items.FullItemSet, model *model.Model) SingleProposedOutput {
	return SingleProposedOutput{Exists: true, Spec: model.Spec, OutputId: uuid.NewString(), ResultRating: float64(model.CalcRatingFullAsFloat(&itemSet)), FullSet: itemSet, Model: model}
}

func (single *SingleProposedOutput) Equals(b *SingleProposedOutput) bool {
	return single.Exists == b.Exists && single.ResultRating == b.ResultRating && single.FullSet.Equals(&b.FullSet)
}

func (single *SingleProposedOutput) Report(printer *util.PrintRecorder) {
	printer.Println(single.OutputId)
	tools.ReportSet(single.Model, &single.FullSet, single.ResultRating, printer)
}
