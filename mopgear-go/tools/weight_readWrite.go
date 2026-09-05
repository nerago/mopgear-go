package tools

import (
	"errors"
	"io/fs"
	"os"
	"slices"

	"github.com/nerago/mopgear-go/gearproto"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func WriteWeightString(weight weight_types.IWeight, printer *util.PrintRecorder) error {
	var str string

	switch cast := weight.(type) {
	case *weight_types.Weight1_ScaledSolvable:
		str = FormatWeight1String(cast)
	case *weight_types.Weight2:
		str = FormatWeight2String(cast)
	case *weight_types.Weight3:
		str = FormatWeight3String(cast)
	default:
		return util.ErrorTracedNew("unknown weight type")
	}

	printer.Println(str)
	return nil
}

func ReadWeight1File(filename string) (weight1 *weight_types.Weight1_ScaledSolvable, err error) {
	protoWeight := gearproto.Weight1{}

	goodRead, err := readWeightFileGeneral(filename, &protoWeight)
	if err != nil || !goodRead {
		return nil, err
	}

	weight1 = buildGearWeight1(&protoWeight)
	return weight1, nil
}

func ReadWeight2File(filename string) (*weight_types.Weight2, error) {
	protoWeight := gearproto.Weight2{}

	goodRead, err := readWeightFileGeneral(filename, &protoWeight)
	if err != nil || !goodRead {
		return nil, err
	}

	weight2, err := buildGearWeight2(&protoWeight)
	if err != nil {
		return nil, err
	}
	return weight2, nil
}

func ReadWeight3File(filename string) (*weight_types.Weight3, error) {
	protoWeight := gearproto.Weight3{}

	goodRead, err := readWeightFileGeneral(filename, &protoWeight)
	if err != nil || !goodRead {
		return nil, err
	}

	weight, err := buildGearWeight3(&protoWeight)
	if err != nil {
		return nil, err
	}
	return weight, nil
}

func readWeightFileGeneral[T proto.Message](filename string, outputPointer T) (bool, error) {
	bytes, err := os.ReadFile(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	err = protojson.UnmarshalOptions{}.Unmarshal(bytes, outputPointer)
	if err != nil {
		return false, err
	}
	return true, nil
}

func FormatWeight1String(weight1 *weight_types.Weight1_ScaledSolvable) string {
	protoWeight := gearproto.Weight1{WeightType: 1}
	protoWeight.Weight = convertWeight1Details(weight1)

	scaleOffset := weight1.GetScaleOffset()
	protoWeight.Scale = scaleOffset.Scale
	protoWeight.Offset = scaleOffset.Offset

	return protojson.MarshalOptions{}.Format(&protoWeight)
}

func FormatWeight2String(weight2 *weight_types.Weight2) string {
	protoWeight := gearproto.Weight2{WeightType: 2}
	protoWeight.Weights = convertWeight2Details(weight2)
	protoWeight.Priority = convertPriority(&weight2.SimPriority, weight2.SimList)

	return protojson.MarshalOptions{}.Format(&protoWeight)
}

func FormatWeight3String(weight3 *weight_types.Weight3) string {
	protoWeight := gearproto.Weight3{WeightType: 3}
	protoWeight.Weights = convertWeight3Details(weight3)
	protoWeight.Priority = convertPriority(&weight3.SimPriority, weight3.SimList)

	return protojson.MarshalOptions{}.Format(&protoWeight)
}

func convertWeight1Details(weight1 *weight_types.Weight1_ScaledSolvable) []*gearproto.StatTypeAndFloatWeight {
	weight := make([]*gearproto.StatTypeAndFloatWeight, 0)
	for statType, value := range weight1.SeqPair() {
		if util.FloatNonZero(value) {
			weight = append(weight, &gearproto.StatTypeAndFloatWeight{
				StatType: convertStatType(statType),
				Weight:   value,
			})
		}
	}
	return weight
}

func convertWeight2Details(weight2 *weight_types.Weight2) []*gearproto.StatSimTypesAndFloatWeight {
	weights := make([]*gearproto.StatSimTypesAndFloatWeight, 0)
	for _, simType := range weight2.SimList {
		for _, statType := range weight2.StatList {
			weightEntry, hasEntry := weight2.DetailedWeights.Get(simType, statType)
			if hasEntry {
				protoEntry := gearproto.StatSimTypesAndFloatWeight{}
				protoEntry.SimType = convertSimType(simType)
				protoEntry.StatType = convertStatType(statType)
				protoEntry.Weight = weightEntry
				weights = append(weights, &protoEntry)
			}
		}
	}
	return weights
}

func buildGearWeight1(protoWeight *gearproto.Weight1) *weight_types.Weight1_ScaledSolvable {
	weight1 := &weight_types.Weight1_ScaledSolvable{}
	for _, ent := range protoWeight.Weight {
		weight1.Put(convertStatTypeReverse(ent.StatType), ent.Weight)
	}
	weight1.SetScaleOffset(weight_types.ScaleAndOffset{Scale: protoWeight.Scale, Offset: protoWeight.Offset})
	return weight1
}

func buildGearWeight2(protoWeight *gearproto.Weight2) (*weight_types.Weight2, error) {
	weight2 := &weight_types.Weight2{}
	for _, ent := range protoWeight.GetWeights() {
		weight2.PutWeight(convertSimTypeReverse(ent.SimType), convertStatTypeReverse(ent.StatType), ent.Weight)
	}
	for _, pri := range protoWeight.GetPriority() {
		if err := weight2.SetSimScale(convertSimTypeReverse(pri.SimType), weight_types.ScaleAndOffset{Scale: pri.RangingScale, Offset: pri.RangingOffset}, pri.RatioScale); err != nil {
			return nil, err
		}
	}
	// TODO retain order specified
	weight2.SimList = slices.Collect(weight2.DetailedWeights.SeqKey1())
	weight2.StatList = slices.Collect(weight2.DetailedWeights.SeqKey2())
	if err := weight2.FinishAndValidateNoVerify(); err != nil {
		return nil, err
	}
	return weight2, nil
}

func convertWeight3Details(weight3 *weight_types.Weight3) []*gearproto.StatSimTypesAndNestedEntries {
	weights := make([]*gearproto.StatSimTypesAndNestedEntries, 0)
	for _, simType := range weight3.SimList {
		for _, statType := range weight3.StatList {
			weightEntry, hasEntry := weight3.StatWeights.GetAsSliceClone(simType, statType)
			if hasEntry {
				group := gearproto.StatSimTypesAndNestedEntries{}
				group.SimType = convertSimType(simType)
				group.StatType = convertStatType(statType)
				group.Entries = make([]*gearproto.StatRangeAndWeightOffset, 0)
				for _, entry := range weightEntry {
					group.Entries = append(group.Entries, &gearproto.StatRangeAndWeightOffset{
						StatMinimum: entry.StatRange.Minimum,
						StatMaximum: entry.StatRange.Maximum,
						Weight:      entry.RatingWeight,
						Offset:      entry.RatingOffset,
					})
				}
				weights = append(weights, &group)
			}
		}
	}
	return weights
}

func buildGearWeight3(protoWeight *gearproto.Weight3) (*weight_types.Weight3, error) {
	weight3 := &weight_types.Weight3{}
	for _, group := range protoWeight.GetWeights() {
		for _, ent := range group.Entries {
			weight3.AddDetailWeight(
				convertSimTypeReverse(group.SimType),
				convertStatTypeReverse(group.StatType),
				weight_types.StatRange{Minimum: ent.StatMinimum, Maximum: ent.StatMaximum},
				ent.Weight,
				ent.Offset,
				0,
			)
		}
	}
	for _, pri := range protoWeight.GetPriority() {
		if err := weight3.SetSimScale(convertSimTypeReverse(pri.SimType), weight_types.ScaleAndOffset{pri.RangingScale, pri.RangingOffset}, pri.RatioScale); err != nil {
			return nil, err
		}
	}
	weight3.StatList = slices.Collect(weight3.StatWeights.SeqKey2())
	weight3.SimList = slices.Collect(weight3.StatWeights.SeqKey1())
	if err := weight3.FinishAndValidateNoVerify(); err != nil {
		return nil, err
	}
	return weight3, nil
}

func convertPriority(simPriority *weight_types.SimPriorityExtended, simList []stats.SimType) []*gearproto.SimTypeAndScaleOffsetAndSimRatio {
	priority := make([]*gearproto.SimTypeAndScaleOffsetAndSimRatio, 0)
	for _, simType := range simList {
		prior := simPriority.GetOrPanic(simType)
		priority = append(priority, &gearproto.SimTypeAndScaleOffsetAndSimRatio{
			SimType:       convertSimType(simType),
			RangingScale:  prior.Ranging.Scale,
			RangingOffset: prior.Ranging.Offset,
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
