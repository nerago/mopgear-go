package util

type CSVOutput struct {
	data [][]string
}

func (csv *CSVOutput) AddItem(values []string) {
	if len(csv.data) > 0 && len(csv.data[0]) != len(values) {
		panic("inconsistent value size")
	}
	csv.data = append(csv.data, values)
}

func (csv *CSVOutput) Write(printer *PrintRecorder) {
	var builder StringBuild2
	for rowIndex := range csv.data[0] {
		for _, values := range csv.data {
			val := values[rowIndex]
			builder.WriteString(val)
			builder.WriteRune(',')
		}
		builder.WriteRune('\n')
	}
	printer.Println(builder.String())
}
