package util_highs

import (
	"paladin_gearing_go/util"
	"testing"

	"github.com/bartolsthoorn/gohighs/highs"
)

func TestConstraintIfBoolCopyValueElseZero(test *testing.T) {
	rangeLow := -200.0
	rangeHigh := 200.0

	testValues := func(oneValue, toggleValue float64, outValueSet *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f toggle=%f", oneValue, toggleValue)
		if outValueSet != nil {
			test.Logf("out=%f", *outValueSet)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, InfNeg(), InfPos(), nil)
		toggleColumn := build.CreateColumnBool(nil)
		outColumn := build.CreateColumnGeneral(highs.Continuous, InfNeg(), InfPos(), nil)
		setColumnToConstant(build, oneColumn, oneValue)
		setColumnToConstant(build, toggleColumn, toggleValue)

		build.ConstraintCopyIfBoolElseZero(toggleColumn, oneColumn, outColumn, rangeLow, rangeHigh)

		if outValueSet != nil {
			setColumnToConstant(build, outColumn, *outValueSet)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outColumn]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqualFloat(*expectOutputValue, boolOutput, test)
		}
	}

	// basic copy with toggle enabled
	testValues(0, 1, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(5, 1, nil, highs.ModelStatusOptimal, new(5.0))
	testValues(-5.5, 1, nil, highs.ModelStatusOptimal, new(-5.5))

	// confirm lock-in with toggle enabled
	testValues(0, 1, new(-0.1), highs.ModelStatusInfeasible, nil)
	testValues(0, 1, new(0.1), highs.ModelStatusInfeasible, nil)
	testValues(5, 1, new(4.9), highs.ModelStatusInfeasible, nil)
	testValues(5, 1, new(-5.0), highs.ModelStatusInfeasible, nil)
	testValues(5, 1, new(5.1), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 1, new(-5.4), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 1, new(5.5), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 1, new(-5.6), highs.ModelStatusInfeasible, nil)

	// copy zero with toggle disabled
	testValues(0, 0, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(5, 0, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(-5.5, 0, nil, highs.ModelStatusOptimal, new(0.0))

	// confirm still lock-in with toggle disabled
	testValues(0, 0, new(-0.1), highs.ModelStatusInfeasible, nil)
	testValues(0, 0, new(0.1), highs.ModelStatusInfeasible, nil)
	testValues(5, 0, new(4.9), highs.ModelStatusInfeasible, nil)
	testValues(5, 0, new(-5.0), highs.ModelStatusInfeasible, nil)
	testValues(5, 0, new(5.1), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 0, new(-5.4), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 0, new(5.5), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 0, new(-5.6), highs.ModelStatusInfeasible, nil)
}

func TestConstraintIfBoolCopy(test *testing.T) {
	rangeHigh := 200.0

	testValues := func(oneValue, oneCoeff, toggleValue float64, outValueSet *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f co=%f toggle=%f", oneValue, oneCoeff, toggleValue)
		if outValueSet != nil {
			test.Logf("out=%f", *outValueSet)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, InfNeg(), InfPos(), nil)
		toggleColumn := build.CreateColumnBool(nil)
		outColumn := build.CreateColumnGeneral(highs.Continuous, InfNeg(), InfPos(), nil)
		setColumnToConstant(build, oneColumn, oneValue)
		setColumnToConstant(build, toggleColumn, toggleValue)

		build.ConstraintCopyIfBool(toggleColumn, oneColumn, oneCoeff, outColumn, rangeHigh)

		if outValueSet != nil {
			setColumnToConstant(build, outColumn, *outValueSet)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outColumn]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqualFloat(*expectOutputValue, boolOutput, test)
		}
	}

	// basic copy with toggle enabled
	testValues(0, 1, 1, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(5, 1, 1, nil, highs.ModelStatusOptimal, new(5.0))
	testValues(-5.5, 1, 1, nil, highs.ModelStatusOptimal, new(-5.5))
	testValues(5, 0.5, 1, nil, highs.ModelStatusOptimal, new(2.5))
	testValues(5, 0, 1, nil, highs.ModelStatusOptimal, new(0.0))

	// confirm lock-in with toggle enabled
	testValues(0, 1, 1, new(-0.1), highs.ModelStatusInfeasible, nil)
	testValues(0, 1, 1, new(0.1), highs.ModelStatusInfeasible, nil)
	testValues(5, 1, 1, new(4.9), highs.ModelStatusInfeasible, nil)
	testValues(5, 1, 1, new(-5.0), highs.ModelStatusInfeasible, nil)
	testValues(5, 1, 1, new(5.1), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 1, 1, new(-5.4), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 1, 1, new(5.5), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 1, 1, new(-5.6), highs.ModelStatusInfeasible, nil)
	testValues(5, 0.5, 1, new(2.4), highs.ModelStatusInfeasible, nil)
	testValues(5, 0.5, 1, new(-2.5), highs.ModelStatusInfeasible, nil)
	testValues(5, 0.5, 1, new(2.6), highs.ModelStatusInfeasible, nil)
	testValues(5, 0, 1, new(-0.1), highs.ModelStatusInfeasible, nil)
	testValues(5, 0, 1, new(0.1), highs.ModelStatusInfeasible, nil)

	// free copy with toggle disabled. doesn't really have a true default as this would imply
	// defaultValue := -rangeHigh
	// testValues(0, 1, 0, nil, highs.ModelStatusOptimal, ptr(defaultValue))
	// testValues(5, 1, 0, nil, highs.ModelStatusOptimal, ptr(defaultValue))
	// testValues(-5.5, 1, 0, nil, highs.ModelStatusOptimal, ptr(defaultValue))
	// testValues(5, 0.5, 0, nil, highs.ModelStatusOptimal, ptr(defaultValue))
	// testValues(5, 0, 0, nil, highs.ModelStatusOptimal, ptr(defaultValue))
	testValues(0, 1, 0, nil, highs.ModelStatusOptimal, nil)
	testValues(5, 1, 0, nil, highs.ModelStatusOptimal, nil)
	testValues(-5.5, 1, 0, nil, highs.ModelStatusOptimal, nil)
	testValues(5, 0.5, 0, nil, highs.ModelStatusOptimal, nil)
	testValues(5, 0, 0, nil, highs.ModelStatusOptimal, nil)

	// confirm free with toggle disabled
	testValues(0, 1, 0, new(-0.1), highs.ModelStatusOptimal, new(-0.1))
	testValues(0, 1, 0, new(0.1), highs.ModelStatusOptimal, new(0.1))
	testValues(5, 1, 0, new(4.9), highs.ModelStatusOptimal, new(4.9))
	testValues(5, 1, 0, new(-5.0), highs.ModelStatusOptimal, new(-5.0))
	testValues(5, 1, 0, new(5.1), highs.ModelStatusOptimal, new(5.1))
	testValues(-5.5, 1, 0, new(-5.4), highs.ModelStatusOptimal, new(-5.4))
	testValues(-5.5, 1, 0, new(5.5), highs.ModelStatusOptimal, new(5.5))
	testValues(-5.5, 1, 0, new(-5.6), highs.ModelStatusOptimal, new(-5.6))
	testValues(5, 0.5, 0, new(2.4), highs.ModelStatusOptimal, new(2.4))
	testValues(5, 0.5, 0, new(-2.5), highs.ModelStatusOptimal, new(-2.5))
	testValues(5, 0.5, 0, new(2.6), highs.ModelStatusOptimal, new(2.6))
	testValues(5, 0, 0, new(-0.1), highs.ModelStatusOptimal, new(-0.1))
	testValues(5, 0, 0, new(0.1), highs.ModelStatusOptimal, new(0.1))
}

func TestConstraintAndBuilder(test *testing.T) {
	testValues := func(values []float64, out *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE:")
		for _, val := range values {
			test.Logf("%f", val)
		}
		if out != nil {
			test.Logf("out=%f", *out)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		outCol := build.CreateColumnBool(nil)

		and := ConstraintAndBuilder{}
		for _, val := range values {
			col := build.CreateColumnBool(nil)
			setColumnToConstant(build, col, val)
			and.AddInput(col)
		}
		and.SetOutput(outCol)
		and.Build(build)

		if out != nil {
			setColumnToConstant(build, outCol, *out)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outCol]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
	}

	testValues(makeSlice(), nil, highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1), nil, highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0, 0), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0, 1), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1, 0), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1, 1), nil, highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0, 0, 0), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0, 0, 1), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0, 1, 0), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0, 1, 1), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1, 0, 0), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1, 0, 1), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1, 1, 0), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1, 1, 1), nil, highs.ModelStatusOptimal, new(1.0))

	testValues(makeSlice(), new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1), new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0, 0), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0, 1), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1, 0), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1, 1), new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0, 0, 0), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0, 0, 1), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0, 1, 0), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0, 1, 1), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1, 0, 0), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1, 0, 1), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1, 1, 0), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1, 1, 1), new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(makeSlice(), new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1), new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 0), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 1), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 0), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 1), new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 0, 0), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 0, 1), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 1, 0), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 1, 1), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 0, 0), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 0, 1), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 1, 0), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 1, 1), new(0.0), highs.ModelStatusInfeasible, nil)
}

func TestConstraintOrBuilder(test *testing.T) {
	testValues := func(values []float64, out *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE:")
		for _, val := range values {
			test.Logf("%f", val)
		}
		if out != nil {
			test.Logf("out=%f", *out)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		outCol := build.CreateColumnBool(nil)

		or := ConstraintOrBuilder{}
		for _, val := range values {
			col := build.CreateColumnBool(nil)
			setColumnToConstant(build, col, val)
			or.AddInput(col)
		}
		or.SetOutput(outCol)
		or.Build(build)

		if out != nil {
			setColumnToConstant(build, outCol, *out)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outCol]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
	}

	testValues(makeSlice(), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1), nil, highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0, 0), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0, 1), nil, highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(1, 0), nil, highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(1, 1), nil, highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0, 0, 0), nil, highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0, 0, 1), nil, highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0, 1, 0), nil, highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0, 1, 1), nil, highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(1, 0, 0), nil, highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(1, 0, 1), nil, highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(1, 1, 0), nil, highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(1, 1, 1), nil, highs.ModelStatusOptimal, new(1.0))

	testValues(makeSlice(), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(1), new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0, 0), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0, 1), new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(1, 0), new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(1, 1), new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0, 0, 0), new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(makeSlice(0, 0, 1), new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0, 1, 0), new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(0, 1, 1), new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(1, 0, 0), new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(1, 0, 1), new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(1, 1, 0), new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(makeSlice(1, 1, 1), new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(makeSlice(), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1), new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 0), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 1), new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 0), new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 1), new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 0, 0), new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 0, 1), new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 1, 0), new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 1, 1), new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 0, 0), new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 0, 1), new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 1, 0), new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 1, 1), new(0.0), highs.ModelStatusInfeasible, nil)
}

func makeSlice(vals ...float64) []float64 {
	return vals
}

func TestConstraintNot(test *testing.T) {
	testValues := func(one float64, out *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f", one)
		if out != nil {
			test.Logf("out=%f", *out)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneCol := build.CreateColumnBool(nil)
		outCol := build.CreateColumnBool(nil)
		setColumnToConstant(build, oneCol, one)

		build.ConstraintNot(oneCol, outCol)

		if out != nil {
			setColumnToConstant(build, outCol, *out)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outCol]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
	}

	testValues(0, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(0, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(0, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(1, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(1, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(1, new(1.0), highs.ModelStatusInfeasible, nil)
}

func TestNotAsColumn(test *testing.T) {
	testValues := func(one float64, out *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f", one)
		if out != nil {
			test.Logf("out=%f", *out)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneCol := build.CreateColumnBool(nil)
		setColumnToConstant(build, oneCol, one)

		outCol := build.NotAsColumn(oneCol)

		if out != nil {
			setColumnToConstant(build, outCol, *out)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outCol]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
	}

	testValues(0, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(0, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(0, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(1, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(1, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(1, new(1.0), highs.ModelStatusInfeasible, nil)
}

func TestAbsoluteValue(test *testing.T) {
	maxValue := 100.0

	testValues := func(oneValue float64, outValueSet *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f", oneValue)
		if outValueSet != nil {
			test.Logf("out=%f", *outValueSet)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		outColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		setColumnToConstant(build, oneColumn, oneValue)

		build.AbsoluteValue(oneColumn, outColumn)

		if outValueSet != nil {
			setColumnToConstant(build, outColumn, *outValueSet)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outColumn]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqualFloat(*expectOutputValue, boolOutput, test)
		}
	}

	// standard positive differences
	testValues(0, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(1, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(-1, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(13.1, nil, highs.ModelStatusOptimal, new(13.1))
	testValues(-13.1, nil, highs.ModelStatusOptimal, new(13.1))

	// confirm forced minimum, goes equals, then free when higher
	testValues(1, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(1, new(0.9), highs.ModelStatusInfeasible, nil)
	testValues(1, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(1, new(1.1), highs.ModelStatusOptimal, new(1.1))
	testValues(1, new(77.0), highs.ModelStatusOptimal, new(77.0))
	testValues(-1, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(-1, new(0.9), highs.ModelStatusInfeasible, nil)
	testValues(-1, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(-1, new(1.1), highs.ModelStatusOptimal, new(1.1))
	testValues(-1, new(77.0), highs.ModelStatusOptimal, new(77.0))
}

func TestAbsoluteValueNonFree_NeedMIP(test *testing.T) {
	maxValue := 100.0
	highRange := 200.0

	testValues := func(oneValue float64, outValueSet *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f", oneValue)
		if outValueSet != nil {
			test.Logf("out=%f", *outValueSet)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		outColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		setColumnToConstant(build, oneColumn, oneValue)

		build.AbsoluteValueNonFree_NeedMIP(oneColumn, outColumn, highRange, "")

		if outValueSet != nil {
			setColumnToConstant(build, outColumn, *outValueSet)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outColumn]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqualFloat(*expectOutputValue, boolOutput, test)
		}
	}

	// standard positive differences
	testValues(0, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(1, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(-1, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(13.1, nil, highs.ModelStatusOptimal, new(13.1))
	testValues(-13.1, nil, highs.ModelStatusOptimal, new(13.1))

	// confirm forced minimum and maximum
	testValues(1, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(1, new(0.9), highs.ModelStatusInfeasible, nil)
	testValues(1, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(1, new(1.1), highs.ModelStatusInfeasible, nil)
	testValues(1, new(77.0), highs.ModelStatusInfeasible, nil)
	testValues(-1, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(-1, new(0.9), highs.ModelStatusInfeasible, nil)
	testValues(-1, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(-1, new(1.1), highs.ModelStatusInfeasible, nil)
	testValues(-1, new(77.0), highs.ModelStatusInfeasible, nil)
}

func TestAbsoluteValueFromDiffTwoVars(test *testing.T) {
	maxValue := 100.0

	testValues := func(oneValue, oneCoeff, twoValue, twoCoeff float64, outValueSet *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f co=%f two=%f co=%f", oneValue, oneCoeff, twoValue, twoCoeff)
		if outValueSet != nil {
			test.Logf("out=%f", *outValueSet)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		twoColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		outColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		setColumnToConstant(build, oneColumn, oneValue)
		setColumnToConstant(build, twoColumn, twoValue)

		build.AbsoluteValueFromDiffTwoVars(oneColumn, oneCoeff, twoColumn, twoCoeff, outColumn, "")

		if outValueSet != nil {
			setColumnToConstant(build, outColumn, *outValueSet)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outColumn]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqualFloat(*expectOutputValue, boolOutput, test)
		}
	}

	// standard positive differences
	testValues(3, 1, 3, 1, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(3, 1, 4, 1, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(4, 1, 3, 1, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(25.4, 1, 12.3, 1, nil, highs.ModelStatusOptimal, new(13.1))
	testValues(12.3, 1, 25.4, 1, nil, highs.ModelStatusOptimal, new(13.1))

	// negatives
	testValues(-55, 1, -44, 1, nil, highs.ModelStatusOptimal, new(11.0))
	testValues(-44, 1, -55, 1, nil, highs.ModelStatusOptimal, new(11.0))
	testValues(-44, 1, 10, 1, nil, highs.ModelStatusOptimal, new(54.0))
	testValues(10, 1, -44, 1, nil, highs.ModelStatusOptimal, new(54.0))

	// zeros
	testValues(0, 1, -44, 1, nil, highs.ModelStatusOptimal, new(44.0))
	testValues(-44, 1, 0, 1, nil, highs.ModelStatusOptimal, new(44.0))
	testValues(44, 1, 0, 1, nil, highs.ModelStatusOptimal, new(44.0))
	testValues(0, 1, 44, 1, nil, highs.ModelStatusOptimal, new(44.0))

	// coefficients
	testValues(5, 2, 9, 1, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(5, 2, 9, -1, nil, highs.ModelStatusOptimal, new(19.0))
	testValues(5, 0.5, 9, 1, nil, highs.ModelStatusOptimal, new(6.5))
	testValues(3, 0.1, 4, 0.1, nil, highs.ModelStatusOptimal, new(0.1))

	// confirm forced minimum, goes equals, then free when higher
	testValues(3, 1, 4, 1, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, 1, new(0.9), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, 1, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(3, 1, 4, 1, new(1.1), highs.ModelStatusOptimal, new(1.1))
	testValues(3, 1, 4, 1, new(77.0), highs.ModelStatusOptimal, new(77.0))
}

func TestAbsoluteValueFromDiffTwoVarsNonFree(test *testing.T) {
	maxValue := 100.0
	highRange := 200.0

	testValues := func(oneValue, oneCoeff, twoValue, twoCoeff float64, outValueSet *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f co=%f two=%f co=%f", oneValue, oneCoeff, twoValue, twoCoeff)
		if outValueSet != nil {
			test.Logf("out=%f", *outValueSet)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		twoColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		outColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		setColumnToConstant(build, oneColumn, oneValue)
		setColumnToConstant(build, twoColumn, twoValue)

		build.AbsoluteValueFromDiffTwoVarsNonFree(oneColumn, oneCoeff, twoColumn, twoCoeff, outColumn, highRange, "")

		if outValueSet != nil {
			setColumnToConstant(build, outColumn, *outValueSet)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outColumn]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqualFloat(*expectOutputValue, boolOutput, test)
		}
	}

	// standard positive differences
	testValues(3, 1, 3, 1, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(3, 1, 4, 1, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(4, 1, 3, 1, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(25.4, 1, 12.3, 1, nil, highs.ModelStatusOptimal, new(13.1))
	testValues(12.3, 1, 25.4, 1, nil, highs.ModelStatusOptimal, new(13.1))

	// negatives
	testValues(-55, 1, -44, 1, nil, highs.ModelStatusOptimal, new(11.0))
	testValues(-44, 1, -55, 1, nil, highs.ModelStatusOptimal, new(11.0))
	testValues(-44, 1, 10, 1, nil, highs.ModelStatusOptimal, new(54.0))
	testValues(10, 1, -44, 1, nil, highs.ModelStatusOptimal, new(54.0))

	// zeros
	testValues(0, 1, -44, 1, nil, highs.ModelStatusOptimal, new(44.0))
	testValues(-44, 1, 0, 1, nil, highs.ModelStatusOptimal, new(44.0))
	testValues(44, 1, 0, 1, nil, highs.ModelStatusOptimal, new(44.0))
	testValues(0, 1, 44, 1, nil, highs.ModelStatusOptimal, new(44.0))

	// coefficients
	testValues(5, 2, 9, 1, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(5, 2, 9, -1, nil, highs.ModelStatusOptimal, new(19.0))
	testValues(5, 0.5, 9, 1, nil, highs.ModelStatusOptimal, new(6.5))
	testValues(3, 0.1, 4, 0.1, nil, highs.ModelStatusOptimal, new(0.1))

	// confirm forced minimum, goes equals, then infeasible when higher too
	testValues(3, 1, 4, 1, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, 1, new(0.9), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, 1, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(3, 1, 4, 1, new(1.1), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, 1, new(77.0), highs.ModelStatusInfeasible, nil)

	// same with values flipped
	testValues(4, 1, 3, 1, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 3, 1, new(0.9), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 3, 1, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(4, 1, 3, 1, new(1.1), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 3, 1, new(77.0), highs.ModelStatusInfeasible, nil)

	// and with a zero
	testValues(4, 1, 4, 1, new(-1.0), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 4, 1, new(-0.1), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 4, 1, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(4, 1, 4, 1, new(0.1), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 4, 1, new(77.0), highs.ModelStatusInfeasible, nil)

	// flipped sign was buggy
	testValues(3, 1, 4, 1, new(-1.0), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 3, 1, new(-1.0), highs.ModelStatusInfeasible, nil)
}

func TestAbsoluteValueFromDiffOneToConst(test *testing.T) {
	maxValue := 100.0

	testValues := func(oneValue, oneCoeff, constValue float64, outValueSet *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f co=%f const=%f", oneValue, oneCoeff, constValue)
		if outValueSet != nil {
			test.Logf("out=%f", *outValueSet)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		outColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		setColumnToConstant(build, oneColumn, oneValue)

		build.AbsoluteValueFromDiffOneToConst(oneColumn, oneCoeff, constValue, outColumn, "")

		if outValueSet != nil {
			setColumnToConstant(build, outColumn, *outValueSet)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outColumn]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqualFloat(*expectOutputValue, boolOutput, test)
		}
	}

	// standard positive differences
	testValues(3, 1, 3, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(3, 1, 4, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(4, 1, 3, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(25.4, 1, 12.3, nil, highs.ModelStatusOptimal, new(13.1))
	testValues(12.3, 1, 25.4, nil, highs.ModelStatusOptimal, new(13.1))

	// negatives
	testValues(-55, 1, -44, nil, highs.ModelStatusOptimal, new(11.0))
	testValues(-44, 1, -55, nil, highs.ModelStatusOptimal, new(11.0))
	testValues(-44, 1, 10, nil, highs.ModelStatusOptimal, new(54.0))
	testValues(10, 1, -44, nil, highs.ModelStatusOptimal, new(54.0))

	// zeros
	testValues(0, 1, -44, nil, highs.ModelStatusOptimal, new(44.0))
	testValues(-44, 1, 0, nil, highs.ModelStatusOptimal, new(44.0))
	testValues(44, 1, 0, nil, highs.ModelStatusOptimal, new(44.0))
	testValues(0, 1, 44, nil, highs.ModelStatusOptimal, new(44.0))

	// confirm forced minimum, goes equals, then free when higher
	testValues(3, 1, 4, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, new(0.9), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(3, 1, 4, new(1.1), highs.ModelStatusOptimal, new(1.1))
	testValues(3, 1, 4, new(77.0), highs.ModelStatusOptimal, new(77.0))
}

func TestAbsoluteValueBeforeDiffOneToConst(test *testing.T) {
	maxValue := 100.0

	testValues := func(oneValue, oneCoeff, constValue float64, outValueSet *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f co=%f const=%f", oneValue, oneCoeff, constValue)
		if outValueSet != nil {
			test.Logf("out=%f", *outValueSet)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		outColumn := build.CreateColumnGeneral(highs.Continuous, -maxValue, maxValue, nil)
		setColumnToConstant(build, oneColumn, oneValue)

		build.AbsoluteValueBeforeDiffOneToConst(oneColumn, oneCoeff, constValue, outColumn, "")

		if outValueSet != nil {
			setColumnToConstant(build, outColumn, *outValueSet)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outColumn]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqualFloat(*expectOutputValue, boolOutput, test)
		}
	}

	// standard positive differences
	testValues(3, 1, 3, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(3, 1, 4, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(4, 1, 3, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(25.4, 1, 12.3, nil, highs.ModelStatusOptimal, new(13.1))
	testValues(12.3, 1, 25.4, nil, highs.ModelStatusOptimal, new(13.1))

	testValues(-3, 1, 3, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(-3, 1, 4, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(4, -1, 3, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(-25.4, 1, 12.3, nil, highs.ModelStatusOptimal, new(13.1))
	testValues(-12.3, 1, 25.4, nil, highs.ModelStatusOptimal, new(13.1))

	// negatives
	testValues(-55, 1, -44, nil, highs.ModelStatusOptimal, new(99.0))
	testValues(-44, 1, -55, nil, highs.ModelStatusOptimal, new(99.0))
	testValues(-44, 1, 10, nil, highs.ModelStatusOptimal, new(34.0))
	testValues(10, 1, -44, nil, highs.ModelStatusOptimal, new(54.0))

	// zeros
	testValues(0, 1, -44, nil, highs.ModelStatusOptimal, new(44.0))
	testValues(-44, 1, 0, nil, highs.ModelStatusOptimal, new(44.0))
	testValues(44, 1, 0, nil, highs.ModelStatusOptimal, new(44.0))
	testValues(0, 1, 44, nil, highs.ModelStatusOptimal, new(44.0))

	// confirm forced minimum, goes equals, then free when higher
	testValues(3, 1, 4, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, new(0.9), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(3, 1, 4, new(1.1), highs.ModelStatusOptimal, new(1.1))
	testValues(3, 1, 4, new(77.0), highs.ModelStatusOptimal, new(77.0))
	testValues(-3, 1, 4, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(-3, 1, 4, new(0.9), highs.ModelStatusInfeasible, nil)
	testValues(-3, 1, 4, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(-3, 1, 4, new(1.1), highs.ModelStatusOptimal, new(1.1))
	testValues(-3, 1, 4, new(77.0), highs.ModelStatusOptimal, new(77.0))
}

func TestAbsoluteValueDiffTwoVarsThenDiffConst(test *testing.T) {
	maxValue := 100.0
	highRange := 200.0

	testValues := func(oneValue, oneCoeff, twoValue, twoCoeff, offset float64, outValueSet *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f co=%f two=%f co=%f off=%f", oneValue, oneCoeff, twoValue, twoCoeff, offset)
		if outValueSet != nil {
			test.Logf("out=%f", *outValueSet)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		build.Minimise = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, oneValue, oneValue, nil)
		twoColumn := build.CreateColumnGeneral(highs.Continuous, twoValue, twoValue, nil)
		outColumn := build.CreateColumnWithOutput(highs.Continuous, -maxValue, maxValue, 1, nil)

		build.AbsoluteValueDiffTwoVarsThenDiffConst_NeedMIP(oneColumn, oneCoeff, twoColumn, twoCoeff, outColumn, offset, highRange, "")

		if outValueSet != nil {
			setColumnToConstant(build, outColumn, *outValueSet)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outColumn]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqualFloat(*expectOutputValue, boolOutput, test)
		}
	}

	// standard positive differences
	testValues(3, 1, 3, 1, 0, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(3, 1, 4, 1, 1, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(4, 1, 3, 1, 3, nil, highs.ModelStatusOptimal, new(2.0))
	testValues(25.4, 1, 12.3, 1, -2, nil, highs.ModelStatusOptimal, new(15.1))
	testValues(12.3, 1, 25.4, 1, 11, nil, highs.ModelStatusOptimal, new(2.1))

	// negatives
	testValues(-55, 1, -44, 1, 0, nil, highs.ModelStatusOptimal, new(11.0))
	testValues(-44, 1, -55, 1, 1, nil, highs.ModelStatusOptimal, new(10.0))
	testValues(-44, 1, 10, 1, -1, nil, highs.ModelStatusOptimal, new(55.0))
	testValues(10, 1, -44, 1, 5, nil, highs.ModelStatusOptimal, new(49.0))
	testValues(3, 1, 4, 1, -10, nil, highs.ModelStatusOptimal, new(11.0))

	// zeros
	testValues(0, 1, -44, 1, 0, nil, highs.ModelStatusOptimal, new(44.0))
	testValues(-44, 1, 0, 1, 1, nil, highs.ModelStatusOptimal, new(43.0))
	testValues(44, 1, 0, 1, 7, nil, highs.ModelStatusOptimal, new(37.0))
	testValues(0, 1, 44, 1, -2.5, nil, highs.ModelStatusOptimal, new(46.5))

	// coefficients
	testValues(5, 2, 9, 1, 1, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(5, 2, 9, -1, 1, nil, highs.ModelStatusOptimal, new(18.0))
	testValues(5, 0.5, 9, 1, 1, nil, highs.ModelStatusOptimal, new(5.5))
	testValues(3, 0.1, 4, 0.1, 1, nil, highs.ModelStatusOptimal, new(0.9))

	// confirm forced minimum, goes equals, then free when higher
	testValues(3, 1, 5, 1, 1, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 5, 1, 1, new(0.9), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 5, 1, 1, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(3, 1, 5, 1, 1, new(1.1), highs.ModelStatusOptimal, new(1.1))
	testValues(3, 1, 5, 1, 1, new(77.0), highs.ModelStatusOptimal, new(77.0))
}

func TestAbsoluteValue_WithToggle(test *testing.T) {
	rangeHigh := 200.0

	testValues := func(oneValue, toggleValue float64, outValueSet *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f toggle=%f", oneValue, toggleValue)
		if outValueSet != nil {
			test.Logf("out=%f", *outValueSet)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, InfNeg(), InfPos(), nil)
		toggleColumn := build.CreateColumnBool(nil)
		outColumn := build.CreateColumnGeneral(highs.Continuous, InfNeg(), InfPos(), nil)
		setColumnToConstant(build, oneColumn, oneValue)
		setColumnToConstant(build, toggleColumn, toggleValue)

		build.AbsoluteValue_WithToggle(oneColumn, outColumn, toggleColumn, rangeHigh)

		if outValueSet != nil {
			setColumnToConstant(build, outColumn, *outValueSet)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outColumn]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqualFloat(*expectOutputValue, boolOutput, test)
		}
	}

	// standard positive differences (checking toggle=ON works like normal AbsolueValue)
	testValues(0, 1, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(1, 1, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(-1, 1, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(13.1, 1, nil, highs.ModelStatusOptimal, new(13.1))
	testValues(-13.1, 1, nil, highs.ModelStatusOptimal, new(13.1))

	// confirm forced minimum, goes equals, then free when higher (checking toggle=ON works like normal AbsolueValue)
	testValues(1, 1, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(1, 1, new(0.9), highs.ModelStatusInfeasible, nil)
	testValues(1, 1, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(1, 1, new(1.1), highs.ModelStatusOptimal, new(1.1))
	testValues(1, 1, new(77.0), highs.ModelStatusOptimal, new(77.0))

	// standard positive differences toggle=OFF defaults to zero
	testValues(0, 0, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(1, 0, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(-1, 0, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(13.1, 0, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(-13.1, 0, nil, highs.ModelStatusOptimal, new(0.0))

	// confirm toggle off lets anything go
	testValues(1, 0, new(0.0), highs.ModelStatusOptimal, nil)
	testValues(1, 0, new(0.9), highs.ModelStatusOptimal, nil)
	testValues(1, 0, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(1, 0, new(1.1), highs.ModelStatusOptimal, new(1.1))
	testValues(1, 0, new(77.0), highs.ModelStatusOptimal, new(77.0))
	testValues(-1, 0, new(0.0), highs.ModelStatusOptimal, nil)
	testValues(-1, 0, new(0.9), highs.ModelStatusOptimal, nil)
	testValues(-1, 0, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(-1, 0, new(1.1), highs.ModelStatusOptimal, new(1.1))
	testValues(-1, 0, new(77.0), highs.ModelStatusOptimal, new(77.0))
	testValues(0, 0, new(0.0), highs.ModelStatusOptimal, nil)
	testValues(0, 0, new(0.9), highs.ModelStatusOptimal, nil)
	testValues(0, 0, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(0, 0, new(1.1), highs.ModelStatusOptimal, new(1.1))
	testValues(0, 0, new(77.0), highs.ModelStatusOptimal, new(77.0))
}

func TestIsXor(test *testing.T) {
	testValues := func(one, two float64, out *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f two=%f", one, two)
		if out != nil {
			test.Logf("out=%f", *out)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneCol := build.CreateColumnBool(nil)
		twoCol := build.CreateColumnBool(nil)
		outCol := build.CreateColumnBool(nil)
		setColumnToConstant(build, oneCol, one)
		setColumnToConstant(build, twoCol, two)

		build.IsXor(oneCol, twoCol, outCol)

		if out != nil {
			setColumnToConstant(build, outCol, *out)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[outCol]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
	}

	// for condition not met output is free
	testValues(0, 0, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(0, 0, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(0, 0, new(1.0), highs.ModelStatusOptimal, new(1.0))

	// condition met, output forced 1.0
	testValues(0, 1, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(0, 1, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(0, 1, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(1, 0, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(1, 0, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(1, 0, new(1.0), highs.ModelStatusOptimal, new(1.0))

	// for condition not met output is free
	testValues(1, 1, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(1, 1, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(1, 1, new(1.0), highs.ModelStatusOptimal, new(1.0))
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

		isGreater := build.ColumnIsGreaterOrEqualThanConstant(compareColumn, checkValue, rangeHigh, equalDelta)

		if setBool != nil {
			setColumnToConstant(build, isGreater, *setBool)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[isGreater]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, colValue <= checkValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
		if solution.Status.HasSolution() {
			assertEqual(colValue >= checkValue, util.FloatEqualsOne(boolOutput), test)
		}
	}

	testValues(49, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(49, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(49, 50, new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(50, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(51, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(51, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(51, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))
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

		isLess := build.ColumnIsLessOrEqualThanConstant(compareColumn, checkValue, rangeHigh, equalDelta)

		if setBool != nil {
			setColumnToConstant(build, isLess, *setBool)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[isLess]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, colValue <= checkValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
		if solution.Status.HasSolution() {
			assertEqual(colValue <= checkValue, util.FloatEqualsOne(boolOutput), test)
		}
	}

	testValues(49, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(49, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(50, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(50, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(51, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(51, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(51, 50, new(1.0), highs.ModelStatusInfeasible, nil)
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

		build.ConstantIsBetweenColumns(loColumn, hiColumn, boolColumn, constValue, rangeHigh, equalDelta)

		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
	}

	// logic lo <= const <= hi
	testValues(49, 49, 51, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(49, 49, 51, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 49, 51, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(49, 50, 51, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(49, 50, 51, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 50, 51, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(49, 51, 51, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(49, 51, 51, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 51, 51, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(50, 50, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(50, 50, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 50, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(49, 48, 51, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(49, 48, 51, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(49, 48, 51, new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 52, 51, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(49, 52, 51, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(49, 52, 51, new(1.0), highs.ModelStatusInfeasible, nil)

	testValues(50, 48, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(50, 48, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(50, 48, 50, new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 52, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(50, 52, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(50, 52, 50, new(1.0), highs.ModelStatusInfeasible, nil)

	testValues(50, 49, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(50, 49, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(50, 49, 50, new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 51, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(50, 51, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(50, 51, 50, new(1.0), highs.ModelStatusInfeasible, nil)

	testValues(51, 50, 49, nil, highs.ModelStatusInfeasible, nil)
	testValues(51, 50, 49, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(51, 50, 49, new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 50, 49, nil, highs.ModelStatusInfeasible, nil)
	testValues(50, 50, 49, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 50, 49, new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 49, 49, nil, highs.ModelStatusInfeasible, nil)
	testValues(50, 49, 49, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 49, 49, new(1.0), highs.ModelStatusInfeasible, nil)
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
			build.ColumnIsNotBetweenConstantsVerify(checkColumn, loValue, hiValue, rangeHigh)
		}

		if expectStatus == -1 {
			test.Fatal("expected panic")
		}

		solution := runHighs(build, util.PrintRecorder_Testing(test))
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

func TestColumnIsBetweenConstants(test *testing.T) {
	maxValue := 100.0
	rangeHigh := 200.0
	equalDelta := 1.0

	testValues := func(loValue, checkValue, hiValue float64, setBool *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: lo=%f check=%f hi=%f", loValue, checkValue, hiValue)

		build := new(LinearBuilder)
		build.NoOutput = true
		checkColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		setColumnToConstant(build, checkColumn, checkValue)

		boolColumn := build.ColumnIsBetweenConstants(checkColumn, loValue, hiValue, rangeHigh, equalDelta)

		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f\n", solution.Status.String(), boolOutput)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
	}

	// standard accept, within range
	testValues(49, 50, 51, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(40, 50, 60, nil, highs.ModelStatusOptimal, new(1.0))

	// equal to high or low = ok
	testValues(49, 49, 51, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(49, 51, 51, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(51, 51, 51, nil, highs.ModelStatusOptimal, new(1.0))

	// outside range normal
	testValues(49, 47, 51, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(49, 48, 51, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(49, 52, 51, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(49, 53, 51, nil, highs.ModelStatusOptimal, new(0.0))

	// outside equal pair
	testValues(50, 48, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(50, 49, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(50, 51, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(50, 52, 50, nil, highs.ModelStatusOptimal, new(0.0))

	// all the above as forces
	testValues(49, 50, 51, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(40, 50, 60, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(49, 49, 51, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(49, 51, 51, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(51, 51, 51, new(1.0), highs.ModelStatusOptimal, new(1.0))
	testValues(49, 47, 51, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(49, 48, 51, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(49, 52, 51, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(49, 53, 51, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(50, 48, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(50, 49, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(50, 51, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(50, 52, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))

	// all the above as opposite forces
	testValues(49, 50, 51, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(40, 50, 60, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 49, 51, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 51, 51, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(51, 51, 51, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 47, 51, new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 48, 51, new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 52, 51, new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 53, 51, new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 48, 50, new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 49, 50, new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 51, 50, new(1.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 52, 50, new(1.0), highs.ModelStatusInfeasible, nil)
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

		build.ColumnIsGreaterOrEqualColumn(oneColumn, twoColumn, boolColumn, rangeHigh, equalDelta)

		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, oneValue >= twoValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
		if solution.Status.HasSolution() {
			assertEqual(oneValue >= twoValue, util.FloatEqualsOne(boolOutput), test)
		}
	}

	// logic one >= two
	testValues(49, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(49, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(49, 50, new(1.0), highs.ModelStatusInfeasible, nil)

	testValues(50, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(50, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(51, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(51, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(51, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))
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

		build.ColumnIsLessOrEqualColumn(oneColumn, twoColumn, boolColumn, rangeHigh, equalDelta)

		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, oneValue >= twoValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
		if solution.Status.HasSolution() {
			assertEqual(oneValue <= twoValue, util.FloatEqualsOne(boolOutput), test)
		}
	}

	// logic one <= two
	testValues(49, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(49, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(50, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(50, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(51, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(51, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(51, 50, new(1.0), highs.ModelStatusInfeasible, nil)
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

		build.ColumnIsGreaterThanColumnEqualityFree(oneColumn, twoColumn, boolColumn, rangeHigh)

		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, oneValue >= twoValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
	}

	// logic one > two
	testValues(49, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(49, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(49, 50, new(1.0), highs.ModelStatusInfeasible, nil)

	testValues(50, 50, nil, highs.ModelStatusOptimal, nil)
	testValues(50, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(50, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(51, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(51, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(51, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))
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

		build.ColumnIsLessThanColumnEqualityFree(oneColumn, twoColumn, boolColumn, rangeHigh)

		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}
		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, oneValue >= twoValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
	}

	// logic one < two
	testValues(49, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(49, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(50, 50, nil, highs.ModelStatusOptimal, nil)
	testValues(50, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(50, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(51, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(51, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(51, 50, new(1.0), highs.ModelStatusInfeasible, nil)
}

func TestColumnIsEqualConstant(test *testing.T) {
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
		boolColumn := build.CreateColumnBool(nil)
		setColumnToConstant(build, compareColumn, colValue)
		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}

		build.ColumnIsEqualConstant(compareColumn, boolColumn, checkValue, rangeHigh, equalDelta)

		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, colValue <= checkValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
		if solution.Status.HasSolution() {
			assertEqual(colValue == checkValue, util.FloatEqualsOne(boolOutput), test)
		}
	}

	testValues(49, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(49, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(49, 50, new(1.0), highs.ModelStatusInfeasible, nil)

	testValues(50, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(50, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(50, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(51, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(51, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(51, 50, new(1.0), highs.ModelStatusInfeasible, nil)
}

func TestColumnIsEqualConstant_OneWayEnforceNotSet(test *testing.T) {
	maxValue := 100.0
	rangeHigh := 200.0

	testValues := func(colValue, checkValue float64, setBool *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: col=%f check=%f", colValue, checkValue)
		if setBool != nil {
			test.Logf("bool=%f", *setBool)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		compareColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		boolColumn := build.CreateColumnBool(nil)
		setColumnToConstant(build, compareColumn, colValue)
		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}

		build.ColumnIsEqualConstant_OneWayEnforceNotSet(compareColumn, boolColumn, checkValue, rangeHigh)

		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, colValue <= checkValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
		if solution.Status.HasSolution() {
			//assertEqual(colValue == checkValue, util.FloatEqualsOne(boolOutput), test)
		}
	}

	testValues(49, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(49, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(49, 50, new(1.0), highs.ModelStatusInfeasible, nil)

	// lazy behaviour check
	testValues(50, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(50, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(50, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(51, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(51, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(51, 50, new(1.0), highs.ModelStatusInfeasible, nil)
}

func TestColumnIsNotEqualConstant(test *testing.T) {
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
		boolColumn := build.CreateColumnBool(nil)
		setColumnToConstant(build, compareColumn, colValue)
		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}

		build.ColumnIsNotEqualConstant(compareColumn, boolColumn, checkValue, rangeHigh, equalDelta)

		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, colValue <= checkValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
		if solution.Status.HasSolution() {
			assertEqual(colValue != checkValue, util.FloatEqualsOne(boolOutput), test)
		}
	}

	testValues(49, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(49, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(50, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(50, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(50, 50, new(1.0), highs.ModelStatusInfeasible, nil)

	testValues(51, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(51, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(51, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))
}

func TestColumnIsNotEqualConstant_OneWayEnforceNotSet(test *testing.T) {
	maxValue := 100.0
	rangeHigh := 200.0

	testValues := func(colValue, checkValue float64, setBool *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: col=%f check=%f", colValue, checkValue)
		if setBool != nil {
			test.Logf("bool=%f", *setBool)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		compareColumn := build.CreateColumnGeneral(highs.Continuous, 0, maxValue, nil)
		boolColumn := build.CreateColumnBool(nil)
		setColumnToConstant(build, compareColumn, colValue)
		if setBool != nil {
			setColumnToConstant(build, boolColumn, *setBool)
		}

		build.ColumnIsNotEqualConstant_OneWayEnforceNotSet(compareColumn, boolColumn, checkValue, rangeHigh)

		solution := runHighs(build, util.PrintRecorder_Testing(test))
		build.debugPrintColumnsForce(solution, util.PrintRecorder_Testing(test))

		boolOutput := solution.ColValues[boolColumn]
		test.Logf("%s %f %v\n", solution.Status.String(), boolOutput, colValue <= checkValue)
		assertEqual(expectStatus, solution.Status, test)
		if expectOutputValue != nil {
			assertEqual(*expectOutputValue, boolOutput, test)
		}
	}

	testValues(49, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(49, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(49, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	// lazy check
	testValues(50, 50, nil, highs.ModelStatusOptimal, new(0.0))
	testValues(50, 50, new(0.0), highs.ModelStatusOptimal, new(0.0))
	testValues(50, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))

	testValues(51, 50, nil, highs.ModelStatusOptimal, new(1.0))
	testValues(51, 50, new(0.0), highs.ModelStatusInfeasible, nil)
	testValues(51, 50, new(1.0), highs.ModelStatusOptimal, new(1.0))
}

func setColumnToConstant(build *LinearBuilder, column ColumnIndex, value float64) {
	row := ConstraintRow{}
	row.Add(column, 1)
	row.Build(build, value, value)
}

func assertEqual[T comparable](expect, actual T, test *testing.T) {
	test.Helper()
	if expect != actual {
		test.Fatalf("assertEqual failed expect=%v actual=%v", expect, actual)
	}
}

func assertEqualFloat(expect, actual float64, test *testing.T) {
	test.Helper()
	if !util.FloatsApproxEquals(expect, actual) {
		test.Fatalf("assertEqual failed expect=%v actual=%v", expect, actual)
	}
}

func runHighs(build *LinearBuilder, printer *util.PrintRecorder) *highs.Solution {
	build.Solver = Solver_Flexible
	return build.RunHighsFuture(nil).WaitForResultOrPanic().solution
}
