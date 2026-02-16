package main

import (
	"fmt"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/setup"
	. "paladin_gearing_go/types/items"
	. "paladin_gearing_go/util"
	"time"
	// "os"
	// "runtime/pprof"
)

const (
	gearFileMiti = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-prot-defence.json`
	gearFileDps  = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-prot-dps.json`
)

var printer = PrintRecorder{}

func main() {
	startTime := time.Now()

	// f, err := os.Create("profile.out")
	// if err != nil {
	// 	panic(err)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	core()

	timeTaken := time.Since(startTime)
	fmt.Println("Duration = " + timeTaken.String())
}

func core() {
	// model, itemOptions := setupPallyMitigation()
	itemOptions, model := setupPallyDps()

	// slotRating(itemOptions[Equip_Chest], &model)
	basicReforge(&itemOptions, &model)
}

func setupPallyMitigation() (FullOptionsMap, Model) {
	model := Model_PallyProtMitigation()
	return OptionsSetup_FromGearFile(gearFileMiti, &model, &printer), model
}

func setupPallyDps() (FullOptionsMap, Model) {
	model := Model_PallyProtDps()
	return OptionsSetup_FromGearFile(gearFileDps, &model, &printer), model
}
