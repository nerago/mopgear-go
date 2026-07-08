package main

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
)

func commonComboCurrent() map[items.ItemId]stats.ReforgeRecipe {
	common := make(map[items.ItemId]stats.ReforgeRecipe)
	common[95020] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Mastery)
	common[102306] = stats.ReforgeRecipe_empty
	common[105111] = stats.ReforgeRecipe_empty
	common[94773] = stats.ReforgeRecipe_empty
	common[99126] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Haste)
	common[103734] = stats.ReforgeRecipe_of(stats.Stat_Crit, stats.Stat_Haste)
	common[102250] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Haste)
	common[96666] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Dodge)
	common[102249] = stats.ReforgeRecipe_empty
	common[96542] = stats.ReforgeRecipe_empty
	common[94529] = stats.ReforgeRecipe_empty
	common[103735] = stats.ReforgeRecipe_empty
	common[103872] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Haste)
	common[99129] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Haste)
	common[101947] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Dodge)
	common[96398] = stats.ReforgeRecipe_of(stats.Stat_Haste, stats.Stat_Mastery)
	common[96500] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Mastery)
	common[103916] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Haste)
	common[105785] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Mastery)
	common[103787] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Mastery)
	common[96668] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Haste)
	common[95011] = stats.ReforgeRecipe_of(stats.Stat_Dodge, stats.Stat_Haste)
	common[103738] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Mastery)
	common[101882] = stats.ReforgeRecipe_of(stats.Stat_Expertise, stats.Stat_Haste)
	common[103826] = stats.ReforgeRecipe_of(stats.Stat_Hit, stats.Stat_Haste)
	return common
}
