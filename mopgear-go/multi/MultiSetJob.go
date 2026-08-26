package multi

import (
	"strings"

	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/util"
)

type MainJob struct {
	printer     *util.PrintRecorder
	input       multi_types.JobInputs
	csvFilename string
	setFilename string

	tasks      []multi_types.JobInputTask
	itemPrep   map[string]*specItemPrep
	workGroups []*workingGroup
	bagsGear   loaders.EquippedArray
}

func JobCreate(printer *util.PrintRecorder, input multi_types.JobInputs, tasks ...multi_types.JobInputTask) *MainJob {
	if input.TimeLimitEachSolve == 0 {
		input.TimeLimitEachSolve = c_defaultTimeoutSeconds
	}
	return &MainJob{
		printer:     printer,
		input:       input,
		tasks:       tasks,
		csvFilename: replaceSuffix(printer.GetFileName(), ".csv"),
		setFilename: replaceSuffix(printer.GetFileName(), ".set"),
	}
}

func replaceSuffix(original string, suffix string) string {
	if len(original) < 4 {
		return ""
	}
	if main, hasLog := strings.CutSuffix(original, ".log"); hasLog {
		return main + suffix
	} else if original[len(original)-4] == '.' && original[len(original)-4:] != suffix {
		return original[len(original)-4:] + suffix
	} else {
		return original + suffix
	}
}
