package solve_highs

import (
	gear_model "paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"

	"github.com/bartolsthoorn/gohighs/highs"
)

type StatRequiredExtended map[stats.StatType]util.HiLoInt

type ExtendedModel struct {
	weight    WeightExtended
	require   StatRequiredExtended
	gearModel *gear_model.SpecModel
}

func SingleGearSetExtendedMain(itemOptions *items.SolvableOptionsMap, model *ExtendedModel, printer *util.PrintRecorder) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior

	setup := setupGearSetExtended(&build, model, itemOptions, 1)

	solutionFuture := build.RunHighsFuture(nil)

	return util_async.FutureCancellable_MapValue(solutionFuture, func(result util_highs.LinearResult) (items.SolvableItemSet, bool) {
		solution := result.GetSolutionAndSaveLog(printer)

		printer.Printf("SOLUTION STATUS = %s\n", solution.Status.String())
		debugPrint(solution, setup.build, setup.allColumns, printer)

		if solution.HasSolution() {
			itemSet := setup.buildResultSet(solution)
			//checkSetRatingIsObjective(solution, &itemSet, model) // TODO extended version
			return itemSet, true
		} else {
			return items.SolvableItemSet{}, false
		}
	})
}

func setupGearSetExtended(build *util_highs.LinearBuilder, model *ExtendedModel, itemOptions *items.SolvableOptionsMap, scaleOutputRating float64) *singleGearSetExtended {
	setup := singleGearSetExtended{singleGearSetShared: singleGearSetShared{build: build}}

	setup.prepareStats()
	setup.prepareRequire(&model.require)
	setup.prepareActiveSetCombos(&model.gearModel.SetBonus)
	setup.prepareUniqueEquipped(itemOptions)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		setup.addItem(slot, item, &model.require, &model.gearModel.SetBonus)
	}
	setup.finishItemsCommon(itemOptions)
	setup.finishStats(&model.require)

	var weight *WeightExtended = &model.weight
	setup.calcSimValues(weight)
	setup.calcCombinedSimRating(weight)
	setup.addMainOutputVariable(scaleOutputRating)
	setup.multiplyRatingsByActiveSetCombo(&model.gearModel.SetBonus, setup.combinedRatingVar)
	setup.addSetNeededCounts(model.gearModel.SetBonusRequired)

	return &setup
}

// TODO set multipliers per sim would be better

// CALCULATION:
// itemColumns * statTotalRows -> statTotalColumns
// statTotalColumns * detailedWeights -> simValueTotalColumns
// simValueTotalColumns * simRatioWeighting ->
// combinedRatingVar * entry_permutation_active(column) -> entry_permutation_output_weighted(column)
// entry_permutation_output_weighted(column) * permutation.weight -> mainOutputRow

type singleGearSetExtended struct {
	singleGearSetShared

	requireRows          map[stats.StatType]*util_highs.ConstraintRow // constrains values for the hit/expertise/etc of each item
	statTotalRows        map[stats.StatType]*util_highs.ConstraintRow
	statTotalColumns     map[stats.StatType]*columnInfo
	simValueTotalColumns map[stats.SimType]*columnInfo
	combinedRatingVar    *columnInfo // sum of values for the ratings of selected items
}

func (setup *singleGearSetExtended) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, require *StatRequiredExtended, setBonus *gear_model.SetBonus) util_highs.ColumnIndex {
	columnIndex := setup.addItemCommon(itemSlot, item, setBonus)

	// add to stats via a summation condition
	for statType, value := range item.Total().SeqPairInt() {
		if value != 0 {
			setup.statTotalRows[statType].Add(columnIndex, float64(value))
		}
	}

	// specific hit/expertise/etc values for hi/lo limits
	for statType := range *require {
		setup.requireRows[statType].Add(columnIndex, item.Total().GetFloat(statType))
	}

	return columnIndex
}

func (setup *singleGearSetExtended) prepareRequire(require *StatRequiredExtended) {
	setup.requireRows = make(map[stats.StatType]*util_highs.ConstraintRow, len(*require))
	for statType := range *require {
		setup.requireRows[statType] = &util_highs.ConstraintRow{Debug: "require " + statType.Name()}
	}
}

func (setup *singleGearSetExtended) prepareStats() {
	setup.statTotalRows = make(map[stats.StatType]*util_highs.ConstraintRow)
	setup.statTotalColumns = make(map[stats.StatType]*columnInfo)
	for _, statType := range stats.StatType_List {
		entry := columnInfo{entryType: entry_stat_total, statType: statType}
		entry.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.C_PlusInf, util_highs.DebugText("statTotal "+statType.Name()))
		setup.statTotalColumns[statType] = &entry
		setup.allColumns = append(setup.allColumns, &entry)

		setup.statTotalRows[statType] = &util_highs.ConstraintRow{Debug: "statTotal " + statType.Name()}
	}
}

func (setup *singleGearSetExtended) finishStats(require *StatRequiredExtended) {
	// constrain: total sum of hit/exp/etc are within requested limits
	for statType, hilo := range *require {
		row := setup.requireRows[statType]
		row.Build(setup.build, float64(hilo.Lo), float64(hilo.Hi))
	}

	// constrain: total sum of each stat for input to weights
	for _, statType := range stats.StatType_List {
		column := setup.statTotalColumns[statType]
		row := setup.statTotalRows[statType]
		row.Add(column.columnIndex, -1)
		row.Build(setup.build, 0, 0)
	}
}

func (setup *singleGearSetExtended) calcSimValues(weight *WeightExtended) {
	// calculate each sim value from stats
	setup.simValueTotalColumns = make(map[stats.SimType]*columnInfo)
	for simType, nestedWeights := range weight.DetailedWeights.SeqGroupsKey2NestedKeyValue() {
		simValueColumn := columnInfo{entryType: entry_sim_value}
		simValueColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, &simValueColumn)
		setup.simValueTotalColumns[simType] = &simValueColumn
		setup.allColumns = append(setup.allColumns, &simValueColumn)

		simValueFromStatRow := util_highs.ConstraintRow{}
		for statType, weightValue := range nestedWeights {
			statColumn := setup.statTotalColumns[statType]
			simValueFromStatRow.Add(statColumn.columnIndex, weightValue)
		}
		simValueFromStatRow.Add(simValueColumn.columnIndex, -1)
		simValueFromStatRow.Build(setup.build, 0, 0)
	}
}

func (setup *singleGearSetExtended) calcCombinedSimRating(weight *WeightExtended) {
	// weighted sum of each sim value
	combinedRatingColumn := columnInfo{entryType: entry_sum_rating}
	combinedRatingColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.C_PlusInf, &combinedRatingColumn)
	setup.combinedRatingVar = &combinedRatingColumn
	setup.allColumns = append(setup.allColumns, &combinedRatingColumn)

	// add up the sim values, multiplying corresponding ratio
	combinedRatingRow := util_highs.ConstraintRow{}
	for simType, simValueColumn := range setup.simValueTotalColumns {
		simRatio := weight.SimRatioWeighting.Get(simType)
		combinedRatingRow.Add(simValueColumn.columnIndex, simRatio)
	}
	combinedRatingRow.Add(combinedRatingColumn.columnIndex, -1)
	combinedRatingRow.Build(setup.build, 0, 0)
}
