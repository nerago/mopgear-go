package utilhighs

import "slices"

func ContraintIfBoolCopyValueElseZero(input *InputBuilder, boolSwitchVar, sourceVar, targetVar ColumnIndex, rangeLow, rangeHigh float64) {
	// based on https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d

	valueHigh := ConstraintRowBuild{Debug: "ContraintIfBoolCopyValueElseZero_ValueHigh"}
	valueHigh.Add(targetVar, -1)
	valueHigh.Add(sourceVar, 1)
	valueHigh.Add(boolSwitchVar, rangeHigh)
	valueHigh.Finish(input, C_MinusInf, rangeHigh)

	valueLow := ConstraintRowBuild{Debug: "ContraintIfBoolCopyValueElseZero_ValueLow"}
	valueLow.Add(targetVar, 1)
	valueLow.Add(sourceVar, -1)
	valueLow.Add(boolSwitchVar, -rangeLow)
	valueLow.Finish(input, C_MinusInf, -rangeLow)

	zeroHigh := ConstraintRowBuild{Debug: "ContraintIfBoolCopyValueElseZero_ZeroHigh"}
	zeroHigh.Add(targetVar, 1)
	zeroHigh.Add(boolSwitchVar, -rangeHigh)
	zeroHigh.Finish(input, C_MinusInf, 0)

	zeroLow := ConstraintRowBuild{Debug: "ContraintIfBoolCopyValueElseZero_ZeroLow"}
	zeroLow.Add(targetVar, -1)
	zeroLow.Add(boolSwitchVar, rangeLow)
	zeroLow.Finish(input, C_MinusInf, 0)
}

func constraintNotBool(input *InputBuilder, sourceVar, targetVar ColumnIndex) {
	not := ConstraintRowBuild{Debug: "constraintNotBool"}
	not.Add(sourceVar, 1)
	not.Add(targetVar, 1)
	not.Finish(input, 1, 1)
}

// https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d
type ContraintAndBuilder struct {
	outputVar ColumnIndex
	inputVars []ColumnIndex
}

func (build *ContraintAndBuilder) SetOutput(column ColumnIndex) {
	build.outputVar = column
}

func (build *ContraintAndBuilder) AddInput(column ColumnIndex) {
	if !slices.Contains(build.inputVars, column) {
		build.inputVars = append(build.inputVars, column)
	}
}

func (build *ContraintAndBuilder) FinishAndApply(input *InputBuilder) {
	sumRow := ConstraintRowBuild{Debug: "ContraintAndBuilder_sumRow"}
	sumRow.Add(build.outputVar, -1)

	for _, inputVar := range build.inputVars {
		sumRow.Add(inputVar, 1)

		pullDown := ConstraintRowBuild{Debug: "ContraintAndBuilder_pullDown"}
		pullDown.Add(inputVar, -1)
		pullDown.Add(build.outputVar, 1)
		pullDown.Finish(input, C_MinusInf, 0)
	}

	targetNum := len(build.inputVars) - 1
	sumRow.Finish(input, C_MinusInf, float64(targetNum))
}

type ConstraintOrBuilder struct {
	outputVar ColumnIndex
	inputVars []ColumnIndex
}

func (build *ConstraintOrBuilder) SetOutput(column ColumnIndex) {
	build.outputVar = column
}

func (build *ConstraintOrBuilder) AddInput(column ColumnIndex) {
	if !slices.Contains(build.inputVars, column) {
		build.inputVars = append(build.inputVars, column)
	}
}

func (build *ConstraintOrBuilder) FinishAndApply(input *InputBuilder) {
	zeroIfNone := ConstraintRowBuild{Debug: "ConstraintOrBuilder"}
	zeroIfNone.Add(build.outputVar, -1)

	for _, inputVar := range build.inputVars {
		zeroIfNone.Add(inputVar, 1)

		pullUp := ConstraintRowBuild{Debug: "ConstraintOrBuilder"}
		pullUp.Add(inputVar, -1)
		pullUp.Add(build.outputVar, 1)
		pullUp.Finish(input, 0, 1)
	}

	// maxNum := len(build.inputVars) - 1
	// sumRow.finish(input, 0, float64(maxNum))
	zeroIfNone.Finish(input, 0, C_PlusInf)
}

func ConstraintNot(input *InputBuilder, inputVar, outputVar ColumnIndex) {
	row := ConstraintRowBuild{Debug: "Not"}
	row.Add(inputVar, 1)
	row.Add(outputVar, 1)
	row.Finish(input, 1, 1)
}

func NotAsColumn(input *InputBuilder, inputVar ColumnIndex) ColumnIndex {
	outputVar := input.CreateColumnBool(nil)
	ConstraintNot(input, inputVar, outputVar)
	return outputVar
}

func AbsoluteValue(input *InputBuilder, inputVar, outputVar ColumnIndex) {
	negative := ConstraintRowBuild{Debug: "AbsoluteValueNegative"}
	negative.Add(inputVar, 1)
	negative.Add(outputVar, 1)
	negative.Finish(input, 0, C_PlusInf)

	positive := ConstraintRowBuild{Debug: "AbsoluteValuePositive"}
	positive.Add(inputVar, 1)
	positive.Add(outputVar, -1)
	positive.Finish(input, C_MinusInf, 0)
}

// untested but should be ok in principle
func AbsoluteValueFromDiffTwoVars(input *InputBuilder, inputOneVar ColumnIndex, inputOneCoefficient float64, inputTwoVar ColumnIndex, inputTwoCoefficient float64, outputVar ColumnIndex, debug string) {
	negative := ConstraintRowBuild{Debug: debug + " AbsoluteValueNegative"}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(inputTwoVar, -inputTwoCoefficient)
	negative.Add(outputVar, 1)
	negative.Finish(input, 0, C_PlusInf)

	positive := ConstraintRowBuild{Debug: debug + " AbsoluteValuePositive"}
	positive.Add(inputOneVar, inputOneCoefficient)
	positive.Add(inputTwoVar, -inputTwoCoefficient)
	positive.Add(outputVar, -1)
	positive.Finish(input, C_MinusInf, 0)
}

func AbsoluteValueFromDiffOneToConst(input *InputBuilder, inputOneVar ColumnIndex, inputOneCoefficient float64, constCompare float64, outputVar ColumnIndex, debug string) {
	negative := ConstraintRowBuild{Debug: debug + " AbsoluteValueNegative"}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(outputVar, 1)
	negative.Finish(input, constCompare, C_PlusInf)

	positive := ConstraintRowBuild{Debug: debug + " AbsoluteValuePositive"}
	positive.Add(inputOneVar, inputOneCoefficient)
	positive.Add(outputVar, -1)
	positive.Finish(input, C_MinusInf, constCompare)
}

func AbsoluteValueFromDiffTwoVarsWithOffset(input *InputBuilder, inputOneVar ColumnIndex, inputOneCoefficient float64, inputTwoVar ColumnIndex, inputTwoCoefficient float64, outputVar ColumnIndex, offset float64, debug string) {
	negative := ConstraintRowBuild{Debug: debug + " AbsoluteValueNegative"}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(inputTwoVar, -inputTwoCoefficient)
	negative.Add(outputVar, 1)
	negative.Finish(input, offset, C_PlusInf)

	positive := ConstraintRowBuild{Debug: debug + " AbsoluteValuePositive"}
	positive.Add(inputOneVar, inputOneCoefficient)
	positive.Add(inputTwoVar, -inputTwoCoefficient)
	positive.Add(outputVar, -1)
	positive.Finish(input, C_MinusInf, offset)
}

func AbsoluteValue_WithToggle(input *InputBuilder, inputVar, outputVar, toggleVar ColumnIndex, rangeHigh float64) {
	setIfNegative := ConstraintRowBuild{}
	setIfNegative.Add(inputVar, 1)
	setIfNegative.Add(outputVar, 1)
	setIfNegative.Add(toggleVar, -rangeHigh)
	setIfNegative.Finish(input, -rangeHigh, C_PlusInf)

	setIfPositive := ConstraintRowBuild{}
	setIfPositive.Add(inputVar, 1)
	setIfPositive.Add(outputVar, -1)
	setIfPositive.Add(toggleVar, rangeHigh)
	setIfPositive.Finish(input, C_MinusInf, rangeHigh)
}

// much the same as absolute value logic, just bool optimised
// similarly needs output variable under minimisation pressure
func IsXor(input *InputBuilder, boolOne ColumnIndex, boolTwo ColumnIndex, output ColumnIndex) {
	negative := ConstraintRowBuild{Debug: "Xor"}
	negative.Add(boolOne, 1)
	negative.Add(boolTwo, -1)
	negative.Add(output, 1)
	negative.Finish(input, 0, 2)

	positive := ConstraintRowBuild{Debug: "Xor"}
	positive.Add(boolOne, 1)
	positive.Add(boolTwo, -1)
	positive.Add(output, -1)
	positive.Finish(input, -2, 0)
}

func ConstantIsGreaterOrEqualColumn(input *InputBuilder, minimumColumn ColumnIndex, checkValue float64, rangeHigh float64) ColumnIndex {
	isOverMinimum := input.CreateColumnBool(DebugString{Text: "isOverMinimum"})

	// ORIGINAL
	// if overmin:      1*range + min <= range + stat  ->>  min <= stat
	// if !overmin:     0*range + min <= range + stat  ->>  min is free
	// if stat > min:   x*range + min <= range + stat  ->>  x*range <= range + stat - min  ->>  x*range <= range + small_positive   ->>   x = 0 or 1
	// if stat < min:   x*range + min <= range + stat  ->>  x*range <= range + stat - min  ->>  x*range <= range + small_negative   ->>   x = 0
	// if stat == min:  x*range + min <= range + stat  ->>  x*range <= range   ->>   x = 0 or 1
	checkIsOverMin := ConstraintRowBuild{Debug: "checkIsOverMin"}
	checkIsOverMin.Add(isOverMinimum, rangeHigh)
	checkIsOverMin.Add(minimumColumn, 1)
	checkIsOverMin.Finish(input, C_MinusInf, rangeHigh+checkValue)

	//   min - stat + x.range >= 0   ->>   min + x.range >= stat
	// if stat > min  ->>  min - stat + x.range >= 0  ->>  small_negative + x.range >= 0  ->>  x=1
	// if stat < min  ->>  min - stat + x.range >= 0  ->>  small_positive + x.range >= 0  ->>  x=0 or 1
	// if stat == min  ->> min - stat + x.range >= 1 ->>  x.range >= 1  ->> x=1
	// if overmin     ->>  min - stat + 1.range >= 0  ->>  min >= stat - range   ->>   min is free
	// if !overmin    ->>  min - stat + 0.range >= 0  ->>  min >= stat    ->>   min is free
	// modifiying with the plus one for the equals case, the rest of the math should hold ok
	setIfOverMin := ConstraintRowBuild{Debug: "setIfOverMin"}
	setIfOverMin.Add(minimumColumn, 1)
	setIfOverMin.Add(isOverMinimum, rangeHigh)
	setIfOverMin.Finish(input, checkValue+1, C_PlusInf)

	return isOverMinimum
}

func ConstantIsLessOrEqualColumn(input *InputBuilder, maximumColumn ColumnIndex, checkValue float64, rangeHigh float64) ColumnIndex {
	isUnderMaximum := input.CreateColumnBool(DebugString{Text: "isUnderMaximum"})

	// if undermax:    1*stat - max <= 0     ->>      stat <= max
	// if !undermax:   0*stat - max <= 0     ->>      max >= 0, (max free)
	// if stat <= max   ->>   x.stat - max <= 0    ->>     x.stat <= max   (x is free)
	// if stat > max    ->>   x.stat - max <= 0    ->>     x=0
	checkIsUnderMax := ConstraintRowBuild{Debug: "checkIsUnderMax"}
	checkIsUnderMax.Add(isUnderMaximum, checkValue)
	checkIsUnderMax.Add(maximumColumn, -1)
	checkIsUnderMax.Finish(input, C_MinusInf, 0)

	//    max - stat - x.range <= 0    ->>     max - x.range <= stat
	// if stat < max    ->>   max - stat - x.range <= 0   ->>   small_positive - x.range <= 0   ->>   small_positive <= x.range   ->>   x=1
	// if stat > max    ->>   max - stat - x.range <= 0   ->>   small_negative - x.range <= 0   ->>   small_negative <= x.range   ->>   x=0 or 1
	// if stat == max   ->>   max - stat - x.range <= 0   ->>   0 - x.range <= 0   ->>   0 <= x.range   ->> x=0 or 1
	// if undermax      ->>   max - stat - 1.range <= 0   ->>   max <= 1.range + stat  ->>  max is free
	// if !undermax     ->>   max - stat - 0.range <= 0   ->>   max - stat <= 0   ->>   max <= stat
	setIfUnderMax := ConstraintRowBuild{Debug: "setIfUnderMax"}
	setIfUnderMax.Add(maximumColumn, 1)
	setIfUnderMax.Add(isUnderMaximum, -rangeHigh)
	setIfUnderMax.Finish(input, C_MinusInf, checkValue)

	return isUnderMaximum
}

func ConstantIsBetweenColumns(input *InputBuilder, minimumColumn, maximumColumn, targetBoolColumn ColumnIndex, checkValue float64, rangeHigh float64) {
	isOverMinimum := ConstantIsGreaterOrEqualColumn(input, minimumColumn, checkValue, rangeHigh)
	isUnderMaximum := ConstantIsLessOrEqualColumn(input, maximumColumn, checkValue, rangeHigh)

	and := ContraintAndBuilder{}
	and.AddInput(isOverMinimum)
	and.AddInput(isUnderMaximum)
	and.SetOutput(targetBoolColumn)
	and.FinishAndApply(input)
}

func ColumnIsNotBetweenConstants(input *InputBuilder, checkColumn ColumnIndex, lo, hi float64, rangeHigh float64) {
	isUnderMin := input.CreateColumnBool(DebugString{Text: "isUnderMin"})
	isOverMax := input.CreateColumnBool(DebugString{Text: "isOverMax"})

	// lo <= check + x*range <= range+lo
	// if undermin: lo <= check + range <= range+lo     ->>  lo <= check + range (for sure)
	//                                                  ->>  check <= lo
	// if !undermin: lo <= check + x*range <= range+lo  ->>  lo <= check
	//                                                       check <= range+lo (for sure)
	// if check < lo: lo <= check + x*range <= range+lo ->>  lo <= check + x*range        ->> lo - check <= x*range          ->>  small_positive <= x*range          ->>  x=1
	//                                                       check + x*range <= range+lo  ->> check - lo + x*range <= range  ->>  small_negative + x*range <= range  ->>  x=0 or 1
	// if check > lo: lo <= check + x*range <= range+lo ->>  lo <= check + x*range        ->> lo - check <= x*range          ->>  small_negative <= x*range          ->>  x=0 or 1
	//                                                       check + x*range <= range+lo  ->> check - lo + x*range <= range  ->>  small_positive + x*range <= range  ->>  x=0
	underMin := ConstraintRowBuild{Debug: "setIfOverMin"}
	underMin.Add(checkColumn, 1)
	underMin.Add(isUnderMin, rangeHigh)
	underMin.Finish(input, lo, lo+rangeHigh)

	// hi-range <= check - x*range <= hi
	// if overmax: hi-range <= check - x*range <= hi  ->>  hi-range <= check - range <= hi  ->>  hi-range <= check - range  ->>  hi <= check
	//                                                                                      ->>  check - range <= hi        ->>  whatever
	// if !overmax: hi-range <= check - x*range <= hi  ->>  hi-range <= check <= hi  ->>  hi-range <= check  ->>  whatever
	//                                                                               ->>  check <= hi
	// if check < hi: hi-range <= check - x*range <= hi  ->>  hi-range <= check - x*range  ->>  hi-check <= range - x*range ->>  small_positive <= range - x*range  ->>  x=0
	//                                                        check - x*range <= hi        ->>  check - hi <= x*range       ->>  small_negative <= x*range          ->>  x=0 or 1
	// if check > hi: hi-range <= check - x*range <= hi  ->>  hi-range <= check - x*range  ->>  hi-check <= range - x*range ->>  small_negative <= range - x*range  ->>  x=0 or 1
	//                                                        check - x*range <= hi        ->>  check - hi <= x*range       ->>  small_positive <= x*range          ->>  x=1
	overMax := ConstraintRowBuild{Debug: "setIfUnderMax"}
	overMax.Add(checkColumn, 1)
	overMax.Add(isOverMax, -rangeHigh)
	overMax.Finish(input, hi-rangeHigh, hi)

	or := ConstraintRowBuild{}
	or.Add(isUnderMin, 1)
	or.Add(isOverMax, 1)
	or.Finish(input, 1, 1)
}

func ColumnIsGreaterOrEqualColumn(input *InputBuilder, thresholdLowColumn, checkHighColumn, boolIsGreater ColumnIndex, rangeHigh float64) {
	// -range <= check - thresh - x*range <= 0
	// if check>thresh  ->>  -range <= small_positive - x*range <= 0   ->>  -range <= small_positive - x*range  ->> x=0 or 1
	//                                                                      small_positive - x*range <= 0       ->> x=1
	// if check<thresh  ->>  -range <= small_negative - x*range <= 0   ->>  -range <= small_negative - x*range  ->> x=0
	//                                                                      small_negative - x*range <= 0       ->> x=0 or 1
	// if bool   ->>  -range <= check - thresh - range <= 0  ->>  -range <= check - thresh - range  ->>  0 <= check - thresh  ->>  thresh <= check
	//                                                            check - thresh - range <= 0       ->> whatever
	// if !bool  ->>  -range <= check - thresh <= 0  ->>  -range <= check - thresh  ->>  whatever
	//                                                    check - thresh <= 0       ->>  check <= thresh
	isGreaterEqual := ConstraintRowBuild{Debug: "isGreaterEqual"}
	isGreaterEqual.Add(thresholdLowColumn, -1)
	isGreaterEqual.Add(checkHighColumn, 1)
	isGreaterEqual.Add(boolIsGreater, -rangeHigh)
	isGreaterEqual.Finish(input, -rangeHigh, 0)
}
