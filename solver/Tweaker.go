package solver

import (
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/types/items"
	. "paladin_gearing_go/util"
)

func Tweaker_Run(initialSet SolvableItemSet, solvableOptionsMap *SolvableOptionsMap, model *Model) SolvableItemSet {
	best := BestCollector1[SolvableItemSet]{}
	best.Offer(&initialSet, model.CalcRatingSolve(&initialSet))

	for slot, slotOptions := range solvableOptionsMap {
		existing := best.BestObject.Items[slot]
		if existing.IsEmpty() && slotOptions != nil {
			panic("unexpected empty slot")
		} else if !existing.IsEmpty() && slotOptions == nil {
			panic("unexpected filled slot")
		} else if !existing.IsEmpty() {
			for _, replace := range slotOptions {
				replaceMap := best.BestObject.Items
				replaceMap[slot] = &replace
				possibleSet := SolvableItemSet_Of(replaceMap)
				best.Offer(possibleSet, model.CalcRatingSolve(possibleSet))
			}
		}
	}

	return *best.BestObject
}
