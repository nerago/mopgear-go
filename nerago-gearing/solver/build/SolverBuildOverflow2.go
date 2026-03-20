package build

import (
	. "paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/util"
)

func SolverBuildOverflow2_Run(itemOptions *SolvableOptionsMap, model *model.Model, targetCount uint64, trackProgress *util.TrackProgress, printer *util.PrintRecorder) util.Optional[SolvableItemSet] {
	printer.Printf("SOLVE OVERFLOW2 %d\n", targetCount)
	return evaluateOverflow2(itemOptions, model, targetCount, trackProgress, defaultEvaluateThreadCount, emptyPeekFunc)
}

func evaluateOverflow2(itemOptions *SolvableOptionsMap, model *model.Model, targetCount uint64, trackProgress *util.TrackProgress, threadCount int, peekFunc func(*SolvableItemSet)) util.Optional[SolvableItemSet] {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], threadCount)
	eachThreadCount := max(targetCount/uint64(threadCount), 1)
	counters := make([]uint64, threadCount)

	trackProgress.RunFromArray(&counters, targetCount)

	for threadNum := range threadCount {
		go evaluateOverflow2Worker(resultChannel, model, eachThreadCount, *itemOptions, uint64(threadNum), &counters[threadNum], peekFunc)
	}

	// combine each thread's best result
	return util.BestCollector1_OfChannel(resultChannel, threadCount)
}

func evaluateOverflow2Worker(resultChannel chan util.BestCollector1[SolvableItemSet], model *model.Model, eachThreadCount uint64, itemOptions SolvableOptionsMap, threadNum uint64, processedCounter *uint64, peekFunc func(*SolvableItemSet)) {
	best := util.BestCollector1[SolvableItemSet]{}

	indexes := [16]uint32{}
	advanceArraysInitial(&indexes, itemOptions, threadNum*eachThreadCount)

	itemSet := new(SolvableItemSet)
	best.BestObject = new(SolvableItemSet)
	for range eachThreadCount {
		makeSetFromArraysAndAdvance2(itemOptions, &indexes, itemSet)
		peekFunc(itemSet)
		if model.CheckSet(itemSet) {
			rating := model.CalcRatingSolve(itemSet)
			best.OfferAndSwap(&itemSet, rating)
		}
		*processedCounter++
	}

	resultChannel <- best
}

func makeSetFromArraysAndAdvance2(slotOptions SolvableOptionsMap, slotIndexes *[16]uint32, itemSet *SolvableItemSet) {
	itemSet.ClearTotals()

	slot := Equip_Iter_First
	for ; slot <= Equip_Iter_Last; slot++ {
		options := slotOptions[slot]
		slotSize := uint32(len(options))
		if slotSize == 1 {
			item := &options[0]
			itemSet.AddItem_Mutating(slot, item)
		} else if slotSize > 1 {
			index := slotIndexes[slot]
			item := &options[index]
			itemSet.AddItem_Mutating(slot, item)

			index++
			if index < slotSize {
				slotIndexes[slot] = index
				slot++
				break
			} else {
				slotIndexes[slot] = 0
			}
		}
	}

	for ; slot <= Equip_Iter_Last; slot++ {
		options := slotOptions[slot]
		slotSize := uint32(len(options))
		if slotSize == 1 {
			item := &options[0]
			itemSet.AddItem_Mutating(slot, item)
		} else if slotSize > 1 {
			index := slotIndexes[slot]
			item := &options[index]
			itemSet.AddItem_Mutating(slot, item)
		}
	}
}

func advanceArraysInitial(indexes *[16]uint32, slotOptions SolvableOptionsMap, skip uint64) {
	for slot := range indexes {
		slotSize := uint64(len(slotOptions[slot]))
		if slotSize > 1 {
			value := uint64(indexes[slot]) + skip
			indexes[slot] = uint32(value % slotSize)
			skip = value / slotSize
			if skip == 0 {
				return
			}
		}
	}
}
