package weightfind

import (
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
	spec.prepareSimData(util.TrackProgress_Nop(), cancel)

	sb := util.StringBuild2{}
	sb.WriteString("UPDATE RATIOS FOR ")
	sb.WriteString(spec.Label)
	sb.WriteString("\n")

	sb.WriteString("Baseline Ratio: ")
	spec.Model.SimPriority.AppendString(&sb)
	sb.WriteString("\n")
	sb.WriteString("Baseline ")
	wrp.appendAccuracy(spec, &spec.Model.StatWeights.Weight1, &spec.Model.SimPriority, &sb)

	taskList := make([]func() *weight_types.WeightResult, 0)

	taskList = append(taskList, func() *weight_types.WeightResult {
		return wrp.runRanking(spec, 32, true, false, true, 5)
	})
	taskList = append(taskList, func() *weight_types.WeightResult {
		return wrp.runRanking(spec, 64, true, false, true, 3)
	})
	taskList = append(taskList, func() *weight_types.WeightResult {
		return wrp.runRanking(spec, 256, false, true, false, 0)
	})
	taskList = append(taskList, func() *weight_types.WeightResult {
		return wrp.runRanking(spec, 1000, false, true, false, 0)
	})
	taskList = append(taskList, func() *weight_types.WeightResult {
		return wrp.runRanking(spec, 1000, false, false, false, 0)
	})

	resultList := util_async.Map_SliceToSlice(c_ratioThreadCount, taskList, func(f *func() *weight_types.WeightResult) *weight_types.WeightResult {
		return (*f)()
	})

	for _, result := range resultList {
		wrp.addReport(spec, result, &sb)
	}

	return sb.String()
}

func (wrp *WeightRatioProcess) addReport(spec *WeightSpec, weightResult *weight_types.WeightResult, sb *util.StringBuild2) {
	newRatio := weightResult.NewRatio
	sb.WriteString("Ratio: ")
	newRatio.AppendString(sb)
	sb.WriteString("\n")

	weights1 := weightResult.AsWeight1()
	wrp.appendAccuracy(spec, weights1, newRatio, sb)
}

func (wrp *WeightRatioProcess) appendAccuracy(spec *WeightSpec, weight *weight_types.Weight1Basic, ratio *weight_types.SimPriorityBasic, sb *util.StringBuild2) {
	newAcc := EvaluateAccuracy(weight, ratio.SimTypes(), ratio, spec.dataAll)
	newAccSt := EvaluateAccuracyStatisticalExtended(weight, ratio.SimTypes(), ratio, spec.dataAll)
	sb.Printf("Accuracy: %f %f\n\n", newAcc, newAccSt)
}

func (wrp *WeightRatioProcess) runRanking(spec *WeightSpec, sampleCount int, mip, allPairs, randPairs bool, randPairCount int) *weight_types.WeightResult {
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
	weightsFuture := ranking.RunSinglePassRaw()
	weightResult := weightsFuture.WaitForResultOrPanic()
	return new(weightResult)
}
