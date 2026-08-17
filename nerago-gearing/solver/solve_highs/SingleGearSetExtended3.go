package solve_highs

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
)

const c_gearExtended3StatHigh = 100000
const c_gearExtended3ScoreHigh = 10

type singleGearSetExtended3 struct {
	singleGearSetExtended
}

func SingleGearSetExtended3Main(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior
	build.TimeLimitSeconds = timeout
	build.AddOptionBool("presolve_rule_logging", true)

	se3 := makeGearSetExtended3(&build)
	outputVar := se3.createOutputVariableForSeparateRun()
	se3.setup(model, itemOptions, outputVar)

	return se3.runForFutureResult(itemOptions, model, printer)
}

func makeGearSetExtended3(build *util_highs.LinearBuilder) *singleGearSetExtended3 {
	return &singleGearSetExtended3{
		singleGearSetExtended: singleGearSetExtended{
			singleGearSetShared: singleGearSetShared{
				build:             build,
				ratingPreScale:    1,
				bonusComboHandler: gearBonusComboHandler{build},
			},
		},
	}
}

func (se3 *singleGearSetExtended3) setup(model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap, outputVar *columnInfo) {
	se3.itemSetupCommon.prepare(model, itemOptions, nil)
	se3.itemSetupEx.prepareStatTotals()
	se3.itemSetupEx.prepareRequireEx(&model.StatRequirements)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		columnIndex := se3.itemSetupCommon.addItemCommon(slot, item)
		se3.itemSetupEx.addItem(item, &model.StatRequirements, columnIndex)
	}

	se3.itemSetupCommon.finishItemsEquipped(itemOptions, se3.build)
	se3.itemSetupEx.finishRequireEx(&model.StatRequirements, se3.build)
	countSetItemsCol := se3.itemSetupCommon.finishSetCounts(se3.build)
	statTotalCols := se3.itemSetupEx.finishStatTotals(se3.build)

	weightCalc := gearWeight3Calc{build: se3.build}
	simValueTotalColumns := weightCalc.calc(statTotalCols, model.Weights3)

	// simValueTotalColumns * activeCombo -> simValueComboColumns -> mainOutputVar
	se3.calcFromSimValueToOutput(model, &model.Weights3.SimPriority, outputVar)
}
