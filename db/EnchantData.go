package db

import (
	"fmt"
	. "paladin_gearing_go/types/stats"
)

var enchants = makeEnchants()

func makeEnchants() map[uint32]GemInfo {
	lookup := make(map[uint32]GemInfo)
	addGem0(lookup, 4099)
	addGem1(lookup, 4411, Stat_Mastery, 170)
	addGem1(lookup, 4412, Stat_Dodge, 170)
	addGem1(lookup, 4414, Stat_Intellect, 180)
	addGem1(lookup, 4415, Stat_Strength, 180)
	addGem1(lookup, 4420, Stat_Stamina, 300)
	addGem1(lookup, 4421, Stat_Hit, 180)
	addGem1(lookup, 4422, Stat_Stamina, 200)
	addGem1(lookup, 4423, Stat_Intellect, 180)
	addGem1(lookup, 4424, Stat_Crit, 180)
	addGem1(lookup, 4426, Stat_Haste, 175)
	addGem1(lookup, 4427, Stat_Hit, 175)
	addGem1(lookup, 4429, Stat_Mastery, 140)
	addGem1(lookup, 4430, Stat_Haste, 170)
	addGem1(lookup, 4431, Stat_Expertise, 170)
	addGem1(lookup, 4432, Stat_Strength, 170)
	addGem1(lookup, 4433, Stat_Mastery, 170)
	addGem1(lookup, 4434, Stat_Intellect, 165)
	addGem0(lookup, 4441) // windsong
	addGem0(lookup, 4443) // elemental force
	addGem0(lookup, 4444) // dancing steel
	addGem2(lookup, 4803, Stat_Strength, 200, Stat_Crit, 100)
	addGem2(lookup, 4805, Stat_Stamina, 300, Stat_Dodge, 100)
	addGem2(lookup, 4806, Stat_Intellect, 200, Stat_Crit, 100)
	addGem2(lookup, 4823, Stat_Strength, 285, Stat_Crit, 165)
	addGem2(lookup, 4824, Stat_Stamina, 430, Stat_Dodge, 165)
	addGem2(lookup, 4826, Stat_Intellect, 285, Stat_Spirit, 165)
	addGem0(lookup, 4892) // lightweave
	addGem1(lookup, 4993, Stat_Parry, 170)
	addGem0(lookup, 5001) // shield spike

	chestStats := StatBlock{}
	chestStats[Stat_Agility] = 80
	chestStats[Stat_Strength] = 80
	chestStats[Stat_Intellect] = 80
	chestStats[Stat_Stamina] = 80
	chestStats[Stat_Spirit] = 80
	addGemBlock(lookup, 4419, chestStats)

	return lookup
}

func EnchantData_ById(id uint32) GemInfo {
	gem, found := enchants[id]
	if !found {
		panic("unknown enchant " + fmt.Sprint(id))
	}
	return gem
}
