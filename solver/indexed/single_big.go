package indexed

import (
	"math/big"
	. "paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/util"
)

func mainLoop_singleThread_big(itemOptions *SolvableOptionsMap, max, skip *big.Int, model *model.Model, peekFunc func(*SolvableItemSet)) SolvableItemSet {
	slotSizes := slotSizesBig(itemOptions)

	var index big.Int
	index.Set(big.NewInt(0))
	best := util.BestCollector1[SolvableItemSet]{}

	go util.TrackProgressBig(&index, max)

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
