package main

import (
	"github.com/nerago/mopgear-go/cmd"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/util"
)

func main() {
	cmd.CommandSetupCommon()

	printer := util.PrintRecorder_CreateLogFile(files.LogOutputPath)
	defer printer.Close()

	prof := util.CmdProfilingStart("misc")
	defer prof.Finish()

	sw := util.StopwatchNoisyStart(printer)
	defer sw.Stop()

	core(printer)
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
	//debugBadWeight(printer)
	//
	//testSim(printer)
	//findUpgrades_Sim_PaladinMiti_Run(printer)
	//findUpgrades_Sim_PaladinDps_Run(printer)
	//findSimpleUpgrade(printer)
	//findSimpleUpgrade_ForceEach(printer)
	//relativeRatingsCompromise(printer)
	trinketSimsBoth(printer)
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
	//statWeightsFormula3(printer)
	//statWeightsRanking(printer)
	//statWeightsRanking3a(printer)
	//statWeightsSearchRatio(printer)
	//statWeightsGridIntoRanking(printer)
	//statWeightsSearch(printer)
	//statWeightsSearchExtended(printer)
	//compareAccuracy(printer)
}
