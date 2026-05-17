package withhighs

import "github.com/bartolsthoorn/gohighs/highs"

func contraintIfBoolCopyValueElseZero(input *inputBuilder, boolSwitchVar, sourceVar, targetVar columnIndex, rangeLow, rangeHigh float64) {
	// based on https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d

	valueHigh := constraintRowBuild{}
	valueHigh.add(targetVar, -1)
	valueHigh.add(sourceVar, 1)
	valueHigh.add(boolSwitchVar, rangeHigh)
	valueHigh.finish(input, c_minusInf, rangeHigh)

	valueLow := constraintRowBuild{}
	valueLow.add(targetVar, 1)
	valueLow.add(sourceVar, -1)
	valueLow.add(boolSwitchVar, -rangeLow)
	valueLow.finish(input, c_minusInf, -rangeLow)

	zeroHigh := constraintRowBuild{}
	zeroHigh.add(targetVar, 1)
	zeroHigh.add(boolSwitchVar, -rangeHigh)
	zeroHigh.finish(input, c_minusInf, 0)

	zeroLow := constraintRowBuild{}
	zeroLow.add(targetVar, -1)
	zeroLow.add(boolSwitchVar, rangeLow)
	zeroLow.finish(input, c_minusInf, 0)
}

func constraintNotBool(input *inputBuilder, sourceVar, targetVar columnIndex) {
	not := constraintRowBuild{}
	not.add(sourceVar, 1)
	not.add(targetVar, 1)
	not.finish(input, 1, 1)
}

// https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d
type contraintAndBuilder struct {
	outputVar columnIndex
	inputVars []columnIndex
}

func (build *contraintAndBuilder) setOutput(column columnIndex) {
	build.outputVar = column
}

func (build *contraintAndBuilder) addInput(column columnIndex) {
	build.inputVars = append(build.inputVars, column)
}

func (build *contraintAndBuilder) finishAndApply(input *inputBuilder) {
	sumRow := constraintRowBuild{}
	sumRow.add(build.outputVar, -1)

	for _, inputVar := range build.inputVars {
		sumRow.add(inputVar, 1)

		pullDown := constraintRowBuild{}
		pullDown.add(inputVar, -1)
		pullDown.add(build.outputVar, 1)
		pullDown.finish(input, c_minusInf, 0)
	}

	targetNum := len(build.inputVars) - 1
	sumRow.finish(input, c_minusInf, float64(targetNum))
}

type contraintOrBuilder struct {
	outputVar columnIndex
	inputVars []columnIndex
}

func (build *contraintOrBuilder) setOutput(column columnIndex) {
	build.outputVar = column
}

func (build *contraintOrBuilder) addInput(column columnIndex) {
	build.inputVars = append(build.inputVars, column)
}

func (build *contraintOrBuilder) finishAndApply(input *inputBuilder) {
	zeroIfNone := constraintRowBuild{}
	zeroIfNone.add(build.outputVar, -1)

	for _, inputVar := range build.inputVars {
		zeroIfNone.add(inputVar, 1)

		pullUp := constraintRowBuild{}
		pullUp.add(inputVar, -1)
		pullUp.add(build.outputVar, 1)
		pullUp.finish(input, 0, 1)
	}

	// maxNum := len(build.inputVars) - 1
	// sumRow.finish(input, 0, float64(maxNum))
	zeroIfNone.finish(input, 0, c_plusInf) 
}

func absoluteValue_messy(input *inputBuilder, inputVar, outputVar columnIndex, rangeHigh float64) {
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
	isNegativeCol := input.createColumnBool()
	negativeCheck := constraintRowBuild{}
	negativeCheck.add(inputVar, 1)
	negativeCheck.add(isNegativeCol, rangeHigh)
	negativeCheck.finish(input, 0, rangeHigh)

	isZeroOrPositiveCol := input.createColumnBool()
	constraintNotBool(input, isNegativeCol, isZeroOrPositiveCol)

	inputNegated := input.createColumnGeneral(highs.Continuous, -rangeHigh, rangeHigh) // would be nice to get passed more info
	negateCalc := constraintRowBuild{}
	negateCalc.add(inputVar, 1)
	negateCalc.add(inputNegated, -1)
	negateCalc.finish(input, 0, 0)

	contraintIfBoolCopyValueElseZero(input, isNegativeCol, inputNegated, outputVar, 0, rangeHigh)
	contraintIfBoolCopyValueElseZero(input, isZeroOrPositiveCol, inputVar, outputVar, 0, rangeHigh)
}

func absoluteValue(input *inputBuilder, inputVar, outputVar columnIndex, rangeHigh float64) {
	isNegativeCol := input.createColumnBool()
	negativeCheck := constraintRowBuild{}
	negativeCheck.add(inputVar, 1)
	negativeCheck.add(isNegativeCol, rangeHigh)
	negativeCheck.finish(input, 0, rangeHigh)

	setIfNegative := constraintRowBuild{}
	setIfNegative.add(inputVar, 1)
	setIfNegative.add(outputVar, 1)
	setIfNegative.add(isNegativeCol, rangeHigh)
	setIfNegative.finish(input, 0, rangeHigh)

	setIfPositive := constraintRowBuild{}
	setIfPositive.add(inputVar, -1)
	setIfPositive.add(outputVar, 1)
	setIfPositive.add(isNegativeCol, -rangeHigh)
	setIfPositive.finish(input, -rangeHigh, 0)
}

func absoluteValue2(input *inputBuilder, inputVar, outputVar columnIndex, rangeHigh float64) {
	// isNegativeCol := input.createColumnBool()
	// negativeCheck := constraintRowBuild{}
	// negativeCheck.add(inputVar, 1)
	// negativeCheck.add(isNegativeCol, rangeHigh)
	// negativeCheck.finish(input, 0, rangeHigh)

	setIfNegative := constraintRowBuild{}
	setIfNegative.add(inputVar, 1)
	setIfNegative.add(outputVar, 1)
	setIfNegative.finish(input, 0, c_plusInf)

	setIfPositive := constraintRowBuild{}
	setIfPositive.add(inputVar, 1)
	setIfPositive.add(outputVar, -1)
	setIfPositive.finish(input, c_minusInf, 0)
}
