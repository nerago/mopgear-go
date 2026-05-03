package withhighs

func contraintIfBoolCopyValueElseZero(input *inputBuilder, boolSwitchVar, sourceVar, targetVar int, rangeLow, rangeHigh float64) {
	// based on https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d

	valueHigh := constraintRowBuild{}
	valueHigh.add(targetVar, -1)
	valueHigh.add(sourceVar, 1)
	valueHigh.add(boolSwitchVar, rangeHigh)
	valueHigh.finish(input, c_minusInf, rangeHigh)

	valueLow := constraintRowBuild{}
	valueLow.add(targetVar, 1)
	valueLow.add(sourceVar, -1)
	valueLow.add(boolSwitchVar, -rangeLow)
	valueLow.finish(input, c_minusInf, -rangeLow)

	zeroHigh := constraintRowBuild{}
	zeroHigh.add(targetVar, 1)
	zeroHigh.add(boolSwitchVar, -rangeHigh)
	zeroHigh.finish(input, c_minusInf, 0)

	zeroLow := constraintRowBuild{}
	zeroLow.add(targetVar, -1)
	zeroLow.add(boolSwitchVar, rangeLow)
	zeroLow.finish(input, c_minusInf, 0)
}

// https://medium.com/data-science/a-comprehensive-guide-to-modeling-techniques-in-mixed-integer-linear-programming-3e96cc1bc03d
type contraintAndBuilder struct {
	outputVar int
	inputVars []int
}

func (build *contraintAndBuilder) setOutput(column int) {
	build.outputVar = column
}

func (build *contraintAndBuilder) addInput(column int) {
	build.inputVars = append(build.inputVars, column)
}

func (build *contraintAndBuilder) finishAndApply(input *inputBuilder) {
	sumRow := constraintRowBuild{}
	sumRow.add(build.outputVar, -1)

	for _, inputVar := range build.inputVars {
		sumRow.add(inputVar, 1)

		pullDown := constraintRowBuild{}
		pullDown.add(inputVar, -1)
		pullDown.add(build.outputVar, 1)
		pullDown.finish(input, c_minusInf, 0)
	}

	targetNum := len(build.inputVars) - 1
	sumRow.finish(input, c_minusInf, float64(targetNum))
}
