package solve_highs

import (
	"math"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const c_gearExtended3StatHigh = 100000
const c_gearExtended3ScoreHigh = 10

func SingleGearSetExtended3Main(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior
	build.TimeLimitSeconds = timeout
	build.AddOptionBool("presolve_rule_logging", true)

	se3 := makeGearSetExtended3(&build)
	outputVar := se3.createOutputVariableForSeparateRun()
	se3.setup(model, itemOptions, outputVar)

	return se3.runForFutureResult(itemOptions, model, printer)
}

func makeGearSetExtended3(build *util_highs.LinearBuilder) *singleGearSetExtended3 {
	return &singleGearSetExtended3{
		singleGearSetExtended: singleGearSetExtended{
			singleGearSetShared: singleGearSetShared{
				build:          build,
				ratingPreScale: 1,
			},
		},
	}
}

func (se3 *singleGearSetExtended3) setup(model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap, outputVar *columnInfo) {
	se3.prepareCommon(model, itemOptions)
	se3.prepareStatTotals()
	se3.prepareRequireEx(&model.StatRequirements)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		se3.addItem(slot, item, &model.StatRequirements, model.SetBonus.IndexForItem)
	}

	se3.finishItemsCommon(itemOptions)
	se3.finishRequireEx(&model.StatRequirements)
	se3.finishStatTotals()

	// statTotalColumns[statType] -> simValueTotalColumns[simType]
	se3.calcSimValues(model.Weights3)
	// simValueTotalColumns * activeCombo -> simValueComboColumns -> mainOutputVar
	se3.calcFromSimValueToOutput(model, &model.Weights3.SimPriority, outputVar)
}

type singleGearSetExtended3 struct {
	singleGearSetExtended
}

func (se3 *singleGearSetExtended3) calcSimValues(weight3 *weight_types.Weight3ExtendedRanged) {
	se3.simValueTotalColumns = make(map[stats.SimType]*columnInfo)
	for _, simType := range weight3.SimList {
		se3.calcValueForSimType(simType, weight3)
	}
}

func (se3 *singleGearSetExtended3) calcValueForSimType(simType stats.SimType, weight *weight_types.Weight3ExtendedRanged) {
	simValueColumn := se3.makeSimValueColumn(simType)
	simValueFromStatRow := util_highs.ConstraintRow{}

	for _, statType := range weight.StatList {
		chosenSimStatContribution := se3.calcValueForSimAndStatType(simType, statType, weight)
		simValueFromStatRow.Add(chosenSimStatContribution.columnIndex, 1)
	}

	simEntry := weight.SimPriority.GetOrPanic(simType)
	offset := -simEntry.RangingOffset
	valueScale := -1.0 / simEntry.RangingScale
	simValueFromStatRow.Add(simValueColumn.columnIndex, valueScale)
	simValueFromStatRow.Build(se3.build, offset, offset)
}

func (se3 *singleGearSetExtended3) calcValueForSimAndStatType(simType stats.SimType, statType stats.StatType, weight *weight_types.Weight3ExtendedRanged) columnInfo {
	chosenSimStatContribution := se3.makeSimStatValueColumn(simType, statType)
	statTotalColumn := se3.statTotalColumns[statType]

	checkSingleRangeActive := util_highs.ConstraintRow{}
	for entry := range weight.StatWeights.GetAsSeq(simType, statType) {
		// check if stat total fits into this range
		var rangeCondition util_highs.ColumnIndex
		if entry.StatRange.Maximum < math.MaxUint32 {
			rangeCondition = se3.statIsBetween(statTotalColumn, entry.StatRange)
		} else {
			rangeCondition = se3.build.ColumnIsGreaterOrEqualThanConstant(statTotalColumn.columnIndex, float64(entry.StatRange.Minimum), c_gearExtended3StatHigh, 1.0)
		}
		checkSingleRangeActive.Add(rangeCondition, 1)

		// then copy a score to contributeScoreSimStat
		optionValueCol := se3.makeSimStatOptionColumn(simType, statType, entry.StatRange)
		se3.calcValueForEntry(statTotalColumn, entry, optionValueCol)
		se3.build.ConstraintCopyIfBool(rangeCondition, optionValueCol.columnIndex, 1, chosenSimStatContribution.columnIndex, c_gearExtended3ScoreHigh)
	}
	checkSingleRangeActive.Build(se3.build, 1, 1)

	return chosenSimStatContribution
}

func (se3 *singleGearSetExtended3) calcValueForEntry(statTotalColumn *columnInfo, entry weight_types.Weight3ExtendedStatEntry, entryValue columnInfo) {
	row := util_highs.ConstraintRow{}
	row.Add(statTotalColumn.columnIndex, entry.RatingWeight)
	row.Add(entryValue.columnIndex, -1)
	row.Build(se3.build, -entry.RatingOffset, -entry.RatingOffset)
}

func (se3 *singleGearSetExtended3) makeSimValueColumn(simType stats.SimType) columnInfo {
	simValueColumn := columnInfo{entryType: entry_sim_value, simType: simType}
	simValueColumn.columnIndex = se3.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &simValueColumn)
	se3.simValueTotalColumns[simType] = &simValueColumn
	se3.allColumns = append(se3.allColumns, &simValueColumn)
	return simValueColumn
}

func (se3 *singleGearSetExtended3) makeSimStatValueColumn(simType stats.SimType, statType stats.StatType) columnInfo {
	simStatValueColumn := columnInfo{entryType: entry_sim_stat_value, simType: simType, statType: statType}
	simStatValueColumn.columnIndex = se3.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &simStatValueColumn)
	se3.allColumns = append(se3.allColumns, &simStatValueColumn)
	return simStatValueColumn
}

func (se3 *singleGearSetExtended3) makeSimStatOptionColumn(simType stats.SimType, statType stats.StatType, statRange weight_types.StatRange) columnInfo {
	valueOptionColumn := columnInfo{entryType: entry_sim_stat_value_option, simType: simType, statType: statType, statRange: statRange}
	valueOptionColumn.columnIndex = se3.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &valueOptionColumn)
	se3.allColumns = append(se3.allColumns, &valueOptionColumn)
	return valueOptionColumn
}

func (se3 *singleGearSetExtended3) statIsBetween(statTotalColumn *columnInfo, statRange weight_types.StatRange) util_highs.ColumnIndex {
	minimum := float64(statRange.Minimum)
	maximum := float64(statRange.Maximum)
	return se3.build.ColumnIsBetweenConstants(statTotalColumn.columnIndex, minimum, maximum, c_gearExtended3StatHigh, 1.0)
}
