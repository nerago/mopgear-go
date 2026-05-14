package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

const (
	simThreadCount   = 4
	highsThreadCount = 10
)

type MultiSetJob struct {
	printer    *util.PrintRecorder
	params     []multiSetParamInternal
	fixedForge map[items.ItemId]stats.ReforgeRecipe
	bagsGear   loaders.EquippedArray
	simRunSize simulate.WowSim_RunSize
}

func MultiSetJob_Create(printer *util.PrintRecorder, simRunSize simulate.WowSim_RunSize) MultiSetJob {
	return MultiSetJob{
		printer:    printer,
		simRunSize: simRunSize,
	}
}

func (job *MultiSetJob) AddSetParam(param multi_types.MultiSetParam) {
	job.params = append(job.params, multiSetParamInternal{MultiSetParam: param, job: job})
	job.params[len(job.params)-1].init()
}

func (job *MultiSetJob) AddFixedForge(itemId items.ItemId, reforge stats.ReforgeRecipe) {
	if job.fixedForge == nil {
		job.fixedForge = make(map[items.ItemId]stats.ReforgeRecipe)
	}
	job.fixedForge[itemId] = reforge
}
