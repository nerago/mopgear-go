package stathighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

const c_highRange = 100000.0

type ComplexStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimResultStats
	inputData    []WeightInput

	input *utilhighs.InputBuilder

	linearContribution int
	linearScore        int

	statWeightColumns map[stats.StatType]utilhighs.ColumnIndex
	simWeightColumns  map[simulate.SimResultType]utilhighs.ColumnIndex

	// unitStatValues  util.MapMapSlice[stats.StatType, simulate.SimResultType, gridDataSample]
	// detailedWeights util.MapMap[stats.StatType, simulate.SimResultType, utilhighs.ColumnIndex]
	// finalWeights    map[stats.StatType]utilhighs.ColumnIndex
}

func (compfit *ComplexStatWeightProcess) Init(printer *util.PrintRecorder) {
	compfit.printer = printer
}

func (compfit *ComplexStatWeightProcess) SupplyData(inputData []WeightInput) {
	compfit.inputData = inputData
}

func (compfit *ComplexStatWeightProcess) SetTargetRatios(targetRatios simulate.SimResultStats) {
	compfit.targetRatios = targetRatios
}

func (compfit *ComplexStatWeightProcess) Run() map[stats.StatType]float64 {
	// highRange := 100000.0 // basic sum of stats Mop P4 is about 83k
	// scaleMin := 0.001  // similar number of places, can get scores down to around 1

	compfit.input = new(utilhighs.InputBuilder)
	compfit.input.Minimise = false
	compfit.input.BlendMultiObjectives = true
	compfit.input.Solver = "ipm"
	compfit.linearContribution = compfit.input.AddLinearBlended(1, 0) // contributionDiff
	compfit.linearScore = compfit.input.AddLinearBlended(2, 0)        // scoreDiff
	// compfit.input.AddLinearBlended(1, 0) // data pairs offset

	// compfit.input.AddLinearPrioritised(false, 1000, 0.2, 1) // contributionDiff
	// compfit.input.AddLinearPrioritised(false, 1000, 0.2, 2) // scoreDiff
	// compfit.input.AddLinearPrioritised(false, 1000, 0.2, 3) // data pairs offset

	compfit.createWeightColumns()
	compfit.buildDataEquations()

	solution, log := compfit.input.RunHighs()
	compfit.printer.AppendOther(log)
	compfit.printer.Println(solution.Status.String())

	return compfit.reportWeightSolution(solution)
}

func (compfit *ComplexStatWeightProcess) createWeightColumns() {
	scaleMin := 0.0
	scaleMax := 100.0

	totalStatWeightRow := utilhighs.ConstraintRowBuild{}
	compfit.statWeightColumns = make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, stat := range G_RequiredStats {
		statColIndex := compfit.input.CreateColumnGeneral(highs.Continuous, scaleMin, scaleMax, utilhighs.DebugString{Text: stat.Name()})
		// statColIndex := input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
		compfit.statWeightColumns[stat] = statColIndex
		totalStatWeightRow.Add(statColIndex, 1)
	}
	totalStatWeightRow.Finish(compfit.input, 1, 1)

	// force strength=1
	forceStrength := utilhighs.ConstraintRowBuild{}
	forceStrength.Add(compfit.statWeightColumns[stats.Stat_Strength], 1)
	forceStrength.Finish(compfit.input, 1, 1)

	totalSimWeightRow := utilhighs.ConstraintRowBuild{}
	compfit.simWeightColumns = make(map[simulate.SimResultType]utilhighs.ColumnIndex)
	for _, simType := range G_RequiredSims {
		simColIndex := compfit.input.CreateColumnGeneral(highs.Continuous, scaleMin, scaleMax, utilhighs.DebugString{Text: simType.String()})
		// simColIndex := input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
		compfit.simWeightColumns[simType] = simColIndex
		totalSimWeightRow.Add(simColIndex, 1)
	}
	// totalSimWeightRow.finish(input, 1, 1)
}

func (compfit *ComplexStatWeightProcess) buildDataEquations() {
	for data := range util.ForPointer(compfit.inputData) {
		compfit.buildDataEquationForInput(data)
	}
}

func (compfit *ComplexStatWeightProcess) buildDataEquationForInput(data *WeightInput) {
	// add up weighted gear score for row
	gearScoreTotal := compfit.input.CreateColumnGeneral(highs.Continuous, 0, c_highRange, utilhighs.DebugString{Text: "gearScoreTotal"})
	gearRow := utilhighs.ConstraintRowBuild{}
	for _, statType := range G_RequiredStats {
		statWeightCol := compfit.statWeightColumns[statType]
		gearRow.Add(statWeightCol, float64(data.TotalStat.Get(statType)))
	}
	gearRow.Add(gearScoreTotal, -1)
	gearRow.Finish(compfit.input, 0, 0)

	// add up weighted sim score for row
	simScoreTotal := compfit.input.CreateColumnGeneral(highs.Continuous, 0, c_highRange, utilhighs.DebugString{Text: "simScoreTotal"})
	simScoreRow := utilhighs.ConstraintRowBuild{}
	for _, simType := range G_RequiredSims {
		simWeightCol := compfit.simWeightColumns[simType]
		simScoreRow.Add(simWeightCol, data.SimResult.GetFriendly(simType))
	}
	simScoreRow.Add(simScoreTotal, -1)
	simScoreRow.Finish(compfit.input, 0, 0)

	// contribution of sim part of simScoreTotal
	for _, simType := range G_RequiredSims {
		contributionRow := utilhighs.ConstraintRowBuild{}

		// formula sum(simWeight*simResult[simType]) = simScoreTotal
		// simScoreTotal * 0.4 = simWeightCol * valuue[type]

		// calc simResult[simType]*simWeightCol[simType]
		simWeightCol := compfit.simWeightColumns[simType]
		contributionRow.Add(simWeightCol, data.SimResult.GetFriendly(simType))

		// ideally we want that to equal simScoreTotal*targetRatio[simType]
		contributionRow.Add(simScoreTotal, -compfit.targetRatios.Get(simType))

		// diff = simWeightCol[type] * thisResult[type] - thisScoreTotal * targetRatio[type]
		// ideally for comparison sake the diff should look like the ratio
		// simWeightCol[type] * thisResult[type] - thisScoreTotal * (targetRatio[type] + diff) = 0
		// simWeightCol[type] * thisResult[type] - thisScoreTotal * (targetRatio[type] + diff) = 0

		contributionDiff := compfit.input.CreateColumnGeneral(highs.Continuous, -c_highRange, c_highRange, utilhighs.DebugString{Text: "contributionDiff"})
		// contributionDiff := input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
		contributionRow.Add(contributionDiff, 1)
		contributionRow.Finish(compfit.input, 0, 0)

		contributionDiffAbsOutput := compfit.input.CreateColumnForLinearObjective(highs.Continuous, 0, c_highRange, 1, 0, utilhighs.DebugString{Text: "contributionDiffAbsOutput"})
		// contributionDiffAbsOutput := input.CreateColumnForLinearObjective(highs.Continuous, 0, utilhighs.C_PlusInf, 1, 0)
		utilhighs.AbsoluteValue(compfit.input, contributionDiff, contributionDiffAbsOutput) // could maybe be narrower highRange
		// absoluteValue(input, contributionDiff, contributionDiffAbsOutput, highRange)
		// colNames = append(colNames, "absBool")
	}

	scoreDiffSigned := compfit.input.CreateColumnGeneral(highs.Continuous, -c_highRange, c_highRange, utilhighs.DebugString{Text: "scoreDiffSigned"})
	// scoreDiffSigned := input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
	scoreDiffRow := utilhighs.ConstraintRowBuild{}
	scoreDiffRow.Add(gearScoreTotal, 1)
	scoreDiffRow.Add(simScoreTotal, -1)
	scoreDiffRow.Add(scoreDiffSigned, 1)
	scoreDiffRow.Finish(compfit.input, 0, 0)

	// diffAbsOutput := input.CreateColumnForLinearObjective(highs.Continuous, 0, utilhighs.C_PlusInf, 1, 1)
	diffAbsOutput := compfit.input.CreateColumnForLinearObjective(highs.Continuous, 0, c_highRange, 1, 1, utilhighs.DebugString{Text: "diffAbsOutput"})
	utilhighs.AbsoluteValue(compfit.input, scoreDiffSigned, diffAbsOutput)
	// absoluteValue(input, scoreDiffSigned, diffAbsOutput, highRange)
	// colNames = append(colNames, "absBool")
}

func (compfit *ComplexStatWeightProcess) reportWeightSolution(solution *highs.Solution) map[stats.StatType]float64 {
	compfit.input.DebugPrintColumns(solution, compfit.printer)

	statWeightResult := make(map[stats.StatType]float64)
	simWeightResult := make(map[simulate.SimResultType]float64)
	compfit.printer.Println("WEIGHTS")
	for _, statType := range G_RequiredStats {
		statWeightCol := compfit.statWeightColumns[statType]
		statWeightResult[statType] = solution.ColValues[statWeightCol]
		compfit.printer.Printf("%10s %f\n", statType.Name(), statWeightResult[statType])
	}
	for _, simType := range G_RequiredSims {
		simWeightCol := compfit.simWeightColumns[simType]
		simWeightResult[simType] = solution.ColValues[simWeightCol]
		compfit.printer.Printf("%10s %f\n", simType.String(), simWeightResult[simType])
	}

	for i := range 5 {
		data := compfit.inputData[i]
		compfit.printer.Println("EXAMPLE")

		// formula sum(simWeight*simResult[simType]) = simScoreTotal
		// simScoreTotal * 0.4 = simWeightCol * valuue[type]

		// calc simResult[simType]*simWeightCol[simType]

		// ideally we want that to equal simScoreTotal*targetRatio[simType]

		statSum := 0.0
		for _, statType := range G_RequiredStats {
			val := float64(data.TotalStat.Get(statType))
			weight := statWeightResult[statType]
			compfit.printer.Printf(" %10s %8.2f * %8.4f = %8.2f\n", statType.Name(), val, weight, val*weight)
			statSum += val * weight
		}
		compfit.printer.Printf("%46f\n", statSum)

		simSum := 0.0
		for _, simType := range G_RequiredSims {
			val := data.SimResult.GetFriendly(simType)
			weight := simWeightResult[simType]
			simSum += val * weight
		}
		for _, simType := range G_RequiredSims {
			val := data.SimResult.GetFriendly(simType)
			weight := simWeightResult[simType]
			compfit.printer.Printf(" %10s %12.2f * %8.4f = %8.2f (%.4f)\n", simType.String(), val, weight, val*weight, val*weight/simSum)
		}
		compfit.printer.Printf("%51f\n", simSum)

		compfit.printer.Println0()
	}

	return statWeightResult
}
