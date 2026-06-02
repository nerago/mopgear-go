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

	scaleSims             map[simulate.SimResultType]float64
	scaleStats            map[stats.StatType]float64
	detailedWeightColumns util.MapMap[stats.StatType, simulate.SimResultType, utilhighs.ColumnIndex]
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
	compfit.input.Solver = "ipm"

	compfit.chooseSimScaling()
	compfit.chooseStatScaling()
	compfit.createWeightColumns()
	compfit.buildDataEquations()

	solution, log := compfit.input.RunHighs()
	compfit.printer.AppendOther(log)

	return compfit.exactAndReportSolution(solution)
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

		compfit.printer.Printf("scale %s %e\n", simType.String(), compfit.scaleSims[simType])
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

		compfit.printer.Printf("scale %s %.8f\n", statType.Name(), compfit.scaleStats[statType])
	}
}

func (compfit *ComplexStatWeightProcess) createWeightColumns() {
	minimumStrength := 0.0001

	for _, statType := range G_RequiredStats {
		for _, simType := range G_RequiredSims {
			// we don't want to be dealing with 0 strength since that's our base stat to scale against
			// however could be rejecting some special situtations where it actually is true
			lo := utilhighs.C_MinusInf
			hi := utilhighs.C_PlusInf
			// if statType == stats.Stat_Strength {
			// 	lo = minimumStrength
			// }

			colDetailWeight := compfit.input.CreateColumnGeneral(highs.Continuous, lo, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name() + " " + simType.String()})
			compfit.detailedWeightColumns.Put(statType, simType, colDetailWeight)
		}
	}

	// we don't want to be dealing with 0 strength since that's our base stat to scale against
	// however could be rejecting some special situtations where it actually is true
	for simType, colDetailWeight := range compfit.detailedWeightColumns.SeqInnerWithKey1Value(stats.Stat_Strength) {
		weightAbsoluteValue := compfit.input.CreateColumnGeneral(highs.Continuous, minimumStrength, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "WEIGHT ABS strength " + simType.String()})
		utilhighs.AbsoluteValue(compfit.input, colDetailWeight, weightAbsoluteValue)
	}
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

// equation is: weightA*scaledStatA + weightB*scaledStatB = scaledSimValue - diff
func (compfit *ComplexStatWeightProcess) buildDataEquationForSim(stats *stats.StatBlock, simValue float64, simType simulate.SimResultType) {
	matchSimValue := utilhighs.ConstraintRowBuild{}

	// TODO is there a way to flip the division for TMI DEATH etc, fundamental problem is that they don't increase linearly with stats

	for _, statType := range G_RequiredStats {
		weightDetailCol := compfit.detailedWeightColumns.GetOrPanic(statType, simType)
		statValue := float64(stats.Get(statType))
		statScale := compfit.scaleStats[statType]

		scaledStatValue := statValue * statScale
		matchSimValue.Add(weightDetailCol, scaledStatValue)
	}

	diffSigned := compfit.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "diffSigned"})
	matchSimValue.Add(diffSigned, 1)

	diffOutput := compfit.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugString{Text: "diffOutput"})
	utilhighs.AbsoluteValue(compfit.input, diffSigned, diffOutput)

	simScale := compfit.scaleSims[simType]
	scaledSimValue := simValue * simScale
	matchSimValue.Finish(compfit.input, scaledSimValue, scaledSimValue)
}

func (compfit *ComplexStatWeightProcess) exactAndReportSolution(solution *highs.Solution) map[stats.StatType]float64 {
	compfit.input.DebugPrintColumns(solution, compfit.printer)

	compfit.printer.Println("WEIGHTS")
	detailWeightMap := compfit.extractDetailWeights(solution)
	statWeightResult := compfit.computeFinalWeights(detailWeightMap)

	compfit.reportExamples(detailWeightMap)

	return statWeightResult
}

func (compfit *ComplexStatWeightProcess) extractDetailWeights(solution *highs.Solution) util.MapMap[stats.StatType, simulate.SimResultType, float64] {
	// extract and report on detail weights
	detailWeightMap := util.MapMap[stats.StatType, simulate.SimResultType, float64]{}
	for entry := range compfit.detailedWeightColumns.SeqWithKeys() {
		statType := entry.Key1
		simType := entry.Key2
		column := entry.Value

		modelWeight := solution.ColValues[column]

		// basic equation is: weightA*scaledStatA + weightB*scaledStatB = scaledSim - diff
		// taking one component: modelWeight * scaledStat = scaledSimValue
		// substitute in scaledStat = stat * statScale, scaledSim = sim * simScale, modelWeight = usableWeight * scaleFix
		//   -->   (usableWeight * scaleFix) * (stat * statScale) = (sim * simScale)
		//   -->   (usableWeight * scaleFix) = (sim * simScale) / (stat * statScale)
		//   -->     usableWeight * scaleFix = (sim * simScale) / (stat * statScale)
		//   -->                    scaleFix = ( (sim * simScale) / (stat * statScale) ) / usableWeight
		//   -->                    scaleFix = (sim / stat) * (simScale / statScale) / usableWeight
		// the essential equation we're kinda working on is weight = sim / stat, so that can cancel out
		//   -->                    scaleFix = (sim / stat) * (simScale / statScale) / (sim / stat)
		//   -->                    scaleFix = simScale / statScale
		// substituting back in: modelWeight = usableWeight * scaleFix
		//   --> modelWeight = usableWeight * scaleFix
		//   --> usableWeight = modelWeight / scaleFix

		scaleFix := compfit.scaleSims[simType] / compfit.scaleStats[statType]
		usableWeight := modelWeight / scaleFix

		if !simType.IsHighGood() {
			usableWeight *= -1
		}

		detailWeightMap.Put(statType, simType, usableWeight)

		compfit.printer.Printf("%10s %10s %11.8f (%5.2e) %11.8f (%5.2e)\n", statType.Name(), simType.String(), modelWeight, modelWeight, usableWeight, usableWeight)
	}
	compfit.printer.Println0()

	for entry := range detailWeightMap.SeqWithKeysOtherOrder() {
		usableWeight := entry.Value
		compfit.printer.Printf("%10s %10s %11.8f (%5.2e)\n", entry.Key1.Name(), entry.Key2.String(), usableWeight, usableWeight)
	}
	compfit.printer.Println0()
	return detailWeightMap
}

func (compfit *ComplexStatWeightProcess) reportExamples(detailWeightMap util.MapMap[stats.StatType, simulate.SimResultType, float64]) {
	for i := range 20 {
		data := compfit.inputData[i]
		compfit.printer.Println("EXAMPLE")

		for _, simType := range G_RequiredSims {
			statSum := 0.0
			compfit.printer.Printf(" %10s", simType.String())
			for _, statType := range G_RequiredStats {
				statValue := float64(data.TotalStat.Get(statType))
				weight := detailWeightMap.GetOrPanic(statType, simType)
				if !simType.IsHighGood() {
					weight *= -1 // unflip so equation appears matches what the model saw
				}
				compfit.printer.Printf(" {%s %.2f * %.4e = %.4f}", statType.Name(), statValue, weight, statValue*weight)
				statSum += statValue * weight
			}
			compfit.printer.Printf(" = %.4f (expect %.4f)\n", statSum, data.SimResult.Get(simType))
		}

		compfit.printer.Println0()
	}
}

func (compfit *ComplexStatWeightProcess) computeFinalWeights(detailWeightMap util.MapMap[stats.StatType, simulate.SimResultType, float64]) map[stats.StatType]float64 {
	statWeightResult := make(map[stats.StatType]float64)
	for statType, seqSimPairs := range detailWeightMap.SeqGroupsKey1NestedKeyValue() {
		sumIndividual := 0.0

		for simType, thisDetailWeight := range seqSimPairs {
			strengthDetailWeight := detailWeightMap.GetOrPanic(stats.Stat_Strength, simType)
			targetRatio := compfit.targetRatios.Get(simType)

			componentValue := targetRatio * thisDetailWeight / strengthDetailWeight

			sumIndividual += componentValue
		}

		statWeightResult[statType] = sumIndividual
	}

	return statWeightResult
}
