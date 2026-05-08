package multi

func (job *MultiSetJob) listInitialOutputs(bestOutputs []multiProposedOutput) {
	for _, best := range bestOutputs {
		job.printer.Printf("::::::::: MULTI RATING %.0f :::::::: %s ::::::::\n", best.totalRatingSum, best.id)
		for i, out := range best.parts {
			job.printer.Println(job.params[i].Label)
			out.Report(job.printer)
		}
	}
}
