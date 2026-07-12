package util_highs

import (
	"paladin_gearing_go/util"
	"testing"

	"github.com/bartolsthoorn/gohighs/highs"
)

func TestContraintIfBoolCopyValueElseZero(test *testing.T) {
	rangeLow := -200.0
	rangeHigh := 200.0

	testValues := func(oneValue, toggleValue float64, outValueSet *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f toggle=%f", oneValue, toggleValue)
		if outValueSet != nil {
			test.Logf("out=%f", *outValueSet)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, C_MinusInf, C_PlusInf, nil)
		toggleColumn := build.CreateColumnBool(nil)
		outColumn := build.CreateColumnGeneral(highs.Continuous, C_MinusInf, C_PlusInf, nil)
		setColumnToConstant(build, oneColumn, oneValue)
		setColumnToConstant(build, toggleColumn, toggleValue)

		build.ContraintIfBoolCopyValueElseZero(toggleColumn, oneColumn, outColumn, rangeLow, rangeHigh)

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
	testValues(0, 1, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(5, 1, nil, highs.ModelStatusOptimal, ptr(5))
	testValues(-5.5, 1, nil, highs.ModelStatusOptimal, ptr(-5.5))

	// confirm lock-in with toggle enabled
	testValues(0, 1, ptr(-0.1), highs.ModelStatusInfeasible, nil)
	testValues(0, 1, ptr(0.1), highs.ModelStatusInfeasible, nil)
	testValues(5, 1, ptr(4.9), highs.ModelStatusInfeasible, nil)
	testValues(5, 1, ptr(-5), highs.ModelStatusInfeasible, nil)
	testValues(5, 1, ptr(5.1), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 1, ptr(-5.4), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 1, ptr(5.5), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 1, ptr(-5.6), highs.ModelStatusInfeasible, nil)

	// copy zero with toggle disabled
	testValues(0, 0, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(5, 0, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(-5.5, 0, nil, highs.ModelStatusOptimal, ptr(0))

	// confirm still lock-in with toggle disabled
	testValues(0, 0, ptr(-0.1), highs.ModelStatusInfeasible, nil)
	testValues(0, 0, ptr(0.1), highs.ModelStatusInfeasible, nil)
	testValues(5, 0, ptr(4.9), highs.ModelStatusInfeasible, nil)
	testValues(5, 0, ptr(-5), highs.ModelStatusInfeasible, nil)
	testValues(5, 0, ptr(5.1), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 0, ptr(-5.4), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 0, ptr(5.5), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 0, ptr(-5.6), highs.ModelStatusInfeasible, nil)
}

func TestContraintIfBoolCopy(test *testing.T) {
	rangeHigh := 200.0

	testValues := func(oneValue, oneCoeff, toggleValue float64, outValueSet *float64, expectStatus highs.ModelStatus, expectOutputValue *float64) {
		test.Logf("CASE: one=%f co=%f toggle=%f", oneValue, oneCoeff, toggleValue)
		if outValueSet != nil {
			test.Logf("out=%f", *outValueSet)
		}

		build := new(LinearBuilder)
		build.NoOutput = true
		oneColumn := build.CreateColumnGeneral(highs.Continuous, C_MinusInf, C_PlusInf, nil)
		toggleColumn := build.CreateColumnBool(nil)
		outColumn := build.CreateColumnGeneral(highs.Continuous, C_MinusInf, C_PlusInf, nil)
		setColumnToConstant(build, oneColumn, oneValue)
		setColumnToConstant(build, toggleColumn, toggleValue)

		build.ContraintIfBoolCopy(toggleColumn, oneColumn, oneCoeff, outColumn, rangeHigh)

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
	testValues(0, 1, 1, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(5, 1, 1, nil, highs.ModelStatusOptimal, ptr(5))
	testValues(-5.5, 1, 1, nil, highs.ModelStatusOptimal, ptr(-5.5))
	testValues(5, 0.5, 1, nil, highs.ModelStatusOptimal, ptr(2.5))
	testValues(5, 0, 1, nil, highs.ModelStatusOptimal, ptr(0))

	// confirm lock-in with toggle enabled
	testValues(0, 1, 1, ptr(-0.1), highs.ModelStatusInfeasible, nil)
	testValues(0, 1, 1, ptr(0.1), highs.ModelStatusInfeasible, nil)
	testValues(5, 1, 1, ptr(4.9), highs.ModelStatusInfeasible, nil)
	testValues(5, 1, 1, ptr(-5), highs.ModelStatusInfeasible, nil)
	testValues(5, 1, 1, ptr(5.1), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 1, 1, ptr(-5.4), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 1, 1, ptr(5.5), highs.ModelStatusInfeasible, nil)
	testValues(-5.5, 1, 1, ptr(-5.6), highs.ModelStatusInfeasible, nil)
	testValues(5, 0.5, 1, ptr(2.4), highs.ModelStatusInfeasible, nil)
	testValues(5, 0.5, 1, ptr(-2.5), highs.ModelStatusInfeasible, nil)
	testValues(5, 0.5, 1, ptr(2.6), highs.ModelStatusInfeasible, nil)
	testValues(5, 0, 1, ptr(-0.1), highs.ModelStatusInfeasible, nil)
	testValues(5, 0, 1, ptr(0.1), highs.ModelStatusInfeasible, nil)

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
	testValues(0, 1, 0, ptr(-0.1), highs.ModelStatusOptimal, ptr(-0.1))
	testValues(0, 1, 0, ptr(0.1), highs.ModelStatusOptimal, ptr(0.1))
	testValues(5, 1, 0, ptr(4.9), highs.ModelStatusOptimal, ptr(4.9))
	testValues(5, 1, 0, ptr(-5), highs.ModelStatusOptimal, ptr(-5))
	testValues(5, 1, 0, ptr(5.1), highs.ModelStatusOptimal, ptr(5.1))
	testValues(-5.5, 1, 0, ptr(-5.4), highs.ModelStatusOptimal, ptr(-5.4))
	testValues(-5.5, 1, 0, ptr(5.5), highs.ModelStatusOptimal, ptr(5.5))
	testValues(-5.5, 1, 0, ptr(-5.6), highs.ModelStatusOptimal, ptr(-5.6))
	testValues(5, 0.5, 0, ptr(2.4), highs.ModelStatusOptimal, ptr(2.4))
	testValues(5, 0.5, 0, ptr(-2.5), highs.ModelStatusOptimal, ptr(-2.5))
	testValues(5, 0.5, 0, ptr(2.6), highs.ModelStatusOptimal, ptr(2.6))
	testValues(5, 0, 0, ptr(-0.1), highs.ModelStatusOptimal, ptr(-0.1))
	testValues(5, 0, 0, ptr(0.1), highs.ModelStatusOptimal, ptr(0.1))
}

func TestContraintAndBuilder(test *testing.T) {
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

	testValues(makeSlice(), nil, highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1), nil, highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0, 0), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0, 1), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1, 0), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1, 1), nil, highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0, 0, 0), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0, 0, 1), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0, 1, 0), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0, 1, 1), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1, 0, 0), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1, 0, 1), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1, 1, 0), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1, 1, 1), nil, highs.ModelStatusOptimal, ptr(1))

	testValues(makeSlice(), ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1), ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0, 0), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0, 1), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1, 0), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1, 1), ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0, 0, 0), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0, 0, 1), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0, 1, 0), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0, 1, 1), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1, 0, 0), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1, 0, 1), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1, 1, 0), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1, 1, 1), ptr(1), highs.ModelStatusOptimal, ptr(1))

	testValues(makeSlice(), ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1), ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 0), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 1), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 0), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 1), ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 0, 0), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 0, 1), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 1, 0), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 1, 1), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 0, 0), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 0, 1), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 1, 0), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 1, 1), ptr(0), highs.ModelStatusInfeasible, nil)
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

	testValues(makeSlice(), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1), nil, highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0, 0), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0, 1), nil, highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(1, 0), nil, highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(1, 1), nil, highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0, 0, 0), nil, highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0, 0, 1), nil, highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0, 1, 0), nil, highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0, 1, 1), nil, highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(1, 0, 0), nil, highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(1, 0, 1), nil, highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(1, 1, 0), nil, highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(1, 1, 1), nil, highs.ModelStatusOptimal, ptr(1))

	testValues(makeSlice(), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(1), ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0, 0), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0, 1), ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(1, 0), ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(1, 1), ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0, 0, 0), ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(makeSlice(0, 0, 1), ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0, 1, 0), ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(0, 1, 1), ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(1, 0, 0), ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(1, 0, 1), ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(1, 1, 0), ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(makeSlice(1, 1, 1), ptr(1), highs.ModelStatusOptimal, ptr(1))

	testValues(makeSlice(), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1), ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 0), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 1), ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 0), ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 1), ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 0, 0), ptr(1), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 0, 1), ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 1, 0), ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(0, 1, 1), ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 0, 0), ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 0, 1), ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 1, 0), ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(makeSlice(1, 1, 1), ptr(0), highs.ModelStatusInfeasible, nil)
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

	testValues(0, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(0, ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(0, ptr(1), highs.ModelStatusOptimal, ptr(1))

	testValues(1, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(1, ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(1, ptr(1), highs.ModelStatusInfeasible, nil)
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

	testValues(0, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(0, ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(0, ptr(1), highs.ModelStatusOptimal, ptr(1))

	testValues(1, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(1, ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(1, ptr(1), highs.ModelStatusInfeasible, nil)
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
	testValues(0, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(1, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(-1, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(13.1, nil, highs.ModelStatusOptimal, ptr(13.1))
	testValues(-13.1, nil, highs.ModelStatusOptimal, ptr(13.1))

	// confirm forced minimum, goes equals, then free when higher
	testValues(1, ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(1, ptr(0.9), highs.ModelStatusInfeasible, nil)
	testValues(1, ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(1, ptr(1.1), highs.ModelStatusOptimal, ptr(1.1))
	testValues(1, ptr(77), highs.ModelStatusOptimal, ptr(77))
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
	testValues(3, 1, 3, 1, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(3, 1, 4, 1, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(4, 1, 3, 1, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(25.4, 1, 12.3, 1, nil, highs.ModelStatusOptimal, ptr(13.1))
	testValues(12.3, 1, 25.4, 1, nil, highs.ModelStatusOptimal, ptr(13.1))

	// negatives
	testValues(-55, 1, -44, 1, nil, highs.ModelStatusOptimal, ptr(11))
	testValues(-44, 1, -55, 1, nil, highs.ModelStatusOptimal, ptr(11))
	testValues(-44, 1, 10, 1, nil, highs.ModelStatusOptimal, ptr(54))
	testValues(10, 1, -44, 1, nil, highs.ModelStatusOptimal, ptr(54))

	// zeros
	testValues(0, 1, -44, 1, nil, highs.ModelStatusOptimal, ptr(44))
	testValues(-44, 1, 0, 1, nil, highs.ModelStatusOptimal, ptr(44))
	testValues(44, 1, 0, 1, nil, highs.ModelStatusOptimal, ptr(44))
	testValues(0, 1, 44, 1, nil, highs.ModelStatusOptimal, ptr(44))

	// coefficients
	testValues(5, 2, 9, 1, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(5, 2, 9, -1, nil, highs.ModelStatusOptimal, ptr(19))
	testValues(5, 0.5, 9, 1, nil, highs.ModelStatusOptimal, ptr(6.5))
	testValues(3, 0.1, 4, 0.1, nil, highs.ModelStatusOptimal, ptr(0.1))

	// confirm forced minimum, goes equals, then free when higher
	testValues(3, 1, 4, 1, ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, 1, ptr(0.9), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, 1, ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(3, 1, 4, 1, ptr(1.1), highs.ModelStatusOptimal, ptr(1.1))
	testValues(3, 1, 4, 1, ptr(77), highs.ModelStatusOptimal, ptr(77))
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
	testValues(3, 1, 3, 1, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(3, 1, 4, 1, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(4, 1, 3, 1, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(25.4, 1, 12.3, 1, nil, highs.ModelStatusOptimal, ptr(13.1))
	testValues(12.3, 1, 25.4, 1, nil, highs.ModelStatusOptimal, ptr(13.1))

	// negatives
	testValues(-55, 1, -44, 1, nil, highs.ModelStatusOptimal, ptr(11))
	testValues(-44, 1, -55, 1, nil, highs.ModelStatusOptimal, ptr(11))
	testValues(-44, 1, 10, 1, nil, highs.ModelStatusOptimal, ptr(54))
	testValues(10, 1, -44, 1, nil, highs.ModelStatusOptimal, ptr(54))

	// zeros
	testValues(0, 1, -44, 1, nil, highs.ModelStatusOptimal, ptr(44))
	testValues(-44, 1, 0, 1, nil, highs.ModelStatusOptimal, ptr(44))
	testValues(44, 1, 0, 1, nil, highs.ModelStatusOptimal, ptr(44))
	testValues(0, 1, 44, 1, nil, highs.ModelStatusOptimal, ptr(44))

	// coefficients
	testValues(5, 2, 9, 1, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(5, 2, 9, -1, nil, highs.ModelStatusOptimal, ptr(19))
	testValues(5, 0.5, 9, 1, nil, highs.ModelStatusOptimal, ptr(6.5))
	testValues(3, 0.1, 4, 0.1, nil, highs.ModelStatusOptimal, ptr(0.1))

	// confirm forced minimum, goes equals, then infeasible when higher too
	testValues(3, 1, 4, 1, ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, 1, ptr(0.9), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, 1, ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(3, 1, 4, 1, ptr(1.1), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, 1, ptr(77), highs.ModelStatusInfeasible, nil)

	// same with values flipped
	testValues(4, 1, 3, 1, ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 3, 1, ptr(0.9), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 3, 1, ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(4, 1, 3, 1, ptr(1.1), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 3, 1, ptr(77), highs.ModelStatusInfeasible, nil)

	// and with a zero
	testValues(4, 1, 4, 1, ptr(-1), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 4, 1, ptr(-0.1), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 4, 1, ptr(0), highs.ModelStatusOptimal, ptr(0))
	testValues(4, 1, 4, 1, ptr(0.1), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 4, 1, ptr(77), highs.ModelStatusInfeasible, nil)

	// flipped sign was buggy
	testValues(3, 1, 4, 1, ptr(-1), highs.ModelStatusInfeasible, nil)
	testValues(4, 1, 3, 1, ptr(-1), highs.ModelStatusInfeasible, nil)
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
	testValues(3, 1, 3, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(3, 1, 4, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(4, 1, 3, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(25.4, 1, 12.3, nil, highs.ModelStatusOptimal, ptr(13.1))
	testValues(12.3, 1, 25.4, nil, highs.ModelStatusOptimal, ptr(13.1))

	// negatives
	testValues(-55, 1, -44, nil, highs.ModelStatusOptimal, ptr(11))
	testValues(-44, 1, -55, nil, highs.ModelStatusOptimal, ptr(11))
	testValues(-44, 1, 10, nil, highs.ModelStatusOptimal, ptr(54))
	testValues(10, 1, -44, nil, highs.ModelStatusOptimal, ptr(54))

	// zeros
	testValues(0, 1, -44, nil, highs.ModelStatusOptimal, ptr(44))
	testValues(-44, 1, 0, nil, highs.ModelStatusOptimal, ptr(44))
	testValues(44, 1, 0, nil, highs.ModelStatusOptimal, ptr(44))
	testValues(0, 1, 44, nil, highs.ModelStatusOptimal, ptr(44))

	// confirm forced minimum, goes equals, then free when higher
	testValues(3, 1, 4, ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, ptr(0.9), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 4, ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(3, 1, 4, ptr(1.1), highs.ModelStatusOptimal, ptr(1.1))
	testValues(3, 1, 4, ptr(77), highs.ModelStatusOptimal, ptr(77))
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
	testValues(3, 1, 3, 1, 0, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(3, 1, 4, 1, 1, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(4, 1, 3, 1, 3, nil, highs.ModelStatusOptimal, ptr(2))
	testValues(25.4, 1, 12.3, 1, -2, nil, highs.ModelStatusOptimal, ptr(15.1))
	testValues(12.3, 1, 25.4, 1, 11, nil, highs.ModelStatusOptimal, ptr(2.1))

	// negatives
	testValues(-55, 1, -44, 1, 0, nil, highs.ModelStatusOptimal, ptr(11))
	testValues(-44, 1, -55, 1, 1, nil, highs.ModelStatusOptimal, ptr(10))
	testValues(-44, 1, 10, 1, -1, nil, highs.ModelStatusOptimal, ptr(55))
	testValues(10, 1, -44, 1, 5, nil, highs.ModelStatusOptimal, ptr(49))
	testValues(3, 1, 4, 1, -10, nil, highs.ModelStatusOptimal, ptr(11))

	// zeros
	testValues(0, 1, -44, 1, 0, nil, highs.ModelStatusOptimal, ptr(44))
	testValues(-44, 1, 0, 1, 1, nil, highs.ModelStatusOptimal, ptr(43))
	testValues(44, 1, 0, 1, 7, nil, highs.ModelStatusOptimal, ptr(37))
	testValues(0, 1, 44, 1, -2.5, nil, highs.ModelStatusOptimal, ptr(46.5))

	// coefficients
	testValues(5, 2, 9, 1, 1, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(5, 2, 9, -1, 1, nil, highs.ModelStatusOptimal, ptr(18))
	testValues(5, 0.5, 9, 1, 1, nil, highs.ModelStatusOptimal, ptr(5.5))
	testValues(3, 0.1, 4, 0.1, 1, nil, highs.ModelStatusOptimal, ptr(0.9))

	// confirm forced minimum, goes equals, then free when higher
	testValues(3, 1, 5, 1, 1, ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 5, 1, 1, ptr(0.9), highs.ModelStatusInfeasible, nil)
	testValues(3, 1, 5, 1, 1, ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(3, 1, 5, 1, 1, ptr(1.1), highs.ModelStatusOptimal, ptr(1.1))
	testValues(3, 1, 5, 1, 1, ptr(77), highs.ModelStatusOptimal, ptr(77))
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
		oneColumn := build.CreateColumnGeneral(highs.Continuous, C_MinusInf, C_PlusInf, nil)
		toggleColumn := build.CreateColumnBool(nil)
		outColumn := build.CreateColumnGeneral(highs.Continuous, C_MinusInf, C_PlusInf, nil)
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
	testValues(0, 1, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(1, 1, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(-1, 1, nil, highs.ModelStatusOptimal, ptr(1))
	testValues(13.1, 1, nil, highs.ModelStatusOptimal, ptr(13.1))
	testValues(-13.1, 1, nil, highs.ModelStatusOptimal, ptr(13.1))

	// confirm forced minimum, goes equals, then free when higher (checking toggle=ON works like normal AbsolueValue)
	testValues(1, 1, ptr(0), highs.ModelStatusInfeasible, nil)
	testValues(1, 1, ptr(0.9), highs.ModelStatusInfeasible, nil)
	testValues(1, 1, ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(1, 1, ptr(1.1), highs.ModelStatusOptimal, ptr(1.1))
	testValues(1, 1, ptr(77), highs.ModelStatusOptimal, ptr(77))

	// standard positive differences toggle=OFF defaults to zero
	testValues(0, 0, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(1, 0, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(-1, 0, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(13.1, 0, nil, highs.ModelStatusOptimal, ptr(0))
	testValues(-13.1, 0, nil, highs.ModelStatusOptimal, ptr(0))

	// confirm toggle off lets anything go
	testValues(1, 0, ptr(0), highs.ModelStatusOptimal, nil)
	testValues(1, 0, ptr(0.9), highs.ModelStatusOptimal, nil)
	testValues(1, 0, ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(1, 0, ptr(1.1), highs.ModelStatusOptimal, ptr(1.1))
	testValues(1, 0, ptr(77), highs.ModelStatusOptimal, ptr(77))
	testValues(-1, 0, ptr(0), highs.ModelStatusOptimal, nil)
	testValues(-1, 0, ptr(0.9), highs.ModelStatusOptimal, nil)
	testValues(-1, 0, ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(-1, 0, ptr(1.1), highs.ModelStatusOptimal, ptr(1.1))
	testValues(-1, 0, ptr(77), highs.ModelStatusOptimal, ptr(77))
	testValues(0, 0, ptr(0), highs.ModelStatusOptimal, nil)
	testValues(0, 0, ptr(0.9), highs.ModelStatusOptimal, nil)
	testValues(0, 0, ptr(1), highs.ModelStatusOptimal, ptr(1))
	testValues(0, 0, ptr(1.1), highs.ModelStatusOptimal, ptr(1.1))
	testValues(0, 0, ptr(77), highs.ModelStatusOptimal, ptr(77))
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

	zero := ptr(0.0)
	one := ptr(1.0)

	// for condition not met output is free
	testValues(0, 0, nil, highs.ModelStatusOptimal, zero)
	testValues(0, 0, zero, highs.ModelStatusOptimal, zero)
	testValues(0, 0, one, highs.ModelStatusOptimal, one)

	// condition met, output forced 1.0
	testValues(0, 1, nil, highs.ModelStatusOptimal, one)
	testValues(0, 1, zero, highs.ModelStatusInfeasible, nil)
	testValues(0, 1, one, highs.ModelStatusOptimal, one)
	testValues(1, 0, nil, highs.ModelStatusOptimal, one)
	testValues(1, 0, zero, highs.ModelStatusInfeasible, nil)
	testValues(1, 0, one, highs.ModelStatusOptimal, one)

	// for condition not met output is free
	testValues(1, 1, nil, highs.ModelStatusOptimal, zero)
	testValues(1, 1, zero, highs.ModelStatusOptimal, zero)
	testValues(1, 1, one, highs.ModelStatusOptimal, one)
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

func assertEqualFloat(expect, actual float64, test *testing.T) {
	test.Helper()
	if !util.FloatsApproxEquals(expect, actual) {
		test.Fatalf("assertEqual failed expect=%v actual=%v", expect, actual)
	}
}

func runHighs(build *LinearBuilder, printer *util.PrintRecorder) *highs.Solution {
	return build.RunHighsFuture(nil).WaitForResultOrPanic().solution
}
