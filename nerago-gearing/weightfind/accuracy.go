package weightfind

import (
	"cmp"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
)

// TODO take into account sim's uncertainty ranges

func EvaluateAccuracyRanged(statWeights stathighs.WeightResult, simRatios stats.SimData, inputData []stathighs.WeightInput) float64 {
	return EvaluateAccuracyRangeInner(statWeights, simRatios.NonZeroTypes(), simRatios, inputData)
}

func EvaluateAccuracyRangeInner(statWeights stathighs.WeightResult, requiredSims []stats.SimType, simRatios stats.SimData, inputData []stathighs.WeightInput) float64 {
	if statWeights.IsEmpty() {
		return 0
	}

	data := util.MapSliceAsNew(inputData, func(input *stathighs.WeightInput) *accuracyInfoSimStatRanged {
		return &accuracyInfoSimStatRanged{
			dataSim:       &input.SimResult,
			statScore:     statWeights.CalcStatScore(input),
			simScore:      0,
			statRankRange: nil,
			simRankRange:  nil,
		}
	})

	// rank stats scores
	slices.SortFunc(data, func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(a.statScore, b.statScore)
	})
	data[0].statRankRange = &util.HiLoInt{Lo: 0, Hi: 0}
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].statScore, data[i-1].statScore) {
			prevRange := data[i-1].statRankRange
			data[i].statRankRange = prevRange
			prevRange.Hi = i
		} else {
			data[i].statRankRange = &util.HiLoInt{Lo: i, Hi: i}
		}
	}

	// score each sim
	for _, simType := range requiredSims {
		slices.SortFunc(data, simSortRangedCompares[simType])
		ratio := simRatios.Get(simType)
		for rank := range data {
			entry := data[rank]
			entry.simScore += float64(rank) * ratio
		}
	}

	// rank combined sims
	slices.SortFunc(data, func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(a.simScore, b.simScore)
	})
	data[0].simRankRange = &util.HiLoInt{Lo: 0, Hi: 0}
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].simScore, data[i-1].simScore) {
			prevRange := data[i-1].simRankRange
			data[i].simRankRange = prevRange
			prevRange.Hi = i
		} else {
			data[i].simRankRange = &util.HiLoInt{Lo: i, Hi: i}
		}
	}

	// compute average difference between stat rank and sim rank
	sumRatioScores := 0.0
	for i := range data {
		entry := data[i]
		ratioScore := rangesToAccuracyRatio(*entry.simRankRange, *entry.statRankRange, len(data))
		sumRatioScores += ratioScore
	}
	return 100.0 * sumRatioScores / float64(len(data))
}

// 100% if ranks are equal, 90% if average 10% difference, etc
func rangesToAccuracyRatio(one, two util.HiLoInt, fullLength int) float64 {
	var diff int
	if one.Overlap(two) {
		return 1.0
	} else if one.Hi < two.Lo {
		diff = two.Lo - one.Hi
	} else if two.Hi < one.Lo {
		diff = one.Lo - two.Hi
	} else {
		panic("logic issue")
	}

	return float64(fullLength-diff) / float64(fullLength)
}

type accuracyInfoSimOnly struct {
	simScore float64
	dataSim  *stats.SimData
	dataStat *stats.StatBlock
}

type accuracyInfoSimStatRanged struct {
	statScore     float64
	simScore      float64
	statRankRange *util.HiLoInt
	simRankRange  *util.HiLoInt
	dataSim       *stats.SimData
}

var simSortRangedCompares = [6]func(a, b *accuracyInfoSimStatRanged) int{
	func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_DPS), b.dataSim.Get(stats.Sim_DPS))
	},
	func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_TPS), b.dataSim.Get(stats.Sim_TPS))
	},
	func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_DTPS), a.dataSim.Get(stats.Sim_DTPS))
	},
	func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_HPS), b.dataSim.Get(stats.Sim_HPS))
	},
	func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_TMI), a.dataSim.Get(stats.Sim_TMI))
	},
	func(a, b *accuracyInfoSimStatRanged) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_DEATH), a.dataSim.Get(stats.Sim_DEATH))
	},
}
var simSortSimSingledCompares = [6]func(a, b *accuracyInfoSimOnly) int{
	func(a, b *accuracyInfoSimOnly) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_DPS), b.dataSim.Get(stats.Sim_DPS))
	},
	func(a, b *accuracyInfoSimOnly) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_TPS), b.dataSim.Get(stats.Sim_TPS))
	},
	func(a, b *accuracyInfoSimOnly) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_DTPS), a.dataSim.Get(stats.Sim_DTPS))
	},
	func(a, b *accuracyInfoSimOnly) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_HPS), b.dataSim.Get(stats.Sim_HPS))
	},
	func(a, b *accuracyInfoSimOnly) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_TMI), a.dataSim.Get(stats.Sim_TMI))
	},
	func(a, b *accuracyInfoSimOnly) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_DEATH), a.dataSim.Get(stats.Sim_DEATH))
	},
}

type evaluatePreparedEntry struct {
	statScore    float64
	stats        *stats.StatBlock
	simRankRange *util.HiLoInt
}

type EvaluateAccuracyPrepared struct {
	prepared       []*evaluatePreparedEntry
	statRankRanges []*util.HiLoInt
}

func (ea *EvaluateAccuracyPrepared) Init(inputData []stathighs.WeightInput, simRatios stats.SimData) {
	data := util.MapSliceAsNew(inputData, func(input *stathighs.WeightInput) *accuracyInfoSimOnly {
		return &accuracyInfoSimOnly{
			dataSim:  &input.SimResult,
			dataStat: &input.TotalStat,
			simScore: 0,
		}
	})

	// score each sim
	requiredSims := simRatios.NonZeroTypes()
	for _, simType := range requiredSims {
		ratio := simRatios.Get(simType)
		slices.SortFunc(data, simSortSimSingledCompares[simType])
		for rank := range data {
			data[rank].simScore += float64(rank) * ratio
		}
	}

	// rank combined sims
	slices.SortFunc(data, func(a, b *accuracyInfoSimOnly) int {
		return cmp.Compare(a.simScore, b.simScore)
	})

	// make ranked entries for later, calculating the sim rank as we go
	prepare := make([]*evaluatePreparedEntry, len(data))
	prepare[0] = &evaluatePreparedEntry{
		simRankRange: &util.HiLoInt{Lo: 0, Hi: 0},
		stats:        data[0].dataStat,
	}
	for i := 1; i < len(data); i++ {
		if util.FloatsApproxEquals(data[i].simScore, data[i-1].simScore) {
			prevRange := prepare[i-1].simRankRange
			prevRange.Hi = i
			prepare[i] = &evaluatePreparedEntry{simRankRange: prevRange, stats: data[i].dataStat}
		} else {
			newRange := &util.HiLoInt{Lo: i, Hi: i}
			prepare[i] = &evaluatePreparedEntry{simRankRange: newRange, stats: data[i].dataStat}
		}
	}

	ea.prepared = prepare
	ea.statRankRanges = make([]*util.HiLoInt, len(data))
}

func (ea *EvaluateAccuracyPrepared) Clone() *EvaluateAccuracyPrepared {
	return &EvaluateAccuracyPrepared{
		prepared: util.MapSliceAsNew_NoPointer(ea.prepared, func(x *evaluatePreparedEntry) *evaluatePreparedEntry {
			return &evaluatePreparedEntry{simRankRange: x.simRankRange, stats: x.stats}
		}),
		statRankRanges: make([]*util.HiLoInt, len(ea.statRankRanges)),
	}
}

// fundamentally not thread safe
func (ea *EvaluateAccuracyPrepared) EvaluateWeight(statWeights stathighs.WeightResult) float64 {
	if statWeights.IsEmpty() {
		return 0
	}

	prepared := ea.prepared
	size := len(prepared)

	// calculate stat scores for given weights
	for i := range size {
		prepared[i].statScore = statWeights.CalcStatScore2(prepared[i].stats)
	}

	// rank stats scores
	// TODO could have a pool of HiLo's in a flat buffer
	slices.SortFunc(prepared, func(a, b *evaluatePreparedEntry) int {
		return cmp.Compare(a.statScore, b.statScore)
	})
	statRankRanges := ea.statRankRanges
	statRankRanges[0] = &util.HiLoInt{Lo: 0, Hi: 0}
	for i := 1; i < size; i++ {
		if util.FloatsApproxEquals(prepared[i].statScore, prepared[i-1].statScore) {
			prevRange := statRankRanges[i-1]
			statRankRanges[i] = prevRange
			prevRange.Hi = i
		} else {
			statRankRanges[i] = &util.HiLoInt{Lo: i, Hi: i}
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
