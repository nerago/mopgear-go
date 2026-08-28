package weightfind

import (
	"sync/atomic"
	"time"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type WeightRatioProcess struct {
	timeoutEach int
	printer     *util.PrintRecorder
	specs       []*WeightSpec
}

func (wrp *WeightRatioProcess) Init(timeoutEach int, printer *util.PrintRecorder) {
	wrp.timeoutEach = timeoutEach
	wrp.printer = printer
}
func (wrp *WeightRatioProcess) AddSpec(spec *WeightSpec) {
	wrp.specs = append(wrp.specs, spec)
}
func (wrp *WeightRatioProcess) AddSpecsFrom(wup *WeightUpdateProcess) {
	wrp.specs = append(wrp.specs, wup.specs...)
}

func (wrp *WeightRatioProcess) Run(cancel util_async.CancelSignal) {
	progress := util.TrackProgress_Start()
	progress.RunOuterTracking(len(wrp.specs))
	defer progress.SetDone()

	summaries := util_async.Map_SliceToSlice_Cancellable(c_ratioThreadCount, wrp.specs, cancel, func(spec **WeightSpec) string {
		return wrp.updateSpecRatio(*spec, progress.NewChild(), cancel)
	})

	for _, summary := range summaries {
		wrp.printer.Println(summary)
	}
}

func (wrp *WeightRatioProcess) updateSpecRatio(spec *WeightSpec, tracker *util.TrackProgress, cancel util_async.CancelSignal) string {
	spec.process.forceSkipSim = true
	err := spec.prepareSimData(util.TrackProgress_Nop(), cancel)
	if err != nil {
		util.GlobalErrorHandler(err)
	}

	sb := util.StringBuild2{}
	sb.WriteString("UPDATE RATIOS FOR ")
	sb.WriteString(spec.Label)
	sb.WriteString("\n")

	sb.WriteString("Baseline Ratio: ")
	spec.Model.SimPriority.AppendString(&sb)
	sb.WriteString("\n")
	sb.WriteString("Baseline ")
	wrp.appendAccuracy(spec, &spec.Model.StatWeights.Weight1, &spec.Model.SimPriority, &sb)

	taskList := make([]func() (weight_types.IWeightResult, error), 0)

	taskList = append(taskList, func() (weight_types.IWeightResult, error) {
		return wrp.runRanking(spec, 32, true, false, true, 5, cancel)
	})
	taskList = append(taskList, func() (weight_types.IWeightResult, error) {
		return wrp.runRanking(spec, 128, true, false, true, 3, cancel)
	})
	taskList = append(taskList, func() (weight_types.IWeightResult, error) {
		return wrp.runRanking(spec, 400, false, true, false, 0, cancel)
	})
	//taskList = append(taskList, func() (weight_types.IWeightResult,error) {
	//	return wrp.runRanking(spec, 1000, false, true, false, 0, cancel)
	//})
	taskList = append(taskList, func() (weight_types.IWeightResult, error) {
		return wrp.runRanking(spec, 2000, false, false, false, 0, cancel)
	})
	taskList = append(taskList, func() (weight_types.IWeightResult, error) {
		return wrp.runSearch(spec, cancel)
	})

	countDone := atomic.Uint64{}
	tracker.RunFromAtomicInt(&countDone, uint64(len(taskList)))
	defer tracker.SetDone()

	type weightAndError struct {
		weight weight_types.IWeightResult
		err    error
	}

	resultList := util_async.Map_SliceToSlice_Cancellable(c_ratioThreadCount, taskList, cancel, func(apply *func() (weight_types.IWeightResult, error)) weightAndError {
		result, err3 := (*apply)()
		countDone.Add(1)
		return weightAndError{result, err3}
	})

	for _, result := range resultList {
		wrp.addReport(spec, result.weight, result.err, &sb)
	}

	return sb.String()
}

func (wrp *WeightRatioProcess) addReport(spec *WeightSpec, weightResult weight_types.IWeightResult, err error, sb *util.StringBuild2) {
	newRatio := weightResult.GetNewRatio()
	if err != nil {
		sb.WriteString("Ratio ")
		sb.WriteString(spec.Label)
		sb.WriteString(" ")
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	} else if newRatio != nil {
		sb.WriteString("Ratio ")
		sb.WriteString(spec.Label)
		newRatio.ScaleForTotalSum(1.0)
		newRatio.AppendString(sb)
		sb.WriteString("\n")

		weights1 := weightResult.AsWeight1(spec.dataAll)
		wrp.appendAccuracy(spec, weights1, newRatio, sb)
	} else {
		sb.WriteString("Ratio ")
		sb.WriteString(spec.Label)
		sb.WriteString(" MISSING\n")
	}
}

func (wrp *WeightRatioProcess) appendAccuracy(spec *WeightSpec, weight *weight_types.Weight1Basic, ratio *weight_types.SimPriorityBasic, sb *util.StringBuild2) {
	newAcc := EvaluateAccuracyBasic(weight, ratio.SimTypes(), ratio, spec.dataAll)
	newAccSt := EvaluateAccuracyStatisticalExtended(weight, ratio.SimTypes(), ratio, spec.dataAll)
	sb.Printf("Accuracy: %f %f\n\n", newAcc, newAccSt)
}

func (wrp *WeightRatioProcess) runRanking(spec *WeightSpec, sampleCount int, mip, allPairs, randPairs bool, randPairCount int, cancel util_async.CancelSignal) (*weight_types.WeightResult1, error) {
	data := util_collection.SliceSampleRandom(spec.dataAll, sampleCount)

	ranking := weight_highs.RankingWeightsRatio30{}
	ranking.AllPairs = allPairs
	ranking.RandPairs = randPairs
	ranking.RandPairCount = randPairCount
	ranking.UseMipCompare = mip
	ranking.Init(wrp.printer, wrp.timeoutEach)
	ranking.SetRequiredStats(spec.statTypes)
	ranking.SetTargetRatios(spec.targetRatio)
	ranking.SupplyData(data)
	weightsFuture, err := ranking.RunSinglePassRaw()
	if err != nil {
		return nil, err
	}
	err = util_async.ChainCancel(cancel, weightsFuture)
	if err != nil {
		return nil, err
	}
	weightResult := weightsFuture.WaitForResultOrNilValue()
	return new(weightResult), nil
}

func (wrp *WeightRatioProcess) runSearch(spec *WeightSpec, outerCancel util_async.CancelSignal) (*weight_types.WeightResult1, error) {
	const c_maxRatioChange = 0.1

	statRange := weight_types.StatRangeFloat{Minimum: -1.0, Maximum: 1.0}

	simRangeMap := stats.SimTypeMap[weight_types.StatRangeFloat]{}
	for statType, existingValue := range spec.targetRatio.SeqTypeValue() {
		simRangeMap.Put(statType, weight_types.StatRangeFloat{
			Minimum: max(existingValue-c_maxRatioChange, 0.0),
			Maximum: min(existingValue+c_maxRatioChange, 1.0),
		})
	}

	innerCancel := util_async.CancelSignal_Make()
	err := util_async.ChainCancel(outerCancel, innerCancel)
	if err != nil {
		return nil, err
	}
	util_async.CancelAfterTimeout(innerCancel, time.Duration(wrp.timeoutEach)*time.Second, wrp.printer)

	ranking := WeightSearcherRatio1{}
	ranking.AccuracyStatistical = true
	ranking.Init(spec.statTypes, spec.simTypes)
	ranking.SupplyData(spec.dataAll)
	ranking.SetStatSimRanges(statRange, simRangeMap)
	weightResult := ranking.Run(innerCancel)
	return new(weightResult), nil
}
