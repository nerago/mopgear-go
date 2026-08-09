package weight_types

import (
	"paladin_gearing_go/stats"
)

type WeightType int

type IWeight interface {
	IsEmpty() bool
	CalcStatScore(*stats.StatBlock) float64
	String() string
	CalcStatScoreWithBonus(*stats.StatBlock, *stats.SimTypeMap[float64]) float64
}
