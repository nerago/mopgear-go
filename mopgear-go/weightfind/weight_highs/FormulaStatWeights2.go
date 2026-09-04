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
	c_formula2HighDiff         = 1000.0
	c_formula2OutputPerInclude = -0.1
	c_formula2ScaleTarget      = 1.0
)

type FormulaStatWeightProcess2 struct {
	printer *util.PrintRecorder

	targetRatios  weight_types.SimPriorityBasic
	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	inputData     []weight_types.WeightInput
	BLEND         int // only matters if minimumIncludeRate<1.0

	build *util_highs.LinearBuilder

	objectiveEquationDiff util_highs.ObjectiveIndex
	objectiveInclude      util_highs.ObjectiveIndex

	scaleSims             stats.SimTypeMap[util_weight.ScaleAndOffset]
	scaleStats            stats.StatTypeMap[float64]
	detailedWeightColumns util_collection.MapMap[stats.StatType, stats.SimType, util_highs.ColumnIndex]
	offsetColumns         map[stats.SimType]util_highs.ColumnIndex

	minimumIncludeRate float64
	includeColumns     []util_highs.ColumnIndex
	includeCountRow    util_highs.ConstraintRow
}

func (form *FormulaStatWeightProcess2) Init(printer *util.PrintRecorder) {
	form.printer = printer
}

func (form *FormulaStatWeightProcess2) SupplyData(inputData []weight_types.WeightInput) {
	form.inputData = inputData
}

func (form *FormulaStatWeightProcess2) SetRequiredStats(requiredStats []stats.StatType) {
	form.requiredStats = requiredStats
}

func (form *FormulaStatWeightProcess2) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	form.targetRatios = targetRatios
	form.requiredSims = targetRatios.SimTypes()
}

func (form *FormulaStatWeightProcess2) SetMinimumIncludeRate(percent float64) {
	form.minimumIncludeRate = percent
}

func (form *FormulaStatWeightProcess2) Run(timeout int) (*util_async.FutureCancellable[weight_types.WeightResult2], error) {
	form.build = new(util_highs.LinearBuilder)
	form.build.Minimise = true
	form.build.Solver = util_highs.Solver_MIP_Interior
	form.build.TimeLimitSeconds = timeout

	// comp.linearEquationDiff = -1
	// comp.linearInclude = -1

	if form.minimumIncludeRate == 1.0 {
		form.build.BlendMultiObjectives = true
		form.objectiveEquationDiff = form.build.AddObjectiveBlended(1, 0)
		form.objectiveInclude = form.build.AddObjectiveBlended(0, 0)
	} else if form.BLEND == 0 {
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
		if solution, err := linearResult.GetSolutionAndSaveLog(form.printer); err == nil {
			weight := form.extractAndReportSolution(solution)
			return weight_types.WeightResult2Make(&weight, stopwatch.Elapsed(), solution.Status)
		} else {
			return weight_types.WeightResult2MakeError(stopwatch.Elapsed(), err)
		}
	})
}

func (form *FormulaStatWeightProcess2) chooseScaling() {
	target := c_formula2ScaleTarget
	form.scaleStats = util_weight.ChooseStatScalingBasic(form.inputData, target, true, form.printer)
	form.scaleSims = util_weight.ChooseSimUnfriendlyUnitScaleAndOffset(form.inputData, form.requiredSims)
}

func (form *FormulaStatWeightProcess2) createWeightColumns() {
	for _, statType := range form.requiredStats {
		for _, simType := range form.requiredSims {
			colDetailWeight := form.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "WEIGHT " + statType.Name() + " " + simType.Name()})
			form.detailedWeightColumns.Put(statType, simType, colDetailWeight)
		}
	}

	form.offsetColumns = make(map[stats.SimType]util_highs.ColumnIndex)
	for _, simType := range form.requiredSims {
		form.offsetColumns[simType] = form.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "OFFSET " + simType.Name()})
	}
}

func (form *FormulaStatWeightProcess2) buildDataEquations() {
	for data := range util_collection.ForPointer(form.inputData) {
		form.buildDataEquationForInput(data)
	}
}

func (form *FormulaStatWeightProcess2) buildDataEquationForInput(data *weight_types.WeightInput) {
	includeColumn := form.sampleIncludeToggleColumn()
	for _, simType := range form.requiredSims {
		form.buildDataEquationForSim(&data.TotalStat, data.SimResult.Get(simType), simType, includeColumn)
	}
}

func (form *FormulaStatWeightProcess2) sampleIncludeToggleColumn() util_highs.ColumnIndex {
	// TODO if include rate is 100% skip this and disable MIP
	includeColumn := form.build.CreateColumnBoolWithObjective(c_formula2OutputPerInclude, form.objectiveInclude, util_highs.DebugString{Text: "include"})
	form.includeCountRow.Add(includeColumn, 1)
	form.includeColumns = append(form.includeColumns, includeColumn)
	return includeColumn
}

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

// another alternative is ignoring this complexity and going back closer to standard
// weights*stats = simValue
// that that only works for the positive model
// the better starting point is: death = 100% - weightA*scaledStatA - weightB*scaledStatB
// in more generic form:         simValue = sharedOffset + weightA*scaledStatA + weightB*scaledStatB  (allowing stats weights to go negative as much as they need)
// advantage is this form can be used for everything

// however this ends up with

func (form *FormulaStatWeightProcess2) buildDataEquationForSim(stats *stats.StatBlock, simValue float64, simType stats.SimType, includeColumn util_highs.ColumnIndex) {
	matchSimValue := util_highs.ConstraintRow{}

	for _, statType := range form.requiredStats {
		weightDetailCol := form.detailedWeightColumns.GetOrPanic(statType, simType)
		statValue := stats.GetFloat(statType)
		statScale := form.scaleStats.GetOrPanic(statType)

		scaledStatValue := statValue * statScale
		matchSimValue.Add(weightDetailCol, scaledStatValue)
	}

	deviationSigned := form.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugString{Text: "deviationSigned"})
	deviationAbsOutput := form.build.CreateColumnWithObjective(highs.Continuous, 0, c_formula2HighDiff, 1, form.objectiveEquationDiff, util_highs.DebugString{Text: "deviationAbsOutput"})

	matchSimValue.Add(deviationSigned, 1)
	form.build.AbsoluteValue_WithToggle(deviationSigned, deviationAbsOutput, includeColumn, c_formula2HighDiff)

	sharedOffsetCol := form.offsetColumns[simType]
	matchSimValue.Add(sharedOffsetCol, 1)

	simScale := form.scaleSims.GetOrPanic(simType)
	scaledSimValue := simScale.Apply(simValue)
	matchSimValue.Build(form.build, scaledSimValue, scaledSimValue)
}

func (form *FormulaStatWeightProcess2) extractAndReportSolution(solution *highs.Solution) weight_types.Weight2Extended {
	form.build.DebugPrintColumns(solution, form.printer)

	form.printer.Println("WEIGHTS")
	weightExtended := form.extractDetailWeights(solution)

	form.reportExamples(&weightExtended)
	form.reportInclude(solution)

	return weightExtended
}

func (form *FormulaStatWeightProcess2) extractDetailWeights(solution *highs.Solution) weight_types.Weight2Extended {
	// extract and report on detail weights
	weightExtended := weight_types.Weight2Extended_Make(form.requiredSims, form.requiredStats)
	for entry := range form.detailedWeightColumns.SeqKey1Key2ValueEntries() {
		statType := entry.Key1
		simType := entry.Key2
		column := entry.Value

		modelWeight := solution.ColValues[column]

		scaleStat := form.scaleStats.GetOrPanic(statType)
		usableWeight := modelWeight * scaleStat

		weightExtended.PutWeight(simType, statType, usableWeight)

		form.printer.Printf("%10s %10s %11.8f (%5.2e) %11.8f (%5.2e)\n", statType.Name(), simType.Name(), modelWeight, modelWeight, usableWeight, usableWeight)
	}
	form.printer.Println0()

	for simType, offsetColumn := range form.offsetColumns {
		//simScale := form.scaleSims[simType]
		offsetValue := solution.ColValues[offsetColumn]
		ratio := form.targetRatios.GetOrPanic(simType)
		weightExtended.SetSimScale(simType, 1, offsetValue, ratio)
	}

	for entry := range weightExtended.SeqBySimThenStat() {
		usableWeight := entry.Value
		form.printer.Printf("%10s %10s %11.8f (%5.2e)\n", entry.Key1.Name(), entry.Key2.Name(), usableWeight, usableWeight)
	}
	form.printer.Println0()

	weightExtended.UpdateScaling(form.inputData) // NOTE: process does get close to the mark, but was a bit off, like 0.11 - 0.98 rather than 0.0 - 1.0
	weightExtended.FinishAndValidate(form.inputData)
	return *weightExtended
}

func (form *FormulaStatWeightProcess2) reportExamples(weightExtended *weight_types.Weight2Extended) {
	for i := range min(20, len(form.inputData)) {
		data := form.inputData[i]
		form.printer.Println("EXAMPLE")

		for _, simType := range form.requiredSims {
			rowSum := 0.0
			form.printer.Printf(" %10s", simType.Name())
			for _, statType := range form.requiredStats {
				statValue := data.TotalStat.GetFloat(statType)
				weight := weightExtended.GetWeightOrPanic(simType, statType)
				form.printer.Printf(" {%s %.0f * %.2e = %.4f}", statType.Name(), statValue, weight, statValue*weight)
				rowSum += statValue * weight
			}

			priorityEntry := weightExtended.GetSimPriority().GetOrPanic(simType)
			offset := priorityEntry.RangingOffset
			simScale := form.scaleSims.GetOrPanic(simType)

			rowSum += offset
			simValue := data.SimResult.Get(simType)

			form.printer.Printf(" + {offset %.4f} = %.4f {expect %.4f + %.2e * %.2e = %.4f}\n", offset, rowSum, simValue, simScale.Offset, simScale.Scale, simScale.Apply(simValue))
		}

		form.printer.Println0()
	}
}

func (form *FormulaStatWeightProcess2) reportInclude(solution *highs.Solution) {
	var includeCount uint32 = 0
	for _, col := range form.includeColumns {
		if util.FloatEqualsOne(solution.ColValues[col]) {
			includeCount++
		}
	}
	includePercent := float64(includeCount) / float64(len(form.inputData))
	form.printer.Printf("Include %d %f\n", includeCount, includePercent)
}
