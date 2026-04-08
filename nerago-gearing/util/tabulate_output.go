package util

type TabulateOutput struct {
	data       [][]string
	spacing    int
	alignRight []bool
	// rowBuild   []string
}

func (tab *TabulateOutput) SetColumnSpacing(spacing int) {
	tab.spacing = spacing
}

func (tab *TabulateOutput) ColumnCount() int {
	return len(tab.alignRight)
}

func (tab *TabulateOutput) AddColumnHeader(header string, alignRight bool) {
	if len(tab.data) == 0 {
		tab.data = append(tab.data, make([]string, 0))
	}
	tab.data[0] = append(tab.data[0], header)
	tab.alignRight = append(tab.alignRight, alignRight)
}

func (tab *TabulateOutput) AddRow(row []string) {
	tab.data = append(tab.data, row)
}

// func (tab *TabulateOutput) BeginRow() {
// 	tab.rowBuild = make([]string, 0, tab.ColumnCount())
// }

// func (tab *TabulateOutput) CurrentRowAddValue(value string) {
// 	tab.rowBuild = append(tab.rowBuild, value)
// }

// func (tab *TabulateOutput) FinishRow() {
// 	tab.AddRow(tab.rowBuild)
// 	tab.rowBuild = nil
// }

func (tab *TabulateOutput) columnSizes() []int {
	sizes := make([]int, 0, 20)
	for _, line := range tab.data {
		for len(sizes) < len(line) {
			sizes = append(sizes, 0)
		}
		for col := range len(line) {
			if len(line[col]) > sizes[col] {
				sizes[col] = len(line[col])
			}
		}
	}
	return sizes
}

func (tab *TabulateOutput) Write(printer *PrintRecorder) {
	sizes := tab.columnSizes()
	var builder StringBuild2
	for _, line := range tab.data {
		for col := range len(line) {
			size := sizes[col]
			isRight := tab.alignRight[col]
			writeSpaces(&builder, tab.spacing)
			writeStringAligned(line[col], size, isRight, &builder)
		}
		printer.Println(builder.String())
		builder.Reset()
	}
}

func writeStringAligned(str string, size int, isRight bool, builder *StringBuild2) {
	if isRight {
		writeSpaces(builder, size-len(str))
		builder.WriteString(str)
	} else {
		builder.WriteString(str)
		writeSpaces(builder, size-len(str))
	}
}

func writeSpaces(builder *StringBuild2, repeat int) {
	for range repeat {
		builder.WriteRune(' ')
	}
}
