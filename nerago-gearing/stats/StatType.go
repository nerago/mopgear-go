package stats

import (
	"paladin_gearing_go/util/util_collection"
)

type StatType uint8

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

	Stat_Invalid StatType = 255
	Stat_Count            = 12
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

func (stat StatType) EnumName() string {
	switch stat {
	case Stat_Strength:
		return "Stat_Strength"
	case Stat_Agility:
		return "Stat_Agility"
	case Stat_Stamina:
		return "Stat_Stamina"
	case Stat_Intellect:
		return "Stat_Intellect"
	case Stat_Spirit:
		return "Stat_Spirit"
	case Stat_Hit:
		return "Stat_Hit"
	case Stat_Crit:
		return "Stat_Crit"
	case Stat_Haste:
		return "Stat_Haste"
	case Stat_Expertise:
		return "Stat_Expertise"
	case Stat_Dodge:
		return "Stat_Dodge"
	case Stat_Parry:
		return "Stat_Parry"
	case Stat_Mastery:
		return "Stat_Mastery"
	default:
		panic("unknown stat")
	}
}

var StatType_List = []StatType{
	Stat_Strength, Stat_Agility, Stat_Stamina, Stat_Intellect,
	Stat_Spirit, Stat_Hit, Stat_Crit, Stat_Haste,
	Stat_Expertise, Stat_Dodge, Stat_Parry, Stat_Mastery}

var StatTypeEnum = util_collection.EnumTypeMake[StatType](StatType_List)

type StatTypeMap[V any] struct {
	util_collection.EnumMapTiny[StatType, V, [Stat_Count]V]
}

func (m *StatTypeMap[V]) Clone() *StatTypeMap[V] {
	return &StatTypeMap[V]{m.EnumMapTiny.Clone()}
}

func (stat StatType) EnumNumValues() uint8 {
	return Stat_Count
}
