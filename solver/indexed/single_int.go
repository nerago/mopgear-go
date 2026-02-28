package indexed

import (
	"context"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/util"
)

func mainLoop_singleThread_int(itemOptions *items.SolvableOptionsMap, max, skip uint64, model *model.Model, peekFunc func(*items.SolvableItemSet)) util.Optional[items.SolvableItemSet] {
	var index uint64 = 0
	best := util.BestCollector1[items.SolvableItemSet]{}

	ctx, cancel := context.WithCancel(context.Background())
	go util.TrackProgressInt(ctx, &index, max)
	defer cancel()

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
	}

	return best.GetBestOptional()
}
