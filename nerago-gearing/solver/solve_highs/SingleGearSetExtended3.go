package solve_highs

import (
	"math"
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

// TODO set multipliers per sim would be better
// TODO make process to generate better multipliers

// CALCULATION:
// itemColumns * statTotalRows -> statTotalColumns
// statTotalColumns * rangeWeight -> simStatOptionColumn[]
// simStatOptionColumn[] -> simStatValueColumn
// simStatValueColumn -> simValueColumn
// simValueColumn * simRatioWeighting -> combinedRatingVar
// combinedRatingVar * entry_permutation_active(column) -> entry_permutation_output_weighted(column)
// entry_permutation_output_weighted(column) * permutation.weight -> mainOutputRow
// mainOutputRow -> mainOutputVar

func SingleGearSetExtended3Main(itemOptions *items.SolvableOptionsMap, model *SolverModel, printer *util.PrintRecorder) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior

	setup := makeGearSetExtended3(&build, model, itemOptions, 1)

	solutionFuture := build.RunHighsFuture(nil)

	return util_async.FutureCancellable_MapValue(solutionFuture, func(result util_highs.LinearResult) (items.SolvableItemSet, bool) {
		solution := result.GetSolution2AndSaveLog(printer)

		printer.Printf("SOLUTION STATUS = %s\n", solution.Status().String())
		debugPrint(solution, setup.build, setup.allColumns, printer)

		if solution.HasSolution() {
			itemSet := setup.buildResultSet(solution)
			checkSetRatingIsObjective(solution, &itemSet, model, 1)
			return itemSet, true
		} else {
			return items.SolvableItemSet{}, false
		}
	})
}

func makeGearSetExtended3(build *util_highs.LinearBuilder, model *SolverModel, itemOptions *items.SolvableOptionsMap, scaleOutputRating float64) *singleGearSetExtended3 {
	setup := singleGearSetExtended3{
		singleGearSetShared: singleGearSetShared{build: build},
	}

	setup.prepareStats()
	setup.prepareRequire(&model.StatRequirements)
	setup.prepareActiveSetCombos(model.SetBonusTotalCount)
	setup.prepareUniqueEquipped(itemOptions)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		setup.addItem(slot, item, &model.StatRequirements, model.SetBonusIndexForItem)
	}
	setup.finishItemsCommon(itemOptions)
	setup.finishStats(&model.StatRequirements)

	setup.calcSimValues(model.Weights3)
	setup.calcCombinedSimRating(model.Weights3)
	setup.addMainOutputVariable(scaleOutputRating)
	setup.multiplyRatingsByActiveSetCombo(setup.combinedRatingVar)
	setup.addSetNeededCounts(model.SetBonusRequiredCounts)

	return &setup
}

type singleGearSetExtended3 struct {
	singleGearSetShared

	requireRows          map[stats.StatType]*util_highs.ConstraintRow // constrains values for the hit/expertise/etc of each item
	statTotalRows        map[stats.StatType]*util_highs.ConstraintRow
	statTotalColumns     map[stats.StatType]*columnInfo
	simValueTotalColumns map[stats.SimType]*columnInfo
	combinedRatingVar    *columnInfo // sum of values for the ratings of selected items
}

func (setup *singleGearSetExtended3) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, require *util_collection.EnumMap[stats.StatType, weight_types.StatRangeFloat], activeSet func(id items.ItemId) (int, bool)) util_highs.ColumnIndex {
	columnIndex := setup.addItemCommon(itemSlot, item, activeSet)

	// add to stats via a summation condition
	for statType, value := range item.Total().SeqPairInt() {
		if value != 0 {
			setup.statTotalRows[statType].Add(columnIndex, float64(value))
		}
	}

	// specific hit/expertise/etc values for hi/lo limits
	for statType := range require.SeqKey() {
		setup.requireRows[statType].Add(columnIndex, item.Total().GetFloat(statType))
	}

	return columnIndex
}

func (setup *singleGearSetExtended3) prepareRequire(require *util_collection.EnumMap[stats.StatType, weight_types.StatRangeFloat]) {
	setup.requireRows = make(map[stats.StatType]*util_highs.ConstraintRow, require.Size())
	for statType := range require.SeqKey() {
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

func (setup *singleGearSetExtended3) finishStats(require *util_collection.EnumMap[stats.StatType, weight_types.StatRangeFloat]) {
	// constrain: total sum of hit/exp/etc are within requested limits
	for statType, hilo := range require.SeqKeyValue() {
		row := setup.requireRows[statType]
		row.Build(setup.build, hilo.Minimum, hilo.Maximum)
	}

	// constrain: total sum of each stat for input to weights
	for statType, column := range setup.statTotalColumns {
		row := setup.statTotalRows[statType]
		row.Add(column.columnIndex, -1)
		row.Build(setup.build, 0, 0)
	}
}

func (setup *singleGearSetExtended3) calcSimValues(weight3 *weight_types.Weight3ExtendedRanged) {
	setup.simValueTotalColumns = make(map[stats.SimType]*columnInfo)
	for _, simType := range weight3.SimList {
		setup.calcValueForSimType(simType, weight3)
	}
}

func (setup *singleGearSetExtended3) calcValueForSimType(simType stats.SimType, weight *weight_types.Weight3ExtendedRanged) {
	simValueColumn := setup.makeSimValueColumn(simType)
	simValueFromStatRow := util_highs.ConstraintRow{}

	for _, statType := range weight.StatList {
		chosenSimStatContribution := setup.calcValueForSimAndStatType(simType, statType, weight)
		simValueFromStatRow.Add(chosenSimStatContribution.columnIndex, 1)
	}

	simEntry := weight.SimPriority.GetOrPanic(simType)
	offset := -simEntry.RangingOffset
	valueScale := -1.0 / simEntry.RangingScale
	simValueFromStatRow.Add(simValueColumn.columnIndex, valueScale)
	simValueFromStatRow.Build(setup.build, offset, offset)
}

func (setup *singleGearSetExtended3) calcValueForSimAndStatType(simType stats.SimType, statType stats.StatType, weight *weight_types.Weight3ExtendedRanged) columnInfo {
	chosenSimStatContribution := setup.makeSimStatValueColumn(simType, statType)
	statTotalColumn := setup.statTotalColumns[statType]

	checkSingleRangeActive := util_highs.ConstraintRow{}
	for entry := range weight.StatWeights.GetAsSeq(simType, statType) {
		// check if stat total fits into this range
		var rangeCondition util_highs.ColumnIndex
		if entry.StatRange.Maximum < math.MaxUint32 {
			rangeCondition = setup.statIsBetween(statTotalColumn, entry.StatRange)
		} else {
			rangeCondition = setup.build.ColumnIsGreaterOrEqualThanConstant(statTotalColumn.columnIndex, float64(entry.StatRange.Minimum), c_gearExtended3StatHigh, 1.0)
		}
		checkSingleRangeActive.Add(rangeCondition, 1)

		// then copy a score to contributeScoreSimStat
		optionValueCol := setup.makeSimStatOptionColumn(simType, statType, entry.StatRange)
		setup.calcValueForEntry(statTotalColumn, entry, optionValueCol)
		setup.build.ConstraintIfBoolCopy(rangeCondition, optionValueCol.columnIndex, 1, chosenSimStatContribution.columnIndex, c_gearExtended3ScoreHigh)
	}
	checkSingleRangeActive.Build(setup.build, 1, 1)

	return chosenSimStatContribution
}

func (setup *singleGearSetExtended3) calcValueForEntry(statTotalColumn *columnInfo, entry weight_types.Weight3ExtendedStatEntry, entryValue columnInfo) {
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

func (setup *singleGearSetExtended3) makeSimStatValueColumn(simType stats.SimType, statType stats.StatType) columnInfo {
	simStatValueColumn := columnInfo{entryType: entry_sim_stat_value, simType: simType, statType: statType}
	simStatValueColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &simStatValueColumn)
	setup.allColumns = append(setup.allColumns, &simStatValueColumn)
	return simStatValueColumn
}

func (setup *singleGearSetExtended3) makeSimStatOptionColumn(simType stats.SimType, statType stats.StatType, statRange weight_types.StatRange) columnInfo {
	valueOptionColumn := columnInfo{entryType: entry_sim_stat_value_option, simType: simType, statType: statType, statRange: statRange}
	valueOptionColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &valueOptionColumn)
	setup.allColumns = append(setup.allColumns, &valueOptionColumn)
	return valueOptionColumn
}

func (setup *singleGearSetExtended3) calcCombinedSimRating(weight *weight_types.Weight3ExtendedRanged) {
	// weighted sum of each sim value
	combinedRatingColumn := columnInfo{entryType: entry_sum_rating}
	combinedRatingColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), &combinedRatingColumn)
	setup.combinedRatingVar = &combinedRatingColumn
	setup.allColumns = append(setup.allColumns, &combinedRatingColumn)

	// add up the sim values, multiplying corresponding ratio
	combinedRatingRow := util_highs.ConstraintRow{}
	for simType, simValueColumn := range setup.simValueTotalColumns {
		simEntry := weight.SimPriority.GetOrPanic(simType)
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
