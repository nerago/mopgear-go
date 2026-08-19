package util

import (
	"os"
	"runtime/pprof"

	"github.com/nerago/mopgear-go/files"
)

type CmdProfiling struct {
	file        *os.File
	commandName string
}

func CmdProfilingStart(commandName string) *CmdProfiling {
	prof := &CmdProfiling{commandName: commandName}

	var err error
	prof.file, err = os.CreateTemp("", "cpu-new.pgo")
	if err != nil {
		panic(err)
	}

	err = pprof.StartCPUProfile(prof.file)
	if err != nil {
		panic(err)
	}

	return prof
}

func (prof *CmdProfiling) Finish() {
	pprof.StopCPUProfile()

	err := prof.file.Close()
	if err != nil {
		panic(err)
	}

	err = os.Rename(prof.file.Name(), files.ProfileDir+prof.commandName+".pgo")
	if err != nil {
		panic(err)
	}
}
