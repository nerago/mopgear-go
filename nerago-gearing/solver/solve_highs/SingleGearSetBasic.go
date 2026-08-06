package solve_highs

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_single_basic_scaled_ratings = 10000000.0 // try to make highs happier
	// example rating                           178237915
	//                                          187513497
	c_single_basic_ratings_high_range = 10000000000.0 / c_single_basic_scaled_ratings
)

func SingleGearSetMain(itemOptions *items.SolvableOptionsMap, model *SolverModel, printer *util.PrintRecorder) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior

	setup := makeGearSetBasic(&build, model, itemOptions, 1)

	solutionFuture := build.RunHighsFuture(nil)

	return util_async.FutureCancellable_MapValue(solutionFuture, func(result util_highs.LinearResult) (items.SolvableItemSet, bool) {
		solution := result.GetSolution2AndSaveLog(printer)

		printer.Printf("SOLUTION STATUS = %s\n", solution.Status().String())
		debugPrint(solution, setup.build, setup.allColumns, printer)

		if solution.HasSolution() {
			itemSet := setup.buildResultSet(solution)
			validateNewSet(itemSet, itemOptions, model.CheckSet)
			checkSetRatingIsObjective(solution, &itemSet, model.CalcRatingSet, c_single_basic_scaled_ratings)
			return itemSet, true
		} else {
			return items.SolvableItemSet{}, false
		}
	})
}

func makeGearSetBasic(build *util_highs.LinearBuilder, model *SolverModel, itemOptions *items.SolvableOptionsMap, scaleOutputRating float64) *singleGearSetBasic {
	setup := singleGearSetBasic{singleGearSetShared: singleGearSetShared{build: build}}

	setup.prepareRatingSum()
	setup.prepareActiveSetCombos(model)
	setup.prepareUniqueEquipped(itemOptions)
	setup.prepareRequiredStats(&model.StatRequirements)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		setup.addItem(slot, item, model.SetBonusIndexForItem, model.CalcRatingItem)
	}
	setup.finishItemsCommon(itemOptions)
	setup.finishRequiredStats(&model.StatRequirements)
	setup.finishBaseRating()

	setup.addMainOutputVariable(scaleOutputRating)
	setup.multiplyRatingsByActiveSetCombo(setup.baseRatingSumVar, c_single_basic_ratings_high_range)
	setup.addSetNeededCounts(model.SetBonusRequiredCounts)

	return &setup
}

type singleGearSetBasic struct {
	singleGearSetShared

	hitValueRow      util_highs.ConstraintRow // constrains values for the hits of each item
	expertValueRow   util_highs.ConstraintRow // constrains values for the expertise of each item
	minimumValueType stats.StatType
	minimumValueRow  util_highs.ConstraintRow // when an extra minimum is specified

	baseRatingSumRow util_highs.ConstraintRow // values for the ratings of each item
	baseRatingSumVar *columnInfo              // sum of values for the ratings of selected items
}

func (setup *singleGearSetBasic) prepareRatingSum() {
	entry := columnInfo{entryType: entry_sum_rating}

	// sum of individual selected item ratings
	// doesen't go directly into output rating
	entry.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), &entry)

	// main action of this variable: derive value to match rest of rest of row sum
	setup.baseRatingSumRow.Add(entry.columnIndex, -1)

	// save reference
	setup.baseRatingSumVar = &entry
	setup.allColumns = append(setup.allColumns, &entry)
}

func (setup *singleGearSetBasic) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, activeSet func(id items.ItemId) (int, bool), calcRating func(item *items.SolvableItem) float64) util_highs.ColumnIndex {
	columnIndex := setup.addItemCommon(itemSlot, item, activeSet)

	// add rating via a summation condition
	// scale down ratings to keep numbers small for solver stability
	rating := calcRating(item) / c_single_basic_scaled_ratings
	setup.baseRatingSumRow.Add(columnIndex, rating)

	// specific hit/expertise values for hi/lo limits
	setup.hitValueRow.Add(columnIndex, item.Total().GetFloat(stats.Stat_Hit))
	setup.expertValueRow.Add(columnIndex, item.Total().GetFloat(stats.Stat_Expertise))

	// additional minimum value (e.g. haste)
	if setup.minimumValueType != stats.Stat_Invalid {
		setup.minimumValueRow.Add(columnIndex, item.Total().GetFloat(setup.minimumValueType))
	}

	return columnIndex
}

func (setup *singleGearSetBasic) finishBaseRating() {
	// constrain: matching sum to individual ratings
	setup.baseRatingSumRow.Debug = "baseRatingSumRow"
	setup.baseRatingSumRow.Build(setup.build, 0, 0)
}

func (setup *singleGearSetBasic) finishRequiredStats(require *util_collection.EnumMap[stats.StatType, weight_types.StatRangeFloat]) {
	// constrain: total sum of hit/exp are within requested limits
	if hitRange, hasHit := require.Get(stats.Stat_Hit); hasHit {
		setup.hitValueRow.Debug = "hitValueRow"
		setup.hitValueRow.Build(setup.build, hitRange.Minimum, hitRange.Maximum)
	}

	if expertRange, hasExpert := require.Get(stats.Stat_Expertise); hasExpert {
		setup.expertValueRow.Debug = "expertValueRow"
		setup.expertValueRow.Build(setup.build, expertRange.Minimum, expertRange.Maximum)
	}

	// constrain: additional minimum value if specified has required minimum
	if setup.minimumValueType != stats.Stat_Invalid {
		otherRange := require.GetOrPanic(setup.minimumValueType)
		setup.minimumValueRow.Build(setup.build, otherRange.Minimum, otherRange.Maximum)
	}
}

func (setup *singleGearSetBasic) prepareRequiredStats(statRequirements *util_collection.EnumMap[stats.StatType, weight_types.StatRangeFloat]) {
	setup.minimumValueType = stats.Stat_Invalid
	for statType := range statRequirements.SeqKey() {
		if statType != stats.Stat_Hit && statType != stats.Stat_Expertise {
			if setup.minimumValueType == stats.Stat_Invalid {
				setup.minimumValueType = statType
			} else {
				panic("multiple additional required stats not supported in basic weights mode")
			}
		}
	}
}
