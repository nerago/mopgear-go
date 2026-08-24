package fitting1

import (
	"cmp"
	"math"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	util_weight2 "github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

type FittingEachStatWeightProcess struct {
	printer *util.PrintRecorder
	timeout int

	onlyComputeSingleSegmentEach bool
	inputData                    []weight_types.WeightInput
	requiredStats                []stats.StatType
	requiredSims                 []stats.SimType

	scaleSims  stats.SimTypeMap[util_weight2.ScaleAndOffset]
	scaleStats stats.StatTypeMap[float64]

	each     util_collection.MapMap[stats.StatType, stats.SimType, *fittingEachFields]
	hasError bool
}

type fittingEachFields struct {
	statType    stats.StatType
	simType     stats.SimType
	resultSlice []util_weight2.FittingInterimResult
}

func (fe *FittingEachStatWeightProcess) Init(printer *util.PrintRecorder, timeout int) {
	fe.printer = printer
	fe.timeout = timeout
}

func (fe *FittingEachStatWeightProcess) SetRequiredStats(requiredStats []stats.StatType, requiredSims []stats.SimType) {
	fe.requiredStats = requiredStats
	fe.requiredSims = requiredSims
}

// if just converting directly to weights1
func (fe *FittingEachStatWeightProcess) SetOnlyComputeSingleSegmentEach(lazy bool) {
	fe.onlyComputeSingleSegmentEach = lazy
}

func (fe *FittingEachStatWeightProcess) SupplyData(inputData []weight_types.WeightInput) {
	fe.inputData = inputData
}

func (fe *FittingEachStatWeightProcess) Run(cancel util_async.CancelSignal) weight_types.WeightResult {
	fe.chooseScaling()
	fe.launchEachNested(cancel)
	if !fe.hasError {
		weight := fe.buildResult()
		stopwatch := fe.calcMetrics()
		return weight_types.WeightResult{Weight: weight, SolveTime: stopwatch.Elapsed(), Status: highs.ModelStatusUnknown}
	} else {
		stopwatch := fe.calcMetrics()
		return weight_types.WeightResult{SolveTime: stopwatch.Elapsed(), Status: highs.ModelStatusModelError}
	}
}

func (fe *FittingEachStatWeightProcess) calcMetrics() *util.Stopwatch {
	stopwatch := util.StopwatchMakeStopped()
	for fields := range fe.each.SeqValues() {
		for _, detail := range fields.resultSlice {
			stopwatch.AddElapsedFrom(&detail.StopwatchSolver)
		}
	}
	return stopwatch
}

func (fe *FittingEachStatWeightProcess) buildResult() *weight_types.Weight3ExtendedRanged {
	weights := weight_types.Weight3ExtendedRanged_Make(fe.requiredStats, fe.requiredSims)
	fe.each.Foreach(func(statType stats.StatType, simType stats.SimType, value *fittingEachFields) {
		for _, detail := range value.resultSlice {
			weights.AddDetailWeight(simType, statType, detail.StatRange, detail.LineSlope, detail.LineOffset, detail.IncludePercentOfTotal)
		}
	})
	// TODO final weight multipliers as needed
	for _, simType := range fe.requiredSims {
		weights.SetSimScale(simType, 1, 0, 1)
	}
	weights.FinishAndValidate()
	return weights
}

func (fe *FittingEachStatWeightProcess) launchEachNested(cancel util_async.CancelSignal) {
	for _, statType := range fe.requiredStats {
		for _, simType := range fe.requiredSims {
			fields := fittingEachFields{statType: statType, simType: simType}
			fe.each.Put(statType, simType, &fields)
		}
	}

	channelEach := util_async.SeqToChannel_Cancellable(fe.each.SeqValues(), cancel)
	util_async.ForEach_Channel(c_fitting_each_threadCount, channelEach, func(fields *fittingEachFields) {
		printer := util.PrintRecorder_HoldAll()

		process := FittingSingleStatSegmentsProcess{}
		process.Init(printer, fe.timeout)
		process.SetOnlyComputeSingleSegment(fe.onlyComputeSingleSegmentEach)
		process.SupplyData(fe.prepareSamples(fields.statType, fields.simType))
		initialResult := process.Run(cancel)

		fe.rescaleAndCleanup(initialResult, fields)
		fe.printer.AppendOther(printer)
	})
}

func (fe *FittingEachStatWeightProcess) chooseScaling() {
	fe.scaleStats = util_weight2.ChooseStatScalingBasic(fe.inputData, c_fitting_statScaledRangeHigh, true, fe.printer)
	fe.scaleSims = util_weight2.ChooseSimUnfriendlyUnitScaleAndOffset(fe.inputData, fe.requiredSims)
}

func (fe *FittingEachStatWeightProcess) prepareSamples(statType stats.StatType, simType stats.SimType) []util_weight2.FittingSample {
	scaleStat := fe.scaleStats.GetOrPanic(statType)
	scaleSim := fe.scaleSims.GetOrPanic(simType)

	samples := make([]util_weight2.FittingSample, len(fe.inputData))
	for i := range fe.inputData {
		statValue := fe.inputData[i].TotalStat.GetFloat(statType)
		simResult := fe.inputData[i].SimResult.Get(simType)
		samples[i] = util_weight2.FittingSample{
			StatValue: statValue * scaleStat,
			SimResult: scaleSim.Apply(simResult),
		}
	}
	return samples
}

// inner process ran: simScaled = lineSlopeInternal*statScaled + lineOffsetInternal
// for the sake of Weight3 we want to retain the simScaling, but undo the stats
//
//	want to end up with: simScaled = lineSlopeUsable * stat + lineOffsetUsable
//
// expanding on the first one: simScaled = lineSlopeInternal*(stat*scaleStat) + lineOffsetInternal
// when stat=0: simScaled = lineOffsetInternal. not sure if it's a true equality, but doesn't seem to need scale at least
// to fix the stat scale lineSlopeUsable * stat = lineSlopeInternal * (stat * scaleStat)
//
//	lineSlopeUsable = lineSlopeInternal * scaleStat
func (fe *FittingEachStatWeightProcess) rescaleAndCleanup(initialMap map[weight_types.StatRangeFloat]FittingSingleStatResult, fields *fittingEachFields) {
	if initialMap != nil {
		fields.resultSlice = fe.convertAndScaleResult(initialMap, fields.statType)
		fields.resultSlice = fe.cleanupRanges(fields.resultSlice)
	} else {
		fe.hasError = true
	}
}

func (fe *FittingEachStatWeightProcess) convertAndScaleResult(initialMap map[weight_types.StatRangeFloat]FittingSingleStatResult, statType stats.StatType) []util_weight2.FittingInterimResult {
	scaleStat := fe.scaleStats.GetOrPanic(statType)

	resultSlice := make([]util_weight2.FittingInterimResult, 0, len(initialMap))
	for rangeScaled, resultInitial := range initialMap {
		interim := util_weight2.FittingInterimResult{
			LineSlope:  resultInitial.LineSlope * scaleStat,
			LineOffset: resultInitial.LineOffset,
			StatRange: weight_types.StatRange{
				Minimum: uint32(math.Round(rangeScaled.Minimum / scaleStat)),
				Maximum: uint32(math.Round(rangeScaled.Maximum / scaleStat)),
			},
			IncludeCount:               resultInitial.IncludeCount,
			IncludePercentOfTotal:      float64(resultInitial.IncludeCount) / float64(len(fe.inputData)),
			IncludePercentOfStageInput: resultInitial.IncludePercentOfStageInput,
			BuiltSequence:              []int{resultInitial.BuiltSequence},
			StopwatchSolver:            resultInitial.StopwatchSolver,
		}
		resultSlice = append(resultSlice, interim)
	}
	return resultSlice
}

func (fe *FittingEachStatWeightProcess) cleanupRanges(results []util_weight2.FittingInterimResult) []util_weight2.FittingInterimResult {
	slices.SortFunc(results, func(a, b util_weight2.FittingInterimResult) int {
		return cmp.Or(cmp.Compare(a.StatRange.Minimum, b.StatRange.Minimum), cmp.Compare(a.StatRange.Maximum, b.StatRange.Maximum))
	})

	results[0].StatRange.Minimum = 0
	results[len(results)-1].StatRange.Maximum = math.MaxUint32

	for i := range len(results) - 1 {
		if fe.updateBreakpoint(&results[i], &results[i+1]) {
			util_collection.DeleteIndexInPlace(&results, i+1)
		}
	}

	return results
}

func (fe *FittingEachStatWeightProcess) updateBreakpoint(one, two *util_weight2.FittingInterimResult) (deleteSecond bool) {
	if one.StatRange.RangeSize() < c_fitting_minimum_stat_coverage || two.StatRange.RangeSize() < c_fitting_minimum_stat_coverage {
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
		one.StopwatchSolver.AddElapsedFrom(&two.StopwatchSolver)
		return true
	} else if one.StatRange.Maximum >= two.StatRange.Minimum && one.StatRange.Maximum <= two.StatRange.Minimum+c_fitting_permitted_overlap_fix {
		// if maximum has small overlap into next minimum, fix
		newMinimum := fe.chooseNicerNumber(one.StatRange.Maximum-1, two.StatRange.Minimum+1)
		two.StatRange.Minimum = newMinimum
		one.StatRange.Maximum = newMinimum - 1
	} else if one.StatRange.Overlap(two.StatRange) {
		// more overlap than threshold in previous, fail
		// shouldn't happen, anything more than exact equality due to scaling and rounding shouldn't either
		panic("overlapping ranges created")
	} else if one.StatRange.Maximum < two.StatRange.Minimum-1 {
		// part of range got dropped earlier
		newMinimum := fe.chooseNicerNumber(one.StatRange.Maximum-1, two.StatRange.Minimum+1)
		two.StatRange.Minimum = newMinimum
		one.StatRange.Maximum = newMinimum - 1
	} else if one.StatRange.Minimum < one.StatRange.Maximum && one.StatRange.Maximum+1 == two.StatRange.Minimum && two.StatRange.Minimum < two.StatRange.Maximum {
		// passes all expected conditions
	} else {
		panic("internal logic error in updateBreakpoint")
	}
	return false
}

func (fe *FittingEachStatWeightProcess) chooseNicerNumber(lo uint32, hi uint32) uint32 {
	if lo > hi {
		lo, hi = hi, lo
	}

	if hi-lo > c_fitting_number_nice_number_interval {
		mid := lo + (hi-lo)/2
		mid -= mid % c_fitting_number_nice_number_interval
		return mid
	} else {
		for v := lo; v <= hi; v++ {
			if v%c_fitting_number_nice_number_interval == 0 {
				return v
			}
		}
		mid := lo + (hi-lo)/2
		return mid
	}
}
