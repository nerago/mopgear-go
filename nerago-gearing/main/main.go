package main

import (
	"io"
	"log"
	"os"
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/util"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/wowsims/mop/sim"
)

const (
	enableProfiling = true
)

var printer *util.PrintRecorder

func main() {
	util.CurrentProcessLowerPriority()

	printer = util.PrintRecorder_CreateLogFile(files.LogOutputPath)
	defer printer.Close()

	db.WowSimDB_Read()
	sim.RegisterAll()

	log.SetOutput(io.Discard) // ignore wowsim's internal progress logs

	if enableProfiling {
		f, err := os.Create(files.ProfileDir + "main-new.pgo")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer func() {
			pprof.StopCPUProfile()
			f.Close()
			os.Rename(f.Name(), files.ProfileDir+"main.pgo")
		}()
	}

	startTime := time.Now()
	printer.Println("Started at " + startTime.Format(time.RFC1123))

	core(printer)

	timeTaken := time.Since(startTime)
	printer.Println("Duration = " + timeTaken.String())
	printer.Println("Finished at " + time.Now().Format(time.RFC1123))

	if enableProfiling {
		f, err := os.Create(files.ProfileDir + "main-memory.pgo")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
			panic(err)
		}
	}
}

func core(printer *util.PrintRecorder) {

	// slotRating(printer)
	// basicReforge(printer)
	// findT5BIS(printer)
	// findT5TrinketPermutations(printer)
	// findT5WeightPermutations(printer)
	// statWeightsGridFromInitialT5(printer)
	// basicListRatingEach(printer)
	// solveForRatings(printer)
	// findBestSubjectToCommon(printer)
	// checkHighs(printer)
	// checkHighsAcross(printer)

	// PaladinMultiRun(printer)
	// testSim(printer)
	// findUpgrades_Sim_PaladinMiti_Run(printer)
	// findUpgrades_Sim_PaladinDps_Run(printer)
	// findUpgrades_Paladin_Sim_AllRaid_Run(printer)
	// findSimpleUpgrade_ForceEach(printer)
	// findMitigationWithCapicitance(printer)
	// relativeRatingsCompromise(printer)
	// trinketSims(printer)

	// forBasicStatsGenerateRatingsDataFromSims(printer)
	// forSpreadsheetGenerateRatingsDataFromSims(printer)
	// statWeightsFromHighAndSim(printer)
	// statWeightsBasic(printer)
	// statWeightsGrid(printer)
	statWeightsFitting2(printer)
}

func setupPallyMitigationSet() (items.FullOptionsMap, model.Model) {
	model := model.Model_PallyProtMitigation_WithSet()
	return setup.OptionsSetup_FromGearFile(files.GearFileProtMitigationWithSet, &model, setup.MissingEnchant_Panic, printer), model
}

func setupPallyMitigationNoSet() (items.FullOptionsMap, model.Model) {
	model := model.Model_PallyProtMitigation_WithSet()
	return setup.OptionsSetup_FromGearFile(files.GearFileProtMitigationNoSet, &model, setup.MissingEnchant_Panic, printer), model
}

func setupPallyDps() (items.FullOptionsMap, model.Model) {
	model := model.Model_PallyProtDps()
	return setup.OptionsSetup_FromGearFile(files.GearFileProtDps, &model, setup.MissingEnchant_Panic, printer), model
}
