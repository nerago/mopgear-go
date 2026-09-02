package weight_types

import (
	"math"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_rank"
)

// Weight4Segmented
type Weight4Segmented struct {
	Segments    []Weight4SingleSegment
	StatList    []stats.StatType
	SimList     []stats.SimType
	SimPriority SimPriorityBasic
}

func Weight4Segmented_Make(statList []stats.StatType, simList []stats.SimType, priority SimPriorityBasic) *Weight4Segmented {
	return &Weight4Segmented{
		StatList:    statList,
		SimList:     simList,
		SimPriority: priority,
	}
}

type Weight4SimOffset struct {
	Scale  float64
	Offset float64
}

type Weight4SegmentBound struct {
	stats.StatTypeMap[StatRange]
}

func (bnd *Weight4SegmentBound) Equals(other *Weight4SegmentBound) bool {
	return bnd.StatTypeMap.Equals(&other.StatTypeMap, func(a *StatRange, b *StatRange) bool { return a.Equals(*b) })
}

func (bnd *Weight4SegmentBound) BoundContains(block *stats.StatBlock) bool {
	for statType, statRange := range bnd.SeqKeyValue() {
		value := block.GetUInt(statType)
		if !statRange.Contains(value) {
			return false
		}
	}
	return true
}

type Weight4SingleSegment struct {
	Bound   Weight4SegmentBound
	Weights util_collection.MapMap[stats.SimType, stats.StatType, float64]
	Offsets stats.SimTypeMap[Weight4SimOffset]
}

func (seg *Weight4SingleSegment) Equals(other *Weight4SingleSegment) bool {
	return seg.Bound.Equals(&other.Bound) &&
		seg.Weights.Equals(&other.Weights, func(a *float64, b *float64) bool { return *a == *b }) &&
		seg.Offsets.Equals(&other.Offsets, func(a, b *Weight4SimOffset) bool { return *a == *b })
}

func (seg *Weight4SingleSegment) IsEmpty() bool {
	if seg.Bound.IsEmpty() {
		return true
	}
	for value := range seg.Weights.SeqValues() {
		if stats.IsUsefulWeightNumber(value) {
			return false
		}
	}
	return true
}

func (seg *Weight4SingleSegment) Contains(stats *stats.StatBlock) bool {
	for statType, statRange := range seg.Bound.SeqKeyValue() {
		value := stats.GetUInt(statType)
		if !statRange.Contains(value) {
			return false
		}
	}
	return true
}

func (wer *Weight4Segmented) IsEmpty() bool {
	if wer.SimPriority.IsEmpty() {
		return true
	}
	for _, segment := range wer.Segments {
		if !segment.IsEmpty() {
			return false
		}
	}
	return true
}

func (wer *Weight4Segmented) AddSegment(segment Weight4SingleSegment) {
	wer.Segments = append(wer.Segments, segment)
}

func (wer *Weight4Segmented) AddWeight2AsSegment(weight2 *Weight2Extended, bounds *Weight4SegmentBound) {
	seg := Weight4SingleSegment{
		Bound:   *bounds,
		Weights: weight2.DetailedWeights,
		Offsets: stats.SimTypeMap[Weight4SimOffset]{},
	}
	for simType, entry := range weight2.SimPriority.entries.SeqKeyValue() {
		seg.Offsets.Put(simType,
			Weight4SimOffset{
				Scale:  entry.RangingScale,
				Offset: entry.RangingOffset,
			},
		)
	}
	wer.AddSegment(seg)
}

func (wer *Weight4Segmented) FinishAndValidate() error {
	for _, seg := range wer.Segments {
		for statType := range seg.Weights.SeqKey2() {
			if !slices.Contains(wer.StatList, statType) {
				return util.ErrorTracedNew("weight given for unlisted stat")
			}
		}
	}
	for _, seg := range wer.Segments {
		for simType := range seg.Weights.SeqKey1() {
			if !slices.Contains(wer.SimList, simType) {
				return util.ErrorTracedNew("weight given for unlisted sim")
			}
		}
	}

	for _, seg := range wer.Segments {
		for _, simType := range wer.SimList {
			for _, statType := range wer.StatList {
				if !seg.Weights.Has(simType, statType) {
					return util.ErrorTracedNew("missing weight for " + statType.Name() + " " + simType.Name())
				}
			}
		}
	}
	for _, seg := range wer.Segments {
		for entry := range seg.Weights.SeqKey1Key2ValueEntries() {
			simType := entry.Key1
			statType := entry.Key2
			if !slices.Contains(wer.StatList, statType) {
				return util.ErrorTracedNew("unexpected weight for " + statType.Name())
			}
			if !slices.Contains(wer.SimList, simType) {
				return util.ErrorTracedNew("unexpected weight for " + simType.Name())
			}
			if seg.Bound.IsEmpty() {
				return util.ErrorTracedNew("empty bound")
			}
		}
	}

	for simType, value := range wer.SimPriority.SeqTypeValue() {
		if util.FloatNonZero(value) {
			if !slices.Contains(wer.SimList, simType) {
				return util.ErrorTracedNew("priority given for unlisted sim")
			}
		}
	}
	for _, simType := range wer.SimList {
		value, hasValue := wer.SimPriority.Get(simType)
		if !hasValue || util.FloatEqualsZero(value) {
			return util.ErrorTracedNew("priority missing for " + simType.Name())
		}
	}

	// TODO check overlap?

	return nil
}

func (wer *Weight4Segmented) CalcStatScore(stats *stats.StatBlock) float64 {
	seg := wer.selectSegment(stats)
	if seg == nil {
		return math.NaN()
	}

	totalSum := 0.0
	for _, simType := range wer.SimList {
		subTotal := seg.calcSingleSimScoreUnscaled(stats, simType)

		priorityRatio := wer.SimPriority.GetOrPanic(simType)
		subTotal *= priorityRatio

		totalSum += subTotal
	}
	return totalSum
}

func (wer *Weight4Segmented) CalcStatScoreWithBonus(stats *stats.StatBlock, simBonus *stats.SimTypeMap[float64]) float64 {
	seg := wer.selectSegment(stats)
	if seg == nil {
		return math.NaN()
	}

	totalSum := 0.0
	for _, simType := range wer.SimList {
		subTotal := seg.calcSingleSimScoreUnscaled(stats, simType)

		priorityRatio := wer.SimPriority.GetOrPanic(simType)
		subTotal *= priorityRatio
		subTotal *= simBonus.GetOrDefault(simType, 1)

		totalSum += subTotal
	}
	return totalSum
}

func (wer *Weight4Segmented) selectSegment(stats *stats.StatBlock) *Weight4SingleSegment {
	for seg := range util_collection.ForPointer(wer.Segments) {
		if seg.Contains(stats) {
			return seg
		}
	}
	// TODO something better, find "nearest"
	return nil // ERROR: no valid segment
}

func (seg *Weight4SingleSegment) calcSingleSimScoreUnscaled(statBlock *stats.StatBlock, simType stats.SimType) float64 {
	subTotal := 0.0

	for statType, detailWeight := range seg.Weights.SeqKey2ValueWithKey1(simType) {
		specificValue := detailWeight * statBlock.GetFloat(statType)
		subTotal += specificValue
	}

	return subTotal
}

func (wer *Weight4Segmented) Equals(other *Weight4Segmented) bool {
	return slices.Equal(wer.StatList, other.StatList) &&
		slices.Equal(wer.SimList, other.SimList) &&
		wer.SimPriority.Equals(&other.SimPriority) &&
		util_collection.EqualFunc_IgnoreOrder_Pointer(wer.Segments, other.Segments, (*Weight4SingleSegment).Equals)
}

func (wer *Weight4Segmented) ConvertToWeight2(inputData []WeightInput) (*Weight2Extended, error) {
	weight2 := Weight2Extended_Make(wer.SimList, wer.StatList)
	seg := wer.chooseBest(inputData)
	if seg == nil {
		return nil, util.ErrorTracedNew("no useful segment to convert")
	}

	seg.Weights.Foreach(func(sim stats.SimType, stat stats.StatType, value float64) {
		weight2.PutWeight(sim, stat, value)
	})

	for sim, value := range wer.SimPriority.SeqTypeValue() {
		offset := seg.Offsets.GetOrPanic(sim)
		if err := weight2.SetSimScale(sim, offset.Scale, offset.Offset, value); err != nil {
			return nil, err
		}
	}

	if err := weight2.UpdateScaling(inputData); err != nil {
		return nil, err
	}
	if err := weight2.FinishAndValidate(inputData); err != nil {
		return nil, err
	}
	return weight2, nil
}

// best = most coverage of sample data
func (wer *Weight4Segmented) chooseBest(dataInput []WeightInput) *Weight4SingleSegment {
	best := util_rank.BestCollector1[Weight4SingleSegment]{}
	for segment := range util_collection.ForPointer(wer.Segments) {
		containedCount := 0
		for input := range util_collection.ForPointer(dataInput) {
			if segment.Bound.BoundContains(&input.TotalStat) {
				containedCount++
			}
		}
		best.Offer(segment, float64(containedCount))
	}
	return best.GetBestPointerOrNil()
}

func (wer *Weight4Segmented) String() string {
	sb := util.StringBuild2{}
	wer.AppendString(&sb)
	return sb.String()
}

func (wer *Weight4Segmented) AppendString(sb *util.StringBuild2) {
	sb.WriteString("(pri(")
	for _, simType := range wer.SimList {
		priority := wer.SimPriority.GetOrPanic(simType)
		sb.WriteString(simType.Name())
		sb.WriteString("=")
		sb.WriteFloatScientific64(priority)
		sb.WriteRune(',')
	}
	sb.Rewind(1)
	sb.WriteRune(')')
	for seg := range util_collection.ForPointer(wer.Segments) {
		sb.WriteString("seg(bnd(")
		for statType, statRange := range seg.Bound.SeqKeyValue() {
			sb.WriteString(statType.Name())
			sb.WriteRune(',')
			sb.WriteUint32(statRange.Minimum)
			sb.WriteRune(',')
			sb.WriteUint32(statRange.Maximum)
			sb.WriteRune(',')
		}
		sb.Rewind(1)
		sb.WriteString(")wei(")
		seg.Weights.Foreach(func(simType stats.SimType, statType stats.StatType, value float64) {
			sb.WriteRune('(')
			sb.WriteString(simType.Name())
			sb.WriteRune(',')
			sb.WriteString(statType.Name())
			sb.WriteRune(',')
			sb.WriteFloatScientific64(value)
			sb.WriteRune(')')
		})
		sb.WriteString(")off(")
		for simType, entry := range seg.Offsets.SeqKeyValue() {
			sb.WriteRune('(')
			sb.WriteString(simType.Name())
			sb.WriteRune(',')
			sb.WriteFloatScientific64(entry.Scale)
			sb.WriteRune(',')
			sb.WriteFloatScientific64(entry.Offset)
			sb.WriteRune(')')
		}
		sb.WriteString("))")
	}
	sb.WriteRune(')')
}
