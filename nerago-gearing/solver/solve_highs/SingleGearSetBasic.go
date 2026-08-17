package solve_highs

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
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

func SingleGearSetMain(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior
	build.TimeLimitSeconds = timeout

	sb := makeGearSetBasic(&build)
	outputVar := sb.setup(model, itemOptions)
	build.ChangeColumnOutputWeight(outputVar.columnIndex, 1)
	return sb.runForFutureResult(itemOptions, model, printer)
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

func (sb *singleGearSetBasic) setup(model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap) *columnInfo {
	sb.itemSetupCommon.prepare(model, itemOptions, nil)
	sb.itemSetupBasic.prepareRequire(&model.StatRequirements)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		columnIndex := sb.itemSetupCommon.addItemCommon(slot, item)
		sb.itemSetupBasic.addItem(item, model.CalcRatingItem, columnIndex)
	}

	sb.itemSetupCommon.finishItemsEquipped(itemOptions, sb.build)
	sb.itemSetupBasic.finishRequire1(&model.StatRequirements, sb.build)
	countSetItemsCol := sb.itemSetupCommon.finishSetCounts(sb.build)
	baseRatingSumVar := sb.itemSetupBasic.finishRatingSum(sb.build)

	sb.bonusComboHandler.addSetNeededCounts(model.SetBonus.RequiredCounts, model.SetBonus.CountMode)

	return sb.bonusComboHandler.ProcessBonus(
		baseRatingSumVar,
		util_collection.Optional_Empty[stats.SimType](),
		c_single_basic_ratings_high_range,
		model,
		countSetItemsCol)
}
