package util

import (
	"os"

	"github.com/nerago/mopgear-go/files"
)

func WriteStringToFile(filename, content string) {
	bytes := []byte(content)
	err := os.WriteFile(filename, bytes, 0666)
	if err != nil {
		panic(err)
	}
}

func WriteFuncToFileWithTemp(filename string, apply func(file *os.File)) error {
	tempFile, err := os.CreateTemp(files.TempLog, "temp")
	if err != nil {
		return err
	}

	apply(tempFile)

	err = tempFile.Sync()
	if err != nil {
		return err
	}
	err = tempFile.Close()
	if err != nil {
		return err
	}

	// technically on Windows not perfectly atomic, but fine for our purposes
	err = os.Rename(tempFile.Name(), filename)
	if err != nil {
		return err
	}
	return nil
}
