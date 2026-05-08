//go:build !amd64
// +build !amd64

package items

import "paladin_gearing_go/stats"

func SolvableItemSet_RecalculateTotal(set *SolvableItemSet) {
	set.total = stats.StatBlock{}
	for _, item := range set.items {
		if item != nil {
			for i := range set.total {
				set.total[i] += item.total[i]
			}
		}
	}
}
