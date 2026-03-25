package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/upgrades"
)

func findUpgrades_Sim_PaladinDps_Run() {
	mode := upgrades.Upgrade_Dps_Heroic
	model := model.Model_PallyProtDps()
	gearFile := files.GearFileProtDps
	// upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Heroic)
	substituteItems := []items.ItemId{
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
	upgrades.FindUpgrades_Sim_Run(mode, &model, gearFile, upgradeItems, substituteItems)
}

func findUpgrades_Sim_PaladinMiti_Run() {
	mode := upgrades.Upgrade_Miti_Heroic
	model := model.Model_PallyProtMitigation()
	gearFile := files.GearFileProtMitigation
	// upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Heroic)
	substituteItems := []items.ItemId{
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
	upgrades.FindUpgrades_Sim_Run(mode, &model, gearFile, upgradeItems, substituteItems)
}
