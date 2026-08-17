package solve_highs

import (
	"paladin_gearing_go/gear_model/bonus_set"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
)

type gearBonusComboHandler struct {
	build *util_highs.LinearBuilder
	//bonusCombos []bonusCombo
	//bonusData   []bonusInfo
}

func (bon *gearBonusComboHandler) prepareEnabledSets(model *solve_highs_types.SolverModel, build *util_highs.LinearBuilder, itemSetup *gearItemSetupShared) {
	// constrain: exact item count in each active set
	if model.SetBonus.TotalCount > 0 {
		bon.bonusData = make([]bonusInfo, model.SetBonus.TotalCount)
		for setIndex := range model.SetBonus.TotalCount {
			info := bonusInfo{
				setIndex:            solve_highs_types.SetBonusIndex(setIndex),
				setCountItems:       model.SetBonus.CountItems[setIndex],
				setMultipliers:      model.SetBonus.MultipliersFlat[setIndex],
				setMultipliersBySim: model.SetBonus.MultipliersBySim[setIndex],
			}
			bon.addSetItemCountVariable(&info, build, itemSetup)
			bon.addSetItemsCountExactVariables(&info, build) // might not need
			bon.bonusData[setIndex] = info
		}
	}
}

func (bon *gearBonusComboHandler) multiplyByActiveCombo(combinedRatingVar *columnInfo, outputVar *columnInfo, rangeHigh float64, getMultiplier func(*bonusCombo) float64) {
	if len(bon.bonusData) > 0 {
		for combo := range util_collection.ForPointer(bon.bonusCombos) {
			bon.buildSetMultipliedOutput(combo, combinedRatingVar, outputVar, rangeHigh, getMultiplier)
		}
	} else {
		bon.buildSimpleNoSetsOutput(combinedRatingVar, outputVar)
	}
}

func (bon *gearBonusComboHandler) buildSimpleNoSetsOutput(inputVar *columnInfo, outputVar *columnInfo) {
	// just copy initial rating sum into final if no sets
	bon.build.ConstraintCopy(
		inputVar.columnIndex, 1,
		outputVar.columnIndex,
		"nullComboCopy",
	)
}

func (bon *gearBonusComboHandler) buildSetMultipliedOutput(combo *bonusCombo, inputVar *columnInfo, outputVar *columnInfo, rangeHigh float64, getMultiplier func(*bonusCombo) float64) {
	activatingVar := combo.activatingVar
	bonusMultiplier := getMultiplier(combo)
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

func (bon *gearBonusComboHandler) prepareSetCombos() {
	if len(bon.bonusData) == 0 {
		return
	}

	bon.makeSetCombosRecur(bon.bonusData, nil, 0)

	for combo := range util_collection.ForPointer(bon.bonusCombos) {
		bon.buildCombinationActivatingVar(combo, build)
	}

	checkSingleCombo := util_highs.ConstraintRow{}
	for combo := range util_collection.ForPointer(bon.bonusCombos) {
		checkSingleCombo.Add(combo.activatingVar.columnIndex, 1)
	}
	checkSingleCombo.Build(bon.build, 1, 1)
}

func (bon *gearBonusComboHandler) makeSetCombosRecur(setData []bonusInfo, built []bonusWithCount, builtComboItemCount int) {
	if len(setData) == 0 || builtComboItemCount == c_maxSetItems {
		bon.bonusCombos = append(bon.bonusCombos, bonusCombo{condition: built})
	} else {
		addSet := &setData[0]
		for itemCount := 0; itemCount <= c_maxSetItems-builtComboItemCount; itemCount++ {
			next := bonusWithCount{addSet, itemCount}
			progress := util_collection.CopyAndAppend(built, next)
			bon.makeSetCombosRecur(setData[1:], progress, builtComboItemCount+itemCount)
		}
	}
}

func (bon *gearBonusComboHandler) addSetNeededCounts(setBonusRequired []solve_highs_types.SetBonusRequiredCounts, countMode bonus_set.ItemCountsRequiredMode) {
	if len(setBonusRequired) > 0 {
		if len(bon.bonusData) == 0 {
			panic("no bonusData to use for addSetNeededCounts")
		} else if len(bon.bonusData) == 1 && len(setBonusRequired) == 1 && len(setBonusRequired[0]) == 1 {
			setCountCol := bon.bonusData[0].setTotalCountVar
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
					setInfo := bon.bonusData[setIndex]
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

// logical AND between exact count vars
func (bon *gearBonusComboHandler) buildCombinationActivatingVar(combo *bonusCombo) {
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
}

func (bon *gearBonusComboHandler) addSetItemsCountExactVariables(info *bonusInfo) {
	// compare total number of items previous computed into this constraint
	compareRow := util_highs.ConstraintRow{Debug: "setItemsCompareRow"}
	compareRow.Add(info.setTotalCountVar.columnIndex, -1)

	// constraint so only one of these flags gets set
	singleFlagOnly := util_highs.ConstraintRow{Debug: "setItemsSingleFlagOnly"}

	// make a bool for each possible count in range 0..5
	for itemCount := 0; itemCount <= c_maxSetItems; itemCount++ {
		boolColumn := columnInfo{entryType: entry_set_exact_count, setIndex: info.setIndex, itemCount: itemCount}
		boolColumn.columnIndex = bon.build.CreateColumnBool(&boolColumn)

		// should activate this flag which will match the total count
		compareRow.Add(boolColumn.columnIndex, float64(itemCount))

		// but only one flag at a time
		singleFlagOnly.Add(boolColumn.columnIndex, 1)

		info.setExactCountVars[itemCount] = &boolColumn
	}

	compareRow.Build(bon.build, 0, 0)     // equal
	singleFlagOnly.Build(bon.build, 1, 1) // sum of flags should be just one, should pull the zero flag up if no other set
}

func (bon *gearBonusComboHandler) checkActiveCombo(solution util_highs.ISolution, solvableItemSet *items.SolvableItemSet) {
	if len(bon.bonusCombos) > 0 {
		var activeCombo *bonusCombo

		for _, combo := range bon.bonusCombos {
			if solution.ValueIsOne(combo.activatingVar.columnIndex) {
				if activeCombo == nil {
					activeCombo = &combo
				} else {
					panic("multiple combos active")
				}
			}
		}

		if activeCombo != nil {
			for _, entry := range activeCombo.condition {
				actualCount := entry.setInfo.setCountItems(solvableItemSet.Items())
				if actualCount != uint8(entry.count) {
					panic("number of items not what solver returned")
				}
			}
		} else {
			panic("no combos active")
		}
	}
}
