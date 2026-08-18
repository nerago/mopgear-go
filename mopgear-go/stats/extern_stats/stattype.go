package extern_stats

import (
	"github.com/nerago/mopgear-go/stats"
	gear_stat "github.com/nerago/mopgear-go/stats"

	wowsim_stat "github.com/wowsims/mop/sim/core/stats"
)

func GearStatToSimStat(stat gear_stat.StatType) wowsim_stat.Stat {
	switch stat {
	case gear_stat.Stat_Strength:
		return wowsim_stat.Strength
	case gear_stat.Stat_Agility:
		return wowsim_stat.Agility
	case gear_stat.Stat_Stamina:
		return wowsim_stat.Stamina
	case gear_stat.Stat_Intellect:
		return wowsim_stat.Intellect
	case gear_stat.Stat_Spirit:
		return wowsim_stat.Spirit
	case gear_stat.Stat_Hit:
		return wowsim_stat.HitRating
	case gear_stat.Stat_Crit:
		return wowsim_stat.CritRating
	case gear_stat.Stat_Haste:
		return wowsim_stat.HasteRating
	case gear_stat.Stat_Expertise:
		return wowsim_stat.ExpertiseRating
	case gear_stat.Stat_Dodge:
		return wowsim_stat.DodgeRating
	case gear_stat.Stat_Parry:
		return wowsim_stat.ParryRating
	case gear_stat.Stat_Mastery:
		return wowsim_stat.MasteryRating
	default:
		panic("unknown stat")
	}
}

func SimStatToGearStat(stat wowsim_stat.Stat) (gearStat gear_stat.StatType, inRange bool, supported bool) {
	switch stat {
	case wowsim_stat.Strength:
		return gear_stat.Stat_Strength, true, true
	case wowsim_stat.Agility:
		return gear_stat.Stat_Agility, true, true
	case wowsim_stat.Stamina:
		return gear_stat.Stat_Stamina, true, true
	case wowsim_stat.Intellect:
		return gear_stat.Stat_Intellect, true, true
	case wowsim_stat.Spirit:
		return gear_stat.Stat_Spirit, true, true
	case wowsim_stat.HitRating:
		return gear_stat.Stat_Hit, true, true
	case wowsim_stat.CritRating:
		return gear_stat.Stat_Crit, true, true
	case wowsim_stat.HasteRating:
		return gear_stat.Stat_Haste, true, true
	case wowsim_stat.ExpertiseRating:
		return gear_stat.Stat_Expertise, true, true
	case wowsim_stat.DodgeRating:
		return gear_stat.Stat_Dodge, true, true
	case wowsim_stat.ParryRating:
		return gear_stat.Stat_Parry, true, true
	case wowsim_stat.MasteryRating:
		return gear_stat.Stat_Mastery, true, true
	case 14, 15, 16, 17, 18, 20:
		return 0, true, false
	default:
		return 0, false, false
	}
}

func SimStatIndexToGearStatThrows(num int) stats.StatType {
	gearStat, inRange, supported := SimStatToGearStat(wowsim_stat.Stat(num))
	if inRange && supported {
		return gearStat
	} else if inRange {
		panic("unsupport stat type from wowsim")
	} else {
		panic("unknown stat index")
	}
}

// func SimStatIndexToGearStatThrows(num int) stats.StatType {
// 	// this may be a one-to-one for now, rather not rely on it
// 	switch num {
// 	case 0:
// 		return stats.Stat_Strength
// 	case 1:
// 		return stats.Stat_Agility
// 	case 3:
// 		return stats.Stat_Intellect
// 	case 2:
// 		return stats.Stat_Stamina
// 	case 4:
// 		return stats.Stat_Spirit
// 	case 5:
// 		return stats.Stat_Hit
// 	case 6:
// 		return stats.Stat_Crit
// 	case 7:
// 		return stats.Stat_Haste
// 	case 8:
// 		return stats.Stat_Expertise
// 	case 9:
// 		return stats.Stat_Dodge
// 	case 10:
// 		return stats.Stat_Parry
// 	case 11:
// 		return stats.Stat_Mastery
// 	case 14, 15, 16, 17, 18, 20:
// 		return stats.Stat_Invalid
// 	default:
// 		panic("unknown stat index " + strconv.Itoa(num))
// 	}
// }

// func SimBlockIndexToStatNoThrow(num int) stats.StatType {
// 	// this may be a one-to-one for now, rather not rely on it
// 	switch num {
// 	case 0:
// 		return stats.Stat_Strength
// 	case 1:
// 		return stats.Stat_Agility
// 	case 3:
// 		return stats.Stat_Intellect
// 	case 2:
// 		return stats.Stat_Stamina
// 	case 4:
// 		return stats.Stat_Spirit
// 	case 5:
// 		return stats.Stat_Hit
// 	case 6:
// 		return stats.Stat_Crit
// 	case 7:
// 		return stats.Stat_Haste
// 	case 8:
// 		return stats.Stat_Expertise
// 	case 9:
// 		return stats.Stat_Dodge
// 	case 10:
// 		return stats.Stat_Parry
// 	case 11:
// 		return stats.Stat_Mastery
// 	default:
// 		return stats.Stat_Invalid
// 	}
// }
