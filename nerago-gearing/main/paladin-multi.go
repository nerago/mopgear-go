package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/multi"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/util"
)

func PaladinMultiRun(printer *util.PrintRecorder) {
	job := multi.MultiSetJob_Create(printer)

	ret := multi.MultiSetParam{
		Label:                "Ret",
		GearFile:             files.GearFileRet,
		Model:                model.Model_PallyRet(),
		IncludeInFirstPass:   false,
		RequestRatingPercent: 0.02,
		PhasedAcceptable:     false,
		ExtraUpgradeLevel:    2}
	ret.AddExtraItems([]items.ItemId{
		// 87026, // heroic peacock cloak
		94942, // hydra bloodcloak
		96769, // doomcloak

		95140, // shado assault band
		86957, // heroic bladed tempest ring
		95513, // scaled tyrant normal

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		87024, // null greathelm
		// 86955, // heroic overwhelm assault belt
		94726, // cloudbreaker belt normal
		96373, // cloudbreaker belt heroic

		95535, // normal lightning legs
		// 94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic

		// 85340, // ret tier14 legs
		//    87101, // ret tier14 head [would need to regem, AVOID]
		// 85339, // ret tier14 shoulder
		85343, // ret tier14 chest
		87100, // ret tier14 hands

		95910, // ret tier15 chest celestial
		95281, // ret tier15 gloves normal
		96657, // ret tier15 legs heroic
		96658, // ret tier15 shoulder heroic
		//    95912, // ret tier15 celestial head (don't have yet)

		95142, // striker's battletags
		95205, // terra-cotta neck
		94776, // primal turtle amulet

		87145, // defiled earth
		94820, // caustic spike bracers
		96394, // frozen warlord bracer heroic
	})
	// ret.AddBagsExtra()
	ret.AddFixedSlot(items.Equip_Ring2, 95513) // scaled tyrant normal

	protDps := multi.MultiSetParam{
		Label:                "Prot-Damage",
		GearFile:             files.GearFileProtDps,
		Model:                model.Model_PallyProtDps(),
		IncludeInFirstPass:   true,
		RequestRatingPercent: 0.45,
		PhasedAcceptable:     false,
		ExtraUpgradeLevel:    2}
	protDps.AddExtraItems([]items.ItemId{
		86957, // heroic bladed tempest ring
		95140, // shado assault band
		// 86946, // ruby signet heroic
		95513, // scaled tyrant normal
		96481, // durumu tentacle heroic

		87015, // heroic clawfeet
		86979, // heroic impaling treads
		94726, // cloudbreaker belt normal
		96373, // cloudbreaker belt heroic
		87024, // null greathelm
		94942, // hydra bloodcloak
		96769, // doomcloak

		87026, // heroic peacock cloak
		// 86955, // heroic overwhelm assault belt
		95535, // normal lightning legs
		94773, // centripetal shoulders normal
		96468, // talonrender chest heroic
		96533, // rein-binders fists heroic
		// 96478, // treads of the blind heroic

		// 85343, // White Tiger Battleplate 504

		95912, // ret tier15 celestial head (don't have yet)
		95910, // ret tier15 chest celestial
		// 95281, // ret tier15 gloves normal
		96657, // ret tier15 legs heroic
		96658, // ret tier15 shoulder heroic

		// 95291, // prot tier15 hand normal
		// 95290, // prot tier15 chest normal
		95292, // prot tier15 head normal
		// 96667, // prot tier15 leg heroic
		// 96668, // prot tier15 shoulder heroic

		95142, // striker's battletags
		95205, // terra-cotta neck
		// 87036, // soulgrasp heroic
		94776, // primal turtle amulet

		96182, // ultimate prot of the emperor thunder normal
		94945, // greatshield of the gloaming normal

		// 87145, // defiled earth
		94820, // caustic spike bracers
		96394, // frozen warlord bracer heroic

		96376, // worldbreaker weapon
	})
	// protDps.AddBagsExtra()
	protDps.AddFixedSlot(items.Equip_Ring2, 95513) // scaled tyrant normal
	// protDps.AddFixedSlot(items.Equip_Offhand, 94945) // greatshield of the gloaming normal

	protMitigation := multi.MultiSetParam{
		Label:                "Prot-Mitigation",
		GearFile:             files.GearFileProtMitigation,
		Model:                model.Model_PallyProtMitigation(),
		IncludeInFirstPass:   true,
		RequestRatingPercent: 0.53,
		PhasedAcceptable:     false,
		ExtraUpgradeLevel:    2}
	protMitigation.AddExtraItems([]items.ItemId{
		86979, // heroic impaling treads
		87015, // heroic clawfeet

		86957, // heroic bladed tempest ring
		// 86946, // ruby signet heroic
		95140, // shado assault band
		95513, // scaled tyrant normal
		96481, // durumu tentacle heroic

		94726, // cloudbreaker belt normal
		96373, // cloudbreaker belt heroic
		// 86955, // heroic overwhelm assault belt

		// 87026, // heroic peacock cloak
		94942, // hydra bloodcloak
		96769, // doomcloak

		95535, // normal lightning legs
		// 94773, // centripetal shoulders normal
		// 96468, // talonrender chest heroic
		96533, // rein-binders fists heroic
		96478, // treads of the blind heroic

		// 95910, // ret tier15 chest celestial
		// 95281, // ret tier15 gloves normal
		96657, // ret tier15 legs heroic
		96658, // ret tier15 shoulder heroic

		95291, // prot tier15 hand normal
		95290, // prot tier15 chest normal
		95292, // prot tier15 head normal
		96667, // prot tier15 leg heroic
		96668, // prot tier15 shoulder heroic

		95142, // striker's battletags
		95205, // terra-cotta neck
		95178, // lootraptor amulet
		94776, // primal turtle amulet

		// 96182, // ultimate prot of the emperor thunder normal
		94945, // greatshield of the gloaming normal

		87145, // defiled earth
		// 94820, // caustic spike bracers
		96394, // frozen warlord bracer heroic

		96376, // worldbreaker weapon [could do a duplicate version]
		86387, // kilrak weapon
	})
	// protMitigation.AddBagsExtra()
	protMitigation.AddFixedSlot(items.Equip_Ring1, 96481)   // durumu
	protMitigation.AddFixedSlot(items.Equip_Offhand, 94945) // greatshield of the gloaming normal

	job.AddSetParam(ret)
	job.AddSetParam(protDps)
	job.AddSetParam(protMitigation)

	// job.AddSuppressSlotCheck(86946) // Vizier's Ruby Signet
	job.AddSuppressSlotCheck(86957) // Ring of the Bladed Tempest
	job.AddSuppressSlotCheck(95140) // Band of the Shado-Pan Assault
	job.AddSuppressSlotCheck(96481) // Durumu's Severed Tentacle

	// job.AddSuppressSlotCheck(94529) // Gaze of the Twins
	// job.AddSuppressSlotCheck(94527) // Ji-Kun's Rising Winds

	// job.SuggestCulls(100, 2)
	// job.SuggestCulls(500, 10)
	//job.SuggestCulls(5000, 10)
	// job.SuggestCulls(10000, 100)
	// job.SuggestCulls(25000, 150)

	// job.FindTopAndPassToSim(50, 2, simulate.RunSize_Medium)
	// job.FindTopAndPassToSim(150000, 50, simulate.RunSize_Medium)
	// job.FindTopAndPassToSim(75000, 20, simulate.RunSize_Medium)
	job.FindTopAndPassToSim(150000, 25, simulate.RunSize_Medium)
}
