package fitting2

import (
	"iter"
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
	BaseEachStatProcess[*fitting2EachFields]
}

type fitting2EachFields struct {
	statType    stats.StatType
	simType     stats.SimType
	process     SingleSegmented2
	resultSlice []util_weight.FittingInterimResult2
}

func (f fitting2EachFields) InnerPrinter() *util.PrintRecorder {
	return f.process.Printer
}

func (f fitting2EachFields) Stopwatch() *util.Stopwatch {
	return &f.process.Stopwatch
}

func (f fitting2EachFields) Results() iter.Seq[util_weight.FittingInterimResult2] {
	return slices.Values(f.resultSlice)
}

func (fe *FittingEachStatWeightProcess2) Run(stopwatch *util.Stopwatch, cancel util_async.CancelSignal) util_collection.Optional[weight_types.Weight3ExtendedRanged] {
	fe.ChooseScaling()
	fe.launchEachNested(cancel)
	fe.CalcMetrics(stopwatch)
	if !fe.Failed {
		weights := fe.BuildResult()
		return util_collection.Optional_OfValue(*weights)
	} else {
		return util_collection.Optional_Empty[weight_types.Weight3ExtendedRanged]()
	}
}

func (fe *FittingEachStatWeightProcess2) ChooseScaling() {
	fe.ScaleStats = util_weight.ChooseStatScalingBasic(fe.InputData, c_fitting2_statScaledMaxValue, true, fe.Printer)
	fe.ScaleSims = util_weight.ChooseSimUnfriendlyUnitScaleAndOffset(fe.InputData, fe.RequiredSims)
}

func (fe *FittingEachStatWeightProcess2) prepareSamples(statType stats.StatType, simType stats.SimType) []util_weight.FittingSample {
	scaleStat := fe.ScaleStats.GetOrPanic(statType)
	scaleSim := fe.ScaleSims.GetOrPanic(simType)

	samples := make([]util_weight.FittingSample, len(fe.InputData))
	for i := range fe.InputData {
		statValue := fe.InputData[i].TotalStat.GetFloat(statType)
		simResult := fe.InputData[i].SimResult.Get(simType)
		samples[i] = util_weight.FittingSample{
			StatValue: statValue * scaleStat,
			SimResult: scaleSim.Apply(simResult),
		}
	}
	return samples
}

func (fe *FittingEachStatWeightProcess2) launchEachNested(cancel util_async.CancelSignal) {
	for _, statType := range fe.RequiredStats {
		for _, simType := range fe.RequiredSims {
			innerPrinter := util.PrintRecorder_HoldAll()
			fields := fitting2EachFields{statType: statType, simType: simType}
			scaleStat := fe.ScaleStats.GetOrPanic(statType)
			fields.process.Init(fe.TargetSegmentCount, scaleStat, innerPrinter, fe.Timeout)
			fields.process.SupplyData(fe.prepareSamples(statType, simType))
			fe.Each.Put(statType, simType, &fields)
		}
	}

	channelEach := util_async.SeqToChannel_Cancellable(fe.Each.SeqValues(), cancel)
	util_async.ForEach_Channel(c_fitting2_each_threadCount, channelEach, func(fields *fitting2EachFields) {
		initialResultFuture := fields.process.Run()
		initialResult, hasResult := initialResultFuture.WaitForResult()
		if hasResult && len(initialResult.Segments) > 0 {
			fe.rescaleAndCleanup(initialResult, fields)
		} else {
			fe.Printer.Printf("FAILED FITTING for %s %s\n", fields.statType.Name(), fields.simType.Name())
			fe.Failed = true
		}
	})
}

func (fe *FittingEachStatWeightProcess2) rescaleAndCleanup(initialSet InitialResultSet, fields *fitting2EachFields) {
	fields.resultSlice = fe.ConvertAndScaleResult(initialSet, fields.statType)
	fields.resultSlice = fe.CleanupRanges(fields.resultSlice)
}
