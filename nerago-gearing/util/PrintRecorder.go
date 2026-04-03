package util

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type PrintRecorder struct {
	holdOutput bool
	builder    StringBuild2
	file       *os.File
	mutex      sync.Mutex
}

func PrintRecorder_CreateLogFile(path string) *PrintRecorder {
	timeStr := strings.ReplaceAll(time.Now().Format(time.RFC3339), ":", "-")
	logName := path + "output-" + timeStr + ".log"
	file, err := os.Create(logName)
	if err != nil {
		panic("error creating log")
	}
	return &PrintRecorder{false, nil, file, sync.Mutex{}}
}

func PrintRecorder_HoldAll() *PrintRecorder {
	return &PrintRecorder{true, nil, nil, sync.Mutex{}}
}

var _newline = []byte{'\n'}

func (print *PrintRecorder) outputNewline() {
	print.file.Write(_newline)
	os.Stdout.Write(_newline)
}

func (print *PrintRecorder) outputBytes(bytes []byte) {
	print.file.Write(bytes)
	os.Stdout.Write(bytes)
}

func (print *PrintRecorder) outputString(str string) {
	print.file.WriteString(str)
	os.Stdout.WriteString(str)
}

func (print *PrintRecorder) Println0() {
	print.mutex.Lock()

	if print.holdOutput {
		print.builder.WriteRune('\n')
	} else {
		print.outputNewline()
	}

	print.mutex.Unlock()
}

func (print *PrintRecorder) Println(str string) {
	print.mutex.Lock()

	if print.holdOutput {
		print.builder.WriteString(str)
		print.builder.WriteRune('\n')
	} else {
		print.outputString(str)
		print.outputNewline()
	}

	print.mutex.Unlock()
}

func (print *PrintRecorder) Printf(format string, args ...any) {
	print.mutex.Lock()

	str := fmt.Sprintf(format, args...)
	if print.holdOutput {
		print.builder.WriteString(str)
	} else {
		print.outputString(str)
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
		print.builder.WriteBuilder(other.builder)
	} else {
		print.outputBytes(other.builder)
	}

	print.mutex.Unlock()
	other.mutex.Unlock()
}

func (print *PrintRecorder) Close() {
	print.mutex.Lock()

	print.file.Close()

	// delete if empty
	logName := print.file.Name()
	info, err := os.Stat(logName)
	if err == nil {
		if info.Size() == 0 {
			os.Remove(logName)
		}
	}

	print.mutex.Unlock()
}
