package solver

import (
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/weightfind/weight_types"

	"github.com/google/uuid"
)

type SolveInput struct {
	ItemOptions *items.FullOptionsMap
	Model       *gear_model.SpecModel
	WeightType  weight_types.WeightType
	Printer     *util.PrintRecorder
}

func Solver(input SolveInput) SolveOutput {
	printer, solveOptions := prepareSolve(input)
	solveModel := solve_highs.SolverModelBuild(input.Model, input.WeightType)

	futureSolvedSet := launchSolve(solveOptions, solveModel, printer, input.WeightType)
	solvedResult := futureSolvedSet.WaitForResultAsOptional()

	return finaliseSolve(solvedResult, solveOptions, input, printer)
}

func launchSolve(solveOptions items.SolvableOptionsMap, solveModel *solve_highs.SolverModel, printer *util.PrintRecorder, weightType weight_types.WeightType) *util_async.FutureCancellable[items.SolvableItemSet] {
	switch weightType {
	case 1:
		return solve_highs.SingleGearSetMain(&solveOptions, solveModel, printer)
	case 2:
		return solve_highs.SingleGearSetExtended2Main(&solveOptions, solveModel, printer)
	case 3:
		return solve_highs.SingleGearSetExtended3Main(&solveOptions, solveModel, printer)
	default:
		panic("invalid weight type")
	}
}

func prepareSolve(input SolveInput) (*util.PrintRecorder, items.SolvableOptionsMap) {
	printer := input.Printer
	if printer == nil {
		printer = util.PrintRecorder_HoldAll()
	}

	solveOptions := items.SolvableOptionsMap_of(input.ItemOptions)
	return printer, solveOptions
}

func finaliseSolve(solvedResult util_collection.Optional[items.SolvableItemSet], solveOptions items.SolvableOptionsMap, input SolveInput, printer *util.PrintRecorder) SolveOutput {
	var solvedSet items.SolvableItemSet
	if solvedResult.IsEmpty() {
		fallbackSet, failureSummary := diagnoseFailure(&solveOptions, input.Model)
		if fallbackSet.IsEmpty() {
			return SolveOutput{Success: false, OutputId: uuid.NewString(), Input: &input, ResultRating: 0, FailureSummary: failureSummary, Printer: printer}
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

	fullItem := items.FullItemSet_FromSolved(solvedSet, input.ItemOptions)
	input.Model.ValidateSet(&fullItem)

	rating := input.Model.CalcRatingSolve(&solvedSet, input.WeightType)

	return SolveOutput{
		Success:      true,
		OutputId:     uuid.NewString(),
		Input:        &input,
		SolvedSet:    solvedSet,
		FullSet:      fullItem,
		ResultRating: rating,
		Printer:      printer}
}

type SolveOutput struct {
	Success        bool
	OutputId       string
	Input          *SolveInput
	SolvedSet      items.SolvableItemSet
	FullSet        items.FullItemSet
	ResultRating   float64
	FailureSummary string
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
		fullSet := output.FullSet
		rating := output.ResultRating
		reportPrint.Println(output.OutputId)
		tools.ReportSet(output.Input.Model, &fullSet, rating, reportPrint)
	} else {
		reportPrint.Printf("SET SOLVE FAILED\n")
	}

	printer.AppendOther(reportPrint)
}
