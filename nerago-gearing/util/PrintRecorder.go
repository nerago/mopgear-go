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

func PrintRecorder_Testing() *PrintRecorder {
	return &PrintRecorder{false, nil, nil, sync.Mutex{}}
}

func PrintRecorder_HoldAll() *PrintRecorder {
	return &PrintRecorder{true, nil, nil, sync.Mutex{}}
}

var _newline = []byte{'\n'}

func (print *PrintRecorder) outputNewline() {
	if print.file != nil {
		print.file.Write(_newline)
	}
	os.Stdout.Write(_newline)
}

func (print *PrintRecorder) outputBytes(bytes []byte) {
	if print.file != nil {
		print.file.Write(bytes)
	}
	os.Stdout.Write(bytes)
}

func (print *PrintRecorder) outputString(str string) {
	if print.file != nil {
		print.file.WriteString(str)
	}
	os.Stdout.WriteString(str)
}

func (print *PrintRecorder) Println0() {
	print.mutex.Lock()
	defer print.mutex.Unlock()

	if print.holdOutput {
		print.builder.WriteRune('\n')
	} else {
		print.outputNewline()
	}
}

func (print *PrintRecorder) Println(str string) {
	print.mutex.Lock()
	defer print.mutex.Unlock()

	if print.holdOutput {
		print.builder.WriteString(str)
		print.builder.WriteRune('\n')
	} else {
		print.outputString(str)
		print.outputNewline()
	}
}

func (print *PrintRecorder) PrintlnFromBuild(strBuild StringBuild2) {
	print.mutex.Lock()
	defer print.mutex.Unlock()

	if print.holdOutput {
		print.builder.WriteBuilder(strBuild)
		print.builder.WriteRune('\n')
	} else {
		print.outputString(strBuild.String())
		print.outputNewline()
	}
}

func (print *PrintRecorder) Printf(format string, args ...any) {
	print.mutex.Lock()
	defer print.mutex.Unlock()

	str := fmt.Sprintf(format, args...)
	if print.holdOutput {
		print.builder.WriteString(str)
	} else {
		print.outputString(str)
	}
}

func (print *PrintRecorder) PrintBytes(bytes []byte) {
	print.mutex.Lock()
	defer print.mutex.Unlock()

	if print.holdOutput {
		print.builder.WriteBytes(bytes)
	} else {
		print.outputString(string(bytes))
	}
}

func (print *PrintRecorder) AppendOther(other *PrintRecorder) {
	if !other.holdOutput {
		panic("can't append printer that wasn't holding output")
	}

	other.mutex.Lock()
	defer other.mutex.Unlock()
	print.mutex.Lock()
	defer print.mutex.Unlock()

	if print.holdOutput {
		print.builder.WriteBuilder(other.builder)
	} else {
		print.outputBytes(other.builder)
	}
}

func (print *PrintRecorder) Close() {
	print.mutex.Lock()
	defer print.mutex.Unlock()

	print.file.Close()
	deleteIfEmpty(print.file.Name())
}

func deleteIfEmpty(logName string) {
	info, err := os.Stat(logName)
	if err == nil && info.Size() == 0 {
		os.Remove(logName)
	}
}
