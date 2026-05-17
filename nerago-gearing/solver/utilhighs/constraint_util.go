package utilhighs

import "github.com/bartolsthoorn/gohighs/highs"

func ContraintIfBoolCopyValueElseZero(input *InputBuilder, boolSwitchVar, sourceVar, targetVar ColumnIndex, rangeLow, rangeHigh float64) {
	// based on https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d

	valueHigh := ConstraintRowBuild{}
	valueHigh.Add(targetVar, -1)
	valueHigh.Add(sourceVar, 1)
	valueHigh.Add(boolSwitchVar, rangeHigh)
	valueHigh.Finish(input, C_MinusInf, rangeHigh)

	valueLow := ConstraintRowBuild{}
	valueLow.Add(targetVar, 1)
	valueLow.Add(sourceVar, -1)
	valueLow.Add(boolSwitchVar, -rangeLow)
	valueLow.Finish(input, C_MinusInf, -rangeLow)

	zeroHigh := ConstraintRowBuild{}
	zeroHigh.Add(targetVar, 1)
	zeroHigh.Add(boolSwitchVar, -rangeHigh)
	zeroHigh.Finish(input, C_MinusInf, 0)

	zeroLow := ConstraintRowBuild{}
	zeroLow.Add(targetVar, -1)
	zeroLow.Add(boolSwitchVar, rangeLow)
	zeroLow.Finish(input, C_MinusInf, 0)
}

func constraintNotBool(input *InputBuilder, sourceVar, targetVar ColumnIndex) {
	not := ConstraintRowBuild{}
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
	build.inputVars = append(build.inputVars, column)
}

func (build *ContraintAndBuilder) FinishAndApply(input *InputBuilder) {
	sumRow := ConstraintRowBuild{}
	sumRow.Add(build.outputVar, -1)

	for _, inputVar := range build.inputVars {
		sumRow.Add(inputVar, 1)

		pullDown := ConstraintRowBuild{}
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
	build.inputVars = append(build.inputVars, column)
}

func (build *ConstraintOrBuilder) FinishAndApply(input *InputBuilder) {
	zeroIfNone := ConstraintRowBuild{}
	zeroIfNone.Add(build.outputVar, -1)

	for _, inputVar := range build.inputVars {
		zeroIfNone.Add(inputVar, 1)

		pullUp := ConstraintRowBuild{}
		pullUp.Add(inputVar, -1)
		pullUp.Add(build.outputVar, 1)
		pullUp.Finish(input, 0, 1)
	}

	// maxNum := len(build.inputVars) - 1
	// sumRow.finish(input, 0, float64(maxNum))
	zeroIfNone.Finish(input, 0, C_PlusInf)
}

func absoluteValue_messy(input *InputBuilder, inputVar, outputVar ColumnIndex, rangeHigh float64) {
	// inputVar + outputVar = 0   OR   inputVar - outputVar = 0
	// diff1 = inputVar
	// diff2 = -inputVar

	// H = 10, M = 10
	// 0 <= inputVar + M*b1 <= H
	// if inputVar==0 then b1=0/1
	// if inputVar==5 then b1=0
	// if inputVar==10 then b1=0
	// if inputVar==-5 then b1=1
	// if inputVar==-10 then b1=1
	isNegativeCol := input.CreateColumnBool()
	negativeCheck := ConstraintRowBuild{}
	negativeCheck.Add(inputVar, 1)
	negativeCheck.Add(isNegativeCol, rangeHigh)
	negativeCheck.Finish(input, 0, rangeHigh)

	isZeroOrPositiveCol := input.CreateColumnBool()
	constraintNotBool(input, isNegativeCol, isZeroOrPositiveCol)

	inputNegated := input.CreateColumnGeneral(highs.Continuous, -rangeHigh, rangeHigh) // would be nice to get passed more info
	negateCalc := ConstraintRowBuild{}
	negateCalc.Add(inputVar, 1)
	negateCalc.Add(inputNegated, -1)
	negateCalc.Finish(input, 0, 0)

	ContraintIfBoolCopyValueElseZero(input, isNegativeCol, inputNegated, outputVar, 0, rangeHigh)
	ContraintIfBoolCopyValueElseZero(input, isZeroOrPositiveCol, inputVar, outputVar, 0, rangeHigh)
}

func absoluteValue(input *InputBuilder, inputVar, outputVar ColumnIndex, rangeHigh float64) {
	isNegativeCol := input.CreateColumnBool()
	negativeCheck := ConstraintRowBuild{}
	negativeCheck.Add(inputVar, 1)
	negativeCheck.Add(isNegativeCol, rangeHigh)
	negativeCheck.Finish(input, 0, rangeHigh)

	setIfNegative := ConstraintRowBuild{}
	setIfNegative.Add(inputVar, 1)
	setIfNegative.Add(outputVar, 1)
	setIfNegative.Add(isNegativeCol, rangeHigh)
	setIfNegative.Finish(input, 0, rangeHigh)

	setIfPositive := ConstraintRowBuild{}
	setIfPositive.Add(inputVar, -1)
	setIfPositive.Add(outputVar, 1)
	setIfPositive.Add(isNegativeCol, -rangeHigh)
	setIfPositive.Finish(input, -rangeHigh, 0)
}

func AbsoluteValue2(input *InputBuilder, inputVar, outputVar ColumnIndex, rangeHigh float64) {
	// isNegativeCol := input.createColumnBool()
	// negativeCheck := constraintRowBuild{}
	// negativeCheck.add(inputVar, 1)
	// negativeCheck.add(isNegativeCol, rangeHigh)
	// negativeCheck.finish(input, 0, rangeHigh)

	setIfNegative := ConstraintRowBuild{}
	setIfNegative.Add(inputVar, 1)
	setIfNegative.Add(outputVar, 1)
	setIfNegative.Finish(input, 0, C_PlusInf)

	setIfPositive := ConstraintRowBuild{}
	setIfPositive.Add(inputVar, 1)
	setIfPositive.Add(outputVar, -1)
	setIfPositive.Finish(input, C_MinusInf, 0)
}
