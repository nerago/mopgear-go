package util

import (
	"fmt"
	"slices"
)

type PrintRecorder struct {
	holdOutput bool
	lines      []string
}

func PrintRecorder_AutoOutput() PrintRecorder {
	return PrintRecorder{}
}

func PrintRecorder_HoldAll() PrintRecorder {
	return PrintRecorder{true, nil}
}

func (print *PrintRecorder) Println(str string) {
	if print.holdOutput {
		print.lines = append(print.lines, str)
	} else {
		fmt.Println(str)
	}
}

func (print *PrintRecorder) Printf(format string, a ...any) {
	if print.holdOutput {
		str := fmt.Sprintf(format, a...)
		print.lines = append(print.lines, str)
	} else {
		fmt.Printf(format, a...)
	}
}

func (print *PrintRecorder) OutputNow() {
	for _, line := range print.lines {
		fmt.Println(line)
	}
}

func (print *PrintRecorder) AppendOther(other *PrintRecorder) {
	if other.holdOutput && !print.holdOutput {
		other.OutputNow()
	}
	print.lines = slices.Concat(print.lines, other.lines)
}
