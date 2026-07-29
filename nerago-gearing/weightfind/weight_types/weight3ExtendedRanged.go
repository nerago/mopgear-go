package weight_types

import (
	"cmp"
	"iter"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_rank"
	"slices"
)

// Weight3ExtendedRanged
type Weight3ExtendedRanged struct {
	StatWeights util_collection.MapMapSlice[stats.SimType, stats.StatType, Weight3ExtendedStatEntry]
	StatList    []stats.StatType
	SimList     []stats.SimType
	SimPriority SimPriorityExtended
}

type Weight3ExtendedStatEntry struct {
	RatingWeight      float64
	RatingOffset      float64
	StatRange         StatRange
	EstimationQuality float64 // higher is better
}

func Weight3ExtendedRanged_Make(statList []stats.StatType, simList []stats.SimType) *Weight3ExtendedRanged {
	return &Weight3ExtendedRanged{
		StatList:    statList,
		SimList:     simList,
		SimPriority: SimPriorityExtended_Make(),
	}
}

func (wer *Weight3ExtendedRanged) AddDetailWeight(simType stats.SimType, statType stats.StatType, statRange StatRange, ratingWeight, ratingOffset, estimationQuality float64) {
	wer.StatWeights.Add(simType, statType, Weight3ExtendedStatEntry{
		StatRange:         statRange,
		RatingWeight:      ratingWeight,
		RatingOffset:      ratingOffset,
		EstimationQuality: estimationQuality,
	})

	wer.StatWeights.MapInternalSliceOrPanic(simType, statType, func(entries []Weight3ExtendedStatEntry) []Weight3ExtendedStatEntry {
		slices.SortStableFunc(entries, func(a, b Weight3ExtendedStatEntry) int {
			return cmp.Compare(a.StatRange.Minimum, b.StatRange.Minimum)
		})
		return entries
	})
}

func (wer *Weight3ExtendedRanged) AddSimScale(simType stats.SimType, rangingScale, rangingOffset, ratioScale float64) {
	wer.SimPriority.SetSimScale(simType, rangingScale, rangingOffset, ratioScale)
}

func (wer *Weight3ExtendedRanged) FinishAndValidate() {
	for statType := range wer.StatWeights.SeqKey2() {
		if !slices.Contains(wer.StatList, statType) {
			panic("weight given for unlisted stat")
		}
	}
	for simType := range wer.StatWeights.SeqKey1() {
		if !slices.Contains(wer.SimList, simType) {
			panic("weight given for unlisted sim")
		}
	}

	for _, simType := range wer.SimList {
		for _, statType := range wer.StatList {
			if !wer.StatWeights.Has(simType, statType) {
				panic("missing weight for " + statType.Name() + " " + simType.Name())
			}
		}
	}

	for simType := range wer.SimPriority.entries.SeqKey() {
		if !slices.Contains(wer.SimList, simType) {
			panic("priority given for unlisted sim")
		}
	}
	for _, simType := range wer.SimList {
		_, hasValue := wer.SimPriority.Get(simType)
		if !hasValue {
			panic("priority missing for " + simType.Name())
		}
	}
}

func (wer *Weight3ExtendedRanged) ConvertToWeight2() *Weight2Extended {
	weight2 := Weight2Extended_Make(wer.StatList, wer.SimList)
	for entry := range wer.StatWeights.SeqKey1Key2ValueSeqEntries() {
		bestValue := chooseBest(entry.ValueSeq)
		weight2.PutWeight(entry.Key2, entry.Key1, bestValue)
	}
	weight2.SimPriority = wer.SimPriority.Clone()
	weight2.FinishAndValidate()
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
