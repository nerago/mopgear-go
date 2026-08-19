package main

import (
	"github.com/nerago/mopgear-go/cmd"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/util"
)

func main() {
	cmd.CommandSetupCommon()

	printer := util.PrintRecorder_CreateLogFileNamed(files.LogOutputPath, "statWeights_updateAll")
	defer printer.Close()

	prof := util.CmdProfilingStart("weights")
	defer prof.Finish()

	sw := util.StopwatchNoisyStart(printer)
	defer sw.Stop()

	statWeights_updateAll(printer)
}
