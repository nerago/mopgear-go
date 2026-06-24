package stats

import "paladin_gearing_go/util"

type GemInfo struct {
	Id    uint32
	Stats StatBlock
}

func (gem *GemInfo) AppendString(build *util.StringBuild2) {
	name := gem.Name()
	if name != "" {
		build.WriteRune('{')
		build.WriteString(name)
		build.WriteRune('}')
	} else {
		gem.Stats.AppendString(build)
	}
}

func (gem *GemInfo) Name() string {
	switch gem.Id {
	case 95344:
		return "Indomitable"
	case 95345:
		return "Courageous"
	case 95346:
		return "Capacitive"
	case 95347:
		return "Sinister"
	default:
		return ""
	}
}

type ReforgeRecipe struct {
	hasValue bool
	From, To StatType
}

var ReforgeRecipe_empty ReforgeRecipe = ReforgeRecipe{}

func ReforgeRecipe_of(from, to StatType) ReforgeRecipe {
	return ReforgeRecipe{hasValue: true, From: from, To: to}
}

func ReforgeRecipe_of_pointer(from, to StatType) *ReforgeRecipe {
	return &ReforgeRecipe{hasValue: true, From: from, To: to}
}

func (reforge ReforgeRecipe) IsEmpty() bool {
	return !reforge.hasValue
}

func (reforge ReforgeRecipe) Equals(other *ReforgeRecipe) bool {
	if reforge.hasValue && other.hasValue {
		return reforge.To == other.To && reforge.From == other.From
	} else if !reforge.hasValue && !other.hasValue {
		return true
	} else {
		return false
	}
}

func (reforge ReforgeRecipe) AppendString(builder *util.StringBuild2) {
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

func (socket SocketType) IsStandard() bool {
	switch socket {
	case Socket_Red, Socket_Yellow, Socket_Blue, Socket_General:
		return true
	case Socket_Meta, Socket_Engineering, Socket_Sha:
		return false
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

func (difficulty Difficulty) ExpectedItemLevelThrone() uint16 {
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

func (difficulty Difficulty) ExpectedItemLevelSiege() uint16 {
	switch difficulty {
	case Difficulty_Celestial:
		return 528
	case Difficulty_Normal:
		return 553
	case Difficulty_Heroic:
		return 566
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

type WowSim_Fight int8

const (
	Fight_Unknown           WowSim_Fight = iota
	Fight_Horridon_HighHeal WowSim_Fight = iota
	Fight_Horridon_LowHeal  WowSim_Fight = iota
	Fight_Animus            WowSim_Fight = iota
	Fight_Juggernaut        WowSim_Fight = iota
	// TODO highheal variant
)

type StatAndValue struct {
	StatType StatType
	Value    uint32
}

type OptimiseGoal int8

const (
	OptimiseGoal_Unknown      OptimiseGoal = iota
	OptimiseGoal_Mitigation   OptimiseGoal = iota
	OptimiseGoal_Dps          OptimiseGoal = iota
	OptimiseGoal_HalfMitiDps  OptimiseGoal = iota
	OptimiseGoal_Healing      OptimiseGoal = iota
	OptimiseGoal_HalfMitiHeal OptimiseGoal = iota
)

func (up OptimiseGoal) Name() string {
	switch up {
	case OptimiseGoal_Mitigation:
		return "miti"
	case OptimiseGoal_Dps:
		return "dps"
	case OptimiseGoal_HalfMitiDps:
		return "mit/dps"
	case OptimiseGoal_Healing:
		return "heal"
	case OptimiseGoal_HalfMitiHeal:
		return "mit/heal"
	default:
		panic("unknown")
	}
}
