package multi

import (
	"cmp"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"paladin_gearing_go/util/util_rank"
	"slices"
)

func (job *MultiSetJob) SuggestCulls(targetCount uint64, topCapture uint64) {
	topTracker := util.TrackProgress_Start()
	topTracker.RunOuterTracking(2)
	defer topTracker.Stop()

	bestOutputs := job.runForTopN(targetCount, topCapture, topTracker.MakeNested())
	job.listInitialOutputs(bestOutputs)
	job.cullingMakeRevisions(bestOutputs, topTracker.MakeNested())
	job.cullingReport()
}

func (job *MultiSetJob) listInitialOutputs(bestOutputs []multiProposedOutput) {
	for _, best := range bestOutputs {
		job.printer.Printf("::::::::: MULTI RATING %.0f :::::::: %s ::::::::\n", best.totalRatingSum, best.id)
		for i, out := range best.parts {
			job.printer.Println(job.params[i].Label)
			out.Report(job.printer)
		}
	}
}

func (job *MultiSetJob) cullingMakeRevisions(proposedList []multiProposedOutput, trackProgress *util.TrackProgress) {
	job.printer.Printf("MAKE REVISIONS FOR %d\n", len(proposedList))

	expectedSets := len(proposedList) * len(job.params) * revisedExtraSetsExpectedEach
	trackProgress.RunOuterTracking(expectedSets)
	defer trackProgress.Stop()

	type bestOption struct {
		specIndex int
		set       *items.FullItemSet
	}

	optionChannel := channel_op.Map_SliceToChannel(generateThreadCount, proposedList, func(prior *multiProposedOutput, downstream chan<- bestOption) {
		printer := util.PrintRecorder_HoldAll()
		printer.Printf(">>> PREP REVISIONS %s\n", prior.id)
		revisedCommon := job.revisedComboActuallyUsed(prior.parts, &prior.combo, printer)
		for i := range prior.parts {
			draft := &prior.parts[i]
			param := &job.params[i]

			// TODO refactor common parts with prepareRevisionsForSim better
			printer.Printf("== %s\n", param.Label)
			printer.Println("DRAFT")
			draft.Report(printer)

			param.seenInSolutions.Add(&draft.fullSet)
			downstream <- bestOption{i, &draft.fullSet}

			revised := job.makeRevised(param, &revisedCommon, trackProgress, printer)
			for _, newOutput := range revised {
				param.seenInSolutions.Add(&newOutput.fullSet)
				downstream <- bestOption{i, &draft.fullSet}
			}
		}
		printer.Println0()
		printer.Println0()
		job.printer.AppendOther(printer)
	})

	bestList := []util_rank.BestCollector1[items.FullItemSet]{}
	for range job.params {
		bestList = append(bestList, util_rank.BestCollector1[items.FullItemSet]{})
	}
	for option := range optionChannel {
		rate := job.params[option.specIndex].Model.CalcRatingFull(option.set)
		bestList[option.specIndex].Offer(option.set, rate)
	}
	for i, best := range bestList {
		set := best.GetBestOrPanic()
		job.params[i].seenInSolutions.Add1000(&set)
	}
}

func (job *MultiSetJob) cullingReport() {
	for paramIndex := range job.params {
		job.params[paramIndex].cullingReport()
		job.params[paramIndex].cullingReportBags()
	}
}

func (param *multiSetParamInternal) cullingReport() {
	type extraInfoStruct struct {
		itemId items.ItemId
		count  uint32
	}

	added := make(map[items.ItemId]bool)

	extraInfo := make([]extraInfoStruct, 0, len(param.extraItems))
	for _, itemId := range param.extraItems {
		seenCount := param.seenInSolutions.content[itemId]
		info := extraInfoStruct{itemId: itemId, count: seenCount}
		extraInfo = append(extraInfo, info)
		added[itemId] = true
	}

	for item := range param.exactEquippedGear.AllItemSeq() {
		itemId := item.ItemId()
		if !added[itemId] {
			seenCount := param.seenInSolutions.content[itemId]
			info := extraInfoStruct{itemId: itemId, count: seenCount}
			extraInfo = append(extraInfo, info)
			added[itemId] = true
		}
	}

	slices.SortFunc(extraInfo, func(a, b extraInfoStruct) int {
		return cmp.Or(cmp.Compare(a.count, b.count), cmp.Compare(a.itemId, b.itemId))
	})

	param.job.printer.Printf("EXTRAS USED %s\n", param.Label)
	for _, info := range extraInfo {
		if slices.Contains(param.blockedItems, info.itemId) {
			param.job.printer.Printf("%5d BLOCKED!\n", info.itemId)
			continue
		}
		item, itemFound := param.itemOptions.FindItemIdFirstOptional(info.itemId)
		if itemFound {
			if info.count == 0 {
				param.job.printer.Printf("%5d 0 NONE // %s; %s\n", info.itemId, item.Slot.Name(), item.BaseName)
			} else {
				param.job.printer.Printf("%5d %6d // %s; %s\n", info.itemId, info.count, item.Slot.Name(), item.BaseName)
			}
		} else {
			param.job.printer.Printf("%5d %d MISSING IN OPTIONS\n", info.itemId, info.count)
		}
	}
	param.job.printer.Println0()
}

func (param *multiSetParamInternal) cullingReportBags() {
	for _, itemId := range param.addedFromBags {
		seenCount := param.seenInSolutions.content[itemId]
		if seenCount > 0 {
			item, itemFound := param.itemOptions.FindItemIdFirstOptional(itemId)
			if itemFound {
				param.job.printer.Printf("BAGS SUGGESTION %d %d %s; %s !!\n", itemId, seenCount, item.Slot.Name(), item.BaseName)
			} else {
				param.job.printer.Printf("BAGS SUGGESTION %d %d BUT missing options?!?!?!?!\n", itemId, seenCount)
			}
		}
	}
}
