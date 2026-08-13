package main

import (
	"io"
	"log"
	"os"
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/util"
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
		f, err := os.CreateTemp("", "main-new.pgo")
		if err != nil {
			panic(err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			panic(err)
		}
		defer func() {
			pprof.StopCPUProfile()
			err := f.Close()
			if err != nil {
				panic(err)
			}
			//if simulate.WowSimRanDuringCurrentProcess {
			err = os.Rename(f.Name(), files.ProfileDir+"main.pgo")
			if err != nil {
				panic(err)
			}
			//}
		}()
	}

	startTime := time.Now()
	printer.Println("Started at " + startTime.Format(time.DateTime))

	core(printer)

	timeTaken := time.Since(startTime)
	printer.Println("Duration = " + timeTaken.String())
	printer.Println("Finished at " + time.Now().Format(time.DateTime))
}

func core(printer *util.PrintRecorder) {
	//slotRating(printer)
	//basicReforge(printer)
	//findT5BIS(printer)
	//findT5TrinketPermutations(printer)
	//findT5WeightPermutations(printer)
	//statWeightsGridFromInitialT5(printer)
	//basicListRatingEach(printer)
	//solveForRatings(printer)
	//findBestSubjectToCommon(printer)
	//checkHighs(printer)
	//
	//testSim(printer)
	//findUpgrades_Sim_PaladinMiti_Run(printer)
	//findUpgrades_Sim_PaladinDps_Run(printer)
	//findSimpleUpgrade(printer)
	//findSimpleUpgrade_ForceEach(printer)
	//relativeRatingsCompromise(printer)
	//trinketSimsBoth(printer)
	//currentSimGear(printer)
	//determineSetBonusValueBySim()
	//determineBestUseOfGearSets()
	//
	//statWeightsBasic(printer)
	//statWeightsGrid1Orig(printer)
	//statWeightsGrid1b(printer)
	//statWeightsFitting(printer)
	//statWeightsFitting1eachProper(printer)
	//statWeightsFitting2(printer)
	//statWeightsFitting2eachProper(printer)
	//statWeightsFitting2each(printer)
	//statWeightsFitting3eachProper(printer)
	//statWeightsFormula(printer)
	//statWeightsRanking(printer)
	//statWeightsRanking3b(printer)
	//statWeightsGridIntoRanking(printer)
	//statWeightsCustom(printer)

	//statWeights_CompareAlgorithms()

	//statWeights_updateAll()

	PaladinMultiRun()
	//PaladinMultiRunLite()

	//findUpgrades_Paladin()
}
