package weight_types

import (
	"cmp"
	"iter"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
	"slices"
)

// Weight3ExtendedRanged
type Weight3ExtendedRanged struct {
	StatWeights    util.MapMapSlice[stats.SimType, stats.StatType, Weight3ExtendedStatEntry]
	SimMultipliers map[stats.SimType]Weight3ExtendedSimEntry
}

type Weight3ExtendedStatEntry struct {
	RatingWeight      float64
	RatingOffset      float64
	StatRange         StatRange
	TotalIfGreater    float64
	EstimationQuality float64 // higher is better
}

type Weight3ExtendedSimEntry struct {
	// calculated so that range of values is consistent (e.g. 0-100)
	// is offset needed, would give more real values for tmi/death, but would that change result
	Scale    float64
	Offset   float64
	Minimise bool
}

func Weight3ExtendedRanged_Make() Weight3ExtendedRanged {
	return Weight3ExtendedRanged{}
}

func (wer *Weight3ExtendedRanged) Add(simType stats.SimType, statType stats.StatType, statRange StatRange, ratingWeight, ratingOffset, estimationQuality float64) {
	wer.StatWeights.Add(simType, statType, Weight3ExtendedStatEntry{
		StatRange:         statRange,
		RatingWeight:      ratingWeight,
		RatingOffset:      ratingOffset,
		TotalIfGreater:    float64(statRange.RangeSize()) * ratingWeight,
		EstimationQuality: estimationQuality,
	})

	wer.StatWeights.MapInternalSlice(simType, statType, func(entries []Weight3ExtendedStatEntry) []Weight3ExtendedStatEntry {
		slices.SortStableFunc(entries, func(a, b Weight3ExtendedStatEntry) int {
			return cmp.Compare(a.StatRange.Minimum, b.StatRange.Minimum)
		})
		return entries
	})
}

func (wer *Weight3ExtendedRanged) Validate() {
	// no gaps etc
}

func (wer *Weight3ExtendedRanged) ConvertToWeight2() Weight2Extended {
	simRatio := wer.makeEquivalentSimRatio()

	weight2 := Weight2Extended_Make(simRatio)
	for entry := range wer.StatWeights.SeqGroupsKeysNestedValueSeq() {
		weight2.Put(entry.Key2, entry.Key1, chooseBest(entry.ValueSeq))
	}
	return weight2
}

func chooseBest(statEntrySeq iter.Seq[Weight3ExtendedStatEntry]) float64 {
	best := util_rank.BestCollector1[Weight3ExtendedStatEntry]{}
	for statEntry := range statEntrySeq {
		best.Offer(&statEntry, statEntry.EstimationQuality)
	}

	bestEntry := best.GetBestOrPanic()
	return bestEntry.RatingWeight
}

func (wer *Weight3ExtendedRanged) makeEquivalentSimRatio() stats.SimData {
	simRatio := stats.SimData{}
	for simType, entry := range wer.SimMultipliers {
		simRatio.Set(simType, entry.Scale)
	}
	simRatio = *simRatio.ScaleForTotalSum(1.0)
	return simRatio
}
