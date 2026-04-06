package multi

import (
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"
)

const revisedExtraSetsExpectedEach = 2

func (job *MultiSetJob) makeRevised(param *multiSetParamInternal, revisedCommon *commonCombo, outerTrackProgress *util.TrackProgress, printer *util.PrintRecorder) []singleProposed {
	extraOutputs := make([]singleProposed, 0, 3)

	revisedOutput := job.revisedSolveCombo(revisedCommon, param, param.PhasedAcceptable, outerTrackProgress)
	if revisedOutput.Success {
		printer.Println("REVISED")
		revisedOutput.Report(printer)
		extraOutputs = append(extraOutputs, SingleProposed_FromOutput(&revisedOutput))
	}

	phasedOutput := job.revisedSolveCombo(revisedCommon, param, true, outerTrackProgress)
	if phasedOutput.Success {
		printer.Println("PHASED")
		phasedOutput.Report(printer)
		extraOutputs = append(extraOutputs, SingleProposed_FromOutput(&phasedOutput))
	}

	// TODO reenchant process, maybe?

	return extraOutputs
}

func (job *MultiSetJob) revisedSolveCombo(combo *commonCombo, param *multiSetParamInternal, phased bool, outerTrackProgress *util.TrackProgress) solver.SolveOutput {
	options := buildOptionsGivenCombo(param.itemOptions, combo)
	return solver.Solver(solver.SolveInput{
		ItemOptions:        &options,
		Model:              &param.Model,
		PhasedAcceptable:   phased,
		OuterTrackProgress: outerTrackProgress,
		SolveSize:          job.solveSizeRevised})
}
