package util

import (
	"errors"
	"os"

	"github.com/nerago/mopgear-go/files"
)

func WriteStringToFile(filename, content string) error {
	bytes := []byte(content)
	err := os.WriteFile(filename, bytes, 0666)
	return err
}

func WriteFuncToFileWithTemp(filename string, apply func(file *os.File)) error {
	tempFile, err := os.CreateTemp(files.TempLog, "temp")
	if err != nil {
		return err
	}

	apply(tempFile)

	err = tempFile.Sync()
	if err != nil {
		_ = tempFile.Close()
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

func ReadFileSniffTen(filename string) (rune, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 'e', err
	}

	buff := [10]byte{}
	num, errRead := file.Read(buff[:])
	errClose := file.Close()
	if errRead != nil || errClose != nil {
		return 'e', errors.Join(errRead, errClose)
	}

	for _, c := range buff[:num] {
		if c == '(' {
			return 'P', nil
		} else if c == '{' {
			return 'J', nil
		}
	}

	return 'e', ErrorTracedNew("unknown file type")
}
