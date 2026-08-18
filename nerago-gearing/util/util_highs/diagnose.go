package util_highs

import (
	"github.com/nerago/mopgear-go/util"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

func diagnoseInfeasible(build *LinearBuilder, printer *util.PrintRecorder) {
	printer.Println("INFEASIBLE MODEL TRYING TO FIND PROBLEM ROW")
	// diagnoseSearchRange(input, printer)
	printer.DebugEnableConsole()
	diagnoseInfeasibleOneByOne(build, printer)
}

func diagnoseSearchRange(build *LinearBuilder, printer *util.PrintRecorder) {
	min, max := 0, len(build.mat.lowerBound)-1

	for max-min > 2 {
		pivot := (min + max) / 2
		cmp := diagnoseTryHalves(build, min, pivot, max, printer)
		if cmp < 0 {
			max = pivot - 1
		} else if cmp > 0 {
			min = pivot
		} else {
			return
		}
	}

}

func diagnoseTryHalves(build *LinearBuilder, min, pivot, max int, printer *util.PrintRecorder) int {
	cloneLower := build.Clone()
	cloneLower.NoOutput = true
	cloneUpper := build.Clone()
	cloneUpper.NoOutput = true

	cloneLower.mat.deleteRowRange(min, pivot-1)
	cloneUpper.mat.deleteRowRange(pivot, max)

	solutionLowerFuture := cloneLower.RunHighsFuture(nil)
	solutionUpperFuture := cloneUpper.RunHighsFuture(nil)

	statusLower := highs.ModelStatusInfeasible
	solutionLower, hasLowerResult := solutionLowerFuture.WaitForResult()
	if hasLowerResult {
		statusLower = solutionLower.GetSolutionAndDiscardLog().Status
	}

	statusUpper := highs.ModelStatusInfeasible
	solutionUpper, hasUpperResult := solutionUpperFuture.WaitForResult()
	if hasUpperResult {
		statusUpper = solutionUpper.GetSolutionAndDiscardLog().Status
	}

	printer.Printf("Half minus(%d..%d)=%s, minus(%d..%d)=%s\n", min, pivot-1, statusLower.String(), pivot, max, statusUpper.String())

	if statusLower == highs.ModelStatusOptimal {
		return -1
	} else if statusUpper == highs.ModelStatusOptimal {
		return 1
	} else if statusLower == highs.ModelStatusInfeasible {
		return -1
	} else if statusUpper == highs.ModelStatusInfeasible {
		return 1
	} else if statusLower == highs.ModelStatusUnboundedOrInfeasible {
		return -1
	} else if statusUpper == highs.ModelStatusUnboundedOrInfeasible {
		return 1
	} else {
		return 0
	}
}

func diagnoseInfeasibleOneByOne(build *LinearBuilder, printer *util.PrintRecorder) {
	for rowIndex := range build.mat.lowerBound {
		clone := build.Clone()
		clone.NoOutput = true
		clone.mat.deleteRow(rowIndex)
		innerPrint := util.PrintRecorder_HoldAll()
		result := clone.RunHighsFuture(nil).WaitForResultOrPanic()
		solution := result.GetSolutionAndSaveLog(innerPrint)
		printer.Printf("Removed row %4d (%s) []=%2d --> %s\n", rowIndex, build.mat.debug[rowIndex], len(build.mat.entries[rowIndex]), solution.Status.String())
		if solution.Status == highs.ModelStatusOptimal {
			printer.AppendOther(innerPrint)
			debugPrintRow(build, rowIndex, printer)
			//debugPrintSolutionValuesWithRowContext(solution, build, rowIndex, printer)
			// drillDownColumn(solution, input,
		}
	}
}

func debugPrintRow(build *LinearBuilder, rowIndex int, printer *util.PrintRecorder) {
	debugText := build.mat.debug[rowIndex]
	printer.Printf("ROW %d %f-%f %s\n", rowIndex, build.mat.lowerBound[rowIndex], build.mat.upperBound[rowIndex], debugText)
	for _, entry := range build.mat.entries[rowIndex] {
		printer.Printf("  $ %d %f\n", entry.columnNumber, entry.value)
	}
}

func debugPrintSolutionValues(solution *highs.Solution, build *LinearBuilder, printer *util.PrintRecorder) {
	printer.Printf("OBJECTIVE VALUE %f \n", solution.Objective)

	for columnIndex, outputValue := range solution.ColValues {
		debugText := debugText(build.vars.debug[columnIndex])

		printer.Printf("%d %f %s\r\n", columnIndex, outputValue, debugText)
	}
}

func debugPrintSolutionValuesWithRowContext(solution *highs.Solution, build *LinearBuilder, rowIndex int, printer *util.PrintRecorder) {
	entryLookup := make(map[ColumnIndex]float64)
	for _, entry := range build.mat.entries[rowIndex] {
		entryLookup[entry.columnNumber] = entry.value
	}

	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("index", true)
	tab.AddColumnHeader("output", false)
	tab.AddColumnHeader("row coefficient", false)
	tab.AddColumnHeader("debug text", false)

	for columnIndex, outputValue := range solution.ColValues {
		debugText := debugText(build.vars.debug[columnIndex])

		coefficient, hasValue := entryLookup[ColumnIndex(columnIndex)]
		if hasValue {
			tab.AddRow([]string{
				strconv.FormatInt(int64(columnIndex), 10),
				strconv.FormatFloat(outputValue, 'f', 6, 64),
				strconv.FormatFloat(coefficient, 'f', 6, 64),
				debugText,
			})
		} else {
			tab.AddRow([]string{
				strconv.FormatInt(int64(columnIndex), 10),
				strconv.FormatFloat(outputValue, 'f', 6, 64),
				"",
				debugText,
			})
		}
	}

	tab.Write(printer)
}
