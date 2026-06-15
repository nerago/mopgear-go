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

		build := new(LinearBuilder)
		build.NoOutput = true
		compareColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		setColumnToConstant(build, compareColumn, colValue)

		isGreater := ColumnIsGreaterOrEqualThanConstant(build, compareColumn, checkValue, rangeHigh, equalDelta)

		if setBool != nil {
			setColumnToConstant(build, isGreater, *setBool)
		}
		solution, _ := build.RunHighs()
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[isGreater]
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

		build := new(LinearBuilder)
		build.NoOutput = true
		compareColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		setColumnToConstant(build, compareColumn, colValue)

		isLess := ColumnIsLessOrEqualThanConstant(build, compareColumn, checkValue, rangeHigh, equalDelta)

		if setBool != nil {
			setColumnToConstant(build, isLess, *setBool)
		}
		solution, _ := build.RunHighs()
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

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
	maxValue := 100.0
	rangeHigh := 200.0
	equalDelta := 1.0

	testValues := func(loValue, constValue, hiValue float64, setBool *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: lo=%f const=%f hi=%f", loValue, constValue, hiValue)
		if setBool != nil {
			test.Logf("bool=%f", *setBool)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		loColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		hiColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		boolColumn := build.CreateColumnBool(nil)
		setColumnToConstant(build, loColumn, loValue)
		setColumnToConstant(build, hiColumn, hiValue)

		ConstantIsBetweenColumns(build, loColumn, hiColumn, boolColumn, constValue, rangeHigh, equalDelta)

		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}
		solution, _ := build.RunHighs()
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
	}

	zero := ptr(0.0)
	one := ptr(1.0)

	// logic lo <= const <= hi
	testValues(49, 49, 51, nil, highs.ModelStatusOptimal, one)
	testValues(49, 49, 51, zero, highs.ModelStatusInfeasible, nil)
	testValues(49, 49, 51, one, highs.ModelStatusOptimal, one)
	testValues(49, 50, 51, nil, highs.ModelStatusOptimal, one)
	testValues(49, 50, 51, zero, highs.ModelStatusInfeasible, nil)
	testValues(49, 50, 51, one, highs.ModelStatusOptimal, one)
	testValues(49, 51, 51, nil, highs.ModelStatusOptimal, one)
	testValues(49, 51, 51, zero, highs.ModelStatusInfeasible, nil)
	testValues(49, 51, 51, one, highs.ModelStatusOptimal, one)
	testValues(50, 50, 50, nil, highs.ModelStatusOptimal, one)
	testValues(50, 50, 50, zero, highs.ModelStatusInfeasible, nil)
	testValues(50, 50, 50, one, highs.ModelStatusOptimal, one)

	testValues(49, 48, 51, nil, highs.ModelStatusOptimal, zero)
	testValues(49, 48, 51, zero, highs.ModelStatusOptimal, zero)
	testValues(49, 48, 51, one, highs.ModelStatusInfeasible, nil)
	testValues(49, 52, 51, nil, highs.ModelStatusOptimal, zero)
	testValues(49, 52, 51, zero, highs.ModelStatusOptimal, zero)
	testValues(49, 52, 51, one, highs.ModelStatusInfeasible, nil)

	testValues(50, 48, 50, nil, highs.ModelStatusOptimal, zero)
	testValues(50, 48, 50, zero, highs.ModelStatusOptimal, zero)
	testValues(50, 48, 50, one, highs.ModelStatusInfeasible, nil)
	testValues(50, 52, 50, nil, highs.ModelStatusOptimal, zero)
	testValues(50, 52, 50, zero, highs.ModelStatusOptimal, zero)
	testValues(50, 52, 50, one, highs.ModelStatusInfeasible, nil)

	testValues(50, 49, 50, nil, highs.ModelStatusOptimal, zero)
	testValues(50, 49, 50, zero, highs.ModelStatusOptimal, zero)
	testValues(50, 49, 50, one, highs.ModelStatusInfeasible, nil)
	testValues(50, 51, 50, nil, highs.ModelStatusOptimal, zero)
	testValues(50, 51, 50, zero, highs.ModelStatusOptimal, zero)
	testValues(50, 51, 50, one, highs.ModelStatusInfeasible, nil)

	testValues(51, 50, 49, nil, highs.ModelStatusInfeasible, nil)
	testValues(51, 50, 49, zero, highs.ModelStatusInfeasible, nil)
	testValues(51, 50, 49, one, highs.ModelStatusInfeasible, nil)
	testValues(50, 50, 49, nil, highs.ModelStatusInfeasible, nil)
	testValues(50, 50, 49, zero, highs.ModelStatusInfeasible, nil)
	testValues(50, 50, 49, one, highs.ModelStatusInfeasible, nil)
	testValues(50, 49, 49, nil, highs.ModelStatusInfeasible, nil)
	testValues(50, 49, 49, zero, highs.ModelStatusInfeasible, nil)
	testValues(50, 49, 49, one, highs.ModelStatusInfeasible, nil)
}

func TestColumnIsNotBetweenConstantsVerify(test *testing.T) {
	maxValue := 100.0
	rangeHigh := 200.0

	testValues := func(loValue, checkValue, hiValue float64, expectStatus highs.ModelStatus) {
		test.Logf("CASE: lo=%f check=%f hi=%f", loValue, checkValue, hiValue)

		build := new(LinearBuilder)
		build.NoOutput = true
		checkColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		setColumnToConstant(build, checkColumn, checkValue)

		{
			defer func() {
				if err := recover(); err != nil {
					test.Log(err)
					if expectStatus != -1 {
						test.Fatal("unexpected panic")
					}
				}
			}()
			ColumnIsNotBetweenConstantsVerify(build, checkColumn, loValue, hiValue, rangeHigh)
		}

		if expectStatus == -1 {
			test.Fatal("expected panic")
		}

		solution, _ := build.RunHighs()
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		test.Logf("%s\n", solution.Status.String())
		assertEqual(expectStatus, solution.Status, test)
	}

	// logic not(lo < check < hi)

	// standard rejection, within range
	testValues(49, 50, 51, highs.ModelStatusInfeasible)
	testValues(40, 50, 60, highs.ModelStatusInfeasible)

	// testValues(50, 50, 50, highs.ModelStatusInfeasible)

	// equal to high or low = ok
	testValues(49, 49, 51, highs.ModelStatusOptimal)
	testValues(49, 51, 51, highs.ModelStatusOptimal)

	// outside range normal
	testValues(49, 47, 51, highs.ModelStatusOptimal)
	testValues(49, 48, 51, highs.ModelStatusOptimal)
	testValues(49, 52, 51, highs.ModelStatusOptimal)
	testValues(49, 53, 51, highs.ModelStatusOptimal)

	// outside equal pair
	testValues(50, 48, 50, highs.ModelStatusOptimal)
	testValues(50, 49, 50, highs.ModelStatusOptimal)
	testValues(50, 51, 50, highs.ModelStatusOptimal)
	testValues(50, 52, 50, highs.ModelStatusOptimal)

	// out of order hi/lo
	testValues(51, 50, 49, -1)
	testValues(50, 50, 49, -1)
	testValues(50, 49, 49, -1)
}

func TestColumnIsGreaterOrEqualColumn(test *testing.T) {
	maxValue := 100.0
	rangeHigh := 200.0
	equalDelta := 1.0

	testValues := func(oneValue, twoValue float64, setBool *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f two=%f", oneValue, twoValue)
		if setBool != nil {
			test.Logf("bool=%f", *setBool)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		twoColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		boolColumn := build.CreateColumnBool(nil)
		setColumnToConstant(build, oneColumn, oneValue)
		setColumnToConstant(build, twoColumn, twoValue)

		ColumnIsGreaterOrEqualColumn(build, oneColumn, twoColumn, boolColumn, rangeHigh, equalDelta)

		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}
		solution, _ := build.RunHighs()
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, oneValue >= twoValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
		if solution.Status.HasSolution() {
			assertEqual(oneValue >= twoValue, FloatEqualsOne(boolOutput), test)
		}
	}

	zero := ptr(0.0)
	one := ptr(1.0)

	// logic one >= two
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

func TestColumnIsLessOrEqualColumn(test *testing.T) {
	maxValue := 100.0
	rangeHigh := 200.0
	equalDelta := 1.0

	testValues := func(oneValue, twoValue float64, setBool *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f two=%f", oneValue, twoValue)
		if setBool != nil {
			test.Logf("bool=%f", *setBool)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		twoColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		boolColumn := build.CreateColumnBool(nil)
		setColumnToConstant(build, oneColumn, oneValue)
		setColumnToConstant(build, twoColumn, twoValue)

		ColumnIsLessOrEqualColumn(build, oneColumn, twoColumn, boolColumn, rangeHigh, equalDelta)

		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}
		solution, _ := build.RunHighs()
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, oneValue >= twoValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
		if solution.Status.HasSolution() {
			assertEqual(oneValue <= twoValue, FloatEqualsOne(boolOutput), test)
		}
	}

	zero := ptr(0.0)
	one := ptr(1.0)

	// logic one <= two
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

func TestColumnIsGreaterThanColumnEqualityFree(test *testing.T) {
	maxValue := 100.0
	rangeHigh := 200.0

	testValues := func(oneValue, twoValue float64, setBool *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f two=%f", oneValue, twoValue)
		if setBool != nil {
			test.Logf("bool=%f", *setBool)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		twoColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		boolColumn := build.CreateColumnBool(nil)
		setColumnToConstant(build, oneColumn, oneValue)
		setColumnToConstant(build, twoColumn, twoValue)

		ColumnIsGreaterThanColumnEqualityFree(build, oneColumn, twoColumn, boolColumn, rangeHigh)

		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}
		solution, _ := build.RunHighs()
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, oneValue >= twoValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
	}

	zero := ptr(0.0)
	one := ptr(1.0)

	// logic one > two
	testValues(49, 50, nil, highs.ModelStatusOptimal, zero)
	testValues(49, 50, zero, highs.ModelStatusOptimal, zero)
	testValues(49, 50, one, highs.ModelStatusInfeasible, nil)

	testValues(50, 50, nil, highs.ModelStatusOptimal, nil)
	testValues(50, 50, zero, highs.ModelStatusOptimal, zero)
	testValues(50, 50, one, highs.ModelStatusOptimal, one)

	testValues(51, 50, nil, highs.ModelStatusOptimal, one)
	testValues(51, 50, zero, highs.ModelStatusInfeasible, nil)
	testValues(51, 50, one, highs.ModelStatusOptimal, one)
}

func TestColumnIsLessThanColumnEqualityFree(test *testing.T) {
	maxValue := 100.0
	rangeHigh := 200.0

	testValues := func(oneValue, twoValue float64, setBool *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f two=%f", oneValue, twoValue)
		if setBool != nil {
			test.Logf("bool=%f", *setBool)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		twoColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		boolColumn := build.CreateColumnBool(nil)
		setColumnToConstant(build, oneColumn, oneValue)
		setColumnToConstant(build, twoColumn, twoValue)

		ColumnIsLessThanColumnEqualityFree(build, oneColumn, twoColumn, boolColumn, rangeHigh)

		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}
		solution, _ := build.RunHighs()
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, oneValue >= twoValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
	}

	zero := ptr(0.0)
	one := ptr(1.0)

	// logic one < two
	testValues(49, 50, nil, highs.ModelStatusOptimal, one)
	testValues(49, 50, zero, highs.ModelStatusInfeasible, nil)
	testValues(49, 50, one, highs.ModelStatusOptimal, one)

	testValues(50, 50, nil, highs.ModelStatusOptimal, nil)
	testValues(50, 50, zero, highs.ModelStatusOptimal, zero)
	testValues(50, 50, one, highs.ModelStatusOptimal, one)

	testValues(51, 50, nil, highs.ModelStatusOptimal, zero)
	testValues(51, 50, zero, highs.ModelStatusOptimal, zero)
	testValues(51, 50, one, highs.ModelStatusInfeasible, nil)
}

func setColumnToConstant(build *LinearBuilder, column ColumnIndex, value float64) {
	row := ConstraintRow{}
	row.Add(column, 1)
	row.Build(build, value, value)
}

func ptr(value float64) *float64 {
	return &value
}

func assertEqual[T comparable](expect, actual T, test *testing.T) {
	test.Helper()
	if expect != actual {
		test.Fatalf("assertEqual failed expect=%v actual=%v", expect, actual)
	}
}
