package build

import (
	. "paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/util"
)

func SolverBuildFull_Run(itemOptions *SolvableOptionsMap, model *model.Model, trackProgress *util.TrackProgress, printer *util.PrintRecorder) util.Optional[SolvableItemSet] {
	return evaluateFull(itemOptions, model, trackProgress, printer, emptyPeekFunc)
}

func evaluateFull(itemOptions *SolvableOptionsMap, model *model.Model, trackProgress *util.TrackProgress, printer *util.PrintRecorder, peekFunc func(*SolvableItemSet)) util.Optional[SolvableItemSet] {
	expectedCount := itemOptions.TotalCombinationCountAsInt()
	printer.Printf("SOLVE FULL %d\n", expectedCount)

	splitSlot := chooseSplitSlot(itemOptions)
	threadCount := max(len(itemOptions[splitSlot]), 1)

	counters := make([]uint64, threadCount)
	trackProgress.RunFromArray(&counters, expectedCount)

	resultChannel := make(chan util.BestCollector1[SolvableItemSet])
	for threadNum := range threadCount {
		go evaluateFullWorker(resultChannel, model, *itemOptions, splitSlot, threadNum, &counters[threadNum], peekFunc)
	}

	// combine each thread's best result
	return util.BestCollector1_OfChannel(resultChannel, threadCount)
}

func evaluateFullWorker(resultChannel chan util.BestCollector1[SolvableItemSet], model *model.Model, optionsMap SolvableOptionsMap, splitSlot SlotEquip, threadNum int, processedCounter *uint64, peekFunc func(*SolvableItemSet)) {
	best := util.BestCollector1[SolvableItemSet]{}

	initialSet := SolvableItemSet_SingleItem(splitSlot, &optionsMap[splitSlot][threadNum])
	makeSetsRecur(Equip_Iter_First, initialSet, optionsMap, splitSlot, model, &best, processedCounter, peekFunc)

	resultChannel <- best
}

func makeSetsRecur(checkSlot SlotEquip, baseSet SolvableItemSet, optionsMap SolvableOptionsMap, splitSlot SlotEquip, model *model.Model, best *util.BestCollector1[SolvableItemSet], processedCounter *uint64, peekFunc func(*SolvableItemSet)) {
	for ; checkSlot <= Equip_Iter_Last; checkSlot++ {
		if checkSlot == splitSlot {
			continue // skip
		}

		slotOpts := optionsMap[checkSlot]
		slotSize := len(slotOpts)
		if slotSize > 0 {
			for i := range slotSize {
				nextSet := baseSet.AddItem_CreateNew(checkSlot, &slotOpts[i])
				makeSetsRecur(checkSlot+1, nextSet, optionsMap, splitSlot, model, best, processedCounter, peekFunc)
			}
			return
		}
	}

	if peekFunc != nil {
		peekFunc(&baseSet)
	}

	if model.CheckSet(&baseSet) {
		value := model.CalcRatingSolve(&baseSet)
		best.Offer(&baseSet, value)
	}
	*processedCounter++
}

func chooseSplitSlot(itemOptions *SolvableOptionsMap) SlotEquip {
	biggestSlot := util.BestCollector1[SlotEquip]{}
	for slot := Equip_Iter_First; slot <= Equip_Iter_Last; slot++ {
		biggestSlot.Offer(&slot, uint64(len(itemOptions[slot])))
	}
	return biggestSlot.GetBestOrPanic()
}
