package solve_highs

import (
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_highs"
)

const c_gearExtended2ScoreHigh = 100

type singleGearSetExtended2 struct {
	singleGearSetExtended
}

func SingleGearSetExtended2Main(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int) (*util_async.FutureCancellableWithError[items.SolvableItemSet], error) {
	build := new(util_highs.LinearBuilder)
	build.Solver = util_highs.Solver_MIP_Interior
	build.TimeLimitSeconds = timeout

	se2 := makeGearSetExtended2(build)
	outputVar, err := se2.setup(model, itemOptions, 1)
	if err != nil {
		return nil, err
	}
	build.ChangeColumnOutputWeight(outputVar.columnIndex, 1)

	return se2.runForFutureResult(itemOptions, model, printer, 1), nil
}

func makeGearSetExtended2(build *util_highs.LinearBuilder) *singleGearSetExtended2 {
	return &singleGearSetExtended2{
		singleGearSetExtended: singleGearSetExtended{
			singleGearSetShared: singleGearSetShared{
				build:             build,
				bonusComboHandler: gearBonusComboHandler{build: build},
			},
		},
	}
}

func (se2 *singleGearSetExtended2) setup(model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap, ratingOutputScale float64) (*columnInfo, error) {
	if err := se2.itemSetupCommon.prepare(model, itemOptions, se2.createItemColumn); err != nil {
		return nil, err
	}
	se2.itemSetupEx.prepareStatTotals()
	se2.itemSetupEx.prepareRequireEx(&model.StatRequirements)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		columnIndex := se2.itemSetupCommon.addItemCommon(slot, item)
		if err := se2.itemSetupEx.addItem(item, &model.StatRequirements, columnIndex); err != nil {
			return nil, err
		}
	}

	if err := se2.itemSetupCommon.finishItemsEquipped(itemOptions, se2.build); err != nil {
		return nil, err
	}
	if err := se2.itemSetupEx.finishRequireEx(&model.StatRequirements, se2.build); err != nil {
		return nil, err
	}
	countSetItemsCol := se2.itemSetupCommon.finishSetCounts(se2.build)

	statTotalCols, err := se2.itemSetupEx.finishStatTotals(se2.build)
	if err != nil {
		return nil, err
	}

	weightCalc := gearWeight2Calc{build: se2.build}
	simValueTotalColumns, err := weightCalc.calc(statTotalCols, model.Weights2)
	if err != nil {
		return nil, err
	}

	// multiply combos
	// simValueTotalColumns[simType] * activeCombo.simMultiplier -> simValueComboColumns[simType] -> mainOutputVar
	return se2.multiplySimValuesByCombo(simValueTotalColumns, model, &model.Weights2.SimPriority, countSetItemsCol, c_gearExtended2ScoreHigh, ratingOutputScale)
}
