package solve_highs

import (
	"github.com/nerago/mopgear-go/gear_model/bonus_set"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"

	"github.com/bartolsthoorn/gohighs/highs"
)

const c_bonusItemCountRangeHigh = 10
const c_bonusItemCountEqualDelta = 1.0

type gearBonusComboHandler struct {
	build        *util_highs.LinearBuilder
	_bonusCombos *util_collection.List[bonusCombo]
	_bonusData   map[solve_highs_types.SetBonusIndex]bonusColsByCount
}

func (bon *gearBonusComboHandler) processBonus(combinedRatingVar *columnInfo, simType util_collection.Optional[stats.SimType], ratingBigM, ratingMax float64, model *solve_highs_types.SolverModel, countSetItemsColumns map[solve_highs_types.SetBonusIndex]*columnInfo) (*columnInfo, error) {
	bonusData, bonusCombos, err := bon.init(model, countSetItemsColumns)
	if err != nil {
		return nil, err
	}

	if len(bonusData) > 0 {
		outputVar := bon._makeOutputForComboVariable(simType, ratingMax)

		// TODO consider handling for single combo, would need multiplier
		for combo := range bonusCombos.SeqValuePointers() {
			bonusMultiplier := bon.totalComboMultiplier(combo, simType, &model.SetBonus)
			bon._forComboCopyToOutput(combo, combinedRatingVar, outputVar, bonusMultiplier, ratingBigM)
		}

		err := bon._addSetNeededCounts(countSetItemsColumns, model.SetBonus.RequiredCounts, model.SetBonus.CountMode)
		if err != nil {
			return nil, err
		}

		return outputVar, nil
	} else {
		return combinedRatingVar, nil
	}
}

func (bon *gearBonusComboHandler) init(model *solve_highs_types.SolverModel, countSetItemsCol map[solve_highs_types.SetBonusIndex]*columnInfo) (map[solve_highs_types.SetBonusIndex]bonusColsByCount, *util_collection.List[bonusCombo], error) {
	bonusData := bon._bonusData
	bonusCombos := bon._bonusCombos
	if bonusData == nil {
		allowedRanges := bon.determineAllowedBonusRanges(model)
		bonusData = bon._makeBonusData(model, allowedRanges, countSetItemsCol)

		if len(bonusData) > 0 {
			bonusCombos = bon.combosMakeInitial(bonusData, allowedRanges)
			if err := bon.finishComboRules(bonusCombos); err != nil {
				return nil, nil, err
			}
			bon._bonusCombos = bonusCombos
		}

		bon._bonusData = bonusData
	}
	return bonusData, bonusCombos, nil
}

func (bon *gearBonusComboHandler) determineAllowedBonusRanges(model *solve_highs_types.SolverModel) []util_collection.HiLoUInt32 {
	setTotalCount := model.SetBonus.TotalCount
	requiredOptions := model.SetBonus.RequiredCounts
	countMode := model.SetBonus.CountMode

	if len(requiredOptions) > 0 {
		countRanges := util_collection.RepeatValue(
			util_collection.HiLoUInt32{Lo: c_maxSetItems, Hi: 0},
			setTotalCount)

		for _, option := range requiredOptions {
			for i := range setTotalCount {
				setIndex := solve_highs_types.SetBonusIndex(i)
				hilo := &countRanges[setIndex]
				if needCount, optionHasEntry := option[setIndex]; optionHasEntry {
					switch countMode {
					case bonus_set.CountMode_Exact:
						hilo.Lo = min(hilo.Lo, uint32(needCount))
						hilo.Hi = max(hilo.Hi, uint32(needCount))
					case bonus_set.CountMode_Minimum:
						hilo.Lo = min(hilo.Lo, uint32(needCount))
						hilo.Hi = c_maxSetItems
					case bonus_set.CountMode_AllowPlusOne:
						hilo.Lo = min(hilo.Lo, uint32(needCount))
						hilo.Hi = max(hilo.Hi, uint32(needCount+1))
					}
				} else {
					// if an option has no rule for set then full range allowed
					hilo.Lo = 0
					hilo.Hi = c_maxSetItems
				}
			}
		}

		return countRanges
	} else {
		// no rules then have to allow full ranges
		return util_collection.RepeatValue(
			util_collection.HiLoUInt32{Lo: 0, Hi: c_maxSetItems},
			setTotalCount)
	}
}

func (bon *gearBonusComboHandler) combosMakeInitial(bonusColumns map[solve_highs_types.SetBonusIndex]bonusColsByCount, allowedRanges []util_collection.HiLoUInt32) *util_collection.List[bonusCombo] {
	bonusCombos := &util_collection.List[bonusCombo]{}
	bon._makeCombos(bonusColumns, bonusCombos, allowedRanges)
	bon._bonusCombos = bonusCombos
	return bonusCombos
}

func (bon *gearBonusComboHandler) finishComboRules(bonusCombos *util_collection.List[bonusCombo]) error {
	for combo := range bonusCombos.SeqValuePointers() {
		err := bon._buildComboActivatingVar(combo)
		if err != nil {
			return err
		}
	}

	checkSingleCombo := util_highs.ConstraintRow{Debug: "checkSingleCombo"}
	for combo := range bonusCombos.SeqValuePointers() {
		checkSingleCombo.Add(combo.activatingVar.columnIndex, 1)
	}

	// don't allow overlapping ranges
	// TODO validate these conditions elsewhere for better errors
	checkSingleCombo.Build(bon.build, 1, 1)

	return nil
}

func (bon *gearBonusComboHandler) _makeOutputForComboVariable(simTypeOptional util_collection.Optional[stats.SimType], ratingMax float64) *columnInfo {
	if simType, hasType := simTypeOptional.GetWithFlag(); hasType {
		entry := columnInfo{entryType: entry_sim_value_combo, simType: simType}
		entry.columnIndex = bon.build.CreateColumnGeneral(highs.Continuous, -ratingMax, ratingMax, &entry)
		return &entry
	} else {
		entry := &columnInfo{entryType: entry_main_output}
		entry.columnIndex = bon.build.CreateColumnGeneral(highs.Continuous, -ratingMax, ratingMax, entry)
		return entry
	}
}

func (bon *gearBonusComboHandler) _makeBonusData(model *solve_highs_types.SolverModel, allowedRanges []util_collection.HiLoUInt32, countSetItemsCol map[solve_highs_types.SetBonusIndex]*columnInfo) map[solve_highs_types.SetBonusIndex]bonusColsByCount {
	// constrain: exact item count in each active set
	if model.SetBonus.TotalCount > 0 {
		bonusColsBySetByCount := make(map[solve_highs_types.SetBonusIndex]bonusColsByCount, model.SetBonus.TotalCount)

		for i := range model.SetBonus.TotalCount {
			setIndex := solve_highs_types.SetBonusIndex(i)
			countRange := allowedRanges[setIndex]
			countColumn := countSetItemsCol[setIndex]

			bon.build.ChangeColumnMinMax(countColumn.columnIndex, float64(countRange.Lo), float64(countRange.Hi))

			colsByCount := bon._addSetItemsCountExactVariables(setIndex, countColumn, countRange)
			bonusColsBySetByCount[setIndex] = colsByCount
		}

		return bonusColsBySetByCount
	} else {
		return nil
	}
}

func (bon *gearBonusComboHandler) _addSetItemsCountExactVariables(setIndex solve_highs_types.SetBonusIndex, totalCountColumn *columnInfo, countRange util_collection.HiLoUInt32) bonusColsByCount {
	exactCountVars := bonusColsByCount{}

	if countRange.Size() == 1 {
		// TODO feeling a bit weird, never check its actually active.
		// TODO if this is literally only condition then could change more
		// only one possible count
		itemCount := countRange.Lo
		boolColumn := &columnInfo{entryType: entry_set_exact_count, setIndex: setIndex, itemCount: int(itemCount)}
		boolColumn.columnIndex = bon.build.CreateColumnGeneral(highs.Integer, 1, 1, boolColumn)
		exactCountVars[itemCount] = boolColumn
	} else {
		// compare total number of items previous computed into this constraint
		compareRow := util_highs.ConstraintRow{Debug: "setItemsCompareRow"}
		compareRow.Add(totalCountColumn.columnIndex, -1)

		// constraint so only one of these flags gets set
		singleFlagOnly := util_highs.ConstraintRow{Debug: "setItemsSingleFlagOnly"}

		// make a bool for each possible count in range 0..5 (or less depending on other rules)
		for itemCount := countRange.Lo; itemCount <= countRange.Hi; itemCount++ {
			boolColumn := &columnInfo{entryType: entry_set_exact_count, setIndex: setIndex, itemCount: int(itemCount)}
			boolColumn.columnIndex = bon.build.CreateColumnBool(boolColumn)

			// should activate this flag which will match the total count
			compareRow.Add(boolColumn.columnIndex, float64(itemCount))

			// but only one flag at a time
			singleFlagOnly.Add(boolColumn.columnIndex, 1)

			exactCountVars[itemCount] = boolColumn
		}

		compareRow.Build(bon.build, 0, 0)     // equal
		singleFlagOnly.Build(bon.build, 1, 1) // sum of flags should be just one, should pull the zero flag up if no other set
	}

	return exactCountVars
}

func (bon *gearBonusComboHandler) _forComboCopyToOutput(combo *bonusCombo, inputVar *columnInfo, outputVar *columnInfo, bonusMultiplier float64, ratingBigM float64) {
	activatingVar := combo.activatingVar
	if util.FloatEqualsZero(bonusMultiplier) {
		bonusMultiplier = 1.0
	}

	assertColumnBoolOrLess(bon.build, activatingVar.columnIndex)
	assertColumnRangeSmallerThanBigM(bon.build, inputVar.columnIndex, ratingBigM)
	assertColumnRangeSmallerThanBigM(bon.build, outputVar.columnIndex, ratingBigM)

	bon.build.ConstraintCopyIfBool(
		activatingVar.columnIndex,
		inputVar.columnIndex, bonusMultiplier,
		outputVar.columnIndex,
		ratingBigM,
	)
}

func (bon *gearBonusComboHandler) totalComboMultiplier(combo *bonusCombo, simTypeOptional util_collection.Optional[stats.SimType], modelBonus *solve_highs_types.SolverModelSetBonus) float64 {
	bonusMultiplier := 1.0
	for _, entry := range combo.condition {
		setIndex := entry.bonusSetIndex
		count := entry.count

		var thisBonus float64
		if simType, hasType := simTypeOptional.GetWithFlag(); hasType {
			thisBonus = modelBonus.MultipliersBySim[setIndex][count].GetOrDefault(simType, 1)
		} else {
			thisBonus = modelBonus.MultipliersFlat[setIndex][count]
		}

		bonusMultiplier *= thisBonus
	}
	return bonusMultiplier
}

func (bon *gearBonusComboHandler) _makeCombos(bonusColumns map[solve_highs_types.SetBonusIndex]bonusColsByCount, bonusCombos *util_collection.List[bonusCombo], allowedRanges []util_collection.HiLoUInt32) {
	bonusColumnSlice := make([]bonusColsByCount, len(bonusColumns))
	for bonusIndex, cols := range bonusColumns {
		bonusColumnSlice[bonusIndex] = cols
	}
	bon._makeCombosRecur(bonusColumnSlice, allowedRanges, nil, 0, bonusCombos, 0)
}

func (bon *gearBonusComboHandler) _makeCombosRecur(bonusColumns []bonusColsByCount, allowedRanges []util_collection.HiLoUInt32, built []comboEntry, builtComboItemCount uint32, bonusCombos *util_collection.List[bonusCombo], bonusSetIndex solve_highs_types.SetBonusIndex) {
	if len(bonusColumns) == 0 || builtComboItemCount == c_maxSetItems {
		bonusCombos.AppendLast(bonusCombo{condition: built})
	} else {
		addSet := &bonusColumns[0]
		addRange := allowedRanges[0]
		for itemCount := addRange.Lo; itemCount <= min(addRange.Hi, c_maxSetItems-builtComboItemCount); itemCount++ {
			next := comboEntry{bonusSetIndex, itemCount, addSet[itemCount]}
			progress := util_collection.CopyAndAppend(built, next)
			bon._makeCombosRecur(bonusColumns[1:], allowedRanges[1:], progress, builtComboItemCount+itemCount, bonusCombos, bonusSetIndex+1)
		}
	}
}

// logical AND between exact count vars
func (bon *gearBonusComboHandler) _buildComboActivatingVar(combo *bonusCombo) error {
	if len(combo.condition) > 1 {
		comboActiveBool := &columnInfo{entryType: entry_combo_active, combo: combo}
		comboActiveBool.columnIndex = bon.build.CreateColumnBool(comboActiveBool)
		combo.activatingVar = comboActiveBool

		buildAnd := util_highs.ConstraintAndBuilder{}
		for _, entry := range combo.condition {
			specificExactBool := entry.exactSetCountVar
			buildAnd.AddInput(specificExactBool.columnIndex)
		}
		buildAnd.SetOutput(comboActiveBool.columnIndex)
		buildAnd.Build(bon.build)
	} else if len(combo.condition) == 1 {
		entry := combo.condition[0]
		combo.activatingVar = entry.exactSetCountVar
	} else {
		return util.ErrorTracedNew("empty condition doesn't make sense")
	}
	return nil
}

func (bon *gearBonusComboHandler) _addSetNeededCounts(countSetItemsCol map[solve_highs_types.SetBonusIndex]*columnInfo, setBonusRequired []solve_highs_types.SetBonusRequiredCounts, countMode bonus_set.ItemCountsRequiredMode) error {
	if len(setBonusRequired) > 0 {
		if len(countSetItemsCol) == 0 {
			return util.ErrorTracedNew("no countSetItemsCol to use for addSetNeededCounts")
		} else if len(setBonusRequired) == 1 && len(setBonusRequired[0]) == 1 {
			setIndex, needCount := util_collection.MapFirstEntry(setBonusRequired[0])
			setCountCol := countSetItemsCol[setIndex]

			lo, hi, err := needCountToHiLo(countMode, needCount)
			if err != nil {
				return err
			}

			bon.build.ChangeColumnMinMax(setCountCol.columnIndex, lo, hi)
		} else {
			oneOfTheseOptions := util_highs.ConstraintRow{Debug: "oneOfTheseOptions"}
			for _, option := range setBonusRequired {
				if optionActive, err := bon.addOption(option, countMode, countSetItemsCol); err == nil {
					oneOfTheseOptions.Add(optionActive, 1)
				} else {
					return nil
				}
			}
			oneOfTheseOptions.Build(bon.build, 1, util_highs.InfPos())
		}
	}
	return nil
}

func needCountToHiLo(countMode bonus_set.ItemCountsRequiredMode, needCount uint8) (float64, float64, error) {
	lo := float64(0)
	hi := float64(c_maxSetItems)
	switch countMode {
	case bonus_set.CountMode_Exact:
		lo = float64(needCount)
		hi = float64(needCount)
	case bonus_set.CountMode_Minimum:
		lo = float64(needCount)
	case bonus_set.CountMode_AllowPlusOne:
		lo = float64(needCount)
		hi = float64(needCount + 1)
	default:
		return 0, 0, util.ErrorTracedNew("unknown type")
	}
	return lo, hi, nil
}

func (bon *gearBonusComboHandler) addOption(option solve_highs_types.SetBonusRequiredCounts, countMode bonus_set.ItemCountsRequiredMode, countSetItemsCol map[solve_highs_types.SetBonusIndex]*columnInfo) (util_highs.ColumnIndex, error) {
	if len(option) > 1 {
		optionParts := util_highs.ConstraintAndBuilder{}

		for setIndex, needCount := range option {
			inRange, err := bon.addOptionPart(setIndex, needCount, countSetItemsCol, countMode)
			if err != nil {
				return 0, err
			}
			optionParts.AddInput(inRange)
		}

		optionActive := bon.build.CreateColumnBool(util_highs.DebugText("SetBonusRequired option"))
		optionParts.SetOutput(optionActive)
		optionParts.Build(bon.build)
		return optionActive, nil
	} else if len(option) == 1 {
		setIndex, needCount := util_collection.MapFirstEntry(option)
		return bon.addOptionPart(setIndex, needCount, countSetItemsCol, countMode)
	} else {
		return 0, util.ErrorTracedNew("empty option")
	}
}

func (bon *gearBonusComboHandler) addOptionPart(setIndex solve_highs_types.SetBonusIndex, needCount uint8, countSetItemsCol map[solve_highs_types.SetBonusIndex]*columnInfo, countMode bonus_set.ItemCountsRequiredMode) (util_highs.ColumnIndex, error) {
	setCountCol := countSetItemsCol[setIndex]

	// in order for c_bonusItemCountRangeHigh to work efficiently make sure the column we're using has smaller ranges
	columnMin, columnMax := bon.build.GetColumnMinMax(setCountCol.columnIndex)
	if columnMin < 0 || columnMax > c_maxSetItems {
		panic("setCountCol has excessive range")
	}

	var inRange util_highs.ColumnIndex
	switch countMode {
	case bonus_set.CountMode_Exact:
		inRange = bon.build.CreateColumnBool(nil)
		bon.build.ColumnIsEqualConstant(setCountCol.columnIndex, inRange, float64(needCount), c_bonusItemCountRangeHigh, c_bonusItemCountEqualDelta)
	case bonus_set.CountMode_Minimum:
		inRange = bon.build.ColumnIsGreaterOrEqualThanConstant(setCountCol.columnIndex, float64(needCount), c_bonusItemCountRangeHigh, c_bonusItemCountEqualDelta)
	case bonus_set.CountMode_AllowPlusOne:
		inRange = bon.build.ColumnIsBetweenConstants(setCountCol.columnIndex, float64(needCount), float64(needCount+1), c_bonusItemCountRangeHigh, c_bonusItemCountEqualDelta)
	default:
		return 0, util.ErrorTracedNew("unknown type")
	}

	return inRange, nil
}

func (bon *gearBonusComboHandler) checkActiveCombo(solution util_highs.ISolution, solvableItemSet *items.SolvableItemSet, model *solve_highs_types.SolverModel) error {
	bonusCombos := bon._bonusCombos

	if bonusCombos != nil && bonusCombos.Size() > 0 {
		var activeCombo *bonusCombo

		for combo := range bonusCombos.SeqValuePointers() {
			if solution.ValueIsOne(combo.activatingVar.columnIndex) {
				if activeCombo == nil {
					activeCombo = combo
				} else {
					return util.ErrorTracedNew("multiple combos active")
				}
			}
		}

		if activeCombo != nil {
			for _, entry := range activeCombo.condition {
				countItems := model.SetBonus.CountItems[entry.bonusSetIndex]
				actualCount := countItems(solvableItemSet.Items())
				if actualCount != uint8(entry.count) {
					return util.ErrorTracedNew("number of items not what solver returned")
				}
			}
		} else {
			return util.ErrorTracedNew("no combos active")
		}
	}

	return nil
}
