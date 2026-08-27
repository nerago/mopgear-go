package solve_highs

import (
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_highs"
)

const c_gearExtended3StatHigh = 100000
const c_gearExtended3ScoreHigh = 100

type singleGearSetExtended3 struct {
	singleGearSetExtended
}

func SingleGearSetExtended3Main(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int) (*util_async.FutureCancellableWithError[*items.SolvableItemSet], error) {
	build := new(util_highs.LinearBuilder)
	build.Solver = util_highs.Solver_MIP_Interior
	build.TimeLimitSeconds = timeout
	//build.AddOptionBool("presolve_rule_logging", true)

	se3 := makeGearSetExtended3(build)
	outputVar, err := se3.setup(model, itemOptions)
	if err != nil {
		return nil, err
	}

	build.ChangeColumnOutputWeight(outputVar.columnIndex, 1)

	return se3.runForFutureResult(itemOptions, model, printer), nil
}

func makeGearSetExtended3(build *util_highs.LinearBuilder) *singleGearSetExtended3 {
	return &singleGearSetExtended3{
		singleGearSetExtended: singleGearSetExtended{
			singleGearSetShared: singleGearSetShared{
				build:             build,
				ratingPreScale:    1,
				bonusComboHandler: gearBonusComboHandler{build: build},
			},
		},
	}
}

func (se3 *singleGearSetExtended3) setup(model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap) (*columnInfo, error) {
	se3.itemSetupCommon.prepare(model, itemOptions, se3.createItemColumn)
	se3.itemSetupEx.prepareStatTotals()
	se3.itemSetupEx.prepareRequireEx(&model.StatRequirements)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		columnIndex := se3.itemSetupCommon.addItemCommon(slot, item)
		err := se3.itemSetupEx.addItem(item, &model.StatRequirements, columnIndex)
		if err != nil {
			return nil, err
		}
	}

	if err := se3.itemSetupCommon.finishItemsEquipped(itemOptions, se3.build); err != nil {
		return nil, err
	}

	if err := se3.itemSetupEx.finishRequireEx(&model.StatRequirements, se3.build); err != nil {
		return nil, err
	}
	countSetItemsCol := se3.itemSetupCommon.finishSetCounts(se3.build)
	statTotalCols, err := se3.itemSetupEx.finishStatTotals(se3.build)
	if err != nil {
		return nil, err
	}

	weightCalc := gearWeight3Calc{build: se3.build}
	simValueTotalColumns, err := weightCalc.calc(statTotalCols, model.Weights3)
	if err != nil {
		return nil, err
	}

	// simValueTotalColumns * activeCombo -> simValueComboColumns -> mainOutputVar
	return se3.calcFromSimValueToOutput(simValueTotalColumns, countSetItemsCol, model, &model.Weights3.SimPriority, c_gearExtended3ScoreHigh)
}
