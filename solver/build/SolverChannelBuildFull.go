package build

import (
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/types/common"
	. "paladin_gearing_go/types/items"
	"paladin_gearing_go/util"
)

const (
	bufferSize          = 32
	evaluateThreadCount = 8
)

func SolverChannelBuildFull_Run(itemOptions *SolvableOptionsMap, model *Model) SolvableItemSet {
	setChannel := allSetsChannel(itemOptions)
	return evaluateBestAllInput(setChannel, model)
}

func evaluateBestAllInput(setChannel <-chan SolvableItemSet, model *Model) SolvableItemSet {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], evaluateThreadCount)
	counters := [evaluateThreadCount]uint64{}

	// track progress with cancel
	// ctx, cancel := context.WithCancel(context.Background())
	// go trackProgressIntThreaded(&counters, skip, max, ctx)
	// defer cancel()

	for i := range evaluateThreadCount {
		go evaluateAllInputWorker(setChannel, resultChannel, model, &counters[i])
	}

	// combine each thread's best result
	return util.BestCollector1_OfChannel(resultChannel, evaluateThreadCount)
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
