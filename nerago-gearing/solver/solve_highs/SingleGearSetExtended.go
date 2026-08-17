package solve_highs

import (
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

func (se *singleGearSetExtended) calcFromSimValueToOutput(simValueTotalColumns map[stats.SimType]*columnInfo, priority *weight_types.SimPriorityExtended, outputVar *columnInfo) {
	// simValueTotalColumns[simType] * activeCombo.simMultiplier -> simValueComboColumns[simType] -> mainOutputVar
	se.multiplySimValuesByCombo(simValueTotalColumns, priority, outputVar)
}

func (se *singleGearSetExtended) multiplySimValuesByCombo(simValueTotalColumns map[stats.SimType]*columnInfo, priority *weight_types.SimPriorityExtended, outputVar *columnInfo) {
	sumRow := util_highs.ConstraintRow{}

	for simType, simValueTotal := range simValueTotalColumns {
		simComboCol := se.makeSimValueForComboVariable(simType)
		se.bonusComboHandler.multiplyByActiveCombo(simValueTotal, simComboCol, c_gearExtended2ScoreHigh,
			func(combo *bonusCombo) float64 { return combo.totalMultiplierForSim(simType) },
		)

		simEntry := priority.GetOrPanic(simType)
		sumRow.Add(simComboCol.columnIndex, simEntry.RatioScale)
	}

	sumRow.Add(outputVar.columnIndex, -1)
	sumRow.Build(se.build, 0, 0)
}

func (se *singleGearSetExtended) makeSimValueForComboVariable(simType stats.SimType) *columnInfo {
	entry := columnInfo{entryType: entry_sim_value_combo, simType: simType}
	entry.columnIndex = se.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &entry)
	return &entry
}
