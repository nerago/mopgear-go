package utilhighs

import "slices"

func ContraintIfBoolCopyValueElseZero(build *LinearBuilder, boolSwitchVar, sourceVar, targetVar ColumnIndex, rangeLow, rangeHigh float64) {
	// based on https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d

	valueHigh := ConstraintRow{Debug: "ContraintIfBoolCopyValueElseZero_ValueHigh"}
	valueHigh.Add(targetVar, -1)
	valueHigh.Add(sourceVar, 1)
	valueHigh.Add(boolSwitchVar, rangeHigh)
	valueHigh.Build(build, C_MinusInf, rangeHigh)

	valueLow := ConstraintRow{Debug: "ContraintIfBoolCopyValueElseZero_ValueLow"}
	valueLow.Add(targetVar, 1)
	valueLow.Add(sourceVar, -1)
	valueLow.Add(boolSwitchVar, -rangeLow)
	valueLow.Build(build, C_MinusInf, -rangeLow)

	zeroHigh := ConstraintRow{Debug: "ContraintIfBoolCopyValueElseZero_ZeroHigh"}
	zeroHigh.Add(targetVar, 1)
	zeroHigh.Add(boolSwitchVar, -rangeHigh)
	zeroHigh.Build(build, C_MinusInf, 0)

	zeroLow := ConstraintRow{Debug: "ContraintIfBoolCopyValueElseZero_ZeroLow"}
	zeroLow.Add(targetVar, -1)
	zeroLow.Add(boolSwitchVar, rangeLow)
	zeroLow.Build(build, C_MinusInf, 0)
}

func ContraintIfBoolCopy(build *LinearBuilder, boolSwitchVar, sourceVar, targetVar ColumnIndex, rangeHigh float64) {
	valueHigh := ConstraintRow{Debug: "ContraintIfBoolCopy_ValueHigh"}
	valueHigh.Add(targetVar, -1)
	valueHigh.Add(sourceVar, 1)
	valueHigh.Add(boolSwitchVar, rangeHigh)
	valueHigh.Build(build, C_MinusInf, rangeHigh)

	valueLow := ConstraintRow{Debug: "ContraintIfBoolCopy_ValueLow"}
	valueLow.Add(targetVar, 1)
	valueLow.Add(sourceVar, -1)
	valueLow.Add(boolSwitchVar, rangeHigh)
	valueLow.Build(build, C_MinusInf, rangeHigh)
}

// https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d
type ContraintAndBuilder struct {
	outputVar ColumnIndex
	inputVars []ColumnIndex
}

func (and *ContraintAndBuilder) SetOutput(column ColumnIndex) {
	and.outputVar = column
}

func (and *ContraintAndBuilder) AddInput(column ColumnIndex) {
	if !slices.Contains(and.inputVars, column) {
		and.inputVars = append(and.inputVars, column)
	}
}

func (and *ContraintAndBuilder) Build(build *LinearBuilder) {
	if len(and.inputVars) == 0 {
		panic("no inputs")
	} else if len(and.inputVars) == 1 {
		copyRow := ConstraintRow{Debug: "ContraintAndBuilder_copy"}
		copyRow.Add(and.outputVar, -1)
		copyRow.Add(and.inputVars[0], 1)
		copyRow.Build(build, 0, 0)
	} else {
		sumRow := ConstraintRow{Debug: "ContraintAndBuilder_sumRow"}
		sumRow.Add(and.outputVar, -1)

		for _, inputVar := range and.inputVars {
			sumRow.Add(inputVar, 1)

			pullDown := ConstraintRow{Debug: "ContraintAndBuilder_pullDown"}
			pullDown.Add(inputVar, -1)
			pullDown.Add(and.outputVar, 1)
			pullDown.Build(build, C_MinusInf, 0)
		}

		targetNum := len(and.inputVars) - 1
		sumRow.Build(build, C_MinusInf, float64(targetNum))
	}
}

type ConstraintOrBuilder struct {
	outputVar ColumnIndex
	inputVars []ColumnIndex
}

func (or *ConstraintOrBuilder) SetOutput(column ColumnIndex) {
	or.outputVar = column
}

func (or *ConstraintOrBuilder) AddInput(column ColumnIndex) {
	if !slices.Contains(or.inputVars, column) {
		or.inputVars = append(or.inputVars, column)
	}
}

func (or *ConstraintOrBuilder) Build(build *LinearBuilder) {
	zeroIfNone := ConstraintRow{Debug: "ConstraintOrBuilder"}
	zeroIfNone.Add(or.outputVar, -1)

	for _, inputVar := range or.inputVars {
		zeroIfNone.Add(inputVar, 1)

		pullUp := ConstraintRow{Debug: "ConstraintOrBuilder"}
		pullUp.Add(inputVar, -1)
		pullUp.Add(or.outputVar, 1)
		pullUp.Build(build, 0, 1)
	}

	// maxNum := len(build.inputVars) - 1
	// sumRow.finish(input, 0, float64(maxNum))
	zeroIfNone.Build(build, 0, C_PlusInf)
}

func ConstraintNot(build *LinearBuilder, inputVar, outputVar ColumnIndex) {
	row := ConstraintRow{Debug: "Not"}
	row.Add(inputVar, 1)
	row.Add(outputVar, 1)
	row.Build(build, 1, 1)
}

func NotAsColumn(build *LinearBuilder, inputVar ColumnIndex) ColumnIndex {
	outputVar := build.CreateColumnBool(nil)
	ConstraintNot(build, inputVar, outputVar)
	return outputVar
}

func AbsoluteValue(build *LinearBuilder, inputVar, outputVar ColumnIndex) {
	negative := ConstraintRow{Debug: "AbsoluteValueNegative"}
	negative.Add(inputVar, 1)
	negative.Add(outputVar, 1)
	negative.Build(build, 0, C_PlusInf)

	positive := ConstraintRow{Debug: "AbsoluteValuePositive"}
	positive.Add(inputVar, 1)
	positive.Add(outputVar, -1)
	positive.Build(build, C_MinusInf, 0)
}

func AbsoluteValueFromDiffTwoVars(build *LinearBuilder, inputOneVar ColumnIndex, inputOneCoefficient float64, inputTwoVar ColumnIndex, inputTwoCoefficient float64, outputVar ColumnIndex, debug string) {
	negative := ConstraintRow{Debug: debug + " AbsoluteValueNegative"}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(inputTwoVar, -inputTwoCoefficient)
	negative.Add(outputVar, 1)
	negative.Build(build, 0, C_PlusInf)

	positive := ConstraintRow{Debug: debug + " AbsoluteValuePositive"}
	positive.Add(inputOneVar, inputOneCoefficient)
	positive.Add(inputTwoVar, -inputTwoCoefficient)
	positive.Add(outputVar, -1)
	positive.Build(build, C_MinusInf, 0)
}

func AbsoluteValueFromDiffOneToConst(build *LinearBuilder, inputOneVar ColumnIndex, inputOneCoefficient float64, constCompare float64, outputVar ColumnIndex, debug string) {
	negative := ConstraintRow{Debug: debug + " AbsoluteValueNegative"}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(outputVar, 1)
	negative.Build(build, constCompare, C_PlusInf)

	positive := ConstraintRow{Debug: debug + " AbsoluteValuePositive"}
	positive.Add(inputOneVar, inputOneCoefficient)
	positive.Add(outputVar, -1)
	positive.Build(build, C_MinusInf, constCompare)
}

// what i wanted it to be is ABS(one-two <=> offset)
func AbsoluteValueFromDiffTwoVarsWithOffset(build *LinearBuilder, inputOneVar ColumnIndex, inputOneCoefficient float64, inputTwoVar ColumnIndex, inputTwoCoefficient float64, outputVar ColumnIndex, offset float64, debug string) {
	negative := ConstraintRow{Debug: debug + " AbsoluteValueNegative"}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(inputTwoVar, -inputTwoCoefficient)
	negative.Add(outputVar, 1)
	negative.Build(build, offset, C_PlusInf)

	positive := ConstraintRow{Debug: debug + " AbsoluteValuePositive"}
	positive.Add(inputOneVar, inputOneCoefficient)
	positive.Add(inputTwoVar, -inputTwoCoefficient)
	positive.Add(outputVar, -1)
	positive.Build(build, C_MinusInf, offset)
}

func AbsoluteValue_WithToggle(build *LinearBuilder, inputVar, outputVar, toggleVar ColumnIndex, rangeHigh float64) {
	setIfNegative := ConstraintRow{}
	setIfNegative.Add(inputVar, 1)
	setIfNegative.Add(outputVar, 1)
	setIfNegative.Add(toggleVar, -rangeHigh)
	setIfNegative.Build(build, -rangeHigh, C_PlusInf)

	setIfPositive := ConstraintRow{}
	setIfPositive.Add(inputVar, 1)
	setIfPositive.Add(outputVar, -1)
	setIfPositive.Add(toggleVar, rangeHigh)
	setIfPositive.Build(build, C_MinusInf, rangeHigh)
}

// basic logic: output = one xor two
// however output is free when condition not met, should ideally put output under minimise pressure
// similar to absolute value, just intended for int vars
func IsXor(build *LinearBuilder, boolOne ColumnIndex, boolTwo ColumnIndex, output ColumnIndex) {
	negative := ConstraintRow{Debug: "Xor"}
	negative.Add(boolOne, 1)
	negative.Add(boolTwo, -1)
	negative.Add(output, 1)
	negative.Build(build, 0, 2)

	positive := ConstraintRow{Debug: "Xor"}
	positive.Add(boolOne, 1)
	positive.Add(boolTwo, -1)
	positive.Add(output, -1)
	positive.Build(build, -2, 0)
}

// rangeHigh should be bigger than any possible value
// equalDelta should be just big enough to make unequal (1.0 for ints)
// logic: colValue >= constValue
func ColumnIsGreaterOrEqualThanConstant(build *LinearBuilder, compareColumn ColumnIndex, constValue float64, rangeHigh float64, equalDelta float64) ColumnIndex {
	isGreaterEqual := build.CreateColumnBool(DebugString{Text: "isGreaterEqual"})

	set := ConstraintRow{Debug: "ColumnIsGreaterOrEqualThanConstant_set"}
	set.Add(compareColumn, 1)
	set.Add(isGreaterEqual, -rangeHigh)
	set.Build(build, C_MinusInf, constValue-equalDelta)

	confirm := ConstraintRow{Debug: "ColumnIsGreaterOrEqualThanConstant_confirm"}
	confirm.Add(isGreaterEqual, constValue)
	confirm.Add(compareColumn, -1)
	confirm.Build(build, C_MinusInf, 0)

	return isGreaterEqual
}

// rangeHigh should be bigger than any possible value
// equalDelta should be just big enough to make unequal (1.0 for ints)
// logic: colValue <= constValue
func ColumnIsLessOrEqualThanConstant(build *LinearBuilder, compareColumn ColumnIndex, constValue float64, rangeHigh float64, equalDelta float64) ColumnIndex {
	isLessEqual := build.CreateColumnBool(DebugString{Text: "isLessEqual"})

	set := ConstraintRow{Debug: "ColumnIsLessOrEqualThanConstant_set"}
	set.Add(compareColumn, 1)
	set.Add(isLessEqual, rangeHigh)
	set.Build(build, constValue+equalDelta, C_PlusInf)

	confirm := ConstraintRow{Debug: "ColumnIsLessOrEqualThanConstant_confirm"}
	confirm.Add(isLessEqual, rangeHigh)
	confirm.Add(compareColumn, 1)
	confirm.Build(build, C_MinusInf, rangeHigh+constValue)

	return isLessEqual
}

func ConstantIsBetweenColumns(build *LinearBuilder, minimumColumn, maximumColumn, targetBoolColumn ColumnIndex, constValue float64, rangeHigh float64, equalDelta float64) {
	isOverMinimum := ColumnIsLessOrEqualThanConstant(build, minimumColumn, constValue, rangeHigh, equalDelta)
	isUnderMaximum := ColumnIsGreaterOrEqualThanConstant(build, maximumColumn, constValue, rangeHigh, equalDelta)

	and := ContraintAndBuilder{}
	and.AddInput(isOverMinimum)
	and.AddInput(isUnderMaximum)
	and.SetOutput(targetBoolColumn)
	and.Build(build)

	// NOTE: this may not be needed and hurt performance for some callers, could have alternate version without
	// but generally makes sense for consistency of expectations
	checkSequence := ConstraintRow{}
	checkSequence.Add(maximumColumn, 1)
	checkSequence.Add(minimumColumn, -1)
	checkSequence.Build(build, 0, C_PlusInf)
}

func ColumnIsNotBetweenConstantsVerify(build *LinearBuilder, checkColumn ColumnIndex, lo, hi float64, rangeHigh float64) {
	if lo > hi {
		panic("backwards range")
	}

	isUnderMin := build.CreateColumnBool(DebugString{Text: "isUnderMin"})
	isOverMax := build.CreateColumnBool(DebugString{Text: "isOverMax"})

	// lo <= check + x*range <= range+lo
	// if undermin: lo <= check + range <= range+lo     ->>  lo <= check + range (for sure)
	//                                                  ->>  check <= lo
	// if !undermin: lo <= check + x*range <= range+lo  ->>  lo <= check
	//                                                       check <= range+lo (for sure)
	// if check < lo: lo <= check + x*range <= range+lo ->>  lo <= check + x*range        ->> lo - check <= x*range          ->>  small_positive <= x*range          ->>  x=1
	//                                                       check + x*range <= range+lo  ->> check - lo + x*range <= range  ->>  small_negative + x*range <= range  ->>  x=0 or 1
	// if check > lo: lo <= check + x*range <= range+lo ->>  lo <= check + x*range        ->> lo - check <= x*range          ->>  small_negative <= x*range          ->>  x=0 or 1
	//                                                       check + x*range <= range+lo  ->> check - lo + x*range <= range  ->>  small_positive + x*range <= range  ->>  x=0
	underMin := ConstraintRow{Debug: "setIfOverMin"}
	underMin.Add(checkColumn, 1)
	underMin.Add(isUnderMin, rangeHigh)
	underMin.Build(build, lo, lo+rangeHigh)

	// hi-range <= check - x*range <= hi
	// if overmax: hi-range <= check - x*range <= hi  ->>  hi-range <= check - range <= hi  ->>  hi-range <= check - range  ->>  hi <= check
	//                                                                                      ->>  check - range <= hi        ->>  whatever
	// if !overmax: hi-range <= check - x*range <= hi  ->>  hi-range <= check <= hi  ->>  hi-range <= check  ->>  whatever
	//                                                                               ->>  check <= hi
	// if check < hi: hi-range <= check - x*range <= hi  ->>  hi-range <= check - x*range  ->>  hi-check <= range - x*range ->>  small_positive <= range - x*range  ->>  x=0
	//                                                        check - x*range <= hi        ->>  check - hi <= x*range       ->>  small_negative <= x*range          ->>  x=0 or 1
	// if check > hi: hi-range <= check - x*range <= hi  ->>  hi-range <= check - x*range  ->>  hi-check <= range - x*range ->>  small_negative <= range - x*range  ->>  x=0 or 1
	//                                                        check - x*range <= hi        ->>  check - hi <= x*range       ->>  small_positive <= x*range          ->>  x=1
	overMax := ConstraintRow{Debug: "setIfUnderMax"}
	overMax.Add(checkColumn, 1)
	overMax.Add(isOverMax, -rangeHigh)
	overMax.Build(build, hi-rangeHigh, hi)

	or := ConstraintRow{}
	or.Add(isUnderMin, 1)
	or.Add(isOverMax, 1)
	or.Build(build, 1, 1)
}

// logic: leftSideCol < rightSideCol
func ColumnIsLessThanColumnEqualityFree(build *LinearBuilder, leftSideCol, rightSideCol, boolIsLess ColumnIndex, rangeHigh float64) {
	// -range <= check - thresh - x*range <= 0
	// if check>thresh  ->>  -range <= small_positive - x*range <= 0   ->>  -range <= small_positive - x*range  ->> x=0 or 1
	//                                                                      small_positive - x*range <= 0       ->> x=1
	// if check<thresh  ->>  -range <= small_negative - x*range <= 0   ->>  -range <= small_negative - x*range  ->> x=0
	//                                                                      small_negative - x*range <= 0       ->> x=0 or 1
	// if bool   ->>  -range <= check - thresh - range <= 0  ->>  -range <= check - thresh - range  ->>  0 <= check - thresh  ->>  thresh <= check
	//                                                            check - thresh - range <= 0       ->> whatever
	// if !bool  ->>  -range <= check - thresh <= 0  ->>  -range <= check - thresh  ->>  whatever
	//                                                    check - thresh <= 0       ->>  check <= thresh
	isLess := ConstraintRow{Debug: "ColumnIsLessThanColumnEqualityFree"}
	isLess.Add(leftSideCol, -1)
	isLess.Add(rightSideCol, 1)
	isLess.Add(boolIsLess, -rangeHigh)
	isLess.Build(build, -rangeHigh, 0)
}

// logic: leftSideCol > rightSideCol
func ColumnIsGreaterThanColumnEqualityFree(build *LinearBuilder, leftSideCol, rightSideCol, boolIsGreater ColumnIndex, rangeHigh float64) {
	ColumnIsLessThanColumnEqualityFree(build, rightSideCol, leftSideCol, boolIsGreater, rangeHigh)
}

// logic: leftSideCol >= rightSideCol
func ColumnIsGreaterOrEqualColumn(build *LinearBuilder, leftSideCol, rightSideCol, boolIsGreater ColumnIndex, rangeHigh float64, equalDelta float64) {
	set := ConstraintRow{Debug: "ColumnIsGreaterOrEqualColumn_set"}
	set.Add(leftSideCol, 1)
	set.Add(rightSideCol, -1)
	set.Add(boolIsGreater, -rangeHigh)
	set.Build(build, C_MinusInf, -equalDelta)

	confirm := ConstraintRow{Debug: "ColumnIsGreaterOrEqualColumn_confirm"}
	confirm.Add(leftSideCol, -1)
	confirm.Add(rightSideCol, 1)
	confirm.Add(boolIsGreater, rangeHigh)
	confirm.Build(build, C_MinusInf, rangeHigh)
}

// logic: leftSideCol <= rightSideCol
func ColumnIsLessOrEqualColumn(build *LinearBuilder, leftSideCol, rightSideCol, boolIsLess ColumnIndex, rangeHigh float64, equalDelta float64) {
	set := ConstraintRow{Debug: "ColumnIsLessOrEqualColumn_set"}
	set.Add(leftSideCol, 1)
	set.Add(rightSideCol, -1)
	set.Add(boolIsLess, rangeHigh)
	set.Build(build, equalDelta, C_PlusInf)

	confirm := ConstraintRow{Debug: "ColumnIsLessOrEqualColumn_confirm"}
	confirm.Add(leftSideCol, 1)
	confirm.Add(rightSideCol, -1)
	confirm.Add(boolIsLess, rangeHigh)
	confirm.Build(build, C_MinusInf, rangeHigh)
}
