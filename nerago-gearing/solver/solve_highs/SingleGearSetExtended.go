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

func (se *singleGearSetExtended) addItem(itemSlot items.SlotEquip, item *items.SolvableItem, require *stats.StatTypeMap[weight_types.StatRangeFloat], activeSet func(id items.ItemId) (int, bool)) util_highs.ColumnIndex {
	columnIndex := se.addItemCommon(itemSlot, item, activeSet)

	// add to stats via a summation condition
	for statType, value := range item.Total().SeqPairInt() {
		if value != 0 {
			se.statTotalRows[statType].Add(columnIndex, float64(value))
		}
	}

	// specific hit/expertise/etc values for hi/lo limits
	for statType := range require.SeqKey() {
		se.requireRows[statType].Add(columnIndex, item.Total().GetFloat(statType))
	}

	return columnIndex
}

func (se *singleGearSetExtended) prepareRequireEx(require *stats.StatTypeMap[weight_types.StatRangeFloat]) {
	se.requireRows = make(map[stats.StatType]*util_highs.ConstraintRow, require.Size())
	for statType := range require.SeqKey() {
		se.requireRows[statType] = &util_highs.ConstraintRow{Debug: "require " + statType.Name()}
	}
}

func (se *singleGearSetExtended) finishRequireEx(require *stats.StatTypeMap[weight_types.StatRangeFloat]) {
	// constrain: total sum of hit/exp/etc are within requested limits
	for statType, hilo := range require.SeqKeyValue() {
		row := se.requireRows[statType]
		row.Build(se.build, hilo.Minimum, hilo.Maximum)
	}
}

func (se *singleGearSetExtended) prepareStatTotals() {
	se.statTotalRows = make(map[stats.StatType]*util_highs.ConstraintRow)
	se.statTotalColumns = make(map[stats.StatType]*columnInfo)
	for _, statType := range stats.StatType_List {
		entry := columnInfo{entryType: entry_stat_total, statType: statType}
		entry.columnIndex = se.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), util_highs.DebugText("statTotal "+statType.Name()))
		se.statTotalColumns[statType] = &entry
		se.allColumns = append(se.allColumns, &entry)

		se.statTotalRows[statType] = &util_highs.ConstraintRow{Debug: "statTotal " + statType.Name()}
	}
}

func (se *singleGearSetExtended) finishStatTotals() {
	// constrain: total sum of each stat for input to weights
	for statType, column := range se.statTotalColumns {
		row := se.statTotalRows[statType]
		row.Add(column.columnIndex, -1)
		row.Build(se.build, 0, 0)
	}
}

func (se *singleGearSetExtended) calcFromSimValueToOutput(model *solve_highs_types.SolverModel, priority *weight_types.SimPriorityExtended, outputVar *columnInfo) {
	if model.SetBonus.ExtendedUseSim {
		// simValueTotalColumns[simType] * activeCombo.simMultiplier -> simValueComboColumns[simType]
		se.multiplySimValuesByCombo()
		// simValueComboColumns[simType] -> mainOutputVar
		se.calcOutputValueFromValueComboCols(priority, outputVar)
	} else {
		// simValueTotalColumns[simType] * simPriority -> combinedRatingVar[single]
		se.calcCombinedRatingFromValueTotalCols(priority)
		// combinedRatingVar * activeCombo.flatMultiplier -> mainOutputVar
		se.multiplyByActiveCombo(se.combinedRatingVar, outputVar, c_gearExtended2ScoreHigh,
			func(combo *bonusCombo) float64 { return combo.totalFlatMultiplier() },
		)
	}
}

func (se *singleGearSetExtended) calcCombinedRatingFromValueTotalCols(priority *weight_types.SimPriorityExtended) {
	// weighted sum of each sim value
	combinedRatingColumn := columnInfo{entryType: entry_sum_rating}
	combinedRatingColumn.columnIndex = se.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), &combinedRatingColumn)
	se.combinedRatingVar = &combinedRatingColumn
	se.allColumns = append(se.allColumns, &combinedRatingColumn)

	// add up the sim values, multiplying corresponding ratio
	combinedRatingRow := util_highs.ConstraintRow{}
	for simType, simValueColumn := range se.simValueTotalColumns {
		simEntry := priority.GetOrPanic(simType)
		combinedRatingRow.Add(simValueColumn.columnIndex, simEntry.RatioScale)
	}
	combinedRatingRow.Add(combinedRatingColumn.columnIndex, -1)
	combinedRatingRow.Build(se.build, 0, 0)
}

func (se *singleGearSetExtended) multiplySimValuesByCombo() {
	se.simValueComboColumns = make(map[stats.SimType]*columnInfo)
	for simType, simValueTotal := range se.simValueTotalColumns {
		simComboCol := se.makeSimValueForComboVariable(simType)
		se.multiplyByActiveCombo(simValueTotal, simComboCol, c_gearExtended2ScoreHigh,
			func(combo *bonusCombo) float64 { return combo.totalMultiplierForSim(simType) },
		)
	}
}

func (se *singleGearSetExtended) makeSimValueForComboVariable(simType stats.SimType) *columnInfo {
	entry := columnInfo{entryType: entry_sim_value_combo, simType: simType}
	entry.columnIndex = se.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &entry)
	se.allColumns = append(se.allColumns, &entry)
	se.simValueComboColumns[simType] = &entry
	return &entry
}

func (se *singleGearSetExtended) calcOutputValueFromValueComboCols(priority *weight_types.SimPriorityExtended, outputVar *columnInfo) {
	sumRow := util_highs.ConstraintRow{}
	for simType, simComboCol := range se.simValueComboColumns {
		simEntry := priority.GetOrPanic(simType)
		sumRow.Add(simComboCol.columnIndex, simEntry.RatioScale)
	}
	sumRow.Add(outputVar.columnIndex, -1)
	sumRow.Build(se.build, 0, 0)
}
