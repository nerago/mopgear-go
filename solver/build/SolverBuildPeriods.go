package build

import (
	"context"
	"paladin_gearing_go/model"
	. "paladin_gearing_go/types/items"
	"paladin_gearing_go/util"
)

func SolverBuildPeriodic_Run(itemOptions *SolvableOptionsMap, model *model.Model, targetCount uint64, printer *util.PrintRecorder) SolvableItemSet {
	printer.Printf("SOLVE PERIODIC2 %d\n", targetCount)
	return evaluatePeriodic(itemOptions, model, targetCount, defaultEvaluateThreadCount, emptyPeekFunc)
}

func evaluatePeriodic(itemOptions *SolvableOptionsMap, model *model.Model, targetCount uint64, threadCount int, peekFunc func(*SolvableItemSet)) SolvableItemSet {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], threadCount)
	eachThreadCount := targetCount / uint64(threadCount)
	counters := make([]uint64, threadCount)
	slotIndexBags := makeSlotIndexBags(itemOptions)

	// track progress with cancel
	ctx, cancel := context.WithCancel(context.Background())
	go trackProgressIntThreaded(&counters, targetCount, ctx)
	defer cancel()

	for threadNum := range threadCount {
		go evaluatePeriodicWorker(resultChannel, model, eachThreadCount, itemOptions, &slotIndexBags, uint64(threadNum), &counters[threadNum], peekFunc)
	}

	// combine each thread's best result
	return util.BestCollector1_OfChannel(resultChannel, threadCount)
}

func evaluatePeriodicWorker(resultChannel chan util.BestCollector1[SolvableItemSet], model *model.Model, eachThreadCount uint64, itemOptions *SolvableOptionsMap, slotIndexBags *[16][]int, offset uint64, processedCounter *uint64, peekFunc func(*SolvableItemSet)) {
	best := util.BestCollector1[SolvableItemSet]{}

	indexes := [16]uint64{}
	for i := range indexes {
		if len(slotIndexBags[i]) > 0 {
			indexes[i] = (offset * eachThreadCount) % uint64(len(slotIndexBags[i]))
		}
	}

	for range eachThreadCount {
		// fmt.Printf("build %d\n", x)
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

func makeSetFromArrays(slotOptions *SolvableOptionsMap, slotIndexes *[16]uint64, slotIndexBags *[16][]int) SolvableItemSet {
	equip := SolvableEquipMap{}
	for slot, options := range slotOptions {
		bag := slotIndexBags[slot]
		bagSize := len(bag)
		if bagSize == 1 {
			equip[slot] = &options[0]
			// fmt.Printf("make slot=%d one\n", slot)
		} else if bagSize > 0 {
			outerIndex := slotIndexes[slot]
			innerIndex := bag[outerIndex]
			// fmt.Printf("make slot=%d outer=%d inner=%d\n", slot, outerIndex, innerIndex)
			slotIndexes[slot] = (outerIndex + 1) % uint64(bagSize)

			equip[slot] = &options[innerIndex]
		}
	}
	return SolvableItemSet_Of(equip)
}

func emptyPeekFunc(*SolvableItemSet) {
}
