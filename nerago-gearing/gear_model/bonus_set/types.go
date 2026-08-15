package bonus_set

import "paladin_gearing_go/stats"

type BonusByCountFlat [6]float64
type BonusBySim stats.SimTypeMap[float64]
type BonusByCountBySim [6]*BonusBySim
