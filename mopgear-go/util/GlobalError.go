package util

import "github.com/nerago/mopgear-go/files"

func GlobalErrorHandler(err error) {
	if err != nil {
		if g_mainLog != nil {
			if g_mainLog.writeErrorToFile(err) {
				return
			}
		}

		printer := PrintRecorder_CreateLogFile(files.LogOutputPath)
		printer.writeErrorToFile(err)
		printer.Close()

		panic(err)
	}
}
