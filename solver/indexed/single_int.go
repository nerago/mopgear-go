package indexed

import (
	"paladin_gearing_go/model"
	. "paladin_gearing_go/types/items"
	"paladin_gearing_go/util"
)

func mainLoop_singleThread_int(itemOptions *SolvableOptionsMap, max, skip uint64, model *model.Model, peekFunc func(*SolvableItemSet)) SolvableItemSet {
	var index uint64 = 0
	best := util.BestCollector1[SolvableItemSet]{}

	go util.TrackProgressInt(&index, max)

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

	return best.GetBest()
}
