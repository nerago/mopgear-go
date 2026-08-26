package multi_types

import (
	"iter"
	"maps"

	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type MultiProposedOutput struct {
	Id             string
	TotalRatingSum float64
	Parts          map[string]SingleProposedOutput
	Combo          CommonCombo
	PermuteLabel   string
	WeightType     weight_types.WeightType
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

func (proposed *MultiProposedOutput) SeqItem() iter.Seq[*items.FullItem] {
	return func(yield func(*items.FullItem) bool) {
		for _, part := range proposed.Parts {
			for item := range part.FullSet.Items().AllItemSeq() {
				if !yield(item) {
					return
				}
			}
		}
	}
}

func (proposed *MultiProposedOutput) Equals(other *MultiProposedOutput) bool {
	if proposed.TotalRatingSum != other.TotalRatingSum {
		return false
	}
	return maps.EqualFunc(proposed.Parts, other.Parts, func(a, b SingleProposedOutput) bool {
		return a.Equals(&b)
	})
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
	return single.Exists == b.Exists && single.Spec == b.Spec && single.FullSet.Equals(&b.FullSet)
}

func (single *SingleProposedOutput) Report(model *gear_model.SpecModel, printer util.Printable) {
	if single.OutputId != "" {
		printer.Println(single.OutputId)
	}
	tools.ReportSet(model, &single.FullSet, printer)
}
