package util_highs

import (
	"os"

	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/util"
)

func makeTempFilename() string {
	tempFile, err := os.CreateTemp(files.TempLog, "highslog")
	if err != nil {
		panic(err)
	}
	verifyNoError(tempFile.Close())
	return tempFile.Name()
}

func readLogfile(tempFilename string, printer *util.PrintRecorder) error {
	if tempFilename != "" && printer != nil {
		file, err := os.Open(tempFilename)
		if err != nil {
			return err
		}

		_, err = file.WriteTo(printer)
		if err != nil {
			_ = file.Close()
			return err
		}

		err = file.Close()
		if err != nil {
			return err
		}

		_ = os.Remove(tempFilename)
	} else if tempFilename != "" {
		_ = os.Remove(tempFilename)
	}
	return nil
}
