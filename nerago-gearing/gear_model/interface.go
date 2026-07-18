package gear_model

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

type StatRequirements interface {
	CheckSet(block *stats.StatBlock) bool
	Equals(other any) bool
	IsLow(statType stats.StatType, value uint32) bool
	IsHigh(statType stats.StatType, value uint32) bool
	AsMap() map[stats.StatType]util.HiLoInt
}

type StatRatings interface {
	CalcRating(block *stats.StatBlock) float64
	CreateString() string
}
