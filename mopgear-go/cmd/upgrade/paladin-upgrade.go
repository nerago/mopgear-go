package main

import (
	"github.com/nerago/mopgear-go/cmd/mygear"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/gear_model/model_factory"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/upgrades"
	"github.com/nerago/mopgear-go/util/util_async"
)

const c_upgradeDefaultTimeout = 2000

func findUpgrades_Paladin() {
	//simSizeBaseline := simulate.RunSize_VerySlow
	//simSizeTopN := simulate.RunSize_VerySlow
	simSizeBaseline := simulate.RunSize_Largish
	simSizeTopN := simulate.RunSize_Largish
	//simSizeBaseline := simulate.RunSize_Common
	//simSizeTopN := simulate.RunSize_Common

	simSizePerItem := simulate.RunSize_QuickDirty
	//simSizePerItem := simulate.RunSize_Common
	//simSizePerItem := simulate.RunSize_Largish

	//simSizeBaseline := simulate.RunSize_TestOnly
	//simSizeTopN := simulate.RunSize_TestOnly
	//simSizePerItem := simulate.RunSize_TestOnly

	substituteEmptySlotOnly := make(map[items.SlotItem]items.ItemId)
	substituteEmptySlotOnly[items.Item_Trinket] = 94529 // gaze
	substituteEmptySlotOnly[items.Item_Ring] = 86957    // heroic bladed tempest ring

	//normalBossesConsider := []string{
	//	"SoO Immerseus",
	//	"SoO FallenProtectors",
	//	"SoO Norushen",
	//	"SoO ShaofPride",
	//	"SoO Galakras",
	//	"SoO IronJuggernaut",
	//	"SoO DarkShaman",
	//	"SoO Nazgrim",
	//	"SoO Malkorok",
	//	"SoO Spoils",
	//	"SoO Thok",
	//	"SoO Blackfuse",
	//	"SoO Paragons",
	//	"SoO Garrosh"}
	//heroicBossesConsider := []string{
	//	"SoO Immerseus",
	//	"SoO FallenProtectors",
	//	"SoO Norushen",
	//	"SoO ShaofPride",
	//	"SoO Galakras",
	//	"SoO IronJuggernaut",
	//	"SoO DarkShaman",
	//	"SoO Nazgrim",
	//	"SoO Malkorok",
	//	"SoO Spoils",
	//	"SoO Thok",
	//	"SoO Blackfuse",
	//	"SoO Paragons",
	//	"SoO Garrosh"
	//	}
	//finder := loaders.ItemFinder_HeroicBossFiltered(loaders.ItemFinder_SiegeStrengthPlateTank, heroicBossesConsider)
	//finder := loaders.ItemFinder_NormalHeroicBossFiltered(loaders.ItemFinder_SiegeStrengthPlateTank, normalBossesConsider, heroicBossesConsider)

	//finder := loaders.ItemFinder_SiegeStrengthPlateTank
	finder := loaders.ItemFinder_Ordos
	//finder := loaders.ItemFinder_TimelessPlate
	//finder := loaders.ItemFinder_BagsUpgraded
	//finder := loaders.SiegeClassGearSetMultiple(stats.Spec_PaladinProt, stats.Spec_PaladinRet)

	// celestial world
	//finder := func(_ stats.Difficulty) []loaders.ItemFoundRef {
	//	return []loaders.ItemFoundRef{{ItemId: 99127}, {ItemId: 99137}}
	//}

	// upgraded trash drops
	//finder := func(_ stats.Difficulty) []loaders.ItemFoundRef {
	//	return []loaders.ItemFoundRef{{ItemId: 105851}, {ItemId: 105850}}
	//}

	input := upgrades.FindUpgrades_MultiSpec_Sim{
		FindUpgrades_SimInputs: upgrades.FindUpgrades_SimInputs{
			FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
				IncludeCelestial:   false,
				IncludeNormal:      true,
				IncludeHeroic:      false,
				IgnoredItems:       mygear.IgnoredItems,
				TargetUpgradeLevel: 2,
				WeightType:         2,
				SolverTimeout:      c_upgradeDefaultTimeout,
			},
			SimSizeBaseline:              simSizeBaseline,
			SimSizeItemInitial:           simSizePerItem,
			ExtraSimForTopResultsCount:   2,
			ExtraSimForTopResultsSimSize: simSizeTopN,
		},
		Specs: []upgrades.FindUpgrades_Spec{
			{
				Label:                   "damage",
				Model:                   model_factory.Model_PallyProtDamage(),
				GearFile:                files.GearFileProtDamage,
				ItemFinder:              finder,
				SubstituteItems:         mygear.SubstituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "balance",
				Model:                   model_factory.Model_PallyProtBalanced(),
				GearFile:                files.GearFileProtBalanced,
				ItemFinder:              finder,
				SubstituteItems:         mygear.SubstituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "miti",
				Model:                   model_factory.Model_PallyProtMitigation(),
				GearFile:                files.GearFileProtMitigation,
				ItemFinder:              finder,
				SubstituteItems:         mygear.SubstituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "survival",
				Model:                   model_factory.Model_PallyProtSurvival(),
				GearFile:                files.GearFileProtSurvival,
				ItemFinder:              finder,
				SubstituteItems:         mygear.SubstituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "heal",
				Model:                   model_factory.Model_PallyProtHeal(),
				GearFile:                files.GearFileProtHeal,
				ItemFinder:              finder,
				SubstituteItems:         mygear.SubstituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
		},
	}

	cancel := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancel)

	upgrades.FindUpgrades_Sim_AllRaid_Run(&input, cancel)
}
