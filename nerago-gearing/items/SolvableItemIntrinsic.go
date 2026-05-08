package items

import "paladin_gearing_go/stats"

func SolvableItemSet_RecalculateTotal(set *SolvableItemSet)

func go_SolvableItemSet_RecalculateTotal(set *SolvableItemSet) {
	set.ClearTotals()
	for _, item := range set.items {
		if item != nil {
			stats.StatBlock_Increment_Mutating(&set.total, &item.total)
		}
	}
}

func go2_SolvableItemSet_RecalculateTotal(set *SolvableItemSet) {
	set.total = stats.StatBlock{}
	for _, item := range set.items {
		if item != nil {
			for i := range set.total {
				set.total[i] += item.total[i]
			}
		}
	}
}
