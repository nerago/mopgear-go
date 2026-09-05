package tools

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func WritePawnString(weights *weight_types.Weight1_CompatibleExternal, specType stats.SpecType, printer *util.PrintRecorder) string {
	str := FormatPawnString(weights, specType)
	printer.Println(str)
	return str
}

var pawnKeys = map[stats.StatType]string{
	stats.Stat_Strength:  "Strength",
	stats.Stat_Agility:   "Agility",
	stats.Stat_Stamina:   "Stamina",
	stats.Stat_Intellect: "Intellect",
	stats.Stat_Spirit:    "Spirit",
	stats.Stat_Hit:       "HitRating",
	stats.Stat_Crit:      "CritRating",
	stats.Stat_Haste:     "HasteRating",
	stats.Stat_Expertise: "ExpertiseRating",
	stats.Stat_Dodge:     "DodgeRating",
	stats.Stat_Parry:     "ParryRating",
	stats.Stat_Mastery:   "MasteryRating",
}

func FormatPawnString(weights *weight_types.Weight1_CompatibleExternal, specType stats.SpecType) string {
	str := util.StringBuild2{}
	str.WriteString("( Pawn: v1: \"Gearing Weights\": Class=")
	str.WriteString(specType.ClassName())
	str.WriteString(",")

	for statType, value := range weights.SeqPair() {
		if util.FloatNonZero(value) {
			str.WriteString(pawnKeys[statType])
			str.WriteString("=")
			str.WriteFloat64(value, -1)
			str.WriteString(",")
		}
	}
	str.Rewind(1)
	str.WriteString(" )")
	return str.String()
}
