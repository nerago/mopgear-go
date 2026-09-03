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
	taskPoolSim.SetContinueOnError(false)

	dataGridFuture, err := spec.prepareDataGrid(taskPoolSim, tracker.NewChild(), cancel)
	if err != nil {
		return err
	}

	dataRandFuture, err2 := spec.prepareDataRandom(taskPoolSim, tracker.NewChild(), cancel)
	if err2 != nil {
		return err2
	}

	dataFitFuture, err3 := spec.prepareDataFit(taskPoolSim, tracker.NewChild(), cancel)
	if err3 != nil {
		return err3
	}

	err4 := taskPoolSim.WaitAllComplete()
	if err4 != nil {
		return err4
	}

	spec.inputs.dataGrid = dataGridFuture.WaitForResultOrNilValue()
	spec.inputs.dataRand = dataRandFuture.WaitForResultOrNilValue()
	spec.inputs.dataFit = dataFitFuture.WaitForResultOrNilValue()
	spec.inputs.dataAll = slices.Concat(spec.inputs.dataGrid, spec.inputs.dataRand, spec.inputs.dataFit)
	return nil
}

func readWeightInputFile(filename string) ([]weight_types.WeightInput, time.Duration, error) {
	statInfo, err := os.Stat(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, 0, nil
	} else if err != nil {
		return nil, 0, err
	}

	// only use data from "today"
	dataAge := time.Since(statInfo.ModTime())

	bytes, err := os.ReadFile(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, 0, nil
	} else if err != nil {
		return nil, 0, err
	}

	var weightInputs []weight_types.WeightInput
	err = json.Unmarshal(bytes, &weightInputs)
	if err != nil {
		return nil, 0, err
	}

	return weightInputs, dataAge, nil
}

func (spec *weightSpecInternal) loadInputData(tempPath string) (bool, *util_async.Future[[]weight_types.WeightInput], error) {
	futureData := util_async.Future_Make[[]weight_types.WeightInput]()

	// READ IN ANY RECENT DATA
	inputData, dataAge, err := readWeightInputFile(tempPath)
	if err != nil {
		return true, nil, err
	}

	// DO WE ACCEPT THE OLD DATA
	if inputData != nil && (dataAge < c_simDataAgeMax || spec.process.forceSkipSim) {
		futureData.SetResult(inputData)
		return true, futureData, nil
	}
	return false, futureData, nil
}

func (spec *weightSpecInternal) sendSimData(futureData *util_async.Future[[]weight_types.WeightInput], data []weight_types.WeightInput, tempPathGrid string, err2 error) error {
	if err2 != nil {
		return err2
	}

	err2 = weight_types.WeightInputWriteFile(data, tempPathGrid)
	if err2 != nil {
		return err2
	}

	futureData.SetResult(data)
	return nil
}

func (spec *weightSpecInternal) prepareDataGrid(taskPool *util_async.NestedTaskPoolChild, tracker *util.TrackProgress, cancel util_async.CancelSignal) (*util_async.Future[[]weight_types.WeightInput], error) {
	tempPathGrid := files.TempData + "weightfind-sim-grid-" + spec.param.Label + ".json"
	done, futureData, errLoad := spec.loadInputData(tempPathGrid)
	if done {
		tracker.SetDone()
		return futureData, errLoad
	}

	// SIMULATE STAT CHANGES, SAVE SIM DATA IN CASE WE NEED TO RESTART
	taskPool.Go(func() error {
		param := spec.param
		currentEquip, err := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(loaders.GearFileReader_Read(param.GearFile), &param.Model, setup.MissingEnchant_Panic, spec.process.printer)
		if err != nil {
			return err
		}
		currentItemSet := items.FullItemSet_FromMap(currentEquip)
		data, err2 := weightfind.SimulateSteppedStatChangesForGrid(currentItemSet, spec.process.printer, spec.process.simSpeed,
			param.Model.SimSpeedUp, param.Model.StatsForWeighting, param.Model.Spec, param.Model.Goal,
			param.Model.SimulateAs,
			param.Model.Professions, tracker, param.Label, cancel,
			param.FixStatsMode, c_eachSimTargetGenerateDataCount)
		return spec.sendSimData(futureData, data, tempPathGrid, err2)
	})
	return futureData, nil
}

func (spec *weightSpecInternal) prepareDataRandom(taskPool *util_async.NestedTaskPoolChild, tracker *util.TrackProgress, cancel util_async.CancelSignal) (*util_async.Future[[]weight_types.WeightInput], error) {
	// READ IN ANY RECENT DATA
	tempPathReal := files.TempData + "weightfind-sim-real-" + spec.param.Label + ".json"
	done, futureData, err := spec.loadInputData(tempPathReal)
	if done {
		tracker.SetDone()
		return futureData, err
	}

	// SIMULATE STAT CHANGES, SAVE SIM DATA IN CASE WE NEED TO RESTART
	taskPool.Go(func() error {
		param := spec.param
		data, err2 := weightfind.SimulateRealRandomSets(param.GearFile, param.SubstituteItems, &param.Model, c_eachSimTargetGenerateDataCount,
			spec.process.simSpeed, param.FixStatsMode, spec.process.printer, tracker, param.Label, cancel)
		return spec.sendSimData(futureData, data, tempPathReal, err2)
	})
	return futureData, nil
}

func (spec *weightSpecInternal) prepareDataFit(taskPool *util_async.NestedTaskPoolChild, tracker *util.TrackProgress, cancel util_async.CancelSignal) (*util_async.Future[[]weight_types.WeightInput], error) {
	// READ IN ANY RECENT DATA
	tempPathFit := files.TempData + "weightfind-sim-fit-" + spec.param.Label + ".json"
	done, futureData, errLoad := spec.loadInputData(tempPathFit)
	if done {
		tracker.SetDone()
		return futureData, errLoad
	}

	// SIMULATE STAT CHANGES, SAVE SIM DATA IN CASE WE NEED TO RESTART
	taskPool.Go(func() error {
		param := spec.param
		currentEquip, err := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(loaders.GearFileReader_Read(spec.param.GearFile), &spec.param.Model, setup.MissingEnchant_Panic, spec.process.printer)
		if err != nil {
			return err
		}
		currentItemSet := items.FullItemSet_FromMap(currentEquip)
		data, err2 := weightfind.SimulateSteppedStatChangesForFitting(currentItemSet, spec.process.printer, spec.process.simSpeed,
			param.Model.SimSpeedUp, param.Model.StatsForWeighting, param.Model.Spec, param.Model.Goal, param.Model.SimulateAs,
			param.Model.Professions, tracker, param.Label, cancel)
		return spec.sendSimData(futureData, data, tempPathFit, err2)
	})
	return futureData, nil
}
