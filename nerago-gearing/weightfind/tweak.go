package weightfind

import (
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
)

const (
	c_tweak_start      = 0.01
	c_tweak_limit      = 0.00001
	c_tweak_iter_count = 1000 // limit to avoid infinite loop
)

func WeightTweaker(startWeight stathighs.WeightResult, weightStats []stats.StatType, targetRatio stats.SimData, inputData []stathighs.WeightInput, printer *util.PrintRecorder) (stathighs.WeightResult, float64) {
	updatedWeight, updatedAccuracy := weightTweakerInternal(startWeight, c_tweak_start, weightStats, targetRatio, inputData, printer)
	return updatedWeight, updatedAccuracy
}

func weightTweakerInternal(startWeight stathighs.WeightResult, tweakStart float64, weightStats []stats.StatType, targetRatio stats.SimData, inputData []stathighs.WeightInput, printer *util.PrintRecorder) (stathighs.WeightResult, float64) {
	increment := tweakStart
	factor := 1 + increment

	type weightAndAccuracy struct {
		weight   stathighs.WeightResult
		accuracy float64
	}
	bestEntry := &weightAndAccuracy{
		startWeight.Clone(),
		EvaluateAccuracy(startWeight, inputData, targetRatio),
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
				accuracyMul := EvaluateAccuracy(mul, inputData, targetRatio)
				best.Offer(&weightAndAccuracy{mul, accuracyMul}, accuracyMul)

				div := bestEntry.weight.Clone()
				div.DivideEquals(stat, factor)
				accuracyDiv := EvaluateAccuracy(div, inputData, targetRatio)
				best.Offer(&weightAndAccuracy{div, accuracyDiv}, accuracyDiv)
			}

			add := bestEntry.weight.Clone()
			add.PlusEquals(stat, increment)
			accuracyAdd := EvaluateAccuracy(add, inputData, targetRatio)
			best.Offer(&weightAndAccuracy{add, accuracyAdd}, accuracyAdd)

			sub := bestEntry.weight.Clone()
			sub.MinusEquals(stat, increment)
			accuracySub := EvaluateAccuracy(sub, inputData, targetRatio)
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

func weightTweakerInternal_FastNoRange(startWeight stathighs.WeightResult, tweakStart float64, weightStats []stats.StatType, simTypes []stats.SimType, targetRatio stats.SimData, inputData []stathighs.WeightInput) (stathighs.WeightResult, float64) {
	increment := tweakStart
	factor := 1 + increment

	best := util_rank.BestCollector1[stathighs.WeightResult]{}
	best.Offer(
		&startWeight,
		EvaluateAccuracyNoRangeInlined1(startWeight, simTypes, targetRatio, inputData),
	)

	for range c_tweak_iter_count {
		foundImprovement := false
		for i := 1; i < len(weightStats); i++ {
			stat := weightStats[i]

			if !best.BestObject.IsZero(stat) {
				mul := best.BestObject.Clone()
				mul.MultiplyEquals(stat, factor)
				accuracyMul := EvaluateAccuracyNoRangeInlined1(mul, simTypes, targetRatio, inputData)
				if best.OfferAndIsBetter(&mul, accuracyMul) {
					foundImprovement = true
				}

				div := best.BestObject.Clone()
				div.DivideEquals(stat, factor)
				accuracyDiv := EvaluateAccuracyNoRangeInlined1(div, simTypes, targetRatio, inputData)
				if best.OfferAndIsBetter(&div, accuracyDiv) {
					foundImprovement = true
				}
			}

			add := best.BestObject.Clone()
			add.PlusEquals(stat, increment)
			accuracyAdd := EvaluateAccuracyNoRangeInlined1(add, simTypes, targetRatio, inputData)
			if best.OfferAndIsBetter(&add, accuracyAdd) {
				foundImprovement = true
			}

			sub := best.BestObject.Clone()
			sub.MinusEquals(stat, increment)
			accuracySub := EvaluateAccuracyNoRangeInlined1(sub, simTypes, targetRatio, inputData)
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

func weightTweakerInternal_FastFullRange(startWeight stathighs.WeightResult, tweakStart float64, weightStats []stats.StatType, simTypes []stats.SimType, targetRatio stats.SimData, inputData []stathighs.WeightInput) (stathighs.WeightResult, float64) {
	increment := tweakStart
	factor := 1 + increment

	best := util_rank.BestCollector1[stathighs.WeightResult]{}
	best.Offer(
		&startWeight,
		EvaluateAccuracyFullRangeInlined(startWeight, simTypes, targetRatio, inputData),
	)

	for range c_tweak_iter_count {
		foundImprovement := false
		for i := 1; i < len(weightStats); i++ {
			stat := weightStats[i]

			if !best.BestObject.IsZero(stat) {
				mul := best.BestObject.Clone()
				mul.MultiplyEquals(stat, factor)
				accuracyMul := EvaluateAccuracyFullRangeInlined(mul, simTypes, targetRatio, inputData)
				if best.OfferAndIsBetter(&mul, accuracyMul) {
					foundImprovement = true
				}

				div := best.BestObject.Clone()
				div.DivideEquals(stat, factor)
				accuracyDiv := EvaluateAccuracyFullRangeInlined(div, simTypes, targetRatio, inputData)
				if best.OfferAndIsBetter(&div, accuracyDiv) {
					foundImprovement = true
				}
			}

			add := best.BestObject.Clone()
			add.PlusEquals(stat, increment)
			accuracyAdd := EvaluateAccuracyFullRangeInlined(add, simTypes, targetRatio, inputData)
			if best.OfferAndIsBetter(&add, accuracyAdd) {
				foundImprovement = true
			}

			sub := best.BestObject.Clone()
			sub.MinusEquals(stat, increment)
			accuracySub := EvaluateAccuracyFullRangeInlined(sub, simTypes, targetRatio, inputData)
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
