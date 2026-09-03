package solve_highs

import (
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"
)

const (
	c_single_basic_default_scaled_ratings = 1.0e-9

	// when called from multi solve ratings will get multiplied to range 0.1-10, this has nice margin
	// less clear on single set calls, baseline highish example is 1.822136e+08, with our multiplier above that turns into 0.1822136, more than fine
	// but depends on weights which aren't scaled consistently
	c_single_basic_ratings_high_range = 1000
)

type singleGearSetBasic struct {
	singleGearSetShared
	itemSetupBasic gearItemSetupBasic
}

func SingleGearSetMain(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int) (*util_async.FutureCancellableWithError[items.SolvableItemSet], error) {
	build := new(util_highs.LinearBuilder)
	build.Solver = util_highs.Solver_MIP_Interior
	build.TimeLimitSeconds = timeout

	sb := makeGearSetBasic(build)
	outputVar, err := sb.setup(model, itemOptions, c_single_basic_default_scaled_ratings)
	if err != nil {
		return nil, err
	}
	build.ChangeColumnOutputWeight(outputVar.columnIndex, 1)

	return sb.runForFutureResult(itemOptions, model, printer, c_single_basic_default_scaled_ratings), nil
}

func makeGearSetBasic(build *util_highs.LinearBuilder) *singleGearSetBasic {
	return &singleGearSetBasic{
		singleGearSetShared: singleGearSetShared{
			build:             build,
			bonusComboHandler: gearBonusComboHandler{build: build},
		},
	}
}

func (sb *singleGearSetBasic) setup(model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap, ratingOutputScale float64) (*columnInfo, error) {
	if err := sb.itemSetupCommon.prepare(model, itemOptions, sb.createItemColumn); err != nil {
		return nil, err
	}
	if err := sb.itemSetupBasic.prepareRequire(&model.StatRequirements); err != nil {
		return nil, err
	}

	for slot, item := range itemOptions.AllItemSlotSeq() {
		columnIndex := sb.itemSetupCommon.addItemCommon(slot, item)
		sb.itemSetupBasic.addItem(item, model.CalcRatingItem, columnIndex, ratingOutputScale)
	}

	if err := sb.itemSetupCommon.finishItemsEquipped(itemOptions, sb.build); err != nil {
		return nil, err
	}
	if err := sb.itemSetupBasic.finishRequire1(&model.StatRequirements, sb.build); err != nil {
		return nil, err
	}
	countSetItemsCol := sb.itemSetupCommon.finishSetCounts(sb.build)
	baseRatingSumVar := sb.itemSetupBasic.finishRatingSum(sb.build)

	return sb.bonusComboHandler.processBonus(
		baseRatingSumVar,
		util_collection.Optional_Empty[stats.SimType](),
		c_single_basic_ratings_high_range,
		model,
		countSetItemsCol)
}
