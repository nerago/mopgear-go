package weightfind

import (
	"cmp"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/weightfind/simrank"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
)

type EvaluateAccuracyPrepared struct {
	prepared       []*weight_types.AccuracyInfoPrepared
	statRankRanges []*util_collection.HiLoInt
	hiLoPool       []util_collection.HiLoInt
}

func (ea *EvaluateAccuracyPrepared) Init(inputData []weight_types.WeightInput, simRatios *weight_types.SimPriorityBasic, simStatistical bool) {
	data := util_collection.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *weight_types.AccuracyInfoPrePrepare {
		return &weight_types.AccuracyInfoPrePrepare{
			DataSim:  &input.SimResult,
			DataStat: &input.TotalStat,
			SimScore: 0,
		}
	})

	requiredSims := simRatios.SimTypes()
	if !simStatistical {
		ea.prepared = simrank.AccuracyPrepareRankSimsBasic(requiredSims, simRatios, data)
	} else {
		ea.prepared = simrank.RankSimsStatisticalForAccuracyPrepare(requiredSims, simRatios, data)
	}

	ea.statRankRanges = make([]*util_collection.HiLoInt, len(data))
	ea.hiLoPool = make([]util_collection.HiLoInt, len(data))
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
	slices.SortFunc(prepared, func(a, b *weight_types.AccuracyInfoPrepared) int {
		return cmp.Compare(a.StatScore, b.StatScore)
	})

	// rank stats scores
	statRankRanges := ea.statRankRanges
	statRankRanges[0] = &ea.hiLoPool[0]
	statRankRanges[0].Lo = 0
	statRankRanges[0].Hi = 0
	hiLoAlloc := 1

	for rank := 1; rank < size; rank++ {
		if util.FloatsApproxEquals(prepared[rank].StatScore, prepared[rank-1].StatScore) {
			prevRange := statRankRanges[rank-1]
			statRankRanges[rank] = prevRange
			prevRange.Hi = rank
		} else {
			newRange := &ea.hiLoPool[hiLoAlloc]
			newRange.Lo = rank
			newRange.Hi = rank
			statRankRanges[rank] = newRange
			hiLoAlloc++
		}
	}

	// compute average difference between stat rank and sim rank.
	sumRatioScores := 0.0
	for i := range size {
		ratioScore := rangesToAccuracyRatio(*prepared[i].SimRankRange, *statRankRanges[i], size)
		sumRatioScores += ratioScore
	}
	averagePercent := 100.0 * sumRatioScores / float64(size)
	return checkValue(averagePercent)
}
