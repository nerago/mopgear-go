package weightfind

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
	"paladin_gearing_go/weightfind/weight_highs"
)

const (
	c_tweak_start      = 0.01
	c_tweak_limit      = 0.00001
	c_tweak_iter_count = 1000 // limit to avoid infinite loop
)

func WeightTweakerWithLogging(startWeight weight_highs.WeightResult, weightStats []stats.StatType, targetRatio stats.SimData, inputData []weight_highs.WeightInput, printer *util.PrintRecorder) (weight_highs.WeightResult, float64) {
	updatedWeight, updatedAccuracy := weightTweakerInternalLogged(startWeight, c_tweak_start, weightStats, targetRatio, inputData, printer)
	return updatedWeight, updatedAccuracy
}

func weightTweakerInternalLogged(startWeight weight_highs.WeightResult, tweakStart float64, weightStats []stats.StatType, targetRatio stats.SimData, inputData []weight_highs.WeightInput, printer *util.PrintRecorder) (weight_highs.WeightResult, float64) {
	requiredSims := targetRatio.NonZeroTypes()
	increment := tweakStart
	factor := 1 + increment

	type weightAndAccuracy struct {
		weight   weight_highs.WeightResult
		accuracy float64
	}
	bestEntry := &weightAndAccuracy{
		startWeight.Clone(),
		EvaluateAccuracyRanged(startWeight, requiredSims, targetRatio, inputData),
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
				accuracyMul := EvaluateAccuracyRanged(startWeight, requiredSims, targetRatio, inputData)
				best.Offer(&weightAndAccuracy{mul, accuracyMul}, accuracyMul)

				div := bestEntry.weight.Clone()
				div.DivideEquals(stat, factor)
				accuracyDiv := EvaluateAccuracyRanged(startWeight, requiredSims, targetRatio, inputData)
				best.Offer(&weightAndAccuracy{div, accuracyDiv}, accuracyDiv)
			}

			add := bestEntry.weight.Clone()
			add.PlusEquals(stat, increment)
			accuracyAdd := EvaluateAccuracyRanged(startWeight, requiredSims, targetRatio, inputData)
			best.Offer(&weightAndAccuracy{add, accuracyAdd}, accuracyAdd)

			sub := bestEntry.weight.Clone()
			sub.MinusEquals(stat, increment)
			accuracySub := EvaluateAccuracyRanged(startWeight, requiredSims, targetRatio, inputData)
			best.Offer(&weightAndAccuracy{sub, accuracySub}, accuracySub)
		}

		proposedEntry := best.GetBestPointerOrPanic()
		if bestEntry.weight.Equals(&proposedEntry.weight) {
			increment /= 2
			factor = 1 + increment
			if increment <= c_tweak_limit {
				printer.Printf("DONE\n")
				break
			}
		} else {
			printer.Printf("NEXT %s accuracy=%f\n", proposedEntry.weight.String(), proposedEntry.accuracy)
			bestEntry = proposedEntry
		}
	}

	return bestEntry.weight, bestEntry.accuracy
}

// TODO compare logic
func weightTweakerInternal_Fast(startWeight weight_highs.WeightResult, tweakStart float64, weightStats []stats.StatType, simTypes []stats.SimType, targetRatio stats.SimData, inputData []weight_highs.WeightInput) (weight_highs.WeightResult, float64) {
	increment := tweakStart
	factor := 1 + increment

	best := util_rank.BestCollector1[weight_highs.WeightResult]{}
	best.Offer(
		&startWeight,
		EvaluateAccuracyRanged(startWeight, simTypes, targetRatio, inputData),
	)

	for range c_tweak_iter_count {
		foundImprovement := false
		for i := 1; i < len(weightStats); i++ {
			stat := weightStats[i]

			if !best.BestObject.IsZero(stat) {
				mul := best.BestObject.Clone()
				mul.MultiplyEquals(stat, factor)
				accuracyMul := EvaluateAccuracyRanged(startWeight, simTypes, targetRatio, inputData)
				if best.OfferAndIsBetter(&mul, accuracyMul) {
					foundImprovement = true
				}

				div := best.BestObject.Clone()
				div.DivideEquals(stat, factor)
				accuracyDiv := EvaluateAccuracyRanged(startWeight, simTypes, targetRatio, inputData)
				if best.OfferAndIsBetter(&div, accuracyDiv) {
					foundImprovement = true
				}
			}

			add := best.BestObject.Clone()
			add.PlusEquals(stat, increment)
			accuracyAdd := EvaluateAccuracyRanged(startWeight, simTypes, targetRatio, inputData)
			if best.OfferAndIsBetter(&add, accuracyAdd) {
				foundImprovement = true
			}

			sub := best.BestObject.Clone()
			sub.MinusEquals(stat, increment)
			accuracySub := EvaluateAccuracyRanged(startWeight, simTypes, targetRatio, inputData)
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

	return best.GetBestOrPanic(), best.BestValue
}

func WeightTweaker_FastCached(startWeight weight_highs.WeightResult, weightStats []stats.StatType, evaluate *EvaluateAccuracyPrepared) (weight_highs.WeightResult, float64) {
	updatedWeight, updatedAccuracy := weightTweaker_internal_FastCached(startWeight, c_tweak_start, weightStats, evaluate)
	return updatedWeight, updatedAccuracy
}

func weightTweaker_internal_FastCached(startWeight weight_highs.WeightResult, tweakStart float64, weightStats []stats.StatType, evaluate *EvaluateAccuracyPrepared) (weight_highs.WeightResult, float64) {
	increment := tweakStart
	factor := 1 + increment

	best := util_rank.BestCollector1[weight_highs.WeightResult]{}
	best.Offer(
		&startWeight,
		evaluate.EvaluateWeight(startWeight),
	)

	for range c_tweak_iter_count {
		foundImprovement := false
		for i := 1; i < len(weightStats); i++ {
			stat := weightStats[i]

			if !best.BestObject.IsZero(stat) {
				mul := best.BestObject.Clone()
				mul.MultiplyEquals(stat, factor)
				accuracyMul := evaluate.EvaluateWeight(startWeight)
				if best.OfferAndIsBetter(&mul, accuracyMul) {
					foundImprovement = true
				}

				div := best.BestObject.Clone()
				div.DivideEquals(stat, factor)
				accuracyDiv := evaluate.EvaluateWeight(startWeight)
				if best.OfferAndIsBetter(&div, accuracyDiv) {
					foundImprovement = true
				}
			}

			add := best.BestObject.Clone()
			add.PlusEquals(stat, increment)
			accuracyAdd := evaluate.EvaluateWeight(startWeight)
			if best.OfferAndIsBetter(&add, accuracyAdd) {
				foundImprovement = true
			}

			sub := best.BestObject.Clone()
			sub.MinusEquals(stat, increment)
			accuracySub := evaluate.EvaluateWeight(startWeight)
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

	return best.GetBestOrPanic(), best.BestValue
}
