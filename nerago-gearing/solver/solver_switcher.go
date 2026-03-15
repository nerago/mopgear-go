package solver

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/build"
	"paladin_gearing_go/solver/phased"
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
	printer := input.Printer
	if printer == nil {
		printer = util.PrintRecorder_HoldAll()
	}

	var trackProgress *util.TrackProgress
	if input.OuterTrackProgress != nil {
		trackProgress = input.OuterTrackProgress.MakeNested()
	} else if input.EnableTrackProgress {
		trackProgress = util.TrackProgress_Start()
		defer trackProgress.Stop()
	} else {
		trackProgress = util.TrackProgress_Nop()
	}

	targetCount := uint64(input.SolveSize)
	solveOptions := items.SolvableOptionsMap_of(input.ItemOptions)
	combinationCount := solveOptions.TotalCombinationCount()

	var solvedResult util.Optional[items.SolvableItemSet]
	if combinationCount.IsUint64() && combinationCount.Uint64() < targetCount {
		solvedResult = build.SolverBuildFull_Run(&solveOptions, input.Model, trackProgress, printer)
	} else if input.PhasedAcceptable {
		solvedResult = phased.SolverSkinnyPhasedIndex_Run(&solveOptions, input.Model, targetCount, trackProgress, printer)
	} else {
		solvedResult = build.SolverBuildOverflow2_Run(&solveOptions, input.Model, targetCount, trackProgress, printer)
	}

	if solvedResult.IsEmpty() {
		return SolveOutput{Success: false, Input: &input, ResultRating: 0, Printer: printer}
	}

	solvedSet := solvedResult.GetOrPanic()

	// TODO bury tweaker into find best checks
	solvedSet = tools.Tweaker_Run(solvedSet, &solveOptions, input.Model)

	return SolveOutput{
		true,
		uuid.NewString(),
		&input,
		solvedSet,
		items.FullItemSet_FromSolved(solvedSet, input.ItemOptions),
		input.Model.CalcRatingSolve(&solvedSet),
		printer}
}

type SolveOutput struct {
	Success      bool
	OutputId     string
	Input        *SolveInput
	SolvedSet    items.SolvableItemSet
	FullSet      items.FullItemSet
	ResultRating uint64
	Printer      *util.PrintRecorder
}

func (output *SolveOutput) Equals(b *SolveOutput) bool {
	return output.Success == b.Success && output.ResultRating == b.ResultRating && output.FullSet.Equals(b.FullSet)
}

func (output *SolveOutput) Report(printer *util.PrintRecorder) {
	if printer != output.Printer {
		printer.AppendOther(output.Printer)
	}

	if output.Success {
		fullSet := output.FullSet
		rating := output.ResultRating
		printer.Println(output.OutputId)
		printer.Printf("SET OUTPUT rating %d\n", rating)
		printer.Printf("BONUS %.2f\n", float64(output.Input.Model.SetBonus.CalcAndMultiply(&fullSet.Items, 1000))/1000.0)
		fullSet.PrintStats(printer)
		printEquipMap(&fullSet.Items, printer)
		simulate.WowSimJson_Write(&output.FullSet.Items, output.Input.Model, printer)
	} else {
		printer.Printf("SET SOLVE FAILED\n")
	}
}

func printEquipMap(fullEquipMap *items.FullEquipMap, printer *util.PrintRecorder) {
	for _, item := range fullEquipMap {
		if item != nil {
			printer.Println(item.CreateString())
		}
	}
}
