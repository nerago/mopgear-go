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

	simSpeed := simulate.RunSize_TestOnly
	//simSpeed := simulate.RunSize_QuickDirty
	// simSpeed := simulate.RunSize_QuickDirty/10
	//simSpeed := simulate.RunSize_Largish

	weightfind.StatWeights_updateAll(simSpeed, printer, []weightfind.WeightOptions{
		{
			Label:           "Prot-Mitigation-NoSet",
			WeightFileOut:   files.WeightMitiNoSetFile,
			GearFile:        files.GearFileProtMitigationNoSet,
			Model:           gear_model.Model_PallyProtMitigation_NoSet(),
			SubstituteItems: substituteItemsMiti,
		},
		//{
		//	Label:           "Prot-Mitigation-WithSet",
		//	WeightFileOut:   files.WeightMitiWithSetFile,
		//	GearFile:        files.GearFileProtMitigationWithSet,
		//	Model:           gear_model.Model_PallyProtMitigation_WithSet(),
		//	SubstituteItems: substituteItemsMiti,
		//},
		//{
		//	Label:           "Prot-Damage",
		//	WeightFileOut:   files.WeightDpsFile,
		//	GearFile:        files.GearFileProtDps,
		//	Model:           gear_model.Model_PallyProtDps(),
		//	SubstituteItems: substituteItemsDps,
		//},
		//{
		//	Label:           "Prot-Compromise",
		//	WeightFileOut:   files.WeightCompromiseFile,
		//	GearFile:        files.GearFileProtCompromise,
		//	Model:           gear_model.Model_PallyProtCompromise(),
		//	SubstituteItems: util.RemoveDuplicatesComparable(slices.Concat(substituteItemsDps, substituteItemsMiti)),
		//},
		//{
		//	Label:           "Prot-Heal",
		//	WeightFileOut:   files.WeightHealFile,
		//	GearFile:        files.GearFileProtHeal,
		//	Model:           gear_model.Model_PallyProtHeal(),
		//	SubstituteItems: util.RemoveDuplicatesComparable(slices.Concat(substituteItemsDps, substituteItemsMiti)),
		//},
		//{
		//	Label:           "Ret",
		//	WeightFileOut:   files.WeightRetFile,
		//	GearFile:        files.GearFileRet,
		//	Model:           gear_model.Model_PallyRet(),
		//	SubstituteItems: substituteItemsRet,
		//},
	})
}
