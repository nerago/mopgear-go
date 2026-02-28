package multi

func (job *MultiSetJob) SuggestCulls(targetCount uint64) {
	job.prepareInitial()
	commonOptions := job.determineCommon()
	_ = job.makeCommonChannel(commonOptions, targetCount)
}
