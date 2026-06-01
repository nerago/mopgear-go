package stathighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

// const c_ComplexHighRange = 100000.0

type ComplexStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimResultStats
	inputData    []WeightInput

	input *utilhighs.InputBuilder

	// linearContribution int
	// linearScore        int

	// simWeightColumns  map[simulate.SimResultType]utilhighs.ColumnIndex

	scaleSims        map[simulate.SimResultType]float64
	scaleStats       map[stats.StatType]float64
	detailedWeights  util.MapMap[stats.StatType, simulate.SimResultType, utilhighs.ColumnIndex]
	finalStatWeights map[stats.StatType]utilhighs.ColumnIndex
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
	compfit.input = new(utilhighs.InputBuilder)
	compfit.input.Minimise = true
	// compfit.input.BlendMultiObjectives = true
	compfit.input.Solver = "ipm"
	// compfit.linearContribution = compfit.input.AddLinearBlended(1, 0) // contributionDiff
	// compfit.linearScore = compfit.input.AddLinearBlended(2, 0)        // scoreDiff

	// compfit.input.AddLinearPrioritised(false, 1000, 0.2, 1) // contributionDiff
	// compfit.input.AddLinearPrioritised(false, 1000, 0.2, 2) // scoreDiff
	// compfit.input.AddLinearPrioritised(false, 1000, 0.2, 3) // data pairs offset

	compfit.chooseSimScaling()
	compfit.chooseStatScaling()
	compfit.createWeightColumns()
	compfit.buildDataEquations()

	solution, log := compfit.input.RunHighs()
	compfit.printer.AppendOther(log)
	compfit.printer.Println(solution.Status.String())

	return compfit.reportWeightSolution(solution)
}

func (compfit *ComplexStatWeightProcess) chooseSimScaling() {
	c_targetNumber := 1.0
	compfit.scaleSims = make(map[simulate.SimResultType]float64)
	for _, simType := range G_RequiredSims {
		total := 0.0
		for data := range util.ForPointer(compfit.inputData) {
			total += data.SimResult.Get(simType)
		}

		average := total / float64(len(compfit.inputData))
		if average != 0 {
			scale := c_targetNumber / average
			compfit.scaleSims[simType] = scale
		} else {
			compfit.scaleSims[simType] = 1
		}

		compfit.printer.Printf("scale %s %f\n", simType.String(), compfit.scaleSims[simType])
	}
}

func (compfit *ComplexStatWeightProcess) chooseStatScaling() {
	c_targetNumber := 1.0
	compfit.scaleStats = make(map[stats.StatType]float64)
	for _, statType := range G_RequiredStats {
		total := 0.0
		for data := range util.ForPointer(compfit.inputData) {
			total += float64(data.TotalStat.Get(statType))
		}

		average := total / float64(len(compfit.inputData))
		if average != 0 {
			scale := c_targetNumber / average
			compfit.scaleStats[statType] = scale
		} else {
			compfit.scaleStats[statType] = 1
		}

		compfit.printer.Printf("scale %s %f\n", statType.Name(), compfit.scaleStats[statType])
	}
}

func (compfit *ComplexStatWeightProcess) createWeightColumns() {
	weightRange := 10.0

	// total stat columns
	compfit.finalStatWeights = make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, stat := range G_RequiredStats {
		statColIndex := compfit.input.CreateColumnGeneral(highs.Continuous, -weightRange, weightRange, utilhighs.DebugString{Text: "FINALWEIGHT " + stat.Name()})
		compfit.finalStatWeights[stat] = statColIndex
	}

	// detailed columns
	for _, statType := range G_RequiredStats {
		for _, simType := range G_RequiredSims {
			colDetailWeight := compfit.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "WEIGHT " + statType.Name() + " " + simType.String()})
			compfit.detailedWeights.Put(statType, simType, colDetailWeight)
		}
	}

	// force strength
	// for _, simType := range G_RequiredSims {
	// 	value := compfit.targetRatios.Get(simType)
	// 	column := compfit.detailedWeights.GetOrPanic(stats.Stat_Strength, simType)
	// 	setStrengthFromRatio := utilhighs.ConstraintRowBuild{}
	// 	setStrengthFromRatio.Debug = "setStrengthFromRatio " + simType.String()
	// 	setStrengthFromRatio.Add(column, 1)
	// 	// maybe should flip the sign for simType.IsHighGood???
	// 	setStrengthFromRatio.Finish(compfit.input, value, value)
	// }

	// total is sum of detailed
	// NOTE not sure we can add these without scaling
	for _, stat := range G_RequiredStats {
		sumTotalStat := utilhighs.ConstraintRowBuild{}
		sumTotalStat.Debug = "sumTotalStat " + stat.Name()
		for simType, detailCol := range compfit.detailedWeights.SeqInnerWithKey1Value(stat) {
			if simType.IsHighGood() {
				sumTotalStat.Add(detailCol, 1)
			} else {
				sumTotalStat.Add(detailCol, -1)
			}
		}
		sumTotalStat.Add(compfit.finalStatWeights[stat], -1)
		sumTotalStat.Finish(compfit.input, 0, 0)
	}

	// totalSimWeightRow := utilhighs.ConstraintRowBuild{}
	// compfit.simWeightColumns = make(map[simulate.SimResultType]utilhighs.ColumnIndex)
	// for _, simType := range G_RequiredSims {
	// 	simColIndex := compfit.input.CreateColumnGeneral(highs.Continuous, scaleMin, scaleMax, utilhighs.DebugString{Text: simType.String()})
	// 	// simColIndex := input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
	// 	compfit.simWeightColumns[simType] = simColIndex
	// 	totalSimWeightRow.Add(simColIndex, 1)
	// }
	// // totalSimWeightRow.finish(input, 1, 1)
}

func (compfit *ComplexStatWeightProcess) buildDataEquations() {
	for data := range util.ForPointer(compfit.inputData) {
		compfit.buildDataEquationForInput(data)
	}
}

func (compfit *ComplexStatWeightProcess) buildDataEquationForInput(data *WeightInput) {
	for _, simType := range G_RequiredSims {
		compfit.buildDataEquationForSim(&data.TotalStat, data.SimResult.Get(simType), simType)
	}
}

func (compfit *ComplexStatWeightProcess) buildDataEquationForSim(stats *stats.StatBlock, simValue float64, simType simulate.SimResultType) {
	matchSimValue := utilhighs.ConstraintRowBuild{}

	for _, statType := range G_RequiredStats {
		detailCol := compfit.detailedWeights.GetOrPanic(statType, simType)
		statValue := float64(stats.Get(statType))
		statScale := compfit.scaleStats[statType]
		matchSimValue.Add(detailCol, statValue*statScale)
	}

	diffSigned := compfit.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "diffSigned"})
	matchSimValue.Add(diffSigned, 1)

	diffOutput := compfit.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: "diffOutput"})
	utilhighs.AbsoluteValue(compfit.input, diffSigned, diffOutput)

	simScale := compfit.scaleSims[simType]
	simValue *= simScale
	matchSimValue.Finish(compfit.input, simValue, simValue)
}

// func (compfit *ComplexStatWeightProcess) buildDataEquationForInputOld(data *WeightInput) {
// 	// add up weighted gear score for row
// 	gearScoreTotal := compfit.input.CreateColumnGeneral(highs.Continuous, 0, c_highRange, utilhighs.DebugString{Text: "gearScoreTotal"})
// 	gearRow := utilhighs.ConstraintRowBuild{}
// 	for _, statType := range G_RequiredStats {
// 		statWeightCol := compfit.finalStatWeights[statType]
// 		gearRow.Add(statWeightCol, float64(data.TotalStat.Get(statType)))
// 	}
// 	gearRow.Add(gearScoreTotal, -1)
// 	gearRow.Finish(compfit.input, 0, 0)

// 	// add up weighted sim score for row
// 	simScoreTotal := compfit.input.CreateColumnGeneral(highs.Continuous, 0, c_highRange, utilhighs.DebugString{Text: "simScoreTotal"})
// 	simScoreRow := utilhighs.ConstraintRowBuild{}
// 	for _, simType := range G_RequiredSims {
// 		simWeightCol := compfit.simWeightColumns[simType]
// 		simScoreRow.Add(simWeightCol, data.SimResult.GetFriendly(simType))
// 	}
// 	simScoreRow.Add(simScoreTotal, -1)
// 	simScoreRow.Finish(compfit.input, 0, 0)

// 	// contribution of sim part of simScoreTotal
// 	for _, simType := range G_RequiredSims {
// 		contributionRow := utilhighs.ConstraintRowBuild{}

// 		// formula sum(simWeight*simResult[simType]) = simScoreTotal
// 		// simScoreTotal * 0.4 = simWeightCol * valuue[type]

// 		// calc simResult[simType]*simWeightCol[simType]
// 		simWeightCol := compfit.simWeightColumns[simType]
// 		contributionRow.Add(simWeightCol, data.SimResult.GetFriendly(simType))

// 		// ideally we want that to equal simScoreTotal*targetRatio[simType]
// 		contributionRow.Add(simScoreTotal, -compfit.targetRatios.Get(simType))

// 		// diff = simWeightCol[type] * thisResult[type] - thisScoreTotal * targetRatio[type]
// 		// ideally for comparison sake the diff should look like the ratio
// 		// simWeightCol[type] * thisResult[type] - thisScoreTotal * (targetRatio[type] + diff) = 0
// 		// simWeightCol[type] * thisResult[type] - thisScoreTotal * (targetRatio[type] + diff) = 0

// 		contributionDiff := compfit.input.CreateColumnGeneral(highs.Continuous, -c_highRange, c_highRange, utilhighs.DebugString{Text: "contributionDiff"})
// 		// contributionDiff := input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
// 		contributionRow.Add(contributionDiff, 1)
// 		contributionRow.Finish(compfit.input, 0, 0)

// 		contributionDiffAbsOutput := compfit.input.CreateColumnForLinearObjective(highs.Continuous, 0, c_highRange, 1, 0, utilhighs.DebugString{Text: "contributionDiffAbsOutput"})
// 		// contributionDiffAbsOutput := input.CreateColumnForLinearObjective(highs.Continuous, 0, utilhighs.C_PlusInf, 1, 0)
// 		utilhighs.AbsoluteValue(compfit.input, contributionDiff, contributionDiffAbsOutput) // could maybe be narrower highRange
// 		// absoluteValue(input, contributionDiff, contributionDiffAbsOutput, highRange)
// 		// colNames = append(colNames, "absBool")
// 	}

// 	scoreDiffSigned := compfit.input.CreateColumnGeneral(highs.Continuous, -c_highRange, c_highRange, utilhighs.DebugString{Text: "scoreDiffSigned"})
// 	// scoreDiffSigned := input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
// 	scoreDiffRow := utilhighs.ConstraintRowBuild{}
// 	scoreDiffRow.Add(gearScoreTotal, 1)
// 	scoreDiffRow.Add(simScoreTotal, -1)
// 	scoreDiffRow.Add(scoreDiffSigned, 1)
// 	scoreDiffRow.Finish(compfit.input, 0, 0)

// 	// diffAbsOutput := input.CreateColumnForLinearObjective(highs.Continuous, 0, utilhighs.C_PlusInf, 1, 1)
// 	diffAbsOutput := compfit.input.CreateColumnForLinearObjective(highs.Continuous, 0, c_highRange, 1, 1, utilhighs.DebugString{Text: "diffAbsOutput"})
// 	utilhighs.AbsoluteValue(compfit.input, scoreDiffSigned, diffAbsOutput)
// 	// absoluteValue(input, scoreDiffSigned, diffAbsOutput, highRange)
// 	// colNames = append(colNames, "absBool")
// }

// func dataCompareInPairsEquations(input *utilhighs.InputBuilder, inputData []WeightInput, statWeightColumns map[stats.StatType]utilhighs.ColumnIndex, targetRatios simulate.SimResultStats) {
// 	detailedWeights := make(map[stats.StatType]map[simulate.SimResultType]utilhighs.ColumnIndex)
// 	for _, statType := range G_RequiredStats {
// 		detailedWeights[statType] = make(map[simulate.SimResultType]utilhighs.ColumnIndex)
// 		for _, simType := range G_RequiredSims {
// 			detailedWeights[statType][simType] = input.CreateColumnGeneral(highs.Continuous, -c_highRange, c_highRange, nil)
// 			// detailedWeights[statType][simType] = input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
// 		}
// 	}

// 	inputData = inputData[0:20]
// 	for _, data1 := range inputData {
// 		for _, data2 := range inputData {
// 			if !stats.StatBlock_Equals(&data1.TotalStat, &data2.TotalStat) {
// 				dataCompareInPairsEquations_Single(input, data1, data2, detailedWeights, c_highRange)
// 			}
// 		}
// 	}

// 	for _, statType := range G_RequiredStats {
// 		statRow := utilhighs.ConstraintRowBuild{}
// 		for _, simType := range G_RequiredSims {
// 			statRow.Add(detailedWeights[statType][simType], targetRatios.Get(simType))
// 		}
// 		statRow.Add(statWeightColumns[statType], -1)
// 		statRow.Finish(input, 0, 0)
// 	}
// }

// func dataCompareInPairsEquations_Single(input *utilhighs.InputBuilder, data1, data2 WeightInput, weights map[stats.StatType]map[simulate.SimResultType]utilhighs.ColumnIndex, highRange float64) {

// 	// so i'd normally do this in the area of a single stat and simType
// 	// Death's Haste improvement is F37/C37 = (death_with_0_haste_str - death_with_+600_haste)/(death_with_0_haste_str - death_with_+600_str)
// 	// overall haste improvement is 0.4 * pos_only(F37/C37)

// 	// we consider gearScoreTotal1/gearScoreTotal2? doesn't say much

// 	// MAYBE A NEW SET OF COEFFECTIENCTS FOR HASTE_DEATH etc
// 	// then we can define signedDiffDeath = (haste2-haste1)*hasteDeath + (crit2-crit1)*critDeath			{or this series expands out cleanly}

// 	// maybe we make an equasion for each

// 	// sim's percent difference should make sense for given stat combo
// 	// simTotal * 0.4

// 	for _, simType := range G_RequiredSims {
// 		simRow := utilhighs.ConstraintRowBuild{}

// 		for _, statType := range G_RequiredStats {
// 			statDiff := data1.TotalStat.Get(statType) - data2.TotalStat.Get(statType)
// 			simRow.Add(weights[statType][simType], float64(statDiff))
// 		}

// 		// offset := input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf)
// 		offset := input.CreateColumnGeneral(highs.Continuous, -highRange, highRange, nil)
// 		simRow.Add(offset, 1)
// 		offsetAbs := input.CreateColumnForLinearObjective(highs.Continuous, 0, utilhighs.C_PlusInf, 1, 2, nil)
// 		utilhighs.AbsoluteValue(input, offset, offsetAbs)

// 		var simDiff float64
// 		if simType.IsHighGood() {
// 			simDiff = data1.SimResult.Get(simType) - data2.SimResult.Get(simType)
// 		} else {
// 			simDiff = data2.SimResult.Get(simType) - data1.SimResult.Get(simType)
// 		}
// 		simRow.Finish(input, simDiff, simDiff)
// 	}
// }

func (compfit *ComplexStatWeightProcess) reportWeightSolution(solution *highs.Solution) map[stats.StatType]float64 {
	compfit.input.DebugPrintColumns(solution, compfit.printer)

	statWeightResult := make(map[stats.StatType]float64)
	compfit.printer.Println("WEIGHTS")
	for _, statType := range G_RequiredStats {
		statWeightCol := compfit.finalStatWeights[statType]
		weight := solution.ColValues[statWeightCol]
		weightCorrectedScale := weight / compfit.scaleStats[statType]
		statWeightResult[statType] = weightCorrectedScale
		compfit.printer.Printf("%10s %f %f\n", statType.Name(), weight, weightCorrectedScale)
	}
	compfit.printer.Println0()

	detailWeight := util.MapMap[stats.StatType, simulate.SimResultType, float64]{}
	for entry := range compfit.detailedWeights.SeqWithKeys() {
		column := entry.Value
		weight := solution.ColValues[column]
		weightCorrectedScale := weight * compfit.scaleSims[entry.Key2] / compfit.scaleStats[entry.Key1]
		detailWeight.Put(entry.Key1, entry.Key2, weightCorrectedScale)
		compfit.printer.Printf("%10s %10s %f %f\n", entry.Key1.Name(), entry.Key2.String(), weight, weightCorrectedScale)
	}
	compfit.printer.Println0()

	for entry := range compfit.detailedWeights.SeqWithKeysOtherOrder() {
		column := entry.Value
		weight := solution.ColValues[column]
		weightCorrectedScale := weight * compfit.scaleSims[entry.Key2] / compfit.scaleStats[entry.Key1]
		compfit.printer.Printf("%10s %10s %f %f\n", entry.Key1.Name(), entry.Key2.String(), weight, weightCorrectedScale)
	}
	compfit.printer.Println0()

	for i := range 5 {
		data := compfit.inputData[i]
		compfit.printer.Println("EXAMPLE")

		for _, simType := range G_RequiredSims {
			statSum := 0.0
			compfit.printer.Printf(" %10s", simType.String())
			for _, statType := range G_RequiredStats {
				statValue := float64(data.TotalStat.Get(statType))
				weight := detailWeight.GetOrPanic(statType, simType)
				compfit.printer.Printf(" {%s %.2f * %.4f = %.2f}", statType.Name(), statValue, weight, statValue*weight)
				statSum += statValue * weight
			}
			compfit.printer.Printf(" = %.2f (expect %.2f)\n", statSum, data.SimResult.Get(simType))
		}

		compfit.printer.Println0()
	}

	return statWeightResult
}
