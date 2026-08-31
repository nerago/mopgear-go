package updateProc

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type updateInputs struct {
	simTypes    []stats.SimType
	statTypes   []stats.StatType
	targetRatio weight_types.SimPriorityBasic

	dataGrid, dataRand, dataFit, dataAll []weight_types.WeightInput
}

func (ui *updateInputs) freeData() {
	ui.dataGrid = nil
	ui.dataRand = nil
	ui.dataFit = nil
	ui.dataAll = nil
}
