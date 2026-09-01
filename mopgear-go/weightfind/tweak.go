package weightfind

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_rank"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

const (
	c_tweak_start      = 0.01
	c_tweak_limit      = 0.00001
	c_tweak_iter_count = 1000 // limit to avoid infinite loop
)

func WeightTweakerWithLogging(startWeight weight_types.Weight1Basic, weightStats []stats.StatType, targetRatio *weight_types.SimPriorityBasic, inputData []weight_types.WeightInput, printer *util.PrintRecorder) (weight_types.Weight1Basic, float64) {
	updatedWeight, updatedAccuracy := weightTweakerInternalLogged(startWeight, c_tweak_start, weightStats, targetRatio, inputData, printer)
	updatedWeight.NormalizeForBase(weightStats)
	return updatedWeight, updatedAccuracy
}

func weightTweakerInternalLogged(startWeight weight_types.Weight1Basic, tweakStart float64, weightStats []stats.StatType, targetRatio *weight_types.SimPriorityBasic, inputData []weight_types.WeightInput, printer *util.PrintRecorder) (weight_types.Weight1Basic, float64) {
	requiredSims := targetRatio.SimTypes()
	increment := tweakStart
	factor := 1 + increment

	if startWeight.IsEmpty() {
		return startWeight, EvaluateAccuracyBasic(&startWeight, requiredSims, targetRatio, inputData)
	}

	type weightAndAccuracy struct {
		weight   weight_types.Weight1Basic
		accuracy float64
	}
	bestEntry := &weightAndAccuracy{
		startWeight.Clone(),
		EvaluateAccuracyBasic(&startWeight, requiredSims, targetRatio, inputData),
	}
	printer.Printf("START %s accuracy=%f\n", bestEntry.weight.String(), bestEntry.accuracy)

	for range c_tweak_iter_count {
		best := util_rank.BestCollector1[weightAndAccuracy]{}
		best.Offer(bestEntry, bestEntry.accuracy)

		for i := 1; i < len(weightStats); i++ {
			stat := weightStats[i]

			if !bestEntry.weight.IsZero(stat) {
				mul := bestEntry.weight.Clone()
				mul.MultiplyEquals(stat, factor)
				accuracyMul := EvaluateAccuracyBasic(&mul, requiredSims, targetRatio, inputData)
				best.Offer(&weightAndAccuracy{mul, accuracyMul}, accuracyMul)

				div := bestEntry.weight.Clone()
				div.DivideEquals(stat, factor)
				accuracyDiv := EvaluateAccuracyBasic(&div, requiredSims, targetRatio, inputData)
				best.Offer(&weightAndAccuracy{div, accuracyDiv}, accuracyDiv)
			}

			add := bestEntry.weight.Clone()
			add.PlusEquals(stat, increment)
			accuracyAdd := EvaluateAccuracyBasic(&add, requiredSims, targetRatio, inputData)
			best.Offer(&weightAndAccuracy{add, accuracyAdd}, accuracyAdd)

			sub := bestEntry.weight.Clone()
			sub.MinusEquals(stat, increment)
			accuracySub := EvaluateAccuracyBasic(&sub, requiredSims, targetRatio, inputData)
			best.Offer(&weightAndAccuracy{sub, accuracySub}, accuracySub)
		}

		proposedEntry := best.GetBestPointerOrPanic()
		if bestEntry.weight.Equals(&proposedEntry.weight) {
			increment /= 2
			factor = 1 + increment
			if increment <= c_tweak_limit {
				printer.Printf("TWEAK DONE\n")
				break
			}
		} else {
			printer.Printf("TWEAK %s accuracy=%f\n", proposedEntry.weight.String(), proposedEntry.accuracy)
			bestEntry = proposedEntry
		}
	}

	return bestEntry.weight, bestEntry.accuracy
}

func weightTweaker_internal_FastCached(startWeight weight_types.Weight1Basic, tweakStart float64, weightStats []stats.StatType, evaluate *EvaluateAccuracyPrepared) (weight_types.Weight1Basic, float64) {
	if startWeight.IsEmpty() {
		return startWeight, evaluate.EvaluateWeight1(&startWeight)
	}

	increment := tweakStart
	factor := 1 + increment

	best := util_rank.BestCollector1[weight_types.Weight1Basic]{}
	best.Offer(
		&startWeight,
		evaluate.EvaluateWeight1(&startWeight),
	)

	for range c_tweak_iter_count {
		foundImprovement := false
		for i := 1; i < len(weightStats); i++ {
			stat := weightStats[i]

			if !best.GetBestPointerOrPanic().IsZero(stat) {
				mul := best.GetBestPointerOrPanic().Clone()
				mul.MultiplyEquals(stat, factor)
				accuracyMul := evaluate.EvaluateWeight1(&mul)
				if best.OfferAndIsBetter(&mul, accuracyMul) {
					foundImprovement = true
				}

				div := best.GetBestPointerOrPanic().Clone()
				div.DivideEquals(stat, factor)
				accuracyDiv := evaluate.EvaluateWeight1(&div)
				if best.OfferAndIsBetter(&div, accuracyDiv) {
					foundImprovement = true
				}
			}

			add := best.GetBestPointerOrPanic().Clone()
			add.PlusEquals(stat, increment)
			accuracyAdd := evaluate.EvaluateWeight1(&add)
			if best.OfferAndIsBetter(&add, accuracyAdd) {
				foundImprovement = true
			}

			sub := best.GetBestPointerOrPanic().Clone()
			sub.MinusEquals(stat, increment)
			accuracySub := evaluate.EvaluateWeight1(&sub)
			if best.OfferAndIsBetter(&sub, accuracySub) {
				foundImprovement = true
			}
		}

		if !foundImprovement {
			increment /= 2
			factor = 1 + increment
			if increment <= c_tweak_limit {
				break
			}
		}
	}

	return best.GetBestOrPanic(), best.GetBestScore()
}
