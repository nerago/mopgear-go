package main

import (
	"github.com/nerago/mopgear-go/cmd"
	"github.com/nerago/mopgear-go/util"
)

func main() {
	cmd.CommandSetupCommon()

	prof := util.CmdProfilingStart("upgrade")
	defer prof.Finish()

	findUpgrades_Paladin()
}
