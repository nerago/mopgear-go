package model_factory

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

// for noset - juggernaut, shamans, siegecrafter
var SimPriority_mitigation = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.2489,
	stats.Sim_DEATH, 0.1981,
	stats.Sim_TMI, 0.1530,
	stats.Sim_DTPS, 0.4000,
	// DPS=0.2489 DTPS=0.4000 TMI=0.1530 DEATH=0.1981
)

// for withset - malkrok, thok, nazgrim
var SimPriority_survival = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.1061,
	stats.Sim_DEATH, 0.2801,
	stats.Sim_TMI, 0.1700,
	stats.Sim_DTPS, 0.4438,
	// DPS=0.1061 DTPS=0.4438 TMI=0.1700 DEATH=0.2801
)

// for compromise set - spoils, galakras, paragons, etc
var SimPriority_balanced = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.5000,
	stats.Sim_DEATH, 0.2022,
	stats.Sim_TMI, 0.1978,
	stats.Sim_DTPS, 0.1000,
	// DPS=0.5000 DTPS=0.1000 TMI=0.1978 DEATH=0.2022
)

// for dps set
var SimPriority_dps = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.8500,
	stats.Sim_DEATH, 0.1000,
	stats.Sim_TMI, 0.0400,
	stats.Sim_DTPS, 0.0100,
	// DPS=0.8500 DTPS=0.0000 TMI=0.0400 DEATH=0.1100
)

// for heal set, garrosh, immerseus
var SimPriority_heal = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.0900,
	stats.Sim_HPS, 0.3000,
	stats.Sim_DEATH, 0.0100,
	stats.Sim_TMI, 0.2500,
	stats.Sim_DTPS, 0.3500,
	//DPS=0.0989 DTPS=0.3509 HPS=0.3000 TMI=0.2500 DEATH=0.0003
)

// for ret set
var SimPriority_ret = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 1,
)

var StatsForWeighting_strengthTank = []stats.StatType{
	stats.Stat_Strength,
	stats.Stat_Stamina,
	stats.Stat_Haste,
	stats.Stat_Mastery,
	stats.Stat_Crit,
	stats.Stat_Dodge,
	stats.Stat_Parry,
	stats.Stat_Expertise,
}

var StatsForWeighting_strengthMelee = []stats.StatType{
	stats.Stat_Strength,
	stats.Stat_Haste,
	stats.Stat_Mastery,
	stats.Stat_Crit,
	stats.Stat_Expertise,
}
