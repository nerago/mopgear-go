package build

import (
	"context"
	"fmt"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/types/common"
	. "paladin_gearing_go/types/items"
	"paladin_gearing_go/util"
	"slices"
	"time"
)

func SolverChannelBuildPeriodic_Run(itemOptions *SolvableOptionsMap, model *Model, targetCount uint64) SolvableItemSet {
	fmt.Printf("SOLVE PERIODIC %d\n", targetCount)
	setChannel := periodicSetsChannel(itemOptions)
	return evaluateBestLimitedCount(setChannel, model, targetCount)
}

func periodicSetsChannel(itemOptions *SolvableOptionsMap) <-chan SolvableItemSet {
	var stageChannel <-chan SolvableItemSet
	slotIndexBags := makeSlotIndexBags(itemOptions)
	for slot := Equip_Head; slot <= Equip_Offhand; slot++ {
		if len(itemOptions[slot]) > 0 {
			if stageChannel == nil {
				stageChannel = initialEndlessPeriodicChannel(itemOptions, slot, slotIndexBags[slot])
			} else {
				stageChannel = stepPeriodicChannel(itemOptions, slot, slotIndexBags[slot], stageChannel)
			}
		}
	}
	return stageChannel
}

func makeSlotIndexBags(itemOptions *SolvableOptionsMap) [16][]int {
	result := [16][]int{}
	usedPeriods := make([]int, 0, 16)
	for slot, slotOptions := range itemOptions {
		slotSize := len(slotOptions)
		if slotSize > 0 {
			var slotBag []int
			if !slices.Contains(usedPeriods, slotSize) {
				slotBag = makeSlotBagBasic(slotSize)
				usedPeriods = append(usedPeriods, slotSize)
				fmt.Printf("period %d = %d basic\n", slot, slotSize)
			} else {
				period := choosePeriod(slotSize, &usedPeriods)
				slotBag = makeSlotBagCycling(slotSize, period)
				usedPeriods = append(usedPeriods, period)
				fmt.Printf("period %d = %d cycle period %d\n", slot, slotSize, period)
			}

			result[slot] = slotBag
		}
	}
	return result
}

func choosePeriod(slotSize int, usedPeriods *[]int) int {
	// start with 3 since 2 could make extra elements too statistially significant
	// max-min factor needs to be equal or bigger than number of slots
	for multiplyFactor := 3; multiplyFactor <= 20; multiplyFactor++ {
		for tryPeriod := (slotSize * multiplyFactor) + 1; tryPeriod < slotSize*(multiplyFactor+1); tryPeriod++ {
			if !slices.Contains(*usedPeriods, tryPeriod) {
				return tryPeriod
			}
		}
	}
	panic("didn't find valid period")
}

func makeSlotBagBasic(slotSize int) []int {
	bag := make([]int, slotSize)
	for i := range slotSize {
		bag[i] = i
	}
	return bag

}
func makeSlotBagCycling(slotSize, period int) []int {
	bag := make([]int, period)
	for i := range period {
		bag[i] = i % slotSize
	}
	return bag
}

func makeNextChannel() chan SolvableItemSet {
	return make(chan SolvableItemSet, bufferSize)
	// return make(chan SolvableItemSet)
}

func initialEndlessPeriodicChannel(itemOptions *SolvableOptionsMap, slot SlotEquip, indexBag []int) <-chan SolvableItemSet {
	output := makeNextChannel()
	slotOptions := itemOptions[slot]
	go func() {
		for {
			for _, index := range indexBag {
				item := &slotOptions[index]
				set := SolvableItemSet_SingleItem(slot, item)
				output <- set
			}
		}
	}()
	return output
}

func stepPeriodicChannel(itemOptions *SolvableOptionsMap, slot SlotEquip, indexBag []int, input <-chan SolvableItemSet) <-chan SolvableItemSet {
	output := makeNextChannel()
	slotOptions := itemOptions[slot]
	go func() {
		for {
			for _, index := range indexBag {
				set := <-input

				item := &slotOptions[index]
				set.AddItem_Mutating(slot, item)
				output <- set
			}
		}
	}()
	return output
}

func evaluateBestLimitedCount(setChannel <-chan SolvableItemSet, model *Model, targetCount uint64) SolvableItemSet {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], evaluateThreadCount)
	eachThreadCount := targetCount / evaluateThreadCount
	counters := [evaluateThreadCount]uint64{}

	// track progress with cancel
	ctx, cancel := context.WithCancel(context.Background())
	go trackProgressIntThreaded(&counters, targetCount, ctx)
	defer cancel()

	for i := range evaluateThreadCount {
		go evaluateWorkerLimitedCount(setChannel, resultChannel, model, eachThreadCount, &counters[i])
	}

	// combine each thread's best result
	return util.BestCollector1_OfChannel(resultChannel, evaluateThreadCount)
}

func evaluateWorkerLimitedCount(setChannel <-chan SolvableItemSet, resultChannel chan util.BestCollector1[SolvableItemSet], model *Model, eachThreadCount uint64, doneCounter *uint64) {
	best := util.BestCollector1[SolvableItemSet]{}
	for range eachThreadCount {
		itemSet := <-setChannel
		if model.CheckSet(&itemSet) {
			rating := model.CalcRatingSolve(&itemSet)
			best.Offer(&itemSet, rating)
		}
		(*doneCounter)++
	}
	resultChannel <- best
}

func trackProgressIntThreaded(threadCounters *[evaluateThreadCount]uint64, targetCount uint64, ctx context.Context) {
	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 5):
			var totalCount uint64 = 0
			for _, value := range threadCounters {
				totalCount += value
			}
			percent := float64(totalCount) / float64(targetCount)
			util.PrintProgressInt(startTime, percent, totalCount)
		}
	}
}
