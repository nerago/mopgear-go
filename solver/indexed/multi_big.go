package indexed

import (
	"context"
	"math/big"
	. "paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/solver/solve_util"
	"paladin_gearing_go/util"
)

func mainLoop_multiThread_big(itemOptions *SolvableOptionsMap, max, skip *big.Int, model *model.Model, peekFunc func(*SolvableItemSet)) SolvableItemSet {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], threadCount)
	counters := [threadCount]uint64{}

	// track progress with cancel
	ctx, cancel := context.WithCancel(context.Background())
	go util.TrackProgressBigThreaded(&counters, skip, max, ctx)
	defer cancel()

	// start up workers
	splits := solve_util.IndexSplitsBig(max, skip, threadCount)
	for i := range threadCount {
		go workerThreadRangeBig(itemOptions, model, splits[i], splits[i+1], skip, resultChannel, &counters[i], peekFunc)
	}

	// combine each thread's best result
	return util.BestCollector1_OfChannel(resultChannel, threadCount)
}

func workerThreadRangeBig(itemOptions *SolvableOptionsMap, model *model.Model, start, max, skip *big.Int, resultChannel chan<- util.BestCollector1[SolvableItemSet], doneCounter *uint64, peekFunc func(*SolvableItemSet)) {
	best := util.BestCollector1[SolvableItemSet]{}
	slotSizes := slotSizesBig(itemOptions)

	var index big.Int
	index.Set(start)

	// fmt.Printf("WORKER %020d-%020d\n", start, max)

	for index.Cmp(max) < 0 {
		set := makeSetBig(itemOptions, &slotSizes, &index)
		if peekFunc != nil {
			peekFunc(&set)
		}
		if model.CheckSet(&set) {
			rating := model.CalcRatingSolve(&set)
			best.Offer(&set, rating)
		}

		index.Add(&index, skip)
		(*doneCounter)++
	}

	resultChannel <- best
}
