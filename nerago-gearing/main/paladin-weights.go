package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind"
)

func statWeights_updateAll() {
	printer := util.PrintRecorder_CreateLogFileNamed(files.LogOutputPath, "statWeightsGrid_updateAll")

	//simSpeed := simulate.RunSize_TestOnly
	// simSpeed := simulate.RunSize_QuickDirty/10
	//simSpeed := simulate.RunSize_QuickDirty
	simSpeed := simulate.RunSize_Common
	//simSpeed := simulate.RunSize_Largish
	forceSkipSim := false

	weightfind.StatWeights_updateAll(simSpeed, forceSkipSim, printer, []weightfind.WeightOptions{
		{
			Label:           "Prot-Mitigation-NoSet",
			WeightFileOut:   files.WeightMitiNoSetFile,
			GearFile:        files.GearFileProtMitigationNoSet,
			Model:           gear_model.Model_PallyProtMitigation_NoSet(),
			SubstituteItems: substituteItemsProt,
		},
		{
			Label:           "Prot-Mitigation-WithSet",
			WeightFileOut:   files.WeightMitiWithSetFile,
			GearFile:        files.GearFileProtMitigationWithSet,
			Model:           gear_model.Model_PallyProtMitigation_WithSet(),
			SubstituteItems: substituteItemsProt,
		},
		{
			Label:           "Prot-Damage",
			WeightFileOut:   files.WeightDpsFile,
			GearFile:        files.GearFileProtDps,
			Model:           gear_model.Model_PallyProtDps(),
			SubstituteItems: substituteItemsProt,
		},
		{
			Label:           "Prot-Compromise",
			WeightFileOut:   files.WeightCompromiseFile,
			GearFile:        files.GearFileProtCompromise,
			Model:           gear_model.Model_PallyProtCompromise(),
			SubstituteItems: substituteItemsProt,
		},
		{
			Label:           "Prot-Heal",
			WeightFileOut:   files.WeightHealFile,
			GearFile:        files.GearFileProtHeal,
			Model:           gear_model.Model_PallyProtHeal(),
			SubstituteItems: substituteItemsProt,
		},
		{
			Label:           "Ret",
			WeightFileOut:   files.WeightRetFile,
			GearFile:        files.GearFileRet,
			Model:           gear_model.Model_PallyRet(),
			SubstituteItems: substituteItemsRet,
		},
	})
}
