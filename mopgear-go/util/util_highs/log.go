package util_highs

import (
	"os"

	"github.com/nerago/mopgear-go/util"
)

func makeTempFilename() string {
	tempFile, err := os.CreateTemp("", "highslog")
	if err != nil {
		panic(err)
	}
	verifyNoError(tempFile.Close())
	return tempFile.Name()
}

func readLogfile(tempFilename string, printer *util.PrintRecorder) {
	if tempFilename != "" && printer != nil {
		file, err := os.Open(tempFilename)
		verifyNoError(err)
		_, err = file.WriteTo(printer)
		verifyNoError(err)
		verifyNoError(file.Close())
	} else if tempFilename != "" {
		_ = os.Remove(tempFilename)
	}
}
