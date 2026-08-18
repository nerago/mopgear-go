package util_highs

import (
	"cmp"
	"fmt"
	"math"
	"slices"
)

type indexAndValue struct {
	columnNumber ColumnIndex
	value        float64
}

type ConstraintRow struct {
	entries []indexAndValue
	Debug   string
}

func (row *ConstraintRow) IsEmpty() bool {
	return len(row.entries) == 0
}

func (row *ConstraintRow) HasValues() bool {
	return len(row.entries) > 0
}

func (row *ConstraintRow) Add(columnIndex ColumnIndex, value float64) {
	if value != 0.0 {
		row.entries = append(row.entries, indexAndValue{columnIndex, value})
	}
}

func (row *ConstraintRow) Change(columnIndex ColumnIndex, value float64) {
	for i := range row.entries {
		if row.entries[i].columnNumber == columnIndex {
			row.entries[i].value = value
			return
		}
	}

	panic("column didn't exist")
}

func (row *ConstraintRow) Build(build *LinearBuilder, lowerBound float64, upperBound float64) {
	// couldn't find reference for sure that indexes need to be sorted but probably best
	slices.SortFunc(row.entries, func(a, b indexAndValue) int { return cmp.Compare(a.columnNumber, b.columnNumber) })

	if C_DebugHighs {
		if len(row.entries) == 0 && lowerBound != 0 && upperBound != 0 {
			panic("empty row makes infeasible")
		} else if len(row.entries) == 0 {
			fmt.Printf("warn empty row\n")
		}
		if lowerBound > upperBound {
			panic("row bounds backwards")
		}
		if (0 < math.Abs(lowerBound) && math.Abs(lowerBound) < 1e-6) || (0 < math.Abs(upperBound) && math.Abs(upperBound) < 1e-6) {
			panic("row bounds very small")
		}
		for i := range row.entries {
			value := row.entries[i].value
			if value >= 1e+15 {
				panic("matrix entry very large")
			}
		}
	}

	build.mat.addRow(row.entries, lowerBound, upperBound, row.Debug)
}
