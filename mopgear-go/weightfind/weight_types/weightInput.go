package weight_types

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
)

type WeightInput struct {
	TotalStat stats.StatBlock
	SimResult stats.SimData
}

func weightReadCommon(filename string) ([]WeightInput, bool, error) {
	bytes, err := os.ReadFile(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, true, nil
	} else if err != nil {
		return nil, false, fmt.Errorf("file load error: %s: %w", filename, err)
	}

	var weightInputs []WeightInput
	err = json.Unmarshal(bytes, &weightInputs)
	if err != nil {
		return nil, false, fmt.Errorf("file load error: %s: %w", filename, err)
	}
	return weightInputs, false, nil
}

func WeightInputReadFile(filename string) ([]WeightInput, error) {
	weightInputs, notFound, err := weightReadCommon(filename)
	if notFound && err == nil {
		return nil, util.ErrorTracedNewFormat("file not found: %s", filename)
	} else {
		return weightInputs, err
	}
}

func WeightInputReadFileMultiple(filenameParams ...string) ([]WeightInput, error) {
	fullResultSlice := make([]WeightInput, 0)
	for _, filename := range filenameParams {
		data, err := WeightInputReadFile(filename)
		if err != nil {
			return nil, err
		}
		fullResultSlice = append(fullResultSlice, data...)
	}
	return fullResultSlice, nil
}

func WeightInputReadFileAndCheckAge(filename string) ([]WeightInput, time.Duration, error) {
	statInfo, err := os.Stat(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, 0, nil
	} else if err != nil {
		return nil, 0, err
	}
	dataAge := time.Since(statInfo.ModTime())

	weightInputs, _, err2 := weightReadCommon(filename)
	return weightInputs, dataAge, err2
}

func WeightInputWriteFile(weightInputs []WeightInput, filename string) error {
	bytes, err := json.Marshal(weightInputs)
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, bytes, 0666)
	if err != nil {
		return err
	}
	return nil
}

func WeightInputReadFileOrPanic(filename string) []WeightInput {
	data, err := WeightInputReadFile(filename)
	if err != nil {
		panic(err)
	}
	return data
}

func WeightInputWriteFileOrPanic(weightInputs []WeightInput, filename string) {
	err := WeightInputWriteFile(weightInputs, filename)
	if err != nil {
		panic(err)
	}
}
