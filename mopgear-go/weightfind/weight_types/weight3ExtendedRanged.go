package weight_types

import (
	"cmp"
	"iter"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_rank"
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

func (ese *Weight3ExtendedStatEntry) Equals(other *Weight3ExtendedStatEntry) bool {
	return ese.RatingWeight == other.RatingWeight &&
		ese.RatingOffset == other.RatingOffset &&
		ese.StatRange == other.StatRange
}

func Weight3ExtendedRanged_Make(statList []stats.StatType, simList []stats.SimType) *Weight3ExtendedRanged {
	return &Weight3ExtendedRanged{
		StatList:    statList,
		SimList:     simList,
		SimPriority: SimPriorityExtended_Make(),
	}
}

func (wer *Weight3ExtendedRanged) IsEmpty() bool {
	if wer.SimPriority.IsEmpty() {
		return true
	}
	for entry := range wer.StatWeights.SeqValues() {
		if stats.IsUsefulWeightNumber(entry.RatingWeight) && stats.IsUsefulWeightNumber(entry.RatingOffset) {
			return false
		}
	}
	return true
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

func (wer *Weight3ExtendedRanged) SetSimScale(simType stats.SimType, rangingScale, rangingOffset, ratioScale float64) {
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
	wer.StatWeights.Foreach(func(simType stats.SimType, statType stats.StatType, entry Weight3ExtendedStatEntry) {
		if !slices.Contains(wer.StatList, statType) {
			panic("unexpected weight for " + statType.Name())
		}
		if !slices.Contains(wer.SimList, simType) {
			panic("unexpected weight for " + simType.Name())
		}
		if entry.StatRange.RangeSize() <= 1 {
			panic("empty range")
		}
	})

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

func (wer *Weight3ExtendedRanged) CalcStatScore(stats *stats.StatBlock) float64 {
	totalSum := 0.0
	for _, simType := range wer.SimList {
		subTotal := wer.calcSingleSimScoreUnscaled(stats, simType)

		priorityEntry := wer.SimPriority.GetOrPanic(simType)
		subTotal = priorityEntry.Apply(subTotal)

		totalSum += subTotal
	}
	return totalSum
}

func (wer *Weight3ExtendedRanged) CalcStatScoreWithBonus(stats *stats.StatBlock, simBonus *stats.SimTypeMap[float64]) float64 {
	totalSum := 0.0
	for _, simType := range wer.SimList {
		subTotal := wer.calcSingleSimScoreUnscaled(stats, simType)

		priorityEntry := wer.SimPriority.GetOrPanic(simType)
		subTotal = priorityEntry.Apply(subTotal)

		subTotal *= simBonus.GetOrDefault(simType, 1)

		totalSum += subTotal
	}
	return totalSum
}

func (wer *Weight3ExtendedRanged) calcSingleSimScoreUnscaled(stats *stats.StatBlock, simType stats.SimType) float64 {
	simSubTotal := 0.0

	for statType, entrySeq := range wer.StatWeights.SeqKey2ValueSeqWithKey1(simType) {
		statValue := stats.GetUInt(statType)

		var entry *Weight3ExtendedStatEntry
		for e := range entrySeq {
			if e.StatRange.Contains(statValue) {
				entry = &e
				break
			}
		}
		if entry == nil {
			panic("no matching range")
		}

		calc := float64(statValue)*entry.RatingWeight + entry.RatingOffset
		simSubTotal += calc
	}

	return simSubTotal
}

func (wer *Weight3ExtendedRanged) Equals(other *Weight3ExtendedRanged) bool {
	return slices.Equal(wer.StatList, other.StatList) &&
		slices.Equal(wer.SimList, other.SimList) &&
		wer.SimPriority.Equals(&other.SimPriority) &&
		wer.StatWeights.Equals(&other.StatWeights, (*Weight3ExtendedStatEntry).Equals)
}

func (wer *Weight3ExtendedRanged) ConvertToWeight2(verificationInputs []WeightInput) *Weight2Extended {
	weight2 := Weight2Extended_Make(wer.SimList, wer.StatList)
	for entry := range wer.StatWeights.SeqKey1Key2ValueSeqEntries() {
		bestValue := chooseBest(entry.ValueSeq)
		weight2.PutWeight(entry.Key1, entry.Key2, bestValue)
	}
	weight2.SimPriority = wer.SimPriority.Clone()
	weight2.UpdateScaling(verificationInputs)
	weight2.FinishAndValidate(verificationInputs)
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

func (wer *Weight3ExtendedRanged) String() string {
	sb := util.StringBuild2{}
	wer.AppendString(&sb)
	return sb.String()
}

func (wer *Weight3ExtendedRanged) AppendString(sb *util.StringBuild2) {
	sb.WriteString("(")
	for _, simType := range wer.SimList {
		priority := wer.SimPriority.GetOrPanic(simType)
		sb.WriteString(simType.Name())
		sb.WriteString("(scale1=")
		sb.WriteFloatScientific64(priority.RatioScale)
		sb.WriteString(",scale2=")
		sb.WriteFloatScientific64(priority.RangingScale)
		sb.WriteString(",offset=")
		sb.WriteFloatScientific64(priority.RangingOffset)
		sb.WriteRune(',')
		for statType, valueSeq := range wer.StatWeights.SeqKey2ValueSeqWithKey1(simType) {
			sb.WriteString(statType.Name())
			sb.WriteString("=(")
			for entry := range valueSeq {
				sb.WriteRune('(')
				sb.WriteUint32(entry.StatRange.Minimum)
				sb.WriteRune(',')
				sb.WriteUint32(entry.StatRange.Maximum)
				sb.WriteRune(',')
				sb.WriteFloatScientific64(entry.RatingWeight)
				sb.WriteRune(',')
				sb.WriteFloatScientific64(entry.RatingOffset)
				sb.WriteRune(')')
			}
			sb.WriteString("),")
		}
		sb.Rewind(1)
		sb.WriteRune(')')
	}
	sb.WriteRune(')')
}
