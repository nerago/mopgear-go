package weight_highs

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	//c_formulaHighWeight       = 50.0
	c_formulaHighDiff         = 1000.0
	c_formulaOutputPerInclude = -0.1
)

type FormulaStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios  weight_types.SimPriorityBasic
	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	inputData     []weight_types.WeightInput
	BLEND         int

	build *util_highs.LinearBuilder

	objectiveEquationDiff util_highs.ObjectiveIndex
	objectiveInclude      util_highs.ObjectiveIndex

	scaleSims             stats.SimTypeMap[float64]
	scaleStats            stats.StatTypeMap[float64]
	detailedWeightColumns util_collection.MapMap[stats.StatType, stats.SimType, util_highs.ColumnIndex]

	minimumIncludeRate float64
	includeColumns     []util_highs.ColumnIndex
	includeCountRow    util_highs.ConstraintRow
}

func (form *FormulaStatWeightProcess) Init(printer *util.PrintRecorder) {
	form.printer = printer
}

func (form *FormulaStatWeightProcess) SupplyData(inputData []weight_types.WeightInput) {
	form.inputData = inputData
}

func (form *FormulaStatWeightProcess) SetRequiredStats(requiredStats []stats.StatType) {
	form.requiredStats = requiredStats
}

func (form *FormulaStatWeightProcess) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	form.targetRatios = targetRatios
	form.requiredSims = targetRatios.SimTypes()
}

func (form *FormulaStatWeightProcess) SetMinimumIncludeRate(percent float64) {
	form.minimumIncludeRate = percent
}

func (form *FormulaStatWeightProcess) Run(timeout int) (*util_async.FutureCancellable[weight_types.WeightResult2], error) {
	form.build = new(util_highs.LinearBuilder)
	form.build.Minimise = true
	form.build.Solver = util_highs.Solver_MIP_Interior
	form.build.TimeLimitSeconds = timeout

	// comp.linearEquationDiff = -1
	// comp.linearInclude = -1

	if form.BLEND == 0 {
		form.build.BlendMultiObjectives = false
		form.objectiveEquationDiff = form.build.AddObjectivePrioritised(false, -1, 0.5, 2)
		form.objectiveInclude = form.build.AddObjectivePrioritised(false, -1, -1, 1)
	} else if form.BLEND == 1 {
		form.build.BlendMultiObjectives = false
		form.objectiveEquationDiff = form.build.AddObjectivePrioritised(false, -1, 0.2, 2)
		form.objectiveInclude = form.build.AddObjectivePrioritised(false, -1, -1, 1)
	} else if form.BLEND == 2 {
		form.build.BlendMultiObjectives = false
		form.objectiveEquationDiff = form.build.AddObjectivePrioritised(false, -1, 0.01, 2)
		form.objectiveInclude = form.build.AddObjectivePrioritised(false, -1, -1, 1)
	} else if form.BLEND == 3 {
		form.build.BlendMultiObjectives = true
		form.objectiveEquationDiff = form.build.AddObjectiveBlended(1, 0)
		form.objectiveInclude = form.build.AddObjectiveBlended(1, 0)
	} else if form.BLEND == 4 {
		form.build.BlendMultiObjectives = true
		form.objectiveEquationDiff = form.build.AddObjectiveBlended(5, 0)
		form.objectiveInclude = form.build.AddObjectiveBlended(1, 0)
	} else {
		form.build.BlendMultiObjectives = true
		form.objectiveEquationDiff = form.build.AddObjectiveBlended(1, 0)
		form.objectiveInclude = form.build.AddObjectiveBlended(5, 0)
	}

	form.chooseScaling()
	form.createWeightColumns()
	form.buildDataEquations()

	form.includeCountRow.Build(form.build, float64(len(form.inputData))*form.minimumIncludeRate, util_highs.InfPos())

	stopwatch := util.StopwatchMakeStopped()
	solutionFuture := form.build.RunHighsFuture(stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) weight_types.WeightResult2 {
		solution, err := linearResult.GetSolutionAndSaveLog(form.printer)
		if err == nil {
			weight := form.extractAndReportSolution(solution)
			return weight_types.WeightResult2Make(&weight, stopwatch.Elapsed(), solution.Status)
		} else {
			return weight_types.WeightResult2MakeError(stopwatch.Elapsed(), err)
		}
	})
}

func (form *FormulaStatWeightProcess) chooseScaling() {
	target := 1.0 // TODO consider non-unit range
	form.scaleSims = util_weight.ChooseSimUnfriendlyScalingBasic(form.inputData, target, false, form.printer)
	form.scaleStats = util_weight.ChooseStatScalingBasic(form.inputData, target, false, form.printer)
}

func (form *FormulaStatWeightProcess) createWeightColumns() {
	for _, statType := range form.requiredStats {
		for _, simType := range form.requiredSims {
			lo := util_highs.InfNeg()
			hi := util_highs.InfPos()
			colDetailWeight := form.build.CreateColumnGeneral(highs.Continuous, lo, hi, util_highs.DebugString{Text: "WEIGHT " + statType.Name() + " " + simType.Name()})
			form.detailedWeightColumns.Put(statType, simType, colDetailWeight)
		}
	}

	// we don't want to be dealing with 0 strength since that's our base stat to scale against
	// in general for this algorithm we're aiming for a direct equation again simValue
	// so should be very rare unless it really seems like strength has zero contribution

	// could resurrect old approach:
	// for _, colDetailWeight := range comp.detailedWeightColumns.SeqInnerWithKey1Value(stats.Stat_Strength) {
	// 	comp.makeNotBetween(colDetailWeight, -minimumStrength, minimumStrength)
	// }
}

//func (form *FormulaStatWeightProcess) makeNotBetween(checkColumn utilhighs.ColumnIndex, lo, hi float64) {
//	form.build.ColumnIsNotBetweenConstantsVerify(checkColumn, lo, hi, c_complexHighWeight)
//}

func (form *FormulaStatWeightProcess) buildDataEquations() {
	for data := range util_collection.ForPointer(form.inputData) {
		form.buildDataEquationForInput(data)
	}
}

func (form *FormulaStatWeightProcess) buildDataEquationForInput(data *weight_types.WeightInput) {
	includeColumn := form.sampleIncludeToggleColumn()
	for _, simType := range form.requiredSims {
		form.buildDataEquationForSim(&data.TotalStat, data.SimResult.Get(simType), simType, includeColumn)
	}
}

func (form *FormulaStatWeightProcess) sampleIncludeToggleColumn() util_highs.ColumnIndex {
	includeColumn := form.build.CreateColumnBoolWithObjective(c_formulaOutputPerInclude, form.objectiveInclude, util_highs.DebugString{Text: "include"})
	form.includeCountRow.Add(includeColumn, 1)
	form.includeColumns = append(form.includeColumns, includeColumn)
	return includeColumn
}

// TODO is there a way to flip the division for TMI DEATH etc, fundamental problem is that they don't increase linearly with stats

// for normal dps/hps/tps etc.
// equation is: weightA*scaledStatA + weightB*scaledStatB = scaledSimValue - diff

// if we were to say they decrease linearly:
// death = 100% - weightA*scaledStatA - weightB*scaledStatB
// easy to change in solve, not entirely sure would end up with significantly different result or not, negative already sort of allowed
// but could easily go negative, would need to cap, might add MIP steps later on
// tmi/dtps are harder to scale

// but if they were more multiplicative:
// death = 100 * ratioA * ratioB
// solve for: how much ratioA change would lead to 10% less death
//         (100 * ratioA * ratioB) * 0.9
//       = 100 * (ratioA*0.9) * ratioB
//        ratioANew = ratioA*0.9
//                  = ratioA*(1/1.111...)
//        ratioA = 1 / (1+stat*weight)
// ideally the equation is then death = 100 * 1 / (1 + statA * weightA) * 1 / (1 + statB * weightB)
// can't do that in linear equations really

// if we solve prior 100 * 1 / (1 + statA * weightA) * 1 / (1 + statB * weightB)
//                 = 100 / ((1 + statA * weightA) * (1 + statB * weightB))
//                 = 100 / (1 + (statA * weightA) + (statB * weightB) + (statA * weightA) * (statB * weightB))
//                 = 100 / (1 + (statA * weightA) + (statB * weightB) + (statA * statB * weightA * weightB))
//                   looking even less linear

// not really equivalent but could have similar effect
// what about death = 100 * 1 / (1 + statA * weightA + statB * weightB)
//            death = 100 / (1 + statA * weightA + statB * weightB)
//      100 / death = (1 + statA * weightA + statB * weightB)
//                0 = (1 + statA * weightA + statB * weightB) - (100 / death)
//               -1 = statA * weightA + statB * weightB - 0.01 * death           [possible formula]
//                        or
//  100 / death - 1 = statA * weightA + statB * weightB                          [possible formula here since death const]

func (form *FormulaStatWeightProcess) buildDataEquationForSim(stats *stats.StatBlock, simValue float64, simType stats.SimType, includeColumn util_highs.ColumnIndex) {
	matchSimValue := util_highs.ConstraintRow{}

	for _, statType := range form.requiredStats {
		weightDetailCol := form.detailedWeightColumns.GetOrPanic(statType, simType)
		statValue := stats.GetFloat(statType)
		statScale := form.scaleStats.GetOrPanic(statType)

		scaledStatValue := statValue * statScale
		matchSimValue.Add(weightDetailCol, scaledStatValue)
	}

	diffSigned := form.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "diffSigned"})
	matchSimValue.Add(diffSigned, 1)

	diffOutput := form.build.CreateColumnWithObjective(highs.Continuous, 0, c_formulaHighDiff, 1, form.objectiveEquationDiff, util_highs.DebugString{Text: "diffOutput"})
	form.build.AbsoluteValue_WithToggle(diffSigned, diffOutput, includeColumn, c_formulaHighDiff)

	simScale := form.scaleSims.GetOrPanic(simType)
	scaledSimValue := simValue * simScale
	matchSimValue.Build(form.build, scaledSimValue, scaledSimValue)
}

func (form *FormulaStatWeightProcess) extractAndReportSolution(solution *highs.Solution) weight_types.Weight2 {
	form.build.DebugPrintColumns(solution, form.printer)

	form.printer.Println("WEIGHTS")
	weightExtended := form.extractDetailWeights(solution)

	form.reportExamples(&weightExtended)
	form.reportInclude(solution)

	return weightExtended
}

func (form *FormulaStatWeightProcess) extractDetailWeights(solution *highs.Solution) weight_types.Weight2 {
	// extract and report on detail weights
	weightExtended := weight_types.Weight2Extended_Make(form.requiredSims, form.requiredStats)
	for entry := range form.detailedWeightColumns.SeqKey1Key2ValueEntries() {
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

		scaleFix := form.scaleSims.GetOrPanic(simType) / form.scaleStats.GetOrPanic(statType)
		usableWeight := modelWeight / scaleFix

		weightExtended.PutWeight(simType, statType, usableWeight)

		form.printer.Printf("%10s %10s %11.8f (%5.2e) %11.8f (%5.2e)\n", statType.Name(), simType.Name(), modelWeight, modelWeight, usableWeight, usableWeight)
	}
	form.printer.Println0()

	for entry := range weightExtended.SeqBySimThenStat() {
		usableWeight := entry.Value
		form.printer.Printf("%10s %10s %11.8f (%5.2e)\n", entry.Key1.Name(), entry.Key2.Name(), usableWeight, usableWeight)
	}
	form.printer.Println0()

	for _, simType := range form.requiredSims {
		weightExtended.SetSimScale(simType, 1, 0, form.targetRatios.GetOrPanic(simType))
	}

	weightExtended.UpdateScaling(form.inputData)
	weightExtended.FinishAndValidate(form.inputData)
	return *weightExtended
}

func (form *FormulaStatWeightProcess) reportExamples(weightExtended *weight_types.Weight2) {
	for i := range min(20, len(form.inputData)) {
		data := form.inputData[i]
		form.printer.Println("EXAMPLE")

		for _, simType := range form.requiredSims {
			statSum := 0.0
			form.printer.Printf(" %10s", simType.Name())
			for _, statType := range form.requiredStats {
				statValue := data.TotalStat.GetFloat(statType)
				weight := weightExtended.GetWeightOrPanic(simType, statType)
				form.printer.Printf(" {%s %.2f * %.4e = %.4f}", statType.Name(), statValue, weight, statValue*weight)
				statSum += statValue * weight
			}
			form.printer.Printf(" = %.4f (expect %.4f)\n", statSum, data.SimResult.Get(simType))
		}

		form.printer.Println0()
	}
}

func (form *FormulaStatWeightProcess) reportInclude(solution *highs.Solution) {
	var includeCount uint32 = 0
	for _, col := range form.includeColumns {
		if util.FloatEqualsOne(solution.ColValues[col]) {
			includeCount++
		}
	}
	includePercent := float64(includeCount) / float64(len(form.inputData))
	form.printer.Printf("Include %d %f\n", includeCount, includePercent)
}
