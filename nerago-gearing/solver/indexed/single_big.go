package indexed

import (
	"math/big"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/util"
)

func mainLoop_singleThread_big(itemOptions *items.SolvableOptionsMap, max, skip *big.Int, trackProgress *util.TrackProgress, model *model.Model, peekFunc func(*items.SolvableItemSet)) util.Optional[items.SolvableItemSet] {
	slotSizes := slotSizesBig(itemOptions)

	index := big.NewInt(0)
	best := util.BestCollector1[items.SolvableItemSet]{}

	trackProgress.RunFromBigInt(index, max)

	for index.Cmp(max) < 0 {
		set := makeSetBig(itemOptions, &slotSizes, index)
		if peekFunc != nil {
			peekFunc(&set)
		}
		if model.CheckSet(&set) {
			rating := model.CalcRatingSolve(&set)
			best.Offer(&set, rating)
		}
		index.Add(index, skip)
	}

	return best.GetBestOptional()
}
