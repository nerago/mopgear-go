package cmd

import (
	"io"
	"log"

	"github.com/nerago/mopgear-go/db"
	"github.com/nerago/mopgear-go/util"
	"github.com/wowsims/mop/sim"
)

func CommandSetupCommon() {
	util.CurrentProcessLowerPriority()
	db.WowSimDB_Read()
	sim.RegisterAll()
	log.SetOutput(io.Discard) // ignore wowsim's internal progress logs
}
