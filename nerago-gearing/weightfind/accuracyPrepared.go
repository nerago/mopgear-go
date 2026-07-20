package weightfind

import (
	"cmp"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/simrank"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
)

type EvaluateAccuracyPrepared struct {
	prepared       []*weight_types.AccuracyPreparedEntry
	statRankRanges []*util.HiLoInt
	hiLoPool       []util.HiLoInt
}

func (ea *EvaluateAccuracyPrepared) Init(inputData []weight_types.WeightInput, simRatios weight_types.SimPriorityBasic, simCalcMode int) {
	if simCalcMode == 0 {
		panic("mode not provided")
	}

	data := util.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *weight_types.AccuracyInfoPrePrepare {
		return &weight_types.AccuracyInfoPrePrepare{
			DataSim:  &input.SimResult,
			DataStat: &input.TotalStat,
			SimScore: 0,
		}
	})

	requiredSims := simRatios.SimTypes()
	if simCalcMode == 1 {
		simrank.RankSimsBasicForAccuracyPrepare(simRatios, data, requiredSims)
	} else {
		simrank.RankSimsStatisticalForAccuracyPrepare(simRatios, data, requiredSims)
	}

	ea.calcSimRankAndPrepare(data)

	ea.statRankRanges = make([]*util.HiLoInt, len(data))
	ea.hiLoPool = make([]util.HiLoInt, len(data))
}

func (ea *EvaluateAccuracyPrepared) calcSimRankAndPrepare(data []*weight_types.AccuracyInfoPrePrepare) {
	// make ranked entries for later, calculating the sim rank as we go
	prepare := make([]*weight_types.AccuracyPreparedEntry, len(data))
	simrank.CalcHiLoForAccuracyPrepare(data, prepare)

	ea.prepared = prepare
}

func (ea *EvaluateAccuracyPrepared) Clone() *EvaluateAccuracyPrepared {
	return &EvaluateAccuracyPrepared{
		prepared: util.MapSliceAsNew_NoPointer(ea.prepared, func(x *weight_types.AccuracyPreparedEntry) *weight_types.AccuracyPreparedEntry {
			return &weight_types.AccuracyPreparedEntry{SimRankRange: x.SimRankRange, Stats: x.Stats}
		}),
		statRankRanges: make([]*util.HiLoInt, len(ea.statRankRanges)),
		hiLoPool:       make([]util.HiLoInt, len(ea.hiLoPool)),
	}
}

// fundamentally not thread safe
func (ea *EvaluateAccuracyPrepared) EvaluateWeight(statWeights weight_types.Weight1Basic) float64 {
	if statWeights.IsEmpty() {
		return 0
	}

	prepared := ea.prepared
	size := len(prepared)

	// calculate stat scores for given weights
	for i := range size {
		prepared[i].StatScore = statWeights.CalcStatScore2(prepared[i].Stats)
	}
	slices.SortFunc(prepared, func(a, b *weight_types.AccuracyPreparedEntry) int {
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
