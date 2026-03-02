package indexed

import (
	. "paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/solver/solve_util"
	"paladin_gearing_go/util"
)

func mainLoop_multiThread_int(itemOptions *SolvableOptionsMap, max, skip uint64, trackProgress *util.TrackProgress, model *model.Model, peekFunc func(*SolvableItemSet)) util.Optional[SolvableItemSet] {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], threadCount)
	counters := make([]uint64, threadCount)

	trackProgress.RunFromArray(&counters, max/skip)

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
