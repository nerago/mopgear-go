package build

import (
	"math/rand"
	. "paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/util"
)

func SolverBuildRandom_Run(itemOptions *SolvableOptionsMap, model *model.Model, targetCount uint64, trackProgress *util.TrackProgress, printer *util.PrintRecorder) util.Optional[SolvableItemSet] {
	printer.Printf("SOLVE RANDOM %d\n", targetCount)
	return evaluateRandom(itemOptions, model, targetCount, trackProgress, defaultEvaluateThreadCount, emptyPeekFunc)
}

func evaluateRandom(itemOptions *SolvableOptionsMap, model *model.Model, targetCount uint64, trackProgress *util.TrackProgress, threadCount int, peekFunc func(*SolvableItemSet)) util.Optional[SolvableItemSet] {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], threadCount)
	eachThreadCount := targetCount / uint64(threadCount)
	counters := make([]uint64, threadCount)

	trackProgress.RunFromArray(&counters, targetCount)

	for threadNum := range threadCount {
		go evaluateRandomWorker(resultChannel, model, eachThreadCount, itemOptions, uint64(threadNum), &counters[threadNum], peekFunc)
	}

	// combine each thread's best result
	return util.BestCollector1_OfChannel(resultChannel, threadCount)
}

func evaluateRandomWorker(resultChannel chan util.BestCollector1[SolvableItemSet], model *model.Model, eachThreadCount uint64, itemOptions *SolvableOptionsMap, threadNum uint64, processedCounter *uint64, peekFunc func(*SolvableItemSet)) {
	best := util.BestCollector1[SolvableItemSet]{}
	rng := rand.New(rand.NewSource(int64(threadNum)))

	for range eachThreadCount {
		// fmt.Printf("build %d\n", x)
		itemSet := makeSetFromRandom(itemOptions, rng)
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

func makeSetFromRandom(slotOptions *SolvableOptionsMap, rng *rand.Rand) SolvableItemSet {
	equip := SolvableEquipMap{}
	for slot, options := range slotOptions {
		optionSize := len(options)
		if optionSize > 0 {
			index := rng.Intn(optionSize)
			equip[slot] = &options[index]
		}
	}
	return SolvableItemSet_Of(equip)
}
