package weight_types

import "paladin_gearing_go/stats"

type WeightType int

type IWeight interface {
	IsEmpty() bool
	CalcStatScore(stats *stats.StatBlock) float64
	String() string
}
