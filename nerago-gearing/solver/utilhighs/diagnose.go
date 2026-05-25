package utilhighs

import (
	"paladin_gearing_go/util"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

func diagnoseInfeasible(input *InputBuilder, printer *util.PrintRecorder) {
	printer.Println("INFEASIBLE MODEL TRYING TO FIND PROBLEM ROW")
	// diagnoseSearchRange(input, printer)
	diagnoseInfeasibleOneByOne(input, printer)
}

func diagnoseSearchRange(input *InputBuilder, printer *util.PrintRecorder) {
	min, max := 0, len(input.mat.lowerBound)-1

	for max-min > 2 {
		pivot := (min + max) / 2
		cmp := diagnoseTryHalves(input, min, pivot, max, printer)
		if cmp < 0 {
			max = pivot - 1
		} else if cmp > 0 {
			min = pivot
		} else {
			return
		}
	}

}

func diagnoseTryHalves(input *InputBuilder, min, pivot, max int, printer *util.PrintRecorder) int {
	cloneLower := input.Clone()
	cloneLower.NoOutput = true
	cloneUpper := input.Clone()
	cloneUpper.NoOutput = true

	cloneLower.mat.deleteRowRange(min, pivot-1)
	cloneUpper.mat.deleteRowRange(pivot, max)

	solutionLower, _ := cloneLower.RunHighs()
	solutionUpper, _ := cloneUpper.RunHighs()

	printer.Printf("Half minus(%d..%d)=%s, minus(%d..%d)=%s\n", min, pivot-1, solutionLower.Status.String(), pivot, max, solutionUpper.Status.String())

	if solutionLower.Status == highs.ModelStatusOptimal {
		return -1
	} else if solutionUpper.Status == highs.ModelStatusOptimal {
		return 1
	} else if solutionLower.Status == highs.ModelStatusInfeasible {
		return -1
	} else if solutionUpper.Status == highs.ModelStatusInfeasible {
		return 1
	} else if solutionLower.Status == highs.ModelStatusUnboundedOrInfeasible {
		return -1
	} else if solutionUpper.Status == highs.ModelStatusUnboundedOrInfeasible {
		return 1
	} else {
		return 0
	}
}

func diagnoseInfeasibleOneByOne(input *InputBuilder, printer *util.PrintRecorder) {
	for rowIndex := range input.mat.lowerBound {
		clone := input.Clone()
		clone.NoOutput = true
		clone.mat.deleteRow(rowIndex)
		solution, log := clone.RunHighs()
		printer.Printf("Removed row %4d (%s) []=%2d --> %s\n", rowIndex, input.mat.debug[rowIndex], len(input.mat.entries[rowIndex]), solution.Status.String())
		if solution.Status == highs.ModelStatusOptimal {
			printer.AppendOther(log)
			debugPrintRow(input, rowIndex, printer)
			debugPrintSolutionValuesWithRowContext(solution, input, rowIndex, printer)
			// drillDownColumn(solution, input,
		}
	}
}

func debugPrintRow(input *InputBuilder, rowIndex int, printer *util.PrintRecorder) {
	debugText := input.mat.debug[rowIndex]
	printer.Printf("ROW %d %f-%f %s\n", rowIndex, input.mat.lowerBound[rowIndex], input.mat.upperBound[rowIndex], debugText)
	for _, entry := range input.mat.entries[rowIndex] {
		printer.Printf("  $ %d %f\n", entry.columnNumber, entry.value)
	}
}

func debugPrintSolutionValues(solution *highs.Solution, input *InputBuilder, printer *util.PrintRecorder) {
	printer.Printf("OBJECTIVE VALUE %f \n", solution.Objective)

	for columnIndex, outputValue := range solution.ColValues {
		debugText := debugText(input.vars.debug[columnIndex])

		printer.Printf("%d %f %s\r\n", columnIndex, outputValue, debugText)
	}
}

func debugPrintSolutionValuesWithRowContext(solution *highs.Solution, input *InputBuilder, rowIndex int, printer *util.PrintRecorder) {
	entryLookup := make(map[ColumnIndex]float64)
	for _, entry := range input.mat.entries[rowIndex] {
		entryLookup[entry.columnNumber] = entry.value
	}

	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("index", true)
	tab.AddColumnHeader("output", false)
	tab.AddColumnHeader("row coefficient", false)
	tab.AddColumnHeader("debug text", false)

	for columnIndex, outputValue := range solution.ColValues {
		debugText := debugText(input.vars.debug[columnIndex])

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
