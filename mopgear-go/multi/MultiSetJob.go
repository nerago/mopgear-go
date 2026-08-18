package multi

import (
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type MultiSetJob struct {
	printer *util.PrintRecorder
	input   multi_types.JobInputs

	itemPrep   map[string]*specItemPrep
	workGroups map[weight_types.WeightType]*workingGroup
	bagsGear   loaders.EquippedArray
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
