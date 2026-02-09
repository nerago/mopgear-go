package solver

import (
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/types/items"
)

type SolveInput struct {
	model       Model
	itemOptions SolvableOptionsMap
}
