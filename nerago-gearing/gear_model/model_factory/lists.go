package model_factory

import (
	"paladin_gearing_go/gear_model/bonus_set"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/weightfind/weight_types"
)

const setNameT15Prot = "Plate of the Lightning Emperor"
const setNameT16Prot = "Plate of Winged Triumph"
const setNameT15Ret = "Battlegear of the Lightning Emperor"
const setNameT16Ret = "Battlegear of Winged Triumph"

//var BonusItems_ZeroAll = bonus_set.ItemCountsRequiredMake(
//	setNameT16Prot, 0,
//	setNameT15Prot, 0,
//	setNameT15Ret, 0,
//	setNameT16Ret, 0,
//)

var BonusItems_ProtZero = bonus_set.ItemCountsRequiredMake(
	setNameT16Prot, 0,
	setNameT15Prot, 0,
)

var BonusItems_Prot15_2pcOnly = bonus_set.ItemCountsRequiredMake(
	setNameT16Prot, 0,
	setNameT15Prot, 2,
)

var BonusItems_Prot16_2pcOnly = bonus_set.ItemCountsRequiredMake(
	setNameT16Prot, 2,
	setNameT15Prot, 0,
)

var BonusItems_Prot16_4pc = bonus_set.ItemCountsRequiredMake(
	setNameT16Prot, 4,
)

var BonusItems_Prot15_Prot16_2pcEach = bonus_set.ItemCountsRequiredMake(
	setNameT15Prot, 2,
	setNameT16Prot, 2,
)

var BonusItems_Ret15_Ret16_2pcEach = bonus_set.ItemCountsRequiredMake(
	setNameT15Ret, 2,
	setNameT16Ret, 2,
)

// for noset - juggernaut, shamans, siegecrafter
var SimPriority_mitigation = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.2,
	stats.Sim_DEATH, 0.2,
	stats.Sim_TMI, 0.2,
	stats.Sim_DTPS, 0.4,
)

// for withset - malkrok, thok, nazgrim
var SimPriority_survival = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.01,
	stats.Sim_DEATH, 0.32,
	stats.Sim_TMI, 0.17,
	stats.Sim_DTPS, 0.50,
)

// for compromise set - spoils, galakras, paragons, etc
var SimPriority_balanced = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.40,
	stats.Sim_DEATH, 0.15,
	stats.Sim_TMI, 0.25,
	stats.Sim_DTPS, 0.20,
)

// for dps set
var SimPriority_dps = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.95,
	stats.Sim_DEATH, 0.01,
	stats.Sim_TMI, 0.03,
	stats.Sim_DTPS, 0.01,
)

// for heal set, garrosh, immerseus
var SimPriority_heal = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.15,
	stats.Sim_HPS, 0.20,
	stats.Sim_DEATH, 0.10,
	stats.Sim_TMI, 0.15,
	stats.Sim_DTPS, 0.40,
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

var trinketsStrengthTankOnly = []items.ItemId{
	104572, 105568, 105070, 102306, 105319, 104821, // Vial of Living Corruption
}

var trinketsStrengthMeleeOnly = []items.ItemId{
	102298, 104993, 104495, 105491, 104744, 105242, // Evil Eye of Galakras
}
