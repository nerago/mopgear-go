package items

import "paladin_gearing_go/stats"

func SolvableItemSet_RecalculateTotal(set *SolvableItemSet)

func MakeSetFromArraysAndAdvance4(slotOptions *SolvableOptionsMap, slotIndexes *[16]uint32, itemSet *SolvableItemSet)

func go_SolvableItemSet_RecalculateTotal(set *SolvableItemSet) {
	set.ClearTotals()
	for _, item := range set.items {
		stats.StatBlock_Increment_Mutating(&set.total, &item.total)
	}
}

func go2_SolvableItemSet_RecalculateTotal(set *SolvableItemSet) {
	set.total = stats.StatBlock{}
	for _, item := range set.items {
		for i := range set.total {
			set.total[i] += item.total[i]
		}
	}
}
