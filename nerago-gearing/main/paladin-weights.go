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
	//simSpeed := simulate.RunSize_Common
	simSpeed := simulate.RunSize_Largish
	forceSkipSim := false
	fixStats := weight_types.FixStatsRangeMode_ExpertiseAlways | weight_types.FixStatsRangeMode_HasteGridOnly | weight_types.FixStatsRangeMode_HasteHigherOnly

	process := weightfind.WeightUpdateProcess{}
	process.Init(simSpeed, forceSkipSim, 500, printer)
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Ret",
		WeightFile1:     files.WeightRetFile,
		GearFile:        files.GearFileRet,
		Model:           model_factory.Model_PallyRet(),
		SubstituteItems: substituteItemsRet,
		FixStatsMode:    weight_types.FixStatsRangeMode_None,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Damage",
		WeightFile1:     files.WeightDamageFile,
		GearFile:        files.GearFileProtDamage,
		Model:           model_factory.Model_PallyProtDamage(),
		SubstituteItems: substituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Balanced",
		WeightFile1:     files.WeightBalancedFile,
		GearFile:        files.GearFileProtBalanced,
		Model:           model_factory.Model_PallyProtBalanced(),
		SubstituteItems: substituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Mitigation",
		WeightFile1:     files.WeightMitigationFile,
		GearFile:        files.GearFileProtMitigation,
		Model:           model_factory.Model_PallyProtMitigation(),
		SubstituteItems: substituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Survival",
		WeightFile1:     files.WeightSurvivalFile,
		GearFile:        files.GearFileProtSurvival,
		Model:           model_factory.Model_PallyProtSurvival(),
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
	cancel := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancel)
	process.Run(cancel)
}
