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
	counts := [16]uint8{}
	setCountsGeneric(&counts, &sets.itemToSet, equipMap)
	return sets.countsToFlat(&counts)
}

func (sets *SpecSetsEnable) CalcBonusFullFlat(equipMap *items.FullEquipMap, ratio weight_types.SimPriorityBasic) float64 {
	counts := [16]uint8{}
	setCountsGeneric(&counts, &sets.itemToSet, equipMap)
	return sets.countsToFlat(&counts)
}

func (sets *SpecSetsEnable) CalcBonusSolveBySim(equipMap *items.SolvableEquipMap, output *stats.SimTypeMap[float64]) {
	counts := [16]uint8{}
	setCountsGeneric(&counts, &sets.itemToSet, equipMap)
	sets.countsToSims(&counts, output)
}

func (sets *SpecSetsEnable) CalcBonusFullBySim(equipMap *items.FullEquipMap, output *stats.SimTypeMap[float64]) {
	counts := [16]uint8{}
	setCountsGeneric(&counts, &sets.itemToSet, equipMap)
	sets.countsToSims(&counts, output)
}

func setCountsGeneric[T items.IEquipMap[M], M items.IItem](counts *[16]uint8, itemToSet *[c_maxItemId]uint8, equip T) {
	addToSpecificSet(counts, itemToSet, equip.Get(items.Equip_Head))
	addToSpecificSet(counts, itemToSet, equip.Get(items.Equip_Shoulder))
	addToSpecificSet(counts, itemToSet, equip.Get(items.Equip_Chest))
	addToSpecificSet(counts, itemToSet, equip.Get(items.Equip_Hand))
	addToSpecificSet(counts, itemToSet, equip.Get(items.Equip_Leg))
}

func (sets *SpecSetsEnable) countsToFlat(counts *[16]uint8) float64 {
	value := 1.0
	for index := range sets.EnabledSets {
		value *= sets.EnabledSets[index].flatBonus[counts[index]]
	}
	return value
}

func (sets *SpecSetsEnable) countsToSims(counts *[16]uint8, output *stats.SimTypeMap[float64]) {
	for _, simType := range stats.SimTypeList {
		output.Put(simType, 1)
	}
	for index := range sets.EnabledSets {
		entry := sets.EnabledSets[index].simBonus[counts[index]]
		if entry != nil {
			multiplyInto(output, entry)
		}
	}
}

func multiplyInto(dest *stats.SimTypeMap[float64], src *stats.SimTypeMap[float64]) {
	for key, value := range src.SeqKeyValue() {
		dest.Compute(key, func(oldValue float64) float64 { return oldValue * value })
	}
}

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

func incrementIfInAnySet[M items.IItem](count *uint8, itemToSet *[c_maxItemId]uint8, item M) {
	if item != nil {
		entry := itemToSet[item.ItemId()]
		if entry != 0 {
			*count++
		}
	}
}

func addToSpecificSet[M items.IItem](counts *[16]uint8, itemToSet *[c_maxItemId]uint8, item M) {
	if item != nil {
		entry := itemToSet[item.ItemId()]
		if entry != 0 {
			counts[entry-1]++
		}
	}
}

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
