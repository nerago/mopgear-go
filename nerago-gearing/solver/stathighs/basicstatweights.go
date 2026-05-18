package stathighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

type BasicStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimResultStats
	simBase      simulate.SimResultStats
	simData      map[stats.StatType]map[uint32]simulate.SimResultStats

	input           utilhighs.InputBuilder
	colNames        []string
	detailedWeights map[stats.StatType]map[simulate.SimResultType]utilhighs.ColumnIndex
}

func (basic *BasicStatWeightProcess) Init(printer *util.PrintRecorder) {
	basic.printer = printer
	basic.simData = make(map[stats.StatType]map[uint32]simulate.SimResultStats)
}

func (basic *BasicStatWeightProcess) SetBaseline(simBase simulate.SimResultStats) {
	basic.simBase = simBase
}

func (basic *BasicStatWeightProcess) SetTargetRatios(targetRatios simulate.SimResultStats) {
	sum := 0.0
	for _, simType := range g_requiredSims {
		val := targetRatios.Get(simType)
		if val <= 0 {
			panic("missing ratio")
		}
		sum += val
	}
	if !utilhighs.FloatEqualsOne(sum) {
		panic("ratios don't add to one")
	}

	basic.targetRatios = targetRatios
}

func (basic *BasicStatWeightProcess) AddSimData(statType stats.StatType, statValue uint32, sim simulate.SimResultStats) {
	if basic.simData[statType] == nil {
		basic.simData[statType] = make(map[uint32]simulate.SimResultStats)
	}
	basic.simData[statType][statValue] = sim
}

// alternately we could baseline each other with a full array of +100 perumtations etc

func (basic *BasicStatWeightProcess) Run() {
	basic.detailedWeights = make(map[stats.StatType]map[simulate.SimResultType]utilhighs.ColumnIndex)
	for _, statType := range g_requiredStats {
		basic.detailedWeights[statType] = make(map[simulate.SimResultType]utilhighs.ColumnIndex)
		for _, simType := range g_requiredSims {
			basic.detailedWeights[statType][simType] = basic.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
		}
	}

	for _, simType := range g_requiredSims {
		value := basic.targetRatios.Get(simType)
		strengthSetToRatio := utilhighs.ConstraintRowBuild{}
		strengthSetToRatio.Add(basic.detailedWeights[stats.Stat_Strength][simType], 1)
		strengthSetToRatio.Finish(&basic.input, value, value)
	}

	for statType, statDataMap := range basic.simData {
		for statValue, statData := range statDataMap {
			basic.processStat(statType, statValue, statData)
		}
	}
}

func (basic *BasicStatWeightProcess) processStat(statType stats.StatType, statValue uint32, sim simulate.SimResultStats) {
	// for _, simType := range g_requiredSims {
		// simValueDiff := sim.Get(simType) - basic.simBase.Get(simType)

		// so i'd normally do this in the area of a single stat and simType
		// Death's Haste improvement is F37/C37 = (death_with_0_haste_str - death_with_+600_haste)/(death_with_0_haste_str - death_with_+600_str)

		// death_haste_weight = death_inc[haste] / death_inc[str]

		// death_haste_weight * death_inc[str] = death_inc[haste]

		// death_haste_weight * death_inc[str] - death_inc[haste] = 0

		// dps_haste_weight   = dps_inc[haste]   / dps_inc[str]
		// taken_haste_weight = taken_inc[haste] / taken_inc[str]
		// tmi_haste_weight   = tmi_inc[haste]   / tmi_inc[str]
		// death_haste_weight = death_inc[haste] / death_inc[str]
		

		// final_haste = dps_inc[haste]   / dps_inc[str]   * target[dps]
		//             + taken_inc[haste] / taken_inc[str] * target[taken]
		//             + tmi_inc[haste]   / tmi_inc[str]   * target[tmi]
		//             + death_inc[haste] / death_inc[str] * target[death]
		
		// final_haste = dps_inc[haste]   / dps_inc[str]   * weight[str][dps]
		//             + taken_inc[haste] / taken_inc[str] * weight[str][taken]
		//             + tmi_inc[haste]   / tmi_inc[str]   * weight[str][tmi]
		//             + death_inc[haste] / death_inc[str] * weight[str][death]

		// weight[haste][dps]   = dps_inc[haste]   / dps_inc[str] * weight[str][dps]
		// weight[haste][dps] / dps_inc[haste]  =   weight[str][dps] / dps_inc[str]
	// }

	// so whole new approach
	// unit_dps_haste = (this_dps[haste] - base_dps) / this_haste_value
	// stage2weight_dps_haste = unit_dps_haste / unit_dps_str * stage2weight_str
}
