package weightfind

import (
	"cmp"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_highs"
	"slices"
)

type EvaluateAccuracyPrepared struct {
	prepared       []*accuracyPreparedEntry
	statRankRanges []*util.HiLoInt
	hiLoPool       []util.HiLoInt
}

type accuracyInfoPrePrepare struct {
	simScore float64
	dataSim  *stats.SimData
	dataStat *stats.StatBlock
}

type accuracyPreparedEntry struct {
	statScore    float64
	stats        *stats.StatBlock
	simRankRange *util.HiLoInt
}

func (ea *EvaluateAccuracyPrepared) Init(inputData []weight_highs.WeightInput, simRatios stats.SimData, simCalcMode int) {
	if simCalcMode == 0 {
		panic("mode not provided")
	}

	data := util.MapSliceAsNew(inputData, func(input *weight_highs.WeightInput) *accuracyInfoPrePrepare {
		return &accuracyInfoPrePrepare{
			dataSim:  &input.SimResult,
			dataStat: &input.TotalStat,
			simScore: 0,
		}
	})

	if simCalcMode == 1 {
		ea.scoreSimsBasic(simRatios, data)
	} else {
		ea.scoreSimsStatistical(simRatios, data)
	}

	ea.calcSimRankAndPrepare(data)

	ea.statRankRanges = make([]*util.HiLoInt, len(data))
	ea.hiLoPool = make([]util.HiLoInt, len(data))
}

func (ea *EvaluateAccuracyPrepared) scoreSimsBasic(simRatios stats.SimData, data []*accuracyInfoPrePrepare) {
	// score each sim
	requiredSims := simRatios.NonZeroTypes()
	for _, simType := range requiredSims {
		ratio := simRatios.Get(simType)
		slices.SortFunc(data, simSortSimSingledCompares[simType])
		for rank := range data {
			data[rank].simScore += float64(rank) * ratio
		}
	}
}

func (ea *EvaluateAccuracyPrepared) scoreSimsStatistical(simRatios stats.SimData, data []*accuracyInfoPrePrepare) {
	// score each sim
	requiredSims := simRatios.NonZeroTypes()
	for _, simType := range requiredSims {
		if simType == stats.Sim_DEATH {
			// death data never has detail
			slices.SortFunc(data, simSortSimSingledCompares[simType])
		} else if simType.IsHighGood() {
			slices.SortFunc(data, func(a, b *accuracyInfoPrePrepare) int {
				return deviationCompareSims(a.dataSim, b.dataSim, simType)
			})
		} else {
			slices.SortFunc(data, func(a, b *accuracyInfoPrePrepare) int {
				return -1 * deviationCompareSims(a.dataSim, b.dataSim, simType)
			})
		}
		ratio := simRatios.Get(simType)
		for rank := range data {
			entry := data[rank]
			entry.simScore += float64(rank) * ratio
		}
	}
}

func (ea *EvaluateAccuracyPrepared) calcSimRankAndPrepare(data []*accuracyInfoPrePrepare) {
	// rank combined sims
	slices.SortFunc(data, func(a, b *accuracyInfoPrePrepare) int {
		return cmp.Compare(a.simScore, b.simScore)
	})

	// make ranked entries for later, calculating the sim rank as we go
	prepare := make([]*accuracyPreparedEntry, len(data))
	prepare[0] = &accuracyPreparedEntry{
		simRankRange: &util.HiLoInt{Lo: 0, Hi: 0},
		stats:        data[0].dataStat,
	}
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].simScore, data[i-1].simScore) {
			prevRange := prepare[i-1].simRankRange
			prevRange.Hi = i
			prepare[i] = &accuracyPreparedEntry{simRankRange: prevRange, stats: data[i].dataStat}
		} else {
			newRange := &util.HiLoInt{Lo: i, Hi: i}
			prepare[i] = &accuracyPreparedEntry{simRankRange: newRange, stats: data[i].dataStat}
		}
	}

	ea.prepared = prepare
}

func (ea *EvaluateAccuracyPrepared) Clone() *EvaluateAccuracyPrepared {
	return &EvaluateAccuracyPrepared{
		prepared: util.MapSliceAsNew_NoPointer(ea.prepared, func(x *accuracyPreparedEntry) *accuracyPreparedEntry {
			return &accuracyPreparedEntry{simRankRange: x.simRankRange, stats: x.stats}
		}),
		statRankRanges: make([]*util.HiLoInt, len(ea.statRankRanges)),
		hiLoPool:       make([]util.HiLoInt, len(ea.hiLoPool)),
	}
}

// fundamentally not thread safe
func (ea *EvaluateAccuracyPrepared) EvaluateWeight(statWeights weight_highs.WeightResult) float64 {
	if statWeights.IsEmpty() {
		return 0
	}

	prepared := ea.prepared
	size := len(prepared)

	// calculate stat scores for given weights
	for i := range size {
		prepared[i].statScore = statWeights.CalcStatScore2(prepared[i].stats)
	}
	slices.SortFunc(prepared, func(a, b *accuracyPreparedEntry) int {
		return cmp.Compare(a.statScore, b.statScore)
	})

	// rank stats scores
	statRankRanges := ea.statRankRanges
	statRankRanges[0] = &ea.hiLoPool[0]
	statRankRanges[0].Lo = 0
	statRankRanges[0].Hi = 0
	hiLoAlloc := 1

	for rank := 1; rank < size; rank++ {
		if util.FloatsApproxEquals(prepared[rank].statScore, prepared[rank-1].statScore) {
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
		ratioScore := rangesToAccuracyRatio(*prepared[i].simRankRange, *statRankRanges[i], size)
		sumRatioScores += ratioScore
	}
	averagePercent := 100.0 * sumRatioScores / float64(size)
	return averagePercent
}

//goland:noinspection DuplicatedCode
var simSortSimSingledCompares = [6]func(a, b *accuracyInfoPrePrepare) int{
	func(a, b *accuracyInfoPrePrepare) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_DPS), b.dataSim.Get(stats.Sim_DPS))
	},
	func(a, b *accuracyInfoPrePrepare) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_TPS), b.dataSim.Get(stats.Sim_TPS))
	},
	func(a, b *accuracyInfoPrePrepare) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_DTPS), a.dataSim.Get(stats.Sim_DTPS))
	},
	func(a, b *accuracyInfoPrePrepare) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_HPS), b.dataSim.Get(stats.Sim_HPS))
	},
	func(a, b *accuracyInfoPrePrepare) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_TMI), a.dataSim.Get(stats.Sim_TMI))
	},
	func(a, b *accuracyInfoPrePrepare) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_DEATH), a.dataSim.Get(stats.Sim_DEATH))
	},
}
