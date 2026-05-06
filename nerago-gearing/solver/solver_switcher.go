package solver

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/build"
	"paladin_gearing_go/solver/phased"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"

	"github.com/google/uuid"
)

type SolveSize uint64

const (
	SolveSize_PerItem SolveSize = 100_000
	SolveSize_Medium  SolveSize = 20_000_000
	SolveSize_Long    SolveSize = 1_000_000_000
)

type SolveInput struct {
	ItemOptions         *items.FullOptionsMap
	Model               *model.Model
	PhasedAcceptable    bool
	EnableTrackProgress bool
	OuterTrackProgress  *util.TrackProgress
	SolveSize           SolveSize
	Printer             *util.PrintRecorder
}

func Solver(input SolveInput) SolveOutput {
	printer, trackProgress, solveOptions := prepareSolve(input)
	defer trackProgress.Stop()

	targetCount := uint64(input.SolveSize)
	combinationCount := solveOptions.TotalCombinationCount()

	var solvedResult util.Optional[items.SolvableItemSet]
	if combinationCount.IsUint64() && combinationCount.Uint64() < targetCount {
		solvedResult = build.SolverBuildFull_Run(&solveOptions, input.Model, trackProgress, printer)
	} else if input.PhasedAcceptable {
		solvedResult = phased.SolverSkinnyPhasedIndex_Run(&solveOptions, input.Model, targetCount, trackProgress, printer)
	} else {
		solvedResult = build.SolverBuildOverflow2_Run(&solveOptions, input.Model, targetCount, trackProgress, printer)
		if solvedResult.IsEmpty() {
			printer.Println("Initial Failure with Overflow, trying with phased!!")
			solvedResult = phased.SolverSkinnyPhasedIndex_Run(&solveOptions, input.Model, targetCount, trackProgress, printer)
		}
	}

	return finaliseSolve(solvedResult, solveOptions, input, printer)
}

func Solver_WithHighs(input SolveInput) SolveOutput {
	printer, trackProgress, solveOptions := prepareSolve(input)
	defer trackProgress.Stop()

	var solvedResult util.Optional[items.SolvableItemSet]
	solvedResult = withhighs.RunAllActiveSets(&solveOptions, input.Model, printer)

	return finaliseSolve(solvedResult, solveOptions, input, printer)
}

func prepareSolve(input SolveInput) (*util.PrintRecorder, *util.TrackProgress, items.SolvableOptionsMap) {
	printer := input.Printer
	if printer == nil {
		printer = util.PrintRecorder_HoldAll()
	}

	var trackProgress *util.TrackProgress
	if input.OuterTrackProgress != nil {
		trackProgress = input.OuterTrackProgress.MakeNested()
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

	solvedSet = tools.Tweaker_Run(&solvedSet, &solveOptions, input.Model)
	solvedSet.DebugValidate()

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
	ResultRating   uint64
	FailureSummary string
	Printer        *util.PrintRecorder
}

func (output *SolveOutput) Equals(b *SolveOutput) bool {
	return output.Success == b.Success && output.ResultRating == b.ResultRating && output.FullSet.Equals(&b.FullSet)
}

func (output *SolveOutput) Report(printer *util.PrintRecorder) {
	if printer != output.Printer {
		printer.AppendOther(output.Printer)
	}

	if output.Success {
		fullSet := output.FullSet
		rating := output.ResultRating
		printer.Println(output.OutputId)
		ReportSet(printer, fullSet, rating, output.Input.Model)
	} else {
		printer.Printf("SET SOLVE FAILED\n")
	}
}

func ReportSet(printer *util.PrintRecorder, fullSet items.FullItemSet, rating uint64, modelObj *model.Model) {
	printer.Printf("SET rating %d\n", rating)
	printer.Printf("BONUS counts %s\n", model.AllBonusesText(fullSet.Items()))
	printer.Printf("BONUS multiply %f\n", modelObj.SetBonus.CalcBonusFull(fullSet.Items()))
	fullSet.PrintStats(printer)
	printEquipMap(fullSet.Items(), printer)

	simulate.WowSimJson_Write(fullSet.Items(), modelObj, printer)

	fullSet.DebugValidate()
}

func printEquipMap(fullEquipMap *items.FullEquipMap, printer *util.PrintRecorder) {
	for _, item := range fullEquipMap {
		if item != nil {
			printer.Println(item.CreateString())
		}
	}
}
