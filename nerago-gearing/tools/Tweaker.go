package tools

import (
	"paladin_gearing_go/items"
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model"
	"paladin_gearing_go/util/util_rank"
)

func Tweaker_Run(initialSet *SolvableItemSet, solvableOptionsMap *SolvableOptionsMap, model *Model) SolvableItemSet {
	best := util_rank.BestCollector1[SolvableItemSet]{}
	best.Offer(initialSet, model.CalcRatingSolve(initialSet))

	possibleSet := new(SolvableItemSet)
	for slot := Equip_Iter_First; slot <= Equip_Iter_Last; slot++ {
		slotOptions := solvableOptionsMap.Get(slot)
		existing := best.BestObject.Items().Get(slot)
		if existing == nil && slotOptions != nil {
			panic("unexpected empty slot")
		} else if existing != nil && slotOptions == nil {
			panic("unexpected filled slot")
		} else if existing != nil {
			for i := range slotOptions {
				replaceItem := &slotOptions[i]

				if slot == Equip_Ring1 || slot == Equip_Ring2 || slot == Equip_Trinket1 || slot == Equip_Trinket2 {
					pairItem := best.BestObject.Items().Get(slot.PairedSlot())
					// TODO updated unique stuff
					if items.UniqueEquipViolationSolve(replaceItem, pairItem, solvableOptionsMap) {
						continue
					}
				}

				best.BestObject.ReplaceItem_Into(slot, replaceItem, possibleSet)

				if model.CheckSet(possibleSet) {
					best.OfferAndSwap(&possibleSet, model.CalcRatingSolve(possibleSet))
				}
			}
		}
	}

	return best.GetBestOrPanic()
}
