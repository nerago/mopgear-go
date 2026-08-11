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

func SingleGearSetExtended3Main(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder) *util_async.FutureCancellable[items.SolvableItemSet] {
	build := util_highs.LinearBuilder{}
	build.Solver = util_highs.Solver_MIP_Interior

	setup := makeGearSetExtended3(&build, model, itemOptions, 1)

	return setup.runForFutureResult(itemOptions, model, printer, 1)
}

func makeGearSetExtended3(build *util_highs.LinearBuilder, model *solve_highs_types.SolverModel, itemOptions *items.SolvableOptionsMap, scaleOutputRating float64) *singleGearSetExtended3 {
	setup := singleGearSetExtended3{
		singleGearSetExtended: singleGearSetExtended{
			singleGearSetShared: singleGearSetShared{build: build},
		},
	}

	setup.prepareCommon(model, itemOptions, scaleOutputRating)
	setup.prepareStatTotals()
	setup.prepareRequireEx(&model.StatRequirements)

	for slot, item := range itemOptions.AllItemSlotSeq() {
		setup.addItem(slot, item, &model.StatRequirements, model.SetBonusIndexForItem)
	}

	setup.finishItemsCommon(itemOptions)
	setup.finishRequireEx(&model.StatRequirements)
	setup.finishStatTotals()

	// statTotalColumns[statType] -> simValueTotalColumns[simType]
	setup.calcSimValues(model.Weights3)
	// simValueTotalColumns * activeCombo -> simValueComboColumns -> mainOutputVar
	setup.calcFromSimValueToOutput(model, &model.Weights3.SimPriority)

	return &setup
}

type singleGearSetExtended3 struct {
	singleGearSetExtended
}

func (setup *singleGearSetExtended3) calcSimValues(weight3 *weight_types.Weight3ExtendedRanged) {
	setup.simValueTotalColumns = make(map[stats.SimType]*columnInfo)
	for _, simType := range weight3.SimList {
		setup.calcValueForSimType(simType, weight3)
	}
}

func (setup *singleGearSetExtended3) calcValueForSimType(simType stats.SimType, weight *weight_types.Weight3ExtendedRanged) {
	simValueColumn := setup.makeSimValueColumn(simType)
	simValueFromStatRow := util_highs.ConstraintRow{}

	for _, statType := range weight.StatList {
		chosenSimStatContribution := setup.calcValueForSimAndStatType(simType, statType, weight)
		simValueFromStatRow.Add(chosenSimStatContribution.columnIndex, 1)
	}

	simEntry := weight.SimPriority.GetOrPanic(simType)
	offset := -simEntry.RangingOffset
	valueScale := -1.0 / simEntry.RangingScale
	simValueFromStatRow.Add(simValueColumn.columnIndex, valueScale)
	simValueFromStatRow.Build(setup.build, offset, offset)
}

func (setup *singleGearSetExtended3) calcValueForSimAndStatType(simType stats.SimType, statType stats.StatType, weight *weight_types.Weight3ExtendedRanged) columnInfo {
	chosenSimStatContribution := setup.makeSimStatValueColumn(simType, statType)
	statTotalColumn := setup.statTotalColumns[statType]

	checkSingleRangeActive := util_highs.ConstraintRow{}
	for entry := range weight.StatWeights.GetAsSeq(simType, statType) {
		// check if stat total fits into this range
		var rangeCondition util_highs.ColumnIndex
		if entry.StatRange.Maximum < math.MaxUint32 {
			rangeCondition = setup.statIsBetween(statTotalColumn, entry.StatRange)
		} else {
			rangeCondition = setup.build.ColumnIsGreaterOrEqualThanConstant(statTotalColumn.columnIndex, float64(entry.StatRange.Minimum), c_gearExtended3StatHigh, 1.0)
		}
		checkSingleRangeActive.Add(rangeCondition, 1)

		// then copy a score to contributeScoreSimStat
		optionValueCol := setup.makeSimStatOptionColumn(simType, statType, entry.StatRange)
		setup.calcValueForEntry(statTotalColumn, entry, optionValueCol)
		setup.build.ConstraintCopyIfBool(rangeCondition, optionValueCol.columnIndex, 1, chosenSimStatContribution.columnIndex, c_gearExtended3ScoreHigh)
	}
	checkSingleRangeActive.Build(setup.build, 1, 1)

	return chosenSimStatContribution
}

func (setup *singleGearSetExtended3) calcValueForEntry(statTotalColumn *columnInfo, entry weight_types.Weight3ExtendedStatEntry, entryValue columnInfo) {
	row := util_highs.ConstraintRow{}
	row.Add(statTotalColumn.columnIndex, entry.RatingWeight)
	row.Add(entryValue.columnIndex, -1)
	row.Build(setup.build, -entry.RatingOffset, -entry.RatingOffset)
}

func (setup *singleGearSetExtended3) makeSimValueColumn(simType stats.SimType) columnInfo {
	simValueColumn := columnInfo{entryType: entry_sim_value, simType: simType}
	simValueColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &simValueColumn)
	setup.simValueTotalColumns[simType] = &simValueColumn
	setup.allColumns = append(setup.allColumns, &simValueColumn)
	return simValueColumn
}

func (setup *singleGearSetExtended3) makeSimStatValueColumn(simType stats.SimType, statType stats.StatType) columnInfo {
	simStatValueColumn := columnInfo{entryType: entry_sim_stat_value, simType: simType, statType: statType}
	simStatValueColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &simStatValueColumn)
	setup.allColumns = append(setup.allColumns, &simStatValueColumn)
	return simStatValueColumn
}

func (setup *singleGearSetExtended3) makeSimStatOptionColumn(simType stats.SimType, statType stats.StatType, statRange weight_types.StatRange) columnInfo {
	valueOptionColumn := columnInfo{entryType: entry_sim_stat_value_option, simType: simType, statType: statType, statRange: statRange}
	valueOptionColumn.columnIndex = setup.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &valueOptionColumn)
	setup.allColumns = append(setup.allColumns, &valueOptionColumn)
	return valueOptionColumn
}

func (setup *singleGearSetExtended3) statIsBetween(statTotalColumn *columnInfo, statRange weight_types.StatRange) util_highs.ColumnIndex {
	minimum := float64(statRange.Minimum)
	maximum := float64(statRange.Maximum)
	return setup.build.ColumnIsBetweenConstants(statTotalColumn.columnIndex, minimum, maximum, c_gearExtended3StatHigh, 1.0)
}
