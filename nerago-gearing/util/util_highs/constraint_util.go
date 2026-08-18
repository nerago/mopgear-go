package util_highs

import (
	"math"
	"paladin_gearing_go/util/util_collection"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

func (build *LinearBuilder) ConstraintCopy(sourceVar ColumnIndex, sourceCoefficient float64, targetVar ColumnIndex, debug string) {
	rowCopy := ConstraintRow{Debug: debug}
	rowCopy.Add(sourceVar, sourceCoefficient)
	rowCopy.Add(targetVar, -1)
	rowCopy.Build(build, 0, 0)
}

func (build *LinearBuilder) ConstraintCopyIfBoolElseZero(boolSwitchVar, sourceVar, targetVar ColumnIndex, rangeLow, rangeHigh float64) {
	// based on https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d

	valueHigh := ConstraintRow{Debug: "ContraintIfBoolCopyValueElseZero_ValueHigh"}
	valueHigh.Add(targetVar, -1)
	valueHigh.Add(sourceVar, 1)
	valueHigh.Add(boolSwitchVar, rangeHigh)
	valueHigh.Build(build, InfNeg(), rangeHigh)

	valueLow := ConstraintRow{Debug: "ContraintIfBoolCopyValueElseZero_ValueLow"}
	valueLow.Add(targetVar, 1)
	valueLow.Add(sourceVar, -1)
	valueLow.Add(boolSwitchVar, -rangeLow)
	valueLow.Build(build, InfNeg(), -rangeLow)

	zeroHigh := ConstraintRow{Debug: "ContraintIfBoolCopyValueElseZero_ZeroHigh"}
	zeroHigh.Add(targetVar, 1)
	zeroHigh.Add(boolSwitchVar, -rangeHigh)
	zeroHigh.Build(build, InfNeg(), 0)

	zeroLow := ConstraintRow{Debug: "ContraintIfBoolCopyValueElseZero_ZeroLow"}
	zeroLow.Add(targetVar, -1)
	zeroLow.Add(boolSwitchVar, rangeLow)
	zeroLow.Build(build, InfNeg(), 0)
}

func (build *LinearBuilder) ConstraintCopyIfBool(boolSwitchVar, sourceVar ColumnIndex, sourceCoefficient float64, targetVar ColumnIndex, rangeHigh float64) {
	valueHigh := ConstraintRow{Debug: "ContraintIfBoolCopy_ValueHigh"}
	valueHigh.Add(targetVar, -1)
	valueHigh.Add(sourceVar, sourceCoefficient)
	valueHigh.Add(boolSwitchVar, rangeHigh)
	valueHigh.Build(build, InfNeg(), rangeHigh)

	valueLow := ConstraintRow{Debug: "ContraintIfBoolCopy_ValueLow"}
	valueLow.Add(targetVar, 1)
	valueLow.Add(sourceVar, -sourceCoefficient)
	valueLow.Add(boolSwitchVar, rangeHigh)
	valueLow.Build(build, InfNeg(), rangeHigh)
}

func (build *LinearBuilder) ConstraintAnd(outputVar ColumnIndex, inputVars ...ColumnIndex) {
	and := ConstraintAndBuilder{}
	and.SetOutput(outputVar)
	for _, input := range inputVars {
		and.AddInput(input)
	}
	and.Build(build)
}

func (build *LinearBuilder) ConstraintOr(outputVar ColumnIndex, inputVars ...ColumnIndex) {
	or := ConstraintOrBuilder{}
	or.SetOutput(outputVar)
	for _, input := range inputVars {
		or.AddInput(input)
	}
	or.Build(build)
}

// https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d
type ConstraintAndBuilder struct {
	outputVar ColumnIndex
	inputVars []ColumnIndex
}

func (and *ConstraintAndBuilder) SetOutput(column ColumnIndex) {
	and.outputVar = column
}

func (and *ConstraintAndBuilder) AddInput(column ColumnIndex) {
	if !slices.Contains(and.inputVars, column) {
		and.inputVars = append(and.inputVars, column)
	}
}

func (and *ConstraintAndBuilder) Build(build *LinearBuilder) {
	if len(and.inputVars) == 0 {
		setTrue := ConstraintRow{Debug: "ContraintAndBuilder_setZero"}
		setTrue.Add(and.outputVar, 1)
		setTrue.Build(build, 1, 1)
	} else if len(and.inputVars) == 1 {
		copyRow := ConstraintRow{Debug: "ContraintAndBuilder_copy"}
		copyRow.Add(and.outputVar, -1)
		copyRow.Add(and.inputVars[0], 1)
		copyRow.Build(build, 0, 0)
	} else {
		sumRow := ConstraintRow{Debug: "ConstraintAndBuilder_sumRow"}
		sumRow.Add(and.outputVar, -1)

		for _, inputVar := range and.inputVars {
			sumRow.Add(inputVar, 1)

			pullDown := ConstraintRow{Debug: "ConstraintAndBuilder_pullDown"}
			pullDown.Add(inputVar, -1)
			pullDown.Add(and.outputVar, 1)
			pullDown.Build(build, InfNeg(), 0)
		}

		targetNum := len(and.inputVars) - 1
		sumRow.Build(build, InfNeg(), float64(targetNum))
	}
}

type ConstraintOrBuilder struct {
	outputVar ColumnIndex
	inputVars []ColumnIndex
}

func (or *ConstraintOrBuilder) SetOutput(column ColumnIndex) {
	or.outputVar = column
}

func (or *ConstraintOrBuilder) AddInputs(columns []ColumnIndex) {
	or.inputVars = append(or.inputVars, columns...)
	util_collection.RemoveDuplicatesComparable_InPlace(&or.inputVars)
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
	zeroIfNone.Build(build, 0, InfPos())
}

func (build *LinearBuilder) ConstraintNot(inputVar, outputVar ColumnIndex) {
	row := ConstraintRow{Debug: "Not"}
	row.Add(inputVar, 1)
	row.Add(outputVar, 1)
	row.Build(build, 1, 1)
}

func (build *LinearBuilder) NotAsColumn(inputVar ColumnIndex) ColumnIndex {
	outputVar := build.CreateColumnBool(nil)
	build.ConstraintNot(inputVar, outputVar)
	return outputVar
}

func (build *LinearBuilder) AbsoluteValue(inputVar, outputVar ColumnIndex) {
	negative := ConstraintRow{Debug: "AbsoluteValueNegative"}
	negative.Add(inputVar, 1)
	negative.Add(outputVar, 1)
	negative.Build(build, 0, InfPos())

	positive := ConstraintRow{Debug: "AbsoluteValuePositive"}
	positive.Add(inputVar, 1)
	positive.Add(outputVar, -1)
	positive.Build(build, InfNeg(), 0)
}

func (build *LinearBuilder) AbsoluteValueNonFree_NeedMIP(inputVar, outputVar ColumnIndex, highRange float64, debug string) {
	isNegative := build.CreateColumnBool(DebugString{Text: debug + " isNegative"})

	setIsNegative := ConstraintRow{Debug: debug + " setIsNegative"}
	setIsNegative.Add(inputVar, 1)
	setIsNegative.Add(isNegative, highRange)
	setIsNegative.Build(build, 0, InfPos())

	confirmIsNegative := ConstraintRow{Debug: debug + "confirmIsNegative"}
	confirmIsNegative.Add(inputVar, 1)
	confirmIsNegative.Add(isNegative, highRange)
	confirmIsNegative.Build(build, InfNeg(), highRange)

	copyNegativeHigh := ConstraintRow{Debug: debug + " copyNegativeHigh"}
	copyNegativeHigh.Add(inputVar, -1)
	copyNegativeHigh.Add(isNegative, highRange)
	copyNegativeHigh.Add(outputVar, -1)
	copyNegativeHigh.Build(build, InfNeg(), highRange)

	copyNegativeLow := ConstraintRow{Debug: debug + " copyNegativeLow"}
	copyNegativeLow.Add(inputVar, 1)
	copyNegativeLow.Add(isNegative, highRange)
	copyNegativeLow.Add(outputVar, 1)
	copyNegativeLow.Build(build, InfNeg(), highRange)

	copyPositiveHigh := ConstraintRow{Debug: debug + " copyPositiveHigh"}
	copyPositiveHigh.Add(inputVar, 1)
	copyPositiveHigh.Add(isNegative, -highRange)
	copyPositiveHigh.Add(outputVar, -1)
	copyPositiveHigh.Build(build, InfNeg(), 0)

	copyPositiveLow := ConstraintRow{Debug: debug + " copyPositiveLow"}
	copyPositiveLow.Add(inputVar, -1)
	copyPositiveLow.Add(isNegative, -highRange)
	copyPositiveLow.Add(outputVar, 1)
	copyPositiveLow.Build(build, InfNeg(), 0)
}

func (build *LinearBuilder) AbsoluteValueFromDiffTwoVars(inputOneVar ColumnIndex, inputOneCoefficient float64, inputTwoVar ColumnIndex, inputTwoCoefficient float64, outputVar ColumnIndex, debug string) {
	// one - two + out >= 0
	// so if one - two < 0, out neeeds to pull it up
	negative := ConstraintRow{Debug: debug + " AbsoluteValueNegative"}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(inputTwoVar, -inputTwoCoefficient)
	negative.Add(outputVar, 1)
	negative.Build(build, 0, InfPos())

	// one - two - out <= 0
	// so if one - two > 0, out needed to pull it down
	positive := ConstraintRow{Debug: debug + " AbsoluteValuePositive"}
	positive.Add(inputOneVar, inputOneCoefficient)
	positive.Add(inputTwoVar, -inputTwoCoefficient)
	positive.Add(outputVar, -1)
	positive.Build(build, InfNeg(), 0)
}

const c_minimumAcceptable = 1e-6

func isTiny(coefficient float64) bool {
	pos := math.Abs(coefficient)
	return 0 < pos && pos < c_minimumAcceptable
}
func chooseTinyValueScale(a float64, b float64, c float64) float64 {
	value := c_minimumAcceptable
	if isTiny(a) {
		value = min(value, math.Abs(a))
	}
	if isTiny(b) {
		value = min(value, math.Abs(b))
	}
	if isTiny(c) {
		value = min(value, math.Abs(c))
	}
	return c_minimumAcceptable / value * 2
}

func (build *LinearBuilder) AbsoluteValueFromDiffTwoVars_ScaleOutput(inputOneVar ColumnIndex, inputOneCoefficient float64, inputTwoVar ColumnIndex, inputTwoCoefficient float64, outputVar ColumnIndex, outputCoefficient float64, debug string) {
	if isTiny(inputOneCoefficient) || isTiny(inputTwoCoefficient) || isTiny(outputCoefficient) {
		scale := chooseTinyValueScale(inputOneCoefficient, inputTwoCoefficient, outputCoefficient)
		inputOneCoefficient *= scale
		inputTwoCoefficient *= scale
		outputCoefficient *= scale
	}

	// one - two + out >= 0
	// so if one - two < 0, out needs to pull it up
	negative := ConstraintRow{Debug: debug + " AbsoluteValueNegative"}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(inputTwoVar, -inputTwoCoefficient)
	negative.Add(outputVar, outputCoefficient)
	negative.Build(build, 0, InfPos())

	// one - two - out <= 0
	// so if one - two > 0, out needed to pull it down
	positive := ConstraintRow{Debug: debug + " AbsoluteValuePositive"}
	positive.Add(inputOneVar, inputOneCoefficient)
	positive.Add(inputTwoVar, -inputTwoCoefficient)
	positive.Add(outputVar, -outputCoefficient)
	positive.Build(build, InfNeg(), 0)
}

func (build *LinearBuilder) AbsoluteValueFromDiffTwoVarsNonFree(inputOneVar ColumnIndex, inputOneCoefficient float64, inputTwoVar ColumnIndex, inputTwoCoefficient float64, outputVar ColumnIndex, highRange float64, debug string) {
	isNegative := build.CreateColumnBool(DebugString{Text: debug + " isNegative"})

	setIsNegative := ConstraintRow{Debug: debug + " setIsNegative"}
	setIsNegative.Add(inputOneVar, inputOneCoefficient)
	setIsNegative.Add(inputTwoVar, -inputTwoCoefficient)
	setIsNegative.Add(isNegative, highRange)
	setIsNegative.Build(build, 0, InfPos())

	confirmIsNegative := ConstraintRow{Debug: debug + "confirmIsNegative"}
	confirmIsNegative.Add(inputOneVar, inputOneCoefficient)
	confirmIsNegative.Add(inputTwoVar, -inputTwoCoefficient)
	confirmIsNegative.Add(isNegative, highRange)
	confirmIsNegative.Build(build, InfNeg(), highRange)

	copyNegativeHigh := ConstraintRow{Debug: debug + " copyNegativeHigh"}
	copyNegativeHigh.Add(inputOneVar, -inputOneCoefficient)
	copyNegativeHigh.Add(inputTwoVar, inputTwoCoefficient)
	copyNegativeHigh.Add(isNegative, highRange)
	copyNegativeHigh.Add(outputVar, -1)
	copyNegativeHigh.Build(build, InfNeg(), highRange)

	copyNegativeLow := ConstraintRow{Debug: debug + " copyNegativeLow"}
	copyNegativeLow.Add(inputOneVar, inputOneCoefficient)
	copyNegativeLow.Add(inputTwoVar, -inputTwoCoefficient)
	copyNegativeLow.Add(isNegative, highRange)
	copyNegativeLow.Add(outputVar, 1)
	copyNegativeLow.Build(build, InfNeg(), highRange)

	copyPositiveHigh := ConstraintRow{Debug: debug + " copyPositiveHigh"}
	copyPositiveHigh.Add(inputOneVar, inputOneCoefficient)
	copyPositiveHigh.Add(inputTwoVar, -inputTwoCoefficient)
	copyPositiveHigh.Add(isNegative, -highRange)
	copyPositiveHigh.Add(outputVar, -1)
	copyPositiveHigh.Build(build, InfNeg(), 0)

	copyPositiveLow := ConstraintRow{Debug: debug + " copyPositiveLow"}
	copyPositiveLow.Add(inputOneVar, -inputOneCoefficient)
	copyPositiveLow.Add(inputTwoVar, inputTwoCoefficient)
	copyPositiveLow.Add(isNegative, -highRange)
	copyPositiveLow.Add(outputVar, 1)
	copyPositiveLow.Build(build, InfNeg(), 0)
}

// out = abs(input-const)
func (build *LinearBuilder) AbsoluteValueFromDiffOneToConst(inputOneVar ColumnIndex, inputOneCoefficient float64, constCompare float64, outputVar ColumnIndex, debug string) {
	negative := ConstraintRow{Debug: debug + " AbsoluteValueNegative"}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(outputVar, 1)
	negative.Build(build, constCompare, InfPos())

	positive := ConstraintRow{Debug: debug + " AbsoluteValuePositive"}
	positive.Add(inputOneVar, inputOneCoefficient)
	positive.Add(outputVar, -1)
	positive.Build(build, InfNeg(), constCompare)
}

// out = abs(abs(inputVar)-const)
func (build *LinearBuilder) AbsoluteValueBeforeDiffOneToConst(inputVar ColumnIndex, inputCoefficient float64, constCompare float64, outputVar ColumnIndex, debug string) {
	//firstAbs := build.CreateColumnGeneral(highs.Continuous, 0, InfPos(), DebugText(debug+" abs"))

	// out = abs(abs(inputVar)-const)
	// cases inputVar>=0 inputVar>=const out=inputVar-const      inputVar-out=const
	// cases inputVar>=0 inputVar<const  out=const-inputVar      out+inputVar=const
	// cases inputVar<0 -inputVar>const  out=-inputVar-const     out+inputVar=-const
	// cases inputVar<0 -inputVar<const  out=const+inputVar      out-inputVar=const

	a := ConstraintRow{}
	a.Add(inputVar, inputCoefficient)
	a.Add(outputVar, -1)
	a.Build(build, InfNeg(), constCompare)

	//b := ConstraintRow{}
	//b.Add(inputVar, inputCoefficient)
	//b.Add(outputVar, 1)
	//b.Build(build, -constCompare, constCompare)

	c := ConstraintRow{}
	c.Add(inputVar, inputCoefficient)
	c.Add(outputVar, 1)
	c.Build(build, constCompare, InfPos())

	//d := ConstraintRow{}
	//d.Add(inputVar, -inputCoefficient)
	//d.Add(outputVar, 1)
	//d.Build(build, constCompare, InfPos())

	//b := ConstraintRow{}
	//b.Add(inputVar, inputCoefficient)
	//b.Add(outputVar, 1)
	//b.Build(build, -constCompare, constCompare)

	//negative := ConstraintRow{}
	//negative.Add(inputVar, inputCoefficient)
	//negative.Add(firstAbs, 1)
	//negative.Build(build, 0, InfPos())
	//
	//positive := ConstraintRow{}
	//positive.Add(inputVar, inputCoefficient)
	//positive.Add(firstAbs, -1)
	//positive.Build(build, InfNeg(), 0)
	//
	//negative2 := ConstraintRow{}
	//negative2.Add(firstAbs, 1)
	//negative2.Add(outputVar, 1)
	//negative2.Build(build, constCompare, InfPos())
	//
	//positive2 := ConstraintRow{}
	//positive2.Add(firstAbs, 1)
	//positive2.Add(outputVar, -1)
	//positive2.Build(build, InfNeg(), constCompare)
}

func (build *LinearBuilder) AbsoluteValueFromSumTwoThenDiffToConst(inputOneVar ColumnIndex, inputOneCoefficient float64, inputTwoVar ColumnIndex, inputTwoCoefficient float64, constCompare float64, outputVar ColumnIndex, debug string) {
	negative := ConstraintRow{Debug: debug + " AbsoluteValueNegative"}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(inputTwoVar, inputTwoCoefficient)
	negative.Add(outputVar, 1)
	negative.Build(build, constCompare, InfPos())

	positive := ConstraintRow{Debug: debug + " AbsoluteValuePositive"}
	positive.Add(inputOneVar, inputOneCoefficient)
	positive.Add(inputTwoVar, inputTwoCoefficient)
	positive.Add(outputVar, -1)
	positive.Build(build, InfNeg(), constCompare)
}

func (build *LinearBuilder) AbsoluteValueFromSumSeveralThenDiffToConst(inputVars []ColumnIndex, inputCoefficients []float64, constCompare float64, outputVar ColumnIndex, debug string) {
	if len(inputVars) != len(inputCoefficients) {
		panic("length mismatch")
	}

	negative := ConstraintRow{Debug: debug + " AbsoluteValueNegative"}
	for i := range inputVars {
		negative.Add(inputVars[i], inputCoefficients[i])
	}
	negative.Add(outputVar, 1)
	negative.Build(build, constCompare, InfPos())

	positive := ConstraintRow{Debug: debug + " AbsoluteValuePositive"}
	for i := range inputVars {
		positive.Add(inputVars[i], inputCoefficients[i])
	}
	positive.Add(outputVar, -1)
	positive.Build(build, InfNeg(), constCompare)
}

func (build *LinearBuilder) AbsoluteValueDiffTwoVarsThenDiffConst_NeedMIP(inputOneVar ColumnIndex, inputOneCoefficient float64, inputTwoVar ColumnIndex, inputTwoCoefficient float64, outputVar ColumnIndex, diffConst float64, highRange float64, debug string) {
	diffTwoVars := build.CreateColumnGeneral(highs.Continuous, InfNeg(), InfPos(), DebugText(debug+" diffTwoVars"))

	isNegative := build.CreateColumnBool(DebugString{Text: debug + " isNegative"})
	setIsNegative := ConstraintRow{Debug: debug + " setIsNegative"}
	setIsNegative.Add(inputOneVar, inputOneCoefficient)
	setIsNegative.Add(inputTwoVar, -inputTwoCoefficient)
	setIsNegative.Add(isNegative, highRange)
	setIsNegative.Build(build, 0, InfPos())
	confirmIsNegative := ConstraintRow{Debug: debug + "confirmIsNegative"}
	confirmIsNegative.Add(inputOneVar, inputOneCoefficient)
	confirmIsNegative.Add(inputTwoVar, -inputTwoCoefficient)
	confirmIsNegative.Add(isNegative, highRange)
	confirmIsNegative.Build(build, InfNeg(), highRange)

	copyNegativeHigh := ConstraintRow{Debug: debug + " copyNegativeHigh"}
	copyNegativeHigh.Add(inputOneVar, -inputOneCoefficient)
	copyNegativeHigh.Add(inputTwoVar, inputTwoCoefficient)
	copyNegativeHigh.Add(isNegative, highRange)
	copyNegativeHigh.Add(diffTwoVars, -1)
	copyNegativeHigh.Build(build, InfNeg(), highRange)
	copyNegativeLow := ConstraintRow{Debug: debug + " copyNegativeLow"}
	copyNegativeLow.Add(inputOneVar, inputOneCoefficient)
	copyNegativeLow.Add(inputTwoVar, -inputTwoCoefficient)
	copyNegativeLow.Add(isNegative, highRange)
	copyNegativeLow.Add(diffTwoVars, 1)
	copyNegativeLow.Build(build, InfNeg(), highRange)
	copyPositiveHigh := ConstraintRow{Debug: debug + " copyPositiveHigh"}
	copyPositiveHigh.Add(inputOneVar, inputOneCoefficient)
	copyPositiveHigh.Add(inputTwoVar, -inputTwoCoefficient)
	copyPositiveHigh.Add(isNegative, -highRange)
	copyPositiveHigh.Add(diffTwoVars, -1)
	copyPositiveHigh.Build(build, InfNeg(), 0)
	copyPositiveLow := ConstraintRow{Debug: debug + " copyPositiveLow"}
	copyPositiveLow.Add(inputOneVar, -inputOneCoefficient)
	copyPositiveLow.Add(inputTwoVar, inputTwoCoefficient)
	copyPositiveLow.Add(isNegative, -highRange)
	copyPositiveLow.Add(diffTwoVars, 1)
	copyPositiveLow.Build(build, InfNeg(), 0)

	diffConstNegative := ConstraintRow{Debug: debug + " diffConstNegative"}
	diffConstNegative.Add(diffTwoVars, 1)
	diffConstNegative.Add(outputVar, 1)
	diffConstNegative.Build(build, diffConst, InfPos())
	diffConstPositive := ConstraintRow{Debug: debug + " diffConstPositive"}
	diffConstPositive.Add(diffTwoVars, 1)
	diffConstPositive.Add(outputVar, -1)
	diffConstPositive.Build(build, InfNeg(), diffConst)
}

func (build *LinearBuilder) AbsoluteValue_WithToggle(inputVar, outputVar, toggleVar ColumnIndex, rangeHigh float64) {
	setIfNegative := ConstraintRow{}
	setIfNegative.Add(inputVar, 1)
	setIfNegative.Add(outputVar, 1)
	setIfNegative.Add(toggleVar, -rangeHigh)
	setIfNegative.Build(build, -rangeHigh, InfPos())

	setIfPositive := ConstraintRow{}
	setIfPositive.Add(inputVar, 1)
	setIfPositive.Add(outputVar, -1)
	setIfPositive.Add(toggleVar, rangeHigh)
	setIfPositive.Build(build, InfNeg(), rangeHigh)

	// ideally outputVar should be limited to zero already
	zeroMinimum := ConstraintRow{}
	zeroMinimum.Add(outputVar, 1)
	zeroMinimum.Build(build, 0, InfPos())
}

func (build *LinearBuilder) AbsoluteValue_WithToggle_NoExtraCheck(inputVar, outputVar, toggleVar ColumnIndex, rangeHigh float64) {
	setIfNegative := ConstraintRow{}
	setIfNegative.Add(inputVar, 1)
	setIfNegative.Add(outputVar, 1)
	setIfNegative.Add(toggleVar, -rangeHigh)
	setIfNegative.Build(build, -rangeHigh, InfPos())

	setIfPositive := ConstraintRow{}
	setIfPositive.Add(inputVar, 1)
	setIfPositive.Add(outputVar, -1)
	setIfPositive.Add(toggleVar, rangeHigh)
	setIfPositive.Build(build, InfNeg(), rangeHigh)
}

func (build *LinearBuilder) AbsoluteValueFromDiffTwoVars_WithToggle(inputOneVar ColumnIndex, inputOneCoefficient float64, inputTwoVar ColumnIndex, inputTwoCoefficient float64, toggleVar ColumnIndex, outputVar ColumnIndex, rangeHigh float64) {
	negative := ConstraintRow{}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(inputTwoVar, -inputTwoCoefficient)
	negative.Add(outputVar, 1)
	negative.Add(toggleVar, -rangeHigh)
	negative.Build(build, -rangeHigh, InfPos())

	positive := ConstraintRow{}
	positive.Add(inputOneVar, inputOneCoefficient)
	positive.Add(inputTwoVar, -inputTwoCoefficient)
	positive.Add(outputVar, -1)
	positive.Add(toggleVar, rangeHigh)
	positive.Build(build, InfNeg(), rangeHigh)
}

func (build *LinearBuilder) AbsoluteValueFromSumTwoThenDiffToConst_WithToggle(inputOneVar ColumnIndex, inputOneCoefficient float64, inputTwoVar ColumnIndex, inputTwoCoefficient float64, constCompare float64, toggleVar ColumnIndex, outputVar ColumnIndex, rangeHigh float64) {
	negative := ConstraintRow{}
	negative.Add(inputOneVar, inputOneCoefficient)
	negative.Add(inputTwoVar, inputTwoCoefficient)
	negative.Add(outputVar, 1)
	negative.Add(toggleVar, -rangeHigh)
	negative.Build(build, constCompare-rangeHigh, InfPos())

	positive := ConstraintRow{}
	positive.Add(inputOneVar, inputOneCoefficient)
	positive.Add(inputTwoVar, inputTwoCoefficient)
	positive.Add(outputVar, -1)
	positive.Add(toggleVar, rangeHigh)
	positive.Build(build, InfNeg(), constCompare+rangeHigh)
}

func (build *LinearBuilder) AbsoluteValueFromSumSeveral_WithToggle(inputVars []ColumnIndex, inputCoefficients []float64, constCompare float64, toggleVar ColumnIndex, outputVar ColumnIndex, rangeHigh float64) {
	if len(inputVars) != len(inputCoefficients) {
		panic("length mismatch")
	}

	negative := ConstraintRow{}
	for i := range inputVars {
		negative.Add(inputVars[i], inputCoefficients[i])
	}
	negative.Add(outputVar, 1)
	negative.Add(toggleVar, -rangeHigh)
	negative.Build(build, constCompare-rangeHigh, InfPos())

	positive := ConstraintRow{}
	for i := range inputVars {
		positive.Add(inputVars[i], inputCoefficients[i])
	}
	positive.Add(outputVar, -1)
	positive.Add(toggleVar, rangeHigh)
	positive.Build(build, InfNeg(), constCompare+rangeHigh)
}

// basic logic: output = one xor two
// however output is free when condition not met, should ideally put output under minimise pressure
// similar to absolute value, just intended for int vars
func (build *LinearBuilder) IsXor(boolOne ColumnIndex, boolTwo ColumnIndex, output ColumnIndex) {
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
func (build *LinearBuilder) ColumnIsGreaterOrEqualThanConstant(compareColumn ColumnIndex, constValue float64, rangeHigh float64, equalDelta float64) ColumnIndex {
	isGreaterEqual := build.CreateColumnBool(DebugString{Text: "isGreaterEqual"})
	build.ColumnIsGreaterOrEqualThanConstant_Supplied(isGreaterEqual, compareColumn, constValue, rangeHigh, equalDelta)
	return isGreaterEqual
}

func (build *LinearBuilder) ColumnIsGreaterOrEqualThanConstant_Supplied(isGreaterEqual ColumnIndex, compareColumn ColumnIndex, constValue float64, rangeHigh float64, equalDelta float64) {
	set := ConstraintRow{Debug: "ColumnIsGreaterOrEqualThanConstant_set"}
	set.Add(compareColumn, 1)
	set.Add(isGreaterEqual, -rangeHigh)
	set.Build(build, InfNeg(), constValue-equalDelta)

	confirm := ConstraintRow{Debug: "ColumnIsGreaterOrEqualThanConstant_confirm"}
	confirm.Add(isGreaterEqual, constValue)
	confirm.Add(compareColumn, -1)
	confirm.Build(build, InfNeg(), 0)
}

// rangeHigh should be bigger than any possible value
// equalDelta should be just big enough to make unequal (1.0 for ints)
// logic: colValue <= constValue
func (build *LinearBuilder) ColumnIsLessOrEqualThanConstant(compareColumn ColumnIndex, constValue float64, rangeHigh float64, equalDelta float64) ColumnIndex {
	isLessEqual := build.CreateColumnBool(DebugString{Text: "isLessEqual"})
	build.ColumnIsLessOrEqualThanConstant_Supplied(isLessEqual, compareColumn, constValue, rangeHigh, equalDelta)
	return isLessEqual
}

func (build *LinearBuilder) ColumnIsLessOrEqualThanConstant_Supplied(isLessEqual ColumnIndex, compareColumn ColumnIndex, constValue float64, rangeHigh float64, equalDelta float64) {
	set := ConstraintRow{Debug: "ColumnIsLessOrEqualThanConstant_set"}
	set.Add(compareColumn, 1)
	set.Add(isLessEqual, rangeHigh)
	set.Build(build, constValue+equalDelta, InfPos())

	confirm := ConstraintRow{Debug: "ColumnIsLessOrEqualThanConstant_confirm"}
	confirm.Add(isLessEqual, rangeHigh)
	confirm.Add(compareColumn, 1)
	confirm.Build(build, InfNeg(), rangeHigh+constValue)
}

func (build *LinearBuilder) ConstantIsBetweenColumns(minimumColumn, maximumColumn, targetBoolColumn ColumnIndex, constValue float64, rangeHigh float64, equalDelta float64) {
	isOverMinimum := build.ColumnIsLessOrEqualThanConstant(minimumColumn, constValue, rangeHigh, equalDelta)
	isUnderMaximum := build.ColumnIsGreaterOrEqualThanConstant(maximumColumn, constValue, rangeHigh, equalDelta)

	and := ConstraintAndBuilder{}
	and.AddInput(isOverMinimum)
	and.AddInput(isUnderMaximum)
	and.SetOutput(targetBoolColumn)
	and.Build(build)

	// NOTE: this may not be needed and hurt performance for some callers, could have alternate version without
	// but generally makes sense for consistency of expectations
	checkSequence := ConstraintRow{}
	checkSequence.Add(maximumColumn, 1)
	checkSequence.Add(minimumColumn, -1)
	checkSequence.Build(build, 0, InfPos())
}

func (build *LinearBuilder) ConstantIsBetweenColumns_NoSequenceCheck(minimumColumn, maximumColumn, targetBoolColumn ColumnIndex, constValue float64, rangeHigh float64, equalDelta float64) {
	isOverMinimum := build.ColumnIsLessOrEqualThanConstant(minimumColumn, constValue, rangeHigh, equalDelta)
	isUnderMaximum := build.ColumnIsGreaterOrEqualThanConstant(maximumColumn, constValue, rangeHigh, equalDelta)

	and := ConstraintAndBuilder{}
	and.AddInput(isOverMinimum)
	and.AddInput(isUnderMaximum)
	and.SetOutput(targetBoolColumn)
	and.Build(build)
}

func (build *LinearBuilder) ColumnIsNotBetweenConstantsVerify(checkColumn ColumnIndex, lo, hi float64, rangeHigh float64) {
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

// inclusive ranges
func (build *LinearBuilder) ColumnIsBetweenConstants(checkColumn ColumnIndex, lo, hi float64, rangeHigh float64, equalDelta float64) ColumnIndex {
	if lo > hi {
		panic("backwards range")
	}

	isUnderMin := build.CreateColumnBool(DebugString{Text: "isUnderMin"})
	isOverMax := build.CreateColumnBool(DebugString{Text: "isOverMax"})

	underMin := ConstraintRow{Debug: "setIfOverMin"}
	underMin.Add(checkColumn, 1)
	underMin.Add(isUnderMin, rangeHigh)
	underMin.Build(build, lo, lo+rangeHigh-equalDelta)

	overMax := ConstraintRow{Debug: "setIfUnderMax"}
	overMax.Add(checkColumn, 1)
	overMax.Add(isOverMax, -rangeHigh)
	overMax.Build(build, hi-rangeHigh+equalDelta, hi)

	isBetween := build.CreateColumnBool(nil)
	check := ConstraintRow{}
	check.Add(isUnderMin, 1)
	check.Add(isOverMax, 1)
	check.Add(isBetween, 1)
	check.Build(build, 1, 1)
	return isBetween
}

// logic: leftSideCol < rightSideCol
func (build *LinearBuilder) ColumnIsLessThanColumnEqualityFree(leftSideCol, rightSideCol, boolIsLess ColumnIndex, rangeHigh float64) {
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
func (build *LinearBuilder) ColumnIsGreaterThanColumnEqualityFree(leftSideCol, rightSideCol, boolIsGreater ColumnIndex, rangeHigh float64) {
	build.ColumnIsLessThanColumnEqualityFree(rightSideCol, leftSideCol, boolIsGreater, rangeHigh)
}

// logic: leftSideCol >= rightSideCol
func (build *LinearBuilder) ColumnIsGreaterOrEqualColumn(leftSideCol, rightSideCol, boolIsGreater ColumnIndex, rangeHigh float64, equalDelta float64) {
	set := ConstraintRow{Debug: "ColumnIsGreaterOrEqualColumn_set"}
	set.Add(leftSideCol, 1)
	set.Add(rightSideCol, -1)
	set.Add(boolIsGreater, -rangeHigh)
	set.Build(build, InfNeg(), -equalDelta)

	confirm := ConstraintRow{Debug: "ColumnIsGreaterOrEqualColumn_confirm"}
	confirm.Add(leftSideCol, -1)
	confirm.Add(rightSideCol, 1)
	confirm.Add(boolIsGreater, rangeHigh)
	confirm.Build(build, InfNeg(), rangeHigh)
}

// logic: leftSideCol <= rightSideCol
func (build *LinearBuilder) ColumnIsLessOrEqualColumn(leftSideCol, rightSideCol, boolIsLess ColumnIndex, rangeHigh float64, equalDelta float64) {
	set := ConstraintRow{Debug: "ColumnIsLessOrEqualColumn_set"}
	set.Add(leftSideCol, 1)
	set.Add(rightSideCol, -1)
	set.Add(boolIsLess, rangeHigh)
	set.Build(build, equalDelta, InfPos())

	confirm := ConstraintRow{Debug: "ColumnIsLessOrEqualColumn_confirm"}
	confirm.Add(leftSideCol, 1)
	confirm.Add(rightSideCol, -1)
	confirm.Add(boolIsLess, rangeHigh)
	confirm.Build(build, InfNeg(), rangeHigh)
}

func (build *LinearBuilder) ColumnIsGreaterOrEqualColumnEnforce(lowCol, highCol ColumnIndex) {
	row := ConstraintRow{Debug: "ColumnIsGreaterOrEqualColumnEnforce"}
	row.Add(lowCol, -1)
	row.Add(highCol, 1)
	row.Build(build, 0, InfPos())
}

func (build *LinearBuilder) ColumnIsLessOrEqualColumnEnforce(lowCol, highCol ColumnIndex) {
	row := ConstraintRow{Debug: "ColumnIsLessOrEqualColumnEnforce"}
	row.Add(lowCol, -1)
	row.Add(highCol, 1)
	row.Build(build, InfNeg(), 0)
}

func (build *LinearBuilder) ColumnIsNotEqualConstant(checkColumn, boolIsUnequal ColumnIndex, constValue float64, rangeHigh float64, equalDelta float64) {
	isUnder := build.CreateColumnBool(DebugString{Text: "isUnderMin"})
	isOver := build.CreateColumnBool(DebugString{Text: "isOverMax"})

	under := ConstraintRow{Debug: ""}
	under.Add(checkColumn, 1)
	under.Add(isUnder, rangeHigh)
	under.Build(build, constValue, constValue+rangeHigh-equalDelta)

	over := ConstraintRow{Debug: ""}
	over.Add(checkColumn, 1)
	over.Add(isOver, -rangeHigh)
	over.Build(build, constValue-rangeHigh+equalDelta, constValue)

	or := ConstraintRow{}
	or.Add(isUnder, 1)
	or.Add(isOver, 1)
	or.Add(boolIsUnequal, -1)
	or.Build(build, 0, 0)
}

func (build *LinearBuilder) ColumnIsEqualConstant(checkColumn, boolIsEqual ColumnIndex, constValue float64, rangeHigh float64, equalDelta float64) {
	isUnder := build.CreateColumnBool(DebugString{Text: "isUnderMin"})
	isOver := build.CreateColumnBool(DebugString{Text: "isOverMax"})

	under := ConstraintRow{Debug: ""}
	under.Add(checkColumn, 1)
	under.Add(isUnder, rangeHigh)
	under.Build(build, constValue, constValue+rangeHigh-equalDelta)
	// const <= under*range + check <= const+range-delta

	over := ConstraintRow{Debug: ""}
	over.Add(checkColumn, 1)
	over.Add(isOver, -rangeHigh)
	over.Build(build, constValue-rangeHigh+equalDelta, constValue)
	// const-range+delta <= check - over*range <= const
	// -(const-range+delta) >= -(check - over*range) >= -const
	// -const <= over*range - check <= range-const-delta

	// const <= under*range + check <= const+range-delta
	// const-range+delta <= check - over*range <= const

	nor := ConstraintRow{}
	nor.Add(isUnder, 1)
	nor.Add(isOver, 1)
	nor.Add(boolIsEqual, 1)
	nor.Build(build, 1, 1)
}

// lazy, can enforce based on bool, but not reliably set bool
func (build *LinearBuilder) ColumnIsEqualConstant_OneWayEnforceNotSet(checkColumn, boolIsEqual ColumnIndex, constValue, rangeHigh float64) {
	// https://hegyhati.github.io/IMOLS/pages/modelling_bigM.html
	//If y=1 then LHS=RHS: LHS≥RHS−M⋅(1−y)
	//                     LHS≤RHS+M⋅(1−y)
	//   LHS>=RHS−M*(1−y)  --> check >= const - M(1-bool)
	//                     --> check >= const - M1 + M*bool
	//                     --> check - M*bool >= const - M
	//   LHS<=RHS+M*(1−y)  --> check <= const + M(1-bool)
	//                     --> check <= const + M1 - M*bool
	//                     --> check + M*bool <= const + M

	over := ConstraintRow{Debug: ""}
	over.Add(checkColumn, 1)
	over.Add(boolIsEqual, -rangeHigh)
	over.Build(build, constValue-rangeHigh, InfPos())

	under := ConstraintRow{Debug: ""}
	under.Add(checkColumn, 1)
	under.Add(boolIsEqual, rangeHigh)
	under.Build(build, InfNeg(), constValue+rangeHigh)
}

func (build *LinearBuilder) ColumnIsNotEqualConstant_OneWayEnforceNotSet(checkColumn, boolIsNotEqual ColumnIndex, constValue, rangeHigh float64) {
	//   LHS>=RHS−M*y      --> check >= const - M*bool
	//                     --> check + M*bool >= const
	//   LHS<=RHS+M*y      --> check <= const + M*bool
	//                     --> check - M*bool <= const

	over := ConstraintRow{Debug: ""}
	over.Add(checkColumn, 1)
	over.Add(boolIsNotEqual, rangeHigh)
	over.Build(build, constValue, InfPos())

	under := ConstraintRow{Debug: ""}
	under.Add(checkColumn, 1)
	under.Add(boolIsNotEqual, -rangeHigh)
	under.Build(build, InfNeg(), constValue)
}
