package util

type CSVOutputByColumn struct {
	builds  []StringBuild2
	currRow int
}

func (csv *CSVOutputByColumn) InitRows(count int) {
	csv.builds = make([]StringBuild2, count)
	csv.currRow = 0
}

func (csv *CSVOutputByColumn) FinishColumn() {
	if csv.currRow != len(csv.builds) {
		panic("too few/many values for column")
	}
	csv.currRow = 0
}

func (csv *CSVOutputByColumn) verifyAdd() {
	if csv.currRow >= len(csv.builds) {
		panic("too many values for column")
	}
}

func (csv *CSVOutputByColumn) AddString(value string) {
	csv.verifyAdd()
	build := &csv.builds[csv.currRow]
	build.WriteString(value)
	build.WriteRune(',')
	csv.currRow++
}

func (csv *CSVOutputByColumn) AddStringMany(valueSlice ...string) {
	for _, value := range valueSlice {
		csv.AddString(value)
	}
}

func (csv *CSVOutputByColumn) AddToBuilder(apply func(b *StringBuild2)) {
	csv.verifyAdd()
	build := &csv.builds[csv.currRow]
	apply(build)
	build.WriteRune(',')
	csv.currRow++
}

func (csv *CSVOutputByColumn) AddFloat64(value float64, decimalPlaces int) {
	csv.verifyAdd()
	build := &csv.builds[csv.currRow]
	build.WriteFloat64(value, decimalPlaces)
	build.WriteRune(',')
	csv.currRow++
}

func (csv *CSVOutputByColumn) AddInt(value int) {
	csv.verifyAdd()
	build := &csv.builds[csv.currRow]
	build.WriteInt(value)
	build.WriteRune(',')
	csv.currRow++
}

func (csv *CSVOutputByColumn) Write(printer *PrintRecorder) {
	for index := range csv.builds {
		printer.PrintlnFromBuild(csv.builds[index])
	}
}
