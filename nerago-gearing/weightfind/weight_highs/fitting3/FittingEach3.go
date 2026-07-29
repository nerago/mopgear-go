package fitting3

import (
	"iter"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/weightfind/util_weight"
	"paladin_gearing_go/weightfind/weight_highs/fitting2"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
)

const (
	c_fitting3_each_threadCount = 8

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

func (f fitting3EachFields) Stopwatch() *util.Stopwatch {
	return &f.process.Stopwatch
}

func (f fitting3EachFields) Results() iter.Seq[util_weight.FittingInterimResult2] {
	return slices.Values(f.resultSlice)
}

func (fe *FittingEachStatWeightProcess3) Run(stopwatch *util.Stopwatch, cancel util_async.CancelSignal) weight_types.Weight3ExtendedRanged {
	fe.ChooseScaling()
	fe.launchEachNested(cancel)
	weights := fe.BuildResult()
	fe.CalcMetrics(stopwatch)
	return *weights
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
			simResult = util_weight.FittingSimDetail{
				Average:   scaleSim.Apply(average),
				Min:       scaleSim.Apply(detail.Min),
				Max:       scaleSim.Apply(detail.Max),
				StdDev:    scaleSim.Scale * detail.StdDev,
				HasDetail: true,
			}

			// if scaling has flipped min/max
			simResult.FlipMinMaxAsNeeded()
		} else {
			simResult = util_weight.FittingSimDetail{
				Average:   scaleSim.Apply(average),
				HasDetail: false,
			}
		}

		samples[i] = util_weight.FittingSample3{
			StatValue: statValue * scaleStat,
			SimResult: simResult,
		}
	}
	return samples
}

func (fe *FittingEachStatWeightProcess3) launchEachNested(cancel util_async.CancelSignal) {
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

	channelEach := util_async.SeqToChannel_Cancellable(fe.Each.SeqValues(), cancel)
	util_async.ForEach_Channel(c_fitting3_each_threadCount, channelEach, func(fields *fitting3EachFields) {
		initialResultFuture := fields.process.Run()
		initialResult := initialResultFuture.WaitForResultOrPanic()
		fe.rescaleAndCleanup(initialResult, fields)
	})
}

func (fe *FittingEachStatWeightProcess3) rescaleAndCleanup(initialSet fitting2.InitialResultSet, fields *fitting3EachFields) {
	fields.resultSlice = fe.ConvertAndScaleResult(initialSet, fields.statType)
	fields.resultSlice = fe.CleanupRanges(fields.resultSlice)
}
