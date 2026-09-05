package updateProc

import (
	"slices"

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

func (spec *weightSpecInternal) loadInputData(tempPath string) (bool, *util_async.Future[[]weight_types.WeightInput], error) {
	futureData := util_async.Future_Make[[]weight_types.WeightInput]()

	// READ IN ANY RECENT DATA
	inputData, dataAge, err := weight_types.WeightInputReadFileAndCheckAge(tempPath)
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
	tempPathGrid := spec.param.Model.ModelItems.GetSampleFileGrid()
	done, futureData, errLoad := spec.loadInputData(tempPathGrid)
	if done {
		tracker.SetDone()
		return futureData, errLoad
	}

	// SIMULATE STAT CHANGES, SAVE SIM DATA IN CASE WE NEED TO RESTART
	taskPool.Go(func() error {
		param := spec.param
		currentEquip, err := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(loaders.GearFileReader_Read(param.Model.GearFile), &param.Model, setup.MissingEnchant_Panic, spec.process.printer)
		if err != nil {
			return err
		}
		currentItemSet := items.FullItemSet_FromMap(currentEquip)
		data, err2 := weightfind.SimulateSteppedStatChangesForGrid(currentItemSet, spec.process.printer, spec.process.simSpeed,
			param.Model.SimSpeedUp, param.Model.StatsForWeighting, param.Model.Spec, param.Model.Goal,
			param.Model.SimulateAs,
			param.Model.Professions, tracker, spec.label, cancel,
			param.FixStatsMode, c_eachSimTargetGenerateDataCount)
		return spec.sendSimData(futureData, data, tempPathGrid, err2)
	})
	return futureData, nil
}

func (spec *weightSpecInternal) prepareDataRandom(taskPool *util_async.NestedTaskPoolChild, tracker *util.TrackProgress, cancel util_async.CancelSignal) (*util_async.Future[[]weight_types.WeightInput], error) {
	// READ IN ANY RECENT DATA
	tempPathReal := spec.param.Model.ModelItems.GetSampleFileRand()
	done, futureData, err := spec.loadInputData(tempPathReal)
	if done {
		tracker.SetDone()
		return futureData, err
	}

	// SIMULATE STAT CHANGES, SAVE SIM DATA IN CASE WE NEED TO RESTART
	taskPool.Go(func() error {
		param := spec.param
		data, err2 := weightfind.SimulateRealRandomSets(param.Model.GearFile, param.SubstituteItems, &param.Model, c_eachSimTargetGenerateDataCount,
			spec.process.simSpeed, param.FixStatsMode, spec.process.printer, tracker, spec.label, cancel)
		return spec.sendSimData(futureData, data, tempPathReal, err2)
	})
	return futureData, nil
}

func (spec *weightSpecInternal) prepareDataFit(taskPool *util_async.NestedTaskPoolChild, tracker *util.TrackProgress, cancel util_async.CancelSignal) (*util_async.Future[[]weight_types.WeightInput], error) {
	// READ IN ANY RECENT DATA
	tempPathFit := spec.param.Model.ModelItems.GetSampleFileFit()files.TempData + "weightfind-sim-fit-" + spec.label + ".json"
	done, futureData, errLoad := spec.loadInputData(tempPathFit)
	if done {
		tracker.SetDone()
		return futureData, errLoad
	}

	// SIMULATE STAT CHANGES, SAVE SIM DATA IN CASE WE NEED TO RESTART
	taskPool.Go(func() error {
		param := spec.param
		currentEquip, err := setup.OptionsSetup_FromEquipped_OriginalForgeOnly(loaders.GearFileReader_Read(param.Model.GearFile), &spec.param.Model, setup.MissingEnchant_Panic, spec.process.printer)
		if err != nil {
			return err
		}
		currentItemSet := items.FullItemSet_FromMap(currentEquip)
		data, err2 := weightfind.SimulateSteppedStatChangesForFitting(currentItemSet, spec.process.printer, spec.process.simSpeed,
			param.Model.SimSpeedUp, param.Model.StatsForWeighting, param.Model.Spec, param.Model.Goal, param.Model.SimulateAs,
			param.Model.Professions, tracker, spec.label, cancel)
		return spec.sendSimData(futureData, data, tempPathFit, err2)
	})
	return futureData, nil
}
