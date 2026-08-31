package util

import (
	"errors"
	"fmt"

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
		writeError(toError(err))
	}
}

func toError(err any) error {
	switch cast := err.(type) {
	case error:
		return cast
	case string:
		return errors.New(cast)
	case fmt.Stringer:
		return errors.New(cast.String())
	default:
		return errors.New("unknown error")
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
