package withhighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

type NewWeightInput struct {
	TotalStat stats.StatBlock
	SimResult simulate.SimResultStats
}

var NewStatWeights_defSpreadSheetWeight = simulate.SimResultStats{
	DPS:   0.1,
	DEATH: 0.2,
	TMI:   0.4,
	DTPS:  0.3,
}

var NewStatWeights_defWeight = simulate.SimResultStats{
	DPS:   0.1,
	DEATH: 0.2,
	TMI:   0.3,
	DTPS:  0.4,
}

var NewStatWeights_animusWeight = simulate.SimResultStats{
	DPS:   0.4,
	DEATH: 0.1,
	TMI:   0.4,
	DTPS:  0.1,
}

var NewStatWeights_dpsWeight = simulate.SimResultStats{
	DPS:   0.97,
	DEATH: 0.01,
	TMI:   0.01,
	DTPS:  0.01,
}

var requiredStats = []stats.StatType{stats.Stat_Strength, stats.Stat_Stamina, stats.Stat_Crit, stats.Stat_Haste, stats.Stat_Expertise, stats.Stat_Mastery, stats.Stat_Dodge, stats.Stat_Parry}
var requiredSims = []simulate.SimResultType{simulate.Result_DPS, simulate.Result_DEATH, simulate.Result_TMI, simulate.Result_DTPS}

func CalcNewStatWeights(inputData []NewWeightInput, targetRatios simulate.SimResultStats, printer *util.PrintRecorder) {
	highRange := 100000.0 // basic sum of stats Mop P4 is about 83k
	// scaleMin := 0.001  // similar number of places, can get scores down to around 1
	colNames := make([]string, 0)

	input := new(inputBuilder)

	input.minimise = true
	input.blendMultiObjectives = true
	input.vars.addLinearObjective(1, 0, 1000, 0.2, 1) // contributionDiff
	input.vars.addLinearObjective(1, 0, 1000, 0.2, 2) // scoreDiff
	input.vars.addLinearObjective(0.1, 0, 1000, 0.2, 3) // data pairs offset

	statWeightColumns, colNames, simWeightColumns := weightColumns(input, colNames)

	colNames = dataEquations(inputData, input, highRange, colNames, statWeightColumns, simWeightColumns, targetRatios)

	colNames = dataCompareInPairsEquations(input, colNames, inputData, statWeightColumns, simWeightColumns, targetRatios, highRange)

	solution, log := input.runHighs()
	printer.AppendOther(log)
	printer.Println(solution.Status.String())

	reportWeightSolution(solution, printer, colNames, statWeightColumns, simWeightColumns, inputData)
}

func weightColumns(input *inputBuilder, colNames []string) (map[stats.StatType]columnIndex, []string, map[simulate.SimResultType]columnIndex) {
	// scaleMin := 0.0
	// scaleMax := 100.0

	totalStatWeightRow := constraintRowBuild{}
	statWeightColumns := make(map[stats.StatType]columnIndex)
	for _, stat := range requiredStats {
		// statColIndex := input.createColumnGeneral(highs.Continuous, scaleMin, scaleMax)
		statColIndex := input.createColumnGeneral(highs.Continuous, c_minusInf, c_plusInf)
		colNames = append(colNames, stat.Name())
		statWeightColumns[stat] = statColIndex
		totalStatWeightRow.add(statColIndex, 1)
	}
	totalStatWeightRow.finish(input, 1, 1)

	// force strength=1
	forceStrength := constraintRowBuild{}
	forceStrength.add(statWeightColumns[stats.Stat_Strength], 1)
	forceStrength.finish(input, 1, 1)

	totalSimWeightRow := constraintRowBuild{}
	simWeightColumns := make(map[simulate.SimResultType]columnIndex)
	for _, simType := range requiredSims {
		// simColIndex := input.createColumnGeneral(highs.Continuous, scaleMin, scaleMax)
		simColIndex := input.createColumnGeneral(highs.Continuous, c_minusInf, c_plusInf)
		colNames = append(colNames, simType.String())
		simWeightColumns[simType] = simColIndex
		totalSimWeightRow.add(simColIndex, 1)
	}
	// totalSimWeightRow.finish(input, 1, 1)
	return statWeightColumns, colNames, simWeightColumns
}

func dataEquations(inputData []NewWeightInput, input *inputBuilder, highRange float64, colNames []string, statWeightColumns map[stats.StatType]columnIndex, simWeightColumns map[simulate.SimResultType]columnIndex, targetRatios simulate.SimResultStats) []string {
	for _, data := range inputData {
		// add up weighted gear score for row
		gearScoreTotal := input.createColumnGeneral(highs.Continuous, 0, highRange)
		colNames = append(colNames, "gearScoreTotal")
		gearRow := constraintRowBuild{}
		for _, statType := range requiredStats {
			statWeightCol := statWeightColumns[statType]
			gearRow.add(statWeightCol, float64(data.TotalStat.Get(statType)))
		}
		gearRow.add(gearScoreTotal, -1)
		gearRow.finish(input, 0, 0)

		// add up weighted sim score for row
		simScoreTotal := input.createColumnGeneral(highs.Continuous, 0, highRange)
		colNames = append(colNames, "simScoreTotal")
		simScoreRow := constraintRowBuild{}
		for _, simType := range requiredSims {
			simWeightCol := simWeightColumns[simType]
			simScoreRow.add(simWeightCol, data.SimResult.GetFriendly(simType))
		}
		simScoreRow.add(simScoreTotal, -1)
		simScoreRow.finish(input, 0, 0)

		// contribution of sim part of simScoreTotal
		for _, simType := range requiredSims {
			contributionRow := constraintRowBuild{}

			// formula sum(simWeight*simResult[simType]) = simScoreTotal
			// simScoreTotal * 0.4 = simWeightCol * valuue[type]

			// calc simResult[simType]*simWeightCol[simType]
			simWeightCol := simWeightColumns[simType]
			contributionRow.add(simWeightCol, data.SimResult.GetFriendly(simType))

			// ideally we want that to equal simScoreTotal*targetRatio[simType]
			contributionRow.add(simScoreTotal, -targetRatios.Get(simType))

			// contributionDiff := input.createColumnGeneral(highs.Continuous, -highRange, highRange)
			contributionDiff := input.createColumnGeneral(highs.Continuous, c_minusInf, c_plusInf)
			colNames = append(colNames, "contributionDiff")
			contributionRow.add(contributionDiff, 1)
			contributionRow.finish(input, 0, 0)

			contributionDiffAbsOutput := input.createColumnForLinearObjective(highs.Continuous, 0, c_plusInf, 1, 0)
			colNames = append(colNames, "contributionDiffAbsOutput")
			absoluteValue2(input, contributionDiff, contributionDiffAbsOutput, highRange) // could maybe be narrower highRange
			// absoluteValue(input, contributionDiff, contributionDiffAbsOutput, highRange)
			// colNames = append(colNames, "absBool")
		}

		// scoreDiffSigned := input.createColumnGeneral(highs.Continuous, -highRange, highRange)
		scoreDiffSigned := input.createColumnGeneral(highs.Continuous, c_minusInf, c_plusInf)
		colNames = append(colNames, "scoreDiffSigned")
		scoreDiffRow := constraintRowBuild{}
		scoreDiffRow.add(gearScoreTotal, 1)
		scoreDiffRow.add(simScoreTotal, -1)
		scoreDiffRow.add(scoreDiffSigned, 1)
		scoreDiffRow.finish(input, 0, 0)

		diffAbsOutput := input.createColumnForLinearObjective(highs.Continuous, 0, c_plusInf, 1, 1)
		colNames = append(colNames, "diffAbsOutput")
		absoluteValue2(input, scoreDiffSigned, diffAbsOutput, highRange)
		// absoluteValue(input, scoreDiffSigned, diffAbsOutput, highRange)
		// colNames = append(colNames, "absBool")
	}
	return colNames
}

func dataCompareInPairsEquations(input *inputBuilder, colNames []string, inputData []NewWeightInput, statWeightColumns map[stats.StatType]columnIndex, simWeightColumns map[simulate.SimResultType]columnIndex, targetRatios simulate.SimResultStats, highRange float64) []string {
	detailedWeights := make(map[stats.StatType]map[simulate.SimResultType]columnIndex)
	for _, statType := range requiredStats {
		detailedWeights[statType] = make(map[simulate.SimResultType]columnIndex)
		for _, simType := range requiredSims {
			detailedWeights[statType][simType] = input.createColumnGeneral(highs.Continuous, c_minusInf, c_plusInf)
		}
	}

	inputData = inputData[0:20]
	for _, data1 := range inputData {
		for _, data2 := range inputData {
			if !stats.StatBlock_Equals(&data1.TotalStat, &data2.TotalStat) {
				colNames = dataCompareInPairsEquations_Single(input, colNames, data1, data2, detailedWeights, highRange)
			}
		}
	}

	for _, statType := range requiredStats {
		statRow := constraintRowBuild{}
		for _, simType := range requiredSims {
			statRow.add(detailedWeights[statType][simType], targetRatios.Get(simType))
		}
		statRow.add(statWeightColumns[statType], -1)
		statRow.finish(input, 0, 0)
	}

	return colNames
}

func dataCompareInPairsEquations_Single(input *inputBuilder, colNames []string, data1, data2 NewWeightInput, weights map[stats.StatType]map[simulate.SimResultType]columnIndex, highRange float64) []string {

	// so i'd normally do this in the area of a single stat and simType
	// Death's Haste improvement is F37/C37 = (death_with_0_haste_str - death_with_+600_haste)/(death_with_0_haste_str - death_with_+600_str)
	// overall haste improvement is 0.4 * pos_only(F37/C37)

	// we consider gearScoreTotal1/gearScoreTotal2? doesn't say much

	// MAYBE A NEW SET OF COEFFECTIENCTS FOR HASTE_DEATH etc
	// then we can define signedDiffDeath = (haste2-haste1)*hasteDeath + (crit2-crit1)*critDeath			{or this series expands out cleanly}

	// maybe we make an equasion for each

	// sim's percent difference should make sense for given stat combo
	// simTotal * 0.4

	for _, simType := range requiredSims {
		simRow := constraintRowBuild{}

		for _, statType := range requiredStats {
			statDiff := data1.TotalStat.Get(statType) - data2.TotalStat.Get(statType)
			simRow.add(weights[statType][simType], float64(statDiff))
		}

		offset := input.createColumnGeneral(highs.Continuous, c_minusInf, c_plusInf)
		simRow.add(offset, 1)
		offsetAbs := input.createColumnForLinearObjective(highs.Continuous, 0, c_plusInf, 1, 2)
		absoluteValue2(input, offset, offsetAbs, highRange)

		simDiff := data1.SimResult.Get(simType) - data2.SimResult.Get(simType)
		simRow.finish(input, simDiff, simDiff)
	}

	return colNames
}

func reportWeightSolution(solution *highs.Solution, printer *util.PrintRecorder, colNames []string, statWeightColumns map[stats.StatType]columnIndex, simWeightColumns map[simulate.SimResultType]columnIndex, inputData []NewWeightInput) {
	for i, x := range solution.ColValues {
		if i < len(colNames)  {
			printer.Printf("%3d %14f %s\n", i, x, colNames[i])
		}
	}

	statWeightResult := make(map[stats.StatType]float64)
	simWeightResult := make(map[simulate.SimResultType]float64)
	printer.Println("WEIGHTS")
	for _, statType := range requiredStats {
		statWeightCol := statWeightColumns[statType]
		statWeightResult[statType] = solution.ColValues[statWeightCol]
		printer.Printf("%10s %f\n", statType.Name(), statWeightResult[statType])
	}
	for _, simType := range requiredSims {
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
		for _, statType := range requiredStats {
			val := float64(data.TotalStat.Get(statType))
			weight := statWeightResult[statType]
			printer.Printf(" %10s %8.2f * %8.4f = %8.2f\n", statType.Name(), val, weight, val*weight)
			statSum += val * weight
		}
		printer.Printf("%46f\n", statSum)

		simSum := 0.0
		for _, simType := range requiredSims {
			val := data.SimResult.GetFriendly(simType)
			weight := simWeightResult[simType]
			simSum += val * weight
		}
		for _, simType := range requiredSims {
			val := data.SimResult.GetFriendly(simType)
			weight := simWeightResult[simType]
			printer.Printf(" %10s %12.2f * %8.4f = %8.2f (%.4f)\n", simType.String(), val, weight, val*weight, val*weight/simSum)
		}
		printer.Printf("%51f\n", simSum)

		printer.Println0()
	}
}
