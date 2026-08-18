package fitting3

import (
	"iter"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting2"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_fitting3_each_threadCount = 3

	//c_fitting3_minimum_stat_coverage       = 100
	//c_fitting3_permitted_overlap_fix       = 50
	//c_fitting3_number_nice_number_interval = 5
)

type FittingEachStatWeightProcess3 struct {
	fitting2.BaseEachStatProcess[*fitting3EachFields]
}

type fitting3EachFields struct {
	statType    stats.StatType
	simType     stats.SimType
	process     FittingSingleSegmented3
	resultSlice []util_weight.FittingInterimResult2
}

func (f fitting3EachFields) InnerPrinter() *util.PrintRecorder {
	return f.process.Printer
}

func (f fitting3EachFields) Stopwatch() *util.Stopwatch {
	return &f.process.Stopwatch
}

func (f fitting3EachFields) Results() iter.Seq[util_weight.FittingInterimResult2] {
	return slices.Values(f.resultSlice)
}

func (fe *FittingEachStatWeightProcess3) Run(cancel util_async.CancelSignal, tracker *util.TrackProgress) weight_types.WeightResult {
	util_async.ChainCancel(cancel, &fe.CancelInternal)
	fe.ChooseScaling()
	fe.launchEachNested(tracker)
	stopwatch := fe.CalcMetrics()
	if !fe.Failed {
		weights := fe.BuildResult()
		return weight_types.WeightResult{Weight: weights, SolveTime: stopwatch.Elapsed(), Status: highs.ModelStatusOptimal}
	} else {
		return weight_types.WeightResult{SolveTime: stopwatch.Elapsed(), Status: highs.ModelStatusModelError}
	}
}

func (fe *FittingEachStatWeightProcess3) ChooseScaling() {
	fe.ScaleStats = util_weight.ChooseStatScalingBasic(fe.InputData, c_fitting3_statScaledMaxValue, true, fe.Printer)
	fe.ScaleSims = util_weight.ChooseSimUnfriendlyUnitScaleAndOffset(fe.InputData, fe.RequiredSims)
}

func (fe *FittingEachStatWeightProcess3) prepareSamples(statType stats.StatType, simType stats.SimType) []util_weight.FittingSample3 {
	scaleStat := fe.ScaleStats.GetOrPanic(statType)
	scaleSim := fe.ScaleSims.GetOrPanic(simType)

	samples := make([]util_weight.FittingSample3, len(fe.InputData))
	for i := range fe.InputData {
		statValue := fe.InputData[i].TotalStat.GetFloat(statType)
		average := fe.InputData[i].SimResult.Get(simType)
		detail := fe.InputData[i].SimResult.GetDetailed2(simType)

		var simResult util_weight.FittingSimDetail
		if detail != nil {
			simResult = util_weight.MakeFittingDetail(average, detail, scaleSim)
		} else {
			simResult = util_weight.MakeFittingDetailFromAverage(average, scaleSim)
		}

		samples[i] = util_weight.FittingSample3{
			StatValue: statValue * scaleStat,
			SimResult: simResult,
		}
	}
	return samples
}

func (fe *FittingEachStatWeightProcess3) launchEachNested(tracker *util.TrackProgress) {
	for _, statType := range fe.RequiredStats {
		for _, simType := range fe.RequiredSims {
			printer := util.PrintRecorder_HoldAll()
			fields := fitting3EachFields{statType: statType, simType: simType}
			scaleStat := fe.ScaleStats.GetOrPanic(statType)
			fields.process.Init(fe.TargetSegmentCount, scaleStat, printer, fe.Timeout)
			fields.process.SupplyData(fe.prepareSamples(statType, simType))
			fe.Each.Put(statType, simType, &fields)
		}
	}

	tracker.RunOuterTracking(fe.Each.Size())
	channelEach := util_async.SeqToChannel_Cancellable(fe.Each.SeqValues(), &fe.CancelInternal)
	util_async.ForEach_Channel(c_fitting3_each_threadCount, channelEach, func(fields *fitting3EachFields) {
		initialResultFuture := fields.process.Run()
		util_async.ChainCancel(&fe.CancelInternal, initialResultFuture)
		initialResult, hasResult := initialResultFuture.WaitForResult()
		if hasResult && len(initialResult.Segments) > 0 {
			fe.rescaleAndCleanup(initialResult, fields)
		} else {
			fe.Printer.Printf("FAILED FITTING for %s %s\n", fields.statType.Name(), fields.simType.Name())
			fe.Failed = true
			fe.CancelInternal.Cancel()
		}
		tracker.NewChild().SetDone()
	})
}

func (fe *FittingEachStatWeightProcess3) rescaleAndCleanup(initialSet fitting2.InitialResultSet, fields *fitting3EachFields) {
	fields.resultSlice = fe.ConvertAndScaleResult(initialSet, fields.statType)
	fields.resultSlice = fe.CleanupRanges(fields.resultSlice)
}
