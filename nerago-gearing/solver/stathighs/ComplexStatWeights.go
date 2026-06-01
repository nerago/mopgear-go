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

	scaleSims       map[simulate.SimResultType]float64
	scaleStats      map[stats.StatType]float64
	detailedWeights util.MapMap[stats.StatType, simulate.SimResultType, utilhighs.ColumnIndex]
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

		compfit.printer.Printf("scale %s %f\n", statType.Name(), compfit.scaleStats[statType])
	}
}

func (compfit *ComplexStatWeightProcess) createWeightColumns() {
	// detailed columns
	for _, statType := range G_RequiredStats {
		for _, simType := range G_RequiredSims {
			colDetailWeight := compfit.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "WEIGHT " + statType.Name() + " " + simType.String()})
			compfit.detailedWeights.Put(statType, simType, colDetailWeight)
		}
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

func (compfit *ComplexStatWeightProcess) exactAndReportSolution(solution *highs.Solution) map[stats.StatType]float64 {
	compfit.input.DebugPrintColumns(solution, compfit.printer)

	compfit.printer.Println("WEIGHTS")
	detailWeightMapForCompute, detailWeightMapForExample := compfit.extractDetailWeights(solution)
	statWeightResult := compfit.computeFinalWeights(detailWeightMapForCompute)

	compfit.reportExamples(detailWeightMapForExample)

	return statWeightResult
}

func (compfit *ComplexStatWeightProcess) extractDetailWeights(solution *highs.Solution) (util.MapMap[stats.StatType, simulate.SimResultType, float64], util.MapMap[stats.StatType, simulate.SimResultType, float64]) {
	// extract and report on detail weights
	detailWeightMapForCompute := util.MapMap[stats.StatType, simulate.SimResultType, float64]{}
	detailWeightMapForExample := util.MapMap[stats.StatType, simulate.SimResultType, float64]{}
	for entry := range compfit.detailedWeights.SeqWithKeys() {
		column := entry.Value
		weight := solution.ColValues[column]

		weightCorrectedPreserveRelative := weight * compfit.scaleStats[entry.Key1]
		detailWeightMapForCompute.Put(entry.Key1, entry.Key2, weightCorrectedPreserveRelative)

		weightCorrectedBothScale := weight * compfit.scaleStats[entry.Key1] / compfit.scaleSims[entry.Key2]
		detailWeightMapForExample.Put(entry.Key1, entry.Key2, weightCorrectedBothScale)

		compfit.printer.Printf("%10s %10s %11.8f %11.8f %11.8f\n", entry.Key1.Name(), entry.Key2.String(), weight, weightCorrectedPreserveRelative, weightCorrectedBothScale)
	}
	compfit.printer.Println0()

	for entry := range detailWeightMapForCompute.SeqWithKeysOtherOrder() {
		weightCorrectedPreserveRelative := entry.Value
		weightCorrectedBothScale := detailWeightMapForExample.GetOrPanic(entry.Key1, entry.Key2)
		compfit.printer.Printf("%10s %10s %11.8f %11.8f\n", entry.Key1.Name(), entry.Key2.String(), weightCorrectedPreserveRelative, weightCorrectedBothScale)
	}
	compfit.printer.Println0()
	return detailWeightMapForCompute, detailWeightMapForExample
}

func (compfit *ComplexStatWeightProcess) reportExamples(detailWeightMapForExample util.MapMap[stats.StatType, simulate.SimResultType, float64]) {
	for i := range 5 {
		data := compfit.inputData[i]
		compfit.printer.Println("EXAMPLE")

		for _, simType := range G_RequiredSims {
			statSum := 0.0
			compfit.printer.Printf(" %10s", simType.String())
			for _, statType := range G_RequiredStats {
				statValue := float64(data.TotalStat.Get(statType))
				weight := detailWeightMapForExample.GetOrPanic(statType, simType)
				compfit.printer.Printf(" {%s %.2f * %.4e = %.4f}", statType.Name(), statValue, weight, statValue*weight)
				statSum += statValue * weight
			}
			compfit.printer.Printf(" = %.4f (expect %.4f)\n", statSum, data.SimResult.Get(simType))
		}

		compfit.printer.Println0()
	}
}

func (compfit *ComplexStatWeightProcess) computeFinalWeights(detailWeightMapForCompute util.MapMap[stats.StatType, simulate.SimResultType, float64]) map[stats.StatType]float64 {
	statWeightResult := make(map[stats.StatType]float64)
	for statType, seqSimPairs := range detailWeightMapForCompute.SeqGroupsKey1NestedKeyValue() {
		sumIndividual := 0.0

		for simType, thisDetailWeight := range seqSimPairs {
			targetRatio := compfit.targetRatios.Get(simType)
			componentValue := thisDetailWeight * targetRatio
			if !simType.IsHighGood() {
				componentValue *= -1
			}
			sumIndividual += componentValue
		}

		statWeightResult[statType] = sumIndividual
	}

	// rescale them so that strength is value 1.0
	// only try if positive result (should always be) mostly to guard divide by zero
	// but negative also would flip everything else, and be messy
	strengthResult := statWeightResult[stats.Stat_Strength]
	if strengthResult > 0 {
		for _, statType := range G_RequiredStats {
			statWeightResult[statType] /= strengthResult
		}
	}
	return statWeightResult
}
