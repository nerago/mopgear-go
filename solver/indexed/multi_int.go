package indexed

import (
	"context"
	"fmt"
	"paladin_gearing_go/model"
	"paladin_gearing_go/solver/solve_util"
	. "paladin_gearing_go/types/items"
	"paladin_gearing_go/util"
	"time"
)

func mainLoop_multiThread_int(itemOptions *SolvableOptionsMap, max, skip uint64, model *model.Model, peekFunc func(*SolvableItemSet)) SolvableItemSet {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], threadCount)
	counters := [threadCount]uint64{}

	// track progress with cancel
	ctx, cancel := context.WithCancel(context.Background())
	go trackProgressIntThreaded(&counters, skip, max, ctx)
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

func trackProgressIntThreaded(threadCounters *[12]uint64, skip, max uint64, ctx context.Context) {
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
			// fmt.Printf(">%d\n", totalCount)
			index := totalCount * skip

			percent := float64(index) / float64(max)

			util.PrintProgressInt(startTime, percent, index)
		}
	}
}
