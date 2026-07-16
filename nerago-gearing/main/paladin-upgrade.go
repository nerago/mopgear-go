package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/upgrades"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"slices"
)

var extrasSetSpecific = []items.ItemId{
	86979,
	95142,
	95153,
	95281,
	95291,
	96375,
	96550,
	100644,
	104993,
	86955,
	96657,
	96667,
	104938,
	105090,
	94773,
	95140,
	96394,
	96447,
	96468,
	103791,
	95535,
	96478,
	96533,
	105033,
}

var substituteItemsCommon = slices.Concat(
	extrasSetSpecific,
	legendCloaks, miscOtherP3,
	retT15, retT16, protT15, protT16,
	timeless, celestial, celestialRaden, orgRaidDrops)
var substituteItemsRet = slices.Concat(substituteItemsCommon)
var substituteItemsProt = slices.Concat(substituteItemsCommon, phase3OneHandAndShield, orgOneHandAndShield)

var ignoredItems = []items.ItemId{
	63207, // org port cloak
	84661, // fishing pole
	90042} // straw hat

func findUpgrades_Sim_PaladinDps_Run(printer *util.PrintRecorder) {
	simRunSize := simulate.RunSize_QuickDirty
	goal := stats.OptimiseGoal_Dps
	model := gear_model.Model_PallyProtDps()
	gearFile := files.GearFileProtDps
	// upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	upgradeItems := loaders.ItemFinder_ThroneStrengthPlateTank(stats.Difficulty_Heroic)
	input := upgrades.FindUpgrades_SimInputs{
		FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
			IncludeNormal:      true,
			IncludeHeroic:      true,
			IgnoredItems:       ignoredItems,
			TargetUpgradeLevel: 0,
		},
		SimSizeBaseline:    simRunSize,
		SimSizeItemInitial: simRunSize}
	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, substituteItemsProt, printer)
}

func findUpgrades_Sim_PaladinMiti_Run(printer *util.PrintRecorder) {
	simRunSize := simulate.RunSize_QuickDirty
	goal := stats.OptimiseGoal_Mitigation
	model := gear_model.Model_PallyProtMitigation_WithSet()
	gearFile := files.GearFileProtMitigationWithSet
	// upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	// upgradeItems := loaders.ItemFinder_ThroneStrengthPlateTank(stats.Difficulty_Heroic)
	upgradeItems := []loaders.ItemFoundRef{{ItemId: 96436, UpgradeLevel: 2}} // tortos shell heroic
	input := upgrades.FindUpgrades_SimInputs{
		FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
			IncludeNormal:      true,
			IncludeHeroic:      true,
			IgnoredItems:       ignoredItems,
			TargetUpgradeLevel: 0,
		},
		SimSizeBaseline:    simRunSize,
		SimSizeItemInitial: simRunSize}
	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, substituteItemsProt, printer)
}

func findUpgrades_T5_Sim_PaladinMiti_Run(printer *util.PrintRecorder) {
	simRunSize := simulate.RunSize_QuickDirty
	goal := stats.OptimiseGoal_Mitigation
	model := gear_model.Model_PallyProtMitigation_WithSet()
	gearFile := files.GearFileProtMitigationWithSet
	upgradeItems := loaders.ItemFinder_SiegeStrengthPlateTank(stats.Difficulty_Normal)
	// upgradeItems := loaders.ItemFinder_SiegeStrengthPlateTank(stats.Difficulty_Heroic)
	input := upgrades.FindUpgrades_SimInputs{
		FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
			IncludeNormal:      true,
			IncludeHeroic:      true,
			IgnoredItems:       ignoredItems,
			TargetUpgradeLevel: 0,
		},
		SimSizeBaseline:    simRunSize,
		SimSizeItemInitial: simRunSize}
	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, substituteItemsProt, printer)
}

func findUpgrades_Paladin() {
	//simSizeBaseline := simulate.RunSize_VerySlow
	//simSizeTopN := simulate.RunSize_VerySlow
	simSizeBaseline := simulate.RunSize_Largish
	simSizeTopN := simulate.RunSize_Largish

	simSizePerItem := simulate.RunSize_QuickDirty
	//simSizePerItem := simulate.RunSize_Common
	//simSizePerItem := simulate.RunSize_Largish

	//simSizeBaseline := simulate.RunSize_TestOnly
	//simSizeTopN := simulate.RunSize_TestOnly
	//simSizePerItem := simulate.RunSize_TestOnly

	substituteEmptySlotOnly := make(map[items.SlotItem]items.ItemId)
	substituteEmptySlotOnly[items.Item_Trinket] = 94529 // gaze
	substituteEmptySlotOnly[items.Item_Ring] = 86957    // heroic bladed tempest ring

	//heroicBossesConsider := []string{"SoO Immerseus", "SoO Norushen", "SoO ShaofPride", "SoO FallenProtectors", "SoO Galakras", "SoO Nazgrim"}
	//finder := loaders.ItemFinder_HeroicBossFiltered(loaders.ItemFinder_SiegeStrengthPlateTank, heroicBossesConsider)

	//finder := loaders.ItemFinder_SiegeStrengthPlateTank
	//finder := loaders.ItemFinder_Ordos
	//finder := loaders.ItemFinder_TimelessPlate
	//finder := loaders.ItemFinder_BagsUpgraded
	finder := loaders.SiegeClassGearSetMultiple(stats.Spec_PaladinProt, stats.Spec_PaladinRet)
	//finder := func(_ stats.Difficulty) []loaders.ItemFoundRef { return []loaders.ItemFoundRef{{ItemId: 103735}, {ItemId: 103791}, {ItemId: 103872}}	}

	input := upgrades.FindUpgrades_MultiSpec_Sim{
		FindUpgrades_SimInputs: upgrades.FindUpgrades_SimInputs{
			FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
				IncludeCelestial:   false,
				IncludeNormal:      true,
				IncludeHeroic:      false,
				IgnoredItems:       ignoredItems,
				TargetUpgradeLevel: 2,
			},
			SimSizeBaseline:              simSizeBaseline,
			SimSizeItemInitial:           simSizePerItem,
			ExtraSimForTopResultsCount:   4,
			ExtraSimForTopResultsSimSize: simSizeTopN,
		},
		Specs: []upgrades.FindUpgrades_Spec{
			{
				Label:                   "dps",
				Model:                   gear_model.Model_PallyProtDps(),
				GearFile:                files.GearFileProtDps,
				ItemFinder:              finder,
				SubstituteItems:         substituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "compromise",
				Model:                   gear_model.Model_PallyProtCompromise(),
				GearFile:                files.GearFileProtCompromise,
				ItemFinder:              finder,
				SubstituteItems:         substituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "mit_noset",
				Model:                   gear_model.Model_PallyProtMitigation_NoSet(),
				GearFile:                files.GearFileProtMitigationNoSet,
				ItemFinder:              finder,
				SubstituteItems:         substituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "mit_set",
				Model:                   gear_model.Model_PallyProtMitigation_WithSet(),
				GearFile:                files.GearFileProtMitigationWithSet,
				ItemFinder:              finder,
				SubstituteItems:         substituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "heal",
				Model:                   gear_model.Model_PallyProtHeal(),
				GearFile:                files.GearFileProtHeal,
				ItemFinder:              finder,
				SubstituteItems:         substituteItemsProt,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
		},
	}

	cancel := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancel)

	upgrades.FindUpgrades_Sim_AllRaid_Run(&input, cancel)
}
