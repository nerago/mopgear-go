package weightfind

import (
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/simrank"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type EvaluateAccuracyPrepared struct {
	prepared       []*weight_types.AccuracyInfoPrepared
	statRankRanges []*util_collection.HiLoInt
	hiLoPool       []util_collection.HiLoInt
}

func (ea *EvaluateAccuracyPrepared) Init(inputData []weight_types.WeightInput, simRatios *weight_types.SimPriorityBasic, simStatistical bool, simStatisticalExtended bool) {
	requiredSims := simRatios.SimTypes()

	if simStatisticalExtended {
		data := util_collection.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *weight_types.AccuracyInfoPrePrepareExtended {
			return &weight_types.AccuracyInfoPrePrepareExtended{
				AccuracyInfoPrePrepare: weight_types.AccuracyInfoPrePrepare{
					DataSim:  &input.SimResult,
					DataStat: &input.TotalStat,
					SimScore: 0,
				},
			}
		})
		ea.prepared = simrank.AccuracyPrepareRankSimsStatisticalExtended(requiredSims, simRatios, data)
	} else {
		data := util_collection.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *weight_types.AccuracyInfoPrePrepare {
			return &weight_types.AccuracyInfoPrePrepare{
				DataSim:  &input.SimResult,
				DataStat: &input.TotalStat,
				SimScore: 0,
			}
		})
		if simStatistical {
			ea.prepared = simrank.AccuracyPrepareRankSimsStatistical(requiredSims, simRatios, data)
		} else {
			ea.prepared = simrank.AccuracyPrepareRankSimsBasic(requiredSims, simRatios, data)
		}
	}

	ea.statRankRanges = make([]*util_collection.HiLoInt, len(ea.prepared))
	ea.hiLoPool = make([]util_collection.HiLoInt, len(ea.prepared))
}

func (ea *EvaluateAccuracyPrepared) Clone() *EvaluateAccuracyPrepared {
	return &EvaluateAccuracyPrepared{
		prepared: util_collection.MapSliceAsNew_NoPointer(ea.prepared, func(x *weight_types.AccuracyInfoPrepared) *weight_types.AccuracyInfoPrepared {
			return &weight_types.AccuracyInfoPrepared{SimRankRange: x.SimRankRange, Stats: x.Stats}
		}),
		statRankRanges: make([]*util_collection.HiLoInt, len(ea.statRankRanges)),
		hiLoPool:       make([]util_collection.HiLoInt, len(ea.hiLoPool)),
	}
}

// fundamentally not thread safe
func (ea *EvaluateAccuracyPrepared) EvaluateWeight1(statWeights *weight_types.Weight1Basic) float64 {
	return evaluateWeightGeneral(ea, statWeights)
}
func (ea *EvaluateAccuracyPrepared) EvaluateWeight2(statWeights *weight_types.Weight2Extended) float64 {
	return evaluateWeightGeneral(ea, statWeights)
}
func (ea *EvaluateAccuracyPrepared) EvaluateWeight3(statWeights *weight_types.Weight3ExtendedRanged) float64 {
	return evaluateWeightGeneral(ea, statWeights)
}

func evaluateWeightGeneral[W weight_types.IWeight](ea *EvaluateAccuracyPrepared, statWeights W) float64 {
	if statWeights.IsEmpty() {
		return 0
	}

	prepared := ea.prepared
	size := len(prepared)

	// calculate stat scores for given weights
	for i := range size {
		prepared[i].StatScore = statWeights.CalcStatScore(prepared[i].Stats)
	}
	//slices.SortFunc(prepared, func(a, b *weight_types.AccuracyInfoPrepared) int {
	//	return cmp.Compare(a.StatScore, b.StatScore)
	//})

	// rank stats scores
	//statRankRanges := ea.statRankRanges
	//statRankRanges[0] = &ea.hiLoPool[0]
	//statRankRanges[0].Lo = 0
	//statRankRanges[0].Hi = 0
	//hiLoAlloc := 1
	//
	//for rank := 1; rank < size; rank++ {
	//	if util.FloatsApproxEquals(prepared[rank].StatScore, prepared[rank-1].StatScore) {
	//		prevRange := statRankRanges[rank-1]
	//		statRankRanges[rank] = prevRange
	//		prevRange.Hi = rank
	//	} else {
	//		newRange := &ea.hiLoPool[hiLoAlloc]
	//		newRange.Lo = rank
	//		newRange.Hi = rank
	//		statRankRanges[rank] = newRange
	//		hiLoAlloc++
	//	}
	//}

	deriveStatRanks(prepared)
	return calcAverageDifference(prepared)
}
