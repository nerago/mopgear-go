package solve_highs

import (
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

type gearWeight3Calc struct {
	build *util_highs.LinearBuilder
}

func (calc3 *gearWeight3Calc) calc(statTotalColumns *stats.StatTypeMap[*columnInfo], weight3 *weight_types.Weight3ExtendedRanged) (simValueTotalColumns map[stats.SimType]*columnInfo) {
	simValueTotalColumns = make(map[stats.SimType]*columnInfo)
	for _, simType := range weight3.SimList {
		simValueTotalColumns[simType] = calc3.calcSim(simType, weight3, statTotalColumns)
	}
	return simValueTotalColumns
}

func (calc3 *gearWeight3Calc) calcSim(simType stats.SimType, weight *weight_types.Weight3ExtendedRanged, statTotalColumns *stats.StatTypeMap[*columnInfo]) (simValueColumn *columnInfo) {
	simValueColumn = calc3.makeSimValueColumn(simType)
	simValueFromStatRow := util_highs.ConstraintRow{}

	for _, statType := range weight.StatList {
		chosenSimStatContribution := calc3.calcSimAndStat(simType, statType, weight, statTotalColumns)
		simValueFromStatRow.Add(chosenSimStatContribution.columnIndex, 1)
	}

	simEntry := weight.SimPriority.GetOrPanic(simType)
	offset := -simEntry.RangingOffset
	valueScale := -1.0 / simEntry.RangingScale
	simValueFromStatRow.Add(simValueColumn.columnIndex, valueScale)
	simValueFromStatRow.Build(calc3.build, offset, offset)

	return simValueColumn
}

func (calc3 *gearWeight3Calc) calcSimAndStat(simType stats.SimType, statType stats.StatType, weight *weight_types.Weight3ExtendedRanged, statTotalColumns *stats.StatTypeMap[*columnInfo]) columnInfo {
	chosenSimStatContribution := calc3.makeSimStatValueColumn3(simType, statType)
	statTotalColumn := statTotalColumns.GetOrPanic(statType)

	checkSingleRangeActive := util_highs.ConstraintRow{}
	for entry := range weight.StatWeights.GetAsSeq(simType, statType) {
		rangeCondition := calc3.entryInRange(entry, statTotalColumn)
		checkSingleRangeActive.Add(rangeCondition, 1)

		optionValueCol := calc3.prepareStatOption(simType, statType, entry, statTotalColumn)

		// then copy a score to contributeScoreSimStat
		calc3.build.ConstraintCopyIfBool(rangeCondition, optionValueCol.columnIndex, 1, chosenSimStatContribution.columnIndex, c_gearExtended3ScoreHigh)
	}
	checkSingleRangeActive.Build(calc3.build, 1, 1)

	return chosenSimStatContribution
}

func (calc3 *gearWeight3Calc) prepareStatOption(simType stats.SimType, statType stats.StatType, entry weight_types.Weight3ExtendedStatEntry, statTotalColumn *columnInfo) columnInfo {
	optionValueCol := calc3.makeSimStatOptionColumn3(simType, statType, entry.StatRange)

	row := util_highs.ConstraintRow{}
	row.Add(statTotalColumn.columnIndex, entry.RatingWeight)
	row.Add(optionValueCol.columnIndex, -1)
	row.Build(calc3.build, -entry.RatingOffset, -entry.RatingOffset)

	return optionValueCol
}

func (calc3 *gearWeight3Calc) entryInRange(entry weight_types.Weight3ExtendedStatEntry, statTotalColumn *columnInfo) util_highs.ColumnIndex {
	// check if stat total fits into this range
	var rangeCondition util_highs.ColumnIndex
	if entry.StatRange.Maximum < math.MaxUint32 {
		rangeCondition = calc3.statIsBetween(statTotalColumn, entry.StatRange)
	} else {
		rangeCondition = calc3.build.ColumnIsGreaterOrEqualThanConstant(statTotalColumn.columnIndex, float64(entry.StatRange.Minimum), c_gearExtended3StatHigh, 1.0)
	}
	return rangeCondition
}

func (calc3 *gearWeight3Calc) makeSimValueColumn(simType stats.SimType) *columnInfo {
	simValueColumn := &columnInfo{entryType: entry_sim_value, simType: simType}
	simValueColumn.columnIndex = calc3.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), simValueColumn)
	return simValueColumn
}

func (calc3 *gearWeight3Calc) makeSimStatValueColumn3(simType stats.SimType, statType stats.StatType) columnInfo {
	simStatValueColumn := columnInfo{entryType: entry_sim_stat_value, simType: simType, statType: statType}
	simStatValueColumn.columnIndex = calc3.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &simStatValueColumn)
	return simStatValueColumn
}

func (calc3 *gearWeight3Calc) makeSimStatOptionColumn3(simType stats.SimType, statType stats.StatType, statRange weight_types.StatRange) columnInfo {
	valueOptionColumn := columnInfo{entryType: entry_sim_stat_value_option, simType: simType, statType: statType, statRange: statRange}
	valueOptionColumn.columnIndex = calc3.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &valueOptionColumn)
	return valueOptionColumn
}

func (calc3 *gearWeight3Calc) statIsBetween(statTotalColumn *columnInfo, statRange weight_types.StatRange) util_highs.ColumnIndex {
	minimum := float64(statRange.Minimum)
	maximum := float64(statRange.Maximum)
	return calc3.build.ColumnIsBetweenConstants(statTotalColumn.columnIndex, minimum, maximum, c_gearExtended3StatHigh, 1.0)
}
