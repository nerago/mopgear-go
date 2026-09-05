package main

import (
	"github.com/nerago/mopgear-go/cmd/mygear"
	"github.com/nerago/mopgear-go/gear_model/model_factory"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/upgrades"
	"github.com/nerago/mopgear-go/util/util_async"
)

const c_upgradeDefaultTimeout = 200

func findUpgrades_Paladin() {
	//simSizeBaseline := simulate.RunSize_VerySlow
	//simSizeTopN := simulate.RunSize_VerySlow
	simSizeBaseline := simulate.RunSize_Largish
	simSizeTopN := simulate.RunSize_Largish
	//simSizeBaseline := simulate.RunSize_Common
	//simSizeTopN := simulate.RunSize_Common
	topSimCount := 4

	//simSizePerItem := simulate.RunSize_QuickDirty
	simSizePerItem := simulate.RunSize_Common
	//simSizePerItem := simulate.RunSize_Largish

	//simSizeBaseline := simulate.RunSize_TestOnly
	//simSizeTopN := simulate.RunSize_TestOnly
	//simSizePerItem := simulate.RunSize_TestOnly

	substituteEmptySlotOnly := make(map[items.SlotItem]items.ItemId)
	substituteEmptySlotOnly[items.Item_Trinket] = 94529 // gaze
	substituteEmptySlotOnly[items.Item_Ring] = 86957    // heroic bladed tempest ring

	normalBossesConsider := []string{
		//"SoO Immerseus",
		//"SoO FallenProtectors",
		//"SoO Norushen",
		//"SoO ShaofPride",
		//"SoO Galakras",
		//"SoO IronJuggernaut",
		"SoO DarkShaman",
		//"SoO Nazgrim",
		"SoO Malkorok",
		"SoO Spoils",
		"SoO Thok",
		"SoO Blackfuse",
		"SoO Paragons",
		"SoO Garrosh"}
	heroicBossesConsider := []string{
		"SoO Immerseus",
		"SoO FallenProtectors",
		"SoO Norushen",
		"SoO ShaofPride",
		"SoO Galakras",
		"SoO IronJuggernaut",
		"SoO DarkShaman",
		"SoO Nazgrim",
		"SoO Malkorok",
		"SoO Spoils",
		"SoO Thok",
		//"SoO Blackfuse",
		//"SoO Paragons",
		//"SoO Garrosh"
	}
	finderBosses := loaders.ItemFinder_NormalHeroicBossFiltered(loaders.ItemFinder_SiegeStrengthPlateTank, normalBossesConsider, heroicBossesConsider)

	//finder := loaders.ItemFinder_SiegeStrengthPlateTank
	finderOrdos := loaders.ItemFinder_Ordos
	//finder := loaders.ItemFinder_TimelessPlate
	//finder := loaders.ItemFinder_BagsUpgraded
	//finder := loaders.SiegeClassGearSetMultiple(stats.Spec_PaladinProt, stats.Spec_PaladinRet)

	// celestial world
	finderWorld := func(_ stats.Difficulty) []loaders.ItemFoundRef {
		return []loaders.ItemFoundRef{{ItemId: 99127}, {ItemId: 99137}}
	}

	_ = finderOrdos
	_ = finderWorld
	_ = finderBosses

	//finder := loaders.ItemFinderConcat(finderO, finderW)
	finder := finderWorld

	input := upgrades.FindUpgradesMultiSpec{
		Settings: upgrades.InputSettings{
			IncludeCelestial:             false,
			IncludeNormal:                true,
			IncludeHeroic:                true,
			IgnoredItems:                 mygear.IgnoredItems,
			TargetUpgradeLevel:           2,
			WeightType:                   1,
			SolverTimeout:                c_upgradeDefaultTimeout,
			SimSizeBaseline:              simSizeBaseline,
			SimSizeItemInitial:           simSizePerItem,
			ExtraSimForTopResultsCount:   topSimCount,
			ExtraSimForTopResultsSimSize: simSizeTopN,
		},
		Specs: []upgrades.SpecInput{
			{
				Label:                   "ret",
				Model:                   model_factory.Model_PallyRet(),
				ItemFinder:              finder,
				SubstituteItems:         mygear.SubstituteItemsRet,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "damage",
				Model:                   model_factory.Model_PallyProtDamage(),
				ItemFinder:              finder,
				SubstituteItems:         mygear.SubstituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "balance",
				Model:                   model_factory.Model_PallyProtBalanced(),
				ItemFinder:              finder,
				SubstituteItems:         mygear.SubstituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "miti",
				Model:                   model_factory.Model_PallyProtMitigation(),
				ItemFinder:              finder,
				SubstituteItems:         mygear.SubstituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "survival",
				Model:                   model_factory.Model_PallyProtSurvival(),
				ItemFinder:              finder,
				SubstituteItems:         mygear.SubstituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "heal",
				Model:                   model_factory.Model_PallyProtHeal(),
				ItemFinder:              finder,
				SubstituteItems:         mygear.SubstituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
		},
	}

	cancel := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancel)

	upgrades.FindUpgrades_Run(&input, cancel)
}
