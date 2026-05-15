package withhighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"

	"github.com/bartolsthoorn/gohighs/highs"
)

type NewWeightInput struct {
	totalStat stats.StatBlock
	simResult simulate.SimResultStats
}

var defSpreadSheetWeight = simulate.SimResultStats{
	DPS:   0.1,
	DEATH: 0.2,
	TMI:   0.4,
	DTPS:  0.3,
}

var defWeight = simulate.SimResultStats{
	DPS:   0.1,
	DEATH: 0.2,
	TMI:   0.3,
	DTPS:  0.4,
}

var animusWeight = simulate.SimResultStats{
	DPS:   0.4,
	DEATH: 0.1,
	TMI:   0.4,
	DTPS:  0.1,
}

var dpsWeight = simulate.SimResultStats{
	DPS:   0.97,
	DEATH: 0.01,
	TMI:   0.01,
	DTPS:  0.01,
}

var requiredStats = []stats.StatType{stats.Stat_Strength, stats.Stat_Stamina, stats.Stat_Crit, stats.Stat_Haste, stats.Stat_Expertise, stats.Stat_Mastery, stats.Stat_Dodge, stats.Stat_Parry}
var requiredSims = []simulate.SimResultType{simulate.Result_DPS, simulate.Result_DEATH, simulate.Result_TMI, simulate.Result_DTPS}

func CalcNewStatWeights(inputData []NewWeightInput, targetRatios simulate.SimResultStats) {
	highRange := 100000.0 // basic sum of stats Mop P4 is about 83k

	input := new(inputBuilder)

	totalStatWeightRow := constraintRowBuild{}
	statWeightColumns := make(map[stats.StatType]int)
	for _, stat := range requiredStats {
		statColIndex := input.createColumnGeneral(highs.Continuous, 0.0001, 1)
		statWeightColumns[stat] = statColIndex
		totalStatWeightRow.add(statColIndex, 1)
	}
	totalStatWeightRow.finish(input, 1, 1)

	totalSimWeightRow := constraintRowBuild{}
	simWeightColumns := make(map[simulate.SimResultType]int)
	for _, simType := range requiredSims {
		simColIndex := input.createColumnGeneral(highs.Continuous, 0.000001, 1)
		simWeightColumns[simType] = simColIndex
		totalSimWeightRow.add(simColIndex, 1)
	}
	totalSimWeightRow.finish(input, 1, 1)

	for _, data := range inputData {
		// add up weighted gear score for row
		gearScoreCol := input.createColumnGeneral(highs.Continuous, 1, 1000)
		gearRow := constraintRowBuild{}
		for statType, statCol := range statWeightColumns {
			gearRow.add(statCol, float64(data.totalStat.Get(statType)))
		}
		gearRow.add(gearScoreCol, -1)
		gearRow.finish(input, 0, 0)

		simScoreRow := constraintRowBuild{}
		for simType, simCol := range simWeightColumns {
			simScoreRow.add(simCol, data.simResult.Get(simType))
		}

		// actually want average contribution to the total
		// make the target weights coeffecients, multiply by calculated weights, then do another diff to target?

		// simTotalScore is made up of 0.1 * x + 0.4 * y + 0.5 * z, should add to one
		// x,y,z = is simResult[Type]*simWeightCol[type]
		// contribution is actually simResult[type]*simWeightCol[type]*targetWeight[type]. of those targetWeight and simResult are known

		// alternate method again calc simResult[Type]*simWeightCol, then diff that to 0.1

		// alternate method again calc simResult[Type]*simWeightCol
		//                        calc sum(simResult[type]

		diffSigned := input.createColumnGeneral(highs.Continuous, c_minusInf, c_plusInf)
		dataRow.add(diffSigned, 1)
		dataRow.finish(input, 0, 0)

		diffAbsOutput := input.createColumnWithOutput(highs.Continuous, c_minusInf, c_plusInf, 1)
		absoluteValue(input, diffSigned, diffAbsOutput, highRange)
	}

	// str   * str_weight??   = str_score
	// haste * haste_weight?? = haste_score
	// crit  * crit_weight??  = crit_score
	//                        = total_gear_score
	//
	// dps   * ?? = [0 .. 0.4]
	// death * ?? = [0 .. 0.1]
	// tmi   * ?? = [0 .. 0.4]
	//            = relative_sim_score

	// we could put both in the 0..1 range?

	// what i kinda want is (abs(a_str - b_str) * w_hst + abs(a_hst - b_hst) * w_hst)
	//                  and (abs(a_dps - b_dps) * z_dps + abs(a_tmi - b_tmi) * z_tmi)
	//    (abs(a_str - b_str) * w_hst + abs(a_hst - b_hst) * w_hst) - (abs(a_dps - b_dps) * z_dps + abs(a_tmi - b_tmi) * z_tmi) = discrepancy
	//    (a_str * w_hst - b_str * w_hst + a_hst * w_hst - b_hst * w_hst) - (a_dps * z_dps - b_dps * z_dps + a_tmi * z_tmi - b_tmi * z_tmi) = discrepancy
	//    (a_str * w_str + a_hst * w_hst) - (a_dps * z_dps + a_tmi * z_tmi) = discrepancy

	// but this just makes up some numbers. might be better to do it per sim stat first then combine
	// how do we otherwise include the target desired ratios
	// maybe doing the rows as differences is better?
}
