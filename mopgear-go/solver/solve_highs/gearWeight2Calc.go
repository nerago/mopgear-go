package solve_highs

import (
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
func (se2 *gearWeight2Calc) calc(statTotalColumns *stats.StatTypeMap[*columnInfo], weight2 *weight_types.Weight2Extended) (simValueTotalColumns map[stats.SimType]*columnInfo) {
	// calculate each sim value from stats
	simValueTotalColumns = make(map[stats.SimType]*columnInfo)
	for simType, nestedWeights := range weight2.SeqBySimNestedPairs() {
		simEntry := weight2.GetSimPriority().GetOrPanic(simType)
		simValueTotalColumns[simType] = se2.calcSim(simType, nestedWeights, simEntry, statTotalColumns)
	}
	return simValueTotalColumns
}

func (se2 *gearWeight2Calc) calcSim(simType stats.SimType, nestedWeights iter.Seq2[stats.StatType, float64], simEntry weight_types.SimPriorityEntry, statTotalColumns *stats.StatTypeMap[*columnInfo]) (simValueColumn *columnInfo) {
	simValueColumn = &columnInfo{entryType: entry_sim_value, simType: simType}
	simValueColumn.columnIndex = se2.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), simValueColumn)

	simValueFromStatRow := util_highs.ConstraintRow{}
	for statType, weightValue := range nestedWeights {
		statColumn := statTotalColumns.GetOrPanic(statType)
		simValueFromStatRow.Add(statColumn.columnIndex, weightValue)
	}

	offset := -simEntry.RangingOffset
	valueScale := -1.0 / simEntry.RangingScale
	simValueFromStatRow.Add(simValueColumn.columnIndex, valueScale)
	simValueFromStatRow.Build(se2.build, offset, offset)

	return simValueColumn
}
