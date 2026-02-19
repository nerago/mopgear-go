package indexed

import (
	"paladin_gearing_go/model"
	. "paladin_gearing_go/types/items"
	"paladin_gearing_go/util"
	"time"
)

func mainLoop_singleThread_int(itemOptions *SolvableOptionsMap, max, skip uint64, model *model.Model) SolvableItemSet {
	var index uint64 = 0
	best := util.BestCollector1[SolvableItemSet]{}

	go trackProgressInt(index, max)

	for index < max {
		set := makeSetInt(itemOptions, index)
		if model.CheckSet(&set) {
			rating := model.CalcRatingSolve(&set)
			best.Offer(&set, rating)
		}
		index += skip
	}

	return best.GetBest()
}

func trackProgressInt(index, max uint64) {
	startTime := time.Now()
	for {
		time.Sleep(time.Second * 5)

		percent := float64(index) / float64(max)

		printProgressInt(startTime, percent, index)
	}
}
