package indexed

import (
	"context"
	. "paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/solver/solve_util"
	"paladin_gearing_go/util"
)

func mainLoop_multiThread_int(itemOptions *SolvableOptionsMap, max, skip uint64, model *model.Model, peekFunc func(*SolvableItemSet)) util.Optional[SolvableItemSet] {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], threadCount)
	counters := make([]uint64, threadCount)

	// track progress with cancel
	ctx, cancel := context.WithCancel(context.Background())
	go util.TrackProgressIntThreaded(ctx, &counters, max/skip)
	defer cancel()

	// start up workers
	splits := solve_util.IndexSplitsInt(max, skip, threadCount)
	for i := range threadCount {
		go workerThreadRangeInt(itemOptions, model, splits[i], splits[i+1], skip, resultChannel, &counters[i], peekFunc)
	}

	// combine each thread's best result
	return util.BestCollector1_OfChannel(resultChannel, threadCount)
}

func workerThreadRangeInt(itemOptions *SolvableOptionsMap, model *model.Model, start, max, skip uint64, resultChannel chan<- util.BestCollector1[SolvableItemSet], doneCounter *uint64, peekFunc func(*SolvableItemSet)) {
	best := util.BestCollector1[SolvableItemSet]{}
	index := start

	for index < max {
		set := makeSetInt(itemOptions, index)
		if peekFunc != nil {
			peekFunc(&set)
		}
		if model.CheckSet(&set) {
			rating := model.CalcRatingSolve(&set)
			best.Offer(&set, rating)
		}

		index += skip
		(*doneCounter)++
	}

	resultChannel <- best
}
