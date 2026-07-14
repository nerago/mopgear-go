package simrank

import (
	"cmp"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/weightfind/weight_types"
)

//goland:noinspection DuplicatedCode
var simSortRangedCompares = [6]func(a, b *weight_types.AccuracyInfoSimStatRanged) int{
	func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
		return cmp.Compare(a.DataSim.Get(stats.Sim_DPS), b.DataSim.Get(stats.Sim_DPS))
	},
	func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
		return cmp.Compare(a.DataSim.Get(stats.Sim_TPS), b.DataSim.Get(stats.Sim_TPS))
	},
	func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
		return cmp.Compare(b.DataSim.Get(stats.Sim_DTPS), a.DataSim.Get(stats.Sim_DTPS))
	},
	func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
		return cmp.Compare(a.DataSim.Get(stats.Sim_HPS), b.DataSim.Get(stats.Sim_HPS))
	},
	func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
		return cmp.Compare(b.DataSim.Get(stats.Sim_TMI), a.DataSim.Get(stats.Sim_TMI))
	},
	func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
		return cmp.Compare(b.DataSim.Get(stats.Sim_DEATH), a.DataSim.Get(stats.Sim_DEATH))
	},
}

//goland:noinspection DuplicatedCode
var simSortSimSingledCompares = [6]func(a, b *weight_types.AccuracyInfoPrePrepare) int{
	func(a, b *weight_types.AccuracyInfoPrePrepare) int {
		return cmp.Compare(a.DataSim.Get(stats.Sim_DPS), b.DataSim.Get(stats.Sim_DPS))
	},
	func(a, b *weight_types.AccuracyInfoPrePrepare) int {
		return cmp.Compare(a.DataSim.Get(stats.Sim_TPS), b.DataSim.Get(stats.Sim_TPS))
	},
	func(a, b *weight_types.AccuracyInfoPrePrepare) int {
		return cmp.Compare(b.DataSim.Get(stats.Sim_DTPS), a.DataSim.Get(stats.Sim_DTPS))
	},
	func(a, b *weight_types.AccuracyInfoPrePrepare) int {
		return cmp.Compare(a.DataSim.Get(stats.Sim_HPS), b.DataSim.Get(stats.Sim_HPS))
	},
	func(a, b *weight_types.AccuracyInfoPrePrepare) int {
		return cmp.Compare(b.DataSim.Get(stats.Sim_TMI), a.DataSim.Get(stats.Sim_TMI))
	},
	func(a, b *weight_types.AccuracyInfoPrePrepare) int {
		return cmp.Compare(b.DataSim.Get(stats.Sim_DEATH), a.DataSim.Get(stats.Sim_DEATH))
	},
}
