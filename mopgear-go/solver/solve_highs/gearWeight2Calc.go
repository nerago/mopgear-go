package solve_highs

import (
	"errors"
	"iter"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

type gearWeight2Calc struct {
	build *util_highs.LinearBuilder
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
func (calc2 *gearWeight2Calc) calc(statTotalColumns *stats.StatTypeMap[*columnInfo], weight2 *weight_types.Weight2Extended) (map[stats.SimType]*columnInfo, error) {
	// calculate each sim value from stats
	simValueTotalColumns := make(map[stats.SimType]*columnInfo)
	for simType, nestedWeights := range weight2.SeqBySimNestedPairs() {
		if simEntry, hasEntry := weight2.GetSimPriority().Get(simType); hasEntry {
			simValueTotalColumns[simType] = calc2.calcSim(simType, nestedWeights, simEntry, statTotalColumns)
		} else {
			return nil, errors.New("missing priority for " + simType.Name())
		}
	}
	return simValueTotalColumns, nil
}

func (calc2 *gearWeight2Calc) calcSim(simType stats.SimType, nestedWeights iter.Seq2[stats.StatType, float64], simEntry weight_types.SimPriorityEntry, statTotalColumns *stats.StatTypeMap[*columnInfo]) (*columnInfo, error) {
	simValueColumn := &columnInfo{entryType: entry_sim_value, simType: simType}
	simValueColumn.columnIndex = calc2.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), simValueColumn)

	simValueFromStatRow := util_highs.ConstraintRow{Debug: "simValueFromStat " + simType.Name()}
	for statType, weightValue := range nestedWeights {
		if statColumn, hasColumn := statTotalColumns.Get(statType); hasColumn {
			simValueFromStatRow.Add(statColumn.columnIndex, weightValue)
		} else {
			return nil, errors.New("missing statTotalColumn for " + statType.Name())
		}
	}

	offset := -simEntry.RangingOffset
	valueScale := -1.0 / simEntry.RangingScale
	simValueFromStatRow.Add(simValueColumn.columnIndex, valueScale)
	simValueFromStatRow.Build(calc2.build, offset, offset)

	return simValueColumn, nil
}
