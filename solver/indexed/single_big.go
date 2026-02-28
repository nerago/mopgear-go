package indexed

import (
	"context"
	"math/big"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/util"
)

func mainLoop_singleThread_big(itemOptions *items.SolvableOptionsMap, max, skip *big.Int, model *model.Model, peekFunc func(*items.SolvableItemSet)) util.Optional[items.SolvableItemSet] {
	slotSizes := slotSizesBig(itemOptions)

	var index big.Int
	index.Set(big.NewInt(0))
	best := util.BestCollector1[items.SolvableItemSet]{}

	ctx, cancel := context.WithCancel(context.Background())
	go util.TrackProgressBig(ctx, &index, max)
	defer cancel()

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

	return best.GetBestOptional()
}
