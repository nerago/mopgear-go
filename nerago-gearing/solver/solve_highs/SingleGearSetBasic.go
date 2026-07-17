package solve_highs

import (
	gear_model "paladin_gearing_go/gear_model"
	"paladin_gearing_go/gear_model/requirements"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_maxSetItems    = 5 // fundamental in MoP gear sets
	c_setItemsCounts = c_maxSetItems + 1

	c_scaled_ratings = 10000000.0 // try to make highs happier
	// example rating      178237915
	//                     187513497
	c_ratings_low_range  = 10000000.0 / c_scaled_ratings
	c_ratings_high_range = 1000000000000.0 / c_scaled_ratings
)

func SingleGearSetMain(itemOptions *items.SolvableOptionsMap, gear_model *gear_model.SpecModel, printer *util.PrintRecorder) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior

	setup := setupGearSet(&build, gear_model, itemOptions, 1)

	solutionFuture := build.RunHighsFuture(nil)

	return util_async.FutureCancellable_MapValue(solutionFuture, func(result util_highs.LinearResult) (items.SolvableItemSet, bool) {
		solution := result.GetSolutionAndSaveLog(printer)

		printer.Printf("SOLUTION STATUS = %s\n", solution.Status.String())
		debugPrint(solution, setup.build, setup.allColumns, printer)

		if solution.HasSolution() {
			itemSet := setup.buildResultSet(solution)
			validateNewSet(itemSet, itemOptions, gear_model)
			checkSetRatingIsObjective(solution, &itemSet, gear_model)
			return itemSet, true
		} else {
			return items.SolvableItemSet{}, false
		}
	})
}

func setupGearSet(build *util_highs.LinearBuilder, model *gear_model.SpecModel, itemOptions *items.SolvableOptionsMap, scaleOutputRating float64) *singleGearSetBasic {
	setup := singleGearSetBasic{singleGearSetShared: singleGearSetShared{build: build}}

	setup.prepareRatingSum()
	setup.prepareActiveSetCombos(&model.SetBonus)
	setup.prepareUniqueEquipped(itemOptions)

	require := model.StatRequirements.(*requirements.StatRequirementsHitExpertise)
	for slot, item := range itemOptions.AllItemSlotSeq() {
		setup.addItem(slot, item, model, require)
	}
	setup.finishItemsCommon(itemOptions)
	setup.finishRequiredStats(require)
	setup.finishBaseRating()

	setup.addMainOutputVariable(scaleOutputRating)
	setup.multiplyRatingsByActiveSetCombo(&model.SetBonus, setup.baseRatingSumVar)
	setup.addSetNeededCounts(model.SetBonusRequired)

	return &setup
}

type singleGearSetBasic struct {
	singleGearSetShared

	hitValueRow     util_highs.ConstraintRow // constrains values for the hits of each item
	expertValueRow  util_highs.ConstraintRow // constrains values for the expertise of each item
	minimumValueRow util_highs.ConstraintRow // when an extra minimum is specified

	baseRatingSumRow util_highs.ConstraintRow // values for the ratings of each item
	baseRatingSumVar *columnInfo              // sum of values for the ratings of selected items
}

func (setup *singleGearSetBasic) prepareRatingSum() {
	entry := columnInfo{entryType: entry_sum_rating}

	// sum of individual selected item ratings
	// doesen't go directly into output rating
	entry.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.C_PlusInf, &entry)

	// main action of this variable: derive value to match rest of rest of row sum
	setup.baseRatingSumRow.Add(entry.columnIndex, -1)

	// save reference
	setup.baseRatingSumVar = &entry
	setup.allColumns = append(setup.allColumns, &entry)
}

func (setup *singleGearSetBasic) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, model *gear_model.SpecModel, require *requirements.StatRequirementsHitExpertise) util_highs.ColumnIndex {
	columnIndex := setup.addItemCommon(itemSlot, item, &model.SetBonus)

	// add rating via a summation condition
	// scale down ratings to keep numbers small for solver stability
	rating := float64(model.CalcRatingSolveItem(item)) / c_scaled_ratings
	setup.baseRatingSumRow.Add(columnIndex, rating)

	// specific hit/expertise values for hi/lo limits
	setup.hitValueRow.Add(columnIndex, float64(item.Total().Hit()))
	setup.expertValueRow.Add(columnIndex, float64(item.Total().Expertise()))

	// additional minimum value (e.g. haste)
	additionalMinimum := require.AdditionalMinimumRequirement
	if additionalMinimum != nil {
		setup.minimumValueRow.Add(columnIndex, item.Total().GetFloat(additionalMinimum.StatType))
	}

	return columnIndex
}

func (setup *singleGearSetBasic) finishBaseRating() {
	// constrain: matching sum to individual ratings
	setup.baseRatingSumRow.Debug = "baseRatingSumRow"
	setup.baseRatingSumRow.Build(setup.build, 0, 0)
}

func (setup *singleGearSetBasic) finishRequiredStats(require *requirements.StatRequirementsHitExpertise) {
	// constrain: total sum of hit/exp are within requested limits
	setup.hitValueRow.Debug = "hitValueRow"
	setup.hitValueRow.Build(setup.build, float64(require.HitMin()), float64(require.HitMax()))
	setup.expertValueRow.Debug = "expertValueRow"
	setup.expertValueRow.Build(setup.build, float64(require.ExpertMin()), float64(require.ExpertMax()))

	// constrain: additional minimum value if specified has required minimum
	additionalMinimum := require.AdditionalMinimumRequirement
	if additionalMinimum != nil {
		setup.minimumValueRow.Build(setup.build, float64(additionalMinimum.Value), util_highs.C_PlusInf)
	}
}
