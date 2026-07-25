package util_highs

import "paladin_gearing_go/util/util_collection"

type constraintMatrixBuilder struct {
	entries    [][]indexAndValue
	lowerBound []float64
	upperBound []float64
	debug      []string
}

func (mat *constraintMatrixBuilder) clone() constraintMatrixBuilder {
	return constraintMatrixBuilder{
		entries: util_collection.MapSliceAsNew(mat.entries, func(subSlice *[]indexAndValue) []indexAndValue {
			return slices.Clone(*subSlice)
		}),
		lowerBound: slices.Clone(mat.lowerBound),
		upperBound: slices.Clone(mat.upperBound),
		debug:      slices.Clone(mat.debug),
	}
}

func (mat *constraintMatrixBuilder) addRow(entries []indexAndValue, lowerBound float64, upperBound float64, debug string) RowIndex {
	index := len(mat.entries)
	mat.entries = append(mat.entries, entries)
	mat.lowerBound = append(mat.lowerBound, lowerBound)
	mat.upperBound = append(mat.upperBound, upperBound)
	mat.debug = append(mat.debug, debug)
	return RowIndex(index)
}

func (mat *constraintMatrixBuilder) deleteRow(rowIndex int) {
	util_collection.DeleteIndexInPlace(&mat.entries, rowIndex)
	util_collection.DeleteIndexInPlace(&mat.lowerBound, rowIndex)
	util_collection.DeleteIndexInPlace(&mat.upperBound, rowIndex)
	util_collection.DeleteIndexInPlace(&mat.debug, rowIndex)
}

func (mat *constraintMatrixBuilder) deleteRowRange(firstDelete, lastDelete int) {
	mat.entries = slices.Delete(mat.entries, firstDelete, lastDelete+1)
	mat.lowerBound = slices.Delete(mat.lowerBound, firstDelete, lastDelete+1)
	mat.upperBound = slices.Delete(mat.upperBound, firstDelete, lastDelete+1)
	mat.debug = slices.Delete(mat.debug, firstDelete, lastDelete+1)
}

func (mat *constraintMatrixBuilder) createSolverInputArrays() (numRows int32, lowerBound []float64, upperBound []float64, startArray []int32, indexArray []int32, valuesArray []float64) {
	numRows = int32(len(mat.entries))
	if int32(len(mat.lowerBound)) != numRows || int32(len(mat.upperBound)) != numRows {
		panic("inconsistent row count")
	}

	valueCount := 0
	for i := range numRows {
		valueCount += len(mat.entries[i])
	}
	if valueCount == 0 {
		panic("completely empty model")
	}

	startArray = make([]int32, numRows)
	indexArray = make([]int32, valueCount)
	valuesArray = make([]float64, valueCount)

	var insertIndex int32 = 0
	for rowNum, rowEntries := range mat.entries {
		startArray[rowNum] = insertIndex

		for _, entry := range rowEntries {
			indexArray[insertIndex] = int32(entry.columnNumber)
			valuesArray[insertIndex] = entry.value
			insertIndex++
		}
	}

	return numRows, mat.lowerBound, mat.upperBound, startArray, indexArray, valuesArray
}
