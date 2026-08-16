package solve_highs

import (
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

type singleGearSetExtended struct {
	singleGearSetShared

	itemSetupEx gearItemSetupEx

	//simValueTotalColumns stats.SimTypeMap[*columnInfo]
	//simValueComboColumns stats.SimTypeMap[*columnInfo]
	//combinedRatingVar    *columnInfo // sum of values for the ratings of selected items
}

func (site *gearItemSetupEx) finishStatTotals(build *util_highs.LinearBuilder) (statTotalColumns *stats.StatTypeMap[*columnInfo]) {
	statTotalColumns = new(stats.StatTypeMap[*columnInfo])
	// constrain: total sum of each stat for input to weights
	for _, statType := range stats.StatType_List {
		entry := columnInfo{entryType: entry_stat_total, statType: statType}
		entry.columnIndex = build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), util_highs.DebugText("statTotal "+statType.Name()))
		statTotalColumns.Put(statType, &entry)

		row := site.statTotalRows.GetOrPanic(statType)
		row.Add(entry.columnIndex, -1)
	}
	return statTotalColumns
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
