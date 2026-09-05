package updateProc

import (
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type weightSpecInternal struct {
	label  string
	param  SpecParam
	inputs updateInputs
	out    choiceOutput

	process *WeightUpdateProcess
}

type SpecParam struct {
	LabelOverride   string
	WeightFile1     string
	WeightFile2     string
	WeightFile3     string
	Model           gear_model.SpecModel
	FixStatsMode    weight_types.FixStatsRangeMode
	SubstituteItems []items.ItemId
}
