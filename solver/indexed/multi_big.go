package indexed

import (
	"context"
	"math/big"
	"paladin_gearing_go/model"
	"paladin_gearing_go/solver/solve_util"
	. "paladin_gearing_go/types/items"
	"paladin_gearing_go/util"
	"time"
)

func mainLoop_multiThread_big(itemOptions *SolvableOptionsMap, max, skip *big.Int, model *model.Model, peekFunc func(*SolvableItemSet)) SolvableItemSet {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], threadCount)
	counters := [threadCount]uint64{}

	// track progress with cancel
	ctx, cancel := context.WithCancel(context.Background())
	go trackProgressBigThreaded(&counters, skip, max, ctx)
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

func trackProgressBigThreaded(threadCounters *[12]uint64, skip, max *big.Int, ctx context.Context) {
	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(time.Second * 5)
		}

		var totalCount uint64 = 0
		for _, value := range threadCounters {
			totalCount += value
		}
		totalCountBig := big.NewInt(int64(totalCount))
		index := big.NewInt(0).Mul(totalCountBig, skip)

		var ratio big.Rat
		ratio.SetFrac(index, max)
		percent, _ := ratio.Float64()

		util.PrintProgressBig(startTime, percent, index)
	}
}
