package main

import (
	"paladin_gearing_go/stats"
)

func loadCurrentReforges() map[uint32]stats.ReforgeRecipe {
	common := make(map[uint32]stats.ReforgeRecipe)

	common[87024] = stats.ReforgeRecipe{From: stats.Stat_Crit, To: stats.Stat_Expertise}
	common[94529] = stats.ReforgeRecipe_empty
	common[94776] = stats.ReforgeRecipe{From: stats.Stat_Hit, To: stats.Stat_Haste}
	common[95140] = stats.ReforgeRecipe_empty
	common[95513] = stats.ReforgeRecipe{From: stats.Stat_Hit, To: stats.Stat_Crit}
	common[96182] = stats.ReforgeRecipe{From: stats.Stat_Parry, To: stats.Stat_Haste}
	common[96373] = stats.ReforgeRecipe{From: stats.Stat_Haste, To: stats.Stat_Expertise}
	common[96376] = stats.ReforgeRecipe_empty
	common[96394] = stats.ReforgeRecipe{From: stats.Stat_Expertise, To: stats.Stat_Mastery}
	common[96398] = stats.ReforgeRecipe_empty
	common[96657] = stats.ReforgeRecipe{From: stats.Stat_Crit, To: stats.Stat_Haste}
	common[96769] = stats.ReforgeRecipe{From: stats.Stat_Hit, To: stats.Stat_Haste}

	return common
}
