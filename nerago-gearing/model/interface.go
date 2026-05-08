package model

import (
	. "paladin_gearing_go/stats"
)

type StatRequirements interface {
	CheckSet(block *StatBlock) bool
}

type StatRatings interface {
	CalcRating(block *StatBlock) uint64
}
