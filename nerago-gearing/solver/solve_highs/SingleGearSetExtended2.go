package solve_highs

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
)

const c_gearExtended2ScoreHigh = 10

type singleGearSetExtended2 struct {
	singleGearSetExtended
}

func SingleGearSetExtended2Main(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior
	build.TimeLimitSeconds = timeout

	se2 := makeGearSetExtended2(&build)
	outputVar := se2.setup(model, itemOptions)
	build.ChangeColumnOutputWeight(outputVar.columnIndex, 1)

	return se2.runForFutureResult(itemOptions, model, printer)
}

func makeGearSetExtended2(build *util_highs.LinearBuilder) *singleGearSetExtended2 {
	return &singleGearSetExtended2{
		singleGearSetExtended: singleGearSetExtended{
			singleGearSetShared: singleGearSetShared{
				build:             build,
				ratingPreScale:    1,
				bonusComboHandler: gearBonusComboHandler{build: build},
			},
		},
	}
}

func (se2 *singleGearSetExtended2) setup(model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap) *columnInfo {
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

	weightCalc := gearWeight2Calc{build: se2.build}
	simValueTotalColumns := weightCalc.calc(statTotalCols, model.Weights2)

	// multiply combos
	return se2.calcFromSimValueToOutput(simValueTotalColumns, countSetItemsCol, model, &model.Weights2.SimPriority)
}
