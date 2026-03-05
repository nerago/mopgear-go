package channel

import (
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model"
	"paladin_gearing_go/util"
)

const (
	bufferSize                 = 32
	defaultEvaluateThreadCount = 6
)

func SolverChannelBuildFull_Run(itemOptions *SolvableOptionsMap, model *Model, trackProgress *util.TrackProgress) util.Optional[SolvableItemSet] {
	total := itemOptions.TotalCombinationCount().Uint64()
	setChannel := allSetsChannel(itemOptions)
	return evaluateBestAllInput(setChannel, model, defaultEvaluateThreadCount, total, trackProgress)
}

func evaluateBestAllInput(setChannel <-chan SolvableItemSet, model *Model, threadCount int, totalCount uint64, trackProgress *util.TrackProgress) util.Optional[SolvableItemSet] {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], threadCount)
	counters := make([]uint64, threadCount)

	trackProgress.RunFromArray(&counters, totalCount)

	for i := range threadCount {
		go evaluateAllInputWorker(setChannel, resultChannel, model, &counters[i])
	}

	// combine each thread's best result
	return util.BestCollector1_OfChannel(resultChannel, threadCount)
}

func evaluateAllInputWorker(setChannel <-chan SolvableItemSet, resultChannel chan util.BestCollector1[SolvableItemSet], model *Model, doneCounter *uint64) {
	best := util.BestCollector1[SolvableItemSet]{}
	for itemSet := range setChannel {
		if model.CheckSet(&itemSet) {
			rating := model.CalcRatingSolve(&itemSet)
			best.Offer(&itemSet, rating)
		}
		(*doneCounter)++
	}
	resultChannel <- best
}

func allSetsChannel(itemOptions *SolvableOptionsMap) <-chan SolvableItemSet {
	var stageChannel <-chan SolvableItemSet
	for slot := Equip_Head; slot <= Equip_Offhand; slot++ {
		if len(itemOptions[slot]) > 0 {
			if stageChannel == nil {
				stageChannel = initialAllOnceChannel(itemOptions, slot)
			} else {
				stageChannel = stepAllOnceChannel(itemOptions, slot, stageChannel)
			}
		}
	}
	return stageChannel
}

func initialAllOnceChannel(itemOptions *SolvableOptionsMap, slot SlotEquip) <-chan SolvableItemSet {
	output := make(chan SolvableItemSet, bufferSize)
	slotOptions := itemOptions[slot]
	go func() {
		for _, item := range slotOptions {
			set := SolvableItemSet_SingleItem(slot, &item)
			output <- set
		}
		close(output)
	}()
	return output
}

func stepAllOnceChannel(itemOptions *SolvableOptionsMap, slot SlotEquip, input <-chan SolvableItemSet) <-chan SolvableItemSet {
	output := make(chan SolvableItemSet, bufferSize)
	slotOptions := itemOptions[slot]
	go func() {
		for baseSet := range input {
			for _, item := range slotOptions {
				newSet := baseSet.AddItem_CreateNew(slot, &item)
				output <- newSet
			}
		}
		close(output)
	}()
	return output
}
