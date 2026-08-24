package main

import (
	"github.com/nerago/mopgear-go/cmd/mygear"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/gear_model/model_factory"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/weightfind"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

//goland:noinspection GoBoolExpressions
func statWeights_updateAll(printer *util.PrintRecorder) {

	//simSpeed := simulate.RunSize_TestOnly
	// simSpeed := simulate.RunSize_QuickDirty/10
	//simSpeed := simulate.RunSize_QuickDirty
	//simSpeed := simulate.RunSize_Common
	simSpeed := simulate.RunSize_Largish
	forceSkipSim := true
	skipSolve := true
	fixStats := weight_types.FixStatsRangeMode_ExpertiseAlways | weight_types.FixStatsRangeMode_HasteGridOnly | weight_types.FixStatsRangeMode_HasteHigherOnly

	process := weightfind.WeightUpdateProcess{}
	process.Init(simSpeed, forceSkipSim, skipSolve, 900, printer)
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Ret",
		WeightFile1:     files.WeightRet,
		GearFile:        files.GearFileRet,
		Model:           model_factory.Model_PallyRet(),
		SubstituteItems: mygear.SubstituteItemsRet,
		FixStatsMode:    weight_types.FixStatsRangeMode_None,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Damage",
		WeightFile1:     files.WeightProtDamage,
		GearFile:        files.GearFileProtDamage,
		Model:           model_factory.Model_PallyProtDamage(),
		SubstituteItems: mygear.SubstituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Balanced",
		WeightFile1:     files.WeightProtBalanced,
		GearFile:        files.GearFileProtBalanced,
		Model:           model_factory.Model_PallyProtBalanced(),
		SubstituteItems: mygear.SubstituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Mitigation",
		WeightFile1:     files.WeightProtMitigation,
		GearFile:        files.GearFileProtMitigation,
		Model:           model_factory.Model_PallyProtMitigation(),
		SubstituteItems: mygear.SubstituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Survival",
		WeightFile1:     files.WeightProtSurvival,
		GearFile:        files.GearFileProtSurvival,
		Model:           model_factory.Model_PallyProtSurvival(),
		SubstituteItems: mygear.SubstituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpec(&weightfind.WeightSpec{
		Label:           "Prot-Heal",
		WeightFile1:     files.WeightProtHeal,
		GearFile:        files.GearFileProtHeal,
		Model:           model_factory.Model_PallyProtHeal(),
		SubstituteItems: mygear.SubstituteItemsProt,
		FixStatsMode:    fixStats,
	})

	cancel := util_async.CancelSignal_Make()
	//util_async.CancelOnKeyPress(cancel)
	//process.Run(cancel)

	ratioUpdate := weightfind.WeightRatioProcess{}
	ratioUpdate.Init(3000, printer)
	ratioUpdate.AddSpecsFrom(&process)
	ratioUpdate.Run(cancel)
}
