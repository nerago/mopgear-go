package model_factory

import (
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func SetupModelWeights(model *gear_model.SpecModel) error {
	sampleInputs, err := weight_types.WeightInputReadFileMultiple(
		model.GetSampleFileRand(), model.GetSampleFileGrid(), model.GetSampleFileFit(),
	)
	if err != nil {
		return err
	}
	return SetupModelWeightsHaveSamples(model, sampleInputs)
}

func SetupModelWeightsHaveSamples(model *gear_model.SpecModel, sampleInputs []weight_types.WeightInput) error {
	model.InitDerives()

	weight, err := tools.StatRatingsWeightsExtended_ReadFile(model.WeightFile, model.StatsForWeighting, sampleInputs)
	if err != nil {
		return err
	}
	model.SetStatWeights(weight)
	return nil
}
