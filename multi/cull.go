package multi

import (
	"cmp"
	"paladin_gearing_go/util"
	"slices"
)

func (job *MultiSetJob) SuggestCulls(targetCount uint64, topCapture int) {
	bestOutputs := job.runForTopN(targetCount, topCapture)

	for _, best := range bestOutputs {
		job.printer.Printf("::::::::: MULTI RATING %d ::::::::\n", best.TotalRatingSum)
		for i, out := range best.Outputs {
			job.printer.Println(job.params[i].Label)
			out.Report(&job.printer)
		}
	}

	job.cullingMakeRevisions(bestOutputs)
	job.cullingReport()
}

func (job *MultiSetJob) runForTopN(targetCount uint64, topCapture int) []MultiProposedOutput {
	job.prepareInitial()
	commonOptions := job.determineCommon()

	trackProgress := util.TrackProgress_Start()

	comboChannel := job.makeCommonChannel(commonOptions, targetCount, trackProgress)
	proposedChannel := job.makeProposedChannel(comboChannel)
	bestOutputs := job.evalutateTopN(proposedChannel, topCapture)

	trackProgress.Stop()
	return bestOutputs
}

func (job *MultiSetJob) cullingMakeRevisions(proposedList []MultiProposedOutput) {
	job.printer.Printf("MAKE REVISIONS FOR %d\n", len(proposedList))

	expectedSets := len(proposedList) * len(job.params) * revisedExtraSetsExpectedEach
	trackProgress := util.TrackProgress_Start()
	trackProgress.RunOuterTracking(expectedSets)
	defer trackProgress.Stop()

	util.Void_IterateEach_Multi_Blocking(generateThreadCount, proposedList, func(prior MultiProposedOutput) {
		printer := util.PrintRecorder_HoldAll()
		revisedCommon := job.revisedComboActuallyUsed(prior.Outputs, prior.Combo, printer)
		for i := range prior.Outputs {
			draft := &prior.Outputs[i]
			param := &job.params[i]

			param.seenInSolutions.Add(&draft.FullSet)

			revised := job.makeRevised(param, revisedCommon, trackProgress)
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
		itemId uint32
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

	param.job.printer.Printf("EXTRAS USED %s\n", param.Label)
	for _, info := range extraInfo {
		if info.count == 0 {
			param.job.printer.Printf("%d 0 NONE\n", info.itemId)
		} else {
			param.job.printer.Printf("%d %d\n", info.itemId, info.count)
		}
	}
}
