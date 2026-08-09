package bonus_set

import (
	"iter"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/weightfind/weight_types"
)

const c_maxItemId = 300000

type SpecSetsEnable struct {
	EnabledSets []PreparedBonus
	itemToSet   [c_maxItemId]uint8
}

func (sets *SpecSetsEnable) AllItemIds() iter.Seq[items.ItemId] {
	return func(yield func(items.ItemId) bool) {
		for prep := range util_collection.ForPointer(sets.EnabledSets) {
			for _, itemId := range prep.items {
				if !yield(itemId) {
					return
				}
			}
		}
	}
}

func (sets *SpecSetsEnable) Equals(other *SpecSetsEnable) bool {
	return util_collection.EqualFunc_Pointer(sets.EnabledSets, other.EnabledSets, (*PreparedBonus).Equals)
}

func (sets *SpecSetsEnable) BonusSetIndexForItem(itemId items.ItemId) (int, bool) {
	entry := sets.itemToSet[itemId]
	if entry > 0 {
		return int(entry) - 1, true
	} else {
		return 0, false
	}
}

func (sets *SpecSetsEnable) CalcBonusSolveFlat(equipMap *items.SolvableEquipMap, ratio weight_types.SimPriorityBasic) float64 {
	return 0 // TODO
}

func (sets *SpecSetsEnable) CalcBonusSolveBySim(equipMap *items.SolvableEquipMap) *stats.SimTypeMap[float64] {
	return &stats.SimTypeMap[float64]{} // TODO
}

func (sets *SpecSetsEnable) CalcBonusFullFlat(equipMap *items.FullEquipMap, ratio weight_types.SimPriorityBasic) float64 {
	return 0 // TODO
}

func (sets *SpecSetsEnable) CalcBonusFullBySim(equipMap *items.FullEquipMap) *stats.SimTypeMap[float64] {
	return &stats.SimTypeMap[float64]{} // TODO
}

// // ########################### CalcBonus ###########################
//
//	func (sets *SetBonus) CalcBonusFull(equip *FullEquipMap) float64 {
//		numSets := len(sets.activeSets)
//		switch numSets {
//		case 0:
//			return 1
//		case 1:
//			var count uint8
//			incrementWithSetValue(&count, sets.itemToSet, equip[Equip_Head])
//			incrementWithSetValue(&count, sets.itemToSet, equip[Equip_Shoulder])
//			incrementWithSetValue(&count, sets.itemToSet, equip[Equip_Chest])
//			incrementWithSetValue(&count, sets.itemToSet, equip[Equip_Hand])
//			incrementWithSetValue(&count, sets.itemToSet, equip[Equip_Leg])
//			return sets.activeSets[0].bonuses[count]
//		default:
//			var counts [10]uint8
//			addToSpecificSet(&counts, sets.itemToSet, equip[Equip_Head])
//			addToSpecificSet(&counts, sets.itemToSet, equip[Equip_Shoulder])
//			addToSpecificSet(&counts, sets.itemToSet, equip[Equip_Chest])
//			addToSpecificSet(&counts, sets.itemToSet, equip[Equip_Hand])
//			addToSpecificSet(&counts, sets.itemToSet, equip[Equip_Leg])
//			var value float64 = 1.0
//			for index := range sets.activeSets {
//				value *= sets.activeSets[index].bonuses[counts[index+1]]
//			}
//			return value
//		}
//	}
//
//	func (sets *SetBonus) CalcBonusSolve(equip *SolvableEquipMap) float64 {
//		numSets := len(sets.activeSets)
//		switch numSets {
//		case 0:
//			return 1
//		case 1:
//			var count uint8
//			incrementWithSetValueSolve(&count, sets.itemToSet, equip[Equip_Head])
//			incrementWithSetValueSolve(&count, sets.itemToSet, equip[Equip_Shoulder])
//			incrementWithSetValueSolve(&count, sets.itemToSet, equip[Equip_Chest])
//			incrementWithSetValueSolve(&count, sets.itemToSet, equip[Equip_Hand])
//			incrementWithSetValueSolve(&count, sets.itemToSet, equip[Equip_Leg])
//			return sets.activeSets[0].bonuses[count]
//		default:
//			var counts [10]uint8
//			addToSpecificSetSolve(&counts, sets.itemToSet, equip[Equip_Head])
//			addToSpecificSetSolve(&counts, sets.itemToSet, equip[Equip_Shoulder])
//			addToSpecificSetSolve(&counts, sets.itemToSet, equip[Equip_Chest])
//			addToSpecificSetSolve(&counts, sets.itemToSet, equip[Equip_Hand])
//			addToSpecificSetSolve(&counts, sets.itemToSet, equip[Equip_Leg])
//			var value float64 = 1.0
//			for index := range sets.activeSets {
//				value *= sets.activeSets[index].bonuses[counts[index+1]]
//			}
//			return value
//		}
//	}
//
// ########################### CountInAnySet ###########################
func (sets *SpecSetsEnable) CountInAnySet(itemSet *items.FullEquipMap) uint8 {
	size := len(sets.EnabledSets)
	if size == 0 {
		return 0
	}

	var count uint8 = 0
	incrementIfInAnySet(&count, &sets.itemToSet, itemSet[items.Equip_Head])
	incrementIfInAnySet(&count, &sets.itemToSet, itemSet[items.Equip_Shoulder])
	incrementIfInAnySet(&count, &sets.itemToSet, itemSet[items.Equip_Chest])
	incrementIfInAnySet(&count, &sets.itemToSet, itemSet[items.Equip_Hand])
	incrementIfInAnySet(&count, &sets.itemToSet, itemSet[items.Equip_Leg])
	return count
}

//	func (sets *SetBonus) CountInAnySetSolve(itemSet *SolvableEquipMap) uint8 {
//		size := len(sets.activeSets)
//		if size == 0 {
//			return 0
//		}
//
//		var count uint8 = 0
//		incrementIfInAnySetSolve(&count, sets.itemToSet, itemSet[Equip_Head])
//		incrementIfInAnySetSolve(&count, sets.itemToSet, itemSet[Equip_Shoulder])
//		incrementIfInAnySetSolve(&count, sets.itemToSet, itemSet[Equip_Chest])
//		incrementIfInAnySetSolve(&count, sets.itemToSet, itemSet[Equip_Hand])
//		incrementIfInAnySetSolve(&count, sets.itemToSet, itemSet[Equip_Leg])
//		return count
//	}
//
// // ########################### incrementIfInAnySet ###########################
func incrementIfInAnySet(count *uint8, itemToSet *[c_maxItemId]uint8, item *items.FullItem) {
	if item != nil {
		entry := itemToSet[item.ItemId()]
		if entry != 0 {
			*count++
		}
	}
}

//
//func incrementIfInAnySetSolve(count *uint8, itemToSet []uint8, item *SolvableItem) {
//	if item != nil {
//		entry := itemToSet[item.ItemId()]
//		if entry != 0 {
//			*count++
//		}
//	}
//}
//
//func incrementWithSetValue(count *uint8, itemToSet []uint8, item *FullItem) {
//	if item != nil {
//		*count += itemToSet[item.ItemId()]
//	}
//}
//func incrementWithSetValueSolve(count *uint8, itemToSet []uint8, item *SolvableItem) {
//	if item != nil {
//		*count += itemToSet[item.ItemId()]
//	}
//}

func SpecSetsEnableNone() *SpecSetsEnable {
	sets := &SpecSetsEnable{}
	sets.initMap()
	return sets
}

func SpecSetsEnableNamed(names ...string) *SpecSetsEnable {
	sets := &SpecSetsEnable{}
	for _, name := range names {
		found := false
		for _, common := range g_setData {
			for _, variant := range common.variants {
				if variant.name == name {
					sets.EnabledSets = append(sets.EnabledSets, preparedBonusMake(common, variant))
					found = true
				}
			}
		}
		if !found {
			panic("set not found " + name)
		}
	}
	sets.initMap()
	return sets
}

func SpecSetsEnableForSpec(spec stats.SpecType, goal stats.OptimiseGoal) *SpecSetsEnable {
	if goal == stats.OptimiseGoal_Unknown {
		panic("please specific goal")
	}
	return SpecSetsEnableForSpec_AllowFallback(spec, goal, false)
}

func SpecSetsEnableForSpec_AllowFallback(spec stats.SpecType, goal stats.OptimiseGoal, fallback bool) *SpecSetsEnable {
	sets := &SpecSetsEnable{}

mainEntry:
	for _, common := range g_setData {
		if len(common.variants) == 1 {
			if common.variants[0].spec == spec {
				sets.EnabledSets = append(sets.EnabledSets, preparedBonusMake(common, common.variants[0]))
				continue mainEntry
			}
		} else if len(common.variants) > 1 {
			// exact spec+goal match
			for _, variant := range common.variants {
				if variant.spec == spec && variant.goal == goal {
					sets.EnabledSets = append(sets.EnabledSets, preparedBonusMake(common, variant))
					continue mainEntry
				}
			}
			// fallback entry for spec
			for _, variant := range common.variants {
				if variant.spec == spec && variant.goal == stats.OptimiseGoal_Unknown {
					sets.EnabledSets = append(sets.EnabledSets, preparedBonusMake(common, variant))
					continue mainEntry
				}
			}
			// useful in ItemBuilder, etc
			if fallback {
				for _, variant := range common.variants {
					if variant.spec == spec {
						sets.EnabledSets = append(sets.EnabledSets, preparedBonusMake(common, variant))
						continue mainEntry
					}
				}
			}
		} else {
			panic("no variants")
		}
	}

	if len(sets.EnabledSets) == 0 {
		panic("didn't find any sets")
	}
	sets.initMap()
	return sets
}

func (sets *SpecSetsEnable) initMap() {
	for index, info := range sets.EnabledSets {
		for _, itemId := range info.items {
			if sets.itemToSet[itemId] != 0 {
				panic("overlapping sets")
			}
			sets.itemToSet[itemId] = uint8(index + 1)
		}
	}
}
