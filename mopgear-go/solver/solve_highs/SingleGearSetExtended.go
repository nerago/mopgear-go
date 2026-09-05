package solve_highs

import (
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

type singleGearSetExtended struct {
	singleGearSetShared

	itemSetupEx gearItemSetupEx

	//simValueTotalColumns stats.SimTypeMap[*columnInfo]
	//simValueComboColumns stats.SimTypeMap[*columnInfo]
	//combinedRatingVar    *columnInfo // sum of values for the ratings of selected items
}

func (se *singleGearSetExtended) multiplySimValuesByCombo(simValueTotalColumns map[stats.SimType]*columnInfo, model *solve_highs_types.SolverModel, priority *weight_types.SimPriorityExtended, countSetItemsCol map[solve_highs_types.SetBonusIndex]*columnInfo, scoreHigh float64) (*columnInfo, error) {
	if len(simValueTotalColumns) > 1 {
		sumRow := util_highs.ConstraintRow{Debug: "multiplySimValuesByCombo"}

		for simType, simValueTotal := range simValueTotalColumns {
			simComboCol, err := se.bonusComboHandler.processBonus(
				simValueTotal,
				util_collection.Optional_OfValue(simType),
				scoreHigh,
				model,
				countSetItemsCol,
			)
			if err != nil {
				return nil, err
			}

			if simEntry, hasEntry := priority.Get(simType); hasEntry {
				sumRow.Add(simComboCol.columnIndex, simEntry.RatioScale)
			} else {
				return nil, util.ErrorTracedNew("missing ratio for " + simType.Name())
			}
		}

		outputVar := se.makeOutputVariable()
		sumRow.Add(outputVar.columnIndex, -1)
		sumRow.Build(se.build, 0, 0)
		return outputVar, nil
	} else if len(simValueTotalColumns) == 1 {
		simType, simValueTotal := util_collection.MapFirstEntry(simValueTotalColumns)
		simEntry, hasEntry := priority.Get(simType)
		if !hasEntry {
			return nil, util.ErrorTracedNew("missing priority for " + simType.Name())
		}

		simComboCol, err := se.bonusComboHandler.processBonus(
			simValueTotal,
			util_collection.Optional_OfValue(simType),
			scoreHigh,
			model,
			countSetItemsCol,
		)
		if err != nil {
			return nil, err
		}

		if simEntry.RatioScale == 1.0 {
			return simComboCol, nil
		} else {
			sumRow := util_highs.ConstraintRow{Debug: "multiplySimValuesByComboOne"}
			outputVar := se.makeOutputVariable()
			sumRow.Add(simComboCol.columnIndex, simEntry.RatioScale)
			sumRow.Add(outputVar.columnIndex, -1)
			sumRow.Build(se.build, 0, 0)
			return outputVar, nil
		}
	} else {
		return nil, util.ErrorTracedNew("empty sim types")
	}
}

func (se *singleGearSetExtended) makeOutputVariable() *columnInfo {
	entry := &columnInfo{entryType: entry_main_output}
	entry.columnIndex = se.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), entry)
	return entry
}
