package solver

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/solver/build"
	"paladin_gearing_go/solver/phased"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
)

type SolveSize uint64

const (
	SolveSize_PerItem SolveSize = 100_000
	SolveSize_Medium  SolveSize = 20_000_000
	SolveSize_Long    SolveSize = 100_000_000
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

	solveOptions := items.SolvableOptionsMap_of(input.ItemOptions)
	model := input.Model

	var mode int
	if input.PhasedAcceptable {
		mode = 8
	} else {
		mode = 6
	}

	var targetCount uint64 = uint64(input.SolveSize)

	var solvedSet util.Optional[items.SolvableItemSet]
	switch mode {
	// case 1:
	// 	solvedSet = indexed.SolverIndexed_RunFull(&solveOptions, model, printer)
	// case 2:
	// 	solvedSet = indexed.SolverIndexed_RunSkipping(&solveOptions, model, targetCount, &printer)
	// case 3:
	// 	solvedSet = channel.SolverChannelBuildFull_Run(&solveOptions, model)
	// case 4:
	// 	solvedSet = channel.SolverChannelBuildPeriodic_Run(&solveOptions, model, targetCount, &printer)
	// case 5:
	// 	solvedSet = build.SolverBuildPeriodic_Run(&solveOptions, model, targetCount, &printer)
	case 6:
		solvedSet = build.SolverBuildOverflow_Run(&solveOptions, model, targetCount, trackProgress, printer)
	// case 7:
	// 	solvedSet = build.SolverBuildRandom_Run(&solveOptions, model, targetCount, &printer)
	case 8:
		solvedSet = phased.SolverSkinnyPhasedIndex_Run(&solveOptions, model, targetCount, trackProgress, printer)
	}

	// TODO bury tweaker into find best checks
	solvedSet.MapInPlace(func(set items.SolvableItemSet) items.SolvableItemSet {
		return tools.Tweaker_Run(set, &solveOptions, model)
	})

	return util.Optional_MapAsValueOrEmpty(solvedSet,
		func(set items.SolvableItemSet) SolveOutput {
			return SolveOutput{true, &input, set, items.FullItemSet_FromSolved(set, input.ItemOptions), model.CalcRatingSolve(&set), printer}
		},
		func() SolveOutput {
			return SolveOutput{Success: false, Input: &input, ResultRating: 0, Printer: printer}
		},
	)
}

type SolveOutput struct {
	Success      bool
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
	printer.AppendOther(output.Printer)
	if output.Success {
		fullSet := output.FullSet
		rating := output.ResultRating
		printer.Printf("SET OUTPUT rating %d\n", rating)
		printer.Printf("RATED %s\n", fullSet.TotalRated.String())
		printer.Printf("CAP %s\n", fullSet.TotalCap.String())
		printEquipMap(&fullSet.Items, printer)
		// TODO set bonus
	} else {
		printer.Printf("SET SOLVE FAILED\n")
	}
}

func printEquipMap(fullEquipMap *items.FullEquipMap, printer *util.PrintRecorder) {
	for _, item := range fullEquipMap {
		if item != nil {
			printer.Println(item.String())
		}
	}
}
