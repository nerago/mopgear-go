package solver

import (
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/solver/solve_highs"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/google/uuid"
)

func Solver(itemOptions *items.FullOptionsMap, model *gear_model.SpecModel, printer *util.PrintRecorder, weightType weight_types.WeightType, timeout int, cancel util_async.CancelSignal) SolveOutput {
	if itemOptions == nil || model == nil || printer == nil || weightType == 0 {
		panic("missing option")
	}

	solveOptions := items.SolvableOptionsMap_of(itemOptions)
	solveModel := solve_highs_types.SolverModelBuild(model, weightType, nil)

	futureSolvedSet := LaunchSolve(&solveOptions, solveModel, printer, weightType, timeout)
	if cancel != nil {
		util_async.ChainCancel(cancel, futureSolvedSet)
	}
	solvedResult := futureSolvedSet.WaitForResultAsOptional()

	return finaliseSolve(solvedResult, solveOptions, itemOptions, model, printer, weightType)
}

func LaunchSolve(solveOptions *items.SolvableOptionsMap, solveModel *solve_highs_types.SolverModel, printer *util.PrintRecorder, weightType weight_types.WeightType, timeout int) *util_async.FutureCancellable[items.SolvableItemSet] {
	switch weightType {
	case 1:
		return solve_highs.SingleGearSetMain(solveOptions, solveModel, printer, timeout)
	case 2:
		return solve_highs.SingleGearSetExtended2Main(solveOptions, solveModel, printer, timeout)
	case 3:
		return solve_highs.SingleGearSetExtended3Main(solveOptions, solveModel, printer, timeout)
	default:
		panic("invalid weight type")
	}
}

func finaliseSolve(solvedResult util_collection.Optional[items.SolvableItemSet], solveOptions items.SolvableOptionsMap, itemOptions *items.FullOptionsMap, model *gear_model.SpecModel, printer *util.PrintRecorder, weightType weight_types.WeightType) SolveOutput {
	var solvedSet items.SolvableItemSet
	if solvedResult.IsEmpty() {
		fallbackSet, failureSummary := diagnoseFailure(&solveOptions, model)
		if fallbackSet.IsEmpty() {
			return SolveOutput{Success: false, OutputId: uuid.NewString(), Model: model, ResultRating: 0, FailureSummary: failureSummary, Printer: printer}
		} else {
			printer.Println("USING FALLBACK CAPPED SET!!")
			solvedSet = fallbackSet.GetOrPanic()
		}
	} else {
		solvedSet = solvedResult.GetOrPanic()
	}

	solvedSet.DebugValidate()

	// solvedSet = tools.Tweaker_Run(&solvedSet, &solveOptions, input.Model)
	// solvedSet.DebugValidate()

	fullItem := items.FullItemSet_FromSolved(solvedSet, itemOptions)
	model.ValidateSet(&fullItem)

	rating := model.CalcRatingSolve(&solvedSet, weightType)

	return SolveOutput{
		Success:      true,
		OutputId:     uuid.NewString(),
		Model:        model,
		SolvedSet:    solvedSet,
		FullSet:      fullItem,
		ResultRating: rating,
		Printer:      printer}
}

type SolveOutput struct {
	Success        bool
	OutputId       string
	SolvedSet      items.SolvableItemSet
	FullSet        items.FullItemSet
	ResultRating   float64
	FailureSummary string
	Model          *gear_model.SpecModel
	Printer        *util.PrintRecorder
}

func (output *SolveOutput) Equals(b *SolveOutput) bool {
	return output.Success == b.Success && output.ResultRating == b.ResultRating && output.FullSet.Equals(&b.FullSet)
}

func (output *SolveOutput) Report(printer *util.PrintRecorder) {
	reportPrint := util.PrintRecorder_HoldAll()
	if printer != output.Printer {
		reportPrint.AppendOther(output.Printer)
	}

	if output.Success {
		reportPrint.Println(output.OutputId)
		tools.ReportSet(output.Model, &output.FullSet, reportPrint)
	} else {
		reportPrint.Printf("SET SOLVE FAILED\n")
	}

	printer.AppendOther(reportPrint)
}
