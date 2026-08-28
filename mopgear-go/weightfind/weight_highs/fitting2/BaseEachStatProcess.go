package fitting2

import (
	"cmp"
	"iter"
	"math"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type BaseEachStatProcess[F IEachFields] struct {
	Printer *util.PrintRecorder
	Timeout int

	TargetSegmentCount int
	InputData          []weight_types.WeightInput
	RequiredStats      []stats.StatType
	RequiredSims       []stats.SimType
	targetRatios       weight_types.SimPriorityBasic

	ScaleSims  stats.SimTypeMap[util_weight.ScaleAndOffset]
	ScaleStats stats.StatTypeMap[float64]

	Each           util_collection.MapMap[stats.StatType, stats.SimType, F]
	CancelInternal util_async.CancelSignalBasic
}

type IEachFields interface {
	Stopwatch() *util.Stopwatch
	InnerPrinter() *util.PrintRecorder
	Results() iter.Seq[util_weight.FittingInterimResult2]
}

func (be *BaseEachStatProcess[F]) Init(targetSegmentCount int, printer *util.PrintRecorder, timeout int) {
	be.TargetSegmentCount = targetSegmentCount
	be.Printer = printer
	be.Timeout = timeout
	be.CancelInternal = *util_async.CancelSignal_Make()
}

func (be *BaseEachStatProcess[F]) SetRequiredStats(requiredStats []stats.StatType, requiredSims []stats.SimType) {
	be.RequiredStats = requiredStats
	be.RequiredSims = requiredSims
}

func (be *BaseEachStatProcess[F]) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	be.targetRatios = targetRatios
}

func (be *BaseEachStatProcess[F]) SupplyData(inputData []weight_types.WeightInput) {
	be.InputData = inputData
}

func (be *BaseEachStatProcess[F]) CalcMetrics() *util.Stopwatch {
	stopwatch := util.StopwatchMakeStopped()
	for fields := range be.Each.SeqValues() {
		stopwatch.AddElapsedFrom(fields.Stopwatch())
		be.Printer.AppendOther(fields.InnerPrinter())
	}
	return stopwatch
}

func (be *BaseEachStatProcess[F]) BuildResult() *weight_types.Weight3ExtendedRanged {
	weights := weight_types.Weight3ExtendedRanged_Make(be.RequiredStats, be.RequiredSims)
	be.Each.Foreach(func(statType stats.StatType, simType stats.SimType, value F) {
		for detail := range value.Results() {
			weights.AddDetailWeight(simType, statType, detail.StatRange, detail.LineSlope, detail.LineOffset, detail.IncludePercentOfTotal)
		}
	})
	for _, simType := range be.RequiredSims {
		ratio := be.targetRatios.GetOrPanic(simType)
		weights.SetSimScale(simType, 1, 0, ratio)
	}
	weights.FinishAndValidate()
	return weights
}

func (be *BaseEachStatProcess[F]) ConvertAndScaleResult(initialSet InitialResultSet, statType stats.StatType) []util_weight.FittingInterimResult2 {
	scaleStat := be.ScaleStats.GetOrPanic(statType)

	resultSlice := make([]util_weight.FittingInterimResult2, 0, len(initialSet.Segments))
	for i, resultInitial := range initialSet.Segments {
		interim := util_weight.FittingInterimResult2{
			LineSlope:  resultInitial.LineSlope * scaleStat,
			LineOffset: resultInitial.LineOffset,
			StatRange: weight_types.StatRange{
				Minimum: uint32(math.Round(resultInitial.StatRange.Minimum / scaleStat)),
				Maximum: uint32(math.Round(resultInitial.StatRange.Maximum / scaleStat)),
			},
			IncludeCount:               resultInitial.IncludeCount,
			IncludePercentOfTotal:      resultInitial.IncludePercentOfTotal,
			IncludePercentOfStageInput: 0,
			BuiltSequence:              []int{i},
		}
		resultSlice = append(resultSlice, interim)
	}
	return resultSlice
}

func (be *BaseEachStatProcess[F]) CleanupRanges(results []util_weight.FittingInterimResult2) []util_weight.FittingInterimResult2 {
	slices.SortFunc(results, func(a, b util_weight.FittingInterimResult2) int {
		return cmp.Or(cmp.Compare(a.StatRange.Minimum, b.StatRange.Minimum), cmp.Compare(a.StatRange.Maximum, b.StatRange.Maximum))
	})

	results[0].StatRange.Minimum = 0
	results[len(results)-1].StatRange.Maximum = math.MaxUint32

	for i := 0; i < len(results)-1; i++ {
		if be.updateBreakpoint(&results[i], &results[i+1]) {
			util_collection.DeleteIndexInPlace(&results, i+1)
		}
	}

	return results
}

func (be *BaseEachStatProcess[F]) updateBreakpoint(one, two *util_weight.FittingInterimResult2) (deleteSecond bool) {
	// TODO similar rules with very low include count
	if one.StatRange.RangeSize() < c_fitting2_minimum_stat_coverage || two.StatRange.RangeSize() < c_fitting2_minimum_stat_coverage {
		// if covers less than 100 stat numbers, merge them
		if one.StatRange.RangeSize() < two.StatRange.RangeSize() {
			one.LineSlope = two.LineSlope
			one.LineOffset = two.LineOffset
		}
		one.StatRange.Maximum = two.StatRange.Maximum
		one.IncludePercentOfStageInput += two.IncludePercentOfStageInput
		one.IncludePercentOfTotal += two.IncludePercentOfTotal
		one.IncludeCount += two.IncludeCount
		one.BuiltSequence = slices.Concat(one.BuiltSequence, two.BuiltSequence)
		return true
	} else if one.StatRange.Maximum >= two.StatRange.Minimum && one.StatRange.Maximum <= two.StatRange.Minimum+c_fitting2_permitted_overlap_fix {
		// if maximum has small overlap into next minimum, fix
		newMinimum := be.chooseNicerNumber(one.StatRange.Maximum-1, two.StatRange.Minimum+1)
		two.StatRange.Minimum = newMinimum
		one.StatRange.Maximum = newMinimum - 1
	} else if one.StatRange.Overlap(two.StatRange) {
		// more overlap than threshold in previous, fail
		panic("overlapping ranges created")
	} else if one.StatRange.Maximum < two.StatRange.Minimum-1 {
		// part of range got dropped earlier
		newMinimum := be.chooseNicerNumber(one.StatRange.Maximum-1, two.StatRange.Minimum+1)
		two.StatRange.Minimum = newMinimum
		one.StatRange.Maximum = newMinimum - 1
	} else if one.StatRange.Minimum < one.StatRange.Maximum && one.StatRange.Maximum+1 == two.StatRange.Minimum && two.StatRange.Minimum < two.StatRange.Maximum {
		// passes all expected conditions
	} else {
		panic("internal logic error in updateBreakpoint")
	}
	return false
}

func (be *BaseEachStatProcess[F]) chooseNicerNumber(lo uint32, hi uint32) uint32 {
	if lo > hi {
		lo, hi = hi, lo
	}

	if hi-lo > c_fitting2_number_nice_number_interval {
		mid := lo + (hi-lo)/2
		mid -= mid % c_fitting2_number_nice_number_interval
		return mid
	} else {
		for v := lo; v <= hi; v++ {
			if v%c_fitting2_number_nice_number_interval == 0 {
				return v
			}
		}
		mid := lo + (hi-lo)/2
		return mid
	}
}
