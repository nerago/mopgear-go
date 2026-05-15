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
	scaleMin := 0.0  // similar number of places, can get scores down to around 1
	scaleMax := 100000.0
	colNames := make([]string, 0)

	input := new(inputBuilder)

	totalStatWeightRow := constraintRowBuild{}
	statWeightColumns := make(map[stats.StatType]int)
	for _, stat := range requiredStats {
		statColIndex := input.createColumnGeneral(highs.Continuous, scaleMin, scaleMax)
		colNames = append(colNames, stat.Name())
		statWeightColumns[stat] = statColIndex
		totalStatWeightRow.add(statColIndex, 1)
	}
	// totalStatWeightRow.finish(input, 1, 1)

	totalSimWeightRow := constraintRowBuild{}
	simWeightColumns := make(map[simulate.SimResultType]int)
	for _, simType := range requiredSims {
		simColIndex := input.createColumnGeneral(highs.Continuous, scaleMin, scaleMax)
		colNames = append(colNames, simType.String())
		simWeightColumns[simType] = simColIndex
		totalSimWeightRow.add(simColIndex, 1)
	}
	totalSimWeightRow.finish(input, 1, 1)

	for _, data := range inputData {
		// add up weighted gear score for row
		gearScoreTotal := input.createColumnGeneral(highs.Continuous, 0, highRange)
		colNames = append(colNames, "gearScoreTotal")
		gearRow := constraintRowBuild{}
		for statType, statWeightCol := range statWeightColumns {
			gearRow.add(statWeightCol, float64(data.TotalStat.Get(statType)))
		}
		gearRow.add(gearScoreTotal, -1)
		gearRow.finish(input, 0, 0)

		// add up weighted sim score for row
		simScoreTotal := input.createColumnGeneral(highs.Continuous, 0, highRange)
		colNames = append(colNames, "simScoreTotal")
		simScoreRow := constraintRowBuild{}
		for simType, simWeightCol := range simWeightColumns {
			simScoreRow.add(simWeightCol, data.SimResult.Get(simType))
		}
		simScoreRow.add(simScoreTotal, -1)
		simScoreRow.finish(input, 0, 0)

		// contribution of sim part of simScoreTotal
		for simType, simWeightCol := range simWeightColumns {
			contributionRow := constraintRowBuild{}

			// calc simResult[simType]*simWeightCol[simType]
			contributionRow.add(simWeightCol, data.SimResult.Get(simType))

			// ideally we want that to equal simScoreTotal*targetRatio[simType]
			contributionRow.add(simScoreTotal, targetRatios.Get(simType))

			contributionDiff := input.createColumnGeneral(highs.Continuous, -highRange, highRange)
			colNames = append(colNames, "contributionDiff")
			contributionRow.add(contributionDiff, 1)
			contributionRow.finish(input, 0, 0)

			contributionDiffAbsOutput := input.createColumnWithOutput(highs.Continuous, 0, c_plusInf, 1)
			colNames = append(colNames, "contributionDiffAbsOutput")
			absoluteValue2(input, contributionDiff, contributionDiffAbsOutput, highRange*2) // could maybe be narrower highRange
			// colNames = append(colNames, "absBool")
		}

		scoreDiffSigned := input.createColumnGeneral(highs.Continuous, -highRange, highRange)
		colNames = append(colNames, "scoreDiffSigned")
		scoreDiffRow := constraintRowBuild{}
		scoreDiffRow.add(gearScoreTotal, 1)
		scoreDiffRow.add(simScoreTotal, -1)
		scoreDiffRow.add(scoreDiffSigned, 1)
		scoreDiffRow.finish(input, 0, 0)

		diffAbsOutput := input.createColumnWithOutput(highs.Continuous, 0, c_plusInf, 1)
		colNames = append(colNames, "diffAbsOutput")
		absoluteValue2(input, scoreDiffSigned, diffAbsOutput, highRange*2)
		// colNames = append(colNames, "absBool")
	}

	// solution, log := input.runHighs()
	solver, logFilename := input.preHighsRun()
	checkError(solver.SetMaximize(false))
	solution, err := highsPool.RunSolverUnderMutex(solver)
	checkError(err)
	log := input.postHighsRun(solver, logFilename)

	printer.AppendOther(log)
	printer.Println(solution.Status.String())

	for i, x := range solution.ColValues {
		printer.Printf("%3d %14f %s\n", i, x, colNames[i])
	}
}
