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
func AbsoluteValueFromDiff(input *InputBuilder, inputOneVar ColumnIndex, inputOneCoefficient float64, inputTwoVar ColumnIndex, inputTwoCoefficient float64, outputVar ColumnIndex, debug string) {
	negative := ConstraintRowBuild{Debug: debug + " AbsoluteValueNegative"}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(inputTwoVar, -inputTwoCoefficient)
	negative.Add(outputVar, 1)
	negative.Finish(input, 0, C_PlusInf)

	positive := ConstraintRowBuild{Debug: debug + " AbsoluteValuePositive"}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(inputTwoVar, -inputTwoCoefficient)
	positive.Add(outputVar, -1)
	positive.Finish(input, C_MinusInf, 0)
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
