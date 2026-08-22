package multi

import (
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/util"
)

type MultiSetJob struct {
	printer *util.PrintRecorder
	input   multi_types.JobInputs

	tasks      []multi_types.JobInputTask
	itemPrep   map[string]*specItemPrep
	workGroups []*workingGroup
	bagsGear   loaders.EquippedArray
}

func JobCreate(printer *util.PrintRecorder, input multi_types.JobInputs, tasks ...multi_types.JobInputTask) *MultiSetJob {
	if input.TimeLimitEachSolve == 0 {
		input.TimeLimitEachSolve = c_defaultTimeoutSeconds
	}
	return &MultiSetJob{
		printer: printer,
		input:   input,
		tasks:   tasks,
	}
}
