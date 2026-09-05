package model_factory

import (
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func SetupModel(model *gear_model.SpecModel, verificationInputs []weight_types.WeightInput, requiredStats []stats.StatType) *gear_model.SpecModel {
	weight, err := tools.StatRatingsWeightsExtended_ReadFile(model.WeightFile, verificationInputs, requiredStats)
	if err != nil {
		panic(err)
	}
	model.StatWeights = weight

	model.ModelItems.Init()

	model.Initialized = true
}
