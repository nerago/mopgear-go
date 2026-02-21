package build

import (
	"context"
	"fmt"
	"paladin_gearing_go/model"
	. "paladin_gearing_go/types/items"
	"paladin_gearing_go/util"
)

func SolverBuildPeriodic_Run(itemOptions *SolvableOptionsMap, model *model.Model, targetCount uint64) SolvableItemSet {
	fmt.Printf("SOLVE PERIODIC2 %d\n", targetCount)
	return evaluatePeriodic(itemOptions, model, targetCount, emptyPeekFunc)
}

func evaluatePeriodic(itemOptions *SolvableOptionsMap, model *model.Model, targetCount uint64, peekFunc func(*SolvableItemSet)) SolvableItemSet {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], evaluateThreadCount)
	eachThreadCount := targetCount / evaluateThreadCount
	counters := [evaluateThreadCount]uint64{}
	slotIndexBags := makeSlotIndexBags(itemOptions)

	// track progress with cancel
	ctx, cancel := context.WithCancel(context.Background())
	go trackProgressIntThreaded(&counters, targetCount, ctx)
	defer cancel()

	for i := range evaluateThreadCount {
		go evaluatePeriodicWorker(resultChannel, model, eachThreadCount, itemOptions, &slotIndexBags, i, &counters[i], peekFunc)
	}

	// combine each thread's best result
	return util.BestCollector1_OfChannel(resultChannel, evaluateThreadCount)
}

func evaluatePeriodicWorker(resultChannel chan util.BestCollector1[SolvableItemSet], model *model.Model, eachThreadCount uint64, itemOptions *SolvableOptionsMap, slotIndexBags *[16][]int, offset int, processedCounter *uint64, peekFunc func(*SolvableItemSet)) {
	best := util.BestCollector1[SolvableItemSet]{}

	indexes := [16]int{}
	for i := range indexes {
		if len(slotIndexBags[i]) > 0 {
			indexes[i] = offset % len(slotIndexBags[i])
		}
	}

	for range eachThreadCount {
		itemSet := makeSetFromArrays(itemOptions, &indexes, slotIndexBags)
		if peekFunc != nil {
			peekFunc(&itemSet)
		}
		if model.CheckSet(&itemSet) {
			rating := model.CalcRatingSolve(&itemSet)
			best.Offer(&itemSet, rating)
		}
		(*processedCounter)++
	}

	resultChannel <- best
}

func makeSetFromArrays(slotOptions *SolvableOptionsMap, slotIndexes *[16]int, slotIndexBags *[16][]int) SolvableItemSet {
	equip := SolvableEquipMap{}
	for slot, options := range slotOptions {
		bag := slotIndexBags[slot]
		bagSize := len(bag)
		if bagSize == 1 {
			equip[slot] = &options[0]
		} else if bagSize > 0 {
			outerIndex := slotIndexes[slot]
			innerIndex := bag[outerIndex]
			slotIndexes[slot] = (outerIndex + 1) % bagSize

			equip[slot] = &options[innerIndex]
		}
	}
	return SolvableItemSet_Of(equip)
}

func emptyPeekFunc(*SolvableItemSet) {
}
