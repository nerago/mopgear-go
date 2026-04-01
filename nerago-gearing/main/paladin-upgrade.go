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
)

const (
	// simRunSize     = simulate.RunSize_Medium
	// simRunSize  = simulate.RunSize_QuickDirty

	// itemSolveSize  = solver.SolveSize_Medium
	// simRunSize     = simulate.RunSize_QuickDirty

	itemSolveSize = solver.SolveSize_Medium
	simRunSize    = simulate.RunSize_Medium

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
	96182, // ultimate prot of the emperor thunder normal
	94945, // greatshield of the gloaming normal
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
	96182, // ultimate prot of the emperor thunder normal
	94945, // greatshield of the gloaming normal
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
			IncludeRaden: false,
			IgnoredItems: ignoredItems,
			SolveSize:    itemSolveSize},
		SimSize: simRunSize}
	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, substituteItemsDps, printer)
}

func findUpgrades_Sim_PaladinMiti_Run(printer *util.PrintRecorder) {
	goal := upgrades.UpgradeGoal_Mitigation
	model := model.Model_PallyProtMitigation()
	gearFile := files.GearFileProtMitigation
	// upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	upgradeItems := loaders.ItemFinder_ThroneStrengthPlateTank(stats.Difficulty_Heroic)
	input := upgrades.FindUpgrades_SimInputs{
		FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
			IncludeRaden: false,
			IgnoredItems: ignoredItems,
			SolveSize:    itemSolveSize},
		SimSize: simRunSize}
	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, substituteItemsMiti, printer)
}

func findUpgrades_Paladin_Sim_AllRaid_Run(printer *util.PrintRecorder) {
	input := upgrades.FindUpgrades_MultiSpec_Sim{
		FindUpgrades_SimInputs: upgrades.FindUpgrades_SimInputs{
			FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
				IncludeRaden: false,
				IgnoredItems: ignoredItems,
				SolveSize:    itemSolveSize,
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
				ItemFinder:      loaders.ItemFinder_CelestialCloak,
				SubstituteItems: substituteItemsDps},
			{
				Label:           "mit",
				Goal:            upgrades.UpgradeGoal_Mitigation,
				Model:           model.Model_PallyProtMitigation(),
				GearFile:        files.GearFileProtMitigation,
				ItemFinder:      loaders.ItemFinder_CelestialCloak,
				SubstituteItems: substituteItemsMiti},
		}}
	upgrades.FindUpgrades_Sim_AllRaid_Run(&input, printer)
}

func findNeededUpgradeLevel(printer *util.PrintRecorder) {
// db.WowSimDB_ByIdFindMaxUpgrade(94942)
db.WowSimDB_ByIdFindMaxUpgrade(87024)

	itemOptions1, _ := setupPallyMitigation()
	itemOptions2, _ := setupPallyDps()

	found := map[items.ItemId]*items.FullItem{}
	for item := range itemOptions1.AllItems() {
		found[item.ItemId()] = item
	}
	for item := range itemOptions2.AllItems() {
		found[item.ItemId()] = item
	}

	printer.Println("NEED UPGRADE")
	for itemId, oldItem := range found {
		bestVersion := db.WowSimDB_ByIdFindMaxUpgrade(itemId)
		if bestVersion.Ref.ItemLevel > oldItem.Ref.ItemLevel {
			printer.Printf("%d (%d) -> %d (%d) ==> %s\n",
				oldItem.Ref.ItemLevel, oldItem.Ref.UpgradeLevel,
				bestVersion.Ref.ItemLevel, bestVersion.Ref.UpgradeLevel,
				bestVersion.CreateString())
		}
	}
}
