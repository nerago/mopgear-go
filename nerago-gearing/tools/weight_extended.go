package tools

import (
	"paladin_gearing_go/gearproto"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_types"

	"google.golang.org/protobuf/encoding/protojson"
)

func WriteWeight3String(weight3 weight_types.Weight3ExtendedRanged, printer *util.PrintRecorder) string {
	protoWeight := gearproto.Weight3Extended{}

	protoWeight.Weights = make([]*gearproto.Weight3Group, 0)
	for weightEntry := range weight3.StatWeights.SeqKey1Key2ValueSeqEntries() {
		group := gearproto.Weight3Group{}
		group.SimType = convertSimType(weightEntry.Key1)
		group.StatType = convertStatType(weightEntry.Key2)
		group.Entries = make([]*gearproto.Weight3Entry, 0)
		for entry := range weightEntry.ValueSeq {
			group.Entries = append(group.Entries, &gearproto.Weight3Entry{
				StatRange: &gearproto.StatRange{
					Minimum: entry.StatRange.Minimum,
					Maximum: entry.StatRange.Maximum,
				},
				RatingWeight: entry.RatingWeight,
				RatingOffset: entry.RatingOffset,
			})
		}
		protoWeight.Weights = append(protoWeight.Weights, &group)
	}

	protoWeight.Priority = make([]*gearproto.Priority2Entry, 0)
	for _, simType := range weight3.SimList {
		prior := weight3.SimPriority.GetOrPanic(simType)
		protoWeight.Priority = append(protoWeight.Priority, &gearproto.Priority2Entry{
			SimType:       convertSimType(simType),
			RangingScale:  prior.RangingScale,
			RangingOffset: prior.RangingOffset,
			RatioScale:    prior.RatioScale,
		})
	}

	str := protojson.MarshalOptions{}.Format(&protoWeight)
	//str := protojson.Format(&protoWeight)
	printer.Println(str)
	return str
}

func convertStatType(stat stats.StatType) gearproto.StatType {
	switch stat {
	case stats.Stat_Strength:
		return gearproto.StatType_Strength
	case stats.Stat_Agility:
		return gearproto.StatType_Agility
	case stats.Stat_Stamina:
		return gearproto.StatType_Stamina
	case stats.Stat_Intellect:
		return gearproto.StatType_Intellect
	case stats.Stat_Spirit:
		return gearproto.StatType_Spirit
	case stats.Stat_Hit:
		return gearproto.StatType_Hit
	case stats.Stat_Crit:
		return gearproto.StatType_Crit
	case stats.Stat_Haste:
		return gearproto.StatType_Haste
	case stats.Stat_Expertise:
		return gearproto.StatType_Expertise
	case stats.Stat_Dodge:
		return gearproto.StatType_Dodge
	case stats.Stat_Parry:
		return gearproto.StatType_Parry
	case stats.Stat_Mastery:
		return gearproto.StatType_Mastery
	default:
		panic("invalid StatType")
	}
}

func convertSimType(sim stats.SimType) gearproto.SimType {
	switch sim {
	case stats.Sim_DPS:
		return gearproto.SimType_DPS
	case stats.Sim_TPS:
		return gearproto.SimType_TPS
	case stats.Sim_DTPS:
		return gearproto.SimType_DTPS
	case stats.Sim_HPS:
		return gearproto.SimType_HPS
	case stats.Sim_TMI:
		return gearproto.SimType_TMI
	case stats.Sim_DEATH:
		return gearproto.SimType_DEATH
	default:
		panic("invalid SimType")
	}
}
