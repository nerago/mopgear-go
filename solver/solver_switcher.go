package solver

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/solver/build"
	"paladin_gearing_go/solver/channel"
	"paladin_gearing_go/solver/indexed"
	"paladin_gearing_go/solver/phased"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
)

// type SolveInput struct {
// 	model       Model
// 	itemOptions SolvableOptionsMap
// }

const targetCount uint64 = 100_000_000

func Solver(itemOptions *items.FullOptionsMap, model *model.Model, phasedAcceptable bool) SolveOutput {
	printer := util.PrintRecorder_HoldAll()
	solveOptions := items.SolvableOptionsMap_of(itemOptions)

	var mode int
	if phasedAcceptable {
		mode = 8
	} else {
		mode = 6
	}

	// TODO run types and size multipliers
	// TODO trackprogress parameter

	var solvedSet util.Optional[items.SolvableItemSet]
	switch mode {
	// case 1:
	// 	solvedSet = indexed.SolverIndexed_RunFull(&solveOptions, model, printer)
	case 2:
		solvedSet = indexed.SolverIndexed_RunSkipping(&solveOptions, model, targetCount, &printer)
	// case 3:
	// 	solvedSet = channel.SolverChannelBuildFull_Run(&solveOptions, model)
	case 4:
		solvedSet = channel.SolverChannelBuildPeriodic_Run(&solveOptions, model, targetCount, &printer)
	case 5:
		solvedSet = build.SolverBuildPeriodic_Run(&solveOptions, model, targetCount, &printer)
	case 6:
		solvedSet = build.SolverBuildOverflow_Run(&solveOptions, model, targetCount, &printer)
	case 7:
		solvedSet = build.SolverBuildRandom_Run(&solveOptions, model, targetCount, &printer)
	case 8:
		solvedSet = phased.SolverSkinnyPhasedIndex_Run(&solveOptions, model, targetCount, &printer)
	}

	// TODO bury tweaker into find best checks
	solvedSet.MapInPlace(func(set items.SolvableItemSet) items.SolvableItemSet {
		return tools.Tweaker_Run(set, &solveOptions, model)
	})

	return util.Optional_MapAsValueOrEmpty(solvedSet,
		func(set items.SolvableItemSet) SolveOutput {
			return SolveOutput{true, set, items.FullItemSet_FromSolved(set, itemOptions), model.CalcRatingSolve(&set), printer}
		},
		func() SolveOutput {
			return SolveOutput{Success: false, ResultRating: 0, Printer: printer}
		},
	)
}

type SolveOutput struct {
	Success      bool
	SolvedSet    items.SolvableItemSet
	FullSet      items.FullItemSet
	ResultRating uint64
	Printer      util.PrintRecorder
}

func (output *SolveOutput) Report(printer *util.PrintRecorder) {
	printer.AppendOther(&output.Printer)
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
		printer.Println(item.String())
	}
}
