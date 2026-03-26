package stats

import "strings"

type GemInfo struct {
	Id    uint32
	Stats StatBlock
}

type ReforgeRecipe struct {
	From, To StatType
}

var ReforgeRecipe_empty ReforgeRecipe = ReforgeRecipe{-1, -1}

func (reforge *ReforgeRecipe) IsEmpty() bool {
	return reforge.From < 0 || reforge.To < 0
}

func (reforge *ReforgeRecipe) AppendString(builder *strings.Builder) {
	builder.WriteRune('(')
	builder.WriteString(reforge.From.Name())
	builder.WriteString("->")
	builder.WriteString(reforge.To.Name())
	builder.WriteRune(')')
}

type ArmorType int8

const (
	Armor_None    ArmorType = -1
	Armor_Cloth   ArmorType = 1
	Armor_Leather ArmorType = 2
	Armor_Mail    ArmorType = 3
	Armor_Plate   ArmorType = 4
)

func (armor ArmorType) Matches(test ArmorType) bool {
	return armor == test || armor == Armor_None || test == Armor_None
}

type SocketType int8

const (
	Socket_Meta        SocketType = 1
	Socket_Red         SocketType = 2
	Socket_Blue        SocketType = 3
	Socket_Yellow      SocketType = 4
	Socket_General     SocketType = 8
	Socket_Engineering SocketType = 9
	Socket_Sha         SocketType = 10
)

func (socket SocketType) SocketMatch(gemStat *StatBlock) bool {
	switch socket {
	case Socket_Red:
		return gemStat[Stat_Agility] != 0 || gemStat[Stat_Strength] != 0 || gemStat[Stat_Intellect] != 0 || gemStat[Stat_Expertise] != 0
	case Socket_Yellow:
		return gemStat[Stat_Crit] != 0 || gemStat[Stat_Haste] != 0 || gemStat[Stat_Mastery] != 0
	case Socket_Blue:
		return gemStat[Stat_Hit] != 0 || gemStat[Stat_Spirit] != 0 || gemStat[Stat_Stamina] != 0
	case Socket_General, Socket_Meta, Socket_Engineering, Socket_Sha:
		return true
	default:
		panic("unexpected common.SocketType")
	}
}

type PrimaryStatType int8

const (
	PrimaryStat_None      PrimaryStatType = iota
	PrimaryStat_Strength                  = iota
	PrimaryStat_Agility                   = iota
	PrimaryStat_Intellect                 = iota
)

type Difficulty int8

const (
	Difficulty_Celestial Difficulty = iota
	Difficulty_Normal               = iota
	Difficulty_Heroic               = iota
)

func (difficulty Difficulty) ExpectedItemLevel() uint16 {
	switch difficulty {
	case Difficulty_Celestial:
		return 502
	case Difficulty_Normal:
		return 522
	case Difficulty_Heroic:
		return 535
	default:
		panic("unknown Difficulty")
	}
}

func (difficulty Difficulty) Name() string {
	switch difficulty {
	case Difficulty_Celestial:
		return "Celestial"
	case Difficulty_Normal:
		return "Normal"
	case Difficulty_Heroic:
		return "Heroic"
	default:
		panic("unknown Difficulty")
	}
}
