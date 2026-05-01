package withhighs

func contraintIfBoolCopyValueElseZero(mat *constraintMatrixBuilder, boolSwitchVar, sourceVar, targetVar int, rangeLow, rangeHigh float64) {
	// based on https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d

	valueHigh := constraintRowSparse{}
	valueHigh.add(targetVar, -1)
	valueHigh.add(sourceVar, 1)
	valueHigh.add(boolSwitchVar, rangeHigh)
	valueHigh.finish(mat, c_minusInf, rangeHigh)

	valueLow := constraintRowSparse{}
	valueLow.add(targetVar, 1)
	valueLow.add(sourceVar, -1)
	valueLow.add(boolSwitchVar, -rangeLow)
	valueLow.finish(mat, c_minusInf, -rangeLow)

	zeroHigh := constraintRowSparse{}
	zeroHigh.add(targetVar, 1)
	zeroHigh.add(boolSwitchVar, -rangeHigh)
	zeroHigh.finish(mat, c_minusInf, 0)

	zeroLow := constraintRowSparse{}
	zeroLow.add(targetVar, -1)
	zeroLow.add(boolSwitchVar, rangeLow)
	zeroLow.finish(mat, c_minusInf, 0)
}

// https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d
type contraintAndBuilder struct {
	outputVar int
	inputVar  []int
}

func (build *contraintAndBuilder) setOutput(column int) {
	build.outputVar = column
}

func (build *contraintAndBuilder) addInput(column int) {
	build.inputVar = append(build.inputVar, column)
}

func (build *contraintAndBuilder) finishAndApply(mat *constraintMatrixBuilder) {
	sumRow := constraintRowSparse{}
	sumRow.add(build.outputVar, -1)

	for _, input := range build.inputVar {
		sumRow.add(input, 1)

		pullDown := constraintRowSparse{}
		pullDown.add(input, -1)
		pullDown.add(build.outputVar, 1)
		pullDown.finish(mat, c_minusInf, 0)
	}

	targetNum := len(build.inputVar) - 1
	sumRow.finish(mat, c_minusInf, float64(targetNum))
}
