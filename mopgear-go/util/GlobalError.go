package util

import (
	"fmt"
	"runtime/debug"

	"github.com/nerago/mopgear-go/files"
)

func GlobalFatalErrorHandler(err error) {
	if err != nil {
		writeError(err)
		panic(err)
	}
}

func GlobalWarnHandler(err any) {
	if err != nil {
		writeError(ErrorFromAny(err))
	}
}

func ErrorFromAny(err any) error {
	switch cast := err.(type) {
	case error:
		return ErrorTracedWrapNew(cast)
	case string:
		return ErrorTracedNew(cast)
	case fmt.Stringer:
		return ErrorTracedNew(cast.String())
	case nil:
		return nil
	default:
		return ErrorTracedNew("unknown error")
	}
}

func writeError(err error) {
	if g_mainLog != nil {
		if g_mainLog.writeErrorToFile(err) {
			return
		}
	}

	printer := PrintRecorder_CreateLogFileNamed(files.LogOutputPath, "error")
	printer.writeErrorToFile(err)
	printer.Close()
}

type IErrorTraced interface {
	error
	Stack() []byte
}

type ErrorTracedString struct {
	str   string
	stack []byte
}

func ErrorTracedNew(str string) IErrorTraced {
	return &ErrorTracedString{
		str:   str,
		stack: debug.Stack(),
	}
}

func (et *ErrorTracedString) Error() string {
	return et.str
}

func (et *ErrorTracedString) Stack() []byte {
	return et.stack
}

type ErrorTracedWrap struct {
	wrapped error
	stack   []byte
}

func ErrorTracedWrapNew(err error) IErrorTraced {
	return &ErrorTracedWrap{
		wrapped: err,
		stack:   debug.Stack(),
	}
}

func (et *ErrorTracedWrap) Error() string {
	return et.wrapped.Error()
}

func (et *ErrorTracedWrap) Stack() []byte {
	return et.stack
}

func (et *ErrorTracedWrap) Unwrap() error {
	return et.wrapped
}
