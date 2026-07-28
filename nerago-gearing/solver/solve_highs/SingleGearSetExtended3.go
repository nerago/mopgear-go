package solve_highs

import (
	"math"
	gear_model "paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const c_gearExtended3StatHigh = 100000
const c_gearExtended3ScoreHigh = 10

type StatRequiredExtended3 map[stats.StatType]util_collection.HiLoUInt32

type ExtendedModel3 struct {
	weight    weight_types.Weight3ExtendedRanged
	require   StatRequiredExtended3
	gearModel *gear_model.SpecModel
}

func SingleGearSetExtended3Main(itemOptions *items.SolvableOptionsMap, model *ExtendedModel3, printer *util.PrintRecorder) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior

	setup := setupGearSetExtended3(&build, model, itemOptions, 1)

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

func setupGearSetExtended3(build *util_highs.LinearBuilder, model *ExtendedModel3, itemOptions *items.SolvableOptionsMap, scaleOutputRating float64) *singleGearSetExtended3 {
	setup := singleGearSetExtended3{singleGearSetShared: singleGearSetShared{build: build}}

	setup.prepareStats()
	setup.prepareRequire(&model.require)
	setup.prepareActiveSetCombos(&model.gearModel.SetBonus)
	setup.prepareUniqueEquipped(itemOptions)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		setup.addItem(slot, item, &model.require, &model.gearModel.SetBonus)
	}
	setup.finishItemsCommon(itemOptions)
	setup.finishStats(&model.require)

	setup.calcSimValues()
	setup.calcCombinedSimRating()
	setup.addMainOutputVariable(scaleOutputRating)
	setup.multiplyRatingsByActiveSetCombo(&model.gearModel.SetBonus, setup.combinedRatingVar)
	setup.addSetNeededCounts(model.gearModel.SetBonusRequired)

	return &setup
}

// TODO set multipliers per sim would be better
// TODO make process to generate better multipliers

// CALCULATION:
// itemColumns * statTotalRows -> statTotalColumns
// statTotalColumns * detailedWeights -> simValueTotalColumns
// simValueTotalColumns * simRatioWeighting ->
// combinedRatingVar * entry_permutation_active(column) -> entry_permutation_output_weighted(column)
// entry_permutation_output_weighted(column) * permutation.weight -> mainOutputRow

type singleGearSetExtended3 struct {
	singleGearSetShared

	model *ExtendedModel3

	requireRows          map[stats.StatType]*util_highs.ConstraintRow // constrains values for the hit/expertise/etc of each item
	statTotalRows        map[stats.StatType]*util_highs.ConstraintRow
	statTotalColumns     map[stats.StatType]*columnInfo
	simValueTotalColumns map[stats.SimType]*columnInfo
	combinedRatingVar    *columnInfo // sum of values for the ratings of selected items
}

func (setup *singleGearSetExtended3) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, require *StatRequiredExtended3, setBonus *gear_model.SetBonus) util_highs.ColumnIndex {
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

func (setup *singleGearSetExtended3) prepareRequire(require *StatRequiredExtended3) {
	setup.requireRows = make(map[stats.StatType]*util_highs.ConstraintRow, len(*require))
	for statType := range *require {
		setup.requireRows[statType] = &util_highs.ConstraintRow{Debug: "require " + statType.Name()}
	}
}

func (setup *singleGearSetExtended3) prepareStats() {
	setup.statTotalRows = make(map[stats.StatType]*util_highs.ConstraintRow)
	setup.statTotalColumns = make(map[stats.StatType]*columnInfo)
	for _, statType := range stats.StatType_List {
		entry := columnInfo{entryType: entry_stat_total, statType: statType}
		entry.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), util_highs.DebugText("statTotal "+statType.Name()))
		setup.statTotalColumns[statType] = &entry
		setup.allColumns = append(setup.allColumns, &entry)

		setup.statTotalRows[statType] = &util_highs.ConstraintRow{Debug: "statTotal " + statType.Name()}
	}
}

func (setup *singleGearSetExtended3) finishStats(require *StatRequiredExtended3) {
	// constrain: total sum of hit/exp/etc are within requested limits
	for statType, hilo := range *require {
		row := setup.requireRows[statType]
		row.Build(setup.build, float64(hilo.Lo), convertHigh(hilo.Hi))
	}

	// constrain: total sum of each stat for input to weights
	for statType, column := range setup.statTotalColumns {
		row := setup.statTotalRows[statType]
		row.Add(column.columnIndex, -1)
		row.Build(setup.build, 0, 0)
	}
}

func (setup *singleGearSetExtended3) calcSimValues() {
	weight := setup.model.weight

	setup.simValueTotalColumns = make(map[stats.SimType]*columnInfo)
	for _, simType := range weight.SimList {
		setup.calcSimValueForType(simType, weight)
	}
}

func (setup *singleGearSetExtended3) calcSimValueForType(simType stats.SimType, weight weight_types.Weight3ExtendedRanged) {
	simValueColumn := setup.makeSimValueColumn(simType)
	simValueFromStatRow := util_highs.ConstraintRow{}

	for _, statType := range weight.StatList {
		statTotalColumn := setup.statTotalColumns[statType]
		contributeScoreSimStat := setup.makeContributeScoreSimStatColumn(simType, statType)

		for entry := range weight.StatWeights.GetAsSeq(simType, statType) {
			// essentially if stat total fits into this range then copy a score to contributeScoreSimStat
			if entry.StatRange.Maximum < math.MaxUint32 {
				setup.calcScoreFromEntryIfStatFits(entry, contributeScoreSimStat, simType, statType, statTotalColumn)
			} else {
				setup.calcScoreFromLastEntryIfGreater(entry, contributeScoreSimStat, simType, statType, statTotalColumn)
			}
		}

		simValueFromStatRow.Add(contributeScoreSimStat.columnIndex, 1)
	}

	simEntry := weight.SimPriority.GetOrPanic(simType)
	offset := -simEntry.RangingOffset
	valueScale := -1.0 / simEntry.RangingScale
	simValueFromStatRow.Add(simValueColumn.columnIndex, valueScale)
	simValueFromStatRow.Build(setup.build, offset, offset)
}

func (setup *singleGearSetExtended3) calcScoreFromEntryIfStatFits(entry weight_types.Weight3ExtendedStatEntry, contributeScoreSimStat columnInfo, simType stats.SimType, statType stats.StatType, statTotalColumn *columnInfo) {
	entryValue := setup.makeContributeScoreSimStatColumn(simType, statType) // need another entry type
	isBetween := setup.statIsBetween(statTotalColumn, entry.StatRange)
	setup.build.ConstraintIfBoolCopy(isBetween, entryValue.columnIndex, 1, contributeScoreSimStat.columnIndex, c_gearExtended3ScoreHigh)

	row := util_highs.ConstraintRow{}
	row.Add(statTotalColumn.columnIndex, entry.RatingWeight)
	row.Add(entryValue.columnIndex, -1)
	row.Build(setup.build, -entry.RatingOffset, -entry.RatingOffset)
}

func (setup *singleGearSetExtended3) calcScoreFromLastEntryIfGreater(entry weight_types.Weight3ExtendedStatEntry, contributeScoreSimStat columnInfo, simType stats.SimType, statType stats.StatType, statTotalColumn *columnInfo) {
	entryValue := setup.makeContributeScoreSimStatColumn(simType, statType) // need another entry type
	isOverMinimum := setup.build.ColumnIsGreaterOrEqualThanConstant(statTotalColumn.columnIndex, float64(entry.StatRange.Minimum), c_gearExtended3StatHigh, 1.0)
	setup.build.ConstraintIfBoolCopy(isOverMinimum, entryValue.columnIndex, 1, contributeScoreSimStat.columnIndex, c_gearExtended3ScoreHigh)

	row := util_highs.ConstraintRow{}
	row.Add(statTotalColumn.columnIndex, entry.RatingWeight)
	row.Add(entryValue.columnIndex, -1)
	row.Build(setup.build, -entry.RatingOffset, -entry.RatingOffset)
}

func (setup *singleGearSetExtended3) makeSimValueColumn(simType stats.SimType) columnInfo {
	simValueColumn := columnInfo{entryType: entry_sim_value, simType: simType}
	simValueColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &simValueColumn)
	setup.simValueTotalColumns[simType] = &simValueColumn
	setup.allColumns = append(setup.allColumns, &simValueColumn)
	return simValueColumn
}

func (setup *singleGearSetExtended3) makeContributeScoreSimStatColumn(simType stats.SimType, statType stats.StatType) columnInfo {
	simStatValueColumn := columnInfo{entryType: entry_sim_stat_value, simType: simType, statType: statType}
	simStatValueColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &simStatValueColumn)
	//setup.simValueTotalColumns[simType] = &simValueColumn
	setup.allColumns = append(setup.allColumns, &simStatValueColumn)
	return simStatValueColumn
}

func (setup *singleGearSetExtended3) calcCombinedSimRating() {
	// weighted sum of each sim value
	combinedRatingColumn := columnInfo{entryType: entry_sum_rating}
	combinedRatingColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), &combinedRatingColumn)
	setup.combinedRatingVar = &combinedRatingColumn
	setup.allColumns = append(setup.allColumns, &combinedRatingColumn)

	// add up the sim values, multiplying corresponding ratio
	combinedRatingRow := util_highs.ConstraintRow{}
	for simType, simValueColumn := range setup.simValueTotalColumns {
		simEntry := setup.model.weight.SimPriority.GetOrPanic(simType)
		combinedRatingRow.Add(simValueColumn.columnIndex, simEntry.RatioScale)
	}
	combinedRatingRow.Add(combinedRatingColumn.columnIndex, -1)
	combinedRatingRow.Build(setup.build, 0, 0)
}

func (setup *singleGearSetExtended3) statIsBetween(statTotalColumn *columnInfo, statRange weight_types.StatRange) util_highs.ColumnIndex {
	isOverMinimum := setup.build.ColumnIsGreaterOrEqualThanConstant(statTotalColumn.columnIndex, float64(statRange.Minimum), c_gearExtended3StatHigh, 1.0)
	isUnderMaximum := setup.build.ColumnIsLessOrEqualThanConstant(statTotalColumn.columnIndex, float64(statRange.Maximum), c_gearExtended3StatHigh, 1.0)
	isBetween := setup.build.CreateColumnBool(nil)
	and := util_highs.ConstraintAndBuilder{}
	and.AddInput(isOverMinimum)
	and.AddInput(isUnderMaximum)
	and.SetOutput(isBetween)
	and.Build(setup.build)
	return isBetween
}
