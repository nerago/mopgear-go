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

func SingleGearSetExtended2Main(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior

	setup := makeGearSetExtended2(&build, model, itemOptions, 1)

	return setup.runForFutureResult(itemOptions, model, printer, 1)
}

func makeGearSetExtended2(build *util_highs.LinearBuilder, model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap, scaleOutputRating float64) *singleGearSetExtended2 {
	setup := singleGearSetExtended2{
		singleGearSetExtended: singleGearSetExtended{
			singleGearSetShared: singleGearSetShared{build: build},
		},
	}

	setup.prepareCommon(model, itemOptions, scaleOutputRating)
	setup.prepareStatTotals()
	setup.prepareRequireEx(&model.StatRequirements)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		setup.addItem(slot, item, &model.StatRequirements, model.SetBonusIndexForItem)
	}

	setup.finishItemsCommon(itemOptions)
	setup.finishRequireEx(&model.StatRequirements)
	setup.finishStatTotals()

	// statTotalColumns[statType] -> simValueTotalColumns[simType]
	setup.calcSimValues(model.Weights2)
	// simValueTotalColumns * activeCombo -> simValueComboColumns -> mainOutputVar
	setup.calcFromSimValueToOutput(model, &model.Weights2.SimPriority)

	return &setup
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
func (setup *singleGearSetExtended2) calcSimValues(weight2 *weight_types.Weight2Extended) {
	// calculate each sim value from stats
	setup.simValueTotalColumns = make(map[stats.SimType]*columnInfo)
	for simType, nestedWeights := range weight2.SeqBySimNestedPairs() {
		simEntry := weight2.GetSimPriority().GetOrPanic(simType)
		setup.calcSimValue(simType, nestedWeights, simEntry)
	}
}

func (setup *singleGearSetExtended2) calcSimValue(simType stats.SimType, nestedWeights iter.Seq2[stats.StatType, float64], simEntry weight_types.SimPriorityEntry) {
	simValueColumn := &columnInfo{entryType: entry_sim_value, simType: simType}
	simValueColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), simValueColumn)
	setup.simValueTotalColumns[simType] = simValueColumn
	setup.allColumns = append(setup.allColumns, simValueColumn)

	simValueFromStatRow := util_highs.ConstraintRow{}
	for statType, weightValue := range nestedWeights {
		statColumn := setup.statTotalColumns[statType]
		simValueFromStatRow.Add(statColumn.columnIndex, weightValue)
	}

	offset := -simEntry.RangingOffset
	valueScale := -1.0 / simEntry.RangingScale
	simValueFromStatRow.Add(simValueColumn.columnIndex, valueScale)
	simValueFromStatRow.Build(setup.build, offset, offset)
}
