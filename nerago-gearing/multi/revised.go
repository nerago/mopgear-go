package multi

import (
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"
)

const revisedExtraSetsExpectedEach = 2

func (job *MultiSetJob) makeRevised(param *MultiSetParam, filteredCombo CommonCombo, outerTrackProgress *util.TrackProgress, printer *util.PrintRecorder) []solver.SolveOutput {
	extraOutputs := make([]solver.SolveOutput, 0, 3)

	revisedOutput := job.revisedSolveCombo(filteredCombo, param, param.PhasedAcceptable, outerTrackProgress)
	if revisedOutput.Success {
		printer.Println("REVISED")
		revisedOutput.Report(printer)
		extraOutputs = append(extraOutputs, revisedOutput)
	}

	phasedOutput := job.revisedSolveCombo(filteredCombo, param, true, outerTrackProgress)
	if phasedOutput.Success {
		printer.Println("PHASED")
		phasedOutput.Report(printer)
		extraOutputs = append(extraOutputs, phasedOutput)
	}

	// TODO reenchant process, maybe?

	return extraOutputs
}

func (job *MultiSetJob) revisedSolveCombo(combo CommonCombo, param *MultiSetParam, phased bool, outerTrackProgress *util.TrackProgress) solver.SolveOutput {
	options := buildOptionsGivenCombo(param.itemOptions, combo)
	return solver.Solver(solver.SolveInput{
		ItemOptions:        &options,
		Model:              &param.Model,
		PhasedAcceptable:   phased,
		OuterTrackProgress: outerTrackProgress,
		SolveSize:          solver.SolveSize_Medium})
}
