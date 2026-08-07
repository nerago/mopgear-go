package multi_types

import (
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
)

type MultiProposedOutput struct {
	Id             string
	TotalRatingSum float64
	Parts          []SingleProposedOutput
	Combo          CommonCombo
	PermuteLabel   string
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
	SpecLabel    string
	OutputId     string
	ResultRating float64
}

func SingleProposed_FromItemSet(itemSet items.FullItemSet, outputId string, spec stats.SpecType, label string, rating float64) SingleProposedOutput {
	return SingleProposedOutput{Exists: true, Spec: spec, OutputId: outputId, ResultRating: rating, FullSet: itemSet, SpecLabel: label}
}

func (single *SingleProposedOutput) Equals(b *SingleProposedOutput) bool {
	return single.Exists == b.Exists && single.ResultRating == b.ResultRating && single.FullSet.Equals(&b.FullSet)
}

func (single *SingleProposedOutput) Report(model *gear_model.SpecModel, printer *util.PrintRecorder) {
	if single.OutputId != "" {
		printer.Println(single.OutputId)
	}
	tools.ReportSet(model, &single.FullSet, printer)
}
