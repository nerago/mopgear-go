package weightfind

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
)

var tweakerChangeStats = []stats.StatType{
	stats.Stat_Stamina, stats.Stat_Crit, stats.Stat_Haste,
	stats.Stat_Expertise, stats.Stat_Mastery, stats.Stat_Dodge, stats.Stat_Parry}

func WeightTweaker(startWeight stathighs.WeightResult, changeStats []stats.StatType, targetRatio simulate.SimData, inputData []stathighs.WeightInput, printer *util.PrintRecorder) stathighs.WeightResult {
	mult := 1.01
	bestWeight := startWeight.Clone()

	for {
		best := util_rank.BestCollector1[stathighs.WeightResult]{}
		best.Offer(&bestWeight, EvaluateAccuracy(bestWeight, inputData, targetRatio))
		for _, stat := range changeStats {
			hi := bestWeight.Clone()
			hi[stat] *= mult
			best.Offer(&hi, EvaluateAccuracy(hi, inputData, targetRatio))

			lo := bestWeight.Clone()
			lo[stat] /= mult
			best.Offer(&lo, EvaluateAccuracy(lo, inputData, targetRatio))
		}
		updateWeight := best.GetBestOrPanic()

		if updateWeight.Equals(bestWeight) {
			printer.Printf("DONE\n")
			break
		} else {
			printer.Printf("NEXT %s accuracy=%f\n", updateWeight.String(), EvaluateAccuracy(updateWeight, inputData, targetRatio))
			bestWeight = updateWeight
		}
	}

	return bestWeight
}
