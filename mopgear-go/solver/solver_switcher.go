package solver

import (
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/solver/solve_highs"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func Solver(itemOptions *items.FullOptionsMap, model *gear_model.SpecModel, printer *util.PrintRecorder, weightType weight_types.WeightType, timeout int, cancel util_async.CancelSignal) SolveOutput {
	if itemOptions == nil || model == nil || printer == nil || weightType == 0 {
		return SolveOutput{Success: false, Error: util.ErrorTracedNew("solver: missing parameter")}
	}

	solveOptions := items.SolvableOptionsMap_of(itemOptions)
	solveModel := solve_highs_types.SolverModelBuild(model, weightType, nil)

	futureSolvedSet, err := LaunchSolve(&solveOptions, solveModel, printer, weightType, timeout)
	if err != nil {
		return SolveOutput{Success: false, Error: err}
	}

	if cancel != nil {
		err = util_async.ChainCancel(cancel, futureSolvedSet)
		if err != nil {
			return SolveOutput{Success: false, Error: err}
		}
	}

	solvedResult, err := futureSolvedSet.WaitForResultOrError()
	if err != nil {
		return SolveOutput{Success: false, Error: err}
	}

	return finaliseSolve(solvedResult, itemOptions, model, weightType)
}

func LaunchSolve(solveOptions *items.SolvableOptionsMap, solveModel *solve_highs_types.SolverModel, printer *util.PrintRecorder, weightType weight_types.WeightType, timeout int) (*util_async.FutureCancellableWithError[items.SolvableItemSet], error) {
	switch weightType {
	case 1:
		return solve_highs.SingleGearSetMain(solveOptions, solveModel, printer, timeout)
	case 2:
		return solve_highs.SingleGearSetExtended2Main(solveOptions, solveModel, printer, timeout)
	case 3:
		return solve_highs.SingleGearSetExtended3Main(solveOptions, solveModel, printer, timeout)
	default:
		return nil, util.ErrorTracedNew("invalid weight type")
	}
}

func finaliseSolve(solvedSet items.SolvableItemSet, itemOptions *items.FullOptionsMap, model *gear_model.SpecModel, weightType weight_types.WeightType) SolveOutput {
	if err := solvedSet.DebugValidate(); err != nil {
		return SolveOutput{Success: false, Error: err}
	}

	fullSet, err := items.FullItemSet_FromSolved(solvedSet, itemOptions)
	if err != nil {
		return SolveOutput{Success: false, Error: err}
	}

	if err := model.ValidateSet(&fullSet); err != nil {
		return SolveOutput{Success: false, Error: err}
	}

	rating := model.CalcRatingFull(&fullSet, weightType)

	return SolveOutput{Success: true, SolvedSet: solvedSet, FullSet: fullSet, Model: model, ResultRating: rating}
}

type SolveOutput struct {
	Success      bool
	Error        error
	SolvedSet    items.SolvableItemSet
	FullSet      items.FullItemSet
	Model        *gear_model.SpecModel
	ResultRating float64
}

func (output *SolveOutput) Equals(b *SolveOutput) bool {
	return output.Success == b.Success && output.ResultRating == b.ResultRating && output.FullSet.Equals(&b.FullSet)
}

func (output *SolveOutput) Report(printer *util.PrintRecorder) {
	reportPrint := util.PrintRecorder_HoldAll()

	if output.Success {
		tools.ReportSet(output.Model, &output.FullSet, reportPrint)
	} else {
		reportPrint.Printf("SET SOLVE FAILED: %s\n", output.Error.Error())
	}

	printer.AppendOther(reportPrint)
}
