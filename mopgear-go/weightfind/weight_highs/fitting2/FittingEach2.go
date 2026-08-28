package fitting2

import (
	"errors"
	"fmt"
	"iter"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
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
	statType stats.StatType
	simType  stats.SimType

	innerPrinter *util.PrintRecorder
	stopwatch    util.Stopwatch

	resultSlice []util_weight.FittingInterimResult2
}

func (f fitting2EachFields) InnerPrinter() *util.PrintRecorder {
	return f.innerPrinter
}

func (f fitting2EachFields) Stopwatch() *util.Stopwatch {
	return &f.stopwatch
}

func (f fitting2EachFields) Results() iter.Seq[util_weight.FittingInterimResult2] {
	return slices.Values(f.resultSlice)
}

func (fe *FittingEachStatWeightProcess2) Run(cancel util_async.CancelSignal) weight_types.WeightResult3 {
	err := util_async.ChainCancel(cancel, &fe.CancelInternal)
	if err != nil {
		return weight_types.WeightResult3Make(nil, 0, highs.ModelStatusModelError, err)
	}

	fe.chooseScaling()
	err = fe.launchEachNested()
	stopwatch := fe.CalcMetrics()

	if err == nil {
		weight := fe.BuildResult()
		return weight_types.WeightResult3Make(weight, stopwatch.Elapsed(), highs.ModelStatusOptimal, err)
	} else {
		return weight_types.WeightResult3Make(nil, stopwatch.Elapsed(), highs.ModelStatusModelError, err)
	}
}

func (fe *FittingEachStatWeightProcess2) chooseScaling() {
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

func (fe *FittingEachStatWeightProcess2) launchEachNested() error {
	for _, statType := range fe.RequiredStats {
		for _, simType := range fe.RequiredSims {
			innerPrinter := util.PrintRecorder_HoldAll()
			fields := fitting2EachFields{statType: statType, simType: simType, innerPrinter: innerPrinter}
			fe.Each.Put(statType, simType, &fields)
		}
	}

	channelEach := util_async.SeqToChannel_Cancellable(fe.Each.SeqValues(), &fe.CancelInternal)
	return util_async.ForEach_Channel_PassError(c_fitting2_each_threadCount, channelEach, func(fields *fitting2EachFields) error {
		scaleStat := fe.ScaleStats.GetOrPanic(fields.statType)

		process := SingleSegmented2{}
		process.Init(fe.TargetSegmentCount, scaleStat, fields.innerPrinter, fe.Timeout)
		process.SupplyData(fe.prepareSamples(fields.statType, fields.simType))

		initialResultFuture, err := process.Run()
		if err != nil {
			return err
		}

		err = util_async.ChainCancel(&fe.CancelInternal, initialResultFuture)
		if err != nil {
			err = errors.Join(err, fe.CancelInternal.Cancel(), initialResultFuture.Cancel())
			fe.Printer.Printf("FAILED FITTING for %s %s: %v\n", fields.statType.Name(), fields.simType.Name(), err)
			return err
		}

		initialResult, hasResult := initialResultFuture.WaitForResult()
		fields.stopwatch = process.Stopwatch

		if hasResult && len(initialResult.Segments) > 0 {
			fe.rescaleAndCleanup(initialResult, fields)
			return nil
		} else {
			message := fmt.Sprintf("FAILED FITTING for %s %s", fields.statType.Name(), fields.simType.Name())
			fe.Printer.Println(message)
			err = fe.CancelInternal.Cancel()
			return errors.Join(err, errors.New(message))
		}
	})
}

func (fe *FittingEachStatWeightProcess2) rescaleAndCleanup(initialSet InitialResultSet, fields *fitting2EachFields) {
	fields.resultSlice = fe.ConvertAndScaleResult(initialSet, fields.statType)
	fields.resultSlice = fe.CleanupRanges(fields.resultSlice)
}
