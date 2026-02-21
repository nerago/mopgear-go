package build

import (
	"fmt"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/types/common"
	. "paladin_gearing_go/types/items"
	"paladin_gearing_go/util"
)

func SolverChannelBuildPrime_Run(itemOptions *SolvableOptionsMap, model *Model, targetCount uint64) SolvableItemSet {
	fmt.Printf("SOLVE PRIME %d\n", targetCount)
	setChannel := primeSetsChannel(itemOptions)
	return evaluateBestLimitedCount(setChannel, model, targetCount)
}

func evaluateBestLimitedCount(setChannel <-chan SolvableItemSet, model *Model, targetCount uint64) SolvableItemSet {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], evaluateThreadCount)
	eachThreadCount := targetCount / evaluateThreadCount
	counters := [evaluateThreadCount]uint64{}

	// track progress with cancel
	// ctx, cancel := context.WithCancel(context.Background())
	// go trackProgressIntThreaded(&counters, skip, max, ctx)
	// defer cancel()

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

func primeSetsChannel(itemOptions *SolvableOptionsMap) <-chan SolvableItemSet {
	var stageChannel <-chan SolvableItemSet
	// maxSlot := maxSlotOptionSize(itemOptions)
	// primeIncrements := util.SmallPrimes(maxSlot)
	primeIncrements := []int{1, 1, 1, 3}
	for slot := Equip_Head; slot <= Equip_Offhand; slot++ {
		if len(itemOptions[slot]) > 0 {
			if stageChannel == nil {
				stageChannel = initialEndlessPrimeChannel(itemOptions, slot, primeIncrements[slot])
			} else {
				stageChannel = stepPrimeChannel(itemOptions, slot, primeIncrements[slot], stageChannel)
			}
		}
	}
	return stageChannel
}

func maxSlotOptionSize(itemOptions *SolvableOptionsMap) int {
	max := 0
	for _, slotOptions := range itemOptions {
		if len(slotOptions) > max {
			max = len(slotOptions)
		}
	}
	return max
}

func initialEndlessPrimeChannel(itemOptions *SolvableOptionsMap, slot SlotEquip, indexIncrement int) <-chan SolvableItemSet {
	output := make(chan SolvableItemSet, bufferSize)
	slotOptions := itemOptions[slot]
	optionsSize := len(slotOptions)
	go func() {
		for {
			index := 0
			item := &slotOptions[index]

			set := SolvableItemSet_SingleItem(slot, item)
			output <- set

			index = (index + indexIncrement) % optionsSize
		}
	}()
	return output
}

func stepPrimeChannel(itemOptions *SolvableOptionsMap, slot SlotEquip, indexIncrement int, input <-chan SolvableItemSet) <-chan SolvableItemSet {
	output := make(chan SolvableItemSet, bufferSize)
	slotOptions := itemOptions[slot]
	optionsSize := len(slotOptions)
	go func() {
		index := 0
		for baseSet := range input {
			item := &slotOptions[index]

			newSet := baseSet.AddItem_CreateNew(slot, item)
			output <- newSet

			index = (index + indexIncrement) % optionsSize
		}
	}()
	return output
}

// func validatePrimeLooping() {
// 	for arraySize := 1; arraySize <= 300; arraySize++ {
// 		primes := util.SmallPrimes(arraySize)
// 		for _, indexIncrement := range primes {
// 			checkPrimeSize(arraySize, indexIncrement)
// 		}
// 	}
// }

// func checkPrimeSize(arraySize, indexIncrement int) {
// 	array := make([]bool, arraySize)
// 	index := 0
// 	for range 2000 {
// 		array[index] = true
// 		index = (index + indexIncrement) % arraySize
// 	}
// 	for _, val := range array {
// 		if !val {
// 			panic("missing true in " + strconv.Itoa(arraySize) + " " + strconv.Itoa(indexIncrement))
// 		}
// 	}
// 	fmt.Printf("ok %d %d\n", arraySize, indexIncrement)
// }
