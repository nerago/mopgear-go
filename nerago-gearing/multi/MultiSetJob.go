package multi

import (
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/weightfind/weight_types"
)

type MultiSetJob struct {
	printer *util.PrintRecorder
	input   multi_types.JobInputs

	itemPrep map[string]*specItemPrep
	working  util_collection.MapMap[string, weight_types.WeightType, *specWorking]
	bagsGear loaders.EquippedArray
}

func JobCreate(printer *util.PrintRecorder, input multi_types.JobInputs) MultiSetJob {
	if input.TimeLimitEachSolve == 0 {
		input.TimeLimitEachSolve = c_defaultTimeoutSeconds
	}
	return MultiSetJob{
		printer: printer,
		input:   input,
	}
}
