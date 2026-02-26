package stats

type StatType int8

const (
	Stat_Strength  StatType = 0
	Stat_Agility   StatType = 1
	Stat_Stamina   StatType = 2
	Stat_Intellect StatType = 3
	Stat_Spirit    StatType = 4
	Stat_Hit       StatType = 5
	Stat_Crit      StatType = 6
	Stat_Haste     StatType = 7
	Stat_Expertise StatType = 8
	Stat_Dodge     StatType = 9
	Stat_Parry     StatType = 10
	Stat_Mastery   StatType = 11
	Stat_Invalid   StatType = -1
)

func (stat StatType) Name() string {
	switch stat {
	case Stat_Strength:
		return "str"
	case Stat_Agility:
		return "agi"
	case Stat_Stamina:
		return "stam"
	case Stat_Intellect:
		return "int"
	case Stat_Spirit:
		return "spi"
	case Stat_Hit:
		return "hit"
	case Stat_Crit:
		return "crit"
	case Stat_Haste:
		return "haste"
	case Stat_Expertise:
		return "expert"
	case Stat_Dodge:
		return "dodge"
	case Stat_Parry:
		return "parry"
	case Stat_Mastery:
		return "master"
	default:
		panic("unknown stat")
	}
}