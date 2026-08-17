package solve_highs

import (
	"paladin_gearing_go/gear_model/bonus_set"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"

	"github.com/bartolsthoorn/gohighs/highs"
)

type gearBonusComboHandler struct {
	build               *util_highs.LinearBuilder
	_remove_bonusCombos *util_collection.List[bonusCombo]
	_remove_bonusData   []bonusInfo
}

func (bon *gearBonusComboHandler) ProcessBonus(combinedRatingVar *columnInfo, simType util_collection.Optional[stats.SimType], rangeHigh float64, model *solve_highs_types.SolverModel, countSetItemsCol map[solve_highs_types.SetBonusIndex]*columnInfo) *columnInfo {
	bonusData := bon.makeBonusData(model, countSetItemsCol, simType)
	if len(bonusData) > 0 {
		bonusCombos := bon.makeCombos(bonusData)

		outputVar := bon.makeOutputForComboVariable(simType)

		for combo := range bonusCombos.SeqValuePointers() {
			bon.forComboCopyToOutput(combo, combinedRatingVar, outputVar, rangeHigh)
		}

		bon._remove_bonusData = bonusData
		bon._remove_bonusCombos = bonusCombos

		return outputVar

		// TODO could sometimes return without processing but need to check multiplier
	} else {
		return combinedRatingVar
	}
}

func (bon *gearBonusComboHandler) makeOutputForComboVariable(simTypeOptional util_collection.Optional[stats.SimType]) *columnInfo {
	if simType, hasType := simTypeOptional.GetWithFlag(); hasType {
		entry := columnInfo{entryType: entry_sim_value_combo, simType: simType}
		entry.columnIndex = bon.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), &entry)
		return &entry
	} else {
		entry := &columnInfo{entryType: entry_main_output}
		entry.columnIndex = bon.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), entry)
		return entry
	}
}

func (bon *gearBonusComboHandler) makeBonusData(model *solve_highs_types.SolverModel, countSetItemsCol map[solve_highs_types.SetBonusIndex]*columnInfo, simTypeOptional util_collection.Optional[stats.SimType]) (bonusData []bonusInfo) {
	// constrain: exact item count in each active set
	if model.SetBonus.TotalCount > 0 {
		bonusData = make([]bonusInfo, model.SetBonus.TotalCount)
		for i := range model.SetBonus.TotalCount {
			setIndex := solve_highs_types.SetBonusIndex(i)
			info := bonusInfo{
				setIndex:          setIndex,
				setTotalCountVar:  countSetItemsCol[setIndex],
				setExactCountVars: bon.addSetItemsCountExactVariables(setIndex, countSetItemsCol[setIndex]),
			}

			if simType, hasType := simTypeOptional.GetWithFlag(); hasType {
				bySim := model.SetBonus.MultipliersBySim[setIndex]
				for c := range bySim {
					info.setMultipliers[c] = bySim[c].GetOrDefault(simType, 1)
				}
			} else {
				info.setMultipliers = model.SetBonus.MultipliersFlat[setIndex]
			}

			bonusData[setIndex] = info
		}
	}
	return bonusData
}

func (bon *gearBonusComboHandler) forComboCopyToOutput(combo *bonusCombo, inputVar *columnInfo, outputVar *columnInfo, rangeHigh float64) {
	activatingVar := combo.activatingVar
	bonusMultiplier := combo.totalMultiplier()
	if util.FloatEqualsZero(bonusMultiplier) {
		bonusMultiplier = 1.0
	}

	bon.build.ConstraintCopyIfBool(
		activatingVar.columnIndex,
		inputVar.columnIndex, bonusMultiplier,
		outputVar.columnIndex,
		rangeHigh,
	)
}

func (bon *gearBonusComboHandler) makeCombos(bonusData []bonusInfo) (bonusCombos *util_collection.List[bonusCombo]) {
	bonusCombos = &util_collection.List[bonusCombo]{}
	bon.makeCombosRecur(bonusData, nil, 0, bonusCombos)

	for combo := range bonusCombos.SeqValuePointers() {
		bon.buildComboActivatingVar(combo)
	}

	checkSingleCombo := util_highs.ConstraintRow{}
	for combo := range bonusCombos.SeqValuePointers() {
		checkSingleCombo.Add(combo.activatingVar.columnIndex, 1)
	}
	checkSingleCombo.Build(bon.build, 1, 1)

	return bonusCombos
}

func (bon *gearBonusComboHandler) makeCombosRecur(setData []bonusInfo, built []bonusWithCount, builtComboItemCount int, bonusCombos *util_collection.List[bonusCombo]) {
	if len(setData) == 0 || builtComboItemCount == c_maxSetItems {
		bonusCombos.AppendLast(bonusCombo{condition: built})
	} else {
		addSet := &setData[0]
		for itemCount := 0; itemCount <= c_maxSetItems-builtComboItemCount; itemCount++ {
			next := bonusWithCount{addSet, itemCount}
			progress := util_collection.CopyAndAppend(built, next)
			bon.makeCombosRecur(setData[1:], progress, builtComboItemCount+itemCount, bonusCombos)
		}
	}
}

// logical AND between exact count vars
func (bon *gearBonusComboHandler) buildComboActivatingVar(combo *bonusCombo) {
	// TODO combining with required lookup could simplify

	if len(combo.condition) > 1 {
		comboActiveBool := &columnInfo{entryType: entry_combo_active, combo: combo}
		comboActiveBool.columnIndex = bon.build.CreateColumnBool(comboActiveBool)
		combo.activatingVar = comboActiveBool

		buildAnd := util_highs.ConstraintAndBuilder{}
		buildAnd.SetOutput(comboActiveBool.columnIndex)
		for _, setAndCount := range combo.condition {
			setInfo := setAndCount.setInfo
			count := setAndCount.count
			specificExactBool := setInfo.setExactCountVars[count]
			buildAnd.AddInput(specificExactBool.columnIndex)
		}
		buildAnd.Build(bon.build)
	} else if len(combo.condition) == 1 {
		setAndCount := combo.condition[0]
		setInfo := setAndCount.setInfo
		count := setAndCount.count
		specificExactBool := setInfo.setExactCountVars[count]
		combo.activatingVar = specificExactBool
	} else {
		panic("empty condition doesn't make sense")
	}
}

func (bon *gearBonusComboHandler) addSetItemsCountExactVariables(setIndex solve_highs_types.SetBonusIndex, countSetItems *columnInfo) [c_setItemsCounts]*columnInfo {
	exactCountVars := [c_setItemsCounts]*columnInfo{}

	// compare total number of items previous computed into this constraint
	compareRow := util_highs.ConstraintRow{Debug: "setItemsCompareRow"}
	compareRow.Add(countSetItems.columnIndex, -1)

	// constraint so only one of these flags gets set
	singleFlagOnly := util_highs.ConstraintRow{Debug: "setItemsSingleFlagOnly"}

	// make a bool for each possible count in range 0..5
	for itemCount := 0; itemCount <= c_maxSetItems; itemCount++ {
		boolColumn := &columnInfo{entryType: entry_set_exact_count, setIndex: setIndex, itemCount: itemCount}
		boolColumn.columnIndex = bon.build.CreateColumnBool(boolColumn)

		// should activate this flag which will match the total count
		compareRow.Add(boolColumn.columnIndex, float64(itemCount))

		// but only one flag at a time
		singleFlagOnly.Add(boolColumn.columnIndex, 1)

		exactCountVars[itemCount] = boolColumn
	}

	compareRow.Build(bon.build, 0, 0)     // equal
	singleFlagOnly.Build(bon.build, 1, 1) // sum of flags should be just one, should pull the zero flag up if no other set

	return exactCountVars
}

func (bon *gearBonusComboHandler) addSetNeededCounts(setBonusRequired []solve_highs_types.SetBonusRequiredCounts, countMode bonus_set.ItemCountsRequiredMode) {
	bonusData := bon._remove_bonusData

	if len(setBonusRequired) > 0 {
		if len(bonusData) == 0 {
			panic("no bonusData to use for addSetNeededCounts")
		} else if len(bonusData) == 1 && len(setBonusRequired) == 1 && len(setBonusRequired[0]) == 1 {
			setCountCol := bonusData[0].setTotalCountVar
			needCount := setBonusRequired[0][0]

			rowSetCountRequired := util_highs.ConstraintRow{Debug: "rowSetCountRequired"}
			rowSetCountRequired.Add(setCountCol.columnIndex, 1)
			switch countMode {
			case bonus_set.CountMode_Exact:
				rowSetCountRequired.Build(bon.build, float64(needCount), float64(needCount))
			case bonus_set.CountMode_Minimum:
				rowSetCountRequired.Build(bon.build, float64(needCount), util_highs.InfPos())
			case bonus_set.CountMode_AllowPlusOne:
				rowSetCountRequired.Build(bon.build, float64(needCount), float64(needCount+1))
			default:
				panic("unknown type")
			}
		} else {
			oneOfTheseOptions := util_highs.ConstraintRow{}

			for _, option := range setBonusRequired {
				optionParts := util_highs.ConstraintAndBuilder{}

				for setIndex, needCount := range option {
					setInfo := bonusData[setIndex]
					setCountCol := setInfo.setTotalCountVar

					var inRange util_highs.ColumnIndex
					switch countMode {
					case bonus_set.CountMode_Exact:
						inRange = bon.build.CreateColumnBool(nil)
						bon.build.ColumnIsEqualConstant(setCountCol.columnIndex, inRange, float64(needCount), 10, 1.0)
					case bonus_set.CountMode_Minimum:
						inRange = bon.build.ColumnIsGreaterOrEqualThanConstant(setCountCol.columnIndex, float64(needCount), 10, 1.0)
					case bonus_set.CountMode_AllowPlusOne:
						inRange = bon.build.ColumnIsBetweenConstants(setCountCol.columnIndex, float64(needCount), float64(needCount+1), 10, 1.0)
					default:
						panic("unknown type")
					}
					optionParts.AddInput(inRange)
				}

				optionActive := bon.build.CreateColumnBool(util_highs.DebugText("SetBonusRequired option"))
				optionParts.SetOutput(optionActive)
				optionParts.Build(bon.build)

				oneOfTheseOptions.Add(optionActive, 1)
			}

			oneOfTheseOptions.Build(bon.build, 1, util_highs.InfPos())
		}
	}
}

func (bon *gearBonusComboHandler) checkActiveCombo(solution util_highs.ISolution, solvableItemSet *items.SolvableItemSet, model *solve_highs_types.SolverModel) {
	bonusCombos := bon._remove_bonusCombos

	if bonusCombos.Size() > 0 {
		var activeCombo *bonusCombo

		for combo := range bonusCombos.SeqValuePointers() {
			if solution.ValueIsOne(combo.activatingVar.columnIndex) {
				if activeCombo == nil {
					activeCombo = combo
				} else {
					panic("multiple combos active")
				}
			}
		}

		if activeCombo != nil {
			for _, entry := range activeCombo.condition {
				countItems := model.SetBonus.CountItems[entry.setInfo.setIndex]
				actualCount := countItems(solvableItemSet.Items())
				if actualCount != uint8(entry.count) {
					panic("number of items not what solver returned")
				}
			}
		} else {
			panic("no combos active")
		}
	}
}
