package weight_types

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

// WeightExtended
type WeightExtended struct {
	DetailedWeights   util.MapMap[stats.StatType, stats.SimType, float64]
	SimRatioWeighting stats.SimData // should also include component to bring values to similar size
}

// WeightExtended2Ranged
type WeightExtended2Ranged struct {
	StatWeights    util.MapMapSlice[stats.StatType, stats.SimType, WeightExtendedStatEntry]
	SimMultipliers map[stats.SimType]WeightExtendedSimEntry
}
type WeightExtendedStatEntry struct {
	StatWeight     float64
	RangeMinimum   uint32
	RangeMaximum   uint32
	TotalIfGreater float64
}
type WeightExtendedSimEntry struct {
	// calculated so that range of values is consistent (e.g. 0-100)
	// is offset needed, would give more real values for tmi/death, but would that change result
	Offset   float64
	Scale    float64
	Minimise bool
}

type WeightAlternateSimPriority struct {
	orderedList []AlternateSimPriority
}
type AlternateSimPriority struct {
	SimType                    stats.SimType
	CompromisePermittedPercent float64
}
