package fitting4

import (
	"errors"
	"fmt"
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
	c_fitting4_each_threadCount = 1

	//c_fitting3_minimum_stat_coverage       = 100
	//c_fitting3_permitted_overlap_fix       = 50
	//c_fitting3_number_nice_number_interval = 5
)

type FittingEachStatWeightProcess4 struct {
	fitting2.BaseEachStatProcess[*fitting4EachFields]
	SegmentOnData bool
}

type fitting4EachFields struct {
	statType stats.StatType
	simType  stats.SimType

	innerPrinter *util.PrintRecorder
	stopwatch    util.Stopwatch

	resultSlice []util_weight.FittingInterimResult2
}

func (f fitting4EachFields) InnerPrinter() *util.PrintRecorder {
	return f.innerPrinter
}

func (f fitting4EachFields) Stopwatch() *util.Stopwatch {
	return &f.stopwatch
}

func (f fitting4EachFields) Results() iter.Seq[util_weight.FittingInterimResult2] {
	return slices.Values(f.resultSlice)
}

func (fe *FittingEachStatWeightProcess4) Run(cancel util_async.CancelSignal) weight_types.WeightResult3 {
	err := util_async.ChainCancel(cancel, &fe.CancelInternal)
	if err != nil {
		return weight_types.WeightResult3MakeError(0, err)
	}

	fe.chooseScaling()
	err = fe.launchEachNested()
	stopwatch := fe.CalcMetrics()

	if err == nil {
		weight := fe.BuildResult()
		return weight_types.WeightResult3Make(weight, stopwatch.Elapsed(), highs.ModelStatusOptimal)
	} else {
		return weight_types.WeightResult3MakeError(stopwatch.Elapsed(), err)
	}
}

func (fe *FittingEachStatWeightProcess4) chooseScaling() {
	fe.ScaleStats = util_weight.ChooseStatScalingBasic(fe.InputData, c_fitting4_statScaledMaxValue, true, fe.Printer)
	fe.ScaleSims = util_weight.ChooseSimUnfriendlyUnitScaleAndOffset(fe.InputData, fe.RequiredSims)
}

func (fe *FittingEachStatWeightProcess4) prepareSamples(statType stats.StatType, simType stats.SimType) []util_weight.FittingSample3 {
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

func (fe *FittingEachStatWeightProcess4) launchEachNested() error {
	for _, statType := range fe.RequiredStats {
		for _, simType := range fe.RequiredSims {
			printer := util.PrintRecorder_HoldAll()
			fields := fitting4EachFields{statType: statType, simType: simType, innerPrinter: printer}
			fe.Each.Put(statType, simType, &fields)
		}
	}

	channelEach := util_async.SeqToChannel_Cancellable(fe.Each.SeqValues(), &fe.CancelInternal)
	return util_async.ForEach_Channel_PassError(c_fitting4_each_threadCount, channelEach, func(fields *fitting4EachFields) error {
		process := FittingSingleSegmented4{}
		scaleStat := fe.ScaleStats.GetOrPanic(fields.statType)
		process.Init(fe.TargetSegmentCount, scaleStat, fields.innerPrinter, fe.Timeout)
		process.segmentOnData = fe.SegmentOnData
		process.SupplyData(fe.prepareSamples(fields.statType, fields.simType))

		initialResultFuture := process.Run()

		err1 := util_async.ChainCancel(&fe.CancelInternal, initialResultFuture)
		if err1 != nil {
			err1 = errors.Join(err1, fe.CancelInternal.Cancel(), initialResultFuture.Cancel())
			fe.Printer.Printf("FAILED FITTING for %s %s: %v\n", fields.statType.Name(), fields.simType.Name(), err1)
			return err1
		}

		initialResult, err2 := initialResultFuture.WaitForResultOrError()
		fields.stopwatch = process.Stopwatch

		if err2 == nil && len(initialResult.Segments) > 0 {
			fe.rescaleAndCleanup(initialResult, fields)
			return nil
		} else {
			err2 = errors.Join(err2, fe.CancelInternal.Cancel())
			errOut := fmt.Errorf("FAILED FITTING for %s %s: %w", fields.statType.Name(), fields.simType.Name(), err2)
			fe.Printer.Println(errOut.Error())
			return errOut
		}
	})
}

func (fe *FittingEachStatWeightProcess4) rescaleAndCleanup(initialSet *fitting2.InitialResultSet, fields *fitting4EachFields) {
	fields.resultSlice = fe.ConvertAndScaleResult(initialSet, fields.statType)
	fields.resultSlice = fe.CleanupRanges(fields.resultSlice)
}
