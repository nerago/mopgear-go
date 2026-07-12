package solver

import (
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"

	"github.com/google/uuid"
)

type SolveInput struct {
	ItemOptions         *items.FullOptionsMap
	Model               *gear_model.SpecModel
	EnableTrackProgress bool
	OuterTrackProgress  *util.TrackProgress
	Printer             *util.PrintRecorder
}

func Solver(input SolveInput) SolveOutput {
	printer, trackProgress, solveOptions := prepareSolve(input)
	defer trackProgress.SetDone()

	futureSolvedResult := solve_highs.SingleGearSetMain(&solveOptions, input.Model, printer)
	solvedResult := futureSolvedResult.WaitForResultAsOptional()

	return finaliseSolve(solvedResult, solveOptions, input, printer)
}

func Solver_Lite(itemOptions *items.FullOptionsMap, model *gear_model.SpecModel, printer *util.PrintRecorder) items.FullItemSet {
	solveOptions := items.SolvableOptionsMap_of(itemOptions)
	futureSolvedSet := solve_highs.SingleGearSetMain(&solveOptions, model, printer)
	solvedSet := futureSolvedSet.WaitForResultAsOptional()
	itemSet := items.FullItemSet_FromSolved(solvedSet.GetOrPanic(), itemOptions)
	itemSet.DebugValidate()
	itemSet.ValidateItemRules()
	return itemSet
}

func prepareSolve(input SolveInput) (*util.PrintRecorder, *util.TrackProgress, items.SolvableOptionsMap) {
	printer := input.Printer
	if printer == nil {
		printer = util.PrintRecorder_HoldAll()
	}

	var trackProgress *util.TrackProgress
	if input.OuterTrackProgress != nil {
		trackProgress = input.OuterTrackProgress.NewChild()
	} else if input.EnableTrackProgress {
		trackProgress = util.TrackProgress_Start()
	} else {
		trackProgress = util.TrackProgress_Nop()
	}

	solveOptions := items.SolvableOptionsMap_of(input.ItemOptions)
	return printer, trackProgress, solveOptions
}

func finaliseSolve(solvedResult util.Optional[items.SolvableItemSet], solveOptions items.SolvableOptionsMap, input SolveInput, printer *util.PrintRecorder) SolveOutput {
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
	fullItem.DebugValidate()
	fullItem.ValidateItemRules()

	return SolveOutput{
		Success:      true,
		OutputId:     uuid.NewString(),
		Input:        &input,
		SolvedSet:    solvedSet,
		FullSet:      fullItem,
		ResultRating: input.Model.CalcRatingSolve(&solvedSet),
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
