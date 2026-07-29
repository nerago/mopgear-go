package fitting2

import (
	"cmp"
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/weightfind/util_weight"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
)

const (
	c_fitting2_each_threadCount = 8

	c_fitting2_minimum_stat_coverage       = 100
	c_fitting2_permitted_overlap_fix       = 50
	c_fitting2_number_nice_number_interval = 5
)

type FittingEachStatWeightProcess2 struct {
	printer *util.PrintRecorder
	timeout int

	targetSegmentCount int
	inputData          []weight_types.WeightInput
	requiredStats      []stats.StatType
	requiredSims       []stats.SimType
	targetRatios       weight_types.SimPriorityBasic

	scaleSims  util_collection.EnumMap[stats.SimType, util_weight.ScaleAndOffset]
	scaleStats util_collection.EnumMap[stats.StatType, float64]

	each util_collection.MapMap[stats.StatType, stats.SimType, *fitting2EachFields]
}

type fitting2EachFields struct {
	statType stats.StatType
	simType  stats.SimType
	process  SingleSegmented2
	//resultSet InitialResultSet
	resultSlice []util_weight.FittingInterimResult2
}

func (fe *FittingEachStatWeightProcess2) Init(targetSegmentCount int, printer *util.PrintRecorder, timeout int) {
	fe.targetSegmentCount = targetSegmentCount
	fe.printer = printer
	fe.timeout = timeout
}

func (fe *FittingEachStatWeightProcess2) SetRequiredStats(requiredStats []stats.StatType, requiredSims []stats.SimType) {
	fe.requiredStats = requiredStats
	fe.requiredSims = requiredSims
}

func (fe *FittingEachStatWeightProcess2) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	fe.targetRatios = targetRatios
}

func (fe *FittingEachStatWeightProcess2) SupplyData(inputData []weight_types.WeightInput) {
	fe.inputData = inputData
}

func (fe *FittingEachStatWeightProcess2) Run(stopwatch *util.Stopwatch, cancel util_async.CancelSignal) weight_types.Weight3ExtendedRanged {
	fe.chooseScaling()
	fe.launchEachNested(cancel)
	weights := fe.buildResult()
	fe.calcMetrics(stopwatch)
	return *weights
}

func (fe *FittingEachStatWeightProcess2) calcMetrics(stopwatch *util.Stopwatch) {
	for fields := range fe.each.SeqValues() {
		stopwatch.AddElapsedFrom(&fields.process.stopwatch)
	}
}

func (fe *FittingEachStatWeightProcess2) buildResult() *weight_types.Weight3ExtendedRanged {
	weights := weight_types.Weight3ExtendedRanged_Make(fe.requiredStats, fe.requiredSims)
	fe.each.ForeachWithKeys(func(statType stats.StatType, simType stats.SimType, value *fitting2EachFields) {
		for _, detail := range value.resultSlice {
			weights.AddDetailWeight(simType, statType, detail.StatRange, detail.LineSlope, detail.LineOffset, detail.IncludePercentOfTotal)
		}
	})
	for _, simType := range fe.requiredSims {
		ratio := fe.targetRatios.GetOrPanic(simType)
		weights.AddSimScale(simType, 1, 0, ratio)
	}
	weights.FinishAndValidate()
	return weights
}

func (fe *FittingEachStatWeightProcess2) launchEachNested(cancel util_async.CancelSignal) {
	for _, statType := range fe.requiredStats {
		for _, simType := range fe.requiredSims {
			printer := util.PrintRecorder_HoldAll()
			fields := fitting2EachFields{statType: statType, simType: simType}
			scaleStat := fe.scaleStats.GetOrPanic(statType)
			fields.process.Init(fe.targetSegmentCount, scaleStat, printer, fe.timeout)
			fields.process.SupplyData(fe.prepareSamples(statType, simType))
			fe.each.Put(statType, simType, &fields)
		}
	}

	channelEach := util_async.SeqToChannel_Cancellable(fe.each.SeqValues(), cancel)
	util_async.ForEach_Channel(c_fitting2_each_threadCount, channelEach, func(fields *fitting2EachFields) {
		initialResultFuture := fields.process.Run()
		initialResult := initialResultFuture.WaitForResultOrPanic()
		fe.rescaleAndCleanup(initialResult, fields)
	})
}

func (fe *FittingEachStatWeightProcess2) chooseScaling() {
	fe.scaleStats = util_weight.ChooseStatScalingBasic(fe.inputData, c_fitting2_statScaledMaxValue, true, fe.printer)
	fe.scaleSims = util_weight.ChooseSimUnfriendlyUnitScaleAndOffset(fe.inputData, fe.requiredSims)
}

func (fe *FittingEachStatWeightProcess2) prepareSamples(statType stats.StatType, simType stats.SimType) []util_weight.FittingSample {
	scaleStat := fe.scaleStats.GetOrPanic(statType)
	scaleSim := fe.scaleSims.GetOrPanic(simType)

	samples := make([]util_weight.FittingSample, len(fe.inputData))
	for i := range fe.inputData {
		statValue := fe.inputData[i].TotalStat.GetFloat(statType)
		simResult := fe.inputData[i].SimResult.Get(simType)
		samples[i] = util_weight.FittingSample{
			StatValue: statValue * scaleStat,
			SimResult: scaleSim.Apply(simResult),
		}
	}
	return samples
}

func (fe *FittingEachStatWeightProcess2) rescaleAndCleanup(initialSet InitialResultSet, fields *fitting2EachFields) {
	fields.resultSlice = fe.convertAndScaleResult(initialSet, fields.statType)
	fields.resultSlice = fe.cleanupRanges(fields.resultSlice)
}

func (fe *FittingEachStatWeightProcess2) convertAndScaleResult(initialSet InitialResultSet, statType stats.StatType) []util_weight.FittingInterimResult2 {
	scaleStat := fe.scaleStats.GetOrPanic(statType)

	resultSlice := make([]util_weight.FittingInterimResult2, 0, len(initialSet.Segments))
	for i, resultInitial := range initialSet.Segments {
		interim := util_weight.FittingInterimResult2{
			LineSlope:  resultInitial.LineSlope,
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

func (fe *FittingEachStatWeightProcess2) cleanupRanges(results []util_weight.FittingInterimResult2) []util_weight.FittingInterimResult2 {
	slices.SortFunc(results, func(a, b util_weight.FittingInterimResult2) int {
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

func (fe *FittingEachStatWeightProcess2) updateBreakpoint(one, two *util_weight.FittingInterimResult2) (deleteSecond bool) {
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
		newMinimum := fe.chooseNicerNumber(one.StatRange.Maximum-1, two.StatRange.Minimum+1)
		two.StatRange.Minimum = newMinimum
		one.StatRange.Maximum = newMinimum - 1
	} else if one.StatRange.Overlap(two.StatRange) {
		// more overlap than threshold in previous, fail
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

func (fe *FittingEachStatWeightProcess2) chooseNicerNumber(lo uint32, hi uint32) uint32 {
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
