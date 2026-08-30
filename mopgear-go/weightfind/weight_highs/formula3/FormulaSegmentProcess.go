package formula3

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type FormulaSegmentedProcess struct {
	printer *util.PrintRecorder

	targetRatios  weight_types.SimPriorityBasic
	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	scaleSims     stats.SimTypeMap[util_weight.ScaleAndOffset]
	scaleStats    stats.StatTypeMap[float64]

	inputData []weight_types.WeightInput
	UseStdDev bool
}

// TODO at least one sample each on two arbitrary threshold "edges" passes through sample?
// or we could go back on another run to recalculate with meeting lines

func (proc *FormulaSegmentedProcess) Init(printer *util.PrintRecorder) {
	proc.printer = printer
}

func (proc *FormulaSegmentedProcess) SupplyData(inputData []weight_types.WeightInput) {
	proc.inputData = inputData
}

func (proc *FormulaSegmentedProcess) SetRequiredStats(requiredStats []stats.StatType) {
	proc.requiredStats = requiredStats
}

func (proc *FormulaSegmentedProcess) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	proc.targetRatios = targetRatios
	proc.requiredSims = targetRatios.SimTypes()
}

func (proc *FormulaSegmentedProcess) Run(timeout *util_highs.TimeLimitToken) (*util_async.FutureCancellable[weight_types.WeightResult4], error) {
	proc.chooseScaling()

	futureResult := util_async.FutureCancellable_Make[weight_types.WeightResult4]()

	go func() {
		sect := formulaSection3{process: proc}
		sect.init()
		sect.setMinimumIncludeRate(0.3)
		resultFuture := sect.run(timeout)
		if result, err := resultFuture.WaitForResultPointerOrError(); err == nil {
			proc.reportOnDataAvailable(result)
			weight4 := proc.buildWeight4(result)
			futureResult.SetResult(weight_types.WeightResult4Make(weight4, result.elapsed, result.status))
		} else {
			futureResult.SetResult(weight_types.WeightResult4MakeError(result.elapsed, err))
		}
	}()

	return futureResult, nil
}

func (proc *FormulaSegmentedProcess) chooseScaling() {
	target := c_statMax
	proc.scaleStats = util_weight.ChooseStatScalingBasic(proc.inputData, target, true, proc.printer)
	proc.scaleSims = util_weight.ChooseSimUnfriendlyUnitScaleAndOffset(proc.inputData, proc.requiredSims)
}

func (proc *FormulaSegmentedProcess) reportOnDataAvailable(result *sectionResult) {
	for statType, statRange := range result.bounds.SeqKeyValue() {
		lo, inc, hi := 0, 0, 0
		for input := range util_collection.ForPointer(proc.inputData) {
			value := input.TotalStat.GetUInt(statType)
			if statRange.Contains(value) {
				inc++
			} else if value < statRange.Minimum {
				lo++
			} else {
				hi++
			}
		}
		proc.printer.Printf("DATA %7s lo=%3d inc=%3d hi=%3d rate=%f\n", statType.Name(), lo, inc, hi, float64(inc)/float64(len(proc.inputData))*100)
	}
}

func (proc *FormulaSegmentedProcess) buildWeight4(result *sectionResult) *weight_types.Weight4Segmented {
	weight4 := weight_types.Weight4Segmented_Make(proc.requiredStats, proc.requiredSims, proc.targetRatios)
	weight4.AddWeight2AsSegment(&result.weights, (*stats.StatTypeMap[weight_types.StatRange])(&result.bounds))
	return weight4
}
