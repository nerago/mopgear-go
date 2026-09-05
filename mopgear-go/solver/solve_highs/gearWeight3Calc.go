package solve_highs

import (
	"cmp"
	"slices"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

type gearWeight3Calc struct {
	build *util_highs.LinearBuilder
}

func (calc3 *gearWeight3Calc) calc(statTotalColumns *stats.StatTypeMap[*columnInfo], weight3 *weight_types.Weight3) (map[stats.SimType]*columnInfo, error) {
	simValueTotalColumns := make(map[stats.SimType]*columnInfo)
	for _, simType := range weight3.SimList {
		simValueCol, err := calc3.calcSim(simType, weight3, statTotalColumns)
		if err != nil {
			return nil, err
		}
		simValueTotalColumns[simType] = simValueCol
	}
	return simValueTotalColumns, nil
}

func (calc3 *gearWeight3Calc) calcSim(simType stats.SimType, weight *weight_types.Weight3, statTotalColumns *stats.StatTypeMap[*columnInfo]) (*columnInfo, error) {
	simValueColumn := calc3.makeSimValueColumn(simType)
	simValueFromStatRow := util_highs.ConstraintRow{}

	for _, statType := range weight.StatList {
		chosenSimStatContribution, err := calc3.calcSimAndStat(simType, statType, weight, statTotalColumns)
		if err != nil {
			return nil, err
		}
		simValueFromStatRow.Add(chosenSimStatContribution.columnIndex, 1)
	}

	if simEntry, hasEntry := weight.SimPriority.Get(simType); hasEntry {
		offset := -simEntry.Ranging.Offset
		valueScale := -1.0 / simEntry.Ranging.Scale
		simValueFromStatRow.Add(simValueColumn.columnIndex, valueScale)
		simValueFromStatRow.Build(calc3.build, offset, offset)
	} else {
		return nil, util.ErrorTracedNew("missing priority for " + simType.Name())
	}

	return simValueColumn, nil
}

func (calc3 *gearWeight3Calc) calcSimAndStat(simType stats.SimType, statType stats.StatType, weight *weight_types.Weight3, statTotalColumns *stats.StatTypeMap[*columnInfo]) (*columnInfo, error) {
	statTotalColumn, hasStatTotalColumn := statTotalColumns.Get(statType)
	if !hasStatTotalColumn {
		return nil, util.ErrorTracedNew("missing statTotalColumn for " + statType.Name())
	}

	entries := slices.Collect(weight.StatWeights.GetAsSeq(simType, statType))
	slices.SortFunc(entries, func(a, b weight_types.Weight3ExtendedStatEntry) int {
		return cmp.Compare(a.StatRange.Minimum, b.StatRange.Minimum)
	})

	if len(entries) == 0 {
		return nil, util.ErrorTracedNew("missing weight entry")
	} else if len(entries) == 1 {
		entry := entries[0]
		optionValueCol := calc3.prepareStatOption(simType, statType, entry, statTotalColumn)
		return optionValueCol, nil
	}

	chosenSimStatContribution := calc3.makeSimStatValueColumn3(simType, statType)
	checkSingleRangeActive := util_highs.ConstraintRow{}
	var previousRangeMaxColumn util_highs.ColumnIndex

	// first entry copy its version of calc if stat < entry.max
	{
		firstEntry := entries[0]

		rangeCondition := calc3.build.ColumnIsLessOrEqualThanConstant(statTotalColumn.columnIndex, float64(firstEntry.StatRange.Maximum), c_gearExtended3StatHigh, 1.0)
		checkSingleRangeActive.Add(rangeCondition, 1)

		optionValueCol := calc3.prepareStatOption(simType, statType, firstEntry, statTotalColumn)
		calc3.build.ConstraintCopyIfBool(rangeCondition, optionValueCol.columnIndex, 1, chosenSimStatContribution.columnIndex, c_gearExtended3ScoreHigh)

		previousRangeMaxColumn = rangeCondition
	}

	for index := 1; index < len(entries)-1; index++ {
		entry := entries[index]

		nextMax := calc3.build.ColumnIsLessOrEqualThanConstant(statTotalColumn.columnIndex, float64(entry.StatRange.Maximum), c_gearExtended3StatHigh, 1.0)
		rangeCondition := columnBoolNotAAndB(calc3.build, previousRangeMaxColumn, nextMax)
		checkSingleRangeActive.Add(rangeCondition, 1)

		optionValueCol := calc3.prepareStatOption(simType, statType, entry, statTotalColumn)
		calc3.build.ConstraintCopyIfBool(rangeCondition, optionValueCol.columnIndex, 1, chosenSimStatContribution.columnIndex, c_gearExtended3ScoreHigh)

		previousRangeMaxColumn = nextMax
	}

	// last entry copy calc if stat > previous entry's max
	{
		lastEntry := entries[len(entries)-1]

		rangeCondition := calc3.build.NotAsColumn(previousRangeMaxColumn)
		checkSingleRangeActive.Add(rangeCondition, 1)

		optionValueCol := calc3.prepareStatOption(simType, statType, lastEntry, statTotalColumn)
		calc3.build.ConstraintCopyIfBool(rangeCondition, optionValueCol.columnIndex, 1, chosenSimStatContribution.columnIndex, c_gearExtended3ScoreHigh)
	}
	//checkSingleRangeActive.Build(calc3.build, 1, 1) // TODO see if this helps

	return chosenSimStatContribution, nil
}

func (calc3 *gearWeight3Calc) prepareStatOption(simType stats.SimType, statType stats.StatType, entry weight_types.Weight3ExtendedStatEntry, statTotalColumn *columnInfo) *columnInfo {
	optionValueCol := calc3.makeSimStatOptionColumn3(simType, statType, entry.StatRange)

	row := util_highs.ConstraintRow{}
	row.Add(statTotalColumn.columnIndex, entry.RatingWeight)
	row.Add(optionValueCol.columnIndex, -1)
	row.Build(calc3.build, -entry.RatingOffset, -entry.RatingOffset)

	return optionValueCol
}

//func (calc3 *gearWeight3Calc) entryInRange(entry weight_types.Weight3ExtendedStatEntry, statTotalColumn *columnInfo) util_highs.ColumnIndex {
//	// check if stat total fits into this range
//	var rangeCondition util_highs.ColumnIndex
//	if entry.StatRange.Maximum < math.MaxUint32 {
//		rangeCondition = calc3.statIsBetween(statTotalColumn, entry.StatRange)
//	} else {
//		rangeCondition = calc3.build.ColumnIsGreaterOrEqualThanConstant(statTotalColumn.columnIndex, float64(entry.StatRange.Minimum), c_gearExtended3StatHigh, 1.0)
//	}
//	return rangeCondition
//}

func (calc3 *gearWeight3Calc) makeSimValueColumn(simType stats.SimType) *columnInfo {
	simValueColumn := &columnInfo{entryType: entry_sim_value, simType: simType}
	simValueColumn.columnIndex = calc3.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), simValueColumn)
	return simValueColumn
}

func (calc3 *gearWeight3Calc) makeSimStatValueColumn3(simType stats.SimType, statType stats.StatType) *columnInfo {
	simStatValueColumn := &columnInfo{entryType: entry_sim_stat_value, simType: simType, statType: statType}
	simStatValueColumn.columnIndex = calc3.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), simStatValueColumn)
	return simStatValueColumn
}

func (calc3 *gearWeight3Calc) makeSimStatOptionColumn3(simType stats.SimType, statType stats.StatType, statRange weight_types.StatRange) *columnInfo {
	valueOptionColumn := &columnInfo{entryType: entry_sim_stat_value_option, simType: simType, statType: statType, statRange: statRange}
	valueOptionColumn.columnIndex = calc3.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), valueOptionColumn)
	return valueOptionColumn
}

//func (calc3 *gearWeight3Calc) statIsBetween(statTotalColumn *columnInfo, statRange weight_types.StatRange) util_highs.ColumnIndex {
//	minimum := float64(statRange.Minimum)
//	maximum := float64(statRange.Maximum)
//	return calc3.build.ColumnIsBetweenConstants(statTotalColumn.columnIndex, minimum, maximum, c_gearExtended3StatHigh, 1.0)
//}

// (!A) && B == out
func columnBoolNotAAndB(build *util_highs.LinearBuilder, a util_highs.ColumnIndex, b util_highs.ColumnIndex) util_highs.ColumnIndex {
	out := build.CreateColumnBool(nil)
	rowSet := util_highs.ConstraintRow{}
	/*
	   first row only logic:
	   abx
	   000 out is free
	   011 out forced 1
	   100 out forced 0
	   110 out is free
	*/
	rowSet.Add(a, -1)
	rowSet.Add(b, 1)
	rowSet.Add(out, -1)
	rowSet.Build(build, -1, 0)

	rowCheckA := util_highs.ConstraintRow{}
	rowCheckA.Add(a, 1)
	rowCheckA.Add(out, 1)
	rowCheckA.Build(build, util_highs.InfNeg(), 1)

	rowCheckB := util_highs.ConstraintRow{}
	rowCheckB.Add(b, -1)
	rowCheckB.Add(out, 1)
	rowCheckB.Build(build, util_highs.InfNeg(), 0)

	return out
}
