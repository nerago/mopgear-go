package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/upgrades"
)

const (
	// simRunSize     = simulate.RunSize_Medium
	// simRunSize  = simulate.RunSize_QuickDirty

	// itemSolveSize  = solver.SolveSize_Medium
	// simRunSize     = simulate.RunSize_QuickDirty

	itemSolveSize = solver.SolveSize_Medium
	simRunSize  = simulate.RunSize_Medium

	// itemSolveSize = solver.SolveSize_PerItem
	// simRunSize    = simulate.RunSize_TestOnly
)

var substituteItemsDps = []items.ItemId{
	87026, // heroic peacock cloak
	96394, // frozen warlord bracer heroic
	95281, // ret tier15 gloves normal
	95205, // terra-cotta neck
	96481, // durumu tentacle heroic
	95910, // ret tier15 chest celestial
	86955, // heroic overwhelm assault belt
	86957, // heroic bladed tempest ring
	87015, // heroic clawfeet
	95140, // shado assault band
	86979, // heroic impaling treads
	96373, // cloudbreaker belt heroic
	96468, // talonrender chest heroic
	94776, // primal turtle amulet
	96533, // rein-binders fists heroic
	94820, // caustic spike bracers
	94942, // hydra bloodcloak
	87024, // null greathelm
	94773, // centripetal shoulders normal
	95513, // scaled tyrant normal
	95535, // normal lightning legs
}
var substituteItemsMiti = []items.ItemId{
	95291, // prot tier15 hand normal
	95290, // prot tier15 chest normal
	95292, // prot tier15 head normal
	96667, // prot tier15 leg heroic
	96668, // prot tier15 shoulder heroic
	96657, // ret tier15 legs heroic
	96769, // doomcloak
	96394, // frozen warlord bracer heroic
	96373, // cloudbreaker belt heroic
	96478, // treads of the blind heroic
	95142, // striker's battletags
	95205, // terra-cotta neck
	95178, // lootraptor amulet
	96533, // rein-binders fists heroic
	86957, // heroic bladed tempest ring
	86955, // heroic overwhelm assault belt
	95535, // normal lightning legs
	87015, // heroic clawfeet
	96481, // durumu tentacle heroic
	95513, // scaled tyrant normal
	95140, // shado assault band
}

var ignoredItems = []items.ItemId{
	63207, // org port cloak
	84661, // fishing pole
	90042} // straw hat

func findUpgrades_Sim_PaladinDps_Run() {
	goal := upgrades.UpgradeGoal_Dps
	model := model.Model_PallyProtDps()
	gearFile := files.GearFileProtDps
	// upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Heroic)
	input := upgrades.FindUpgrades_SimInputs{
		FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
			IgnoredItems: ignoredItems,
			SolveSize:    itemSolveSize},
		SimSize: simRunSize}
	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, substituteItemsDps)
}

func findUpgrades_Sim_PaladinMiti_Run() {
	goal := upgrades.UpgradeGoal_Mitigation
	model := model.Model_PallyProtMitigation()
	gearFile := files.GearFileProtMitigation
	// upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Heroic)
	input := upgrades.FindUpgrades_SimInputs{
		FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
			IgnoredItems: ignoredItems,
			SolveSize:    itemSolveSize},
		SimSize: simRunSize}
	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, substituteItemsMiti)
}

func findUpgrades_Paladin_Sim_AllRaid_Run() {
	input := upgrades.FindUpgrades_MultiSpec_Sim{
		FindUpgrades_SimInputs: upgrades.FindUpgrades_SimInputs{
			FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
				IgnoredItems:        ignoredItems,
				SolveSize:           itemSolveSize,
				// PositiveResultsOnly: true,
			},
			SimSize: simRunSize,
		},
		Specs: []upgrades.FindUpgrades_Spec{
			{
				Label:           "dps",
				Goal:            upgrades.UpgradeGoal_Dps,
				Model:           model.Model_PallyProtDps(),
				GearFile:        files.GearFileProtDps,
				ItemFinder:      loaders.ItemFinder_ThroneProtMinusRaden,
				SubstituteItems: substituteItemsDps},
			{
				Label:           "mit",
				Goal:            upgrades.UpgradeGoal_Mitigation,
				Model:           model.Model_PallyProtMitigation(),
				GearFile:        files.GearFileProtMitigation,
				ItemFinder:      loaders.ItemFinder_ThroneProtMinusRaden,
				SubstituteItems: substituteItemsMiti},
		}}
	upgrades.FindUpgrades_Sim_AllRaid_Run(&input)
}
