package main

import (
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/solver"
)

func main() {
	model := Model_PallyProtMitigation()
	solvedSet := SolverIndexed_RunFull(itemOptions, &model)
}
