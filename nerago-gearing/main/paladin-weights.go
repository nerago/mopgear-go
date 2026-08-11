package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model/model_factory"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/weightfind"
	"paladin_gearing_go/weightfind/weight_types"
)

//goland:noinspection GoBoolExpressions
func statWeights_updateAll() {
	printer := util.PrintRecorder_CreateLogFileNamed(files.LogOutputPath, "statWeights_updateAll")

	//simSpeed := simulate.RunSize_TestOnly
	// simSpeed := simulate.RunSize_QuickDirty/10
	//simSpeed := simulate.RunSize_QuickDirty
	simSpeed := simulate.RunSize_Common
	//simSpeed := simulate.RunSize_Largish
	forceSkipSim := true
	fixStats := weight_types.FixStatsRangeMode_ExpertiseAlways | weight_types.FixStatsRangeMode_HasteGridOnly | weight_types.FixStatsRangeMode_HasteHigherOnly

	process := weightfind.WeightUpdateProcess{}
	process.Init(simSpeed, forceSkipSim, printer)
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Mitigation-NoSet",
		WeightFile1:     files.WeightMitiNoSetFile,
		GearFile:        files.GearFileProtMitigationNoSet,
		Model:           model_factory.Model_PallyProtMitigation_NoSet(),
		SubstituteItems: substituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Mitigation-WithSet",
		WeightFile1:     files.WeightMitiWithSetFile,
		GearFile:        files.GearFileProtMitigationWithSet,
		Model:           model_factory.Model_PallyProtMitigation_WithSet(),
		SubstituteItems: substituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Damage",
		WeightFile1:     files.WeightDpsFile,
		GearFile:        files.GearFileProtDps,
		Model:           model_factory.Model_PallyProtDps(),
		SubstituteItems: substituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Compromise",
		WeightFile1:     files.WeightCompromiseFile,
		GearFile:        files.GearFileProtCompromise,
		Model:           model_factory.Model_PallyProtCompromise(),
		SubstituteItems: substituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Heal",
		WeightFile1:     files.WeightHealFile,
		GearFile:        files.GearFileProtHeal,
		Model:           model_factory.Model_PallyProtHeal(),
		SubstituteItems: substituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Ret",
		WeightFile1:     files.WeightRetFile,
		GearFile:        files.GearFileRet,
		Model:           model_factory.Model_PallyRet(),
		SubstituteItems: substituteItemsRet,
		FixStatsMode:    weight_types.FixStatsRangeMode_None,
	})

	cancel := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancel)
	process.Run(cancel)
}
