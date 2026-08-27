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
	c_single_basic_scaled_ratings = 1.0e-8 // try to make highs happier
	// example rating 138082172 1.38e8
	c_single_basic_ratings_high_range = 1.0e10 * c_single_basic_scaled_ratings
)

type singleGearSetBasic struct {
	singleGearSetShared
	itemSetupBasic gearItemSetupBasic
}

func SingleGearSetMain(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int) (*util_async.FutureCancellableWithError[*items.SolvableItemSet], error) {
	build := new(util_highs.LinearBuilder)
	build.Solver = util_highs.Solver_MIP_Interior
	build.TimeLimitSeconds = timeout

	sb := makeGearSetBasic(build)
	outputVar, err := sb.setup(model, itemOptions)
	if err != nil {
		return nil, err
	}
	build.ChangeColumnOutputWeight(outputVar.columnIndex, 1)

	return sb.runForFutureResult(itemOptions, model, printer), nil
}

func makeGearSetBasic(build *util_highs.LinearBuilder) *singleGearSetBasic {
	return &singleGearSetBasic{
		singleGearSetShared: singleGearSetShared{
			build:             build,
			ratingPreScale:    c_single_basic_scaled_ratings,
			bonusComboHandler: gearBonusComboHandler{build: build},
		},
	}
}

func (sb *singleGearSetBasic) setup(model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap) (*columnInfo, error) {
	sb.itemSetupCommon.prepare(model, itemOptions, sb.createItemColumn)
	sb.itemSetupBasic.prepareRequire(&model.StatRequirements)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		columnIndex := sb.itemSetupCommon.addItemCommon(slot, item)
		sb.itemSetupBasic.addItem(item, model.CalcRatingItem, columnIndex)
	}

	sb.itemSetupCommon.finishItemsEquipped(itemOptions, sb.build)
	sb.itemSetupBasic.finishRequire1(&model.StatRequirements, sb.build)
	countSetItemsCol := sb.itemSetupCommon.finishSetCounts(sb.build)
	baseRatingSumVar := sb.itemSetupBasic.finishRatingSum(sb.build)

	return sb.bonusComboHandler.processBonus(
		baseRatingSumVar,
		util_collection.Optional_Empty[stats.SimType](),
		c_single_basic_ratings_high_range,
		model,
		countSetItemsCol)
}
