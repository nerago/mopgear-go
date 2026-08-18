package tools

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func WritePawnString(weights weight_types.Weight1Basic, printer *util.PrintRecorder) string {
	str := util.StringBuild2{}
	str.WriteString("( Pawn: v1: \"Gearing Weights\": Class=Paladin,Strength=")
	str.WriteFloat64(weights.Get(stats.Stat_Strength), 10)
	str.WriteString(",Stamina=")
	str.WriteFloat64(weights.Get(stats.Stat_Stamina), 10)
	str.WriteString(",CritRating=")
	str.WriteFloat64(weights.Get(stats.Stat_Crit), 10)
	str.WriteString(",HasteRating=")
	str.WriteFloat64(weights.Get(stats.Stat_Haste), 10)
	str.WriteString(",ExpertiseRating=")
	str.WriteFloat64(weights.Get(stats.Stat_Expertise), 10)
	str.WriteString(",MasteryRating=")
	str.WriteFloat64(weights.Get(stats.Stat_Mastery), 10)
	str.WriteString(",DodgeRating=")
	str.WriteFloat64(weights.Get(stats.Stat_Dodge), 10)
	str.WriteString(",ParryRating=")
	str.WriteFloat64(weights.Get(stats.Stat_Parry), 10)
	str.WriteString(", )")
	printer.PrintlnFromBuild(str)
	return str.String()
}
