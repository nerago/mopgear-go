package multi

import (
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"
)

const revisedExtraSetsExpectedEach = 2

func (job *MultiSetJob) makeRevised(param *MultiSetParam, filteredCombo commonCombo, outerTrackProgress *util.TrackProgress) []solver.SolveOutput {
	extraOutputs := make([]solver.SolveOutput, 0)

	revisedOutput := job.revisedSolveCombo(filteredCombo, param, param.PhasedAcceptable, outerTrackProgress)
	if revisedOutput.Success {
		extraOutputs = append(extraOutputs, revisedOutput)
	}

	phasedOutput := job.revisedSolveCombo(filteredCombo, param, true, outerTrackProgress)
	if phasedOutput.Success {
		extraOutputs = append(extraOutputs, phasedOutput)
	}

	// TODO reenchant process, maybe?

	return extraOutputs
}

func (job *MultiSetJob) revisedSolveCombo(combo commonCombo, param *MultiSetParam, phased bool, outerTrackProgress *util.TrackProgress) solver.SolveOutput {
	options := buildOptionsGivenCombo(param.itemOptions, combo)
	return solver.Solver(solver.SolveInput{
		ItemOptions:        &options,
		Model:              &param.Model,
		PhasedAcceptable:   phased,
		OuterTrackProgress: outerTrackProgress,
		SolveSize:          solver.SolveSize_Medium})
}
