package updateProc

import (
	"encoding/json/v2"
	"errors"
	"io/fs"
	"os"
	"slices"
	"time"

	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/weightfind"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func (spec *weightSpecInternal) prepareSimData(taskPoolSim *util_async.NestedTaskPoolChild, tracker *util.TrackProgress, cancel util_async.CancelSignal) error {
	inputDataGrid, err := spec.prepareDataGrid(tracker, cancel)
	if err != nil {
		return err
	}

	inputDataReal, err2 := spec.prepareDataRandom(tracker, cancel)
	if err2 != nil {
		return err2
	}

	inputDataFit, err3 := spec.prepareDataFit(tracker, cancel)
	if err3 != nil {
		return err3
	}

	spec.inputs.dataAll = slices.Concat(inputDataGrid, inputDataReal, inputDataFit)
	return nil
}

func readWeightInputFile(filename string) ([]weight_types.WeightInput, time.Duration) {
	statInfo, err := os.Stat(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, 0
	} else if err != nil {
		panic(err)
	}

	// only use data from "today"
	dataAge := time.Since(statInfo.ModTime())

	bytes, err := os.ReadFile(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, 0
	} else if err != nil {
		panic(err)
	}

	var weightInputs []weight_types.WeightInput
	err = json.Unmarshal(bytes, &weightInputs)
	if err != nil {
		panic(err)
	}

	return weightInputs, dataAge
}

func (spec *weightSpecInternal) prepareDataGrid(taskPool *util_async.NestedTaskPoolChild, tracker *util.TrackProgress, cancel util_async.CancelSignal) ([]weight_types.WeightInput, error) {
	param := spec.param

	// READ IN ANY RECENT DATA
	tempPathGrid := files.TempData + "weightfind-sim-grid-" + param.Label + ".json"
	inputDataGrid, dataAgeGrid := readWeightInputFile(tempPathGrid)

	// DO WE ACCEPT THE OLD DATA
	if dataAgeGrid > c_simDataAgeMax && !spec.process.forceSkipSim {
		inputDataGrid = nil
	}

	// SIMULATE STAT CHANGES, SAVE SIM DATA IN CASE WE NEED TO RESTART
	if inputDataGrid == nil {
		taskPool.Go(func() {
			currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(param.GearFile), &param.Model, setup.MissingEnchant_Panic, spec.process.printer)
			currentItemSet := items.FullItemSet_FromMap(currentEquip)
			data, err := weightfind.SimulateSteppedStatChangesForGrid(currentItemSet, spec.process.printer, spec.process.simSpeed,
				param.Model.SimSpeedUp, param.Model.StatsForWeighting, param.Model.Spec, param.Model.Goal,
				param.Model.SimulateAs,
				param.Model.Professions, tracker.NewChild(), param.Label, cancel,
				param.FixStatsMode, c_eachSimTargetGenerateDataCount)
			if err != nil {
				return nil, err
			}

			inputDataGrid = data
			err = weight_types.WeightInputWriteFile(inputDataGrid, tempPathGrid)
			if err != nil {
				return nil, err
			}
		})
	} else {
		tracker.NewChild().SetDone()
	}

	spec.inputs.dataGrid = inputDataGrid
	return inputDataGrid, nil
}

func (spec *weightSpecInternal) prepareDataRandom(taskPool *util_async.NestedTaskPoolChild, tracker *util.TrackProgress, cancel util_async.CancelSignal) ([]weight_types.WeightInput, error) {
	param := spec.param

	// READ IN ANY RECENT DATA
	tempPathReal := files.TempData + "weightfind-sim-real-" + param.Label + ".json"
	inputDataReal, dataAgeReal := readWeightInputFile(tempPathReal)

	// DO WE ACCEPT THE OLD DATA
	if dataAgeReal > c_simDataAgeMax && !spec.process.forceSkipSim {
		inputDataReal = nil
	}

	// SIMULATE STAT CHANGES, SAVE SIM DATA IN CASE WE NEED TO RESTART
	if inputDataReal == nil {
		data, err := weightfind.SimulateRealRandomSets(param.GearFile, param.SubstituteItems, &param.Model, c_eachSimTargetGenerateDataCount,
			spec.process.simSpeed, param.FixStatsMode, spec.process.printer, tracker.NewChild(), param.Label, cancel)
		if err != nil {
			return nil, err
		}

		inputDataReal = data
		err = weight_types.WeightInputWriteFile(inputDataReal, tempPathReal)
		if err != nil {
			return nil, err
		}
	} else {
		tracker.NewChild().SetDone()
	}

	spec.inputs.dataRand = inputDataReal
	return inputDataReal, nil
}

func (spec *weightSpecInternal) prepareDataFit(taskPool *util_async.NestedTaskPoolChild, tracker *util.TrackProgress, cancel util_async.CancelSignal) ([]weight_types.WeightInput, error) {
	param := spec.param

	// READ IN ANY RECENT DATA
	tempPathFit := files.TempData + "weightfind-sim-fit-" + param.Label + ".json"
	inputDataFit, dataAgeFit := readWeightInputFile(tempPathFit)
	// DO WE ACCEPT THE OLD DATA
	if dataAgeFit > c_simDataAgeMax && !spec.process.forceSkipSim {
		inputDataFit = nil
	}
	// SIMULATE STAT CHANGES, SAVE SIM DATA IN CASE WE NEED TO RESTART
	if inputDataFit == nil {
		currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(spec.param.GearFile), &spec.param.Model, setup.MissingEnchant_Panic, spec.process.printer)
		currentItemSet := items.FullItemSet_FromMap(currentEquip)
		data, err := weightfind.SimulateSteppedStatChangesForFitting(currentItemSet, spec.process.printer, spec.process.simSpeed,
			param.Model.SimSpeedUp, param.Model.StatsForWeighting, param.Model.Spec, param.Model.Goal, param.Model.SimulateAs,
			param.Model.Professions, tracker.NewChild(), param.Label, cancel)
		if err != nil {
			return nil, err
		}

		inputDataFit = data
		err = weight_types.WeightInputWriteFile(inputDataFit, tempPathFit)
		if err != nil {
			return nil, err
		}
	} else {
		tracker.NewChild().SetDone()
	}
	spec.inputs.dataFit = inputDataFit
	return inputDataFit, nil
}
