package util

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type PrintRecorder struct {
	holdOutput    bool
	consoleOutput bool
	debugConsole  bool
	builder       StringBuild2
	file          *os.File
	test          *testing.T
	parent        *PrintRecorder
	prefix        string
	mutex         sync.Mutex
}

func PrintRecorder_CreateLogFile(directory string) *PrintRecorder {
	timeStr := strings.ReplaceAll(time.Now().Format(time.DateTime), ":", "-")
	logName := directory + "output-" + timeStr + ".log"
	file, err := os.Create(logName)
	if err != nil {
		panic(err)
	}
	return &PrintRecorder{false, true, false, nil, file, nil, nil, "", sync.Mutex{}}
}

func PrintRecorder_CreateLogFileNamed(directory string, tag string) *PrintRecorder {
	timeStr := strings.ReplaceAll(time.Now().Format(time.DateTime), ":", "-")
	logName := directory + "output-" + timeStr + "-" + tag + ".log"
	file, err := os.Create(logName)
	if err != nil {
		panic(err)
	}
	return &PrintRecorder{false, true, false, nil, file, nil, nil, "", sync.Mutex{}}
}

func PrintRecorder_Testing(test *testing.T) *PrintRecorder {
	return &PrintRecorder{false, true, false, nil, nil, test, nil, "", sync.Mutex{}}
}

func PrintRecorder_HoldAll() *PrintRecorder {
	return &PrintRecorder{true, false, false, nil, nil, nil, nil, "", sync.Mutex{}}
}

func PrintRecorder_Nop() *PrintRecorder {
	return &PrintRecorder{false, false, false, nil, nil, nil, nil, "", sync.Mutex{}}
}

func (print *PrintRecorder) NewChildPrefixed(prefixLines string) *PrintRecorder {
	return &PrintRecorder{false, false, false, nil, nil, nil, print, prefixLines, sync.Mutex{}}
}

func (print *PrintRecorder) DebugEnableConsole() {
	print.debugConsole = true
}

var _newline = []byte{'\n'}

func (print *PrintRecorder) outputNewline() {
	if print.file != nil {
		_, err := print.file.Write(_newline)
		if err != nil {
			panic(err)
		}
	} else if print.test != nil {
		print.test.Log()
	} else if print.parent != nil {
		print.parent.Println(print.prefix)
	}

	if print.consoleOutput {
		_, err := os.Stdout.Write(_newline)
		if err != nil {
			panic(err)
		}
	}
}

func (print *PrintRecorder) outputBytes(bytes []byte) {
	if print.file != nil {
		_, err := print.file.Write(bytes)
		if err != nil {
			panic(err)
		}
	} else if print.test != nil {
		print.test.Log(bytes)
	} else if print.parent != nil {
		print.parent.mutex.Lock()
		defer print.parent.mutex.Unlock()

		if print.parent.holdOutput {
			print.parent.builder.WriteString(print.prefix)
			print.parent.builder.WriteRune(' ')
			print.parent.builder.WriteBytes(bytes)
		} else {
			print.parent.outputString(print.prefix)
			print.parent.outputString(" ")
			print.parent.outputBytes(bytes)
		}
	}

	if print.consoleOutput {
		_, err := os.Stdout.Write(bytes)
		if err != nil {
			panic(err)
		}
	}
}

func (print *PrintRecorder) outputString(str string) {
	if print.file != nil {
		_, err := print.file.WriteString(str)
		if err != nil {
			panic(err)
		}
	} else if print.test != nil {
		print.test.Log(str)
	} else if print.parent != nil {
		print.parent.Printf("%s %v", print.prefix, str)
	}

	if print.consoleOutput {
		_, err := os.Stdout.WriteString(str)
		if err != nil {
			panic(err)
		}
	}
}

func (print *PrintRecorder) Println0() {
	print.mutex.Lock()
	defer print.mutex.Unlock()

	if print.holdOutput {
		print.builder.WriteRune('\n')
	} else {
		print.outputNewline()
	}

	if print.debugConsole {
		_, _ = os.Stdout.Write([]byte{'\n'})
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

	if print.debugConsole {
		_, _ = os.Stdout.WriteString(str)
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

	if print.debugConsole {
		_, _ = os.Stdout.WriteString(strBuild.String())
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

	if print.debugConsole {
		_, _ = os.Stdout.WriteString(str)
	}
}

func (print *PrintRecorder) PrintBytes(bytes []byte) {
	print.mutex.Lock()
	defer print.mutex.Unlock()

	if print.holdOutput {
		print.builder.WriteBytes(bytes)
	} else {
		print.outputBytes(bytes)
	}

	if print.debugConsole {
		_, _ = os.Stdout.Write(bytes)
	}
}

// Write for Writer interface
func (print *PrintRecorder) Write(p []byte) (n int, err error) {
	print.PrintBytes(p)
	return len(p), nil
}

func (print *PrintRecorder) AppendOther(other *PrintRecorder) {
	if !other.holdOutput {
		panic("can't append printer that wasn't holding output")
	}

	other.mutex.Lock()
	defer other.mutex.Unlock()
	print.mutex.Lock()
	defer print.mutex.Unlock()

	if other.prefix != "" {
		panic("can't add prefix, lines already written to buffer")
	}

	if print.holdOutput {
		print.builder.WriteBuilder(other.builder)
	} else {
		print.outputBytes(other.builder)
	}

	other.builder.Free()
}

func (print *PrintRecorder) Close() {
	print.mutex.Lock()
	defer print.mutex.Unlock()

	err := print.file.Close()
	if err != nil {
		panic(err)
	}

	deleteIfEmpty(print.file.Name())
}

func deleteIfEmpty(logName string) {
	info, err := os.Stat(logName)
	if err == nil && info.Size() <= 32 {
		_ = os.Remove(logName)
	}
}
