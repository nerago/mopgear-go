package stathighs

import (
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_complexHighWeight       = 50.0
	c_complexHighDiff         = 1000.0
	c_complexOutputPerInclude = -0.1
)

type FormulaStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios stats.SimData
	requiredSims []stats.SimType
	inputData    []WeightInput

	build *utilhighs.LinearBuilder

	objectiveEquationDiff utilhighs.ObjectiveIndex
	objectiveInclude      utilhighs.ObjectiveIndex

	scaleSims             map[stats.SimType]float64
	scaleStats            map[stats.StatType]float64
	detailedWeightColumns util.MapMap[stats.StatType, stats.SimType, utilhighs.ColumnIndex]

	minimumIncludeRate float64
	includeColumns     []utilhighs.ColumnIndex
	includeCountRow    utilhighs.ConstraintRow
}

func (form *FormulaStatWeightProcess) Init(printer *util.PrintRecorder) {
	form.printer = printer
}

func (form *FormulaStatWeightProcess) SupplyData(inputData []WeightInput) {
	form.inputData = inputData
}

func (form *FormulaStatWeightProcess) SetTargetRatios(targetRatios stats.SimData) {
	form.targetRatios = targetRatios
	form.requiredSims = targetRatios.NonZeroTypes()
}

func (form *FormulaStatWeightProcess) SetMinimumIncludeRate(percent float64) {
	form.minimumIncludeRate = percent
}

func (form *FormulaStatWeightProcess) Run() WeightResult {
	form.build = new(utilhighs.LinearBuilder)
	form.build.Minimise = true
	form.build.Solver = utilhighs.Solver_MIP_Interior

	// comp.linearEquationDiff = -1
	// comp.linearInclude = -1

	form.build.BlendMultiObjectives = false
	form.objectiveEquationDiff = form.build.AddObjectivePrioritised(false, -1, 0.5, 2)
	form.objectiveInclude = form.build.AddObjectivePrioritised(false, -1, -1, 1)

	form.chooseScaling()
	form.createWeightColumns()
	form.buildDataEquations()

	form.includeCountRow.Build(form.build, float64(len(form.inputData))*form.minimumIncludeRate, utilhighs.C_PlusInf)

	solution, log := form.build.RunHighs()
	form.printer.AppendOther(log)

	return form.extractAndReportSolution(solution)
}

func (form *FormulaStatWeightProcess) chooseScaling() {
	form.scaleSims = chooseSimScaling(form.inputData, form.printer)
	form.scaleStats = chooseStatScaling(form.inputData, form.printer)
}

func (form *FormulaStatWeightProcess) createWeightColumns() {
	// minimumStrength := 0.0001

	for _, statType := range G_RequiredStats {
		for _, simType := range form.requiredSims {
			lo := utilhighs.C_MinusInf
			hi := utilhighs.C_PlusInf
			colDetailWeight := form.build.CreateColumnGeneral(highs.Continuous, lo, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name() + " " + simType.Name()})
			form.detailedWeightColumns.Put(statType, simType, colDetailWeight)
		}
	}

	// TODO do a positive sum for each?

	// we don't want to be dealing with 0 strength since that's our base stat to scale against
	// however could be rejecting some special situtations where it actually is true
	// for _, colDetailWeight := range comp.detailedWeightColumns.SeqInnerWithKey1Value(stats.Stat_Strength) {
	// 	comp.makeNotBetween(colDetailWeight, -minimumStrength, minimumStrength)
	// }
}

func (form *FormulaStatWeightProcess) makeNotBetween(checkColumn utilhighs.ColumnIndex, lo, hi float64) {
	form.build.ColumnIsNotBetweenConstantsVerify(checkColumn, lo, hi, c_complexHighWeight)
}

func (form *FormulaStatWeightProcess) buildDataEquations() {
	for data := range util.ForPointer(form.inputData) {
		form.buildDataEquationForInput(data)
	}
}

func (form *FormulaStatWeightProcess) buildDataEquationForInput(data *WeightInput) {
	includeColumn := form.sampleIncludeToggleColumn()
	for _, simType := range form.requiredSims {
		form.buildDataEquationForSim(&data.TotalStat, data.SimResult.Get(simType), simType, includeColumn)
	}
}

func (form *FormulaStatWeightProcess) sampleIncludeToggleColumn() utilhighs.ColumnIndex {
	includeColumn := form.build.CreateColumnBoolWithObjective(c_complexOutputPerInclude, form.objectiveInclude, utilhighs.DebugString{Text: "include"})
	form.includeCountRow.Add(includeColumn, 1)
	form.includeColumns = append(form.includeColumns, includeColumn)
	return includeColumn
}

// equation is: weightA*scaledStatA + weightB*scaledStatB = scaledSimValue - diff
func (form *FormulaStatWeightProcess) buildDataEquationForSim(stats *stats.StatBlock, simValue float64, simType stats.SimType, includeColumn utilhighs.ColumnIndex) {
	matchSimValue := utilhighs.ConstraintRow{}

	// TODO is there a way to flip the division for TMI DEATH etc, fundamental problem is that they don't increase linearly with stats

	for _, statType := range G_RequiredStats {
		weightDetailCol := form.detailedWeightColumns.GetOrPanic(statType, simType)
		statValue := stats.GetFloat(statType)
		statScale := form.scaleStats[statType]

		scaledStatValue := statValue * statScale
		matchSimValue.Add(weightDetailCol, scaledStatValue)
	}

	diffSigned := form.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "diffSigned"})
	matchSimValue.Add(diffSigned, 1)

	diffOutput := form.build.CreateColumnWithObjective(highs.Continuous, 0, c_complexHighDiff, 1, form.objectiveEquationDiff, utilhighs.DebugString{Text: "diffOutput"})
	form.build.AbsoluteValue_WithToggle(diffSigned, diffOutput, includeColumn, c_complexHighDiff)

	simScale := form.scaleSims[simType]
	scaledSimValue := simValue * simScale
	matchSimValue.Build(form.build, scaledSimValue, scaledSimValue)
}

func (form *FormulaStatWeightProcess) extractAndReportSolution(solution *highs.Solution) WeightResult {
	form.build.DebugPrintColumns(solution, form.printer)

	form.printer.Println("WEIGHTS")
	detailWeightMap := form.extractDetailWeights(solution)
	statWeightResult := form.computeFinalWeights(detailWeightMap)

	form.reportExamples(detailWeightMap)
	form.reportInclude(solution)

	return statWeightResult
}

func (form *FormulaStatWeightProcess) extractDetailWeights(solution *highs.Solution) util.MapMap[stats.StatType, stats.SimType, float64] {
	// extract and report on detail weights
	detailWeightMap := util.MapMap[stats.StatType, stats.SimType, float64]{}
	for entry := range form.detailedWeightColumns.SeqWithKeys() {
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

		scaleFix := form.scaleSims[simType] / form.scaleStats[statType]
		usableWeight := modelWeight / scaleFix

		if !simType.IsHighGood() {
			usableWeight *= -1
		}

		detailWeightMap.Put(statType, simType, usableWeight)

		form.printer.Printf("%10s %10s %11.8f (%5.2e) %11.8f (%5.2e)\n", statType.Name(), simType.Name(), modelWeight, modelWeight, usableWeight, usableWeight)
	}
	form.printer.Println0()

	for entry := range detailWeightMap.SeqWithKeysOtherOrder() {
		usableWeight := entry.Value
		form.printer.Printf("%10s %10s %11.8f (%5.2e)\n", entry.Key1.Name(), entry.Key2.Name(), usableWeight, usableWeight)
	}
	form.printer.Println0()
	return detailWeightMap
}

func (form *FormulaStatWeightProcess) reportExamples(detailWeightMap util.MapMap[stats.StatType, stats.SimType, float64]) {
	for i := range min(20, len(form.inputData)) {
		data := form.inputData[i]
		form.printer.Println("EXAMPLE")

		for _, simType := range form.requiredSims {
			statSum := 0.0
			form.printer.Printf(" %10s", simType.Name())
			for _, statType := range G_RequiredStats {
				statValue := data.TotalStat.GetFloat(statType)
				weight := detailWeightMap.GetOrPanic(statType, simType)
				if !simType.IsHighGood() {
					weight *= -1 // unflip so equation appears matches what the model saw
				}
				form.printer.Printf(" {%s %.2f * %.4e = %.4f}", statType.Name(), statValue, weight, statValue*weight)
				statSum += statValue * weight
			}
			form.printer.Printf(" = %.4f (expect %.4f)\n", statSum, data.SimResult.Get(simType))
		}

		form.printer.Println0()
	}
}

func (form *FormulaStatWeightProcess) computeFinalWeights(detailWeightMap util.MapMap[stats.StatType, stats.SimType, float64]) WeightResult {
	statWeightResult := WeightResult_Make()
	for statType, seqSimPairs := range detailWeightMap.SeqGroupsKey1NestedKeyValue() {
		sumIndividual := 0.0

		for simType, thisDetailWeight := range seqSimPairs {
			strengthDetailWeight := detailWeightMap.GetOrPanic(stats.Stat_Strength, simType)
			targetRatio := form.targetRatios.Get(simType)

			componentValue := targetRatio * thisDetailWeight / strengthDetailWeight

			sumIndividual += componentValue
		}

		statWeightResult.Put(statType, sumIndividual)
	}

	return statWeightResult
}

func (form *FormulaStatWeightProcess) reportInclude(solution *highs.Solution) {
	var includeCount uint32 = 0
	for _, col := range form.includeColumns {
		if utilhighs.FloatEqualsOne(solution.ColValues[col]) {
			includeCount++
		}
	}
	includePercent := float64(includeCount) / float64(len(form.inputData))
	form.printer.Printf("Include %d %f\n", includeCount, includePercent)
}
