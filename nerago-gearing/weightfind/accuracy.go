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
	return EvaluateAccuracyRanged0(statWeights, simRatios.NonZeroTypes(), simRatios, inputData)
}

func EvaluateAccuracyRanged0(statWeights stathighs.WeightResult, requiredSims []stats.SimType, simRatios stats.SimData, inputData []stathighs.WeightInput) float64 {
	if statWeights.IsEmpty() {
		return 0
	}
	// make structures
	type accuracyInfo struct {
		input *stathighs.WeightInput

		combinedSimRankScore float64

		statRankRange util.HiLoInt
		simRankRange  util.HiLoInt
	}
	accuracyData := util.MapSliceAsNew(inputData, func(input *stathighs.WeightInput) accuracyInfo {
		return accuracyInfo{
			input: input,
		}
	})

	// score stats
	for entry, statRank := range util.CalculateRankingRanges(true, accuracyData, func(x *accuracyInfo) float64 { return statWeights.CalcStatScore(x.input) }) {
		entry.statRankRange = statRank
	}

	// score each sim
	for _, simType := range requiredSims {
		for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), accuracyData, func(x *accuracyInfo) float64 { return x.input.SimResult.Get(simType) }) {
			entry.combinedSimRankScore += float64(simDetailRank) * simRatios.Get(simType)
		}
	}

	// rank combined sims
	for entry, simRank := range util.CalculateRankingRanges(true, accuracyData, func(x *accuracyInfo) float64 { return x.combinedSimRankScore }) {
		entry.simRankRange = simRank
	}

	// compute average difference between stat rank and sim rank
	totalComparePercents := 0.0
	for info := range util.ForPointer(accuracyData) {
		percentScore := rangesToAccuracyRatio(info.simRankRange, info.statRankRange, len(accuracyData))
		totalComparePercents += percentScore
	}
	return totalComparePercents / float64(len(accuracyData))
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

type accuracyInfoSimOnlySingles struct {
	simScore float64
	dataSim  *stats.SimData
	dataStat *stats.StatBlock
}

type accuracyInfoSimStatSingles struct {
	statScore float64
	simScore  float64
	statRank  int
	dataSim   *stats.SimData
}

type accuracyInfoSimStatRanged struct {
	statScore     float64
	simScore      float64
	statRankRange *util.HiLoInt
	simRankRange  *util.HiLoInt
	dataSim       *stats.SimData
}

func EvaluateAccuracyNoRange(statWeights stathighs.WeightResult, requiredSims []stats.SimType, simRatios stats.SimData, inputData []stathighs.WeightInput) float64 {
	accuracyData := util.MapSliceAsNew(inputData, func(input *stathighs.WeightInput) *accuracyInfoSimStatSingles {
		return &accuracyInfoSimStatSingles{
			dataSim:   &input.SimResult,
			statScore: statWeights.CalcStatScore(input),
			simScore:  0,
			statRank:  0,
		}
	})

	// rank stats scores
	slices.SortFunc(accuracyData, func(a, b *accuracyInfoSimStatSingles) int {
		return cmp.Compare(a.statScore, b.statScore)
	})
	for rank := range accuracyData {
		accuracyData[rank].statRank = rank
	}

	// score each sim
	for _, simType := range requiredSims {
		ratio := simRatios.Get(simType)
		slices.SortFunc(accuracyData, simSortSingleCompares[simType])
		for rank := range accuracyData {
			accuracyData[rank].simScore += float64(rank) * ratio
		}
	}

	// rank combined sims
	slices.SortFunc(accuracyData, func(a, b *accuracyInfoSimStatSingles) int {
		return cmp.Compare(a.simScore, b.simScore)
	})

	// compute average difference between stat rank and sim rank.
	// sim rank is just the current order.
	size := len(accuracyData)
	totalComparePercents := 0.0
	for simRank := range accuracyData {
		entry := accuracyData[simRank]
		diff := util.AbsIntDiff(simRank, entry.statRank)
		diffAsRatio := float64(size-diff) / float64(size)
		totalComparePercents += diffAsRatio
	}
	averagePercent := 100.0 * totalComparePercents / float64(size)
	return averagePercent
}

func EvaluateAccuracyWithRange2(statWeights stathighs.WeightResult, requiredSims []stats.SimType, simRatios stats.SimData, inputData []stathighs.WeightInput) float64 {
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
	totalComparePercents := 0.0
	for i := range data {
		entry := data[i]
		percentScore := rangesToAccuracyRatio(*entry.simRankRange, *entry.statRankRange, len(data))
		totalComparePercents += percentScore
	}
	return 100.0 * totalComparePercents / float64(len(data))
}

func EvaluateAccuracyWithRangePartialRefactor3(statWeights stathighs.WeightResult, requiredSims []stats.SimType, simRatios stats.SimData, inputData []stathighs.WeightInput) float64 {
	data := util.MapSliceAsNew(inputData, func(input *stathighs.WeightInput) *accuracyInfoSimStatRanged {
		return &accuracyInfoSimStatRanged{
			dataSim:   &input.SimResult,
			statScore: statWeights.CalcStatScore(input),
		}
	})

	// rank stats scores
	rankDataRanged(data,
		func(a *accuracyInfoSimStatRanged) float64 { return a.statScore },
		func(a *accuracyInfoSimStatRanged, hiLo *util.HiLoInt) { a.statRankRange = hiLo },
	)

	// score each sim
	for _, simType := range requiredSims {
		slices.SortFunc(data, simSortRangedCompares[simType])
		ratio := simRatios.Get(simType)
		for rank := range data {
			data[rank].simScore += float64(rank) * ratio
		}
	}

	// rank combined sims
	rankDataRanged(data,
		func(a *accuracyInfoSimStatRanged) float64 { return a.simScore },
		func(a *accuracyInfoSimStatRanged, hiLo *util.HiLoInt) { a.simRankRange = hiLo },
	)

	// compute average difference between stat rank and sim rank
	totalComparePercents := 0.0
	for i := range data {
		entry := data[i]
		percentScore := rangesToAccuracyRatio(*entry.simRankRange, *entry.statRankRange, len(data))
		totalComparePercents += percentScore
	}
	return 100.0 * totalComparePercents / float64(len(data))
}

func rankDataRanged[T any](data []*T, toScore func(*T) float64, setRange func(*T, *util.HiLoInt)) {
	sortDataRangedGeneric(data, toScore)

	prevScore := toScore(data[0])
	prevHiLo := &util.HiLoInt{Lo: 0, Hi: 0}
	setRange(data[0], prevHiLo)

	for index := 1; index < len(data); index++ {
		if util.FloatsApproxEquals(toScore(data[index]), prevScore) {
			prevHiLo.Hi = index
		} else {
			prevHiLo = &util.HiLoInt{Lo: index, Hi: index}
		}
		setRange(data[index], prevHiLo)
	}
}

func sortDataRangedGeneric[T any](data []*T, toScore func(*T) float64) {
	slices.SortFunc(data, func(a, b *T) int {
		return cmp.Compare(toScore(a), toScore(b))
	})
}

var simSortSingleCompares = [6]func(a, b *accuracyInfoSimStatSingles) int{
	func(a, b *accuracyInfoSimStatSingles) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_DPS), b.dataSim.Get(stats.Sim_DPS))
	},
	func(a, b *accuracyInfoSimStatSingles) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_TPS), b.dataSim.Get(stats.Sim_TPS))
	},
	func(a, b *accuracyInfoSimStatSingles) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_DTPS), a.dataSim.Get(stats.Sim_DTPS))
	},
	func(a, b *accuracyInfoSimStatSingles) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_HPS), b.dataSim.Get(stats.Sim_HPS))
	},
	func(a, b *accuracyInfoSimStatSingles) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_TMI), a.dataSim.Get(stats.Sim_TMI))
	},
	func(a, b *accuracyInfoSimStatSingles) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_DEATH), b.dataSim.Get(stats.Sim_DEATH))
	},
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
		return cmp.Compare(b.dataSim.Get(stats.Sim_DEATH), b.dataSim.Get(stats.Sim_DEATH))
	},
}
var simSortSimSingledCompares = [6]func(a, b *accuracyInfoSimOnlySingles) int{
	func(a, b *accuracyInfoSimOnlySingles) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_DPS), b.dataSim.Get(stats.Sim_DPS))
	},
	func(a, b *accuracyInfoSimOnlySingles) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_TPS), b.dataSim.Get(stats.Sim_TPS))
	},
	func(a, b *accuracyInfoSimOnlySingles) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_DTPS), a.dataSim.Get(stats.Sim_DTPS))
	},
	func(a, b *accuracyInfoSimOnlySingles) int {
		return cmp.Compare(a.dataSim.Get(stats.Sim_HPS), b.dataSim.Get(stats.Sim_HPS))
	},
	func(a, b *accuracyInfoSimOnlySingles) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_TMI), a.dataSim.Get(stats.Sim_TMI))
	},
	func(a, b *accuracyInfoSimOnlySingles) int {
		return cmp.Compare(b.dataSim.Get(stats.Sim_DEATH), b.dataSim.Get(stats.Sim_DEATH))
	},
}

func EvaluateAccuracyFullRangeInlined(statWeights stathighs.WeightResult, requiredSims []stats.SimType, simRatios stats.SimData, inputData []stathighs.WeightInput) float64 {
	type accuracyInfo struct {
		statScore float64
		simScore  float64
		statRank  *util.HiLoInt
		simRank   *util.HiLoInt
		input     *stathighs.WeightInput
	}
	accuracyData := make([]*accuracyInfo, len(inputData))
	for i := range len(inputData) {
		input := &inputData[i]
		accuracyData[i] = &accuracyInfo{
			input:     input,
			statScore: statWeights.CalcStatScore(input),
		}
	}

	// rank stats scores
	slices.SortFunc(accuracyData, func(a, b *accuracyInfo) int { return cmp.Compare(a.statScore, b.statScore) })
	{
		prevScore := accuracyData[0].statScore
		prevHiLo := &util.HiLoInt{Lo: 0, Hi: 0}
		for index := range accuracyData {
			entry := accuracyData[index]
			// TODO FloatsApproxEqualsFast
			if util.FloatsApproxEquals(entry.statScore, prevScore) {
				prevHiLo.Hi = index
				entry.statRank = prevHiLo
			} else {
				prevHiLo = &util.HiLoInt{Lo: index, Hi: index}
				entry.statRank = prevHiLo
			}
			prevScore = entry.statScore
		}
	}

	// score each sim
	for _, simType := range requiredSims {
		// TODO use sortSimType
		if simType.IsHighGood() {
			slices.SortFunc(accuracyData, func(a, b *accuracyInfo) int {
				return cmp.Compare(a.input.SimResult.Get(simType), b.input.SimResult.Get(simType))
			})
		} else {
			slices.SortFunc(accuracyData, func(a, b *accuracyInfo) int {
				return cmp.Compare(b.input.SimResult.Get(simType), a.input.SimResult.Get(simType))
			})
		}

		ratio := simRatios.Get(simType)
		for rank := range accuracyData {
			entry := accuracyData[rank]
			entry.simScore += float64(rank) * ratio
		}
	}

	// rank combined sims
	slices.SortFunc(accuracyData, func(a, b *accuracyInfo) int { return cmp.Compare(a.simScore, b.simScore) })
	{
		prevScore := accuracyData[0].simScore
		prevHiLo := &util.HiLoInt{Lo: 0, Hi: 0}
		for index := range accuracyData {
			entry := accuracyData[index]
			// TODO FloatsApproxEqualsFast
			if util.FloatsApproxEquals(entry.simScore, prevScore) {
				prevHiLo.Hi = index
				entry.simRank = prevHiLo
			} else {
				prevHiLo = &util.HiLoInt{Lo: index, Hi: index}
				entry.simRank = prevHiLo
			}
			prevScore = entry.simScore
		}
	}

	// compute average difference between stat rank and sim rank
	totalComparePercents := 0.0
	for i := range accuracyData {
		entry := accuracyData[i]
		percentScore := rangesToAccuracyRatio(*entry.simRank, *entry.statRank, len(accuracyData))
		totalComparePercents += percentScore
	}

	return 100.0 * totalComparePercents / float64(len(accuracyData))
}

type evaluatePreparedEntry struct {
	statScore  float64
	stats      *stats.StatBlock
	targetRank int
}

type EvaluateAccuracyPrepared struct {
	entries []*evaluatePreparedEntry
}

func (ea *EvaluateAccuracyPrepared) Init(inputData []stathighs.WeightInput, simRatios stats.SimData) {
	data := util.MapSliceAsNew(inputData, func(input *stathighs.WeightInput) *accuracyInfoSimOnlySingles {
		return &accuracyInfoSimOnlySingles{
			dataSim:  &input.SimResult,
			dataStat: &input.TotalStat,
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
	slices.SortFunc(data, func(a, b *accuracyInfoSimOnlySingles) int {
		return cmp.Compare(a.simScore, b.simScore)
	})

	// make ranked entries for later, recording the sim rank
	prepare := make([]*evaluatePreparedEntry, len(data))
	for simRank := range data {
		entry := data[simRank]
		prepare[simRank] = &evaluatePreparedEntry{targetRank: simRank, stats: entry.dataStat}
	}
	ea.entries = prepare
}

func (ea *EvaluateAccuracyPrepared) Clone() *EvaluateAccuracyPrepared {
	return &EvaluateAccuracyPrepared{util.MapSliceAsNew_NoPointer(ea.entries, func(x *evaluatePreparedEntry) *evaluatePreparedEntry {
		return &evaluatePreparedEntry{targetRank: x.targetRank, stats: x.stats}
	})}
}

// fundamentally not thread safe
func (ea *EvaluateAccuracyPrepared) EvaluateWeight(statWeights stathighs.WeightResult) float64 {
	entries := ea.entries

	// calculate stat scores for given weights
	for i := range len(entries) {
		entries[i].statScore = statWeights.CalcStatScore2(entries[i].stats)
	}

	// rank stats scores
	slices.SortFunc(entries, func(a, b *evaluatePreparedEntry) int {
		return cmp.Compare(a.statScore, b.statScore)
	})

	// compute average difference between stat rank and sim rank.
	// stat rank is just the current order.
	size := len(entries)
	totalComparePercents := 0.0
	for statRank := range entries {
		entry := entries[statRank]
		diff := util.AbsIntDiff(entry.targetRank, statRank)
		diffAsRatio := float64(size-diff) / float64(size)
		totalComparePercents += diffAsRatio
	}
	averagePercent := 100.0 * totalComparePercents / float64(size)
	return averagePercent
}
