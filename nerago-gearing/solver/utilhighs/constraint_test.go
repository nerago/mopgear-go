package utilhighs

import (
	"paladin_gearing_go/util"
	"testing"

	"github.com/bartolsthoorn/gohighs/highs"
)

func TestContraintIfBoolCopyValueElseZero(test *testing.T) {
	//TODO
}

func TestContraintIfBoolCopy(test *testing.T) {
	//TODO
}

func TestContraintAndBuilder(test *testing.T) {
	//TODO
}

func TestConstraintOrBuilder(test *testing.T) {
	//TODO
}

func TestConstraintNot(test *testing.T) {
	//TODO
}

func TestNotAsColumn(test *testing.T) {
	//TODO
}

func TestAbsoluteValue(test *testing.T) {
	//TODO
}

func TestAbsoluteValueFromDiffTwoVars(test *testing.T) {
	//TODO
}

func TestAbsoluteValueFromDiffOneToConst(test *testing.T) {
	//TODO
}

func TestAbsoluteValueFromDiffTwoVarsWithOffset(test *testing.T) {
	//TODO
}

func TestAbsoluteValue_WithToggle(test *testing.T) {
	//TODO
}

func TestIsXor(test *testing.T) {
	//TODO
}

func TestColumnIsGreaterOrEqualThanConstant(test *testing.T) {
maxValue := 100.0
	rangeHigh := 200.0
	equalDelta := 1.0

	testValues := func(colValue, checkValue float64, setBool *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: col=%f check=%f", colValue, checkValue)
		if setBool != nil {
			test.Logf("bool=%f", *setBool)
		}

		build := new(InputBuilder)
		build.NoOutput = true
		compareColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		setColumnToConstant(build, compareColumn, colValue)

		isLess := ColumnIsGreaterOrEqualThanConstant(build, compareColumn, checkValue, rangeHigh, equalDelta)

		if setBool != nil {
			setColumnToConstant(build, isLess, *setBool)
		}
		solution, _ := build.RunHighs()
		build.DebugPrintColumns(solution, util.PrintRecorder_Testing())

		boolOutput := solution.ColValues[isLess]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, colValue <= checkValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
		if solution.Status.HasSolution() {
			assertEqual(colValue >= checkValue, FloatEqualsOne(boolOutput), test)
		}
	}

	zero := ptr(0.0)
	one := ptr(1.0)

	testValues(49, 50, nil, highs.ModelStatusOptimal, zero)
	testValues(49, 50, zero, highs.ModelStatusOptimal, zero)
	testValues(49, 50, one, highs.ModelStatusInfeasible, nil)
	testValues(50, 50, nil, highs.ModelStatusOptimal, one)
	testValues(50, 50, zero, highs.ModelStatusInfeasible, nil)
	testValues(50, 50, one, highs.ModelStatusOptimal, one)
	testValues(51, 50, nil, highs.ModelStatusOptimal, one)
	testValues(51, 50, zero, highs.ModelStatusInfeasible, nil)
	testValues(51, 50, one, highs.ModelStatusOptimal, one)
}

func TestColumnIsLessOrEqualThanConstant(test *testing.T) {
	maxValue := 100.0
	rangeHigh := 200.0
	equalDelta := 1.0

	testValues := func(colValue, checkValue float64, setBool *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: col=%f check=%f", colValue, checkValue)
		if setBool != nil {
			test.Logf("bool=%f", *setBool)
		}

		build := new(InputBuilder)
		build.NoOutput = true
		compareColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		setColumnToConstant(build, compareColumn, colValue)

		isLess := ColumnIsLessOrEqualThanConstant(build, compareColumn, checkValue, rangeHigh, equalDelta)

		if setBool != nil {
			setColumnToConstant(build, isLess, *setBool)
		}
		solution, _ := build.RunHighs()
		build.DebugPrintColumns(solution, util.PrintRecorder_Testing())

		boolOutput := solution.ColValues[isLess]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, colValue <= checkValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
		if solution.Status.HasSolution() {
			assertEqual(colValue <= checkValue, FloatEqualsOne(boolOutput), test)
		}
	}

	zero := ptr(0.0)
	one := ptr(1.0)

	testValues(49, 50, nil, highs.ModelStatusOptimal, one)
	testValues(49, 50, zero, highs.ModelStatusInfeasible, nil)
	testValues(49, 50, one, highs.ModelStatusOptimal, one)

	testValues(50, 50, nil, highs.ModelStatusOptimal, one)
	testValues(50, 50, zero, highs.ModelStatusInfeasible, nil)
	testValues(50, 50, one, highs.ModelStatusOptimal, one)

	testValues(51, 50, nil, highs.ModelStatusOptimal, zero)
	testValues(51, 50, zero, highs.ModelStatusOptimal, zero)
	testValues(51, 50, one, highs.ModelStatusInfeasible, nil)
}

func TestConstantIsBetweenColumns(test *testing.T) {
	//TODO
}

func TestColumnIsNotBetweenConstants(test *testing.T) {
	//TODO
}

func TestColumnIsGreaterOrEqualColumn(test *testing.T) {
	//TODO
}

func setColumnToConstant(build *InputBuilder, column ColumnIndex, value float64) {
	row := ConstraintRowBuild{}
	row.Add(column, 1)
	row.Finish(build, value, value)
}

func ptr(value float64) *float64 {
	return &value
}

func assertEqual[T comparable](expect, actual T, test *testing.T) {
	if expect != actual {
		test.Fatalf("assertEqual failed expect=%v actual=%v", expect, actual)
	}
}
