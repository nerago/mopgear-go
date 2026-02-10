package solver

import (
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/types/items"
)

// type SolveInput struct {
// 	model       Model
// 	itemOptions SolvableOptionsMap
// }

func Solver(itemOptions *FullOptionsMap, model *Model) FullItemSet {
	solveOptions := SolvableOptionsMap_of(itemOptions)
	solvedSet := SolverIndexed_RunFull(&solveOptions, model)
	return FullItemSet_FromSolved(solvedSet, itemOptions)
}
