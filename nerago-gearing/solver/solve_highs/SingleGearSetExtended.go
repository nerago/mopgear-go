package solve_highs

import (
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_collection"
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

func (se *singleGearSetExtended) calcFromSimValueToOutput(simValueTotalColumns map[stats.SimType]*columnInfo, countSetItemsCol map[solve_highs_types.SetBonusIndex]*columnInfo, model *solve_highs_types.SolverModel, priority *weight_types.SimPriorityExtended) *columnInfo {
	// simValueTotalColumns[simType] * activeCombo.simMultiplier -> simValueComboColumns[simType] -> mainOutputVar
	return se.multiplySimValuesByCombo(simValueTotalColumns, model, priority, countSetItemsCol)
}

func (se *singleGearSetExtended) multiplySimValuesByCombo(simValueTotalColumns map[stats.SimType]*columnInfo, model *solve_highs_types.SolverModel, priority *weight_types.SimPriorityExtended, countSetItemsCol map[solve_highs_types.SetBonusIndex]*columnInfo) *columnInfo {
	if len(simValueTotalColumns) > 1 {
		sumRow := util_highs.ConstraintRow{}

		for simType, simValueTotal := range simValueTotalColumns {
			simComboCol := se.bonusComboHandler.ProcessBonus(
				simValueTotal,
				util_collection.Optional_OfValue(simType),
				c_gearExtended2ScoreHigh,
				model,
				countSetItemsCol,
			)

			simEntry := priority.GetOrPanic(simType)
			sumRow.Add(simComboCol.columnIndex, simEntry.RatioScale)
		}

		outputVar := se.makeOutputVariable()
		sumRow.Add(outputVar.columnIndex, -1)
		sumRow.Build(se.build, 0, 0)
		return outputVar
	} else if len(simValueTotalColumns) == 1 {
		simType, simValueTotal := util_collection.MapFirstEntry(simValueTotalColumns)
		simComboCol := se.bonusComboHandler.ProcessBonus(
			simValueTotal,
			util_collection.Optional_OfValue(simType),
			c_gearExtended2ScoreHigh,
			model,
			countSetItemsCol,
		)
		// TODO is it okay we ignore RatioScale
		return simComboCol
	} else {
		panic("empty sim types")
	}
}

func (se *singleGearSetExtended) makeOutputVariable() *columnInfo {
	entry := &columnInfo{entryType: entry_main_output}
	entry.columnIndex = se.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), entry)
	return entry
}
