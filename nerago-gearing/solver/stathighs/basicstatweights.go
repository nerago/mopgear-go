package stathighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	rangeHigh = 100.0
)

type BasicStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimResultStats
	simBase      simulate.SimResultStats
	simData      util.MapMap[stats.StatType, uint32, simulate.SimResultStats]

	input           utilhighs.InputBuilder
	colNames        []string
	unitStatValue   util.MapMap[stats.StatType, simulate.SimResultType, utilhighs.ColumnIndex]
	detailedWeights util.MapMap[stats.StatType, simulate.SimResultType, utilhighs.ColumnIndex]
}

func (basic *BasicStatWeightProcess) Init(printer *util.PrintRecorder) {
	basic.printer = printer
	basic.input.Minimise = true
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
	basic.simData.Put(statType, statValue, sim)
}

// alternately we could baseline each other with a full array of +100 perumtations etc

func (basic *BasicStatWeightProcess) Run() {
	for _, statType := range g_requiredStats {
		for _, simType := range g_requiredSims {
			colWeight := basic.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
			basic.detailedWeights.Put(statType, simType, colWeight)
			basic.colNames = append(basic.colNames, "WEIGHT: "+statType.Name()+" "+simType.String())

			colUnit := basic.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
			basic.unitStatValue.Put(statType, simType, colUnit)
			basic.colNames = append(basic.colNames, "UNIT: "+statType.Name()+" "+simType.String())
		}
	}

	for _, simType := range g_requiredSims {
		value := basic.targetRatios.Get(simType)
		strengthSetToRatio := utilhighs.ConstraintRowBuild{}
		strengthSetToRatio.Add(basic.detailedWeights.GetOrPanic(stats.Stat_Strength, simType), 1)
		strengthSetToRatio.Finish(&basic.input, value, value)
	}

	basic.simData.ForeachWithKeys(func(statType stats.StatType, statValue uint32, sim simulate.SimResultStats) {
		basic.incorporateSample(statType, statValue, sim)
	})
	basic.convertUnitValuesToDetailedRatings()

	solution, log := basic.input.RunHighs()
	basic.printer.AppendOther(log)
	basic.printer.Println(solution.Status.String())

	for i, x := range solution.ColValues {
		if i < len(basic.colNames) {
			basic.printer.Printf("%3d %14f %s\n", i, x, basic.colNames[i])
		}
	}
}

func (basic *BasicStatWeightProcess) convertUnitValuesToDetailedRatings() {
	basic.unitStatValue.ForeachGroupForKey2(func(simType simulate.SimResultType, lookupStat func(stats.StatType) utilhighs.ColumnIndex) {
		baseStatType := stats.Stat_Strength
		// baseUnitColumn := lookupStat(baseStatType)
		baseWeightColumn := basic.detailedWeights.GetOrPanic(baseStatType, simType)
		for _, thisStatType := range g_requiredStats {
			if thisStatType != baseStatType {
				thisUnitColumn := lookupStat(thisStatType)
				thisWeightColumn := basic.detailedWeights.GetOrPanic(thisStatType, simType)
				value := basic.targetRatios.Get(simType) // effectively str weight
			}
		}

		// stage2weight_dps_haste = unit_dps_haste / unit_dps_str * stage2weight_str
		// stage2weight_dps_haste / stage2weight_str = unit_dps_haste / unit_dps_str

		// stage2weight_dps_haste / 0.4 - unit_dps_haste / unit_dps_str = 0

		// ax + by = z[const]
		// x + by/a = z/a
		// x = z/a - by/a


		// ratio = unit_dps_haste / unit_dps_str
		// ratio * unit_dps_str = unit_dps_haste 
		// ratio * unit_dps_str - unit_dps_haste = 0


		// diff = stage2weight_dps_haste / stage2weight_str - unit_dps_haste / unit_dps_str
	})

	//
}


func (basic *BasicStatWeightProcess) unitDiffValue(sim simulate.SimResultStats, simType simulate.SimResultType, statValue uint32) float64 {
	simValueDiff := sim.Get(simType) - basic.simBase.Get(simType)
	simValueDiffPerStat := simValueDiff / float64(statValue)
	return simValueDiffPerStat

	// this is a single diff value, ideally we want to push data in and average across multiple
	// therefore ideally a variable not const
}

func (basic *BasicStatWeightProcess) incorporateSample(statType stats.StatType, statValue uint32, sim simulate.SimResultStats) {

	for _, simType := range g_requiredSims {
		unitColumn := basic.unitStatValue.GetOrPanic(statType, simType)
		// TODO validateIsRelevantBase(basic.simBase // can't valiidate since we don't actually have full stat blocks, just this one difference value

		row := utilhighs.ConstraintRowBuild{}
		row.Add(unitColumn, 1)

		offset := basic.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
		basic.colNames = append(basic.colNames, "diff to unit: "+statType.Name()+" "+simType.String())
		row.Add(offset, 1)

		simValueDiffPerStat := basic.unitDiffValue(sim, simType, statValue)
		row.Finish(&basic.input, simValueDiffPerStat, simValueDiffPerStat)

		offsetAbsolute := basic.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1)
		basic.colNames = append(basic.colNames, "abs diff to unit: "+statType.Name()+" "+simType.String()+" {output}")
		utilhighs.AbsoluteValue2(&basic.input, offset, offsetAbsolute, rangeHigh)
	}

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
