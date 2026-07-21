package simrank

import (
	"cmp"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
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

func sortAccuracyFast(inputData []*weight_types.AccuracyInfoSimStatRanged, simType stats.SimType) {
	slices.SortFunc(inputData, simSortRangedCompares[simType])
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

func sortAccuracyPrepareFast(inputData []*weight_types.AccuracyInfoPrePrepare, simType stats.SimType) {
	slices.SortFunc(inputData, simSortSimSingledCompares[simType])
}

func sortAccuracyWithDeviation(simType stats.SimType, inputData []*weight_types.AccuracyInfoSimStatRanged) {
	switch simType {
	case stats.Sim_DPS:
		slices.SortFunc(inputData, func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
			return deviationCompareSims(a.DataSim, b.DataSim, stats.Sim_DPS)
		})
	case stats.Sim_TPS:
		slices.SortFunc(inputData, func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
			return deviationCompareSims(a.DataSim, b.DataSim, stats.Sim_TPS)
		})
	case stats.Sim_DTPS:
		slices.SortFunc(inputData, func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
			return deviationCompareSims(b.DataSim, a.DataSim, stats.Sim_DTPS)
		})
	case stats.Sim_HPS:
		slices.SortFunc(inputData, func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
			return deviationCompareSims(a.DataSim, b.DataSim, stats.Sim_HPS)
		})
	case stats.Sim_TMI:
		slices.SortFunc(inputData, func(a, b *weight_types.AccuracyInfoSimStatRanged) int {
			return deviationCompareSims(b.DataSim, a.DataSim, stats.Sim_TMI)
		})
	case stats.Sim_DEATH:
		// death data never has detail
		slices.SortFunc(inputData, simSortRangedCompares[stats.Sim_DEATH])
	}
}

func sortAccuracyPrepareWithDeviation(simType stats.SimType, inputData []*weight_types.AccuracyInfoPrePrepare) {
	switch simType {
	case stats.Sim_DPS:
		slices.SortFunc(inputData, func(a, b *weight_types.AccuracyInfoPrePrepare) int {
			return deviationCompareSims(a.DataSim, b.DataSim, stats.Sim_DPS)
		})
	case stats.Sim_TPS:
		slices.SortFunc(inputData, func(a, b *weight_types.AccuracyInfoPrePrepare) int {
			return deviationCompareSims(a.DataSim, b.DataSim, stats.Sim_TPS)
		})
	case stats.Sim_DTPS:
		slices.SortFunc(inputData, func(a, b *weight_types.AccuracyInfoPrePrepare) int {
			return deviationCompareSims(b.DataSim, a.DataSim, stats.Sim_DTPS)
		})
	case stats.Sim_HPS:
		slices.SortFunc(inputData, func(a, b *weight_types.AccuracyInfoPrePrepare) int {
			return deviationCompareSims(a.DataSim, b.DataSim, stats.Sim_HPS)
		})
	case stats.Sim_TMI:
		slices.SortFunc(inputData, func(a, b *weight_types.AccuracyInfoPrePrepare) int {
			return deviationCompareSims(b.DataSim, a.DataSim, stats.Sim_TMI)
		})
	case stats.Sim_DEATH:
		// death data never has detail
		slices.SortFunc(inputData, simSortSimSingledCompares[stats.Sim_DEATH])
	}
}

func sortGeneral[T weight_types.IRankEntryCommon](simType stats.SimType, inputData []T) {
	switch simType {
	case stats.Sim_DPS:
		slices.SortFunc(inputData, func(a, b T) int {
			return cmp.Compare(a.GetSimData().Get(stats.Sim_DPS), b.GetSimData().Get(stats.Sim_DPS))
		})
	case stats.Sim_TPS:
		slices.SortFunc(inputData, func(a, b T) int {
			return cmp.Compare(a.GetSimData().Get(stats.Sim_TPS), b.GetSimData().Get(stats.Sim_TPS))
		})
	case stats.Sim_DTPS:
		slices.SortFunc(inputData, func(a, b T) int {
			return cmp.Compare(b.GetSimData().Get(stats.Sim_DTPS), a.GetSimData().Get(stats.Sim_DTPS))
		})
	case stats.Sim_HPS:
		slices.SortFunc(inputData, func(a, b T) int {
			return cmp.Compare(a.GetSimData().Get(stats.Sim_HPS), b.GetSimData().Get(stats.Sim_HPS))
		})
	case stats.Sim_TMI:
		slices.SortFunc(inputData, func(a, b T) int {
			return cmp.Compare(b.GetSimData().Get(stats.Sim_TMI), a.GetSimData().Get(stats.Sim_TMI))
		})
	case stats.Sim_DEATH:
		slices.SortFunc(inputData, func(a, b T) int {
			return cmp.Compare(b.GetSimData().Get(stats.Sim_DEATH), a.GetSimData().Get(stats.Sim_DEATH))
		})
	}
}
