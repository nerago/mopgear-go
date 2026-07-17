package gear_model

import (
	. "paladin_gearing_go/stats"
)

type StatRequirements interface {
	CheckSet(block *StatBlock) bool
	Equals(other any) bool
	IsLow(statType StatType, value uint32) bool
	IsHigh(statType StatType, value uint32) bool
}

type StatRatings interface {
	CalcRating(block *StatBlock) float64
	CreateString() string
}
