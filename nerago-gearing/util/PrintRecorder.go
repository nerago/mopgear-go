package util

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type PrintRecorder struct {
	holdOutput bool
	lines      []string
	writer     *os.File
}

func PrintRecorder_CreateLogFile() *PrintRecorder {
	timeStr := strings.Replace(time.Now().Format(time.RFC3339), ":", "-", -1)
	logName := "output-" + timeStr + ".log"
	file, err := os.Create(logName)
	if err != nil {
		panic("error creating log")
	}
	return &PrintRecorder{false, nil, file}
}

func PrintRecorder_HoldAll() *PrintRecorder {
	return &PrintRecorder{true, nil, nil}
}

func (print *PrintRecorder) Println0() {
	print.Println("")
}

func (print *PrintRecorder) Println(str string) {
	if print.holdOutput {
		print.lines = append(print.lines, str)
	} else {
		print.writer.WriteString(str)
		print.writer.WriteString("\n")
		fmt.Println(str)
	}
}

func (print *PrintRecorder) Printf(format string, a ...any) {
	str := fmt.Sprintf(format, a...)
	if print.holdOutput {
		print.lines = append(print.lines, str)
	} else {
		print.writer.WriteString(str)
		fmt.Print(str)
	}
}

func (print *PrintRecorder) AppendOther(other *PrintRecorder) {
	if !other.holdOutput {
		panic("can't append printer that wasn't holding output")
	}

	if print.holdOutput {
		print.lines = append(print.lines, other.lines...)
	} else {
		for _, line := range other.lines {
			if len(line) > 0 && line[len(line)-1] == '\n' {
				print.writer.WriteString(line)
				fmt.Print(line)
			} else {
				print.writer.WriteString(line)
				print.writer.WriteString("\n")
				fmt.Println(line)
			}
		}
	}
}

func (print *PrintRecorder) Close() {
	print.writer.Close()
}
