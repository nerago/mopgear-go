package solver

import (
	. "paladin_gearing_go/model"
	"paladin_gearing_go/solver/build"
	"paladin_gearing_go/solver/indexed"
	"paladin_gearing_go/solver/phased"
	. "paladin_gearing_go/types/items"
)

// type SolveInput struct {
// 	model       Model
// 	itemOptions SolvableOptionsMap
// }

const targetCount uint64 = 100_000_000

func Solver(itemOptions *FullOptionsMap, model *Model) FullItemSet {
	solveOptions := SolvableOptionsMap_of(itemOptions)

	var solvedSet SolvableItemSet
	mode := 5
	switch mode {
	case 1:
		solvedSet = indexed.SolverIndexed_RunFull(&solveOptions, model)
	case 2:
		solvedSet = indexed.SolverIndexed_RunSkipping(&solveOptions, model, targetCount)
	case 3:
		solvedSet = build.SolverChannelBuildFull_Run(&solveOptions, model)
	case 4:
		solvedSet = build.SolverChannelBuildPeriodic_Run(&solveOptions, model, targetCount)
	case 5:
		solvedSet = build.SolverBuildPeriodic_Run(&solveOptions, model, targetCount)
	case 6:
		solvedSet = phased.SolverSkinnyPhasedIndex_Run(&solveOptions, model, targetCount)
	}

	// solvedSet = Tweaker_Run(solvedSet, &solveOptions, model)
	return FullItemSet_FromSolved(solvedSet, itemOptions)
}
