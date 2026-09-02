package formula3

import (
	"math"
	"time"

	"github.com/bartolsthoorn/gohighs/highs"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/util/util_rank"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

const c_collapseRangePercent = 0.05
const c_initialIncludePercent = 0.3
const c_otherPrimaryIncludePercent = 0.1
const c_otherPrimaryIncludePercentMax = 0.8

type FormulaSegmentedProcess3 struct {
	printer *util.PrintRecorder

	targetRatios  weight_types.SimPriorityBasic
	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	scaleSims     stats.SimTypeMap[util_weight.ScaleAndOffset]
	scaleStats    stats.StatTypeMap[float64]
	inputData     []weight_types.WeightInput
	UseStdDev     bool

	statInfoMap stats.StatTypeMap[*statInfo]
	primaryInfo *statInfo
	allSegments []*sectionResult
	elapsed     time.Duration
	complete    bool

	timeoutToken  *util_highs.TimeLimitToken
	innerCancel   util_async.CancelSignal
	processFuture *util_async.FutureCancellable[weight_types.WeightResult4]
}

// TODO at least one sample each on two arbitrary threshold "edges" passes through sample?
// or we could go back on another run to recalculate with meeting lines

func (proc *FormulaSegmentedProcess3) Init(printer *util.PrintRecorder) {
	proc.printer = printer
	proc.innerCancel = util_async.CancelSignal_Make()
}

func (proc *FormulaSegmentedProcess3) SupplyData(inputData []weight_types.WeightInput) {
	proc.inputData = inputData
}

func (proc *FormulaSegmentedProcess3) SetRequiredStats(requiredStats []stats.StatType) {
	proc.requiredStats = requiredStats

}

func (proc *FormulaSegmentedProcess3) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	proc.targetRatios = targetRatios
	proc.requiredSims = targetRatios.SimTypes()
}

func (proc *FormulaSegmentedProcess3) Run(timeout *util_highs.TimeLimitToken, outerCancel util_async.CancelSignal) (*util_async.FutureCancellable[weight_types.WeightResult4], error) {
	proc.timeoutToken = timeout
	err := util_async.ChainCancel(outerCancel, proc.innerCancel)
	if err != nil {
		return nil, err
	}

	proc.chooseScaling()

	proc.processFuture = util_async.FutureCancellable_Make[weight_types.WeightResult4]()

	go proc.runThread()

	return proc.processFuture, nil
}

func (proc *FormulaSegmentedProcess3) runThread() {
	err := proc.runInitialSection()
	if err != nil {
		proc.sendErrorResult(err)
		return
	}

	if proc.complete {
		weight4 := proc.buildWeight4FromSingle()
		proc.sendSuccessResult(weight4)
	} else {
		err = proc.primaryCompletionLoop()
		if err != nil {
			proc.sendErrorResult(err)
			return
		}
	}
}

func (proc *FormulaSegmentedProcess3) runInitialSection() error {
	sect := formulaSection3{process: proc}
	sect.init()
	sect.setMinimumIncludeRate(c_initialIncludePercent)

	sectionFuture, err := sect.runSection(proc.timeoutToken)
	if err != nil {
		return err
	}

	if err := util_async.ChainCancel(proc.innerCancel, sectionFuture); err != nil {
		return err
	}

	sectResult, hasResult := sectionFuture.WaitForResult()
	if !hasResult {
		return util.ErrorTracedNew("missing section result")
	} else if sectResult.err != nil {
		return sectResult.err
	}

	err = proc.processInitialResult(&sectResult)
	return err
}

func (proc *FormulaSegmentedProcess3) sendSuccessResult(weight4 *weight_types.Weight4Segmented) {
	proc.processFuture.SetResult(weight_types.WeightResult4Make(weight4, proc.elapsed, highs.ModelStatusOptimal))

}

func (proc *FormulaSegmentedProcess3) sendErrorResult(err error) {
	proc.processFuture.SetResult(weight_types.WeightResult4MakeError(proc.elapsed, err))
}

func (proc *FormulaSegmentedProcess3) chooseScaling() {
	target := c_statMax
	proc.scaleStats = util_weight.ChooseStatScalingBasic(proc.inputData, target, true, proc.printer)
	proc.scaleSims = util_weight.ChooseSimUnfriendlyUnitScaleAndOffset(proc.inputData, proc.requiredSims)
}

// todo would be nice to have intercept data for later
func (proc *FormulaSegmentedProcess3) processInitialResult(result *sectionResult) error {
	proc.elapsed += result.elapsed
	proc.allSegments = append(proc.allSegments, result)

	smallestUndecidedInclude, allFullRange := proc.setupInitialStatInfo(result)

	if allFullRange {
		proc.complete = true
		return nil
	} else if smallestUsedStat, hasSmallest := smallestUndecidedInclude.GetBest(); hasSmallest {
		info := proc.statInfoMap.GetOrPanic(smallestUsedStat)
		info.status = statusPrimaryBreakpoints
		proc.primaryInfo = info
		return nil
	} else {
		return util.ErrorTracedNew("unexpected result: neither full range or smallestUndecided set")
	}
}

func (proc *FormulaSegmentedProcess3) setupInitialStatInfo(result *sectionResult) (util_rank.BestCollector1Lite[stats.StatType], bool) {
	smallestUndecidedInclude := util_rank.BestCollector1Lite[stats.StatType]{}
	allFullRange := true

	for statType, statRange := range result.bounds.SeqKeyValue() {
		inputDataCount := float64(len(proc.inputData))
		loCount, includeCount, hiCount := proc.countDataInRange(statType, statRange)
		proc.printer.Printf("DATA %7s lo=%3d inc=%3d hi=%3d rate=%f\n", statType.Name(), loCount, includeCount, hiCount, float64(includeCount)/inputDataCount*100)

		info := &statInfo{
			statType:    statType,
			status:      statusUndecided,
			sections:    []*sectionResult{result},
			usedRange:   statRange,
			loPercent:   float64(loCount) / inputDataCount,
			usedPercent: float64(includeCount) / inputDataCount,
			hiPercent:   float64(hiCount) / inputDataCount,
		}
		info.initialDerives()

		if info.status == statusUndecided {
			smallestUndecidedInclude.Offer(statType, info.usedPercent)
		}
		if info.status != statusFullRange {
			allFullRange = false
		}

		proc.statInfoMap.Put(statType, info)
	}

	return smallestUndecidedInclude, allFullRange
}

func (info *statInfo) initialDerives() {
	if info.loPercent < c_collapseRangePercent {
		info.usedRange.Minimum = 0
		info.loPercent = 0
	}
	if info.hiPercent < c_collapseRangePercent {
		info.usedRange.Maximum = math.MaxUint32
		info.hiPercent = 0
	}

	if info.usedRange.IsFullRange() {
		info.status = statusFullRange
	}
}

func (proc *FormulaSegmentedProcess3) countDataInRange(statType stats.StatType, statRange weight_types.StatRange) (int, int, int) {
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
	return lo, inc, hi
}

func (proc *FormulaSegmentedProcess3) buildWeight4FromSingle() *weight_types.Weight4Segmented {
	weight4 := weight_types.Weight4Segmented_Make(proc.requiredStats, proc.requiredSims, proc.targetRatios)
	for _, segment := range proc.allSegments {
		weight4.AddWeight2AsSegment(&segment.weights, &segment.bounds)
	}
	return weight4
}

func (proc *FormulaSegmentedProcess3) primaryCompletionLoop() error {
	for !proc.complete {
		tasks := make([]*util_async.FutureCancellable[sectionResult], 0, 10)

		primary := proc.primaryInfo
		if primary.hiPercent > 0 {
			bound := boundsSingleGreaterThan(primary.statType, primary.usedRange.Maximum)
			future, err := proc.startNextSection(bound, true, false, primary.hiPercent)
			if err != nil {
				return err
			}
			tasks = append(tasks, future)
		}

	}
	return nil
}

func (proc *FormulaSegmentedProcess3) startNextSection(bound *weight_types.Weight4SegmentBound, forceLo, forceHi bool, totalPercentAvailable float64) (*util_async.FutureCancellable[sectionResult], error) {
	sect := formulaSection3{process: proc}
	sect.init()
	sect.setFilterBounds(bound)
	sect.setForceThreshold(forceLo, forceHi)

	includeRate := 0.0
	if totalPercentAvailable > 0.60 {
		includeRate = c_initialIncludePercent
	} else if totalPercentAvailable > c_otherPrimaryIncludePercent {
		includeRate = util.Clamp(c_otherPrimaryIncludePercent/totalPercentAvailable, c_otherPrimaryIncludePercent, c_otherPrimaryIncludePercentMax)
	} else if totalPercentAvailable > c_collapseRangePercent {
		includeRate = 1.00
	} else {
		return nil, util.ErrorTracedNew("small range should already be collapsed")
	}
	sect.setMinimumIncludeRate(includeRate)

	future, err := sect.runSection(proc.timeoutToken)
	if err == nil {
		err = util_async.ChainCancel(proc.innerCancel, future)
	}
	return future, err
}
