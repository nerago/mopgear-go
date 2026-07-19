package weight_types

import "paladin_gearing_go/stats"

type WeightAlternateSimPriority struct {
	orderedList []AlternateSimPriority
}
type AlternateSimPriority struct {
	SimType                    stats.SimType
	CompromisePermittedPercent float64
}
