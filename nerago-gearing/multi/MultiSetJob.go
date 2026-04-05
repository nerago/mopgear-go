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
	params             []MultiSetParam
	fixedForge         map[items.ItemId]stats.ReforgeRecipe
	specificAllowRates map[items.ItemId]float32
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
	param.init(job)
	job.params = append(job.params, param)
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

func (job *MultiSetJob) AddSpecificAllowRate(itemId items.ItemId, proportion float32) {
	if job.specificAllowRates == nil {
		job.specificAllowRates = make(map[items.ItemId]float32)
	}
	job.specificAllowRates[itemId] = proportion
}
