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

func Solver(itemOptions *items.FullOptionsMap, model *model.Model, printer *util.PrintRecorder) items.FullItemSet {
	solveOptions := items.SolvableOptionsMap_of(itemOptions)

	var solvedSet items.SolvableItemSet
	mode := 2
	switch mode {
	// case 1:
	// 	solvedSet = indexed.SolverIndexed_RunFull(&solveOptions, model, printer)
	case 2:
		solvedSet = indexed.SolverIndexed_RunSkipping(&solveOptions, model, targetCount, printer)
	// case 3:
	// 	solvedSet = channel.SolverChannelBuildFull_Run(&solveOptions, model)
	case 4:
		solvedSet = channel.SolverChannelBuildPeriodic_Run(&solveOptions, model, targetCount, printer)
	case 5:
		solvedSet = build.SolverBuildPeriodic_Run(&solveOptions, model, targetCount, printer)
	case 6:
		solvedSet = build.SolverBuildOverflow_Run(&solveOptions, model, targetCount, printer)
	case 7:
		solvedSet = build.SolverBuildRandom_Run(&solveOptions, model, targetCount, printer)
	case 8:
		solvedSet = phased.SolverSkinnyPhasedIndex_Run(&solveOptions, model, targetCount, printer)
	}

	// TODO bury tweaker into find best checks
	solvedSet = tools.Tweaker_Run(solvedSet, &solveOptions, model)
	return items.FullItemSet_FromSolved(solvedSet, itemOptions)
}
