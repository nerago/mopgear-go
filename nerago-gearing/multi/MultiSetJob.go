package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

const (
	generateThreadCount = 6
	solveThreadCount    = 6
	evaluateThreadCount = 6
)

type MultiSetJob struct {
	printer            *util.PrintRecorder
	params             []multiSetParamInternal
	fixedForge         map[items.ItemId]stats.ReforgeRecipe
	specificAllowRates map[items.ItemId]specificAllowEntry
	bagsGear           loaders.EquippedArray
	multiSetFilter     func(multiProposedOutput) bool
	solveSizeProposal  solver.SolveSize
	solveSizeRevised   solver.SolveSize
}

func MultiSetJob_Create(printer *util.PrintRecorder, solveSizeProposal solver.SolveSize, solveSizeRevised solver.SolveSize) MultiSetJob {
	return MultiSetJob{
		printer:           printer,
		solveSizeProposal: solveSizeProposal,
		solveSizeRevised:  solveSizeRevised,
	}
}

func (job *MultiSetJob) AddSetParam(param MultiSetParam) {
	job.params = append(job.params, multiSetParamInternal{MultiSetParam: param, job: job})
	job.params[len(job.params)-1].init()
}

func (job *MultiSetJob) AddFixedForge(itemId items.ItemId, reforge stats.ReforgeRecipe) {
	if job.fixedForge == nil {
		job.fixedForge = make(map[items.ItemId]stats.ReforgeRecipe)
	}
	job.fixedForge[itemId] = reforge
}

func (job *MultiSetJob) SetMultiSetFilter(filter func(multiProposedOutput) bool) {
	job.multiSetFilter = filter
}

func (job *MultiSetJob) AddSpecificAllowRate(itemId items.ItemId, modeOff forceItemMode, modeOn forceItemMode, proportion float32) {
	if job.specificAllowRates == nil {
		job.specificAllowRates = make(map[items.ItemId]specificAllowEntry)
	}
	job.specificAllowRates[itemId] = specificAllowEntry{itemId: itemId, modeOff: modeOff, modeOn: modeOn, proportion: proportion}
}

type specificAllowEntry struct {
	itemId     items.ItemId
	modeOff    forceItemMode
	modeOn     forceItemMode
	proportion float32
}
