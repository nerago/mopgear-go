package tools

import (
	"paladin_gearing_go/gearproto"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_types"

	"google.golang.org/protobuf/encoding/protojson"
)

func WriteWeight3String(weight3 weight_types.Weight3ExtendedRanged, printer *util.PrintRecorder) string {
	protoWeight := gearproto.Weight3Extended{}

	protoWeight.Weights = make([]*gearproto.Weight3Group, 0)
	for weightEntry := range weight3.StatWeights.SeqKey1Key2ValueSeqEntries() {
		group := gearproto.Weight3Group{}
		group.SimType = gearproto.SimType(weightEntry.Key1)
		group.StatType = gearproto.StatType(weightEntry.Key2)
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
			SimType:       gearproto.SimType(simType),
			RangingScale:  prior.RangingScale,
			RangingOffset: prior.RangingOffset,
			RatioScale:    prior.RatioScale,
		})
	}

	str := protojson.Format(&protoWeight)
	printer.Println(str)
	return str
}
