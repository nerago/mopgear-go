package solve_highs

import (
	"iter"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const c_gearExtended2ScoreHigh = 10

// TODO set multipliers per sim would be better

// CALCULATION:
// itemColumns * statTotalRows -> statTotalColumns
// statTotalColumns * detailedWeights[sim][stat] -> simValueTotalColumns
// simValueTotalColumns -> combinedRatingVar
// combinedRatingVar * entry_combo_active(column) -> entry_combo_output_weighted(column)
// entry_combo_output_weighted(column) * combo.multiply -> mainOutputRow
// mainOutputRow -> mainOutputVar

func SingleGearSetExtended2Main(itemOptions *items.SolvableOptionsMap, model *SolverModel, printer *util.PrintRecorder) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior

	setup := makeGearSetExtended2(&build, model, itemOptions, 1)

	solutionFuture := build.RunHighsFuture(nil)

	return util_async.FutureCancellable_MapValue(solutionFuture, func(result util_highs.LinearResult) (items.SolvableItemSet, bool) {
		solution := result.GetSolution2AndSaveLog(printer)

		printer.Printf("SOLUTION STATUS = %s\n", solution.Status().String())
		debugPrint(solution, setup.build, setup.allColumns, printer)

		if solution.HasSolution() {
			itemSet := setup.buildResultSet(solution)
			checkSetRatingIsObjective(solution, &itemSet, model.CalcRatingSet, 1)
			return itemSet, true
		} else {
			return items.SolvableItemSet{}, false
		}
	})
}

func makeGearSetExtended2(build *util_highs.LinearBuilder, model *SolverModel, itemOptions *items.SolvableOptionsMap, scaleOutputRating float64) *singleGearSetExtended2 {
	setup := singleGearSetExtended2{
		singleGearSetShared: singleGearSetShared{build: build},
	}

	setup.prepareStats()
	setup.prepareRequire(&model.StatRequirements)
	setup.prepareActiveSetCombos(model)
	setup.prepareUniqueEquipped(itemOptions)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		setup.addItem(slot, item, &model.StatRequirements, model.SetBonusIndexForItem)
	}
	setup.finishItemsCommon(itemOptions)
	setup.finishStats(&model.StatRequirements)

	setup.calcSimValues(model.Weights2)
	setup.calcCombinedSimRating(model.Weights2)
	setup.addMainOutputVariable(scaleOutputRating)
	setup.multiplyRatingsByActiveSetCombo(setup.combinedRatingVar, c_gearExtended2ScoreHigh)
	setup.addSetNeededCounts(model.SetBonusRequiredCounts, model.SetBonusAvoidNextStep)

	return &setup
}

type singleGearSetExtended2 struct {
	singleGearSetShared

	//model *ExtendedModel

	requireRows          map[stats.StatType]*util_highs.ConstraintRow // constrains values for the hit/expertise/etc of each item
	statTotalRows        map[stats.StatType]*util_highs.ConstraintRow
	statTotalColumns     map[stats.StatType]*columnInfo
	simValueTotalColumns map[stats.SimType]*columnInfo
	combinedRatingVar    *columnInfo // sum of values for the ratings of selected items
}

func (setup *singleGearSetExtended2) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, require *stats.StatTypeMap[weight_types.StatRangeFloat], activeSet func(id items.ItemId) (int, bool)) util_highs.ColumnIndex {
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

func (setup *singleGearSetExtended2) prepareRequire(require *stats.StatTypeMap[weight_types.StatRangeFloat]) {
	setup.requireRows = make(map[stats.StatType]*util_highs.ConstraintRow, require.Size())
	for statType := range require.SeqKey() {
		setup.requireRows[statType] = &util_highs.ConstraintRow{Debug: "require " + statType.Name()}
	}
}

func (setup *singleGearSetExtended2) prepareStats() {
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

func (setup *singleGearSetExtended2) finishStats(require *stats.StatTypeMap[weight_types.StatRangeFloat]) {
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

// for extended stats planned calculation is:
// statA*weight1A + statB*weight1B + statC*weight1C = sim1
// statA*weight2A + statB*weight2B + statC*weight2C = sim2
// (sim1+offset)*scale1 = 0-1.0  (better is higher)
// SimPriorityEntry.Apply is: (subtotal + se.RangingOffset) * se.RangingScale * se.RatioScale
// statA*weight1A + statB*weight1B + statC*weight1C = simValue
// ((statA*weight1A + statB*weight1B + statC*weight1C)+offset) * scales = simValue
// (statA*weight1A + statB*weight1B + statC*weight1C)+offset = simValue/scales
// statA*weight1A + statB*weight1B + statC*weight1C = simValue/scales - offset
// statA*weight1A + statB*weight1B + statC*weight1C - simValue/scales = -offset
func (setup *singleGearSetExtended2) calcSimValues(weight2 *weight_types.Weight2Extended) {
	// calculate each sim value from stats
	setup.simValueTotalColumns = make(map[stats.SimType]*columnInfo)
	for simType, nestedWeights := range weight2.SeqBySimNestedPairs() {
		simEntry := weight2.GetSimPriority().GetOrPanic(simType)
		setup.calcSimValue(simType, nestedWeights, simEntry)
	}
}

func (setup *singleGearSetExtended2) calcSimValue(simType stats.SimType, nestedWeights iter.Seq2[stats.StatType, float64], simEntry weight_types.SimPriorityEntry) {
	simValueColumn := &columnInfo{entryType: entry_sim_value, simType: simType}
	simValueColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), simValueColumn)
	setup.simValueTotalColumns[simType] = simValueColumn
	setup.allColumns = append(setup.allColumns, simValueColumn)

	simValueFromStatRow := util_highs.ConstraintRow{}
	for statType, weightValue := range nestedWeights {
		statColumn := setup.statTotalColumns[statType]
		simValueFromStatRow.Add(statColumn.columnIndex, weightValue)
	}

	offset := -simEntry.RangingOffset
	valueScale := -1.0 / simEntry.RangingScale
	simValueFromStatRow.Add(simValueColumn.columnIndex, valueScale)
	simValueFromStatRow.Build(setup.build, offset, offset)
}

func (setup *singleGearSetExtended2) calcCombinedSimRating(weight *weight_types.Weight2Extended) {
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
