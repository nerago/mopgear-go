package main

import (
	"os"
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/setup"
	. "paladin_gearing_go/util"
	"runtime/pprof"
	"time"
)

const (
	gearFileMiti    = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-prot-defence.json`
	gearFileDps     = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-prot-dps.json`
	enableProfiling = true
)

var printer = PrintRecorder{}

func main() {
	startTime := time.Now()

	if enableProfiling {
		f, err := os.Create("default.pgo")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	core()

	timeTaken := time.Since(startTime)
	printer.Println("Duration = " + timeTaken.String())
}

func core() {
	itemOptions, model := setupPallyMitigation()
	// itemOptions, model := setupPallyDps()

	// slotRating(itemOptions[Equip_Chest], &model)
	basicReforge(&itemOptions, &model, &printer)
}

func setupPallyMitigation() (FullOptionsMap, Model) {
	model := Model_PallyProtMitigation()
	return OptionsSetup_FromGearFile(gearFileMiti, &model, &printer), model
}

func setupPallyDps() (FullOptionsMap, Model) {
	model := Model_PallyProtDps()
	return OptionsSetup_FromGearFile(gearFileDps, &model, &printer), model
}
