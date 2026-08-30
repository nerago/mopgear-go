package formula3

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
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

func (proc *FormulaSegmentedProcess) Run(timeout *util_highs.TimeLimitToken) (*util_async.FutureCancellable[weight_types.WeightResult2], error) {

	proc.chooseScaling()

	sect := formulaSection3{process: proc}
	sect.init()
	sect.setMinimumIncludeRate(1.0)
	return sect.run(timeout)

	//return nil, nil
}

func (proc *FormulaSegmentedProcess) chooseScaling() {
	target := c_statMax
	proc.scaleStats = util_weight.ChooseStatScalingBasic(proc.inputData, target, true, proc.printer)
	proc.scaleSims = util_weight.ChooseSimUnfriendlyUnitScaleAndOffset(proc.inputData, proc.requiredSims)
}
