package simrank

import (
	"cmp"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func sortGenericSimStatistical[T weight_types.IRankEntry](simType stats.SimType, inputData []T) {
	switch simType {
	case stats.Sim_DPS:
		slices.SortFunc(inputData, func(a, b T) int {
			return compareSimsStatisticalByType(a.GetSimData(), b.GetSimData(), stats.Sim_DPS)
		})
	case stats.Sim_TPS:
		slices.SortFunc(inputData, func(a, b T) int {
			return compareSimsStatisticalByType(a.GetSimData(), b.GetSimData(), stats.Sim_TPS)
		})
	case stats.Sim_DTPS:
		slices.SortFunc(inputData, func(a, b T) int {
			return compareSimsStatisticalByType(b.GetSimData(), a.GetSimData(), stats.Sim_DTPS)
		})
	case stats.Sim_HPS:
		slices.SortFunc(inputData, func(a, b T) int {
			return compareSimsStatisticalByType(a.GetSimData(), b.GetSimData(), stats.Sim_HPS)
		})
	case stats.Sim_TMI:
		slices.SortFunc(inputData, func(a, b T) int {
			return compareSimsStatisticalByType(b.GetSimData(), a.GetSimData(), stats.Sim_TMI)
		})
	case stats.Sim_DEATH:
		// death data never has detail
		slices.SortFunc(inputData, func(a, b T) int {
			return cmp.Compare(b.GetSimData().Get(stats.Sim_DEATH), a.GetSimData().Get(stats.Sim_DEATH))
		})
	default:
		panic("unknown type")
	}
}

func sortGenericSimAverages[T weight_types.IRankEntry](simType stats.SimType, inputData []T) {
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
	default:
		panic("unknown type")
	}
}
