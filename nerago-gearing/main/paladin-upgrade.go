package main

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/upgrades"
	"paladin_gearing_go/util"
	"slices"
)

const (
	// simRunSize     = simulate.RunSize_Medium
	// simRunSize  = simulate.RunSize_QuickDirty

	itemSolveSize = solver.SolveSize_Medium
	simRunSize    = simulate.RunSize_QuickDirty

	// itemSolveSize = solver.SolveSize_Medium
	// simRunSize    = simulate.RunSize_Medium

	// itemSolveSize = solver.SolveSize_PerItem
	// simRunSize    = simulate.RunSize_TestOnly
)

var substituteItemsDps = []items.ItemId{
	96394, // frozen warlord bracer heroic
	95281, // ret tier15 gloves normal
	95205, // terra-cotta neck
	96481, // durumu tentacle heroic
	95910, // ret tier15 chest celestial
	86955, // heroic overwhelm assault belt
	86957, // heroic bladed tempest ring
	87015, // heroic clawfeet
	95140, // shado assault band
	95141, // shado assault loop
	96500, // scaled tyrant heroic
	86979, // heroic impaling treads
	96373, // cloudbreaker belt heroic
	96468, // talonrender chest heroic
	94776, // primal turtle amulet
	96533, // rein-binders fists heroic
	94820, // caustic spike bracers
	87024, // null greathelm
	94773, // centripetal shoulders normal
	95535, // normal lightning legs
	96182, // ultimate prot of the emperor thunder normal
	94945, // greatshield of the gloaming normal
	96436, // tortos shell heroic
	95778, // golden golem celestial [would need gem]
	96376, // worldbreaker weapon
	96534, // qon's scimitar
	86387, // kilrak weapon
	98147, // pre-legend strength dps
	98146, // pre-legend strength tank
	96478, // treads of the blind heroic
	95153, // Tyrant King Battleplate
	95282, // ret tier15 normal head
	95292, // prot tier15 head normal
	96657, // ret tier15 legs heroic
	96658, // ret tier15 shoulder heroic
	95142, // striker's battletags
	94776, // primal turtle amulet
	// 94527, // ji-kun trinket [not sure about trinkets here]
	// 94529, // gaze of the twins [not sure about trinkets here]
	96436, // tortos shell heroic
	96420, // talisman of angry spirits
}
var substituteItemsMiti = []items.ItemId{
	95291, // prot tier15 hand normal
	95290, // prot tier15 chest normal
	95292, // prot tier15 head normal
	96667, // prot tier15 leg heroic
	96668, // prot tier15 shoulder heroic
	96657, // ret tier15 legs heroic
	96658, // ret tier15 shoulder heroic
	95281, // ret tier15 gloves normal
	96394, // frozen warlord bracer heroic
	96373, // cloudbreaker belt heroic
	96478, // treads of the blind heroic
	95142, // striker's battletags
	95205, // terra-cotta neck
	95178, // lootraptor amulet
	96468, // talonrender chest heroic
	96533, // rein-binders fists heroic
	86957, // heroic bladed tempest ring
	86955, // heroic overwhelm assault belt
	95535, // normal lightning legs
	87015, // heroic clawfeet
	96481, // durumu tentacle heroic
	95140, // shado assault band
	95141, // shado assault loop
	96500, // scaled tyrant heroic
	96182, // ultimate prot of the emperor thunder normal
	94945, // greatshield of the gloaming normal
	96436, // tortos shell heroic
	96428, // shell-coated wrists
	96447, // rot-proof greatplate
	96376, // worldbreaker weapon
	96534, // qon's scimitar
	86387, // kilrak weapon
	98146, // pre-legend strength tank
	98147, // pre-legend strength dps
	94776, // primal turtle amulet
	94820, // caustic spike bracers
	95141, // shado assault loop
	96420, // talisman of angry spirits
}

var ignoredItems = []items.ItemId{
	63207, // org port cloak
	84661, // fishing pole
	90042} // straw hat

func findUpgrades_Sim_PaladinDps_Run(printer *util.PrintRecorder) {
	goal := upgrades.UpgradeGoal_Dps
	model := model.Model_PallyProtDps()
	gearFile := files.GearFileProtDps
	// upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	upgradeItems := loaders.ItemFinder_ThroneStrengthPlateTank(stats.Difficulty_Heroic)
	input := upgrades.FindUpgrades_SimInputs{
		FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
			IncludeNormal: true,
			IncludeHeroic: true,
			IncludeRaden:  false,
			IgnoredItems:  ignoredItems,
			SolveSize:     itemSolveSize},
		SimSize: simRunSize}
	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, substituteItemsDps, printer)
}

func findUpgrades_Sim_PaladinMiti_Run(printer *util.PrintRecorder) {
	goal := upgrades.UpgradeGoal_Mitigation
	model := model.Model_PallyProtMitigation_WithSet()
	gearFile := files.GearFileProtMitigationSet
	// upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	// upgradeItems := loaders.ItemFinder_ThroneStrengthPlateTank(stats.Difficulty_Heroic)
	upgradeItems := []*items.FullItem{db.WowSimDB_ByIdAndUpgrade_AllowFallback(96436, 2, printer)} // tortos shell heroic
	input := upgrades.FindUpgrades_SimInputs{
		FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
			IncludeNormal: true,
			IncludeHeroic: true,
			IncludeRaden:  false,
			IgnoredItems:  ignoredItems,
			SolveSize:     itemSolveSize},
		SimSize: simRunSize}
	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, substituteItemsMiti, printer)
}

func findUpgrades_Paladin_Sim_AllRaid_Run(printer *util.PrintRecorder) {
	substituteItemsDpsMiti := slices.Concat(substituteItemsDps, substituteItemsMiti)
	substituteItemsDpsMiti = util.RemoveDuplicatesComparable(substituteItemsDpsMiti)

	substituteEmptySlotOnly := make(map[items.SlotItem]items.ItemId)
	substituteEmptySlotOnly[items.Item_Trinket] = 94519
	substituteEmptySlotOnly[items.Item_Ring] = 86957

	input := upgrades.FindUpgrades_MultiSpec_Sim{
		FindUpgrades_SimInputs: upgrades.FindUpgrades_SimInputs{
			FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
				IncludeNormal: false,
				IncludeHeroic: true,
				IncludeRaden:  true,
				IgnoredItems:  ignoredItems,
				SolveSize:     itemSolveSize,
				// PositiveResultsOnly: true,
			},
			SimSize: simRunSize,
		},
		Specs: []upgrades.FindUpgrades_Spec{
			{
				Label:                   "dps",
				Goal:                    upgrades.UpgradeGoal_Dps,
				Model:                   model.Model_PallyProtDps(),
				GearFile:                files.GearFileProtDps,
				ItemFinder:              loaders.ItemFinder_ThroneStrengthPlateTank_MinusConflictStuff,
				SubstituteItems:         substituteItemsDps,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "compromise",
				Goal:                    upgrades.UpgradeGoal_HalfMitiDps,
				Model:                   model.Model_PallyProtCompromise(),
				GearFile:                files.GearFileProtCompromise,
				ItemFinder:              loaders.ItemFinder_ThroneStrengthPlateTank_MinusConflictStuff,
				SubstituteItems:         substituteItemsDpsMiti,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "mit_noset",
				Goal:                    upgrades.UpgradeGoal_Mitigation,
				Model:                   model.Model_PallyProtMitigation_NoSet(),
				GearFile:                files.GearFileProtMitigationNoSet,
				ItemFinder:              loaders.ItemFinder_ThroneStrengthPlateTank_MinusConflictStuff,
				SubstituteItems:         substituteItemsMiti,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "mit_set",
				Goal:                    upgrades.UpgradeGoal_Mitigation,
				Model:                   model.Model_PallyProtMitigation_WithSet(),
				GearFile:                files.GearFileProtMitigationSet,
				ItemFinder:              loaders.ItemFinder_ThroneStrengthPlateTank_MinusConflictStuff,
				SubstituteItems:         substituteItemsMiti,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
		},
	}
	upgrades.FindUpgrades_Sim_AllRaid_Run(&input, printer)
}
