package util_highs

import (
	"os"
	"paladin_gearing_go/util"
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
		defer file.Close()

		_, err = file.WriteTo(printer)
		verifyNoError(err)
	}
}
