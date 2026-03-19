package multi

import (
	"cmp"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"slices"
)

func (job *MultiSetJob) SuggestCulls(targetCount uint64, topCapture int) {
	topTracker := util.TrackProgress_Start()
	topTracker.RunOuterTracking(2)
	defer topTracker.Stop()

	bestOutputs := job.runForTopN(targetCount, topCapture, topTracker.MakeNested())
	job.listInitialOutputs(bestOutputs)
	job.cullingMakeRevisions(bestOutputs, topTracker.MakeNested())
	job.cullingReport()
}

func (job *MultiSetJob) listInitialOutputs(bestOutputs []MultiProposedOutput) {
	for _, best := range bestOutputs {
		job.printer.Printf("::::::::: MULTI RATING %d :::::::: %s ::::::::\n", best.TotalRatingSum, best.Id)
		for i, out := range best.Outputs {
			job.printer.Println(job.params[i].Label)
			out.Report(job.printer)
		}
	}
}

func (job *MultiSetJob) runForTopN(targetCount uint64, topCapture int, trackProgress *util.TrackProgress) []MultiProposedOutput {
	job.prepareInitial()
	commonOptions := job.determineCommon()

	comboChannel := job.makeCommonChannel(commonOptions, targetCount, trackProgress)
	proposedChannel := job.makeProposedChannel(comboChannel)
	bestOutputs := job.evalutateTopN(proposedChannel, topCapture)

	trackProgress.Stop()
	return bestOutputs
}

func (job *MultiSetJob) cullingMakeRevisions(proposedList []MultiProposedOutput, trackProgress *util.TrackProgress) {
	job.printer.Printf("MAKE REVISIONS FOR %d\n", len(proposedList))

	expectedSets := len(proposedList) * len(job.params) * revisedExtraSetsExpectedEach
	trackProgress.RunOuterTracking(expectedSets)
	defer trackProgress.Stop()

	util.Void_IterateEach_Multi_Blocking(generateThreadCount, proposedList, func(prior *MultiProposedOutput) {
		printer := util.PrintRecorder_HoldAll()
		revisedCommon := job.revisedComboActuallyUsed(prior.Outputs, prior.Combo, printer)
		for i := range prior.Outputs {
			draft := &prior.Outputs[i]
			param := &job.params[i]

			param.seenInSolutions.Add(&draft.FullSet)

			revised := job.makeRevised(param, revisedCommon, trackProgress, printer)
			for _, newOutput := range revised {
				param.seenInSolutions.Add(&newOutput.FullSet)
			}
		}
		job.printer.AppendOther(printer)
	})
}

func (job *MultiSetJob) cullingReport() {
	for paramIndex := range job.params {
		job.params[paramIndex].cullingReport()
	}
}

func (param *MultiSetParam) cullingReport() {
	type extraInfoStruct struct {
		itemId items.ItemId
		count  uint32
	}

	extraInfo := make([]extraInfoStruct, 0, len(param.extraItems))
	for _, itemId := range param.extraItems {
		// TODO also include equipped?
		seenCount := param.seenInSolutions.content[itemId]
		info := extraInfoStruct{itemId: itemId, count: seenCount}
		extraInfo = append(extraInfo, info)
	}

	slices.SortFunc(extraInfo, func(a, b extraInfoStruct) int {
		return cmp.Or(cmp.Compare(a.count, b.count), cmp.Compare(a.itemId, b.itemId))
	})

	// TODO is there concurrent goroutines updating?
	param.job.printer.Printf("EXTRAS USED %s\n", param.Label)
	for _, info := range extraInfo {
		if info.count == 0 {
			param.job.printer.Printf("%d 0 NONE\n", info.itemId)
		} else {
			param.job.printer.Printf("%d %d\n", info.itemId, info.count)
		}
	}
}
