package multi

import (
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

const (
	generateThreadCount = 6
	solveThreadCount = 6
	evaluateThreadCount = 6
)

type MultiSetJob struct {
	printer        util.PrintRecorder
	params         []MultiSetParam
	fixedForge     map[uint32]stats.ReforgeRecipe
	bagsGear       []loaders.EquippedItem
	multiSetFilter func(MultiProposedOutput) bool
}

func (job *MultiSetJob) AddSetParam(param MultiSetParam) {
	param.init(job)
	job.params = append(job.params, param)
}

func (job *MultiSetJob) AddFixedForge(itemId uint32, reforge stats.ReforgeRecipe) {
	if job.fixedForge == nil {
		job.fixedForge = make(map[uint32]stats.ReforgeRecipe)
	}
	job.fixedForge[itemId] = reforge
}

func (job *MultiSetJob) SetMultiSetFilter(filter func(MultiProposedOutput) bool) {
	job.multiSetFilter = filter
}
