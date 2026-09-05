package main

import (
	"github.com/nerago/mopgear-go/cmd/mygear"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/gear_model/model_factory"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/weightfind/updateProc"
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
	skipSolve := false
	fixStats := weight_types.FixStatsRangeMode_ExpertiseAlways | weight_types.FixStatsRangeMode_HasteGridOnly | weight_types.FixStatsRangeMode_HasteHigherOnly
	_ = fixStats

	process := updateProc.WeightUpdateProcess{}
	process.Init(simSpeed, forceSkipSim, skipSolve, 3600, printer)

	process.AddSpecParam(updateProc.SpecParam{
		WeightFile1:     files.WeightRet,
		Model:           model_factory.Model_PallyRet(),
		SubstituteItems: mygear.SubstituteItemsRet,
		FixStatsMode:    weight_types.FixStatsRangeMode_None,
	})
	process.AddSpecParam(updateProc.SpecParam{
		WeightFile1:     files.WeightProtDamage,
		Model:           model_factory.Model_PallyProtDamage(),
		SubstituteItems: mygear.SubstituteItemsProt,
		FixStatsMode:    fixStats,
	})

	process.AddSpecParam(updateProc.SpecParam{
		WeightFile1:     files.WeightProtBalanced,
		Model:           model_factory.Model_PallyProtBalanced(),
		SubstituteItems: mygear.SubstituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpecParam(updateProc.SpecParam{
		WeightFile1:     files.WeightProtMitigation,
		Model:           model_factory.Model_PallyProtMitigation(),
		SubstituteItems: mygear.SubstituteItemsProt,
		FixStatsMode:    fixStats,
	})

	process.AddSpecParam(updateProc.SpecParam{
		WeightFile1:     files.WeightProtSurvival,
		Model:           model_factory.Model_PallyProtSurvival(),
		SubstituteItems: mygear.SubstituteItemsProt,
		FixStatsMode:    fixStats,
	})
	process.AddSpecParam(updateProc.SpecParam{
		WeightFile1:     files.WeightProtHeal,
		Model:           model_factory.Model_PallyProtHeal(),
		SubstituteItems: mygear.SubstituteItemsProt,
		FixStatsMode:    fixStats,
	})

	cancel := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancel)

	process.Run(cancel, 6)

	//ratioUpdate := weightfind.WeightRatioProcess{}
	//ratioUpdate.Init(300, printer)
	//ratioUpdate.AddSpecsFrom(&process)
	//ratioUpdate.Run(cancel)
}
