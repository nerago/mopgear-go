package model

import (
	. "paladin_gearing_go/stats"
)

type ReforgeRules struct {
	source []StatType
	target []StatType
}

func (rules *ReforgeRules) Source() []StatType {
	return rules.source
}

func (rules *ReforgeRules) Target() []StatType {
	return rules.target
}

var standardSource = []StatType{Stat_Hit, Stat_Spirit, Stat_Expertise, Stat_Mastery, Stat_Haste, Stat_Crit, Stat_Dodge, Stat_Parry}

var targetsTank = []StatType{Stat_Hit, Stat_Expertise, Stat_Mastery, Stat_Haste, Stat_Crit, Stat_Dodge}
var targetsMelee = []StatType{Stat_Hit, Stat_Expertise, Stat_Mastery, Stat_Haste, Stat_Crit}
var targetsCasterPure = []StatType{Stat_Hit, Stat_Mastery, Stat_Haste, Stat_Crit}
var targetsCasterHybrid = []StatType{Stat_Hit, Stat_Spirit, Stat_Mastery, Stat_Haste, Stat_Crit}

var ReforgeRules_tank = ReforgeRules{standardSource, targetsTank}
var ReforgeRules_melee = ReforgeRules{standardSource, targetsMelee}
var ReforgeRules_casterPure = ReforgeRules{standardSource, targetsCasterPure}
var ReforgeRules_casterHybrid = ReforgeRules{standardSource, targetsCasterHybrid}
