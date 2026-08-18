package tools

import (
	"errors"
	"github.com/nerago/mopgear-go/gearproto"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
	"io/fs"
	"os"
	"slices"

	"google.golang.org/protobuf/encoding/protojson"
)

func WriteWeightString(weight weight_types.IWeight, printer *util.PrintRecorder) string {
	var str string

	switch cast := weight.(type) {
	case *weight_types.Weight1Basic:
		str = FormatWeight1String(cast)
	case *weight_types.Weight2Extended:
		str = FormatWeight2String(cast)
	case *weight_types.Weight3ExtendedRanged:
		str = FormatWeight3String(cast)
	default:
		panic("unknown weight type")
	}

	printer.Println(str)
	return str
}

func ReadWeight2File(filename string) (*weight_types.Weight2Extended, bool) {
	bytes, err := os.ReadFile(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false
	} else if err != nil {
		panic(err)
	}

	protoWeight := gearproto.Weight2Extended{}
	err = protojson.UnmarshalOptions{}.Unmarshal(bytes, &protoWeight)
	if err != nil {
		panic(err)
	}

	return buildGearWeight2(&protoWeight), true
}

func ReadWeight3File(filename string) (*weight_types.Weight3ExtendedRanged, bool) {
	bytes, err := os.ReadFile(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false
	} else if err != nil {
		panic(err)
	}

	protoWeight := gearproto.Weight3Extended{}
	err = protojson.UnmarshalOptions{}.Unmarshal(bytes, &protoWeight)
	if err != nil {
		panic(err)
	}

	return buildGearWeight3(&protoWeight), true
}

func FormatWeight1String(weight1 *weight_types.Weight1Basic) string {
	protoWeight := gearproto.Weight1Basic{WeightType: 1}
	protoWeight.Weight = convertWeight1Details(weight1)

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
		if util.FloatNonZero(value) {
			weight = append(weight, &gearproto.Weight1Entry{
				StatType:     convertStatType(statType),
				RatingWeight: value,
			})
		}
	}
	return weight
}

func convertWeight2Details(weight2 *weight_types.Weight2Extended) []*gearproto.Weight2Entry {
	weights := make([]*gearproto.Weight2Entry, 0)
	for _, simType := range weight2.SimList {
		for _, statType := range weight2.StatList {
			weightEntry, hasEntry := weight2.DetailedWeights.Get(statType, simType)
			if hasEntry {
				protoEntry := gearproto.Weight2Entry{}
				protoEntry.SimType = convertSimType(simType)
				protoEntry.StatType = convertStatType(statType)
				protoEntry.RatingWeight = weightEntry
				weights = append(weights, &protoEntry)
			}
		}
	}
	return weights
}

func buildGearWeight2(protoWeight *gearproto.Weight2Extended) *weight_types.Weight2Extended {
	weight2 := &weight_types.Weight2Extended{}
	for _, ent := range protoWeight.GetWeights() {
		weight2.PutWeight(
			convertStatTypeReverse(ent.StatType),
			convertSimTypeReverse(ent.SimType),
			ent.RatingWeight,
		)
	}
	for _, pri := range protoWeight.GetPriority() {
		weight2.SetSimScale(convertSimTypeReverse(pri.SimType), pri.RangingScale, pri.RangingOffset, pri.RatioScale)
	}
	weight2.StatList = slices.Collect(weight2.DetailedWeights.SeqKey1())
	weight2.SimList = slices.Collect(weight2.DetailedWeights.SeqKey2())
	weight2.FinishAndValidate()
	return weight2
}

func convertWeight3Details(weight3 *weight_types.Weight3ExtendedRanged) []*gearproto.Weight3Group {
	weights := make([]*gearproto.Weight3Group, 0)
	for _, simType := range weight3.SimList {
		for _, statType := range weight3.StatList {
			weightEntry, hasEntry := weight3.StatWeights.GetAsSliceClone(simType, statType)
			if hasEntry {
				group := gearproto.Weight3Group{}
				group.SimType = convertSimType(simType)
				group.StatType = convertStatType(statType)
				group.Entries = make([]*gearproto.Weight3Entry, 0)
				for _, entry := range weightEntry {
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
		}
	}
	return weights
}

func buildGearWeight3(protoWeight *gearproto.Weight3Extended) *weight_types.Weight3ExtendedRanged {
	weight3 := &weight_types.Weight3ExtendedRanged{}
	for _, group := range protoWeight.GetWeights() {
		for _, ent := range group.Entries {
			weight3.AddDetailWeight(
				convertSimTypeReverse(group.SimType),
				convertStatTypeReverse(group.StatType),
				weight_types.StatRange{Minimum: ent.StatRange.Minimum, Maximum: ent.StatRange.Maximum},
				ent.RatingWeight,
				ent.RatingOffset,
				0,
			)
		}
	}
	for _, pri := range protoWeight.GetPriority() {
		weight3.SetSimScale(convertSimTypeReverse(pri.SimType), pri.RangingScale, pri.RangingOffset, pri.RatioScale)
	}
	weight3.StatList = slices.Collect(weight3.StatWeights.SeqKey2())
	weight3.SimList = slices.Collect(weight3.StatWeights.SeqKey1())
	weight3.FinishAndValidate()
	return weight3
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

func convertStatTypeReverse(stat gearproto.StatType) stats.StatType {
	switch stat {
	case gearproto.StatType_Strength:
		return stats.Stat_Strength
	case gearproto.StatType_Agility:
		return stats.Stat_Agility
	case gearproto.StatType_Stamina:
		return stats.Stat_Stamina
	case gearproto.StatType_Intellect:
		return stats.Stat_Intellect
	case gearproto.StatType_Spirit:
		return stats.Stat_Spirit
	case gearproto.StatType_Hit:
		return stats.Stat_Hit
	case gearproto.StatType_Crit:
		return stats.Stat_Crit
	case gearproto.StatType_Haste:
		return stats.Stat_Haste
	case gearproto.StatType_Expertise:
		return stats.Stat_Expertise
	case gearproto.StatType_Dodge:
		return stats.Stat_Dodge
	case gearproto.StatType_Parry:
		return stats.Stat_Parry
	case gearproto.StatType_Mastery:
		return stats.Stat_Mastery
	default:
		panic("invalid StatType")
	}
}

func convertSimTypeReverse(sim gearproto.SimType) stats.SimType {
	switch sim {
	case gearproto.SimType_DPS:
		return stats.Sim_DPS
	case gearproto.SimType_TPS:
		return stats.Sim_TPS
	case gearproto.SimType_DTPS:
		return stats.Sim_DTPS
	case gearproto.SimType_HPS:
		return stats.Sim_HPS
	case gearproto.SimType_TMI:
		return stats.Sim_TMI
	case gearproto.SimType_DEATH:
		return stats.Sim_DEATH
	default:
		panic("invalid SimType")
	}
}
