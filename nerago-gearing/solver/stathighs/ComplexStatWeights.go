package stathighs

import (
	"paladin_gearing_go/simulate"
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

type ComplexStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimResultStats
	inputData    []WeightInput

	input *utilhighs.InputBuilder

	linearEquationDiff int
	linearInclude      int

	scaleSims             map[simulate.SimResultType]float64
	scaleStats            map[stats.StatType]float64
	detailedWeightColumns util.MapMap[stats.StatType, simulate.SimResultType, utilhighs.ColumnIndex]

	minimumIncludeRate float64
	includeColumns     []utilhighs.ColumnIndex
	includeCountRow    utilhighs.ConstraintRowBuild
}

func (comp *ComplexStatWeightProcess) Init(printer *util.PrintRecorder) {
	comp.printer = printer
}

func (comp *ComplexStatWeightProcess) SupplyData(inputData []WeightInput) {
	comp.inputData = inputData
}

func (comp *ComplexStatWeightProcess) SetTargetRatios(targetRatios simulate.SimResultStats) {
	comp.targetRatios = targetRatios
}

func (comp *ComplexStatWeightProcess) SetMinimumIncludeRate(percent float64) {
	comp.minimumIncludeRate = percent
}

func (comp *ComplexStatWeightProcess) Run() map[stats.StatType]float64 {
	comp.input = new(utilhighs.InputBuilder)
	comp.input.Minimise = true
	// comp.input.Solver = "ipm"

	// comp.linearEquationDiff = -1
	// comp.linearInclude = -1

	comp.input.BlendMultiObjectives = false
	comp.linearEquationDiff = comp.input.AddLinearPrioritised(false, -1, 0.5, 2)
	comp.linearInclude = comp.input.AddLinearPrioritised(false, -1, -1, 1)

	comp.chooseScaling()
	comp.createWeightColumns()
	comp.buildDataEquations()

	comp.includeCountRow.Finish(comp.input, float64(len(comp.inputData))*comp.minimumIncludeRate, utilhighs.C_PlusInf)

	solution, log := comp.input.RunHighs()
	comp.printer.AppendOther(log)

	return comp.exactAndReportSolution(solution)
}

func (comp *ComplexStatWeightProcess) chooseScaling() {
	comp.scaleSims = chooseSimScaling(comp.inputData, comp.printer)
	comp.scaleStats = chooseStatScaling(comp.inputData, comp.printer)
}

func (comp *ComplexStatWeightProcess) createWeightColumns() {
	minimumStrength := 0.0001

	for _, statType := range G_RequiredStats {
		for _, simType := range G_RequiredSims {
			lo := utilhighs.C_MinusInf
			hi := utilhighs.C_PlusInf
			colDetailWeight := comp.input.CreateColumnGeneral(highs.Continuous, lo, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name() + " " + simType.Name()})
			comp.detailedWeightColumns.Put(statType, simType, colDetailWeight)
		}
	}

	// we don't want to be dealing with 0 strength since that's our base stat to scale against
	// however could be rejecting some special situtations where it actually is true
	for _, colDetailWeight := range comp.detailedWeightColumns.SeqInnerWithKey1Value(stats.Stat_Strength) {
		comp.makeNotBetween(colDetailWeight, -minimumStrength, minimumStrength)
	}
}

// copied from FittingSingleStatWeightProcess.makeIsOverMinimum
func (comp *ComplexStatWeightProcess) makeNotBetween(checkColumn utilhighs.ColumnIndex, lo, hi float64) {
	isUnderMin := comp.input.CreateColumnBool(utilhighs.DebugString{Text: "isUnderMin"})
	isOverMax := comp.input.CreateColumnBool(utilhighs.DebugString{Text: "isOverMax"})

	// lo <= check + x*range <= range+lo
	// if undermin: lo <= check + range <= range+lo     ->>  lo <= check + range (for sure)
	//                                                  ->>  check <= lo
	// if !undermin: lo <= check + x*range <= range+lo  ->>  lo <= check
	//                                                       check <= range+lo (for sure)
	// if check < lo: lo <= check + x*range <= range+lo ->>  lo <= check + x*range        ->> lo - check <= x*range          ->>  small_positive <= x*range          ->>  x=1
	//                                                       check + x*range <= range+lo  ->> check - lo + x*range <= range  ->>  small_negative + x*range <= range  ->>  x=0 or 1
	// if check > lo: lo <= check + x*range <= range+lo ->>  lo <= check + x*range        ->> lo - check <= x*range          ->>  small_negative <= x*range          ->>  x=0 or 1
	//                                                       check + x*range <= range+lo  ->> check - lo + x*range <= range  ->>  small_positive + x*range <= range  ->>  x=0
	underMin := utilhighs.ConstraintRowBuild{Debug: "setIfOverMin"}
	underMin.Add(checkColumn, 1)
	underMin.Add(isUnderMin, c_complexHighWeight)
	underMin.Finish(comp.input, lo, lo+c_complexHighWeight)

	// hi-range <= check - x*range <= hi
	// if overmax: hi-range <= check - x*range <= hi  ->>  hi-range <= check - range <= hi  ->>  hi-range <= check - range  ->>  hi <= check
	//                                                                                      ->>  check - range <= hi        ->>  whatever
	// if !overmax: hi-range <= check - x*range <= hi  ->>  hi-range <= check <= hi  ->>  hi-range <= check  ->>  whatever
	//                                                                               ->>  check <= hi
	// if check < hi: hi-range <= check - x*range <= hi  ->>  hi-range <= check - x*range  ->>  hi-check <= range - x*range ->>  small_positive <= range - x*range  ->>  x=0
	//                                                        check - x*range <= hi        ->>  check - hi <= x*range       ->>  small_negative <= x*range          ->>  x=0 or 1
	// if check > hi: hi-range <= check - x*range <= hi  ->>  hi-range <= check - x*range  ->>  hi-check <= range - x*range ->>  small_negative <= range - x*range  ->>  x=0 or 1
	//                                                        check - x*range <= hi        ->>  check - hi <= x*range       ->>  small_positive <= x*range          ->>  x=1
	overMax := utilhighs.ConstraintRowBuild{Debug: "setIfUnderMax"}
	overMax.Add(checkColumn, 1)
	overMax.Add(isOverMax, -c_complexHighWeight)
	overMax.Finish(comp.input, hi-c_complexHighWeight, hi)

	or := utilhighs.ConstraintRowBuild{}
	or.Add(isUnderMin, 1)
	or.Add(isOverMax, 1)
	or.Finish(comp.input, 1, 1)
}

func (comp *ComplexStatWeightProcess) buildDataEquations() {
	for data := range util.ForPointer(comp.inputData) {
		comp.buildDataEquationForInput(data)
	}
}

func (comp *ComplexStatWeightProcess) buildDataEquationForInput(data *WeightInput) {
	includeColumn := comp.sampleIncludeToggleColumn()
	for _, simType := range G_RequiredSims {
		comp.buildDataEquationForSim(&data.TotalStat, data.SimResult.Get(simType), simType, includeColumn)
	}
}

func (comp *ComplexStatWeightProcess) sampleIncludeToggleColumn() utilhighs.ColumnIndex {
	includeColumn := comp.input.CreateColumnForLinearObjective(highs.Integer, 0, 1, c_complexOutputPerInclude, comp.linearInclude, utilhighs.DebugString{Text: "include"})
	comp.includeCountRow.Add(includeColumn, 1)
	comp.includeColumns = append(comp.includeColumns, includeColumn)
	return includeColumn
}

// equation is: weightA*scaledStatA + weightB*scaledStatB = scaledSimValue - diff
func (comp *ComplexStatWeightProcess) buildDataEquationForSim(stats *stats.StatBlock, simValue float64, simType simulate.SimResultType, includeColumn utilhighs.ColumnIndex) {
	matchSimValue := utilhighs.ConstraintRowBuild{}

	// TODO is there a way to flip the division for TMI DEATH etc, fundamental problem is that they don't increase linearly with stats

	for _, statType := range G_RequiredStats {
		weightDetailCol := comp.detailedWeightColumns.GetOrPanic(statType, simType)
		statValue := stats.GetFloat(statType)
		statScale := comp.scaleStats[statType]

		scaledStatValue := statValue * statScale
		matchSimValue.Add(weightDetailCol, scaledStatValue)
	}

	diffSigned := comp.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "diffSigned"})
	matchSimValue.Add(diffSigned, 1)

	diffOutput := comp.input.CreateColumnForLinearObjective(highs.Continuous, 0, c_complexHighDiff, 1, comp.linearEquationDiff, utilhighs.DebugString{Text: "diffOutput"})
	utilhighs.AbsoluteValue_WithToggle(comp.input, diffSigned, diffOutput, includeColumn, c_complexHighDiff)

	simScale := comp.scaleSims[simType]
	scaledSimValue := simValue * simScale
	matchSimValue.Finish(comp.input, scaledSimValue, scaledSimValue)
}

func (comp *ComplexStatWeightProcess) exactAndReportSolution(solution *highs.Solution) map[stats.StatType]float64 {
	comp.input.DebugPrintColumns(solution, comp.printer)

	comp.printer.Println("WEIGHTS")
	detailWeightMap := comp.extractDetailWeights(solution)
	statWeightResult := comp.computeFinalWeights(detailWeightMap)

	comp.reportExamples(detailWeightMap)
	comp.reportInclude(solution)

	return statWeightResult
}

func (comp *ComplexStatWeightProcess) extractDetailWeights(solution *highs.Solution) util.MapMap[stats.StatType, simulate.SimResultType, float64] {
	// extract and report on detail weights
	detailWeightMap := util.MapMap[stats.StatType, simulate.SimResultType, float64]{}
	for entry := range comp.detailedWeightColumns.SeqWithKeys() {
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

		scaleFix := comp.scaleSims[simType] / comp.scaleStats[statType]
		usableWeight := modelWeight / scaleFix

		if !simType.IsHighGood() {
			usableWeight *= -1
		}

		detailWeightMap.Put(statType, simType, usableWeight)

		comp.printer.Printf("%10s %10s %11.8f (%5.2e) %11.8f (%5.2e)\n", statType.Name(), simType.Name(), modelWeight, modelWeight, usableWeight, usableWeight)
	}
	comp.printer.Println0()

	for entry := range detailWeightMap.SeqWithKeysOtherOrder() {
		usableWeight := entry.Value
		comp.printer.Printf("%10s %10s %11.8f (%5.2e)\n", entry.Key1.Name(), entry.Key2.Name(), usableWeight, usableWeight)
	}
	comp.printer.Println0()
	return detailWeightMap
}

func (comp *ComplexStatWeightProcess) reportExamples(detailWeightMap util.MapMap[stats.StatType, simulate.SimResultType, float64]) {
	for i := range 20 {
		data := comp.inputData[i]
		comp.printer.Println("EXAMPLE")

		for _, simType := range G_RequiredSims {
			statSum := 0.0
			comp.printer.Printf(" %10s", simType.Name())
			for _, statType := range G_RequiredStats {
				statValue := data.TotalStat.GetFloat(statType)
				weight := detailWeightMap.GetOrPanic(statType, simType)
				if !simType.IsHighGood() {
					weight *= -1 // unflip so equation appears matches what the model saw
				}
				comp.printer.Printf(" {%s %.2f * %.4e = %.4f}", statType.Name(), statValue, weight, statValue*weight)
				statSum += statValue * weight
			}
			comp.printer.Printf(" = %.4f (expect %.4f)\n", statSum, data.SimResult.Get(simType))
		}

		comp.printer.Println0()
	}
}

func (comp *ComplexStatWeightProcess) computeFinalWeights(detailWeightMap util.MapMap[stats.StatType, simulate.SimResultType, float64]) map[stats.StatType]float64 {
	statWeightResult := make(map[stats.StatType]float64)
	for statType, seqSimPairs := range detailWeightMap.SeqGroupsKey1NestedKeyValue() {
		sumIndividual := 0.0

		for simType, thisDetailWeight := range seqSimPairs {
			strengthDetailWeight := detailWeightMap.GetOrPanic(stats.Stat_Strength, simType)
			targetRatio := comp.targetRatios.Get(simType)

			componentValue := targetRatio * thisDetailWeight / strengthDetailWeight

			sumIndividual += componentValue
		}

		statWeightResult[statType] = sumIndividual
	}

	return statWeightResult
}

func (comp *ComplexStatWeightProcess) reportInclude(solution *highs.Solution) {
	var includeCount uint32 = 0
	for _, col := range comp.includeColumns {
		if utilhighs.FloatEqualsOne(solution.ColValues[col]) {
			includeCount++
		}
	}
	includePercent := float64(includeCount) / float64(len(comp.inputData))
	comp.printer.Printf("Include %d %f\n", includeCount, includePercent)
}
