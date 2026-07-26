package solve_build

import (
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util/util_rank"
	"slices"
)

func SolverBuildBestWorst(itemOptions *items.FullOptionsMap, model *gear_model.SpecModel) []items.FullItemSet {
	hi := items.FullEquipMap{}
	lo := items.FullEquipMap{}
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		options := itemOptions[slot]
		optionSize := len(options)
		if optionSize > 0 {
			if slot == items.Equip_Ring1 || slot == items.Equip_Trinket1 {
				best2 := chooseBest(options, model)
				worst2 := chooseWorst(options, model)
				hi[slot] = best2[0]
				hi[slot.PairedSlot()] = best2[1]
				lo[slot] = worst2[0]
				lo[slot.PairedSlot()] = worst2[1]
				slot++
			} else {
				hi[slot] = chooseBest(options, model)[0]
				lo[slot] = chooseWorst(options, model)[0]
			}
		}
	}
	return []items.FullItemSet{items.FullItemSet_FromMap(hi), items.FullItemSet_FromMap(lo)}
}

func chooseBest(options []items.FullItem, model *gear_model.SpecModel) []*items.FullItem {
	best := util_rank.HighestCollector_ForN[items.FullItem](2, (*items.FullItem).Equals)
	for _, item := range options {
		rating := model.CalcRatingFullItem(&item)
		best.Offer(&item, rating)
	}
	return slices.Collect(best.ResultsSeq())
}

func chooseWorst(options []items.FullItem, model *gear_model.SpecModel) []*items.FullItem {
	best := util_rank.LowestCollector_ForN[items.FullItem](2, (*items.FullItem).Equals)
	for _, item := range options {
		rating := model.CalcRatingFullItem(&item)
		best.Offer(&item, rating)
	}
	return slices.Collect(best.ResultsSeq())
}
