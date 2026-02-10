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

	fmt.Printf("SOLVE SKIP %d %d %d\n", max, targetCombination, skip)

	return mainLoop(itemOptions, max, skip, model)
}

func SolverIndexed_RunFull(itemOptions *SolvableOptionsMap, model *Model) SolvableItemSet {
	max := itemOptions.TotalCombinationCount()
	fmt.Printf("SOLVE FULL %d\n", max)
	return mainLoop(itemOptions, max, int_one, model)
}

func mainLoop(itemOptions *SolvableOptionsMap, max, skip *big.Int, model *Model) SolvableItemSet {
	return mainLoop_multiThread2(itemOptions, max, skip, model)
	// return mainLoop_singleThread(itemOptions, max, skip, model)
}

const threadCount = 12

func mainLoop_multiThread(itemOptions *SolvableOptionsMap, max, skip *big.Int, model *Model) SolvableItemSet {
	indexChannel := make(chan *big.Int, 128)
	resultChannel := make(chan BestCollector1[SolvableItemSet], threadCount)
	index := big.NewInt(0)

	// thread to track progress
	go trackProgress(index, max)

	// start up workers
	for range threadCount {
		go workerThread(itemOptions, model, indexChannel, resultChannel)
	}

	// generate indexes on main thread
	for index.Cmp(max) < 0 {
		indexChannel <- index

		nextIndex := big.NewInt(0)
		nextIndex.Add(index, skip)
		index = nextIndex
	}
	close(indexChannel)

	// combine each thread's best result
	best := BestCollector1[SolvableItemSet]{}
	for range threadCount {
		threadResult := <-resultChannel
		best.CombineOther(threadResult)
	}
	return best.GetBest()
}

func indexSplits(max, skip *big.Int) []*big.Int {
	// totalSteps := big.NewInt(0).Div(max, skip)
	// eachThreadSteps := big.NewInt(0).Div(totalSteps, big.NewInt(threadCount))
	// eachThreadIndexRange := big.NewInt(0).Mul(eachThreadSteps, skip)

	indexPerThread := big.NewInt(0)
	indexPerThread.Div(max, skip)
	indexPerThread.Div(indexPerThread, big.NewInt(threadCount))
	indexPerThread.Mul(indexPerThread, skip)

	// indexPerThread := big.NewInt(0)
	// indexPerThread.Div(max, big.NewInt(threadCount))

	splitArray := make([]*big.Int, 0)
	start := big.NewInt(0)
	for range threadCount {
		splitArray = append(splitArray, start)
		start = big.NewInt(0).Add(start, indexPerThread)
	}
	splitArray = append(splitArray, max)

	return splitArray
}

func mainLoop_multiThread2(itemOptions *SolvableOptionsMap, max, skip *big.Int, model *Model) SolvableItemSet {
	resultChannel := make(chan BestCollector1[SolvableItemSet], threadCount)

	// thread to track progress
	// go trackProgress(index, max)

	// start up workers
	splits := indexSplits(max, skip)
	for i := range threadCount {
		go workerThreadRange(itemOptions, model, splits[i], splits[i+1], skip, resultChannel)
	}

	// combine each thread's best result
	best := BestCollector1[SolvableItemSet]{}
	for range threadCount {
		threadResult := <-resultChannel
		best.CombineOther(threadResult)
	}
	return best.GetBest()
}

func workerThreadRange(itemOptions *SolvableOptionsMap, model *Model, start, max, skip *big.Int, resultChannel chan BestCollector1[SolvableItemSet]) {
	best := BestCollector1[SolvableItemSet]{}
	slotSizes := slotSizes(itemOptions)

	var index big.Int
	index.Set(start)

	for index.Cmp(max) < 0 {
		set := makeSet(itemOptions, &slotSizes, &index)
		if model.CheckSet(set) {
			rating := model.CalcRatingSolve(set)
			best.Offer(set, rating)
		}

		index.Add(&index, skip)
	}

	resultChannel <- best
}

func workerThread(itemOptions *SolvableOptionsMap, model *Model, indexChannel <-chan *big.Int, resultChannel chan<- BestCollector1[SolvableItemSet]) {
	best := BestCollector1[SolvableItemSet]{}
	slotSizes := slotSizes(itemOptions)

	for index := range indexChannel {
		set := makeSet(itemOptions, &slotSizes, index)
		if model.CheckSet(set) {
			rating := model.CalcRatingSolve(set)
			best.Offer(set, rating)
		}
	}

	resultChannel <- best
}

func mainLoop_singleThread(itemOptions *SolvableOptionsMap, max, skip *big.Int, model *Model) SolvableItemSet {
	slotSizes := slotSizes(itemOptions)
	index := big.NewInt(0)
	best := BestCollector1[SolvableItemSet]{}

	go trackProgress(index, max)

	for index.Cmp(max) < 0 {
		set := makeSet(itemOptions, &slotSizes, index)
		if model.CheckSet(set) {
			rating := model.CalcRatingSolve(set)
			best.Offer(set, rating)
		}
		index.Add(index, skip)
	}

	return best.GetBest()
}

func trackProgress(index, max *big.Int) {
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

func slotSizes(itemOptions *SolvableOptionsMap) [16]*big.Int {
	slotSizes := [16]*big.Int{}
	for i, array := range itemOptions {
		slotSizes[i] = big.NewInt(int64(len(array)))
	}
	return slotSizes
}

func makeSet(itemOptions *SolvableOptionsMap, slotSizes *[16]*big.Int, mainIndex *big.Int) *SolvableItemSet {
	equip := SolvableEquipMap{}

	var div, mod big.Int
	for slot, array := range itemOptions {
		size := slotSizes[slot]
		div.DivMod(mainIndex, size, &mod)

		slotIndex := mod.Int64()
		choice := &array[slotIndex]
		equip[slot] = choice

		mainIndex = &div
	}

	return SolvableItemSet_Of(equip)
}
