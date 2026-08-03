package tools

import (
	"paladin_gearing_go/gearproto"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_types"

	"google.golang.org/protobuf/encoding/protojson"
)

func WriteWeightString(weight weight_types.IWeight, printer *util.PrintRecorder) string {
	var str string
	if weightCast1, isCast1 := weight.(*weight_types.Weight1Basic); isCast1 {
		str = FormatWeight1String(weightCast1)
	} else if weightCast2, isCast2 := weight.(*weight_types.Weight2Extended); isCast2 {
		str = FormatWeight2String(weightCast2)
	} else if weightCast3, isCast3 := weight.(*weight_types.Weight3ExtendedRanged); isCast3 {
		str = FormatWeight3String(weightCast3)
	} else {
		str = ""
	}

	printer.Println(str)
	return str
}

func FormatWeight1String(weight1 *weight_types.Weight1Basic) string {
	protoWeight := gearproto.Weight1Basic{WeightType: 1}
	weight := convertWeight1Details(weight1)
	protoWeight.Weight = weight

	return protojson.MarshalOptions{}.Format(&protoWeight)
}

func FormatWeight2String(weight2 *weight_types.Weight2Extended) string {
	protoWeight := gearproto.Weight2Extended{WeightType: 2}
	protoWeight.Weights = convertWeight2Details(weight2)
	protoWeight.Priority = convertPriority(&weight2.SimPriority, weight2.SimList)

	return protojson.MarshalOptions{}.Format(&protoWeight)
}

func FormatWeight3String(weight3 *weight_types.Weight3ExtendedRanged) string {
	protoWeight := gearproto.Weight3Extended{WeightType: 3}
	protoWeight.Weights = convertWeight3Details(weight3)
	protoWeight.Priority = convertPriority(&weight3.SimPriority, weight3.SimList)

	return protojson.MarshalOptions{}.Format(&protoWeight)
}

func convertWeight1Details(weight1 *weight_types.Weight1Basic) []*gearproto.Weight1Entry {
	weight := make([]*gearproto.Weight1Entry, 0)
	for statType, value := range weight1.SeqPair() {
		weight = append(weight, &gearproto.Weight1Entry{
			StatType:     convertStatType(statType),
			RatingWeight: value,
		})
	}
	return weight
}

func convertWeight2Details(weight2 *weight_types.Weight2Extended) []*gearproto.Weight2Entry {
	weights := make([]*gearproto.Weight2Entry, 0)
	for weightEntry := range weight2.DetailedWeights.SeqKey1Key2ValueEntries() {
		protoEntry := gearproto.Weight2Entry{}
		protoEntry.SimType = convertSimType(weightEntry.Key2)
		protoEntry.StatType = convertStatType(weightEntry.Key1)
		protoEntry.RatingWeight = weightEntry.Value
		weights = append(weights, &protoEntry)
	}
	return weights
}

func convertWeight3Details(weight3 *weight_types.Weight3ExtendedRanged) []*gearproto.Weight3Group {
	weights := make([]*gearproto.Weight3Group, 0)
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
		weights = append(weights, &group)
	}
	return weights
}

func convertPriority(simPriority *weight_types.SimPriorityExtended, simList []stats.SimType) []*gearproto.Priority2Entry {
	priority := make([]*gearproto.Priority2Entry, 0)
	for _, simType := range simList {
		prior := simPriority.GetOrPanic(simType)
		priority = append(priority, &gearproto.Priority2Entry{
			SimType:       convertSimType(simType),
			RangingScale:  prior.RangingScale,
			RangingOffset: prior.RangingOffset,
			RatioScale:    prior.RatioScale,
		})
	}
	return priority
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
