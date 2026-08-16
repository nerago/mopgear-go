package solve_highs

import (
	"iter"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const c_gearExtended2ScoreHigh = 10

func SingleGearSetExtended2Main(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior
	build.TimeLimitSeconds = timeout

	se2 := makeGearSetExtended2(&build)
	outputVar := se2.createOutputVariableForSeparateRun()
	se2.setup(model, itemOptions, outputVar)

	return se2.runForFutureResult(itemOptions, model, printer)
}

func makeGearSetExtended2(build *util_highs.LinearBuilder) *singleGearSetExtended2 {
	return &singleGearSetExtended2{
		singleGearSetExtended: singleGearSetExtended{
			singleGearSetShared: singleGearSetShared{
				build:          build,
				ratingPreScale: 1,
			},
		},
	}
}

func (se2 *singleGearSetExtended2) setup(model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap, outputVar *columnInfo) {
	se2.itemSetupCommon.prepare(model, itemOptions, se2.createItemColumn)
	se2.itemSetupEx.prepareStatTotals()
	se2.itemSetupEx.prepareRequireEx(&model.StatRequirements)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		columnIndex := se2.itemSetupCommon.addItemCommon(slot, item)
		se2.itemSetupEx.addItem(item, &model.StatRequirements, columnIndex)
	}

	se2.itemSetupCommon.finishItemsEquipped(itemOptions, se2.build)
	se2.itemSetupEx.finishRequireEx(&model.StatRequirements, se2.build)
	countSetItemsCol := se2.itemSetupCommon.finishSetCounts(se2.build)
	statTotalCols := se2.itemSetupEx.finishStatTotals(se2.build)

	// statTotalColumns[statType] -> simValueTotalColumns[simType]
	se2.calcSimValues(model.Weights2)
	// simValueTotalColumns * activeCombo -> simValueComboColumns -> mainOutputVar
	se2.calcFromSimValueToOutput(model, &model.Weights2.SimPriority, outputVar)
}

type singleGearSetExtended2 struct {
	singleGearSetExtended
}

// for extended stats planned calculation is:
// statA*weight1A + statB*weight1B + statC*weight1C = sim1
// statA*weight2A + statB*weight2B + statC*weight2C = sim2
// (sim1+offset)*scale1 = 0-1.0  (better is higher)
// SimPriorityEntry.Apply is: (subtotal + se.RangingOffset) * se.RangingScale * se.RatioScale
// statA*weight1A + statB*weight1B + statC*weight1C = simValue
// ((statA*weight1A + statB*weight1B + statC*weight1C)+offset) * scales = simValue
// (statA*weight1A + statB*weight1B + statC*weight1C)+offset = simValue/scales
// statA*weight1A + statB*weight1B + statC*weight1C = simValue/scales - offset
// statA*weight1A + statB*weight1B + statC*weight1C - simValue/scales = -offset
func (se2 *singleGearSetExtended2) calcSimValues(weight2 *weight_types.Weight2Extended) {
	// calculate each sim value from stats
	se2.simValueTotalColumns = make(map[stats.SimType]*columnInfo)
	for simType, nestedWeights := range weight2.SeqBySimNestedPairs() {
		simEntry := weight2.GetSimPriority().GetOrPanic(simType)
		se2.calcSimValue(simType, nestedWeights, simEntry)
	}
}

func (se2 *singleGearSetExtended2) calcSimValue(simType stats.SimType, nestedWeights iter.Seq2[stats.StatType, float64], simEntry weight_types.SimPriorityEntry) {
	simValueColumn := &columnInfo{entryType: entry_sim_value, simType: simType}
	simValueColumn.columnIndex = se2.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), simValueColumn)
	se2.simValueTotalColumns[simType] = simValueColumn

	simValueFromStatRow := util_highs.ConstraintRow{}
	for statType, weightValue := range nestedWeights {
		statColumn := se2.statTotalColumns.GetOrPanic(statType)
		simValueFromStatRow.Add(statColumn.columnIndex, weightValue)
	}

	offset := -simEntry.RangingOffset
	valueScale := -1.0 / simEntry.RangingScale
	simValueFromStatRow.Add(simValueColumn.columnIndex, valueScale)
	simValueFromStatRow.Build(se2.build, offset, offset)
}
