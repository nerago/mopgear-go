package main

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/upgrades"
	"paladin_gearing_go/util"
	"slices"
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
	95291,  // prot tier15 hand normal
	95290,  // prot tier15 chest normal
	95292,  // prot tier15 head normal
	96667,  // prot tier15 leg heroic
	96668,  // prot tier15 shoulder heroic
	96657,  // ret tier15 legs heroic
	96658,  // ret tier15 shoulder heroic
	95281,  // ret tier15 gloves normal
	96394,  // frozen warlord bracer heroic
	96373,  // cloudbreaker belt heroic
	96478,  // treads of the blind heroic
	95142,  // striker's battletags
	95205,  // terra-cotta neck
	95178,  // lootraptor amulet
	96468,  // talonrender chest heroic
	96533,  // rein-binders fists heroic
	86957,  // heroic bladed tempest ring
	86955,  // heroic overwhelm assault belt
	95535,  // normal lightning legs
	87015,  // heroic clawfeet
	96481,  // durumu tentacle heroic
	95140,  // shado assault band
	95141,  // shado assault loop
	96500,  // scaled tyrant heroic
	96182,  // ultimate prot of the emperor thunder normal
	94945,  // greatshield of the gloaming normal
	96436,  // tortos shell heroic
	96428,  // shell-coated wrists
	96447,  // rot-proof greatplate
	96376,  // worldbreaker weapon
	96534,  // qon's scimitar
	86387,  // kilrak weapon
	98146,  // pre-legend strength tank
	98147,  // pre-legend strength dps
	94776,  // primal turtle amulet
	94820,  // caustic spike bracers
	95141,  // shado assault loop
	96420,  // talisman of angry spirits
	101882, // cliffbreaker helm exp/mastery
	103787, // poisonbinder girth
	103742, // blood rage bracers
	99126,  // prot t16 chest normal
	103738, // bubble bracers
}

var ignoredItems = []items.ItemId{
	63207, // org port cloak
	84661, // fishing pole
	90042} // straw hat

func findUpgrades_Sim_PaladinDps_Run(printer *util.PrintRecorder) {
	simRunSize := simulate.RunSize_QuickDirty
	goal := stats.OptimiseGoal_Dps
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
		},
		SimSize: simRunSize}
	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, substituteItemsDps, printer)
}

func findUpgrades_Sim_PaladinMiti_Run(printer *util.PrintRecorder) {
	simRunSize := simulate.RunSize_QuickDirty
	goal := stats.OptimiseGoal_Mitigation
	model := model.Model_PallyProtMitigation_WithSet()
	gearFile := files.GearFileProtMitigationWithSet
	// upgradeItems := loaders.ItemFinder_ThroneProtMinusRaden(stats.Difficulty_Normal)
	// upgradeItems := loaders.ItemFinder_ThroneStrengthPlateTank(stats.Difficulty_Heroic)
	upgradeItems := []*items.FullItem{db.WowSimDB_ByIdAndUpgrade_AllowFallback(96436, 2, printer)} // tortos shell heroic
	input := upgrades.FindUpgrades_SimInputs{
		FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
			IncludeNormal: true,
			IncludeHeroic: true,
			IncludeRaden:  false,
			IgnoredItems:  ignoredItems,
		},
		SimSize: simRunSize}
	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, substituteItemsMiti, printer)
}

func findUpgrades_T5_Sim_PaladinMiti_Run(printer *util.PrintRecorder) {
	simRunSize := simulate.RunSize_QuickDirty
	goal := stats.OptimiseGoal_Mitigation
	model := model.Model_PallyProtMitigation_WithSet()
	gearFile := files.GearFileProtMitigationWithSet
	upgradeItems := loaders.ItemFinder_SiegeStrengthPlateTank(stats.Difficulty_Normal)
	// upgradeItems := loaders.ItemFinder_SiegeStrengthPlateTank(stats.Difficulty_Heroic)
	input := upgrades.FindUpgrades_SimInputs{
		FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
			IncludeNormal: true,
			IncludeHeroic: true,
			IncludeRaden:  false,
			IgnoredItems:  ignoredItems,
		},
		SimSize: simRunSize}
	upgrades.FindUpgrades_Sim_Run(&input, goal, &model, gearFile, upgradeItems, substituteItemsMiti, printer)
}

func findUpgrades_Paladin_Sim_AllRaid_Run(printer *util.PrintRecorder) {
	// simRunSize    = simulate.RunSize_TestOnly
	// var simRunSize simulate.WowSim_RunSize = 1500
	// var simRunSize simulate.WowSim_RunSize = 3000
	simRunSize := simulate.RunSize_QuickDirty
	// simRunSize    := simulate.RunSize_Medium

	substituteItemsDpsAndMiti := slices.Concat(substituteItemsDps, substituteItemsMiti)
	substituteItemsDpsAndMiti = util.RemoveDuplicatesComparable(substituteItemsDpsAndMiti)

	substituteEmptySlotOnly := make(map[items.SlotItem]items.ItemId)
	substituteEmptySlotOnly[items.Item_Trinket] = 94529 // gaze
	substituteEmptySlotOnly[items.Item_Ring] = 86957    // heroic bladed tempest ring

	// finder := func(_ stats.Difficulty) []*items.FullItem {
	// 	return []*items.FullItem{db.WowSimDB_ByIdAndUpgrade(99128, 2), db.WowSimDB_ByIdAndUpgrade(99138, 2)}
	// }
	finder := loaders.ItemFinder_SiegeStrengthPlateTank
	// finder := loaders.ItemFinder_Ordos
	// finder := loaders.ItemFinder_TimelessPlate

	input := upgrades.FindUpgrades_MultiSpec_Sim{
		FindUpgrades_SimInputs: upgrades.FindUpgrades_SimInputs{
			FindUpgrades_BasicInputs: upgrades.FindUpgrades_BasicInputs{
				IncludeCelestial: true,
				IncludeNormal:    false,
				IncludeHeroic:    false,
				IncludeRaden:     false,
				IgnoredItems:     ignoredItems,
			},
			SimSize: simRunSize,
		},
		Specs: []upgrades.FindUpgrades_Spec{
			{
				Label:                   "dps",
				Model:                   model.Model_PallyProtDps(),
				GearFile:                files.GearFileProtDps,
				ItemFinder:              finder,
				SubstituteItems:         substituteItemsDps,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "compromise",
				Model:                   model.Model_PallyProtCompromise(),
				GearFile:                files.GearFileProtCompromise,
				ItemFinder:              finder,
				SubstituteItems:         substituteItemsDpsAndMiti,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "mit_noset",
				Model:                   model.Model_PallyProtMitigation_NoSet(),
				GearFile:                files.GearFileProtMitigationNoSet,
				ItemFinder:              finder,
				SubstituteItems:         substituteItemsMiti,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
			{
				Label:                   "mit_set",
				Model:                   model.Model_PallyProtMitigation_WithSet(),
				GearFile:                files.GearFileProtMitigationWithSet,
				ItemFinder:              finder,
				SubstituteItems:         substituteItemsMiti,
				SubstituteEmptySlotOnly: substituteEmptySlotOnly,
			},
		},
	}
	upgrades.FindUpgrades_Sim_AllRaid_Run(&input, printer)
}
