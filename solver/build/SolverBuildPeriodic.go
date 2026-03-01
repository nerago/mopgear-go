package build

import (
	"context"
	. "paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/solver/channel"
	"paladin_gearing_go/util"
)

const (
	defaultEvaluateThreadCount = 12
)

func SolverBuildPeriodic_Run(itemOptions *SolvableOptionsMap, model *model.Model, targetCount uint64, printer *util.PrintRecorder) util.Optional[SolvableItemSet] {
	printer.Printf("SOLVE PERIODIC2 %d\n", targetCount)
	return evaluatePeriodic(itemOptions, model, targetCount, defaultEvaluateThreadCount, emptyPeekFunc)
}

func evaluatePeriodic(itemOptions *SolvableOptionsMap, model *model.Model, targetCount uint64, threadCount int, peekFunc func(*SolvableItemSet)) util.Optional[SolvableItemSet] {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], threadCount)
	eachThreadCount := max(targetCount/uint64(threadCount), 1)
	counters := make([]uint64, threadCount)
	slotIndexBags := channel.MakeSlotIndexBags(itemOptions)

	// track progress with cancel
	ctx, cancel := context.WithCancel(context.Background())
	go util.TrackProgressIntThreaded(ctx, &counters, targetCount)
	defer cancel()

	for threadNum := range threadCount {
		go evaluatePeriodicWorker(resultChannel, model, eachThreadCount, itemOptions, &slotIndexBags, uint64(threadNum), &counters[threadNum], peekFunc)
	}

	// combine each thread's best result
	return util.BestCollector1_OfChannel(resultChannel, threadCount)
}

func evaluatePeriodicWorker(resultChannel chan util.BestCollector1[SolvableItemSet], model *model.Model, eachThreadCount uint64, itemOptions *SolvableOptionsMap, slotIndexBags *[16][]int, threadNum uint64, processedCounter *uint64, peekFunc func(*SolvableItemSet)) {
	best := util.BestCollector1[SolvableItemSet]{}

	// by starting with last index we try to avoid a first round where they move in sync for most of the bag before diverging
	// then index offset by thread is trying to calculate where previous worker would have left off
	indexes := [16]uint64{}
	for i := range indexes {
		slotLen := uint64(len(slotIndexBags[i]))
		if slotLen > 0 {
			indexes[i] = (slotLen - 1 + (threadNum * eachThreadCount)) % slotLen
			// indexes[i] = ((offset * eachThreadCount)) % slotLen
			// indexes[i] = 0
		}
	}

	for range eachThreadCount {
		// fmt.Printf("build %d\n", x)
		itemSet := makeSetFromArraysBagged(itemOptions, &indexes, slotIndexBags)
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

func makeSetFromArraysBagged(slotOptions *SolvableOptionsMap, slotIndexes *[16]uint64, slotIndexBags *[16][]int) SolvableItemSet {
	equip := SolvableEquipMap{}
	for slot, options := range slotOptions {
		bag := slotIndexBags[slot]
		bagSize := len(bag)
		if bagSize == 1 {
			equip[slot] = &options[0]
		} else if bagSize > 1 {
			outerIndex := slotIndexes[slot]
			innerIndex := bag[outerIndex]
			slotIndexes[slot] = (outerIndex + 1) % uint64(bagSize)

			equip[slot] = &options[innerIndex]
		}
	}
	return SolvableItemSet_Of(equip)
}

func emptyPeekFunc(*SolvableItemSet) {
}
