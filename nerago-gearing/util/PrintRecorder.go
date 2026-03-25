package util

import (
	"fmt"
	"os"
	"paladin_gearing_go/files"
	"strings"
	"sync"
	"time"
)

type PrintRecorder struct {
	holdOutput bool
	lines      []string
	writer     *os.File
	mutex      sync.Mutex
}

func PrintRecorder_CreateLogFile() *PrintRecorder {
	timeStr := strings.ReplaceAll(time.Now().Format(time.RFC3339), ":", "-")
	logName := files.LogOutputPath + "output-" + timeStr + ".log"
	file, err := os.Create(logName)
	if err != nil {
		panic("error creating log")
	}
	return &PrintRecorder{false, nil, file, sync.Mutex{}}
}

func PrintRecorder_HoldAll() *PrintRecorder {
	return &PrintRecorder{true, nil, nil, sync.Mutex{}}
}

func (print *PrintRecorder) Println0() {
	print.Println("")
}

func (print *PrintRecorder) Println(str string) {
	print.mutex.Lock()

	if print.holdOutput {
		print.lines = append(print.lines, str)
	} else {
		print.writer.WriteString(str)
		print.writer.WriteString("\n")
		fmt.Println(str)
	}

	print.mutex.Unlock()
}

func (print *PrintRecorder) Printf(format string, a ...any) {
	print.mutex.Lock()

	str := fmt.Sprintf(format, a...)
	if print.holdOutput {
		print.lines = append(print.lines, str)
	} else {
		print.writer.WriteString(str)
		fmt.Print(str)
	}

	print.mutex.Unlock()
}

func (print *PrintRecorder) AppendOther(other *PrintRecorder) {
	if !other.holdOutput {
		panic("can't append printer that wasn't holding output")
	}

	other.mutex.Lock()
	print.mutex.Lock()

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

	print.mutex.Unlock()
	other.mutex.Unlock()
}

func (print *PrintRecorder) Close() {
	print.mutex.Lock()

	print.writer.Close()

	// delete if empty
	logName := print.writer.Name()
	info, err := os.Stat(logName)
	if err == nil {
		if info.Size() == 0 {
			os.Remove(logName)
		}
	}

	print.mutex.Unlock()
}
