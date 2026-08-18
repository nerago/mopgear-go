package bonus_set

import "github.com/nerago/mopgear-go/stats"

type BonusByCountFlat [6]float64
type BonusBySim stats.SimTypeMap[float64]
type BonusByCountBySim [6]*BonusBySim
