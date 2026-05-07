package withhighs

import (
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

func FindSuggestedRatingMultipliers(oldMultipliers []float64, equippedRatings []float64, bestUnderOldMultsRatings []float64, optimumIndependentRatings []float64, printer *util.PrintRecorder) {
	inputBuilder := inputBuilder{}

	// so we want the actual output to be the multi input like RequestRatingPercent: 0.20
	// what we want to minimise is the change in each value?

	// do we need to use the actual ratings at all?
	// pretend they're all 100

	// so calculation previously done sum(request * scaling * equip) = 100000000
	// well thats done in parts ratingMultiply = 100000000 * requestPercent / baseLineRating
	// then later on ResultRating * ratingMultiply
	// which overall is sum(ResultRating * ratingMultiply)
	//                = sum(ResultRating * 100000000 * requestPercent / baseLineRating)
	//                  sum(100000000 * requestPercent * ratingPenalty) <= 100000000

	// if we want to solve for requestPercent then
	//                sum(requestPercent * ratingPenalty) <= 1
	//                sum(requestPercent * ratingPenalty) <= 1

	percentSum := constraintRowBuild{}
	solvePercent := constraintRowBuild{}

	outputVars := []int{}
	for i := range equippedRatings {
		outputVar := inputBuilder.createColumnGeneral(highs.Continuous, 0.001, 0.9)
		outputVars = append(outputVars, outputVar)

		// penalty := equippedRatings[i] / optimumIndependentRatings[i] // not sure which to use
		penalty := equippedRatings[i] / bestUnderOldMultsRatings[i] // not sure which to use
		solvePercent.add(outputVar, penalty)

		percentSum.add(outputVar, 1)

		oldMultVar := inputBuilder.createColumnGeneral(highs.Continuous, oldMultipliers[i], oldMultipliers[i])
		diff := inputBuilder.createColumnWithOutput(highs.Continuous, 0, 0.4, 1)
		diffRow := constraintRowBuild{}
		diffRow.add(diff, -1)
		diffRow.add(oldMultVar, 1)
		diffRow.add(outputVar, -1)
		diffRow.finish(&inputBuilder, 0, 0)

		// https://math.stackexchange.com/questions/432003/converting-absolute-value-program-into-linear-program
		// simple but "unreliable" version
		diffAbsCalcA := inputBuilder.createColumnGeneral(highs.Continuous, 0, 1)
		diffAbsCalcB := inputBuilder.createColumnGeneral(highs.Continuous, -1, 0)
		diffAbsCalcRow := constraintRowBuild{}
		diffAbsCalcRow.add(diffAbsCalcA, 1)
		diffAbsCalcRow.add(diffAbsCalcB, -1)
		diffAbsCalcRow.add(diff, -1)
		diffAbsCalcRow.finish(&inputBuilder, 0, 0)

		// web suggestion was xi - yi <= erri && xi - yi >= -erri
		//                    xi - yi - erri <= 0 && xi - yi + erri >= 0

		// if x-y is +3, 3 <= err && 3 >= -err

		// lets think of it another way abs(x-y) == diff
		//                              x-y == diff or y-x == diff
		//                              (x-y <= diff && x-y>=diff) or (y-x <= diff && y-x >= diff)
		//                              !((x-y > diff || x-y < diff) && (y-x > diff || y-x >= diff))
		//                              !((x-y > diff || x-y < diff) && (y-x > diff || y-x >= diff))

		// screw that diff has minimisation pressure so we just need to specify the lower bound
		//                              diff >= abs(x-y)
		//                              diff >= x-y or diff >= y-x

	}

	percentSum.finish(&inputBuilder, 1, 1)
	solvePercent.finish(&inputBuilder, 0.8, 1)

	solution, log := inputBuilder.runHighsMinimise()
	printer.AppendOther(log)

	if c_debugHighs {
		for i, v := range solution.ColValues {
			printer.Printf("%d %f\n", i, v)
		}
	}
}
