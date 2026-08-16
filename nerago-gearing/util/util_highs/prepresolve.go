package util_highs

import (
	"fmt"
	"math"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

func (build *LinearBuilder) PrePreSolve(printer *util.PrintRecorder) {
	findSingletons(&build.mat, printer)

	// DoubletonEquations rowUpper==rowLower, rowSize==2
	//   * check if row infeasible when binary variables forced either low or high, then it must be the other
	//   * column can be substituted? prefer non-integer
	//   * for integer dividing and rounding mutual coefficients, does it shift too much, then don't. HPreSolve:3116
	//   * bunch of complicated conditions but boils down to basic equality relationship with 2 non-zeros
	findPossibleSubstitutions(&build.mat, &build.vars, printer)
	// another implied bounds check involving coefficients for Mip HPresolve:3603
	// would then allow all sorts of row substitution

	// row is redundant
	//   * alternate considering possible row and column bounds(hi/lo), building an implied bounds
	//   * it's redundant from the point limits are equal or smaller than row bounds
	//   * but for us that might be okay
	// is a ranged row constraint effectively an equality once implied bounds are considered

	// forced equality if started unequal for implied bounds pushes it to the edge HPresolve:4346

	// sparsify rowUpper==rowLower
	//   * compare rows via mutual columns HPresolve:8263
	//   * looking for rows that can partially cancel each other out where appropriate scaling is used
	//   * trying to create zeros

	// be nice to understand implied ints, getting a handful (1% cols)
}

func findSingletons(mat *constraintMatrixBuilder, printer *util.PrintRecorder) {
	for rowIndex, row := range mat.entries {
		nonZero := 0
		for _, entry := range row {
			if floatNonZeroPreSolver(entry.value) {
				nonZero++
			}
		}
		if nonZero < 2 {
			reportIssueRow(printer, mat, rowIndex, "less than two non-zeros")
		}
	}
}

const c_float_small_value = 1e-9

func floatNonZeroPreSolver(value float64) bool {
	return math.Abs(value) >= c_float_small_value
}

func findPossibleSubstitutions(mat *constraintMatrixBuilder, vars *variableArrayBuilder, printer *util.PrintRecorder) {
	for rowIndex, row := range mat.entries {
		lo := mat.lowerBound[rowIndex]
		hi := mat.upperBound[rowIndex]
		if lo == hi { // highs seems to just equality the floats
			nonZero := 0
			for _, entry := range row {
				if floatNonZeroPreSolver(entry.value) {
					nonZero++
				}
			}
			if nonZero != len(row) {
				panic("unexpected zeros")
			} else if nonZero == 2 {
				col1 := row[0].columnNumber
				col2 := row[1].columnNumber
				if (vars.colTypes[col1] == highs.Continuous && vars.colTypes[col2] == highs.Continuous) || (vars.colTypes[col1] == highs.Integer && vars.colTypes[col2] == highs.Integer) {
					reportIssueCol(printer, vars, col1, fmt.Sprintf("possible substitution with below based on row %d (%s)", rowIndex, mat.debug[rowIndex]))
					reportIssueCol(printer, vars, col2, fmt.Sprintf("possible substitution with above based on row %d (%s)", rowIndex, mat.debug[rowIndex]))
				} else if vars.colTypes[col1] == highs.Continuous && vars.colTypes[col2] == highs.Integer {
					reportIssueCol(printer, vars, col1, fmt.Sprintf("possible substitution with %d (%s) based on row %d (%s)", col2, debugText(vars.debug[col2]), rowIndex, mat.debug[rowIndex]))
				} else if vars.colTypes[col1] == highs.Integer && vars.colTypes[col2] == highs.Continuous {
					reportIssueCol(printer, vars, col2, fmt.Sprintf("possible substitution with %d (%s) based on row %d (%s)", col1, debugText(vars.debug[col1]), rowIndex, mat.debug[rowIndex]))
				} else {
					reportIssueCol(printer, vars, col1, "don't understand")
					reportIssueCol(printer, vars, col2, "don't understand")
				}
			}
		}
	}
}

func reportIssueCol(printer *util.PrintRecorder, vars *variableArrayBuilder, colIndex ColumnIndex, msg string) {
	printer.Printf("Col %d (%s) %s\n", colIndex, debugText(vars.debug[colIndex]), msg)
}

func reportIssueRow(printer *util.PrintRecorder, mat *constraintMatrixBuilder, rowIndex int, msg string) {
	printer.Printf("Row %d (%s) %s\n", rowIndex, mat.debug[rowIndex], msg)
}
