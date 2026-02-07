package db

import (
	. "paladin_gearing_go/types/stats"
	. "paladin_gearing_go/util"
)

func addGem0(lookup map[int32]GemInfo, id int32) {
	lookup[id] = GemInfo{id, StatBlock_empty}
}

func addGem1(lookup map[int32]GemInfo, id int32, stat StatType, value uint32) {
	lookup[id] = GemInfo{id, StatBlock_of(stat, value)}
}

func addGem2(lookup map[int32]GemInfo, id int32, statA StatType, valueA uint32, statB StatType, valueB uint32) {
	lookup[id] = GemInfo{id, StatBlock_of2(statA, valueA, statB, valueB)}
}

func addGemBlock(lookup map[int32]GemInfo, id int32, block StatBlock) {
	lookup[id] = GemInfo{id, block}
}

var standardGems = makeStandardGems()
var engineerGems = makeEngineerGems()
var metaGems = makeMetaGems()
var shaGems = makeShaGems()
var allGems = CombineMaps(standardGems, engineerGems, metaGems, shaGems)

func makeStandardGems() map[int32]GemInfo {
	lookup := make(map[int32]GemInfo)
	addGem2(lookup, 76537, Stat_Strength, 60, Stat_Haste, 120)
	addGem1(lookup, 76570, Stat_Hit, 320)
	addGem2(lookup, 76576, Stat_Hit, 160, Stat_Haste, 160)
	addGem2(lookup, 76585, Stat_Haste, 160, Stat_Spirit, 160)
	addGem2(lookup, 76588, Stat_Stamina, 120, Stat_Haste, 160)
	addGem2(lookup, 76593, Stat_Crit, 160, Stat_Expertise, 160)
	addGem2(lookup, 76601, Stat_Haste, 160, Stat_Expertise, 160)
	addGem2(lookup, 76603, Stat_Strength, 80, Stat_Haste, 160)
	addGem2(lookup, 76606, Stat_Intellect, 80, Stat_Mastery, 160)
	addGem2(lookup, 76615, Stat_Hit, 160, Stat_Expertise, 160)
	addGem2(lookup, 76618, Stat_Strength, 80, Stat_Hit, 160)
	addGem1(lookup, 76627, Stat_Expertise, 320)
	addGem1(lookup, 76628, Stat_Intellect, 160)
	addGem1(lookup, 76633, Stat_Haste, 320)
	addGem1(lookup, 76636, Stat_Hit, 320)
	addGem2(lookup, 76642, Stat_Hit, 160, Stat_Haste, 160)
	addGem2(lookup, 76654, Stat_Stamina, 120, Stat_Haste, 160)
	addGem2(lookup, 76667, Stat_Haste, 160, Stat_Expertise, 160)
	addGem2(lookup, 76668, Stat_Intellect, 80, Stat_Haste, 160)
	addGem2(lookup, 76669, Stat_Strength, 80, Stat_Haste, 160)
	addGem2(lookup, 76681, Stat_Hit, 160, Stat_Expertise, 160)
	addGem2(lookup, 76682, Stat_Intellect, 80, Stat_Hit, 160)
	addGem2(lookup, 76686, Stat_Intellect, 80, Stat_Spirit, 160)
	addGem1(lookup, 76693, Stat_Expertise, 320)
	addGem1(lookup, 76694, Stat_Intellect, 160)
	addGem1(lookup, 76696, Stat_Strength, 320)
	addGem1(lookup, 76697, Stat_Crit, 320)
	addGem1(lookup, 76699, Stat_Haste, 320)
	addGem1(lookup, 76700, Stat_Mastery, 320)
	return lookup
}

func makeEngineerGems() map[int32]GemInfo {
	lookup := make(map[int32]GemInfo)
	addGem1(lookup, 77541, Stat_Crit, 600)
	addGem1(lookup, 77542, Stat_Haste, 600)
	addGem1(lookup, 77543, Stat_Expertise, 600)
	addGem1(lookup, 77545, Stat_Hit, 600)
	addGem1(lookup, 77546, Stat_Spirit, 600)
	addGem1(lookup, 77547, Stat_Mastery, 600)
	return lookup
}

func makeMetaGems() map[int32]GemInfo {
	lookup := make(map[int32]GemInfo)
	addGem1(lookup, 76885, Stat_Intellect, 216)
	addGem1(lookup, 76886, Stat_Strength, 216)
	addGem1(lookup, 76895, Stat_Stamina, 324)
	addGem1(lookup, 95344, Stat_Stamina, 324)
	addGem1(lookup, 95346, Stat_Crit, 324)
	return lookup
}

func makeShaGems() map[int32]GemInfo {
	lookup := make(map[int32]GemInfo)
	addGem1(lookup, 89881, Stat_Strength, 500)
	return lookup
}

func GemData_ById(id int32) GemInfo {
	gem, found := allGems[id]
	if !found {
		panic("unknown gem " + string(id))
	}
	return gem
}

// NOTE only needed with WowHead lookup
// var socketBonus = makeSocketBonus()
// func makeSocketBonus() map[int32]StatBlock {
// 	lookup := make(map[int32]StatBlock)
// 	lookup[4827] = StatBlock_of(Primary, 60)
// 	lookup[4828] = StatBlock_of(Primary, 120)
// 	lookup[4829] = StatBlock_of(Primary, 180)
// 	lookup[4830] = StatBlock_of(Primary, 60)
// 	lookup[4831] = StatBlock_of(Primary, 80)
// 	lookup[4832] = StatBlock_of(Stat_Stamina, 90)
// 	lookup[4833] = StatBlock_of(Stat_Crit, 60)
// 	lookup[4834] = StatBlock_of(Stat_Mastery, 60)
// 	lookup[4835] = StatBlock_of(Stat_Hit, 60)
// 	lookup[4836] = StatBlock_of(Stat_Haste, 60)
// 	lookup[4837] = StatBlock_of(Stat_Expertise, 60)
// 	lookup[4838] = StatBlock_of(Stat_Spirit, 60)
// 	lookup[4839] = StatBlock_of(Stat_Dodge, 60)
// 	lookup[4840] = StatBlock_of(Stat_Parry, 60)
// 	lookup[4842] = StatBlock_empty
// 	lookup[4843] = StatBlock_of(Stat_Crit, 120)
// 	lookup[4844] = StatBlock_of(Stat_Dodge, 120)
// 	lookup[4845] = StatBlock_of(Stat_Expertise, 120)
// 	lookup[4846] = StatBlock_of(Stat_Haste, 120)
// 	lookup[4848] = StatBlock_of(Primary, 120)
// 	lookup[4850] = StatBlock_of(Stat_Parry, 120)
// 	lookup[4851] = StatBlock_empty
// 	lookup[4852] = StatBlock_of(Stat_Spirit, 120)
// 	lookup[4853] = StatBlock_of(Primary, 120)
// 	lookup[4854] = StatBlock_of(Stat_Stamina, 180)
// 	lookup[4855] = StatBlock_of(Stat_Crit, 180)
// 	lookup[4858] = StatBlock_of(Stat_Haste, 180)
// 	lookup[4860] = StatBlock_of(Primary, 180)
// 	lookup[4863] = StatBlock_empty
// 	lookup[4867] = StatBlock_of(Stat_Stamina, 270)
// 	lookup[4868] = StatBlock_of(Primary, 180)
// 	return lookup
// }
