package stathighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

const c_highRange = 100000.0

func CalcComplexStatWeights(inputData []WeightInput, targetRatios simulate.SimResultStats, printer *util.PrintRecorder) map[stats.StatType]float64 {
	// highRange := 100000.0 // basic sum of stats Mop P4 is about 83k
	// scaleMin := 0.001  // similar number of places, can get scores down to around 1
	colNames := make([]string, 0)

	input := new(utilhighs.InputBuilder)

	input.Minimise = false
	input.BlendMultiObjectives = true
	input.Solver = "ipm"
	input.AddLinearObjective(-1, 0, 1000, 0.2, 1) // contributionDiff
	input.AddLinearObjective(-2, 0, 1000, 0.2, 2) // scoreDiff
	input.AddLinearObjective(-1, 0, 1000, 0.2, 3) // data pairs offset

	statWeightColumns, colNames, simWeightColumns := weightColumns(input, colNames)

	// colNames = dataEquations(inputData, input, colNames, statWeightColumns, simWeightColumns, targetRatios)

	colNames = dataCompareInPairsEquations(input, colNames, inputData, statWeightColumns, simWeightColumns, targetRatios)

	solution, log := input.RunHighs()
	printer.AppendOther(log)
	printer.Println(solution.Status.String())

	return reportWeightSolution(solution, printer, colNames, statWeightColumns, simWeightColumns, inputData)
}

func weightColumns(input *utilhighs.InputBuilder, colNames []string) (map[stats.StatType]utilhighs.ColumnIndex, []string, map[simulate.SimResultType]utilhighs.ColumnIndex) {
	scaleMin := 0.0
	scaleMax := 100.0

	totalStatWeightRow := utilhighs.ConstraintRowBuild{}
	statWeightColumns := make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, stat := range G_RequiredStats {
		statColIndex := input.CreateColumnGeneral(highs.Continuous, scaleMin, scaleMax)
		// statColIndex := input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
		colNames = append(colNames, stat.Name())
		statWeightColumns[stat] = statColIndex
		totalStatWeightRow.Add(statColIndex, 1)
	}
	totalStatWeightRow.Finish(input, 1, 1)

	// force strength=1
	forceStrength := utilhighs.ConstraintRowBuild{}
	forceStrength.Add(statWeightColumns[stats.Stat_Strength], 1)
	forceStrength.Finish(input, 1, 1)

	totalSimWeightRow := utilhighs.ConstraintRowBuild{}
	simWeightColumns := make(map[simulate.SimResultType]utilhighs.ColumnIndex)
	for _, simType := range G_RequiredSims {
		simColIndex := input.CreateColumnGeneral(highs.Continuous, scaleMin, scaleMax)
		// simColIndex := input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
		colNames = append(colNames, simType.String())
		simWeightColumns[simType] = simColIndex
		totalSimWeightRow.Add(simColIndex, 1)
	}
	// totalSimWeightRow.finish(input, 1, 1)
	return statWeightColumns, colNames, simWeightColumns
}

func dataEquations(inputData []WeightInput, input *utilhighs.InputBuilder, colNames []string, statWeightColumns map[stats.StatType]utilhighs.ColumnIndex, simWeightColumns map[simulate.SimResultType]utilhighs.ColumnIndex, targetRatios simulate.SimResultStats) []string {
	for _, data := range inputData {
		// add up weighted gear score for row
		gearScoreTotal := input.CreateColumnGeneral(highs.Continuous, 0, c_highRange)
		colNames = append(colNames, "gearScoreTotal")
		gearRow := utilhighs.ConstraintRowBuild{}
		for _, statType := range G_RequiredStats {
			statWeightCol := statWeightColumns[statType]
			gearRow.Add(statWeightCol, float64(data.TotalStat.Get(statType)))
		}
		gearRow.Add(gearScoreTotal, -1)
		gearRow.Finish(input, 0, 0)

		// add up weighted sim score for row
		simScoreTotal := input.CreateColumnGeneral(highs.Continuous, 0, c_highRange)
		colNames = append(colNames, "simScoreTotal")
		simScoreRow := utilhighs.ConstraintRowBuild{}
		for _, simType := range G_RequiredSims {
			simWeightCol := simWeightColumns[simType]
			simScoreRow.Add(simWeightCol, data.SimResult.GetFriendly(simType))
		}
		simScoreRow.Add(simScoreTotal, -1)
		simScoreRow.Finish(input, 0, 0)

		// contribution of sim part of simScoreTotal
		for _, simType := range G_RequiredSims {
			contributionRow := utilhighs.ConstraintRowBuild{}

			// formula sum(simWeight*simResult[simType]) = simScoreTotal
			// simScoreTotal * 0.4 = simWeightCol * valuue[type]

			// calc simResult[simType]*simWeightCol[simType]
			simWeightCol := simWeightColumns[simType]
			contributionRow.Add(simWeightCol, data.SimResult.GetFriendly(simType))

			// ideally we want that to equal simScoreTotal*targetRatio[simType]
			contributionRow.Add(simScoreTotal, -targetRatios.Get(simType))

			// diff = simWeightCol[type] * thisResult[type] - thisScoreTotal * targetRatio[type]
			// ideally for comparison sake the diff should look like the ratio
			// simWeightCol[type] * thisResult[type] - thisScoreTotal * (targetRatio[type] + diff) = 0
			// simWeightCol[type] * thisResult[type] - thisScoreTotal * (targetRatio[type] + diff) = 0

			contributionDiff := input.CreateColumnGeneral(highs.Continuous, -c_highRange, c_highRange)
			// contributionDiff := input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
			colNames = append(colNames, "contributionDiff")
			contributionRow.Add(contributionDiff, 1)
			contributionRow.Finish(input, 0, 0)

			contributionDiffAbsOutput := input.CreateColumnForLinearObjective(highs.Continuous, 0, c_highRange, 1, 0)
			// contributionDiffAbsOutput := input.CreateColumnForLinearObjective(highs.Continuous, 0, utilhighs.C_PlusInf, 1, 0)
			colNames = append(colNames, "contributionDiffAbsOutput")
			utilhighs.AbsoluteValue2(input, contributionDiff, contributionDiffAbsOutput) // could maybe be narrower highRange
			// absoluteValue(input, contributionDiff, contributionDiffAbsOutput, highRange)
			// colNames = append(colNames, "absBool")
		}

		scoreDiffSigned := input.CreateColumnGeneral(highs.Continuous, -c_highRange, c_highRange)
		// scoreDiffSigned := input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
		colNames = append(colNames, "scoreDiffSigned")
		scoreDiffRow := utilhighs.ConstraintRowBuild{}
		scoreDiffRow.Add(gearScoreTotal, 1)
		scoreDiffRow.Add(simScoreTotal, -1)
		scoreDiffRow.Add(scoreDiffSigned, 1)
		scoreDiffRow.Finish(input, 0, 0)

		// diffAbsOutput := input.CreateColumnForLinearObjective(highs.Continuous, 0, utilhighs.C_PlusInf, 1, 1)
		diffAbsOutput := input.CreateColumnForLinearObjective(highs.Continuous, 0, c_highRange, 1, 1)
		colNames = append(colNames, "diffAbsOutput")
		utilhighs.AbsoluteValue2(input, scoreDiffSigned, diffAbsOutput)
		// absoluteValue(input, scoreDiffSigned, diffAbsOutput, highRange)
		// colNames = append(colNames, "absBool")
	}
	return colNames
}

func dataCompareInPairsEquations(input *utilhighs.InputBuilder, colNames []string, inputData []WeightInput, statWeightColumns map[stats.StatType]utilhighs.ColumnIndex, simWeightColumns map[simulate.SimResultType]utilhighs.ColumnIndex, targetRatios simulate.SimResultStats) []string {
	detailedWeights := make(map[stats.StatType]map[simulate.SimResultType]utilhighs.ColumnIndex)
	for _, statType := range G_RequiredStats {
		detailedWeights[statType] = make(map[simulate.SimResultType]utilhighs.ColumnIndex)
		for _, simType := range G_RequiredSims {
			detailedWeights[statType][simType] = input.CreateColumnGeneral(highs.Continuous, -c_highRange, c_highRange)
			// detailedWeights[statType][simType] = input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
		}
	}

	inputData = inputData[0:20]
	for _, data1 := range inputData {
		for _, data2 := range inputData {
			if !stats.StatBlock_Equals(&data1.TotalStat, &data2.TotalStat) {
				colNames = dataCompareInPairsEquations_Single(input, colNames, data1, data2, detailedWeights, c_highRange)
			}
		}
	}

	for _, statType := range G_RequiredStats {
		statRow := utilhighs.ConstraintRowBuild{}
		for _, simType := range G_RequiredSims {
			statRow.Add(detailedWeights[statType][simType], targetRatios.Get(simType))
		}
		statRow.Add(statWeightColumns[statType], -1)
		statRow.Finish(input, 0, 0)
	}

	return colNames
}

func dataCompareInPairsEquations_Single(input *utilhighs.InputBuilder, colNames []string, data1, data2 WeightInput, weights map[stats.StatType]map[simulate.SimResultType]utilhighs.ColumnIndex, highRange float64) []string {

	// so i'd normally do this in the area of a single stat and simType
	// Death's Haste improvement is F37/C37 = (death_with_0_haste_str - death_with_+600_haste)/(death_with_0_haste_str - death_with_+600_str)
	// overall haste improvement is 0.4 * pos_only(F37/C37)

	// we consider gearScoreTotal1/gearScoreTotal2? doesn't say much

	// MAYBE A NEW SET OF COEFFECTIENCTS FOR HASTE_DEATH etc
	// then we can define signedDiffDeath = (haste2-haste1)*hasteDeath + (crit2-crit1)*critDeath			{or this series expands out cleanly}

	// maybe we make an equasion for each

	// sim's percent difference should make sense for given stat combo
	// simTotal * 0.4

	for _, simType := range G_RequiredSims {
		simRow := utilhighs.ConstraintRowBuild{}

		for _, statType := range G_RequiredStats {
			statDiff := data1.TotalStat.Get(statType) - data2.TotalStat.Get(statType)
			simRow.Add(weights[statType][simType], float64(statDiff))
		}

		// offset := input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
		offset := input.CreateColumnGeneral(highs.Continuous, -highRange, highRange)
		simRow.Add(offset, 1)
		offsetAbs := input.CreateColumnForLinearObjective(highs.Continuous, 0, utilhighs.C_PlusInf, 1, 2)
		utilhighs.AbsoluteValue2(input, offset, offsetAbs)

		var simDiff float64
		if simType.IsHighGood() {
			simDiff = data1.SimResult.Get(simType) - data2.SimResult.Get(simType)
		} else {
			simDiff = data2.SimResult.Get(simType) - data1.SimResult.Get(simType)
		}
		simRow.Finish(input, simDiff, simDiff)
	}

	return colNames
}

func reportWeightSolution(solution *highs.Solution, printer *util.PrintRecorder, colNames []string, statWeightColumns map[stats.StatType]utilhighs.ColumnIndex, simWeightColumns map[simulate.SimResultType]utilhighs.ColumnIndex, inputData []WeightInput) map[stats.StatType]float64 {
	for i, x := range solution.ColValues {
		if i < len(colNames) {
			printer.Printf("%3d %14f %s\n", i, x, colNames[i])
		}
	}

	statWeightResult := make(map[stats.StatType]float64)
	simWeightResult := make(map[simulate.SimResultType]float64)
	printer.Println("WEIGHTS")
	for _, statType := range G_RequiredStats {
		statWeightCol := statWeightColumns[statType]
		statWeightResult[statType] = solution.ColValues[statWeightCol]
		printer.Printf("%10s %f\n", statType.Name(), statWeightResult[statType])
	}
	for _, simType := range G_RequiredSims {
		simWeightCol := simWeightColumns[simType]
		simWeightResult[simType] = solution.ColValues[simWeightCol]
		printer.Printf("%10s %f\n", simType.String(), simWeightResult[simType])
	}

	for i := range 5 {
		data := inputData[i]
		printer.Println("EXAMPLE")

		// formula sum(simWeight*simResult[simType]) = simScoreTotal
		// simScoreTotal * 0.4 = simWeightCol * valuue[type]

		// calc simResult[simType]*simWeightCol[simType]

		// ideally we want that to equal simScoreTotal*targetRatio[simType]

		statSum := 0.0
		for _, statType := range G_RequiredStats {
			val := float64(data.TotalStat.Get(statType))
			weight := statWeightResult[statType]
			printer.Printf(" %10s %8.2f * %8.4f = %8.2f\n", statType.Name(), val, weight, val*weight)
			statSum += val * weight
		}
		printer.Printf("%46f\n", statSum)

		simSum := 0.0
		for _, simType := range G_RequiredSims {
			val := data.SimResult.GetFriendly(simType)
			weight := simWeightResult[simType]
			simSum += val * weight
		}
		for _, simType := range G_RequiredSims {
			val := data.SimResult.GetFriendly(simType)
			weight := simWeightResult[simType]
			printer.Printf(" %10s %12.2f * %8.4f = %8.2f (%.4f)\n", simType.String(), val, weight, val*weight, val*weight/simSum)
		}
		printer.Printf("%51f\n", simSum)

		printer.Println0()
	}

	return statWeightResult
}
