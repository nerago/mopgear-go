package multi

func (job *MultiSetJob) SuggestCulls(targetCount uint64) {
	job.prepareInitial()
	commonOptions := job.determineCommon()
	comboChannel := job.makeCommonChannel(commonOptions, targetCount)
	proposedChannel := job.makeProposedChannel(comboChannel)
	bestOutputs := job.evalutateTopN(proposedChannel, 10)
	for _, best := range bestOutputs {
		for i, out := range best.Outputs {
			job.printer.Println(job.params[i].Label)
			out.Report(&job.printer)
		}
	}
}
