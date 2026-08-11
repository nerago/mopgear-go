package solve_highs

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

type singleGearSetExtended struct {
	singleGearSetShared

	requireRows          map[stats.StatType]*util_highs.ConstraintRow // constrains values for the hit/expertise/etc of each item
	statTotalRows        map[stats.StatType]*util_highs.ConstraintRow
	statTotalColumns     map[stats.StatType]*columnInfo
	simValueTotalColumns map[stats.SimType]*columnInfo
	simValueComboColumns map[stats.SimType]*columnInfo
	combinedRatingVar    *columnInfo // sum of values for the ratings of selected items
}

func (setup *singleGearSetExtended) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, require *stats.StatTypeMap[weight_types.StatRangeFloat], activeSet func(id items.ItemId) (int, bool)) util_highs.ColumnIndex {
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

func (setup *singleGearSetExtended) prepareRequireEx(require *stats.StatTypeMap[weight_types.StatRangeFloat]) {
	setup.requireRows = make(map[stats.StatType]*util_highs.ConstraintRow, require.Size())
	for statType := range require.SeqKey() {
		setup.requireRows[statType] = &util_highs.ConstraintRow{Debug: "require " + statType.Name()}
	}
}

func (setup *singleGearSetExtended) finishRequireEx(require *stats.StatTypeMap[weight_types.StatRangeFloat]) {
	// constrain: total sum of hit/exp/etc are within requested limits
	for statType, hilo := range require.SeqKeyValue() {
		row := setup.requireRows[statType]
		row.Build(setup.build, hilo.Minimum, hilo.Maximum)
	}
}

func (setup *singleGearSetExtended) prepareStatTotals() {
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

func (setup *singleGearSetExtended) finishStatTotals() {
	// constrain: total sum of each stat for input to weights
	for statType, column := range setup.statTotalColumns {
		row := setup.statTotalRows[statType]
		row.Add(column.columnIndex, -1)
		row.Build(setup.build, 0, 0)
	}
}

func (setup *singleGearSetExtended) calcFromSimValueToOutput(model *solve_highs_types.SolverModel, priority *weight_types.SimPriorityExtended) {
	if model.SetBonusExtendedUseSim {
		// simValueTotalColumns[simType] * activeCombo.simMultiplier -> simValueComboColumns[simType]
		setup.multiplySimValuesByCombo()
		// simValueComboColumns[simType] -> mainOutputVar
		setup.calcOutputValueFromValueComboCols(priority, setup.mainOutputVar)
	} else {
		// simValueTotalColumns[simType] * simPriority -> combinedRatingVar[single]
		setup.calcCombinedRatingFromValueTotalCols(priority)
		// combinedRatingVar * activeCombo.flatMultiplier -> mainOutputVar
		setup.multiplyByActiveCombo(setup.combinedRatingVar, setup.mainOutputVar, c_gearExtended2ScoreHigh,
			func(combo *bonusCombo) float64 { return combo.totalFlatMultiplier() },
		)
	}
}

func (setup *singleGearSetExtended) calcCombinedRatingFromValueTotalCols(priority *weight_types.SimPriorityExtended) {
	// weighted sum of each sim value
	combinedRatingColumn := columnInfo{entryType: entry_sum_rating}
	combinedRatingColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), &combinedRatingColumn)
	setup.combinedRatingVar = &combinedRatingColumn
	setup.allColumns = append(setup.allColumns, &combinedRatingColumn)

	// add up the sim values, multiplying corresponding ratio
	combinedRatingRow := util_highs.ConstraintRow{}
	for simType, simValueColumn := range setup.simValueTotalColumns {
		simEntry := priority.GetOrPanic(simType)
		combinedRatingRow.Add(simValueColumn.columnIndex, simEntry.RatioScale)
	}
	combinedRatingRow.Add(combinedRatingColumn.columnIndex, -1)
	combinedRatingRow.Build(setup.build, 0, 0)
}

func (setup *singleGearSetExtended) multiplySimValuesByCombo() {
	setup.simValueComboColumns = make(map[stats.SimType]*columnInfo)
	for simType, simValueTotal := range setup.simValueTotalColumns {
		simComboCol := setup.makeSimValueForComboVariable(simType)
		setup.multiplyByActiveCombo(simValueTotal, simComboCol, c_gearExtended2ScoreHigh,
			func(combo *bonusCombo) float64 { return combo.totalMultiplierForSim(simType) },
		)
	}
}

func (setup *singleGearSetExtended) makeSimValueForComboVariable(simType stats.SimType) *columnInfo {
	entry := columnInfo{entryType: entry_sim_value_combo, simType: simType}
	entry.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &entry)
	setup.allColumns = append(setup.allColumns, &entry)
	setup.simValueComboColumns[simType] = &entry
	return &entry
}

func (setup *singleGearSetExtended) calcOutputValueFromValueComboCols(priority *weight_types.SimPriorityExtended, outputVar *columnInfo) {
	sumRow := util_highs.ConstraintRow{}
	for simType, simComboCol := range setup.simValueComboColumns {
		simEntry := priority.GetOrPanic(simType)
		sumRow.Add(simComboCol.columnIndex, simEntry.RatioScale)
	}
	sumRow.Add(outputVar.columnIndex, -1)
	sumRow.Build(setup.build, 0, 0)
}
