package solver

import (
	"fmt"
	"math/big"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/types/items"
	. "paladin_gearing_go/util"
	"time"
)

var int_one = big.NewInt(1)

func SolverIndexed_RunSkipping(itemOptions *SolvableOptionsMap, model *Model) SolvableItemSet {
	max := itemOptions.TotalCombinationCount()
	targetCombination := big.NewInt(10000000)

	skip := big.NewInt(0)
	skip.Div(max, targetCombination)
	skip = nextPrime(skip)

	fmt.Printf("SOLVE SKIP %d %d %d\n", max, targetCombination, skip)

	return mainLoop(itemOptions, max, skip, model)
}

func nextPrime(skip *big.Int) *big.Int {
	if skip.Cmp(int_one) <= 0 {
		return int_one
	}

	for !skip.ProbablyPrime(100) {
		skip.Add(skip, int_one)
	}
	return skip
}

func SolverIndexed_RunFull(itemOptions *SolvableOptionsMap, model *Model) SolvableItemSet {
	max := itemOptions.TotalCombinationCount()
	fmt.Printf("SOLVE FULL %d\n", max)
	return mainLoop(itemOptions, max, int_one, model)
}

func mainLoop(itemOptions *SolvableOptionsMap, max, skip *big.Int, model *Model) SolvableItemSet {
	// if max.IsUint64() && skip.IsUint64() {
	// 	return mainLoop_multiThread_int(itemOptions, max.Uint64(), skip.Uint64(), model)
	// } else {
	// 	return mainLoop_multiThread_big(itemOptions, max, skip, model)
	// }

	if max.IsUint64() && skip.IsUint64() {
		return mainLoop_singleThread_int(itemOptions, max.Uint64(), skip.Uint64(), model)
	} else {
		return mainLoop_singleThread_big(itemOptions, max, skip, model)
	}
}

const threadCount = 12

func indexSplitsBig(max, skip *big.Int) []*big.Int {
	indexPerThread := big.NewInt(0)
	indexPerThread.Div(max, skip)
	indexPerThread.Div(indexPerThread, big.NewInt(threadCount))
	indexPerThread.Mul(indexPerThread, skip)

	splitArray := make([]*big.Int, 0)
	start := big.NewInt(0)
	for range threadCount {
		splitArray = append(splitArray, start)
		start = big.NewInt(0).Add(start, indexPerThread)
	}
	splitArray = append(splitArray, max)

	return splitArray
}

func indexSplitsInt(max, skip uint64) []uint64 {
	indexPerThread := max / skip
	indexPerThread /= threadCount
	indexPerThread *= skip

	splitArray := make([]uint64, 0)
	var start uint64 = 0
	for range threadCount {
		splitArray = append(splitArray, start)
		start += indexPerThread
	}
	splitArray = append(splitArray, max)

	return splitArray
}

func mainLoop_multiThread_big(itemOptions *SolvableOptionsMap, max, skip *big.Int, model *Model) SolvableItemSet {
	resultChannel := make(chan BestCollector1[SolvableItemSet], threadCount)

	// thread to track progress
	counters := [threadCount]uint64{}
	go trackProgressBigThreaded(&counters, skip, max)

	// start up workers
	splits := indexSplitsBig(max, skip)
	for i := range threadCount {
		go workerThreadRangeBig(itemOptions, model, splits[i], splits[i+1], skip, resultChannel, &counters[i])
	}

	// combine each thread's best result
	best := BestCollector1[SolvableItemSet]{}
	for range threadCount {
		threadResult := <-resultChannel
		best.CombineOther(threadResult)
	}
	return best.GetBest()
}

func mainLoop_multiThread_int(itemOptions *SolvableOptionsMap, max, skip uint64, model *Model) SolvableItemSet {
	resultChannel := make(chan BestCollector1[SolvableItemSet], threadCount)

	// thread to track progress
	counters := [threadCount]uint64{}
	go trackProgressIntThreaded(&counters, skip, max)

	// start up workers
	splits := indexSplitsInt(max, skip)
	for i := range threadCount {
		go workerThreadRangeInt(itemOptions, model, splits[i], splits[i+1], skip, resultChannel, &counters[i])
	}

	// combine each thread's best result
	best := BestCollector1[SolvableItemSet]{}
	for range threadCount {
		threadResult := <-resultChannel
		best.CombineOther(threadResult)
	}
	return best.GetBest()
}

func workerThreadRangeBig(itemOptions *SolvableOptionsMap, model *Model, start, max, skip *big.Int,
	resultChannel chan BestCollector1[SolvableItemSet], counter *uint64) {
	best := BestCollector1[SolvableItemSet]{}
	slotSizes := slotSizesBig(itemOptions)

	var index big.Int
	index.Set(start)

	fmt.Printf("WORKER %020d-%020d\n", start, max)

	for index.Cmp(max) < 0 {
		set := makeSetBig(itemOptions, &slotSizes, &index)
		if model.CheckSet(&set) {
			rating := model.CalcRatingSolve(&set)
			best.Offer(&set, rating)
		}

		index.Add(&index, skip)
		(*counter)++
	}

	resultChannel <- best
}

func workerThreadRangeInt(itemOptions *SolvableOptionsMap, model *Model, start, max, skip uint64,
	resultChannel chan BestCollector1[SolvableItemSet], counter *uint64) {
	best := BestCollector1[SolvableItemSet]{}

	index := start

	fmt.Printf("WORKER %020d-%020d\n", start, max)

	for index < max {
		set := makeSetInt(itemOptions, index)
		if model.CheckSet(&set) {
			rating := model.CalcRatingSolve(&set)
			best.Offer(&set, rating)
		}

		index += skip
		(*counter)++
	}

	resultChannel <- best
}

func workerThread(itemOptions *SolvableOptionsMap, model *Model, indexChannel <-chan *big.Int, resultChannel chan<- BestCollector1[SolvableItemSet]) {
	best := BestCollector1[SolvableItemSet]{}
	slotSizes := slotSizesBig(itemOptions)

	for index := range indexChannel {
		set := makeSetBig(itemOptions, &slotSizes, index)
		if model.CheckSet(&set) {
			rating := model.CalcRatingSolve(&set)
			best.Offer(&set, rating)
		}
	}

	resultChannel <- best
}

func mainLoop_singleThread_int(itemOptions *SolvableOptionsMap, max, skip uint64, model *Model) SolvableItemSet {
	var index uint64 = 0
	best := BestCollector1[SolvableItemSet]{}

	go trackProgressInt(index, max)

	for index < max {
		set := makeSetInt(itemOptions, index)
		if model.CheckSet(&set) {
			rating := model.CalcRatingSolve(&set)
			best.Offer(&set, rating)
		}
		index += skip
	}

	return best.GetBest()
}

func mainLoop_singleThread_big(itemOptions *SolvableOptionsMap, max, skip *big.Int, model *Model) SolvableItemSet {
	slotSizes := slotSizesBig(itemOptions)

	var index big.Int
	index.Set(big.NewInt(0))
	best := BestCollector1[SolvableItemSet]{}

	go trackProgressBig(&index, max)

	for index.Cmp(max) < 0 {
		set := makeSetBig(itemOptions, &slotSizes, &index)
		if model.CheckSet(&set) {
			rating := model.CalcRatingSolve(&set)
			best.Offer(&set, rating)
		}
		index.Add(&index, skip)
	}

	return best.GetBest()
}

func trackProgressInt(index, max uint64) {
	startTime := time.Now()
	for {
		time.Sleep(time.Second * 5)

		percent := float64(index) / float64(max)

		if percent > 0 {
			timeTaken := time.Since(startTime)
			totalEstimate := time.Duration(float64(timeTaken) / percent)
			estimateRemain := totalEstimate - timeTaken
			fmt.Printf("%d %.1f%% %s\n", index, percent*100, estimateRemain.String())
		}
	}
}

func trackProgressIntThreaded(threadCounters *[12]uint64, skip, max uint64) {
	startTime := time.Now()
	for {
		time.Sleep(time.Second * 5)

		var totalCount uint64 = 0
		for _, value := range threadCounters {
			totalCount += value
		}
		index := totalCount * skip

		percent := float64(index) / float64(max)

		if percent > 0 {
			timeTaken := time.Since(startTime)
			totalEstimate := time.Duration(float64(timeTaken) / percent)
			estimateRemain := totalEstimate - timeTaken
			fmt.Printf("%d %.1f%% %s\n", index, percent*100, estimateRemain.String())
		}
	}
}

func trackProgressBig(index, max *big.Int) {
	startTime := time.Now()
	for {
		time.Sleep(time.Second * 5)

		var ratio big.Rat
		ratio.SetFrac(index, max)
		percent, _ := ratio.Float64()

		if percent > 0 {
			timeTaken := time.Since(startTime)
			totalEstimate := time.Duration(float64(timeTaken) / percent)
			estimateRemain := totalEstimate - timeTaken
			fmt.Printf("%d %.1f%% %s\n", index, percent*100, estimateRemain.String())
		}
	}
}

func trackProgressBigThreaded(threadCounters *[12]uint64, skip, max *big.Int) {
	startTime := time.Now()
	for {
		time.Sleep(time.Second * 5)

		var totalCount uint64 = 0
		for _, value := range threadCounters {
			totalCount += value
		}
		totalCountBig := big.NewInt(int64(totalCount))
		index := big.NewInt(0).Mul(totalCountBig, skip)

		var ratio big.Rat
		ratio.SetFrac(index, max)
		percent, _ := ratio.Float64()

		if percent > 0 {
			timeTaken := time.Since(startTime)
			totalEstimate := time.Duration(float64(timeTaken) / percent)
			estimateRemain := totalEstimate - timeTaken
			fmt.Printf("%d %.1f%% %s\n", index, percent*100, estimateRemain.String())
		}
	}
}

func slotSizesBig(itemOptions *SolvableOptionsMap) [16]*big.Int {
	slotSizes := [16]*big.Int{}
	for i, array := range itemOptions {
		slotSizes[i] = big.NewInt(int64(len(array)))
	}
	return slotSizes
}

func makeSetBig(itemOptions *SolvableOptionsMap, slotSizes *[16]*big.Int, mainIndex *big.Int) SolvableItemSet {
	equip := SolvableEquipMap{}

	currIndex := big.NewInt(0)
	currIndex.Set(mainIndex)
	mod := big.NewInt(0)

	for slot, array := range itemOptions {
		size := slotSizes[slot]

		currIndex.DivMod(currIndex, size, mod)
		slotIndex := mod.Int64()

		equip[slot] = &array[slotIndex]
	}

	return SolvableItemSet_Of(equip)
}

func makeSetInt(itemOptions *SolvableOptionsMap, mainIndex uint64) SolvableItemSet {
	equip := SolvableEquipMap{}

	currIndex := mainIndex

	for slot, array := range itemOptions {
		size := uint64(len(array))

		slotIndex := currIndex % size
		currIndex /= size

		equip[slot] = &array[slotIndex]
	}

	return SolvableItemSet_Of(equip)
}
