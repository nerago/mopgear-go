package indexed

import (
	"math/big"
	"paladin_gearing_go/model"
	. "paladin_gearing_go/types/items"
	"paladin_gearing_go/util"
	"time"
)

func mainLoop_singleThread_big(itemOptions *SolvableOptionsMap, max, skip *big.Int, model *model.Model, peekFunc func(*SolvableItemSet)) SolvableItemSet {
	slotSizes := slotSizesBig(itemOptions)

	var index big.Int
	index.Set(big.NewInt(0))
	best := util.BestCollector1[SolvableItemSet]{}

	go trackProgressBig(&index, max)

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
	}

	return best.GetBest()
}

func trackProgressBig(index, max *big.Int) {
	startTime := time.Now()
	for {
		time.Sleep(time.Second * 5)

		var ratio big.Rat
		ratio.SetFrac(index, max)
		percent, _ := ratio.Float64()

		util.PrintProgressBig(startTime, percent, index)
	}
}
